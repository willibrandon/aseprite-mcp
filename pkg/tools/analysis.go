package tools

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/mtlog/core"
)

// AnalyzeReferenceInput defines the input parameters for the analyze_reference tool.
type AnalyzeReferenceInput struct {
	ReferencePath    string `json:"reference_path" jsonschema:"Path to reference image (.jpg, .png, .gif, .bmp, .aseprite)"`
	TargetWidth      int    `json:"target_width" jsonschema:"Pixel art target width (1-65535)"`
	TargetHeight     int    `json:"target_height" jsonschema:"Pixel art target height (1-65535)"`
	PaletteSize      int    `json:"palette_size,omitempty" jsonschema:"Number of colors to extract (5-32, default: 16)"`
	BrightnessLevels int    `json:"brightness_levels,omitempty" jsonschema:"Quantize brightness into N levels (2-10, default: 5)"`
	EdgeThreshold    int    `json:"edge_threshold,omitempty" jsonschema:"Edge detection sensitivity (0-255, default: 30)"`
}

// AnalyzeReferenceOutput defines the output for the analyze_reference tool.
type AnalyzeReferenceOutput struct {
	Palette        []aseprite.PaletteColor `json:"palette"`
	BrightnessMap  *aseprite.BrightnessMap `json:"brightness_map"`
	EdgeMap        *aseprite.EdgeMap       `json:"edge_map"`
	Composition    *aseprite.Composition   `json:"composition"`
	DitheringZones []DitheringZone         `json:"dithering_zones"`
	Metadata       *AnalysisMetadata       `json:"metadata"`
}

// DitheringZone represents a suggested area for dithering.
type DitheringZone struct {
	Region  aseprite.Region `json:"region"`
	Type    string          `json:"type"`    // "gradient", "texture"
	Colors  []string        `json:"colors"`  // Suggested palette colors for dithering
	Pattern string          `json:"pattern"` // "bayer_4x4", "checkerboard", etc.
	Reason  string          `json:"reason"`  // Explanation for suggestion
}

// AnalysisMetadata contains metadata about the analysis.
type AnalysisMetadata struct {
	SourceDimensions Dimensions `json:"source_dimensions"`
	TargetDimensions Dimensions `json:"target_dimensions"`
	ScaleFactor      float64    `json:"scale_factor"`
	DominantHue      float64    `json:"dominant_hue"`   // 0-360 degrees
	ColorHarmony     string     `json:"color_harmony"`  // "complementary", "analogous", etc.
	ContrastRatio    string     `json:"contrast_ratio"` // "low", "medium", "high"
}

