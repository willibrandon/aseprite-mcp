package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
)

// createTransformTestSession creates an MCP session with transform tools registered
func createTransformTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "aseprite-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterTransformTools(server, client, gen, cfg, logger)
	// Also register canvas and drawing tools for setup
	RegisterCanvasTools(server, client, gen, cfg, logger)
	RegisterDrawingTools(server, client, gen, cfg, logger)

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

func TestFlipSprite_ViaMCP(t *testing.T) {
	_, session, _ := createTransformTestSession(t)
	defer session.Close()

	// Create a sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      32,
			"height":     32,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Flip horizontally
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "flip_sprite",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"direction":   "horizontal",
			"target":      "sprite",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Flip should succeed")
}

func TestRotateSprite_ViaMCP(t *testing.T) {
	_, session, _ := createTransformTestSession(t)
	defer session.Close()

	// Create a sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      32,
			"height":     32,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Rotate 90 degrees
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "rotate_sprite",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"angle":       90,
			"target":      "sprite",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Rotate should succeed")
}

func TestScaleSprite_ViaMCP(t *testing.T) {
	_, session, _ := createTransformTestSession(t)
	defer session.Close()

	// Create a sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      32,
			"height":     32,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Scale to 2x
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "scale_sprite",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"scale_x":     2.0,
			"scale_y":     2.0,
			"algorithm":   "nearest",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		NewWidth  int `json:"new_width"`
		NewHeight int `json:"new_height"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	assert.Equal(t, 64, output.NewWidth, "Width should be doubled")
	assert.Equal(t, 64, output.NewHeight, "Height should be doubled")
}

func TestCropSprite_ViaMCP(t *testing.T) {
	_, session, _ := createTransformTestSession(t)
	defer session.Close()

	// Create a sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      64,
			"height":     64,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Crop to 32x32 from top-left
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "crop_sprite",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"x":           0,
			"y":           0,
			"width":       32,
			"height":      32,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Crop should succeed")
}

func TestResizeCanvas_ViaMCP(t *testing.T) {
	_, session, _ := createTransformTestSession(t)
	defer session.Close()

	// Create a sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      32,
			"height":     32,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Resize canvas to 64x64 with center anchor
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "resize_canvas",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"width":       64,
			"height":      64,
			"anchor":      "center",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Resize canvas should succeed")
}

func TestDownsampleImage_ViaMCP(t *testing.T) {
	_, session, _ := createTransformTestSession(t)
	defer session.Close()

	// Create a source sprite to downsample
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      64,
			"height":     64,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Downsample to 32x32
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "downsample_image",
		Arguments: map[string]any{
			"source_path":   createOutput.FilePath,
			"target_width":  16,
			"target_height": 16,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		OutputPath string `json:"output_path"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	// Verify downsampled file exists
	_, err = os.Stat(output.OutputPath)
	assert.NoError(t, err, "Downsampled sprite should exist")
	defer os.Remove(output.OutputPath)
}

func TestApplyOutline_ViaMCP(t *testing.T) {
	_, session, _ := createTransformTestSession(t)
	defer session.Close()

	// Create a test sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      32,
			"height":     32,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Draw some pixels first so outline has something to work with
	_, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_rectangle",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"x":            8,
			"y":            8,
			"width":        16,
			"height":       16,
			"color":        "#FF0000FF",
			"filled":       true,
		},
	})
	require.NoError(t, err)

	// Apply outline
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "apply_outline",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"color":        "#000000FF",
			"thickness":    2,
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
