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
	SpritePath  string       `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string       `json:"layer_name" jsonschema:"Name of the layer to draw on"`
	FrameNumber int          `json:"frame_number" jsonschema:"Frame number to draw on (1-based)"`
	Pixels      []PixelInput `json:"pixels" jsonschema:"Array of pixels to draw"`
}

// DrawPixelsOutput defines the output for the draw_pixels tool.
type DrawPixelsOutput struct {
	PixelsDrawn int `json:"pixels_drawn" jsonschema:"Number of pixels successfully drawn"`
}

// DrawLineInput defines the input parameters for the draw_line tool.
type DrawLineInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to draw on"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to draw on (1-based)"`
	X1          int    `json:"x1" jsonschema:"X coordinate of line start point"`
	Y1          int    `json:"y1" jsonschema:"Y coordinate of line start point"`
	X2          int    `json:"x2" jsonschema:"X coordinate of line end point"`
	Y2          int    `json:"y2" jsonschema:"Y coordinate of line end point"`
	Color       string `json:"color" jsonschema:"Hex color string in format #RRGGBB or #RRGGBBAA"`
	Thickness   int    `json:"thickness" jsonschema:"Line thickness in pixels (1-100)"`
}

// DrawLineOutput defines the output for the draw_line tool.
type DrawLineOutput struct {
	Success bool `json:"success" jsonschema:"Whether the line was drawn successfully"`
}

// DrawRectangleInput defines the input parameters for the draw_rectangle tool.
type DrawRectangleInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to draw on"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to draw on (1-based)"`
	X           int    `json:"x" jsonschema:"X coordinate of rectangle top-left corner"`
	Y           int    `json:"y" jsonschema:"Y coordinate of rectangle top-left corner"`
	Width       int    `json:"width" jsonschema:"Width of rectangle in pixels"`
	Height      int    `json:"height" jsonschema:"Height of rectangle in pixels"`
	Color       string `json:"color" jsonschema:"Hex color string in format #RRGGBB or #RRGGBBAA"`
	Filled      bool   `json:"filled" jsonschema:"Fill interior (true) or draw outline only (false)"`
}

// DrawRectangleOutput defines the output for the draw_rectangle tool.
type DrawRectangleOutput struct {
	Success bool `json:"success" jsonschema:"Whether the rectangle was drawn successfully"`
}

// DrawCircleInput defines the input parameters for the draw_circle tool.
type DrawCircleInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to draw on"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to draw on (1-based)"`
	CenterX     int    `json:"center_x" jsonschema:"X coordinate of circle center"`
	CenterY     int    `json:"center_y" jsonschema:"Y coordinate of circle center"`
	Radius      int    `json:"radius" jsonschema:"Radius of circle in pixels"`
	Color       string `json:"color" jsonschema:"Hex color string in format #RRGGBB or #RRGGBBAA"`
	Filled      bool   `json:"filled" jsonschema:"Fill interior (true) or draw outline only (false)"`
}

// DrawCircleOutput defines the output for the draw_circle tool.
type DrawCircleOutput struct {
	Success bool `json:"success" jsonschema:"Whether the circle was drawn successfully"`
}

