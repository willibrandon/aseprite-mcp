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

// ExportSpritesheetInput defines the input parameters for the export_spritesheet tool.
type ExportSpritesheetInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	OutputPath  string `json:"output_path" jsonschema:"Output file path for spritesheet"`
	Layout      string `json:"layout" jsonschema:"Spritesheet layout: horizontal, vertical, rows, columns, or packed"`
	Padding     int    `json:"padding" jsonschema:"Padding between frames in pixels (0-100)"`
	IncludeJSON bool   `json:"include_json" jsonschema:"Include JSON metadata file"`
}

// ExportSpritesheetOutput defines the output for the export_spritesheet tool.
type ExportSpritesheetOutput struct {
	SpritesheetPath string  `json:"spritesheet_path" jsonschema:"Path to exported spritesheet"`
	MetadataPath    *string `json:"metadata_path,omitempty" jsonschema:"Path to JSON metadata if included"`
	FrameCount      int     `json:"frame_count" jsonschema:"Number of frames in spritesheet"`
}

// ImportImageInput defines the input parameters for the import_image tool.
type ImportImageInput struct {
	SpritePath  string          `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	ImagePath   string          `json:"image_path" jsonschema:"Path to image file to import"`
	LayerName   string          `json:"layer_name" jsonschema:"Layer name for imported image"`
	FrameNumber int             `json:"frame_number" jsonschema:"Frame number to place image (1-based)"`
	Position    *aseprite.Point `json:"position,omitempty" jsonschema:"Position to place image (optional defaults to 0,0)"`
}

// ImportImageOutput defines the output for the import_image tool.
type ImportImageOutput struct {
	Success bool `json:"success" jsonschema:"Import success status"`
}

// SaveAsInput defines the input parameters for the save_as tool.
type SaveAsInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	OutputPath string `json:"output_path" jsonschema:"New .aseprite file path"`
}

// SaveAsOutput defines the output for the save_as tool.
type SaveAsOutput struct {
	Success  bool   `json:"success" jsonschema:"Save success status"`
	FilePath string `json:"file_path" jsonschema:"Path to saved file"`
}

// RegisterExportTools registers all export and import tools with the MCP server.
//
// Registers the following tools:
//   - export_sprite: Export individual frames to PNG, GIF, JPG, or BMP
//   - export_spritesheet: Export all frames as a spritesheet (horizontal/vertical/grid layout)
//   - import_image: Import images as new layers or cels
//   - save_as: Save sprite to a different path
//
// Export tools support multiple output formats and frame selection.
// Import tools can create new layers or merge into existing ones.
func RegisterExportTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register export_sprite tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "export_sprite",
			Description: "Export sprite to common image formats (PNG, GIF, JPG, BMP).",
		},
		maybeWrapWithTiming("export_sprite", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input ExportSpriteInput) (*mcp.CallToolResult, *ExportSpriteOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("export_sprite tool called", "sprite_path", input.SpritePath, "output_path", input.OutputPath, "format", input.Format, "frame_number", input.FrameNumber)

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
				opLogger.Error("Failed to export sprite", "error", err)
				return nil, nil, fmt.Errorf("failed to export sprite: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Exported successfully") {
				opLogger.Warning("Unexpected output from export_sprite", "output", output)
			}

			// Get file size
			fileInfo, err := os.Stat(input.OutputPath)
			if err != nil {
				opLogger.Warning("Failed to stat exported file", "error", err)
				return nil, &ExportSpriteOutput{
					ExportedPath: input.OutputPath,
					FileSize:     0,
				}, nil
			}

			opLogger.Information("Sprite exported successfully", "sprite", input.SpritePath, "output", input.OutputPath, "format", format, "frame", input.FrameNumber, "size", fileInfo.Size())

			return nil, &ExportSpriteOutput{
				ExportedPath: input.OutputPath,
				FileSize:     fileInfo.Size(),
			}, nil
		}),
	)

	// Register export_spritesheet tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "export_spritesheet",
			Description: "Export animation frames as spritesheet with layout options.",
		},
		maybeWrapWithTiming("export_spritesheet", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input ExportSpritesheetInput) (*mcp.CallToolResult, *ExportSpritesheetOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("export_spritesheet tool called", "sprite_path", input.SpritePath, "output_path", input.OutputPath, "layout", input.Layout)

			// Validate inputs
			if input.OutputPath == "" {
				return nil, nil, fmt.Errorf("output_path cannot be empty")
			}

			if input.SpritePath == "" {
				return nil, nil, fmt.Errorf("sprite_path cannot be empty")
			}

			// Validate layout
			validLayouts := map[string]bool{
				"horizontal": true,
				"vertical":   true,
				"rows":       true,
				"columns":    true,
				"packed":     true,
			}
			layout := input.Layout
			if layout == "" {
				layout = "horizontal"
			}
			if !validLayouts[layout] {
				return nil, nil, fmt.Errorf("invalid layout: %s (valid: horizontal, vertical, rows, columns, packed)", input.Layout)
			}

			// Validate padding
			if input.Padding < 0 || input.Padding > 100 {
				return nil, nil, fmt.Errorf("padding must be between 0 and 100, got %d", input.Padding)
			}

			// Ensure output directory exists
			outputDir := filepath.Dir(input.OutputPath)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
			}

			// Generate Lua script
			script := gen.ExportSpritesheet(input.OutputPath, layout, input.Padding, input.IncludeJSON)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to export spritesheet", "error", err)
				return nil, nil, fmt.Errorf("failed to export spritesheet: %w", err)
			}

			// Parse JSON output
			var result ExportSpritesheetOutput
			if err := parseJSON(output, &result); err != nil {
				opLogger.Error("Failed to parse spritesheet output", "error", err, "output", output)
				return nil, nil, fmt.Errorf("failed to parse output: %w", err)
			}

			opLogger.Information("Spritesheet exported successfully", "sprite", input.SpritePath, "output", result.SpritesheetPath, "frames", result.FrameCount)

			return nil, &result, nil
		}),
	)

	// Register import_image tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "import_image",
			Description: "Import an external image file as a layer in the sprite.",
		},
		maybeWrapWithTiming("import_image", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input ImportImageInput) (*mcp.CallToolResult, *ImportImageOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("import_image tool called", "sprite_path", input.SpritePath, "image_path", input.ImagePath, "layer", input.LayerName)

			// Validate inputs
			if input.SpritePath == "" {
				return nil, nil, fmt.Errorf("sprite_path cannot be empty")
			}

			if input.ImagePath == "" {
				return nil, nil, fmt.Errorf("image_path cannot be empty")
			}

			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name cannot be empty")
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be >= 1, got %d", input.FrameNumber)
			}

			// Check if image file exists
			if _, err := os.Stat(input.ImagePath); os.IsNotExist(err) {
				return nil, nil, fmt.Errorf("image file not found: %s", input.ImagePath)
			}

			// Extract position if provided
			var x, y *int
			if input.Position != nil {
				x = &input.Position.X
				y = &input.Position.Y
			}

			// Generate Lua script
			script := gen.ImportImage(input.ImagePath, input.LayerName, input.FrameNumber, x, y)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to import image", "error", err)
				return nil, nil, fmt.Errorf("failed to import image: %w", err)
			}

			// Check for success message
			if !strings.Contains(output, "Image imported successfully") {
				opLogger.Warning("Unexpected output from import_image", "output", output)
			}

			opLogger.Information("Image imported successfully", "sprite", input.SpritePath, "image", input.ImagePath, "layer", input.LayerName)

			return nil, &ImportImageOutput{Success: true}, nil
		}),
	)

	// Register save_as tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "save_as",
			Description: "Save sprite to a new .aseprite file path.",
		},
		maybeWrapWithTiming("save_as", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input SaveAsInput) (*mcp.CallToolResult, *SaveAsOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("save_as tool called", "sprite_path", input.SpritePath, "output_path", input.OutputPath)

			// Validate inputs
			if input.SpritePath == "" {
				return nil, nil, fmt.Errorf("sprite_path cannot be empty")
			}

			if input.OutputPath == "" {
				return nil, nil, fmt.Errorf("output_path cannot be empty")
			}

			// Ensure output ends with .aseprite
			if !strings.HasSuffix(input.OutputPath, ".aseprite") && !strings.HasSuffix(input.OutputPath, ".ase") {
				return nil, nil, fmt.Errorf("output_path must have .aseprite or .ase extension")
			}

			// Ensure output directory exists
			outputDir := filepath.Dir(input.OutputPath)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
			}

			// Generate Lua script
			script := gen.SaveAs(input.OutputPath)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to save sprite", "error", err)
				return nil, nil, fmt.Errorf("failed to save sprite: %w", err)
			}

			// Parse JSON output
			var result SaveAsOutput
			if err := parseJSON(output, &result); err != nil {
				opLogger.Error("Failed to parse save_as output", "error", err, "output", output)
				return nil, nil, fmt.Errorf("failed to parse output: %w", err)
			}

			opLogger.Information("Sprite saved successfully", "sprite", input.SpritePath, "new_path", result.FilePath)

			return nil, &result, nil
		}),
	)
}
