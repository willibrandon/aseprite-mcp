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

// SelectRectangleInput defines the input parameters for the select_rectangle tool.
type SelectRectangleInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	X          int    `json:"x" jsonschema:"X coordinate of selection rectangle"`
	Y          int    `json:"y" jsonschema:"Y coordinate of selection rectangle"`
	Width      int    `json:"width" jsonschema:"Width of selection rectangle (minimum 1)"`
	Height     int    `json:"height" jsonschema:"Height of selection rectangle (minimum 1)"`
	Mode       string `json:"mode" jsonschema:"Selection mode: replace, add, subtract, or intersect (default: replace)"`
}

// SelectRectangleOutput defines the output for the select_rectangle tool.
type SelectRectangleOutput struct {
	Success bool `json:"success" jsonschema:"Whether the selection was created successfully"`
}

// SelectEllipseInput defines the input parameters for the select_ellipse tool.
type SelectEllipseInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	X          int    `json:"x" jsonschema:"X coordinate of selection ellipse bounding box"`
	Y          int    `json:"y" jsonschema:"Y coordinate of selection ellipse bounding box"`
	Width      int    `json:"width" jsonschema:"Width of selection ellipse (minimum 1)"`
	Height     int    `json:"height" jsonschema:"Height of selection ellipse (minimum 1)"`
	Mode       string `json:"mode" jsonschema:"Selection mode: replace, add, subtract, or intersect (default: replace)"`
}

// SelectEllipseOutput defines the output for the select_ellipse tool.
type SelectEllipseOutput struct {
	Success bool `json:"success" jsonschema:"Whether the selection was created successfully"`
}

// SelectAllInput defines the input parameters for the select_all tool.
type SelectAllInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
}

// SelectAllOutput defines the output for the select_all tool.
type SelectAllOutput struct {
	Success bool `json:"success" jsonschema:"Whether select all was successful"`
}

// DeselectInput defines the input parameters for the deselect tool.
type DeselectInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
}

// DeselectOutput defines the output for the deselect tool.
type DeselectOutput struct {
	Success bool `json:"success" jsonschema:"Whether deselect was successful"`
}

// MoveSelectionInput defines the input parameters for the move_selection tool.
type MoveSelectionInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	DX         int    `json:"dx" jsonschema:"Horizontal offset in pixels (can be negative)"`
	DY         int    `json:"dy" jsonschema:"Vertical offset in pixels (can be negative)"`
}

// MoveSelectionOutput defines the output for the move_selection tool.
type MoveSelectionOutput struct {
	Success bool `json:"success" jsonschema:"Whether the selection was moved successfully"`
}

// CutSelectionInput defines the input parameters for the cut_selection tool.
type CutSelectionInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to cut from"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to cut from (1-based index)"`
}

// CutSelectionOutput defines the output for the cut_selection tool.
type CutSelectionOutput struct {
	Success bool `json:"success" jsonschema:"Whether the cut was successful"`
}

// CopySelectionInput defines the input parameters for the copy_selection tool.
type CopySelectionInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
}

// CopySelectionOutput defines the output for the copy_selection tool.
type CopySelectionOutput struct {
	Success bool `json:"success" jsonschema:"Whether the copy was successful"`
}

// PasteClipboardInput defines the input parameters for the paste_clipboard tool.
type PasteClipboardInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to paste onto"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to paste onto (1-based index)"`
	X           *int   `json:"x,omitempty" jsonschema:"X coordinate for paste position (optional)"`
	Y           *int   `json:"y,omitempty" jsonschema:"Y coordinate for paste position (optional)"`
}

// PasteClipboardOutput defines the output for the paste_clipboard tool.
type PasteClipboardOutput struct {
	Success bool `json:"success" jsonschema:"Whether the paste was successful"`
}

