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

// DownsampleImageInput defines the input parameters for the downsample_image tool.
type DownsampleImageInput struct {
	SourcePath   string `json:"source_path" jsonschema:"Path to source image file (.aseprite, .png, .jpg, .bmp, .gif)"`
	TargetWidth  int    `json:"target_width" jsonschema:"Target width in pixels (1-65535)"`
	TargetHeight int    `json:"target_height" jsonschema:"Target height in pixels (1-65535)"`
	OutputPath   string `json:"output_path,omitempty" jsonschema:"Optional output path for downsampled sprite (defaults to temp directory)"`
}

// DownsampleImageOutput defines the output for the downsample_image tool.
type DownsampleImageOutput struct {
	OutputPath   string `json:"output_path" jsonschema:"Path to the downsampled sprite file"`
	SourceWidth  int    `json:"source_width" jsonschema:"Original source image width"`
	SourceHeight int    `json:"source_height" jsonschema:"Original source image height"`
	TargetWidth  int    `json:"target_width" jsonschema:"Downsampled image width"`
	TargetHeight int    `json:"target_height" jsonschema:"Downsampled image height"`
}

// RegisterTransformTools registers all image transformation tools with the MCP server.
func RegisterTransformTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register downsample_image tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "downsample_image",
			Description: "Downsample an image to smaller dimensions using box filter (area averaging) algorithm. Accepts any image format supported by Aseprite (.aseprite, .png, .jpg, .bmp, .gif) and creates a new downsampled sprite. This is useful for creating pixel art versions of high-resolution images or reducing image size while maintaining quality.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input DownsampleImageInput) (*mcp.CallToolResult, *DownsampleImageOutput, error) {
			logger.Debug("downsample_image tool called", "source", input.SourcePath, "target_width", input.TargetWidth, "target_height", input.TargetHeight)

			// Validate dimensions
			if input.TargetWidth < 1 || input.TargetWidth > 65535 {
				return nil, nil, fmt.Errorf("target_width must be between 1 and 65535, got %d", input.TargetWidth)
			}
			if input.TargetHeight < 1 || input.TargetHeight > 65535 {
				return nil, nil, fmt.Errorf("target_height must be between 1 and 65535, got %d", input.TargetHeight)
			}

			// Determine output path
			outputPath := input.OutputPath
			if outputPath == "" {
				outputPath = filepath.Join(cfg.TempDir, fmt.Sprintf("downsampled-%d.aseprite", time.Now().UnixNano()))
			}

			// First, get source image dimensions using get_sprite_info
			infoScript := gen.GetSpriteInfo()
			infoOutput, err := client.ExecuteLua(ctx, infoScript, input.SourcePath)
			if err != nil {
				logger.Error("Failed to get source image info", "error", err)
				return nil, nil, fmt.Errorf("failed to get source image info: %w", err)
			}

			// Parse dimensions from JSON output (simple extraction)
			var sourceWidth, sourceHeight int
			// The output is JSON like {"width":500,"height":756,...}
			_, err = fmt.Sscanf(infoOutput, `{"width":%d,"height":%d`, &sourceWidth, &sourceHeight)
			if err != nil {
				// Try alternate parsing
				if strings.Contains(infoOutput, `"width":`) {
					parts := strings.Split(infoOutput, `"width":`)
					if len(parts) > 1 {
						_, _ = fmt.Sscanf(parts[1], "%d", &sourceWidth)
					}
					parts = strings.Split(infoOutput, `"height":`)
					if len(parts) > 1 {
						_, _ = fmt.Sscanf(parts[1], "%d", &sourceHeight)
					}
				}
				if sourceWidth == 0 || sourceHeight == 0 {
					logger.Error("Failed to parse source dimensions", "error", err, "output", infoOutput)
					return nil, nil, fmt.Errorf("failed to parse source dimensions: %w", err)
				}
			}

			logger.Information("Source image dimensions", "width", sourceWidth, "height", sourceHeight)

			// Generate downsampling script
			script := gen.DownsampleImage(input.SourcePath, outputPath, input.TargetWidth, input.TargetHeight)

			// Execute Lua script
			output, err := client.ExecuteLua(ctx, script, "")
			if err != nil {
				logger.Error("Failed to downsample image", "error", err)
				return nil, nil, fmt.Errorf("failed to downsample image: %w", err)
			}

			// Parse output path
			resultPath := strings.TrimSpace(output)

			logger.Information("Image downsampled successfully",
				"source", input.SourcePath,
				"output", resultPath,
				"source_size", fmt.Sprintf("%dx%d", sourceWidth, sourceHeight),
				"target_size", fmt.Sprintf("%dx%d", input.TargetWidth, input.TargetHeight))

			return nil, &DownsampleImageOutput{
				OutputPath:   resultPath,
				SourceWidth:  sourceWidth,
				SourceHeight: sourceHeight,
				TargetWidth:  input.TargetWidth,
				TargetHeight: input.TargetHeight,
			}, nil
		},
	)
}
