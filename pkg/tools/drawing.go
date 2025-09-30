package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

type DrawPixelsInput struct {
	SpritePath  string        `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
	LayerName   string        `json:"layer_name" jsonschema:"required,description=Layer to draw on"`
	FrameNumber int           `json:"frame_number" jsonschema:"required,minimum=1,description=Frame number (1-indexed)"`
	Pixels      []PixelInput  `json:"pixels" jsonschema:"required,description=Array of pixels to draw"`
}

type PixelInput struct {
	X     int    `json:"x" jsonschema:"required,description=X coordinate"`
	Y     int    `json:"y" jsonschema:"required,description=Y coordinate"`
	Color string `json:"color" jsonschema:"required,description=Hex color (#RRGGBB or #RRGGBBAA)"`
}

type DrawPixelsOutput struct {
	PixelsDrawn int `json:"pixels_drawn" jsonschema:"description=Number of pixels drawn"`
}

type DrawLineInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
	LayerName   string `json:"layer_name" jsonschema:"required,description=Layer to draw on"`
	FrameNumber int    `json:"frame_number" jsonschema:"required,minimum=1,description=Frame number (1-indexed)"`
	X1          int    `json:"x1" jsonschema:"required,description=Start X coordinate"`
	Y1          int    `json:"y1" jsonschema:"required,description=Start Y coordinate"`
	X2          int    `json:"x2" jsonschema:"required,description=End X coordinate"`
	Y2          int    `json:"y2" jsonschema:"required,description=End Y coordinate"`
	Color       string `json:"color" jsonschema:"required,description=Hex color (#RRGGBB or #RRGGBBAA)"`
	Thickness   int    `json:"thickness" jsonschema:"minimum=1,maximum=100,description=Line thickness in pixels"`
}

type DrawRectangleInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
	LayerName   string `json:"layer_name" jsonschema:"required,description=Layer to draw on"`
	FrameNumber int    `json:"frame_number" jsonschema:"required,minimum=1,description=Frame number (1-indexed)"`
	X           int    `json:"x" jsonschema:"required,description=X coordinate"`
	Y           int    `json:"y" jsonschema:"required,description=Y coordinate"`
	Width       int    `json:"width" jsonschema:"required,minimum=1,description=Rectangle width"`
	Height      int    `json:"height" jsonschema:"required,minimum=1,description=Rectangle height"`
	Color       string `json:"color" jsonschema:"required,description=Hex color (#RRGGBB or #RRGGBBAA)"`
	Filled      bool   `json:"filled" jsonschema:"description=Fill interior or outline only"`
}

type DrawCircleInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
	LayerName   string `json:"layer_name" jsonschema:"required,description=Layer to draw on"`
	FrameNumber int    `json:"frame_number" jsonschema:"required,minimum=1,description=Frame number (1-indexed)"`
	CenterX     int    `json:"center_x" jsonschema:"required,description=Center X coordinate"`
	CenterY     int    `json:"center_y" jsonschema:"required,description=Center Y coordinate"`
	Radius      int    `json:"radius" jsonschema:"required,minimum=1,description=Circle radius"`
	Color       string `json:"color" jsonschema:"required,description=Hex color (#RRGGBB or #RRGGBBAA)"`
	Filled      bool   `json:"filled" jsonschema:"description=Fill interior or outline only"`
}

type FillAreaInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
	LayerName   string `json:"layer_name" jsonschema:"required,description=Layer to draw on"`
	FrameNumber int    `json:"frame_number" jsonschema:"required,minimum=1,description=Frame number (1-indexed)"`
	X           int    `json:"x" jsonschema:"required,description=Starting X coordinate"`
	Y           int    `json:"y" jsonschema:"required,description=Starting Y coordinate"`
	Color       string `json:"color" jsonschema:"required,description=Hex color (#RRGGBB or #RRGGBBAA)"`
}

type DrawingOutput struct {
	Success bool `json:"success" jsonschema:"description=Whether the operation succeeded"`
}

func RegisterDrawingTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator) {
	// draw_pixels tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "draw_pixels",
		Description: "Draw individual pixels at specified coordinates",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DrawPixelsInput) (*mcp.CallToolResult, *DrawPixelsOutput, error) {
		// Convert input pixels to aseprite.Pixel
		pixels := make([]aseprite.Pixel, len(input.Pixels))
		for i, p := range input.Pixels {
			var color aseprite.Color
			if err := color.FromHex(p.Color); err != nil {
				return nil, nil, fmt.Errorf("invalid color at pixel %d: %w", i, err)
			}
			pixels[i] = aseprite.Pixel{
				Point: aseprite.Point{X: p.X, Y: p.Y},
				Color: color,
			}
		}

		// Generate Lua script
		script := gen.DrawPixels(input.LayerName, input.FrameNumber, pixels)

		// Execute
		_, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to draw pixels: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Drew %d pixels successfully", len(pixels))},
			},
		}, &DrawPixelsOutput{PixelsDrawn: len(pixels)}, nil
	})

	// draw_line tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "draw_line",
		Description: "Draw a line between two points",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DrawLineInput) (*mcp.CallToolResult, *DrawingOutput, error) {
		var color aseprite.Color
		if err := color.FromHex(input.Color); err != nil {
			return nil, nil, fmt.Errorf("invalid color: %w", err)
		}

		thickness := input.Thickness
		if thickness == 0 {
			thickness = 1
		}

		// Generate Lua script
		script := gen.DrawLine(input.LayerName, input.FrameNumber, input.X1, input.Y1, input.X2, input.Y2, color, thickness)

		// Execute
		_, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to draw line: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Drew line from (%d,%d) to (%d,%d)", input.X1, input.Y1, input.X2, input.Y2)},
			},
		}, &DrawingOutput{Success: true}, nil
	})

	// draw_rectangle tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "draw_rectangle",
		Description: "Draw a rectangle",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DrawRectangleInput) (*mcp.CallToolResult, *DrawingOutput, error) {
		var color aseprite.Color
		if err := color.FromHex(input.Color); err != nil {
			return nil, nil, fmt.Errorf("invalid color: %w", err)
		}

		// Generate Lua script
		script := gen.DrawRectangle(input.LayerName, input.FrameNumber, input.X, input.Y, input.Width, input.Height, color, input.Filled)

		// Execute
		_, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to draw rectangle: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Drew %dx%d rectangle at (%d,%d)", input.Width, input.Height, input.X, input.Y)},
			},
		}, &DrawingOutput{Success: true}, nil
	})

	// draw_circle tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "draw_circle",
		Description: "Draw a circle",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DrawCircleInput) (*mcp.CallToolResult, *DrawingOutput, error) {
		var color aseprite.Color
		if err := color.FromHex(input.Color); err != nil {
			return nil, nil, fmt.Errorf("invalid color: %w", err)
		}

		// Generate Lua script
		script := gen.DrawCircle(input.LayerName, input.FrameNumber, input.CenterX, input.CenterY, input.Radius, color, input.Filled)

		// Execute
		_, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to draw circle: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Drew circle with radius %d at (%d,%d)", input.Radius, input.CenterX, input.CenterY)},
			},
		}, &DrawingOutput{Success: true}, nil
	})

	// fill_area tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "fill_area",
		Description: "Flood fill an area (paint bucket tool)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input FillAreaInput) (*mcp.CallToolResult, *DrawingOutput, error) {
		var color aseprite.Color
		if err := color.FromHex(input.Color); err != nil {
			return nil, nil, fmt.Errorf("invalid color: %w", err)
		}

		// Generate Lua script
		script := gen.FillArea(input.LayerName, input.FrameNumber, input.X, input.Y, color)

		// Execute
		_, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fill area: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Filled area starting at (%d,%d)", input.X, input.Y)},
			},
		}, &DrawingOutput{Success: true}, nil
	})
}