// Dimensions represents width and height.
type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// RegisterAnalysisTools registers all image analysis tools with the MCP server.
func RegisterAnalysisTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register analyze_reference tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "analyze_reference",
			Description: "Extract structured data from reference images to guide pixel art creation. Performs k-means palette extraction, brightness/edge detection, and composition analysis. Returns palette sorted by hue/lightness, brightness map with quantized levels, edge map with major contours, composition guides (rule of thirds, focal points), and suggested dithering zones.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeReferenceInput) (*mcp.CallToolResult, *AnalyzeReferenceOutput, error) {
			logger.Debug("analyze_reference tool called",
				"reference", input.ReferencePath,
				"target_width", input.TargetWidth,
				"target_height", input.TargetHeight,
				"palette_size", input.PaletteSize)

			// Validate inputs
			if input.TargetWidth < 1 || input.TargetWidth > 65535 {
				return nil, nil, fmt.Errorf("target_width must be between 1 and 65535, got %d", input.TargetWidth)
			}
			if input.TargetHeight < 1 || input.TargetHeight > 65535 {
				return nil, nil, fmt.Errorf("target_height must be between 1 and 65535, got %d", input.TargetHeight)
			}

			// Set defaults
			paletteSize := input.PaletteSize
			if paletteSize == 0 {
				paletteSize = 16
			}
			if paletteSize < 5 || paletteSize > 32 {
				return nil, nil, fmt.Errorf("palette_size must be between 5 and 32, got %d", paletteSize)
			}

			brightnessLevels := input.BrightnessLevels
			if brightnessLevels == 0 {
				brightnessLevels = 5
			}
			if brightnessLevels < 2 || brightnessLevels > 10 {
				return nil, nil, fmt.Errorf("brightness_levels must be between 2 and 10, got %d", brightnessLevels)
			}

			edgeThreshold := input.EdgeThreshold
			if edgeThreshold == 0 {
				edgeThreshold = 30
			}
			if edgeThreshold < 0 || edgeThreshold > 255 {
				return nil, nil, fmt.Errorf("edge_threshold must be between 0 and 255, got %d", edgeThreshold)
			}

			// Load reference image
			file, err := os.Open(input.ReferencePath)
			if err != nil {
				logger.Error("Failed to open reference image", "error", err)
				return nil, nil, fmt.Errorf("failed to open reference image: %w", err)
			}
			defer file.Close()

			img, _, err := image.Decode(file)
			if err != nil {
				logger.Error("Failed to decode reference image", "error", err)
				return nil, nil, fmt.Errorf("failed to decode reference image: %w", err)
			}

			bounds := img.Bounds()
			sourceWidth := bounds.Dx()
			sourceHeight := bounds.Dy()

			logger.Information("Analyzing reference image",
				"source_size", fmt.Sprintf("%dx%d", sourceWidth, sourceHeight),
				"target_size", fmt.Sprintf("%dx%d", input.TargetWidth, input.TargetHeight))

			// Extract palette using k-means clustering
			palette, err := aseprite.ExtractPalette(img, paletteSize)
			if err != nil {
				logger.Error("Failed to extract palette", "error", err)
				return nil, nil, fmt.Errorf("failed to extract palette: %w", err)
			}

			logger.Debug("Palette extracted", "colors", len(palette))

			// Generate brightness map
			brightnessMap, err := aseprite.GenerateBrightnessMap(img, input.TargetWidth, input.TargetHeight, brightnessLevels)
			if err != nil {
				logger.Error("Failed to generate brightness map", "error", err)
				return nil, nil, fmt.Errorf("failed to generate brightness map: %w", err)
			}

			// Detect edges
			edgeMap, err := aseprite.DetectEdges(img, edgeThreshold)
			if err != nil {
				logger.Error("Failed to detect edges", "error", err)
				return nil, nil, fmt.Errorf("failed to detect edges: %w", err)
			}

			logger.Debug("Edge detection complete", "major_edges", len(edgeMap.MajorEdges))

			// Analyze composition
			composition, err := aseprite.AnalyzeComposition(img, edgeMap)
			if err != nil {
				logger.Error("Failed to analyze composition", "error", err)
				return nil, nil, fmt.Errorf("failed to analyze composition: %w", err)
			}

			// Suggest dithering zones
			ditheringZones := suggestDitheringZones(palette, brightnessMap, edgeMap)

			// Calculate metadata
			metadata := calculateMetadata(palette, sourceWidth, sourceHeight, input.TargetWidth, input.TargetHeight)

			logger.Information("Reference analysis complete",
				"palette_colors", len(palette),
				"focal_points", len(composition.FocalPoints),
				"dithering_zones", len(ditheringZones),
				"dominant_hue", metadata.DominantHue,
				"color_harmony", metadata.ColorHarmony)

			return nil, &AnalyzeReferenceOutput{
				Palette:        palette,
				BrightnessMap:  brightnessMap,
				EdgeMap:        edgeMap,
				Composition:    composition,
				DitheringZones: ditheringZones,
				Metadata:       metadata,
			}, nil
		},
	)
}

// suggestDitheringZones analyzes the image and suggests areas that would benefit from dithering.
func suggestDitheringZones(palette []aseprite.PaletteColor, brightnessMap *aseprite.BrightnessMap, edgeMap *aseprite.EdgeMap) []DitheringZone {
	zones := make([]DitheringZone, 0)

	// Analyze brightness map for gradients
	// Look for horizontal or vertical gradients
	gridHeight := len(brightnessMap.Grid)
	if gridHeight == 0 {
		return zones
	}
	gridWidth := len(brightnessMap.Grid[0])

	// Check for horizontal gradients (sky, horizons)
	for y := 0; y < gridHeight-2; y++ {
		gradientStart := -1
		for x := 0; x < gridWidth; x++ {
			if x == 0 {
				continue
			}

			// Check if brightness increases across this row
			diff := brightnessMap.Grid[y][x] - brightnessMap.Grid[y][x-1]
			if diff > 0 && gradientStart == -1 {
				gradientStart = x - 1
			} else if diff <= 0 && gradientStart != -1 {
				// End of gradient
				if x-gradientStart >= 3 { // Minimum gradient width
					// Find suitable palette colors for this brightness range
					startLevel := brightnessMap.Grid[y][gradientStart]
					endLevel := brightnessMap.Grid[y][x-1]
					colors := findColorsForBrightnessRange(palette, startLevel, endLevel, len(brightnessMap.Legend))

					if len(colors) >= 2 {
						zones = append(zones, DitheringZone{
							Region: aseprite.Region{
								X:      gradientStart,
								Y:      y,
								Width:  x - gradientStart,
								Height: 1,
							},
							Type:    "gradient",
							Colors:  colors[:2], // Use two colors for dithering
							Pattern: "bayer_4x4",
							Reason:  "horizontal brightness gradient detected",
						})
					}
				}
				gradientStart = -1
			}
		}
	}

	// Check for uniform areas that need texture
	for y := 0; y < gridHeight-3; y++ {
		for x := 0; x < gridWidth-3; x++ {
			// Check 3x3 region for uniformity
			baseLevel := brightnessMap.Grid[y][x]
			uniform := true
			for dy := 0; dy < 3 && y+dy < gridHeight; dy++ {
				for dx := 0; dx < 3 && x+dx < gridWidth; dx++ {
					if brightnessMap.Grid[y+dy][x+dx] != baseLevel {
						uniform = false
						break
					}
				}
				if !uniform {
					break
				}
			}

			// Check if area has low edge density (flat area)
			edgeCount := 0
			for dy := 0; dy < 3 && y+dy < len(edgeMap.Grid); dy++ {
				for dx := 0; dx < 3 && x+dx < len(edgeMap.Grid[y+dy]); dx++ {
					if edgeMap.Grid[y+dy][x+dx] == 1 {
						edgeCount++
					}
				}
			}

			if uniform && edgeCount < 2 { // Low edge density
				colors := findColorsForBrightnessRange(palette, baseLevel, baseLevel, len(brightnessMap.Legend))
				if len(colors) >= 2 {
					zones = append(zones, DitheringZone{
						Region: aseprite.Region{
							X:      x,
							Y:      y,
							Width:  3,
							Height: 3,
						},
						Type:    "texture",
						Colors:  colors[:2],
						Pattern: "checkerboard",
						Reason:  "uniform area needs texture",
					})
				}
			}
		}
	}

	// Limit to most significant zones
	if len(zones) > 5 {
		zones = zones[:5]
	}

	return zones
}

