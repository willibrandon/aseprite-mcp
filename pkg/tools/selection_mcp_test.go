package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
)

// createSelectionTestSession creates an MCP session with selection tools registered
func createSelectionTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "aseprite-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterSelectionTools(server, client, gen, cfg, logger)
	RegisterCanvasTools(server, client, gen, cfg, logger)
	RegisterDrawingTools(server, client, gen, cfg, logger)
	RegisterInspectionTools(server, client, gen, cfg, logger)

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

func TestSelectRectangle_ViaMCP(t *testing.T) {
	_, session, _ := createSelectionTestSession(t)
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

	// Select rectangle
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "select_rectangle",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"x":           5,
			"y":           5,
			"width":       10,
			"height":      10,
			"mode":        "replace",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Select rectangle should succeed")
}

func TestSelectEllipse_ViaMCP(t *testing.T) {
	_, session, _ := createSelectionTestSession(t)
	defer session.Close()

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

	// Select ellipse
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "select_ellipse",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"x":           8,
			"y":           8,
			"width":       16,
			"height":      16,
			"mode":        "replace",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Select ellipse should succeed")
}

func TestSelectAll_ViaMCP(t *testing.T) {
	_, session, _ := createSelectionTestSession(t)
	defer session.Close()

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

	// Select all
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "select_all",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Select all should succeed")
}

func TestDeselect_ViaMCP(t *testing.T) {
	_, session, _ := createSelectionTestSession(t)
	defer session.Close()

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

	// Deselect
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "deselect",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Deselect should succeed")
}

func TestMoveSelection_ViaMCP(t *testing.T) {
	_, session, _ := createSelectionTestSession(t)
	defer session.Close()

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

	// Create selection - persists to sprite.data
	selectResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "select_rectangle",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"x":           5,
			"y":           5,
			"width":       10,
			"height":      10,
			"mode":        "replace",
		},
	})
	require.NoError(t, err)
	require.False(t, selectResult.IsError)

	// Move selection - restored from sprite.data, works across processes!
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "move_selection",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"dx":          5,
			"dy":          5,
		},
	})

	require.NoError(t, err)
	if result.IsError {
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				t.Logf("Error from move_selection: %s", textContent.Text)
			}
		}
	}
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Move selection should succeed with persisted state")
}

func TestCopyPasteWorkflow_ViaMCP(t *testing.T) {
	_, session, _ := createSelectionTestSession(t)
	defer session.Close()

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

	// Draw content to copy
	session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_rectangle",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"x":            5,
			"y":            5,
			"width":        10,
			"height":       10,
			"color":        "#FF0000",
			"filled":       true,
		},
	})

	// Select rectangle
	session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "select_rectangle",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"x":           5,
			"y":           5,
			"width":       10,
			"height":      10,
			"mode":        "replace",
		},
	})

	// Copy - selection restored, copied to hidden clipboard layer
	copyResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "copy_selection",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
		},
	})

	require.NoError(t, err)
	if copyResult.IsError {
		if len(copyResult.Content) > 0 {
			if textContent, ok := copyResult.Content[0].(*mcp.TextContent); ok {
				t.Logf("Error from copy_selection: %s", textContent.Text)
			}
		}
	} else {
		// Log successful output to see debug messages
		if len(copyResult.Content) > 0 {
			if textContent, ok := copyResult.Content[0].(*mcp.TextContent); ok {
				t.Logf("Copy output: %s", textContent.Text)
			}
		}
	}
	require.False(t, copyResult.IsError)

	// Debug: Check if clipboard layer was created
	debugResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_sprite_info",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
		},
	})
	require.NoError(t, err)
	if len(debugResult.Content) > 0 {
		if textContent, ok := debugResult.Content[0].(*mcp.TextContent); ok {
			t.Logf("Sprite info after copy: %s", textContent.Text)
		}
	}

	// Debug: Check if Layer 1 has pixels
	pixelsResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_pixels",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"x":            5,
			"y":            5,
			"width":        10,
			"height":       10,
		},
	})
	require.NoError(t, err)
	if len(pixelsResult.Content) > 0 {
		if textContent, ok := pixelsResult.Content[0].(*mcp.TextContent); ok {
			t.Logf("Layer 1 pixels: %s", textContent.Text[:100]) // First 100 chars
		}
	}

	// Paste - retrieves from hidden clipboard layer, works across processes!
	pasteResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "paste_clipboard",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
		},
	})

	require.NoError(t, err)
	if pasteResult.IsError {
		if len(pasteResult.Content) > 0 {
			if textContent, ok := pasteResult.Content[0].(*mcp.TextContent); ok {
				t.Logf("Error from paste_clipboard: %s", textContent.Text)
			}
		}
	}
	require.False(t, pasteResult.IsError)
}

func TestCutPasteWorkflow_ViaMCP(t *testing.T) {
	_, session, _ := createSelectionTestSession(t)
	defer session.Close()

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

	// Draw content to cut
	session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_rectangle",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"x":            10,
			"y":            10,
			"width":        8,
			"height":       8,
			"color":        "#00FF00",
			"filled":       true,
		},
	})

	// Select rectangle
	session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "select_rectangle",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"x":           10,
			"y":           10,
			"width":       8,
			"height":      8,
			"mode":        "replace",
		},
	})

	// Cut - selection restored, copied to hidden clipboard, pixels cleared
	cutResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "cut_selection",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
		},
	})

	require.NoError(t, err)
	require.False(t, cutResult.IsError)

	// Paste - retrieves from hidden clipboard layer, works across processes!
	pasteResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "paste_clipboard",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
		},
	})

	require.NoError(t, err)
	require.False(t, pasteResult.IsError)
}
