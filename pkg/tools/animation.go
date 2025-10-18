package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
	"github.com/willibrandon/pixel-mcp/pkg/config"
)

// SetFrameDurationInput defines the input parameters for the set_frame_duration tool.
type SetFrameDurationInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to modify (1-based)"`
	DurationMs  int    `json:"duration_ms" jsonschema:"Frame duration in milliseconds (1-65535)"`
}

// SetFrameDurationOutput defines the output for the set_frame_duration tool.
type SetFrameDurationOutput struct {
	Success bool `json:"success" jsonschema:"Whether the frame duration was set successfully"`
}

// CreateTagInput defines the input parameters for the create_tag tool.
type CreateTagInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	TagName    string `json:"tag_name" jsonschema:"Name for the animation tag"`
	FromFrame  int    `json:"from_frame" jsonschema:"Starting frame number (1-based, inclusive)"`
	ToFrame    int    `json:"to_frame" jsonschema:"Ending frame number (1-based, inclusive)"`
	Direction  string `json:"direction" jsonschema:"Playback direction: forward, reverse, or pingpong"`
}

// CreateTagOutput defines the output for the create_tag tool.
type CreateTagOutput struct {
	Success bool `json:"success" jsonschema:"Whether the tag was created successfully"`
}

// DuplicateFrameInput defines the input parameters for the duplicate_frame tool.
type DuplicateFrameInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	SourceFrame int    `json:"source_frame" jsonschema:"Frame number to duplicate (1-based)"`
	InsertAfter int    `json:"insert_after" jsonschema:"Insert duplicated frame after this frame number (1-based, 0 = insert at end)"`
}

// DuplicateFrameOutput defines the output for the duplicate_frame tool.
type DuplicateFrameOutput struct {
	NewFrameNumber int `json:"new_frame_number" jsonschema:"Index of the newly created frame (1-based)"`
}

// LinkCelInput defines the input parameters for the link_cel tool.
type LinkCelInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer"`
	SourceFrame int    `json:"source_frame" jsonschema:"Source frame with the cel to link (1-based)"`
	TargetFrame int    `json:"target_frame" jsonschema:"Target frame where linked cel will be created (1-based)"`
}

// DeleteTagInput defines the input parameters for the delete_tag tool.
type DeleteTagInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	TagName    string `json:"tag_name" jsonschema:"Name of the tag to delete"`
}

// DeleteTagOutput defines the output for the delete_tag tool.
type DeleteTagOutput struct {
	Success bool `json:"success" jsonschema:"Deletion success status"`
}

// LinkCelOutput defines the output for the link_cel tool.
type LinkCelOutput struct {
	Success bool `json:"success" jsonschema:"Whether the cel was linked successfully"`
}

