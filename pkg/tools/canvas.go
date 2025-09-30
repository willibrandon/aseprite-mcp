package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

type CreateCanvasInput struct {
	Width     int    `json:"width" jsonschema:"required,minimum=1,maximum=65535,description=Canvas width in pixels"`
	Height    int    `json:"height" jsonschema:"required,minimum=1,maximum=65535,description=Canvas height in pixels"`
	ColorMode string `json:"color_mode" jsonschema:"enum=rgb,enum=grayscale,enum=indexed,description=Color mode for the sprite"`
}

type CreateCanvasOutput struct {
	FilePath string `json:"file_path" jsonschema:"description=Path to the created Aseprite file"`
}

type AddLayerInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
	LayerName  string `json:"layer_name" jsonschema:"required,description=Name for the new layer"`
}

type AddLayerOutput struct {
	Success bool `json:"success" jsonschema:"description=Whether the operation succeeded"`
}

type AddFrameInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
	DurationMs int    `json:"duration_ms" jsonschema:"required,minimum=1,maximum=65535,description=Frame duration in milliseconds"`
}

type AddFrameOutput struct {
	FrameNumber int `json:"frame_number" jsonschema:"description=Index of the created frame"`
}

type GetSpriteInfoInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
}

func RegisterCanvasTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator) {
	// create_canvas tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_canvas",
		Description: "Create a new Aseprite sprite with specified dimensions and color mode",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateCanvasInput) (*mcp.CallToolResult, *CreateCanvasOutput, error) {
		// Validate color mode
		var colorMode aseprite.ColorMode
		switch input.ColorMode {
		case "rgb":
			colorMode = aseprite.ColorModeRGB
		case "grayscale":
			colorMode = aseprite.ColorModeGrayscale
		case "indexed":
			colorMode = aseprite.ColorModeIndexed
		default:
			return nil, nil, fmt.Errorf("invalid color mode: %s (must be rgb, grayscale, or indexed)", input.ColorMode)
		}

		// Generate Lua script
		script := gen.CreateCanvas(input.Width, input.Height, colorMode)

		// Execute
		output, err := client.ExecuteLua(ctx, script, "")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create canvas: %w", err)
		}

		// Parse output (file path)
		filePath := strings.TrimSpace(output)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Created sprite at: %s", filePath)},
			},
		}, &CreateCanvasOutput{FilePath: filePath}, nil
	})

	// add_layer tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_layer",
		Description: "Add a new layer to an existing sprite",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AddLayerInput) (*mcp.CallToolResult, *AddLayerOutput, error) {
		// Generate Lua script
		script := gen.AddLayer(input.LayerName)

		// Execute
		_, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to add layer: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Added layer '%s' successfully", input.LayerName)},
			},
		}, &AddLayerOutput{Success: true}, nil
	})

	// add_frame tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_frame",
		Description: "Add a new frame to the sprite timeline",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AddFrameInput) (*mcp.CallToolResult, *AddFrameOutput, error) {
		// Generate Lua script
		script := gen.AddFrame(input.DurationMs)

		// Execute
		output, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to add frame: %w", err)
		}

		// Parse frame number from output
		frameNum := 0
		fmt.Sscanf(strings.TrimSpace(output), "%d", &frameNum)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Added frame %d with duration %dms", frameNum, input.DurationMs)},
			},
		}, &AddFrameOutput{FrameNumber: frameNum}, nil
	})

	// get_sprite_info tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_sprite_info",
		Description: "Retrieve metadata about a sprite",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSpriteInfoInput) (*mcp.CallToolResult, *aseprite.SpriteInfo, error) {
		// Generate Lua script
		script := gen.GetSpriteInfo()

		// Execute
		output, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get sprite info: %w", err)
		}

		// Parse JSON output
		var info aseprite.SpriteInfo
		if err := json.Unmarshal([]byte(output), &info); err != nil {
			return nil, nil, fmt.Errorf("failed to parse sprite info: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Sprite: %dx%d, %d frames, %d layers", info.Width, info.Height, info.FrameCount, info.LayerCount)},
			},
		}, &info, nil
	})
}