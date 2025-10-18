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

// createAntialiasingTestSession creates an MCP session with antialiasing tools registered
func createAntialiasingTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pixel-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterAntialiasingTools(server, client, gen, cfg, logger)
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

func TestSuggestAntialiasing_ViaMCP(t *testing.T) {
	_, session, _ := createAntialiasingTestSession(t)
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

	// Draw a jagged diagonal line to create edges that need antialiasing
	// Create a diagonal stair-step pattern
	_, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_pixels",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"pixels": []map[string]any{
				{"x": 10, "y": 10, "color": "#FF0000"},
				{"x": 11, "y": 10, "color": "#FF0000"},
				{"x": 11, "y": 11, "color": "#FF0000"},
				{"x": 12, "y": 11, "color": "#FF0000"},
				{"x": 12, "y": 12, "color": "#FF0000"},
				{"x": 13, "y": 12, "color": "#FF0000"},
			},
		},
	})
	require.NoError(t, err)

	// Suggest antialiasing without auto-apply
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "suggest_antialiasing",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"threshold":    128,
			"auto_apply":   false,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output AntialiasingResult
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	assert.False(t, output.Applied, "Should not be applied when auto_apply is false")
	assert.Greater(t, output.TotalEdges, 0, "Should detect jagged edges in diagonal pattern")
	assert.Equal(t, len(output.Suggestions), output.TotalEdges, "Suggestions count should match total edges")
}

func TestSuggestAntialiasingAutoApply_ViaMCP(t *testing.T) {
	_, session, _ := createAntialiasingTestSession(t)
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

	// Draw a jagged diagonal line
	_, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_pixels",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"pixels": []map[string]any{
				{"x": 15, "y": 15, "color": "#00FF00"},
				{"x": 16, "y": 15, "color": "#00FF00"},
				{"x": 16, "y": 16, "color": "#00FF00"},
				{"x": 17, "y": 16, "color": "#00FF00"},
			},
		},
	})
	require.NoError(t, err)

	// Suggest antialiasing with auto-apply
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "suggest_antialiasing",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"auto_apply":   true,
			"use_palette":  false,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output AntialiasingResult
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	assert.True(t, output.Applied, "Should be applied when auto_apply is true")
	assert.Greater(t, output.TotalEdges, 0, "Should detect and smooth jagged edges")
}

func TestSuggestAntialiasingWithRegion_ViaMCP(t *testing.T) {
	_, session, _ := createAntialiasingTestSession(t)
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

	// Draw jagged patterns in different regions
	_, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_pixels",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"pixels": []map[string]any{
				// First diagonal (will be outside region)
				{"x": 5, "y": 5, "color": "#FF0000"},
				{"x": 6, "y": 5, "color": "#FF0000"},
				{"x": 6, "y": 6, "color": "#FF0000"},
				// Second diagonal (will be inside region)
				{"x": 20, "y": 20, "color": "#0000FF"},
				{"x": 21, "y": 20, "color": "#0000FF"},
				{"x": 21, "y": 21, "color": "#0000FF"},
			},
		},
	})
	require.NoError(t, err)

	// Analyze only a specific region
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "suggest_antialiasing",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"region": map[string]any{
				"x":      15,
				"y":      15,
				"width":  20,
				"height": 20,
			},
			"auto_apply": false,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output AntialiasingResult
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	// Should only detect edges in the specified region (blue diagonal), not the red one
	assert.Greater(t, output.TotalEdges, 0, "Should detect edges in specified region")

	// Verify suggestions are within region bounds
	for _, sug := range output.Suggestions {
		assert.GreaterOrEqual(t, sug.X, 15, "Suggestion X should be within region")
		assert.GreaterOrEqual(t, sug.Y, 15, "Suggestion Y should be within region")
		assert.Less(t, sug.X, 35, "Suggestion X should be within region")
		assert.Less(t, sug.Y, 35, "Suggestion Y should be within region")
	}
}