// RegisterAnimationTools registers all animation tools with the MCP server.
func RegisterAnimationTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register set_frame_duration tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "set_frame_duration",
			Description: "Set the duration of an existing animation frame in milliseconds.",
		},
		maybeWrapWithTiming("set_frame_duration", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input SetFrameDurationInput) (*mcp.CallToolResult, *SetFrameDurationOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("set_frame_duration tool called", "sprite_path", input.SpritePath, "frame_number", input.FrameNumber, "duration_ms", input.DurationMs)

			// Validate inputs
			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be at least 1, got %d", input.FrameNumber)
			}

			if input.DurationMs < 1 || input.DurationMs > 65535 {
				return nil, nil, fmt.Errorf("duration_ms must be between 1 and 65535, got %d", input.DurationMs)
			}

			// Generate Lua script
			script := gen.SetFrameDuration(input.FrameNumber, input.DurationMs)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to set frame duration", "error", err)
				return nil, nil, fmt.Errorf("failed to set frame duration: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Frame duration set successfully") {
				opLogger.Warning("Unexpected output from set_frame_duration", "output", output)
			}

			opLogger.Information("Frame duration set successfully", "sprite", input.SpritePath, "frame", input.FrameNumber, "duration_ms", input.DurationMs)

			return nil, &SetFrameDurationOutput{Success: true}, nil
		}),
	)

	// Register create_tag tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "create_tag",
			Description: "Create an animation tag to define a named frame range with playback direction.",
		},
		maybeWrapWithTiming("create_tag", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input CreateTagInput) (*mcp.CallToolResult, *CreateTagOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("create_tag tool called", "sprite_path", input.SpritePath, "tag_name", input.TagName, "from_frame", input.FromFrame, "to_frame", input.ToFrame, "direction", input.Direction)

			// Validate inputs
			if input.TagName == "" {
				return nil, nil, fmt.Errorf("tag_name cannot be empty")
			}

			if input.FromFrame < 1 {
				return nil, nil, fmt.Errorf("from_frame must be at least 1, got %d", input.FromFrame)
			}

			if input.ToFrame < input.FromFrame {
				return nil, nil, fmt.Errorf("to_frame must be >= from_frame, got from=%d to=%d", input.FromFrame, input.ToFrame)
			}

			// Validate direction
			validDirections := map[string]bool{
				"forward":  true,
				"reverse":  true,
				"pingpong": true,
			}
			if !validDirections[input.Direction] {
				return nil, nil, fmt.Errorf("invalid direction: %s (valid: forward, reverse, pingpong)", input.Direction)
			}

			// Generate Lua script
			script := gen.CreateTag(input.TagName, input.FromFrame, input.ToFrame, input.Direction)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to create tag", "error", err)
				return nil, nil, fmt.Errorf("failed to create tag: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Tag created successfully") {
				opLogger.Warning("Unexpected output from create_tag", "output", output)
			}

			opLogger.Information("Tag created successfully", "sprite", input.SpritePath, "tag_name", input.TagName, "from_frame", input.FromFrame, "to_frame", input.ToFrame, "direction", input.Direction)

			return nil, &CreateTagOutput{Success: true}, nil
		}),
	)

	// Register duplicate_frame tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "duplicate_frame",
			Description: "Duplicate an existing frame and insert it at the specified position.",
		},
		maybeWrapWithTiming("duplicate_frame", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input DuplicateFrameInput) (*mcp.CallToolResult, *DuplicateFrameOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("duplicate_frame tool called", "sprite_path", input.SpritePath, "source_frame", input.SourceFrame, "insert_after", input.InsertAfter)

			// Validate inputs
			if input.SourceFrame < 1 {
				return nil, nil, fmt.Errorf("source_frame must be at least 1, got %d", input.SourceFrame)
			}

			if input.InsertAfter < 0 {
				return nil, nil, fmt.Errorf("insert_after must be non-negative, got %d", input.InsertAfter)
			}

			// Generate Lua script
			script := gen.DuplicateFrame(input.SourceFrame, input.InsertAfter)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to duplicate frame", "error", err)
				return nil, nil, fmt.Errorf("failed to duplicate frame: %w", err)
			}

			// Parse frame number from output
			var newFrameNumber int
			_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &newFrameNumber)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse new frame number from output: %w", err)
			}

			opLogger.Information("Frame duplicated successfully", "sprite", input.SpritePath, "source_frame", input.SourceFrame, "new_frame_number", newFrameNumber)

			return nil, &DuplicateFrameOutput{NewFrameNumber: newFrameNumber}, nil
		}),
	)

	// Register link_cel tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "link_cel",
			Description: "Create a linked cel that references another cel's image data, useful for animation optimization.",
		},
		maybeWrapWithTiming("link_cel", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input LinkCelInput) (*mcp.CallToolResult, *LinkCelOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("link_cel tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName, "source_frame", input.SourceFrame, "target_frame", input.TargetFrame)

			// Validate inputs
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.SourceFrame < 1 {
				return nil, nil, fmt.Errorf("source_frame must be at least 1, got %d", input.SourceFrame)
			}

			if input.TargetFrame < 1 {
				return nil, nil, fmt.Errorf("target_frame must be at least 1, got %d", input.TargetFrame)
			}

			if input.SourceFrame == input.TargetFrame {
				return nil, nil, fmt.Errorf("source_frame and target_frame cannot be the same")
			}

			// Generate Lua script
			script := gen.LinkCel(input.LayerName, input.SourceFrame, input.TargetFrame)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to link cel", "error", err)
				return nil, nil, fmt.Errorf("failed to link cel: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Cel linked successfully") {
				opLogger.Warning("Unexpected output from link_cel", "output", output)
			}

			opLogger.Information("Cel linked successfully", "sprite", input.SpritePath, "layer", input.LayerName, "source_frame", input.SourceFrame, "target_frame", input.TargetFrame)

			return nil, &LinkCelOutput{Success: true}, nil
		}),
	)

	// Register delete_tag tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "delete_tag",
			Description: "Delete an animation tag by name.",
		},
		maybeWrapWithTiming("delete_tag", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteTagInput) (*mcp.CallToolResult, *DeleteTagOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("delete_tag tool called", "sprite_path", input.SpritePath, "tag_name", input.TagName)

			// Validate inputs
			if input.SpritePath == "" {
				return nil, nil, fmt.Errorf("sprite_path cannot be empty")
			}

			if input.TagName == "" {
				return nil, nil, fmt.Errorf("tag_name cannot be empty")
			}

			// Generate Lua script
			script := gen.DeleteTag(input.TagName)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to delete tag", "error", err)
				return nil, nil, fmt.Errorf("failed to delete tag: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Tag deleted successfully") {
				opLogger.Warning("Unexpected output from delete_tag", "output", output)
			}

			opLogger.Information("Tag deleted successfully", "sprite", input.SpritePath, "tag", input.TagName)

			return nil, &DeleteTagOutput{Success: true}, nil
		}),
	)
}
