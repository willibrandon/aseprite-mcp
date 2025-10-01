package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/mtlog/core"
)

// SetPaletteInput defines the input parameters for the set_palette tool.
type SetPaletteInput struct {
	SpritePath string   `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	Colors     []string `json:"colors" jsonschema:"Array of hex colors to set as palette (#RRGGBB format)"`
}

// ApplyShadingInput defines the input parameters for the apply_shading tool.
type ApplyShadingInput struct {
	SpritePath     string   `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName      string   `json:"layer_name" jsonschema:"Name of the layer to apply shading to"`
	FrameNumber    int      `json:"frame_number" jsonschema:"Frame number (1-based index)"`
	Region         Region   `json:"region" jsonschema:"Rectangular region to apply shading to"`
	Palette        []string `json:"palette" jsonschema:"Array of hex colors ordered darkest to lightest (#RRGGBB format)"`
	LightDirection string   `json:"light_direction" jsonschema:"Light direction: top_left, top, top_right, left, right, bottom_left, bottom, bottom_right"`
	Intensity      float64  `json:"intensity" jsonschema:"Shading intensity (0.0-1.0)"`
	Style          string   `json:"style" jsonschema:"Shading style: pillow, smooth, or hard"`
}

// Region defines a rectangular region.
type Region struct {
	X      int `json:"x" jsonschema:"X coordinate of top-left corner"`
	Y      int `json:"y" jsonschema:"Y coordinate of top-left corner"`
	Width  int `json:"width" jsonschema:"Width of region"`
	Height int `json:"height" jsonschema:"Height of region"`
}

// AnalyzePaletteHarmoniesInput defines the input parameters for the analyze_palette_harmonies tool.
type AnalyzePaletteHarmoniesInput struct {
	Palette []string `json:"palette" jsonschema:"Array of hex colors to analyze (#RRGGBB format)"`
}

// PaletteHarmonyResult contains the result of palette harmony analysis.
type PaletteHarmonyResult struct {
	Complementary []ComplementaryPair `json:"complementary"`
	Triadic       []TriadicSet        `json:"triadic"`
	Analogous     []AnalogousSet      `json:"analogous"`
	Temperature   TemperatureAnalysis `json:"temperature"`
}

// ComplementaryPair represents two complementary colors.
type ComplementaryPair struct {
	Color1      string  `json:"color1"`
	Color2      string  `json:"color2"`
	Contrast    float64 `json:"contrast"`
	Description string  `json:"description"`
}

// TriadicSet represents three colors evenly spaced on the color wheel.
type TriadicSet struct {
	Colors      []string `json:"colors"`
	Balance     float64  `json:"balance"`
	Description string   `json:"description"`
}

// AnalogousSet represents adjacent colors on the color wheel.
type AnalogousSet struct {
	Colors      []string `json:"colors"`
	Harmony     float64  `json:"harmony"`
	Description string   `json:"description"`
}

// TemperatureAnalysis contains color temperature analysis.
type TemperatureAnalysis struct {
	WarmColors  []string `json:"warm_colors"`
	CoolColors  []string `json:"cool_colors"`
	NeutralColors []string `json:"neutral_colors"`
	Dominant    string   `json:"dominant"`
	Description string   `json:"description"`
}

