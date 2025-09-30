package tools

import (
	"context"
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
}

// generateTimestamp returns a Unix timestamp in nanoseconds suitable for unique filenames.
func generateTimestamp() int64 {
	return time.Now().UnixNano()
}