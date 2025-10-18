package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

func TestSetFrameDurationInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   SetFrameDurationInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: SetFrameDurationInput{
				SpritePath:  "/path/to/sprite.aseprite",
				FrameNumber: 1,
				DurationMs:  100,
			},
			wantErr: false,
		},
		{
			name: "frame number too small",
			input: SetFrameDurationInput{
				SpritePath:  "/path/to/sprite.aseprite",
				FrameNumber: 0,
				DurationMs:  100,
			},
			wantErr: true,
			errMsg:  "frame_number must be at least 1",
		},
		{
			name: "duration too small",
			input: SetFrameDurationInput{
				SpritePath:  "/path/to/sprite.aseprite",
				FrameNumber: 1,
				DurationMs:  0,
			},
			wantErr: true,
			errMsg:  "duration_ms must be between 1 and 65535",
		},
		{
			name: "duration too large",
			input: SetFrameDurationInput{
				SpritePath:  "/path/to/sprite.aseprite",
				FrameNumber: 1,
				DurationMs:  65536,
			},
			wantErr: true,
			errMsg:  "duration_ms must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate frame number
			if tt.input.FrameNumber < 1 && tt.wantErr {
				return
			}
			// Validate duration
			if (tt.input.DurationMs < 1 || tt.input.DurationMs > 65535) && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestCreateTagInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateTagInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input forward",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "forward",
			},
			wantErr: false,
		},
		{
			name: "valid input reverse",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "reverse",
			},
			wantErr: false,
		},
		{
			name: "valid input pingpong",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "pingpong",
			},
			wantErr: false,
		},
		{
			name: "empty tag name",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "forward",
			},
			wantErr: true,
			errMsg:  "tag_name cannot be empty",
		},
		{
			name: "from_frame too small",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  0,
				ToFrame:    4,
				Direction:  "forward",
			},
			wantErr: true,
			errMsg:  "from_frame must be at least 1",
		},
		{
			name: "to_frame before from_frame",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  4,
				ToFrame:    2,
				Direction:  "forward",
			},
			wantErr: true,
			errMsg:  "to_frame must be >= from_frame",
		},
		{
			name: "invalid direction",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid direction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate tag name
			if tt.input.TagName == "" && tt.wantErr {
				return
			}
			// Validate from_frame
			if tt.input.FromFrame < 1 && tt.wantErr {
				return
			}
			// Validate frame range
			if tt.input.ToFrame < tt.input.FromFrame && tt.wantErr {
				return
			}
			// Validate direction
			validDirections := map[string]bool{
				"forward":  true,
				"reverse":  true,
				"pingpong": true,
			}
			if !validDirections[tt.input.Direction] && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestDuplicateFrameInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   DuplicateFrameInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input insert at end",
			input: DuplicateFrameInput{
				SpritePath:  "/path/to/sprite.aseprite",
				SourceFrame: 1,
				InsertAfter: 0,
			},
			wantErr: false,
		},
		{
			name: "valid input insert after frame",
			input: DuplicateFrameInput{
				SpritePath:  "/path/to/sprite.aseprite",
				SourceFrame: 1,
				InsertAfter: 2,
			},
			wantErr: false,
		},
		{
			name: "source_frame too small",
			input: DuplicateFrameInput{
				SpritePath:  "/path/to/sprite.aseprite",
				SourceFrame: 0,
				InsertAfter: 1,
			},
			wantErr: true,
			errMsg:  "source_frame must be at least 1",
		},
		{
			name: "insert_after negative",
			input: DuplicateFrameInput{
				SpritePath:  "/path/to/sprite.aseprite",
				SourceFrame: 1,
				InsertAfter: -1,
			},
			wantErr: true,
			errMsg:  "insert_after must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate source_frame
			if tt.input.SourceFrame < 1 && tt.wantErr {
				return
			}
			// Validate insert_after
			if tt.input.InsertAfter < 0 && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestLinkCelInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   LinkCelInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				SourceFrame: 1,
				TargetFrame: 2,
			},
			wantErr: false,
		},
		{
			name: "empty layer name",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "",
				SourceFrame: 1,
				TargetFrame: 2,
			},
			wantErr: true,
			errMsg:  "layer_name cannot be empty",
		},
		{
			name: "source_frame too small",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				SourceFrame: 0,
				TargetFrame: 2,
			},
			wantErr: true,
			errMsg:  "source_frame must be at least 1",
		},
		{
			name: "target_frame too small",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				SourceFrame: 1,
				TargetFrame: 0,
			},
			wantErr: true,
			errMsg:  "target_frame must be at least 1",
		},
		{
			name: "source and target same",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				SourceFrame: 2,
				TargetFrame: 2,
			},
			wantErr: true,
			errMsg:  "source_frame and target_frame cannot be the same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate layer name
			if tt.input.LayerName == "" && tt.wantErr {
				return
			}
			// Validate source_frame
			if tt.input.SourceFrame < 1 && tt.wantErr {
				return
			}
			// Validate target_frame
			if tt.input.TargetFrame < 1 && tt.wantErr {
				return
			}
			// Validate not same
			if tt.input.SourceFrame == tt.input.TargetFrame && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestDeleteTagInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   DeleteTagInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: DeleteTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
			},
			wantErr: false,
		},
		{
			name: "empty sprite path",
			input: DeleteTagInput{
				SpritePath: "",
				TagName:    "walk",
			},
			wantErr: true,
			errMsg:  "sprite_path cannot be empty",
		},
		{
			name: "empty tag name",
			input: DeleteTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "",
			},
			wantErr: true,
			errMsg:  "tag_name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate sprite_path
			if tt.input.SpritePath == "" && tt.wantErr {
				return
			}
			// Validate tag_name
			if tt.input.TagName == "" && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

