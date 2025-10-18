package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
	"github.com/willibrandon/pixel-mcp/pkg/config"
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

// FlipSpriteInput defines the input parameters for the flip_sprite tool.
type FlipSpriteInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	Direction  string `json:"direction" jsonschema:"Flip direction: horizontal or vertical"`
	Target     string `json:"target" jsonschema:"What to flip: sprite, layer, or cel (default: sprite)"`
}

// FlipSpriteOutput defines the output for the flip_sprite tool.
type FlipSpriteOutput struct {
	Success bool `json:"success"`
}

// RotateSpriteInput defines the input parameters for the rotate_sprite tool.
type RotateSpriteInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	Angle      int    `json:"angle" jsonschema:"Rotation angle: 90, 180, or 270 degrees"`
	Target     string `json:"target" jsonschema:"What to rotate: sprite, layer, or cel (default: sprite)"`
}

// RotateSpriteOutput defines the output for the rotate_sprite tool.
type RotateSpriteOutput struct {
	Success bool `json:"success"`
}

// ScaleSpriteInput defines the input parameters for the scale_sprite tool.
type ScaleSpriteInput struct {
	SpritePath string  `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	ScaleX     float64 `json:"scale_x" jsonschema:"Horizontal scale factor (0.01 to 100.0)"`
	ScaleY     float64 `json:"scale_y" jsonschema:"Vertical scale factor (0.01 to 100.0)"`
	Algorithm  string  `json:"algorithm" jsonschema:"Scaling algorithm: nearest, bilinear, or rotsprite (default: nearest)"`
}

// ScaleSpriteOutput defines the output for the scale_sprite tool.
type ScaleSpriteOutput struct {
	Success   bool `json:"success"`
	NewWidth  int  `json:"new_width" jsonschema:"New sprite width after scaling"`
	NewHeight int  `json:"new_height" jsonschema:"New sprite height after scaling"`
}

// CropSpriteInput defines the input parameters for the crop_sprite tool.
type CropSpriteInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	X          int    `json:"x" jsonschema:"Crop region X coordinate"`
	Y          int    `json:"y" jsonschema:"Crop region Y coordinate"`
	Width      int    `json:"width" jsonschema:"Crop region width (must be positive)"`
	Height     int    `json:"height" jsonschema:"Crop region height (must be positive)"`
}

// CropSpriteOutput defines the output for the crop_sprite tool.
type CropSpriteOutput struct {
	Success bool `json:"success"`
}

// ResizeCanvasInput defines the input parameters for the resize_canvas tool.
type ResizeCanvasInput struct {
	SpritePath string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	Width      int    `json:"width" jsonschema:"New canvas width (1-65535)"`
	Height     int    `json:"height" jsonschema:"New canvas height (1-65535)"`
	Anchor     string `json:"anchor" jsonschema:"Anchor position: center, top_left, top_right, bottom_left, or bottom_right (default: center)"`
}

// ResizeCanvasOutput defines the output for the resize_canvas tool.
type ResizeCanvasOutput struct {
	Success bool `json:"success"`
}

// ApplyOutlineInput defines the input parameters for the apply_outline tool.
type ApplyOutlineInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to apply outline to"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number (1-based index)"`
	Color       string `json:"color" jsonschema:"Outline color in hex format (#RRGGBB or #RRGGBBAA)"`
	Thickness   int    `json:"thickness" jsonschema:"Outline thickness in pixels (1-10)"`
}