// findColorsForBrightnessRange finds palette colors matching a brightness range.
func findColorsForBrightnessRange(palette []aseprite.PaletteColor, minLevel, maxLevel, numLevels int) []string {
	colors := make([]string, 0)

	// Map levels to lightness range
	minLightness := float64(minLevel) / float64(numLevels) * 100.0
	maxLightness := float64(maxLevel+1) / float64(numLevels) * 100.0

	// Find palette colors in this lightness range
	for _, pColor := range palette {
		if pColor.Lightness >= minLightness-10 && pColor.Lightness <= maxLightness+10 {
			colors = append(colors, pColor.Color)
		}
	}

	// If no colors found, use darkest and lightest
	if len(colors) == 0 && len(palette) > 0 {
		colors = append(colors, palette[0].Color) // Darkest
		if len(palette) > 1 {
			colors = append(colors, palette[len(palette)-1].Color) // Lightest
		}
	}

	return colors
}

// calculateMetadata calculates analysis metadata.
func calculateMetadata(palette []aseprite.PaletteColor, sourceWidth, sourceHeight, targetWidth, targetHeight int) *AnalysisMetadata {
	// Calculate dominant hue (weighted average)
	totalWeight := 0.0
	weightedHue := 0.0
	for _, c := range palette {
		totalWeight += c.UsagePercent
		weightedHue += c.Hue * c.UsagePercent
	}

	dominantHue := 0.0
	if totalWeight > 0 {
		dominantHue = weightedHue / totalWeight
	}

	// Determine color harmony
	harmony := determineColorHarmony(palette)

	// Calculate contrast ratio
	contrast := "medium"
	if len(palette) > 0 {
		lightestColor := palette[len(palette)-1]
		darkestColor := palette[0]
		lightnessRange := lightestColor.Lightness - darkestColor.Lightness

		if lightnessRange < 30 {
			contrast = "low"
		} else if lightnessRange > 60 {
			contrast = "high"
		}
	}

	scaleFactor := float64(targetWidth) / float64(sourceWidth)

	return &AnalysisMetadata{
		SourceDimensions: Dimensions{Width: sourceWidth, Height: sourceHeight},
		TargetDimensions: Dimensions{Width: targetWidth, Height: targetHeight},
		ScaleFactor:      scaleFactor,
		DominantHue:      dominantHue,
		ColorHarmony:     harmony,
		ContrastRatio:    contrast,
	}
}

// determineColorHarmony analyzes the palette to determine color relationships.
func determineColorHarmony(palette []aseprite.PaletteColor) string {
	if len(palette) < 2 {
		return "monochromatic"
	}

	// Calculate hue spread
	hues := make([]float64, len(palette))
	for i, c := range palette {
		hues[i] = c.Hue
	}

	// Check for complementary (opposite on color wheel, ~180 degrees apart)
	maxHueDiff := 0.0
	for i := 0; i < len(hues)-1; i++ {
		for j := i + 1; j < len(hues); j++ {
			diff := absDiff(hues[i], hues[j])
			if diff > maxHueDiff {
				maxHueDiff = diff
			}
		}
	}

	if maxHueDiff > 150 && maxHueDiff < 210 {
		return "complementary"
	}

	// Check for analogous (adjacent on color wheel, within 60 degrees)
	hueRange := maxHueDiff
	if hueRange < 60 {
		return "analogous"
	}

	// Check for triadic (120 degrees apart)
	// This is simplified - a full check would look for three 120° segments
	if len(palette) >= 3 {
		// Count colors in three 120° segments
		segment1, segment2, segment3 := 0, 0, 0
		for _, h := range hues {
			if h < 120 {
				segment1++
			} else if h < 240 {
				segment2++
			} else {
				segment3++
			}
		}

		if segment1 > 0 && segment2 > 0 && segment3 > 0 {
			return "triadic"
		}
	}

	return "diverse"
}

// absDiff calculates the absolute difference between two hue values (0-360).
func absDiff(a, b float64) float64 {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}

	// Handle wrap-around (e.g., 10° and 350° are actually 20° apart)
	if diff > 180 {
		diff = 360 - diff
	}

	return diff
}
