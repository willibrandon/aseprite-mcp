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

func TestSetPaletteInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   SetPaletteInput
		wantErr bool
	}{
		{
			name: "valid palette with 8 colors",
			input: SetPaletteInput{
				SpritePath: "/tmp/test.aseprite",
				Colors:     []string{"#000000", "#FF0000", "#00FF00", "#0000FF", "#FFFF00", "#FF00FF", "#00FFFF", "#FFFFFF"},
			},
			wantErr: false,
		},
		{
			name: "valid palette with 16 colors",
			input: SetPaletteInput{
				SpritePath: "/tmp/test.aseprite",
				Colors:     []string{"#000000", "#111111", "#222222", "#333333", "#444444", "#555555", "#666666", "#777777", "#888888", "#999999", "#AAAAAA", "#BBBBBB", "#CCCCCC", "#DDDDDD", "#EEEEEE", "#FFFFFF"},
			},
			wantErr: false,
		},
		{
			name: "single color palette",
			input: SetPaletteInput{
				SpritePath: "/tmp/test.aseprite",
				Colors:     []string{"#FF5733"},
			},
			wantErr: false,
		},
		{
			name: "maximum 256 colors",
			input: SetPaletteInput{
				SpritePath: "/tmp/test.aseprite",
				Colors:     make([]string, 256),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate palette size
			if len(tt.input.Colors) == 0 {
				if !tt.wantErr {
					t.Error("Expected valid input but colors array is empty")
				}
			}
			if len(tt.input.Colors) > 256 {
				if !tt.wantErr {
					t.Error("Expected valid input but colors array exceeds 256")
				}
			}

			// Validate hex color format
			for _, color := range tt.input.Colors {
				if color != "" && !isValidHexColor(color) {
					if !tt.wantErr {
						t.Errorf("Expected valid input but color format is invalid: %s", color)
					}
				}
			}
		})
	}
}

func TestApplyShadingInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   ApplyShadingInput
		wantErr bool
	}{
		{
			name: "valid shading with smooth style",
			input: ApplyShadingInput{
				SpritePath:     "/tmp/test.aseprite",
				LayerName:      "Layer 1",
				FrameNumber:    1,
				Region:         Region{X: 0, Y: 0, Width: 32, Height: 32},
				Palette:        []string{"#000000", "#808080", "#FFFFFF"},
				LightDirection: "top_left",
				Intensity:      0.5,
				Style:          "smooth",
			},
			wantErr: false,
		},
		{
			name: "valid shading with hard style",
			input: ApplyShadingInput{
				SpritePath:     "/tmp/test.aseprite",
				LayerName:      "Layer 1",
				FrameNumber:    1,
				Region:         Region{X: 10, Y: 10, Width: 20, Height: 20},
				Palette:        []string{"#FF0000", "#FF8080", "#FFC0C0"},
				LightDirection: "top",
				Intensity:      0.8,
				Style:          "hard",
			},
			wantErr: false,
		},
		{
			name: "all light directions valid",
			input: ApplyShadingInput{
				SpritePath:     "/tmp/test.aseprite",
				LayerName:      "Layer 1",
				FrameNumber:    1,
				Region:         Region{X: 0, Y: 0, Width: 10, Height: 10},
				Palette:        []string{"#000000", "#FFFFFF"},
				LightDirection: "bottom_right",
				Intensity:      0.3,
				Style:          "smooth",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate region
			if tt.input.Region.Width <= 0 || tt.input.Region.Height <= 0 {
				if !tt.wantErr {
					t.Error("Expected valid input but region dimensions are invalid")
				}
			}

			// Validate intensity
			if tt.input.Intensity < 0.0 || tt.input.Intensity > 1.0 {
				if !tt.wantErr {
					t.Errorf("Expected valid input but intensity out of range: %f", tt.input.Intensity)
				}
			}

			// Validate light direction
			validDirections := map[string]bool{
				"top_left": true, "top": true, "top_right": true,
				"left": true, "right": true,
				"bottom_left": true, "bottom": true, "bottom_right": true,
			}
			if !validDirections[tt.input.LightDirection] {
				if !tt.wantErr {
					t.Errorf("Expected valid input but light direction is invalid: %s", tt.input.LightDirection)
				}
			}

			// Validate style
			validStyles := map[string]bool{"pillow": true, "smooth": true, "hard": true}
			if !validStyles[tt.input.Style] {
				if !tt.wantErr {
					t.Errorf("Expected valid input but style is invalid: %s", tt.input.Style)
				}
			}
		})
	}
}

func TestAnalyzePaletteHarmoniesInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   AnalyzePaletteHarmoniesInput
		wantErr bool
	}{
		{
			name: "valid palette with complementary colors",
			input: AnalyzePaletteHarmoniesInput{
				Palette: []string{"#FF0000", "#00FFFF"}, // Red and Cyan (complementary)
			},
			wantErr: false,
		},
		{
			name: "valid palette with triadic colors",
			input: AnalyzePaletteHarmoniesInput{
				Palette: []string{"#FF0000", "#00FF00", "#0000FF"}, // Red, Green, Blue (triadic)
			},
			wantErr: false,
		},
		{
			name: "large palette",
			input: AnalyzePaletteHarmoniesInput{
				Palette: []string{"#000000", "#111111", "#222222", "#333333", "#444444", "#555555", "#666666", "#777777", "#888888", "#999999", "#AAAAAA", "#BBBBBB", "#CCCCCC", "#DDDDDD", "#EEEEEE", "#FFFFFF"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate palette size
			if len(tt.input.Palette) == 0 {
				if !tt.wantErr {
					t.Error("Expected valid input but palette array is empty")
				}
			}
			if len(tt.input.Palette) > 256 {
				if !tt.wantErr {
					t.Error("Expected valid input but palette array exceeds 256")
				}
			}

			// Validate hex color format
			for _, color := range tt.input.Palette {
				if !isValidHexColor(color) {
					if !tt.wantErr {
						t.Errorf("Expected valid input but color format is invalid: %s", color)
					}
				}
			}
		})
	}
}

func TestHexToHSL(t *testing.T) {
	tests := []struct {
		name      string
		hexColor  string
		wantH     float64
		wantS     float64
		wantL     float64
		tolerance float64
	}{
		{
			name:      "pure red",
			hexColor:  "#FF0000",
			wantH:     0,
			wantS:     1.0,
			wantL:     0.5,
			tolerance: 1.0,
		},
		{
			name:      "pure green",
			hexColor:  "#00FF00",
			wantH:     120,
			wantS:     1.0,
			wantL:     0.5,
			tolerance: 1.0,
		},
		{
			name:      "pure blue",
			hexColor:  "#0000FF",
			wantH:     240,
			wantS:     1.0,
			wantL:     0.5,
			tolerance: 1.0,
		},
		{
			name:      "white",
			hexColor:  "#FFFFFF",
			wantH:     0,
			wantS:     0,
			wantL:     1.0,
			tolerance: 0.01,
		},
		{
			name:      "black",
			hexColor:  "#000000",
			wantH:     0,
			wantS:     0,
			wantL:     0,
			tolerance: 0.01,
		},
		{
			name:      "gray",
			hexColor:  "#808080",
			wantH:     0,
			wantS:     0,
			wantL:     0.5,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s, l := hexToHSL(tt.hexColor)

			// Check hue (with wraparound tolerance for 0/360)
			hueDiff := absFloat(h - tt.wantH)
			if hueDiff > 360-tt.tolerance {
				hueDiff = 360 - hueDiff
			}
			if hueDiff > tt.tolerance {
				t.Errorf("hexToHSL(%s) hue = %v, want %v (±%v)", tt.hexColor, h, tt.wantH, tt.tolerance)
			}

			// Check saturation
			if absFloat(s-tt.wantS) > tt.tolerance {
				t.Errorf("hexToHSL(%s) saturation = %v, want %v (±%v)", tt.hexColor, s, tt.wantS, tt.tolerance)
			}

			// Check lightness
			if absFloat(l-tt.wantL) > tt.tolerance {
				t.Errorf("hexToHSL(%s) lightness = %v, want %v (±%v)", tt.hexColor, l, tt.wantL, tt.tolerance)
			}
		})
	}
}