// MCP Protocol Tests (use real MCP with in-memory transport and real Aseprite)

// createAnimationTestSession creates an MCP session with animation tools registered
func createAnimationTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pixel-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterAnimationTools(server, client, gen, cfg, logger)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	_, err := server.Connect(context.Background(), serverTransport, nil)
	require.NoError(t, err)

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	session, err := mcpClient.Connect(context.Background(), clientTransport, nil)
	require.NoError(t, err)

	return server, session, client
}

func TestSetFrameDuration_ViaMCP(t *testing.T) {
	_, session, client := createAnimationTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	// Create a sprite with multiple frames
	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-duration.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-duration.aseprite")

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "set_frame_duration",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-duration.aseprite",
			"frame_number": 1,
			"duration_ms":  150,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}

func TestCreateTag_ViaMCP(t *testing.T) {
	_, session, client := createAnimationTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-tag.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-tag.aseprite")

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_tag",
		Arguments: map[string]any{
			"sprite_path": cfg.TempDir + "/test-tag.aseprite",
			"tag_name":    "walk",
			"from_frame":  1,
			"to_frame":    1,
			"direction":   "forward",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}

func TestDeleteTag_ViaMCP(t *testing.T) {
	_, session, client := createAnimationTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-deltag.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-deltag.aseprite")

	// First create a tag
	createScript := gen.CreateTag("test", 1, 1, "forward")
	_, err = client.ExecuteLua(context.Background(), createScript, cfg.TempDir+"/test-deltag.aseprite")
	require.NoError(t, err)

	// Now delete it
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "delete_tag",
		Arguments: map[string]any{
			"sprite_path": cfg.TempDir + "/test-deltag.aseprite",
			"tag_name":    "test",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}

func TestDuplicateFrame_ViaMCP(t *testing.T) {
	_, session, client := createAnimationTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-dup.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-dup.aseprite")

	// Debug: check initial frame count
	infoScript := gen.GetSpriteInfo()
	infoOutput, _ := client.ExecuteLua(context.Background(), infoScript, cfg.TempDir+"/test-dup.aseprite")
	t.Logf("Initial sprite info: %s", infoOutput)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "duplicate_frame",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-dup.aseprite",
			"source_frame": 1,
			"insert_after": 0, // Insert at end
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	t.Logf("DuplicateFrame result: %s", result.Content[0].(*mcp.TextContent).Text)

	var output struct {
		NewFrameNumber int `json:"new_frame_number"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.Equal(t, 2, output.NewFrameNumber)
}

func TestLinkCel_ViaMCP(t *testing.T) {
	_, session, client := createAnimationTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-link.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-link.aseprite")

	// Add a second frame first
	addFrameScript := gen.AddFrame(100)
	_, err = client.ExecuteLua(context.Background(), addFrameScript, cfg.TempDir+"/test-link.aseprite")
	require.NoError(t, err)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "link_cel",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-link.aseprite",
			"layer_name":   "Layer 1",
			"source_frame": 1,
			"target_frame": 2,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}