// RegisterSelectionTools registers all selection-related MCP tools with the server.
func RegisterSelectionTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register select_rectangle tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "select_rectangle",
			Description: "Create a rectangular selection with specified mode (replace/add/subtract/intersect). Selections define which pixels will be affected by editing operations.",
		},
		maybeWrapWithTiming("select_rectangle", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input SelectRectangleInput) (*mcp.CallToolResult, *SelectRectangleOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("select_rectangle tool called", "sprite_path", input.SpritePath, "x", input.X, "y", input.Y, "width", input.Width, "height", input.Height, "mode", input.Mode)

			// Validate inputs
			if input.Width < 1 {
				return nil, nil, fmt.Errorf("width must be at least 1, got %d", input.Width)
			}

			if input.Height < 1 {
				return nil, nil, fmt.Errorf("height must be at least 1, got %d", input.Height)
			}

			// Validate mode
			mode := input.Mode
			if mode == "" {
				mode = "replace"
			}
			validModes := map[string]bool{"replace": true, "add": true, "subtract": true, "intersect": true}
			if !validModes[mode] {
				return nil, nil, fmt.Errorf("invalid mode %q, must be one of: replace, add, subtract, intersect", mode)
			}

			// Generate Lua script
			script := gen.SelectRectangle(input.X, input.Y, input.Width, input.Height, mode)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to create rectangle selection", "error", err)
				return nil, nil, fmt.Errorf("failed to create rectangle selection: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Rectangle selection created successfully") {
				opLogger.Warning("Unexpected output from select_rectangle", "output", output)
			}

			opLogger.Information("Rectangle selection created successfully", "sprite", input.SpritePath, "mode", mode)

			return nil, &SelectRectangleOutput{Success: true}, nil
		}),
	)

	// Register select_ellipse tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "select_ellipse",
			Description: "Create an elliptical selection with specified mode (replace/add/subtract/intersect). The ellipse is defined by a bounding box.",
		},
		maybeWrapWithTiming("select_ellipse", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input SelectEllipseInput) (*mcp.CallToolResult, *SelectEllipseOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("select_ellipse tool called", "sprite_path", input.SpritePath, "x", input.X, "y", input.Y, "width", input.Width, "height", input.Height, "mode", input.Mode)

			// Validate inputs
			if input.Width < 1 {
				return nil, nil, fmt.Errorf("width must be at least 1, got %d", input.Width)
			}

			if input.Height < 1 {
				return nil, nil, fmt.Errorf("height must be at least 1, got %d", input.Height)
			}

			// Validate mode
			mode := input.Mode
			if mode == "" {
				mode = "replace"
			}
			validModes := map[string]bool{"replace": true, "add": true, "subtract": true, "intersect": true}
			if !validModes[mode] {
				return nil, nil, fmt.Errorf("invalid mode %q, must be one of: replace, add, subtract, intersect", mode)
			}

			// Generate Lua script
			script := gen.SelectEllipse(input.X, input.Y, input.Width, input.Height, mode)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to create ellipse selection", "error", err)
				return nil, nil, fmt.Errorf("failed to create ellipse selection: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Ellipse selection created successfully") {
				opLogger.Warning("Unexpected output from select_ellipse", "output", output)
			}

			opLogger.Information("Ellipse selection created successfully", "sprite", input.SpritePath, "mode", mode)

			return nil, &SelectEllipseOutput{Success: true}, nil
		}),
	)

	// Register select_all tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "select_all",
			Description: "Select the entire canvas. This selects all pixels in the sprite, regardless of layers or frames.",
		},
		maybeWrapWithTiming("select_all", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input SelectAllInput) (*mcp.CallToolResult, *SelectAllOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("select_all tool called", "sprite_path", input.SpritePath)

			// Generate Lua script
			script := gen.SelectAll()

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to select all", "error", err)
				return nil, nil, fmt.Errorf("failed to select all: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Select all completed successfully") {
				opLogger.Warning("Unexpected output from select_all", "output", output)
			}

			opLogger.Information("Select all completed successfully", "sprite", input.SpritePath)

			return nil, &SelectAllOutput{Success: true}, nil
		}),
	)

	// Register deselect tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "deselect",
			Description: "Clear the current selection. Removes any active selection from the sprite.",
		},
		maybeWrapWithTiming("deselect", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input DeselectInput) (*mcp.CallToolResult, *DeselectOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("deselect tool called", "sprite_path", input.SpritePath)

			// Generate Lua script
			script := gen.Deselect()

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to deselect", "error", err)
				return nil, nil, fmt.Errorf("failed to deselect: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Deselect completed successfully") {
				opLogger.Warning("Unexpected output from deselect", "output", output)
			}

			opLogger.Information("Deselect completed successfully", "sprite", input.SpritePath)

			return nil, &DeselectOutput{Success: true}, nil
		}),
	)

	// Register move_selection tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "move_selection",
			Description: "Move the current selection by a specified offset. Does not move the pixel content, only the selection bounds. Requires an active selection.",
		},
		maybeWrapWithTiming("move_selection", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input MoveSelectionInput) (*mcp.CallToolResult, *MoveSelectionOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("move_selection tool called", "sprite_path", input.SpritePath, "dx", input.DX, "dy", input.DY)

			// Generate Lua script
			script := gen.MoveSelection(input.DX, input.DY)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to move selection", "error", err)
				return nil, nil, fmt.Errorf("failed to move selection: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Selection moved successfully") {
				opLogger.Warning("Unexpected output from move_selection", "output", output)
			}

			opLogger.Information("Selection moved successfully", "sprite", input.SpritePath, "dx", input.DX, "dy", input.DY)

			return nil, &MoveSelectionOutput{Success: true}, nil
		}),
	)

	// Register cut_selection tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "cut_selection",
			Description: "Cut the selected pixels to clipboard. Removes pixels from the specified layer and frame, placing them on the clipboard. Requires an active selection.",
		},
		maybeWrapWithTiming("cut_selection", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input CutSelectionInput) (*mcp.CallToolResult, *CutSelectionOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("cut_selection tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName, "frame_number", input.FrameNumber)

			// Validate inputs
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be at least 1, got %d", input.FrameNumber)
			}

			// Generate Lua script
			script := gen.CutSelection(input.LayerName, input.FrameNumber)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to cut selection", "error", err)
				return nil, nil, fmt.Errorf("failed to cut selection: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Cut selection completed successfully") {
				opLogger.Warning("Unexpected output from cut_selection", "output", output)
			}

			opLogger.Information("Cut selection completed successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber)

			return nil, &CutSelectionOutput{Success: true}, nil
		}),
	)

	// Register copy_selection tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "copy_selection",
			Description: "Copy the selected pixels to clipboard without removing them. Requires an active selection.",
		},
		maybeWrapWithTiming("copy_selection", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input CopySelectionInput) (*mcp.CallToolResult, *CopySelectionOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("copy_selection tool called", "sprite_path", input.SpritePath)

			// Generate Lua script
			script := gen.CopySelection()

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to copy selection", "error", err)
				return nil, nil, fmt.Errorf("failed to copy selection: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Copy selection completed successfully") {
				opLogger.Warning("Unexpected output from copy_selection", "output", output)
			}

			opLogger.Information("Copy selection completed successfully", "sprite", input.SpritePath)

			return nil, &CopySelectionOutput{Success: true}, nil
		}),
	)

	// Register paste_clipboard tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "paste_clipboard",
			Description: "Paste clipboard content onto the specified layer and frame. Optionally specify paste position (x, y). Requires clipboard to contain image data.",
		},
		maybeWrapWithTiming("paste_clipboard", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input PasteClipboardInput) (*mcp.CallToolResult, *PasteClipboardOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("paste_clipboard tool called", "sprite_path", input.SpritePath, "layer_name", input.LayerName, "frame_number", input.FrameNumber, "x", input.X, "y", input.Y)

			// Validate inputs
			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be at least 1, got %d", input.FrameNumber)
			}

			// Generate Lua script
			script := gen.PasteClipboard(input.LayerName, input.FrameNumber, input.X, input.Y)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to paste clipboard", "error", err)
				return nil, nil, fmt.Errorf("failed to paste clipboard: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Paste completed successfully") {
				opLogger.Warning("Unexpected output from paste_clipboard", "output", output)
			}

			opLogger.Information("Paste completed successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber)

			return nil, &PasteClipboardOutput{Success: true}, nil
		}),
	)
}
