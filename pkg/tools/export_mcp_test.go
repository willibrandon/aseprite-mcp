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
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// createExportTestSession creates an MCP session with export tools registered
func createExportTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client, string) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pixel-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterExportTools(server, client, gen, cfg, logger)
	// Also register canvas tools for setup
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

	return server, session, client, cfg.TempDir
}

func TestExportSprite_ViaMCP(t *testing.T) {
	_, session, _, tempDir := createExportTestSession(t)
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

	// Export to PNG
	exportPath := filepath.Join(tempDir, "test_export.png")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "export_sprite",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"output_path":  exportPath,
			"format":       "png",
			"frame_number": 1,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)
	defer os.Remove(exportPath)

	// Verify file was created
	_, err = os.Stat(exportPath)
	assert.NoError(t, err, "Exported PNG should exist")
}

func TestExportSpritesheet_ViaMCP(t *testing.T) {
	_, session, _, tempDir := createExportTestSession(t)
	defer session.Close()

	// Create a sprite with 2 frames
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      16,
			"height":     16,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Add a second frame
	session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "add_frame",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"duration_ms": 100,
		},
	})

	// Export spritesheet
	exportPath := filepath.Join(tempDir, "test_spritesheet.png")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "export_spritesheet",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"output_path":  exportPath,
			"layout":       "horizontal",
			"padding":      0,
			"include_json": false,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)
	defer os.Remove(exportPath)

	// Verify file was created
	_, err = os.Stat(exportPath)
	assert.NoError(t, err, "Exported spritesheet should exist")
}

func TestImportImage_ViaMCP(t *testing.T) {
	_, session, _, tempDir := createExportTestSession(t)
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

	// First export a layer as image to use for import
	exportPath := filepath.Join(tempDir, "layer_export.png")
	session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "export_sprite",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"output_path":  exportPath,
			"format":       "png",
			"frame_number": 1,
		},
	})
	defer os.Remove(exportPath)

	// Now import it back as a new layer
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "import_image",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"image_path":   exportPath,
			"layer_name":   "ImportedLayer",
			"frame_number": 1,
			"position": map[string]any{
				"x": 0,
				"y": 0,
			},
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Import should succeed")
}

func TestSaveAs_ViaMCP(t *testing.T) {
	_, session, _, tempDir := createExportTestSession(t)
	defer session.Close()

	// Create a sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      16,
			"height":     16,
			"color_mode": "rgb",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Save as new file
	newPath := filepath.Join(tempDir, "saved_copy.aseprite")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "save_as",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"output_path": newPath,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)
	defer os.Remove(newPath)

	var output struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	// Verify new file exists
	_, err = os.Stat(output.FilePath)
	assert.NoError(t, err, "Saved file should exist")
	assert.Equal(t, newPath, output.FilePath, "Output path should match")
}