// ApplyOutlineOutput defines the output for the apply_outline tool.
type ApplyOutlineOutput struct {
	Success bool `json:"success"`
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
		maybeWrapWithTiming("downsample_image", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input DownsampleImageInput) (*mcp.CallToolResult, *DownsampleImageOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("downsample_image tool called", "source", input.SourcePath, "target_width", input.TargetWidth, "target_height", input.TargetHeight)

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
				opLogger.Error("Failed to get source image info", "error", err)
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
					opLogger.Error("Failed to parse source dimensions", "error", err, "output", infoOutput)
					return nil, nil, fmt.Errorf("failed to parse source dimensions: %w", err)
				}
			}

			opLogger.Information("Source image dimensions", "width", sourceWidth, "height", sourceHeight)

			// Generate downsampling script
			script := gen.DownsampleImage(input.SourcePath, outputPath, input.TargetWidth, input.TargetHeight)

			// Execute Lua script
			output, err := client.ExecuteLua(ctx, script, "")
			if err != nil {
				opLogger.Error("Failed to downsample image", "error", err)
				return nil, nil, fmt.Errorf("failed to downsample image: %w", err)
			}

			// Parse output path
			resultPath := strings.TrimSpace(output)

			opLogger.Information("Image downsampled successfully",
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
		}),
	)

	// Register flip_sprite tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "flip_sprite",
			Description: "Flip a sprite, layer, or cel horizontally or vertically. This operation mirrors the image content along the specified axis.",
		},
		maybeWrapWithTiming("flip_sprite", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input FlipSpriteInput) (*mcp.CallToolResult, *FlipSpriteOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("flip_sprite tool called", "sprite", input.SpritePath, "direction", input.Direction, "target", input.Target)

			// Validate direction
			if input.Direction != "horizontal" && input.Direction != "vertical" {
				return nil, nil, fmt.Errorf("direction must be 'horizontal' or 'vertical', got: %s", input.Direction)
			}

			// Validate target
			if input.Target == "" {
				input.Target = "sprite"
			}
			if input.Target != "sprite" && input.Target != "layer" && input.Target != "cel" {
				return nil, nil, fmt.Errorf("target must be 'sprite', 'layer', or 'cel', got: %s", input.Target)
			}

			// Generate Lua script
			script := gen.FlipSprite(input.Direction, input.Target)

			// Execute Lua script
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to flip sprite", "error", err)
				return nil, nil, fmt.Errorf("failed to flip sprite: %w", err)
			}

			opLogger.Information("Sprite flipped successfully", "direction", input.Direction, "target", input.Target)

			return nil, &FlipSpriteOutput{Success: true}, nil
		}),
	)

	// Register rotate_sprite tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "rotate_sprite",
			Description: "Rotate a sprite, layer, or cel by 90, 180, or 270 degrees clockwise.",
		},
		maybeWrapWithTiming("rotate_sprite", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input RotateSpriteInput) (*mcp.CallToolResult, *RotateSpriteOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("rotate_sprite tool called", "sprite", input.SpritePath, "angle", input.Angle, "target", input.Target)

			// Validate angle
			if input.Angle != 90 && input.Angle != 180 && input.Angle != 270 {
				return nil, nil, fmt.Errorf("angle must be 90, 180, or 270, got: %d", input.Angle)
			}

			// Validate target
			if input.Target == "" {
				input.Target = "sprite"
			}
			if input.Target != "sprite" && input.Target != "layer" && input.Target != "cel" {
				return nil, nil, fmt.Errorf("target must be 'sprite', 'layer', or 'cel', got: %s", input.Target)
			}

			// Generate Lua script
			script := gen.RotateSprite(input.Angle, input.Target)

			// Execute Lua script
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to rotate sprite", "error", err)
				return nil, nil, fmt.Errorf("failed to rotate sprite: %w", err)
			}

			opLogger.Information("Sprite rotated successfully", "angle", input.Angle, "target", input.Target)

			return nil, &RotateSpriteOutput{Success: true}, nil
		}),
	)

	// Register scale_sprite tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "scale_sprite",
			Description: "Scale a sprite by specified X and Y factors using a chosen algorithm (nearest, bilinear, or rotsprite). Returns the new dimensions.",
		},
		maybeWrapWithTiming("scale_sprite", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input ScaleSpriteInput) (*mcp.CallToolResult, *ScaleSpriteOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("scale_sprite tool called", "sprite", input.SpritePath, "scale_x", input.ScaleX, "scale_y", input.ScaleY, "algorithm", input.Algorithm)

			// Validate scale factors
			if input.ScaleX < 0.01 || input.ScaleX > 100.0 {
				return nil, nil, fmt.Errorf("scale_x must be between 0.01 and 100.0, got: %.3f", input.ScaleX)
			}
			if input.ScaleY < 0.01 || input.ScaleY > 100.0 {
				return nil, nil, fmt.Errorf("scale_y must be between 0.01 and 100.0, got: %.3f", input.ScaleY)
			}

			// Validate algorithm
			if input.Algorithm == "" {
				input.Algorithm = "nearest"
			}
			if input.Algorithm != "nearest" && input.Algorithm != "bilinear" && input.Algorithm != "rotsprite" {
				return nil, nil, fmt.Errorf("algorithm must be 'nearest', 'bilinear', or 'rotsprite', got: %s", input.Algorithm)
			}

			// Generate Lua script
			script := gen.ScaleSprite(input.ScaleX, input.ScaleY, input.Algorithm)

			// Execute Lua script
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to scale sprite", "error", err)
				return nil, nil, fmt.Errorf("failed to scale sprite: %w", err)
			}

			// Parse JSON output
			var result ScaleSpriteOutput
			if err := parseJSON(output, &result); err != nil {
				return nil, nil, fmt.Errorf("failed to parse scale result: %w", err)
			}

			opLogger.Information("Sprite scaled successfully",
				"scale_x", input.ScaleX,
				"scale_y", input.ScaleY,
				"algorithm", input.Algorithm,
				"new_size", fmt.Sprintf("%dx%d", result.NewWidth, result.NewHeight))

			return nil, &result, nil
		}),
	)

	// Register crop_sprite tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "crop_sprite",
			Description: "Crop a sprite to a specified rectangular region. The crop bounds must be within the sprite dimensions.",
		},
		maybeWrapWithTiming("crop_sprite", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input CropSpriteInput) (*mcp.CallToolResult, *CropSpriteOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("crop_sprite tool called", "sprite", input.SpritePath, "bounds", fmt.Sprintf("%d,%d,%dx%d", input.X, input.Y, input.Width, input.Height))

			// Validate crop bounds
			if input.X < 0 || input.Y < 0 {
				return nil, nil, fmt.Errorf("crop position must be non-negative, got: x=%d, y=%d", input.X, input.Y)
			}
			if input.Width <= 0 || input.Height <= 0 {
				return nil, nil, fmt.Errorf("crop dimensions must be positive, got: width=%d, height=%d", input.Width, input.Height)
			}

			// Generate Lua script (validation of bounds against sprite dimensions happens in Lua)
			script := gen.CropSprite(input.X, input.Y, input.Width, input.Height)

			// Execute Lua script
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to crop sprite", "error", err)
				return nil, nil, fmt.Errorf("failed to crop sprite: %w", err)
			}

			opLogger.Information("Sprite cropped successfully", "bounds", fmt.Sprintf("%d,%d,%dx%d", input.X, input.Y, input.Width, input.Height))

			return nil, &CropSpriteOutput{Success: true}, nil
		}),
	)

	// Register resize_canvas tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "resize_canvas",
			Description: "Resize the canvas without scaling content. Content is positioned according to the anchor point (center, top_left, top_right, bottom_left, or bottom_right).",
		},
		maybeWrapWithTiming("resize_canvas", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input ResizeCanvasInput) (*mcp.CallToolResult, *ResizeCanvasOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("resize_canvas tool called", "sprite", input.SpritePath, "size", fmt.Sprintf("%dx%d", input.Width, input.Height), "anchor", input.Anchor)

			// Validate dimensions
			if input.Width < 1 || input.Width > 65535 {
				return nil, nil, fmt.Errorf("width must be between 1 and 65535, got: %d", input.Width)
			}
			if input.Height < 1 || input.Height > 65535 {
				return nil, nil, fmt.Errorf("height must be between 1 and 65535, got: %d", input.Height)
			}

			// Validate anchor
			if input.Anchor == "" {
				input.Anchor = "center"
			}
			validAnchors := map[string]bool{
				"center":       true,
				"top_left":     true,
				"top_right":    true,
				"bottom_left":  true,
				"bottom_right": true,
			}
			if !validAnchors[input.Anchor] {
				return nil, nil, fmt.Errorf("anchor must be 'center', 'top_left', 'top_right', 'bottom_left', or 'bottom_right', got: %s", input.Anchor)
			}

			// Generate Lua script
			script := gen.ResizeCanvas(input.Width, input.Height, input.Anchor)

			// Execute Lua script
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to resize canvas", "error", err)
				return nil, nil, fmt.Errorf("failed to resize canvas: %w", err)
			}

			opLogger.Information("Canvas resized successfully", "size", fmt.Sprintf("%dx%d", input.Width, input.Height), "anchor", input.Anchor)

			return nil, &ResizeCanvasOutput{Success: true}, nil
		}),
	)

	// Register apply_outline tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "apply_outline",
			Description: "Apply an outline effect to a layer at a specified frame. The outline is drawn around non-transparent pixels with configurable color and thickness.",
		},
		maybeWrapWithTiming("apply_outline", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input ApplyOutlineInput) (*mcp.CallToolResult, *ApplyOutlineOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("apply_outline tool called", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "color", input.Color, "thickness", input.Thickness)

			// Validate frame number
			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be >= 1, got: %d", input.FrameNumber)
			}

			// Validate thickness
			if input.Thickness < 1 || input.Thickness > 10 {
				return nil, nil, fmt.Errorf("thickness must be between 1 and 10, got: %d", input.Thickness)
			}

			// Parse color
			var color aseprite.Color
			if err := color.FromHex(input.Color); err != nil {
				return nil, nil, fmt.Errorf("invalid color format: %w", err)
			}

			// Generate Lua script
			script := gen.ApplyOutline(input.LayerName, input.FrameNumber, color, input.Thickness)

			// Execute Lua script
			_, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to apply outline", "error", err)
				return nil, nil, fmt.Errorf("failed to apply outline: %w", err)
			}

			opLogger.Information("Outline applied successfully", "layer", input.LayerName, "frame", input.FrameNumber, "thickness", input.Thickness)

			return nil, &ApplyOutlineOutput{Success: true}, nil
		}),
	)
}