// RegisterPaletteTools registers all palette management tools with the MCP server.
func RegisterPaletteTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register set_palette tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "set_palette",
			Description: "Set the sprite's color palette to the specified colors. Useful for applying extracted palettes from analyze_reference or creating custom limited palettes for pixel art. Colors should be in #RRGGBB hex format.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input SetPaletteInput) (*mcp.CallToolResult, *struct{ Success bool }, error) {
			logger.Debug("set_palette tool called",
				"sprite", input.SpritePath,
				"color_count", len(input.Colors))

			// Validate inputs
			if len(input.Colors) == 0 {
				return nil, nil, fmt.Errorf("colors array cannot be empty")
			}

			if len(input.Colors) > 256 {
				return nil, nil, fmt.Errorf("palette can have at most 256 colors, got %d", len(input.Colors))
			}

			// Validate hex color format
			for i, color := range input.Colors {
				if !isValidHexColor(color) {
					return nil, nil, fmt.Errorf("invalid color at index %d: %s (expected #RRGGBB format)", i, color)
				}
			}

			// Generate Lua script to set palette
			script := gen.SetPalette(input.Colors)

			// Execute Lua script
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to set palette", "error", err)
				return nil, nil, fmt.Errorf("failed to set palette: %w", err)
			}

			logger.Information("Palette set successfully",
				"sprite", input.SpritePath,
				"colors", len(input.Colors))

			return nil, &struct{ Success bool }{Success: true}, nil
		},
	)

	// Register apply_shading tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "apply_shading",
			Description: "Apply palette-constrained shading to a region based on light direction. Automatically adjusts pixel colors to create highlights and shadows while staying within the provided palette. Supports smooth, hard, and pillow shading styles. Essential for adding depth and dimension to pixel art.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input ApplyShadingInput) (*mcp.CallToolResult, *struct{ Success bool }, error) {
			logger.Debug("apply_shading tool called",
				"sprite", input.SpritePath,
				"layer", input.LayerName,
				"frame", input.FrameNumber,
				"palette_size", len(input.Palette))

			// Validate inputs
			if len(input.Palette) == 0 {
				return nil, nil, fmt.Errorf("palette array cannot be empty")
			}

			if len(input.Palette) > 256 {
				return nil, nil, fmt.Errorf("palette can have at most 256 colors, got %d", len(input.Palette))
			}

			// Validate hex color format for palette
			for i, color := range input.Palette {
				if !isValidHexColor(color) {
					return nil, nil, fmt.Errorf("invalid palette color at index %d: %s (expected #RRGGBB format)", i, color)
				}
			}

			// Validate light direction
			validDirections := map[string]bool{
				"top_left": true, "top": true, "top_right": true,
				"left": true, "right": true,
				"bottom_left": true, "bottom": true, "bottom_right": true,
			}
			if !validDirections[input.LightDirection] {
				return nil, nil, fmt.Errorf("invalid light direction: %s (must be one of: top_left, top, top_right, left, right, bottom_left, bottom, bottom_right)", input.LightDirection)
			}

			// Validate intensity
			if input.Intensity < 0.0 || input.Intensity > 1.0 {
				return nil, nil, fmt.Errorf("intensity must be between 0.0 and 1.0, got %f", input.Intensity)
			}

			// Validate style
			validStyles := map[string]bool{
				"pillow": true, "smooth": true, "hard": true,
			}
			if !validStyles[input.Style] {
				return nil, nil, fmt.Errorf("invalid style: %s (must be one of: pillow, smooth, hard)", input.Style)
			}

			// Validate region
			if input.Region.Width <= 0 || input.Region.Height <= 0 {
				return nil, nil, fmt.Errorf("region dimensions must be positive, got width=%d height=%d", input.Region.Width, input.Region.Height)
			}

			// Generate Lua script to apply shading
			script := gen.ApplyShading(
				input.LayerName,
				input.FrameNumber,
				input.Region.X,
				input.Region.Y,
				input.Region.Width,
				input.Region.Height,
				input.Palette,
				input.LightDirection,
				input.Intensity,
				input.Style,
			)

			// Execute Lua script
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to apply shading", "error", err)
				return nil, nil, fmt.Errorf("failed to apply shading: %w", err)
			}

			logger.Information("Shading applied successfully",
				"sprite", input.SpritePath,
				"layer", input.LayerName,
				"direction", input.LightDirection,
				"style", input.Style)

			return nil, &struct{ Success bool }{Success: true}, nil
		},
	)

	// Register analyze_palette_harmonies tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "analyze_palette_harmonies",
			Description: "Analyze color palette for harmonious relationships. Identifies complementary pairs (opposite colors on color wheel), triadic sets (3 evenly spaced colors), analogous groups (adjacent colors), and color temperature (warm/cool/neutral). Essential for creating professional, cohesive pixel art palettes based on color theory.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input AnalyzePaletteHarmoniesInput) (*mcp.CallToolResult, *PaletteHarmonyResult, error) {
			logger.Debug("analyze_palette_harmonies tool called",
				"palette_size", len(input.Palette))

			// Validate inputs
			if len(input.Palette) == 0 {
				return nil, nil, fmt.Errorf("palette array cannot be empty")
			}

			if len(input.Palette) > 256 {
				return nil, nil, fmt.Errorf("palette can have at most 256 colors, got %d", len(input.Palette))
			}

			// Validate hex color format
			for i, color := range input.Palette {
				if !isValidHexColor(color) {
					return nil, nil, fmt.Errorf("invalid color at index %d: %s (expected #RRGGBB format)", i, color)
				}
			}

			// Analyze harmonies
			result := analyzePaletteHarmonies(input.Palette)

			logger.Information("Palette harmonies analyzed successfully",
				"complementary_pairs", len(result.Complementary),
				"triadic_sets", len(result.Triadic),
				"analogous_sets", len(result.Analogous),
				"dominant_temp", result.Temperature.Dominant)

			output := &PaletteHarmonyResult{
				Complementary: result.Complementary,
				Triadic:       result.Triadic,
				Analogous:     result.Analogous,
				Temperature:   result.Temperature,
			}

			return nil, output, nil
		},
	)
}