func TestAnalyzePaletteHarmonies_Complementary(t *testing.T) {
	// Test with known complementary pair
	palette := []string{"#FF0000", "#00FFFF"} // Red and Cyan (180° apart)
	result := analyzePaletteHarmonies(palette)

	if len(result.Complementary) == 0 {
		t.Error("Expected to find complementary pair, found none")
	}

	found := false
	for _, pair := range result.Complementary {
		if (pair.Color1 == "#FF0000" && pair.Color2 == "#00FFFF") ||
			(pair.Color1 == "#00FFFF" && pair.Color2 == "#FF0000") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find red-cyan complementary pair")
	}
}

func TestAnalyzePaletteHarmonies_Temperature(t *testing.T) {
	tests := []struct {
		name         string
		palette      []string
		wantDominant string
	}{
		{
			name:         "warm palette",
			palette:      []string{"#FF0000", "#FF8000", "#FFFF00"}, // Red, Orange, Yellow
			wantDominant: "warm",
		},
		{
			name:         "cool palette",
			palette:      []string{"#0000FF", "#0080FF", "#00FFFF"}, // Blue, Light Blue, Cyan
			wantDominant: "cool",
		},
		{
			name:         "neutral palette",
			palette:      []string{"#808080", "#A0A0A0", "#C0C0C0"}, // Grays
			wantDominant: "neutral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzePaletteHarmonies(tt.palette)

			if result.Temperature.Dominant != tt.wantDominant {
				t.Errorf("analyzePaletteHarmonies() dominant temperature = %v, want %v",
					result.Temperature.Dominant, tt.wantDominant)
			}
		})
	}
}

// createPaletteTestSession creates an MCP session with palette tools registered
func createPaletteTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pixel-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterPaletteTools(server, client, gen, cfg, logger)
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

func TestGetPalette_ViaMCP(t *testing.T) {
	_, session, _ := createPaletteTestSession(t)
	defer session.Close()

	// Create indexed sprite (has palette)
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      16,
			"height":     16,
			"color_mode": "indexed",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Get palette
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_palette",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Colors []string `json:"colors"`
		Size   int      `json:"size"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	assert.Greater(t, output.Size, 0, "Should have palette colors")
	assert.Equal(t, output.Size, len(output.Colors), "Size should match colors length")
}

func TestSetPalette_ViaMCP(t *testing.T) {
	_, session, _ := createPaletteTestSession(t)
	defer session.Close()

	// Create indexed sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      16,
			"height":     16,
			"color_mode": "indexed",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Set custom palette
	customPalette := []string{"#FF0000", "#00FF00", "#0000FF", "#FFFFFF", "#000000"}
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "set_palette",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"colors":      customPalette,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	// Verify palette was set
	getResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_palette",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
		},
	})
	require.NoError(t, err)

	var getOutput struct {
		Colors []string `json:"colors"`
	}
	json.Unmarshal([]byte(getResult.Content[0].(*mcp.TextContent).Text), &getOutput)

	assert.Equal(t, 5, len(getOutput.Colors), "Should have 5 palette colors")
}

func TestSetPaletteColor_ViaMCP(t *testing.T) {
	_, session, _ := createPaletteTestSession(t)
	defer session.Close()

	// Create indexed sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      16,
			"height":     16,
			"color_mode": "indexed",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Set color at index 0 to red
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "set_palette_color",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"index":       0,
			"color":       "#FF0000",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)
}

func TestAddPaletteColor_ViaMCP(t *testing.T) {
	_, session, _ := createPaletteTestSession(t)
	defer session.Close()

	// Create indexed sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      16,
			"height":     16,
			"color_mode": "indexed",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Set a smaller palette first (indexed sprites start with 256 colors)
	session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "set_palette",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"colors":      []string{"#FF0000", "#00FF00", "#0000FF"},
		},
	})

	// Now add a new color to palette
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "add_palette_color",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"color":       "#FF00FF",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		ColorIndex int `json:"color_index"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	assert.GreaterOrEqual(t, output.ColorIndex, 0, "Should return valid palette index")
}

func TestSortPalette_ViaMCP(t *testing.T) {
	_, session, _ := createPaletteTestSession(t)
	defer session.Close()

	// Create indexed sprite
	createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_canvas",
		Arguments: map[string]any{
			"width":      16,
			"height":     16,
			"color_mode": "indexed",
		},
	})
	require.NoError(t, err)

	var createOutput struct {
		FilePath string `json:"file_path"`
	}
	json.Unmarshal([]byte(createResult.Content[0].(*mcp.TextContent).Text), &createOutput)
	defer os.Remove(createOutput.FilePath)

	// Set a known palette
	session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "set_palette",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"colors":      []string{"#0000FF", "#FF0000", "#00FF00", "#FFFFFF", "#000000"},
		},
	})

	// Sort by hue
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "sort_palette",
		Arguments: map[string]any{
			"sprite_path": createOutput.FilePath,
			"method":      "hue",
			"ascending":   true,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)
}

