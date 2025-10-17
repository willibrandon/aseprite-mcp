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

// createDitheringTestSession creates an MCP session with dithering tools registered
func createDitheringTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "aseprite-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterDitheringTools(server, client, gen, cfg, logger)
	// Also register canvas tools for setup
	RegisterCanvasTools(server, client, gen, cfg, logger)

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

func TestDrawWithDither_ViaMCP(t *testing.T) {
	_, session, _ := createDitheringTestSession(t)
	defer session.Close()

	// Create a 64x64 sprite
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

	tests := []struct {
		name    string
		pattern string
		density float64
	}{
		{"Bayer 2x2", "bayer_2x2", 0.5},
		{"Bayer 4x4", "bayer_4x4", 0.5},
		{"Bayer 8x8", "bayer_8x8", 0.5},
		{"Checkerboard", "checkerboard", 0.5},
		{"Grass texture", "grass", 0.7},
		{"Water texture", "water", 0.6},
		{"Stone texture", "stone", 0.5},
		{"Cloud texture", "cloud", 0.4},
		{"Brick texture", "brick", 0.5},
		{"Dots texture", "dots", 0.5},
		{"Diagonal texture", "diagonal", 0.5},
		{"Cross texture", "cross", 0.5},
		{"Noise texture", "noise", 0.5},
		{"Horizontal lines", "horizontal_lines", 0.5},
		{"Vertical lines", "vertical_lines", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
				Name: "draw_with_dither",
				Arguments: map[string]any{
					"sprite_path":  createOutput.FilePath,
					"layer_name":   "Layer 1",
					"frame_number": 1,
					"region": map[string]any{
						"x":      0,
						"y":      0,
						"width":  16,
						"height": 16,
					},
					"color1":  "#FF0000FF",
					"color2":  "#0000FFFF",
					"pattern": tt.pattern,
					"density": tt.density,
				},
			})

			require.NoError(t, err)
			require.False(t, result.IsError)

			var output struct {
				Success bool `json:"success"`
			}
			json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
			assert.True(t, output.Success, "Dithering should succeed")
		})
	}
}

func TestDrawWithDitherDensity_ViaMCP(t *testing.T) {
	_, session, _ := createDitheringTestSession(t)
	defer session.Close()

	// Create a 32x32 sprite
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

	// Test different density values
	densities := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, density := range densities {
		result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "draw_with_dither",
			Arguments: map[string]any{
				"sprite_path":  createOutput.FilePath,
				"layer_name":   "Layer 1",
				"frame_number": 1,
				"region": map[string]any{
					"x":      0,
					"y":      0,
					"width":  8,
					"height": 8,
				},
				"color1":  "#FFFFFFFF",
				"color2":  "#000000FF",
				"pattern": "bayer_4x4",
				"density": density,
			},
		})

		require.NoError(t, err)
		require.False(t, result.IsError, "Density %f should succeed", density)
	}
}

// Unit tests for isValidHexColor helper function
func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{
			name:  "empty string",
			color: "",
			want:  false,
		},
		{
			name:  "valid RGB with hash",
			color: "#FF0000",
			want:  true,
		},
		{
			name:  "valid RGB without hash",
			color: "00FF00",
			want:  true,
		},
		{
			name:  "valid RGBA with hash",
			color: "#0000FFAA",
			want:  true,
		},
		{
			name:  "valid RGBA without hash",
			color: "FF00FF80",
			want:  true,
		},
		{
			name:  "too short",
			color: "#FFF",
			want:  false,
		},
		{
			name:  "too long",
			color: "#FF00FF00AA",
			want:  false,
		},
		{
			name:  "invalid character G",
			color: "#GGGGGG",
			want:  false,
		},
		{
			name:  "invalid character Z",
			color: "#FF00ZZ",
			want:  false,
		},
		{
			name:  "mixed case valid",
			color: "#Ff00Aa",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidHexColor(tt.color)
			assert.Equal(t, tt.want, got, "isValidHexColor(%q)", tt.color)
		})
	}
}