// analyzePaletteHarmonies performs color harmony analysis on a palette.
func analyzePaletteHarmonies(palette []string) PaletteHarmonyResult {
	result := PaletteHarmonyResult{
		Complementary: []ComplementaryPair{},
		Triadic:       []TriadicSet{},
		Analogous:     []AnalogousSet{},
	}

	// Convert hex colors to HSL
	type ColorHSL struct {
		Hex        string
		H, S, L    float64
	}
	colors := make([]ColorHSL, 0, len(palette))

	for _, hexColor := range palette {
		h, s, l := hexToHSL(hexColor)
		colors = append(colors, ColorHSL{
			Hex: hexColor,
			H:   h,
			S:   s,
			L:   l,
		})
	}

	// Find complementary pairs (opposite on color wheel, ~180° apart)
	for i := 0; i < len(colors); i++ {
		for j := i + 1; j < len(colors); j++ {
			hueDiff := normalizeHueDiff(colors[i].H - colors[j].H)
			if hueDiff >= 150 && hueDiff <= 210 {
				// Complementary colors
				contrast := (colors[i].L + colors[j].L) / 2.0
				result.Complementary = append(result.Complementary, ComplementaryPair{
					Color1:      colors[i].Hex,
					Color2:      colors[j].Hex,
					Contrast:    contrast,
					Description: fmt.Sprintf("High contrast pair (%.0f° apart)", hueDiff),
				})
			}
		}
	}

	// Find triadic sets (3 colors evenly spaced, ~120° apart)
	for i := 0; i < len(colors); i++ {
		for j := i + 1; j < len(colors); j++ {
			for k := j + 1; k < len(colors); k++ {
				diff1 := normalizeHueDiff(colors[j].H - colors[i].H)
				diff2 := normalizeHueDiff(colors[k].H - colors[j].H)
				diff3 := normalizeHueDiff(colors[i].H - colors[k].H)

				// Check if roughly evenly spaced (±30° tolerance)
				if isNear(diff1, 120, 30) && isNear(diff2, 120, 30) && isNear(diff3, 120, 30) {
					avgDiff := (diff1 + diff2 + diff3) / 3.0
					balance := 1.0 - (absFloat(avgDiff-120) / 120.0)
					result.Triadic = append(result.Triadic, TriadicSet{
						Colors:      []string{colors[i].Hex, colors[j].Hex, colors[k].Hex},
						Balance:     balance,
						Description: fmt.Sprintf("Balanced triadic set (%.1f balance)", balance),
					})
				}
			}
		}
	}

	// Find analogous sets (adjacent colors, within 30-60°)
	for i := 0; i < len(colors); i++ {
		analogous := []string{colors[i].Hex}
		for j := 0; j < len(colors); j++ {
			if i != j {
				hueDiff := normalizeHueDiff(colors[j].H - colors[i].H)
				if hueDiff > 0 && hueDiff <= 60 {
					analogous = append(analogous, colors[j].Hex)
				}
			}
		}

		if len(analogous) >= 3 {
			// Calculate harmony score based on hue similarity
			harmony := 1.0 / float64(len(analogous))
			result.Analogous = append(result.Analogous, AnalogousSet{
				Colors:      analogous,
				Harmony:     harmony,
				Description: fmt.Sprintf("Harmonious adjacent colors (%d colors)", len(analogous)),
			})
		}
	}

	// Analyze color temperature
	warm := []string{}
	cool := []string{}
	neutral := []string{}

	for _, color := range colors {
		// Warm: red-orange-yellow (0-60° and 300-360°)
		// Cool: cyan-blue-purple (180-300°)
		// Neutral: green-yellow-green (60-180°) with low saturation
		if color.S < 0.2 {
			neutral = append(neutral, color.Hex)
		} else if (color.H >= 0 && color.H <= 60) || (color.H >= 300 && color.H <= 360) {
			warm = append(warm, color.Hex)
		} else if color.H >= 180 && color.H <= 300 {
			cool = append(cool, color.Hex)
		} else {
			neutral = append(neutral, color.Hex)
		}
	}

	dominant := "neutral"
	if len(warm) > len(cool) && len(warm) > len(neutral) {
		dominant = "warm"
	} else if len(cool) > len(warm) && len(cool) > len(neutral) {
		dominant = "cool"
	}

	result.Temperature = TemperatureAnalysis{
		WarmColors:    warm,
		CoolColors:    cool,
		NeutralColors: neutral,
		Dominant:      dominant,
		Description:   fmt.Sprintf("Palette is predominantly %s (%d warm, %d cool, %d neutral)", dominant, len(warm), len(cool), len(neutral)),
	}

	return result
}

