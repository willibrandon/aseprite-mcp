package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/mtlog/core"
)

// PixelInput represents a single pixel to be drawn.
type PixelInput struct {
	X     int    `json:"x" jsonschema:"X coordinate of the pixel"`
	Y     int    `json:"y" jsonschema:"Y coordinate of the pixel"`
	Color string `json:"color" jsonschema:"Hex color string in format #RRGGBB or #RRGGBBAA"`
}

// DrawPixelsInput defines the input parameters for the draw_pixels tool.
type DrawPixelsInput struct {
	SpritePath  string        `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string        `json:"layer_name" jsonschema:"Name of the layer to draw on"`
	FrameNumber int           `json:"frame_number" jsonschema:"Frame number to draw on (1-based)"`
	Pixels      []PixelInput  `json:"pixels" jsonschema:"Array of pixels to draw"`
}

// DrawPixelsOutput defines the output for the draw_pixels tool.
type DrawPixelsOutput struct {
	PixelsDrawn int `json:"pixels_drawn" jsonschema:"Number of pixels successfully drawn"`
}

// RegisterDrawingTools registers all drawing tools with the MCP server.
func RegisterDrawingTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register draw_pixels tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "draw_pixels",
			Description: "Draw individual pixels at specified coordinates with colors. Supports batch operations for efficiency.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input DrawPixelsInput) (*mcp.CallToolResult, *DrawPixelsOutput, error) {
			logger.Debug("draw_pixels tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName, "frame_number", input.FrameNumber, "pixel_count", len(input.Pixels))

			// Validate inputs
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be at least 1, got %d", input.FrameNumber)
			}

			if len(input.Pixels) == 0 {
				return nil, nil, fmt.Errorf("pixels array cannot be empty")
			}

			// Convert pixel inputs to aseprite.Pixel types
			pixels := make([]aseprite.Pixel, len(input.Pixels))
			for i, p := range input.Pixels {
				var color aseprite.Color
				if err := color.FromHex(p.Color); err != nil {
					return nil, nil, fmt.Errorf("invalid color format for pixel %d: %w", i, err)
				}

				pixels[i] = aseprite.Pixel{
					Point: aseprite.Point{X: p.X, Y: p.Y},
					Color: color,
				}
			}

			// Generate Lua script
			script := gen.DrawPixels(input.LayerName, input.FrameNumber, pixels)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to draw pixels", "error", err)
				return nil, nil, fmt.Errorf("failed to draw pixels: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Pixels drawn successfully") {
				logger.Warning("Unexpected output from draw_pixels", "output", output)
			}

			logger.Information("Pixels drawn successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "count", len(pixels))

			return nil, &DrawPixelsOutput{PixelsDrawn: len(pixels)}, nil
		},
	)
}