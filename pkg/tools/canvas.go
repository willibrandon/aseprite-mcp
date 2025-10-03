package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/mtlog/core"
)

// CreateCanvasInput defines the input parameters for the create_canvas tool.
type CreateCanvasInput struct {
	Width     int    `json:"width" jsonschema:"Canvas width in pixels (1-65535)"`
	Height    int    `json:"height" jsonschema:"Canvas height in pixels (1-65535)"`
	ColorMode string `json:"color_mode" jsonschema:"Color mode: rgb, grayscale, or indexed"`
}

// CreateCanvasOutput defines the output for the create_canvas tool.
type CreateCanvasOutput struct {
	FilePath string `json:"file_path" jsonschema:"Absolute path to the created Aseprite file"`
}

// AddLayerInput defines the input parameters for the add_layer tool.
type AddLayerInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName  string `json:"layer_name" jsonschema:"Name for the new layer"`
}

// AddLayerOutput defines the output for the add_layer tool.
type AddLayerOutput struct {
	Success bool `json:"success" jsonschema:"Whether the layer was added successfully"`
}

// AddFrameInput defines the input parameters for the add_frame tool.
type AddFrameInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	DurationMs int    `json:"duration_ms" jsonschema:"Frame duration in milliseconds (1-65535)"`
}

// AddFrameOutput defines the output for the add_frame tool.
type AddFrameOutput struct {
	FrameNumber int `json:"frame_number" jsonschema:"Index of the created frame (1-based)"`
}

// GetSpriteInfoInput defines the input parameters for the get_sprite_info tool.
type GetSpriteInfoInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
}

// GetSpriteInfoOutput defines the output for the get_sprite_info tool.
type GetSpriteInfoOutput struct {
	Width      int      `json:"width" jsonschema:"Sprite width in pixels"`
	Height     int      `json:"height" jsonschema:"Sprite height in pixels"`
	ColorMode  string   `json:"color_mode" jsonschema:"Color mode (rgb, grayscale, or indexed)"`
	FrameCount int      `json:"frame_count" jsonschema:"Number of frames in the sprite"`
	LayerCount int      `json:"layer_count" jsonschema:"Number of layers in the sprite"`
	Layers     []string `json:"layers" jsonschema:"Names of all layers in the sprite"`
}