// hexToHSL converts a hex color to HSL values.
func hexToHSL(hexColor string) (h, s, l float64) {
	// Remove # prefix
	hexColor = hexColor[1:]

	// Parse RGB
	var r, g, b int
	fmt.Sscanf(hexColor[:2], "%x", &r)
	fmt.Sscanf(hexColor[2:4], "%x", &g)
	fmt.Sscanf(hexColor[4:6], "%x", &b)

	// Normalize to 0-1
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	// Find max and min
	max := maxFloat(rf, gf, bf)
	min := minFloat(rf, gf, bf)
	delta := max - min

	// Calculate lightness
	l = (max + min) / 2.0

	// Calculate saturation
	if delta == 0 {
		h = 0
		s = 0
	} else {
		// Calculate saturation
		if l < 0.5 {
			s = delta / (max + min)
		} else {
			s = delta / (2.0 - max - min)
		}

		// Calculate hue
		switch max {
		case rf:
			h = ((gf - bf) / delta)
			if gf < bf {
				h += 6.0
			}
		case gf:
			h = ((bf - rf) / delta) + 2.0
		case bf:
			h = ((rf - gf) / delta) + 4.0
		}

		h *= 60.0
	}

	return h, s, l
}

// normalizeHueDiff normalizes hue difference to 0-360 range.
func normalizeHueDiff(diff float64) float64 {
	for diff < 0 {
		diff += 360
	}
	for diff > 360 {
		diff -= 360
	}
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

// isNear checks if a value is near a target within a tolerance.
func isNear(value, target, tolerance float64) bool {
	return absFloat(value-target) <= tolerance
}

// absFloat returns the absolute value of a float.
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// maxFloat returns the maximum of three floats.
func maxFloat(a, b, c float64) float64 {
	if a > b && a > c {
		return a
	}
	if b > c {
		return b
	}
	return c
}

// minFloat returns the minimum of three floats.
func minFloat(a, b, c float64) float64 {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
