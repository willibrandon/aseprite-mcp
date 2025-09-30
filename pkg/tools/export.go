package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/mtlog/core"
)

// ExportSpriteInput defines the input parameters for the export_sprite tool.
type ExportSpriteInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	OutputPath  string `json:"output_path" jsonschema:"Output file path for exported image"`
	Format      string `json:"format" jsonschema:"Export format: png, gif, jpg, bmp"`
	FrameNumber int    `json:"frame_number" jsonschema:"Specific frame to export (0 = all frames, 1-based)"`
}

// ExportSpriteOutput defines the output for the export_sprite tool.
type ExportSpriteOutput struct {
	ExportedPath string `json:"exported_path" jsonschema:"Path to the exported file"`
	FileSize     int64  `json:"file_size" jsonschema:"Size of exported file in bytes"`
}

// RegisterExportTools registers all export tools with the MCP server.
func RegisterExportTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register export_sprite tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "export_sprite",
			Description: "Export sprite to common image formats (PNG, GIF, JPG, BMP).",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input ExportSpriteInput) (*mcp.CallToolResult, *ExportSpriteOutput, error) {
			logger.Debug("export_sprite tool called", "sprite_path", input.SpritePath, "output_path", input.OutputPath, "format", input.Format, "frame_number", input.FrameNumber)

			// Validate inputs
			if input.OutputPath == "" {
				return nil, nil, fmt.Errorf("output_path cannot be empty")
			}

			// Validate format
			validFormats := map[string]bool{
				"png": true,
				"gif": true,
				"jpg": true,
				"bmp": true,
			}
			format := strings.ToLower(input.Format)
			if !validFormats[format] {
				return nil, nil, fmt.Errorf("invalid format: %s (valid: png, gif, jpg, bmp)", input.Format)
			}

			// Validate frame number
			if input.FrameNumber < 0 {
				return nil, nil, fmt.Errorf("frame_number must be non-negative, got %d", input.FrameNumber)
			}

			// Ensure output directory exists
			outputDir := filepath.Dir(input.OutputPath)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
			}

			// Generate Lua script
			script := gen.ExportSprite(input.OutputPath, input.FrameNumber)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				logger.Error("Failed to export sprite", "error", err)
				return nil, nil, fmt.Errorf("failed to export sprite: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Exported successfully") {
				logger.Warning("Unexpected output from export_sprite", "output", output)
			}

			// Get file size
			fileInfo, err := os.Stat(input.OutputPath)
			if err != nil {
				logger.Warning("Failed to stat exported file", "error", err)
				return nil, &ExportSpriteOutput{
					ExportedPath: input.OutputPath,
					FileSize:     0,
				}, nil
			}

			logger.Information("Sprite exported successfully", "sprite", input.SpritePath, "output", input.OutputPath, "format", format, "frame", input.FrameNumber, "size", fileInfo.Size())

			return nil, &ExportSpriteOutput{
				ExportedPath: input.OutputPath,
				FileSize:     fileInfo.Size(),
			}, nil
		},
	)
}