// RegisterCanvasTools registers all canvas management tools with the MCP server.
func RegisterCanvasTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register create_canvas tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "create_canvas",
			Description: "Create a new Aseprite sprite with specified dimensions and color mode. Returns the path to the created .aseprite file.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input CreateCanvasInput) (*mcp.CallToolResult, *CreateCanvasOutput, error) {
			logger.Debug("create_canvas tool called", "width", input.Width, "height", input.Height, "color_mode", input.ColorMode)

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

			// Generate filename in temp directory
			filename := filepath.Join(cfg.TempDir, fmt.Sprintf("sprite-%d.aseprite", generateTimestamp()))

			// Generate Lua script
			script := gen.CreateCanvas(input.Width, input.Height, colorMode, filename)

			// Execute Lua script with Aseprite
			output, err := client.ExecuteLua(ctx, script, "")
			if err != nil {
				logger.Error("Failed to create canvas", "error", err)
				return nil, nil, fmt.Errorf("failed to create canvas: %w", err)
			}

			// Parse output to get file path
			filePath := strings.TrimSpace(output)

			logger.Information("Canvas created successfully", "path", filePath, "width", input.Width, "height", input.Height)

			return nil, &CreateCanvasOutput{FilePath: filePath}, nil
		},
	)

	// Register add_layer tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "add_layer",
			Description: "Add a new layer to an existing Aseprite sprite. Returns success status.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input AddLayerInput) (*mcp.CallToolResult, *AddLayerOutput, error) {
			logger.Debug("add_layer tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName)

			// Validate layer name
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			// Generate Lua script
			script := gen.AddLayer(input.LayerName)

			// Execute Lua script with the sprite
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to add layer", "error", err)
				return nil, nil, fmt.Errorf("failed to add layer: %w", err)
			}

			logger.Information("Layer added successfully", "sprite", input.SpritePath, "layer", input.LayerName)

			return nil, &AddLayerOutput{Success: true}, nil
		},
	)

	// Register add_frame tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "add_frame",
			Description: "Add a new frame to an existing Aseprite sprite. Returns the frame number (1-based index).",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input AddFrameInput) (*mcp.CallToolResult, *AddFrameOutput, error) {
			logger.Debug("add_frame tool called", "sprite_path", input.SpritePath, "duration_ms", input.DurationMs)

			// Validate duration
			if input.DurationMs < 1 || input.DurationMs > 65535 {
				return nil, nil, fmt.Errorf("duration_ms must be between 1 and 65535, got %d", input.DurationMs)
			}

			// Generate Lua script
			script := gen.AddFrame(input.DurationMs)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to add frame", "error", err)
				return nil, nil, fmt.Errorf("failed to add frame: %w", err)
			}

			// Parse frame number from output
			var frameNumber int
			_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &frameNumber)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse frame number from output: %w", err)
			}

			logger.Information("Frame added successfully", "sprite", input.SpritePath, "frame_number", frameNumber)

			return nil, &AddFrameOutput{FrameNumber: frameNumber}, nil
		},
	)

	// Register get_sprite_info tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_sprite_info",
			Description: "Retrieve metadata about an existing Aseprite sprite including dimensions, color mode, frame count, layer count, and layer names.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input GetSpriteInfoInput) (*mcp.CallToolResult, *GetSpriteInfoOutput, error) {
			logger.Debug("get_sprite_info tool called", "sprite_path", input.SpritePath)

			// Generate Lua script
			script := gen.GetSpriteInfo()

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to get sprite info", "error", err)
				return nil, nil, fmt.Errorf("failed to get sprite info: %w", err)
			}

			// Parse JSON output
			var info GetSpriteInfoOutput
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &info); err != nil {
				logger.Error("Failed to parse sprite info", "error", err, "output", output)
				return nil, nil, fmt.Errorf("failed to parse sprite info: %w", err)
			}

			logger.Information("Sprite info retrieved successfully", "sprite", input.SpritePath, "width", info.Width, "height", info.Height)

			return nil, &info, nil
		},
	)

	// Register delete_layer tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "delete_layer",
			Description: "Delete a layer from an existing sprite. Cannot delete the last remaining layer.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input DeleteLayerInput) (*mcp.CallToolResult, *DeleteLayerOutput, error) {
			logger.Debug("delete_layer tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName)

			// Generate Lua script
			script := gen.DeleteLayer(input.LayerName)

			// Execute Lua script with the sprite
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to delete layer", "error", err)
				return nil, nil, fmt.Errorf("failed to delete layer: %w", err)
			}

			logger.Information("Layer deleted successfully", "sprite", input.SpritePath, "layer", input.LayerName)

			return nil, &DeleteLayerOutput{Success: true}, nil
		},
	)

	// Register delete_frame tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "delete_frame",
			Description: "Delete a frame from an existing sprite. Cannot delete the last remaining frame.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input DeleteFrameInput) (*mcp.CallToolResult, *DeleteFrameOutput, error) {
			logger.Debug("delete_frame tool called", "sprite_path", input.SpritePath, "frame_number", input.FrameNumber)

			// Generate Lua script
			script := gen.DeleteFrame(input.FrameNumber)

			// Execute Lua script with the sprite
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to delete frame", "error", err)
				return nil, nil, fmt.Errorf("failed to delete frame: %w", err)
			}

			logger.Information("Frame deleted successfully", "sprite", input.SpritePath, "frame_number", input.FrameNumber)

			return nil, &DeleteFrameOutput{Success: true}, nil
		},
	)
}

// DeleteLayerInput defines the input parameters for the delete_layer tool.
type DeleteLayerInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the sprite file"`
	LayerName  string `json:"layer_name" jsonschema:"Name of the layer to delete"`
}

// DeleteLayerOutput defines the output for the delete_layer tool.
type DeleteLayerOutput struct {
	Success bool `json:"success" jsonschema:"Whether the layer was deleted successfully"`
}

// DeleteFrameInput defines the input parameters for the delete_frame tool.
type DeleteFrameInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the sprite file"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to delete (1-based)"`
}

// DeleteFrameOutput defines the output for the delete_frame tool.
type DeleteFrameOutput struct {
	Success bool `json:"success" jsonschema:"Whether the frame was deleted successfully"`
}

// generateTimestamp returns a Unix timestamp in nanoseconds suitable for unique filenames.
func generateTimestamp() int64 {
	return time.Now().UnixNano()
}