func TestAnalyzePaletteHarmonies_ViaMCP(t *testing.T) {
	_, session, _ := createPaletteTestSession(t)
	defer session.Close()

	// Test with known complementary colors
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "analyze_palette_harmonies",
		Arguments: map[string]any{
			"palette": []string{"#FF0000", "#00FFFF", "#00FF00"},
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Complementary []struct {
			Color1 string `json:"color1"`
			Color2 string `json:"color2"`
		} `json:"complementary"`
		Triadic []struct {
			Color1 string `json:"color1"`
			Color2 string `json:"color2"`
			Color3 string `json:"color3"`
		} `json:"triadic"`
		Analogous []struct {
			Colors []string `json:"colors"`
		} `json:"analogous"`
		Temperature struct {
			Warm    []string `json:"warm"`
			Cool    []string `json:"cool"`
			Neutral []string `json:"neutral"`
		} `json:"temperature"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	assert.NotNil(t, output.Complementary, "Should analyze complementary colors")
	assert.NotNil(t, output.Temperature, "Should analyze color temperature")
}

func TestApplyShading_ViaMCP(t *testing.T) {
	_, session, _ := createPaletteTestSession(t)
	defer session.Close()

	// Create RGB sprite for shading test
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

	// Draw a filled rectangle to apply shading to
	_, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_rectangle",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"x":            16,
			"y":            16,
			"width":        32,
			"height":       32,
			"color":        "#808080",
			"filled":       true,
		},
	})
	require.NoError(t, err)

	// Apply smooth shading with top-left light
	palette := []string{"#000000", "#404040", "#808080", "#C0C0C0", "#FFFFFF"}
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "apply_shading",
		Arguments: map[string]any{
			"sprite_path":  createOutput.FilePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"region": map[string]any{
				"x":      16,
				"y":      16,
				"width":  32,
				"height": 32,
			},
			"palette":         palette,
			"light_direction": "top_left",
			"intensity":       0.8,
			"style":           "smooth",
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success, "Apply shading should succeed")
}

// Unit tests for parseJSON helper function
func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			output:  `{"success": true}`,
			wantErr: false,
		},
		{
			name:    "valid JSON with extra text before",
			output:  `some text before {"success": true}`,
			wantErr: false,
		},
		{
			name:    "valid JSON with extra text after",
			output:  `{"success": true} some text after`,
			wantErr: false,
		},
		{
			name:    "valid JSON with extra text before and after",
			output:  `prefix {"success": true} suffix`,
			wantErr: false,
		},
		{
			name:    "no JSON object",
			output:  `no json here`,
			wantErr: true,
		},
		{
			name:    "empty string",
			output:  ``,
			wantErr: true,
		},
		{
			name:    "only opening brace",
			output:  `{`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			output:  `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				Success bool `json:"success"`
			}
			err := parseJSON(tt.output, &result)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, result.Success)
			}
		})
	}
}

// Unit tests for normalizeHueDiff helper function
func TestNormalizeHueDiff(t *testing.T) {
	tests := []struct {
		name string
		diff float64
		want float64
	}{
		{
			name: "positive difference < 180",
			diff: 100.0,
			want: 100.0,
		},
		{
			name: "negative difference",
			diff: -50.0,
			want: 50.0, // -50 + 360 = 310, which is > 180, so 360 - 310 = 50
		},
		{
			name: "difference > 180",
			diff: 200.0,
			want: 160.0, // 360 - 200 = 160
		},
		{
			name: "difference > 360",
			diff: 400.0,
			want: 40.0, // 400 - 360 = 40
		},
		{
			name: "exact 180",
			diff: 180.0,
			want: 180.0,
		},
		{
			name: "multiple loops negative",
			diff: -720.0,
			want: 0.0, // -720 + 360 + 360 = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeHueDiff(tt.diff)
			assert.Equal(t, tt.want, got)
		})
	}
}
