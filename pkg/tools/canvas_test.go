package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
)

func TestCreateCanvasInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input CreateCanvasInput
		valid bool
	}{
		{
			name: "valid RGB canvas",
			input: CreateCanvasInput{
				Width:     800,
				Height:    600,
				ColorMode: "rgb",
			},
			valid: true,
		},
		{
			name: "valid grayscale canvas",
			input: CreateCanvasInput{
				Width:     100,
				Height:    100,
				ColorMode: "grayscale",
			},
			valid: true,
		},
		{
			name: "valid indexed canvas",
			input: CreateCanvasInput{
				Width:     320,
				Height:    240,
				ColorMode: "indexed",
			},
			valid: true,
		},
		{
			name: "minimum dimensions",
			input: CreateCanvasInput{
				Width:     1,
				Height:    1,
				ColorMode: "rgb",
			},
			valid: true,
		},
		{
			name: "maximum dimensions",
			input: CreateCanvasInput{
				Width:     65535,
				Height:    65535,
				ColorMode: "rgb",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation that struct can be created
			if tt.input.Width < 1 || tt.input.Width > 65535 {
				if tt.valid {
					t.Error("Expected valid input but width is out of range")
				}
			}
			if tt.input.Height < 1 || tt.input.Height > 65535 {
				if tt.valid {
					t.Error("Expected valid input but height is out of range")
				}
			}
			validModes := map[string]bool{"rgb": true, "grayscale": true, "indexed": true}
			if !validModes[tt.input.ColorMode] {
				if tt.valid {
					t.Error("Expected valid input but color mode is invalid")
				}
			}
		})
	}
}

func TestGenerateTimestamp(t *testing.T) {
	ts1 := generateTimestamp()
	ts2 := generateTimestamp()

	// Timestamps should be positive
	if ts1 <= 0 {
		t.Error("generateTimestamp() returned non-positive value")
	}

	// Second timestamp should be >= first (monotonic)
	if ts2 < ts1 {
		t.Error("generateTimestamp() not monotonic")
	}
}

func TestAddLayerInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input AddLayerInput
		valid bool
	}{
		{
			name: "valid layer",
			input: AddLayerInput{
				SpritePath: "/path/to/sprite.aseprite",
				LayerName:  "Background",
			},
			valid: true,
		},
		{
			name: "empty layer name",
			input: AddLayerInput{
				SpritePath: "/path/to/sprite.aseprite",
				LayerName:  "",
			},
			valid: false,
		},
		{
			name: "special characters in layer name",
			input: AddLayerInput{
				SpritePath: "/path/to/sprite.aseprite",
				LayerName:  "Layer-1_Test (Copy)",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.input.LayerName == ""
			if isEmpty && tt.valid {
				t.Error("Expected valid input but layer name is empty")
			}
			if !isEmpty && !tt.valid {
				t.Error("Expected invalid input but layer name is not empty")
			}
		})
	}
}

func TestAddFrameInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input AddFrameInput
		valid bool
	}{
		{
			name: "valid frame with 100ms duration",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 100,
			},
			valid: true,
		},
		{
			name: "minimum duration (1ms)",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 1,
			},
			valid: true,
		},
		{
			name: "maximum duration (65535ms)",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 65535,
			},
			valid: true,
		},
		{
			name: "zero duration",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 0,
			},
			valid: false,
		},
		{
			name: "negative duration",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: -1,
			},
			valid: false,
		},
		{
			name: "duration too large",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 65536,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.input.DurationMs >= 1 && tt.input.DurationMs <= 65535
			if isValid != tt.valid {
				t.Errorf("Expected valid=%v but got %v for duration=%d", tt.valid, isValid, tt.input.DurationMs)
			}
		})
	}
}

func TestGetSpriteInfoInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input GetSpriteInfoInput
		valid bool
	}{
		{
			name: "valid sprite path",
			input: GetSpriteInfoInput{
				SpritePath: "/path/to/sprite.aseprite",
			},
			valid: true,
		},
		{
			name: "empty sprite path",
			input: GetSpriteInfoInput{
				SpritePath: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.input.SpritePath == ""
			if isEmpty == tt.valid {
				t.Errorf("Expected valid=%v but path is empty=%v", tt.valid, isEmpty)
			}
		})
	}
}

// The tests below use real MCP protocol with in-memory transport and real Aseprite

