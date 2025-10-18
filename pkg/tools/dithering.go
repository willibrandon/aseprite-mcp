package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
	"github.com/willibrandon/pixel-mcp/pkg/config"
)

// DrawWithDitherInput defines the input parameters for the draw_with_dither tool.
type DrawWithDitherInput struct {
	SpritePath  string      `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string      `json:"layer_name" jsonschema:"Name of the layer to draw on"`
	FrameNumber int         `json:"frame_number" jsonschema:"Frame number to draw on (1-based)"`
	Region      RegionInput `json:"region" jsonschema:"Rectangular region to fill with dithering"`
	Color1      string      `json:"color1" jsonschema:"First color (hex #RRGGBB or #RRGGBBAA)"`
	Color2      string      `json:"color2" jsonschema:"Second color (hex #RRGGBB or #RRGGBBAA)"`
	Pattern     string      `json:"pattern" jsonschema:"Dithering pattern: bayer_2x2|bayer_4x4|bayer_8x8|checkerboard|floyd_steinberg|grass|water|stone|cloud|brick|dots|diagonal|cross|noise|horizontal_lines|vertical_lines"`
	Density     float64     `json:"density,omitempty" jsonschema:"Ratio of color1 to color2 (0.0-1.0, default: 0.5)"`
}

// RegionInput defines a rectangular region.
type RegionInput struct {
	X      int `json:"x" jsonschema:"X coordinate of top-left corner"`
	Y      int `json:"y" jsonschema:"Y coordinate of top-left corner"`
	Width  int `json:"width" jsonschema:"Width of region"`
	Height int `json:"height" jsonschema:"Height of region"`
}

// RegisterDitheringTools registers all dithering tools with the MCP server.
func RegisterDitheringTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register draw_with_dither tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "draw_with_dither",
			Description: "Fill a region with a dithering pattern to create smooth gradients and textures. Supports 16 patterns: Bayer matrix (bayer_2x2, bayer_4x4, bayer_8x8) for ordered dithering, Floyd-Steinberg error diffusion (floyd_steinberg) for high-quality gradients, checkerboard for 50/50 blends, and texture patterns (grass, water, stone, cloud, brick, dots, diagonal, cross, noise, horizontal_lines, vertical_lines) for organic effects. Use density parameter to control the ratio of color1 to color2 (0.0 = all color1, 1.0 = all color2, 0.5 = even mix). Essential for professional pixel art gradients and textures.",
		},
		maybeWrapWithTiming("draw_with_dither", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input DrawWithDitherInput) (*mcp.CallToolResult, *struct{ Success bool }, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("draw_with_dither tool called",
				"sprite", input.SpritePath,
				"layer", input.LayerName,
				"frame", input.FrameNumber,
				"pattern", input.Pattern,
				"density", input.Density)

			// Validate inputs
			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be >= 1, got %d", input.FrameNumber)
			}

			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name is required")
			}

			if input.Region.Width <= 0 || input.Region.Height <= 0 {
				return nil, nil, fmt.Errorf("region width and height must be positive, got %dx%d",
					input.Region.Width, input.Region.Height)
			}

			// Validate pattern
			validPatterns := map[string]bool{
				"bayer_2x2":        true,
				"bayer_4x4":        true,
				"bayer_8x8":        true,
				"checkerboard":     true,
				"grass":            true,
				"water":            true,
				"stone":            true,
				"cloud":            true,
				"brick":            true,
				"dots":             true,
				"diagonal":         true,
				"cross":            true,
				"noise":            true,
				"horizontal_lines": true,
				"vertical_lines":   true,
				"floyd_steinberg":  true,
			}
			if !validPatterns[input.Pattern] {
				return nil, nil, fmt.Errorf("invalid pattern: %s (must be one of: bayer_2x2, bayer_4x4, bayer_8x8, checkerboard, grass, water, stone, cloud, brick, dots, diagonal, cross, noise, horizontal_lines, vertical_lines, floyd_steinberg)", input.Pattern)
			}

			// Validate colors (basic hex format check)
			if !isValidHexColor(input.Color1) {
				return nil, nil, fmt.Errorf("invalid color1 format: %s (expected #RRGGBB or #RRGGBBAA)", input.Color1)
			}
			if !isValidHexColor(input.Color2) {
				return nil, nil, fmt.Errorf("invalid color2 format: %s (expected #RRGGBB or #RRGGBBAA)", input.Color2)
			}

			// Set default density
			density := input.Density
			if density == 0 {
				density = 0.5
			}
			if density < 0 || density > 1 {
				return nil, nil, fmt.Errorf("density must be between 0.0 and 1.0, got %f", density)
			}

			// Generate Lua script for dithering
			script := gen.DrawWithDither(
				input.LayerName,
				input.FrameNumber,
				input.Region.X,
				input.Region.Y,
				input.Region.Width,
				input.Region.Height,
				input.Color1,
				input.Color2,
				input.Pattern,
				density,
			)

			// Execute Lua script
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to apply dithering", "error", err)
				return nil, nil, fmt.Errorf("failed to apply dithering: %w", err)
			}

			opLogger.Information("Dithering applied successfully",
				"sprite", input.SpritePath,
				"layer", input.LayerName,
				"pattern", input.Pattern,
				"region", fmt.Sprintf("%dx%d at (%d,%d)", input.Region.Width, input.Region.Height, input.Region.X, input.Region.Y))

			return nil, &struct{ Success bool }{Success: true}, nil
		}),
	)
}

// isValidHexColor checks if a string is a valid hex color format.
func isValidHexColor(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Remove # prefix if present
	if s[0] == '#' {
		s = s[1:]
	}

	// Check length (6 for RGB, 8 for RGBA)
	if len(s) != 6 && len(s) != 8 {
		return false
	}

	// Check all characters are hex digits
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}