// FillAreaInput defines the input parameters for the fill_area tool.
type FillAreaInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to draw on"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to draw on (1-based)"`
	X           int    `json:"x" jsonschema:"X coordinate of starting point"`
	Y           int    `json:"y" jsonschema:"Y coordinate of starting point"`
	Color       string `json:"color" jsonschema:"Hex color string in format #RRGGBB or #RRGGBBAA"`
	Tolerance   int    `json:"tolerance" jsonschema:"Color matching tolerance (0-255, default 0)"`
}

// FillAreaOutput defines the output for the fill_area tool.
type FillAreaOutput struct {
	Success bool `json:"success" jsonschema:"Whether the fill operation was successful"`
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

	// Register draw_line tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "draw_line",
			Description: "Draw a line between two points with specified color and thickness.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input DrawLineInput) (*mcp.CallToolResult, *DrawLineOutput, error) {
			logger.Debug("draw_line tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName, "frame_number", input.FrameNumber)

			// Validate inputs
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be at least 1, got %d", input.FrameNumber)
			}

			if input.Thickness < 1 || input.Thickness > 100 {
				return nil, nil, fmt.Errorf("thickness must be between 1 and 100, got %d", input.Thickness)
			}

			// Parse color
			var color aseprite.Color
			if err := color.FromHex(input.Color); err != nil {
				return nil, nil, fmt.Errorf("invalid color format: %w", err)
			}

			// Generate Lua script
			script := gen.DrawLine(input.LayerName, input.FrameNumber, input.X1, input.Y1, input.X2, input.Y2, color, input.Thickness)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to draw line", "error", err)
				return nil, nil, fmt.Errorf("failed to draw line: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Line drawn successfully") {
				logger.Warning("Unexpected output from draw_line", "output", output)
			}

			logger.Information("Line drawn successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber)

			return nil, &DrawLineOutput{Success: true}, nil
		},
	)

	// Register draw_rectangle tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "draw_rectangle",
			Description: "Draw a rectangle with specified position, size, color, and fill option.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input DrawRectangleInput) (*mcp.CallToolResult, *DrawRectangleOutput, error) {
			logger.Debug("draw_rectangle tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName, "frame_number", input.FrameNumber, "filled", input.Filled)

			// Validate inputs
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be at least 1, got %d", input.FrameNumber)
			}

			if input.Width < 1 || input.Height < 1 {
				return nil, nil, fmt.Errorf("width and height must be at least 1, got width=%d height=%d", input.Width, input.Height)
			}

			// Parse color
			var color aseprite.Color
			if err := color.FromHex(input.Color); err != nil {
				return nil, nil, fmt.Errorf("invalid color format: %w", err)
			}

			// Generate Lua script
			script := gen.DrawRectangle(input.LayerName, input.FrameNumber, input.X, input.Y, input.Width, input.Height, color, input.Filled)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to draw rectangle", "error", err)
				return nil, nil, fmt.Errorf("failed to draw rectangle: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Rectangle drawn successfully") {
				logger.Warning("Unexpected output from draw_rectangle", "output", output)
			}

			logger.Information("Rectangle drawn successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "filled", input.Filled)

			return nil, &DrawRectangleOutput{Success: true}, nil
		},
	)

	// Register draw_circle tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "draw_circle",
			Description: "Draw a circle with specified center, radius, color, and fill option.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input DrawCircleInput) (*mcp.CallToolResult, *DrawCircleOutput, error) {
			logger.Debug("draw_circle tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName, "frame_number", input.FrameNumber, "radius", input.Radius, "filled", input.Filled)

			// Validate inputs
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be at least 1, got %d", input.FrameNumber)
			}

			if input.Radius < 1 {
				return nil, nil, fmt.Errorf("radius must be at least 1, got %d", input.Radius)
			}

			// Parse color
			var color aseprite.Color
			if err := color.FromHex(input.Color); err != nil {
				return nil, nil, fmt.Errorf("invalid color format: %w", err)
			}

			// Generate Lua script
			script := gen.DrawCircle(input.LayerName, input.FrameNumber, input.CenterX, input.CenterY, input.Radius, color, input.Filled)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to draw circle", "error", err)
				return nil, nil, fmt.Errorf("failed to draw circle: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Circle drawn successfully") {
				logger.Warning("Unexpected output from draw_circle", "output", output)
			}

			logger.Information("Circle drawn successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "radius", input.Radius, "filled", input.Filled)

			return nil, &DrawCircleOutput{Success: true}, nil
		},
	)

	// Register fill_area tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "fill_area",
			Description: "Flood fill from a starting point with specified color (paint bucket tool).",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input FillAreaInput) (*mcp.CallToolResult, *FillAreaOutput, error) {
			logger.Debug("fill_area tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName, "frame_number", input.FrameNumber, "x", input.X, "y", input.Y, "tolerance", input.Tolerance)

			// Validate inputs
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be at least 1, got %d", input.FrameNumber)
			}

			if input.Tolerance < 0 || input.Tolerance > 255 {
				return nil, nil, fmt.Errorf("tolerance must be between 0 and 255, got %d", input.Tolerance)
			}

			// Parse color
			var color aseprite.Color
			if err := color.FromHex(input.Color); err != nil {
				return nil, nil, fmt.Errorf("invalid color format: %w", err)
			}

			// Generate Lua script
			script := gen.FillArea(input.LayerName, input.FrameNumber, input.X, input.Y, color, input.Tolerance)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to fill area", "error", err)
				return nil, nil, fmt.Errorf("failed to fill area: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Area filled successfully") {
				logger.Warning("Unexpected output from fill_area", "output", output)
			}

			logger.Information("Area filled successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "x", input.X, "y", input.Y, "tolerance", input.Tolerance)

			return nil, &FillAreaOutput{Success: true}, nil
		},
	)
}