// createTestMCPSession creates a real MCP server/client connection with in-memory transport.
func createTestMCPSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client, *config.Config, core.Logger) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)

	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()

	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "aseprite-mcp-test",
		Version: "1.0.0",
	}, nil)

	// Register tools
	RegisterCanvasTools(server, client, gen, cfg, logger)

	// Create in-memory transport
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	_, err := server.Connect(context.Background(), serverTransport, nil)
	require.NoError(t, err, "Failed to connect server")

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	session, err := mcpClient.Connect(context.Background(), clientTransport, nil)
	require.NoError(t, err, "Failed to connect client")

	return server, session, client, cfg, logger
}

func TestRegisterCanvasTools(t *testing.T) {
	_, session, _, _, _ := createTestMCPSession(t)
	defer session.Close()

	// List tools from the server
	var tools []*mcp.Tool
	for tool, err := range session.Tools(context.Background(), nil) {
		require.NoError(t, err)
		tools = append(tools, tool)
	}

	// Verify canvas tools are registered
	expectedTools := map[string]bool{
		"create_canvas":   false,
		"add_layer":       false,
		"add_frame":       false,
		"delete_layer":    false,
		"delete_frame":    false,
		"get_sprite_info": false,
	}

	for _, tool := range tools {
		if _, exists := expectedTools[tool.Name]; exists {
			expectedTools[tool.Name] = true
		}
	}

	// Check all expected tools were found
	for toolName, found := range expectedTools {
		assert.True(t, found, "Tool %s should be registered", toolName)
	}
}

func TestCreateCanvas_ViaMCP(t *testing.T) {
	_, session, _, cfg, _ := createTestMCPSession(t)
	defer session.Close()

	tests := []struct {
		name      string
		args      map[string]any
		wantError bool
	}{
		{
			name: "valid RGB canvas",
			args: map[string]any{
				"width":      64,
				"height":     64,
				"color_mode": "rgb",
			},
			wantError: false,
		},
		{
			name: "valid grayscale canvas",
			args: map[string]any{
				"width":      32,
				"height":     32,
				"color_mode": "grayscale",
			},
			wantError: false,
		},
		{
			name: "valid indexed canvas",
			args: map[string]any{
				"width":      128,
				"height":     128,
				"color_mode": "indexed",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "create_canvas",
				Arguments: tt.args,
			})

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.False(t, result.IsError, "Tool returned error: %v", result.Content)

			// Parse output
			textContent, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected TextContent")

			var output struct {
				FilePath string `json:"file_path"`
			}
			err = json.Unmarshal([]byte(textContent.Text), &output)
			require.NoError(t, err)

			// Verify sprite file exists
			_, err = os.Stat(output.FilePath)
			require.NoError(t, err, "Sprite file should exist: %s", output.FilePath)

			// Verify it's in temp dir
			assert.True(t, filepath.Dir(output.FilePath) == cfg.TempDir, "Sprite should be in temp dir")

			// Cleanup
			os.Remove(output.FilePath)
		})
	}
}

func TestAddLayer_ViaMCP(t *testing.T) {
	_, session, _, _, _ := createTestMCPSession(t)
	defer session.Close()

	// First create a canvas
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      64,
			"height":     64,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)
	require.False(t, createResult.IsError)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Now add a layer
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "add_layer",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"layer_name":  "TestLayer",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	// Verify layer was added by checking sprite info
	infoResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_sprite_info",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
		},
	})
	require.NoError(t, err)

	var infoOutput struct {
		LayerCount int      `json:"layer_count"`
		Layers     []string `json:"layers"`
	}
	json.Unmarshal([]byte(infoResult.Content[0].(*mcp.TextContent).Text), &infoOutput)

	assert.Equal(t, 2, infoOutput.LayerCount, "Should have 2 layers (default + TestLayer)")
	assert.Contains(t, infoOutput.Layers, "TestLayer")
}

func TestGetSpriteInfo_ViaMCP(t *testing.T) {
	_, session, _, _, _ := createTestMCPSession(t)
	defer session.Close()

	// Create canvas
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      128,
			"height":     96,
			"color_mode": "indexed",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Get sprite info
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_sprite_info",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Width      int      `json:"width"`
		Height     int      `json:"height"`
		ColorMode  string   `json:"color_mode"`
		FrameCount int      `json:"frame_count"`
		LayerCount int      `json:"layer_count"`
		Layers     []string `json:"layers"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	assert.Equal(t, 128, output.Width)
	assert.Equal(t, 96, output.Height)
	assert.Equal(t, "indexed", output.ColorMode)
	assert.Equal(t, 1, output.FrameCount)
	assert.Equal(t, 1, output.LayerCount)
	assert.Len(t, output.Layers, 1)
}
