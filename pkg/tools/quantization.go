package tools

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
	"github.com/willibrandon/pixel-mcp/pkg/config"
)

// QuantizePaletteInput defines the input parameters for the quantize_palette tool.
type QuantizePaletteInput struct {
	SpritePath           string `json:"sprite_path" jsonschema:"Path to source .aseprite file"`
	TargetColors         int    `json:"target_colors" jsonschema:"Target palette size (2-256)"`
	Algorithm            string `json:"algorithm" jsonschema:"Quantization algorithm: median_cut (default), kmeans, or octree"`
	Dither               bool   `json:"dither" jsonschema:"Apply Floyd-Steinberg dithering during quantization (default: false)"`
	PreserveTransparency *bool  `json:"preserve_transparency,omitempty" jsonschema:"Keep transparent pixels transparent (default: true)"`
	ConvertToIndexed     *bool  `json:"convert_to_indexed,omitempty" jsonschema:"Convert sprite to indexed color mode (default: true)"`
}

// QuantizePaletteOutput defines the output for the quantize_palette tool.
type QuantizePaletteOutput struct {
	Success         bool     `json:"success" jsonschema:"Whether the operation succeeded"`
	OriginalColors  int      `json:"original_colors" jsonschema:"Number of unique colors in original sprite"`
	QuantizedColors int      `json:"quantized_colors" jsonschema:"Number of colors in quantized palette"`
	ColorMode       string   `json:"color_mode" jsonschema:"Color mode after quantization (indexed or rgb)"`
	Palette         []string `json:"palette" jsonschema:"Array of hex colors in the quantized palette"`
	AlgorithmUsed   string   `json:"algorithm_used" jsonschema:"Quantization algorithm that was used"`
}

// RegisterQuantizationTools registers the quantize_palette tool with the MCP server.
func RegisterQuantizationTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "quantize_palette",
			Description: "Automatically reduce sprite colors using industry-standard quantization algorithms. Supports three algorithms: median_cut (fast, balanced quality), kmeans (highest quality, slower), octree (very fast, good for photos). Can apply Floyd-Steinberg dithering for smoother gradients. Optionally converts to indexed color mode for true palette constraint or keeps RGB mode for flexible multi-pass workflows.",
		},
		maybeWrapWithTiming("quantize_palette", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input QuantizePaletteInput) (*mcp.CallToolResult, *QuantizePaletteOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("quantize_palette tool called",
				"sprite", input.SpritePath,
				"target_colors", input.TargetColors,
				"algorithm", input.Algorithm)

			// Set defaults
			if input.Algorithm == "" {
				input.Algorithm = "median_cut"
			}
			if input.PreserveTransparency == nil {
				defaultTrue := true
				input.PreserveTransparency = &defaultTrue
			}
			if input.ConvertToIndexed == nil {
				defaultTrue := true
				input.ConvertToIndexed = &defaultTrue
			}

			// Validate inputs
			if input.TargetColors < 2 || input.TargetColors > 256 {
				return nil, nil, fmt.Errorf("target_colors must be between 2 and 256, got %d", input.TargetColors)
			}

			validAlgorithms := map[string]bool{
				"median_cut": true,
				"kmeans":     true,
				"octree":     true,
			}
			if !validAlgorithms[input.Algorithm] {
				return nil, nil, fmt.Errorf("invalid algorithm: %s (must be median_cut, kmeans, or octree)", input.Algorithm)
			}

			// Check sprite file exists
			if _, err := os.Stat(input.SpritePath); os.IsNotExist(err) {
				return nil, nil, fmt.Errorf("sprite file not found: %s", input.SpritePath)
			}

			// Step 1: Export sprite to temporary PNG for analysis
			tempDir, err := os.MkdirTemp("", "pixel-mcp-quantize-*")
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create temp directory: %w", err)
			}
			defer os.RemoveAll(tempDir)

			tempPNG := filepath.Join(tempDir, "sprite.png")

			// Export sprite to PNG using gen.ExportSprite for consistency
			exportScript := gen.ExportSprite(tempPNG, 0)

			_, err = client.ExecuteLua(ctx, exportScript, input.SpritePath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to export sprite to PNG: %w", err)
			}

			// Step 2: Load PNG and perform quantization in Go
			imgFile, err := os.Open(tempPNG)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to open exported PNG: %w", err)
			}
			defer imgFile.Close()

			img, err := png.Decode(imgFile)
			if err != nil {
				// Try to decode as any image format
				imgFile.Seek(0, 0)
				img, _, err = image.Decode(imgFile)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to decode image: %w", err)
				}
			}

			// Perform quantization
			palette, originalColors, err := aseprite.QuantizePalette(
				img,
				input.TargetColors,
				input.Algorithm,
				*input.PreserveTransparency,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("quantization failed: %w", err)
			}

			opLogger.Information("Quantization completed",
				"original_colors", originalColors,
				"quantized_colors", len(palette),
				"algorithm", input.Algorithm)

			// Step 3: If dithering is requested, remap pixels to quantized palette with dithering
			if input.Dither {
				// Convert hex palette to color.Color slice
				paletteColors := make([]color.Color, len(palette))
				for i, hexColor := range palette {
					c, err := colorful.Hex(hexColor)
					if err != nil {
						return nil, nil, fmt.Errorf("invalid palette color %s: %w", hexColor, err)
					}
					r, g, b := c.RGB255()
					paletteColors[i] = color.RGBA{R: r, G: g, B: b, A: 255}
				}

				// Remap image with dithering
				ditheredImg := aseprite.RemapPixelsWithDithering(img, paletteColors, true)

				// Save dithered image to temp file
				ditheredPNG := filepath.Join(tempDir, "dithered.png")
				ditheredFile, err := os.Create(ditheredPNG)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to create dithered PNG: %w", err)
				}
				defer ditheredFile.Close()

				if err := png.Encode(ditheredFile, ditheredImg); err != nil {
					return nil, nil, fmt.Errorf("failed to encode dithered PNG: %w", err)
				}

				// Replace sprite content with dithered image
				replaceScript := gen.ReplaceWithImage(ditheredPNG)
				_, err = client.ExecuteLua(ctx, replaceScript, input.SpritePath)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to replace sprite with dithered image: %w", err)
				}

				opLogger.Information("Dithering applied successfully",
					"sprite", input.SpritePath)
			}
			// Step 4: Generate and execute Lua script to apply quantized palette
			applyScript := gen.ApplyQuantizedPalette(
				palette,
				originalColors,
				input.Algorithm,
				*input.ConvertToIndexed,
				input.Dither,
			)

			output, err := client.ExecuteLua(ctx, applyScript, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to apply quantized palette", "error", err)
				return nil, nil, fmt.Errorf("failed to apply quantized palette: %w", err)
			}

			// Parse JSON output from Lua
			var result QuantizePaletteOutput
			if err := parseJSON(output, &result); err != nil {
				return nil, nil, fmt.Errorf("failed to parse quantization output: %w", err)
			}

			opLogger.Information("Palette quantized and applied successfully",
				"sprite", input.SpritePath,
				"original_colors", result.OriginalColors,
				"quantized_colors", result.QuantizedColors,
				"color_mode", result.ColorMode,
				"algorithm", result.AlgorithmUsed)

			return nil, &result, nil
		}),
	)
}
