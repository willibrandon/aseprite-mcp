// Package main demonstrates the quantize_palette tool with all algorithms.
// This example creates a sprite with many colors and reduces them using
// different quantization algorithms (median_cut, kmeans, octree).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/mtlog/sinks"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
	"github.com/willibrandon/pixel-mcp/pkg/config"
)

func main() {
	logger := createLogger()
	if err := run(logger); err != nil {
		logger.Fatal("Application error: {Error}", err)
	}
}

func run(logger core.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	logger.Information("Palette Quantization Example")
	logger.Information("============================")
	logger.Information("")

	// Start MCP server
	serverPath := os.Getenv("ASEPRITE_MCP_PATH")
	if serverPath == "" {
		serverPath = findServerBinary()
		if serverPath == "" {
			return fmt.Errorf("ASEPRITE_MCP_PATH not set and could not find pixel-mcp binary")
		}
	}

	logger.Information("Starting server: {ServerPath}", serverPath)
	cmd := exec.Command(serverPath)

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "quantization-example",
		Version: "1.0.0",
	}, nil)

	logger.Information("Connecting to server...")
	transport := &mcp.CommandTransport{Command: cmd}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer session.Close()

	logger.Information("Connected!")
	logger.Information("")

	// Load config for Aseprite path
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Aseprite client and Lua generator for direct Lua execution
	// cfg.Timeout is already a time.Duration, use it directly
	asepriteClient := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()

	// Create output directory
	outputDir := "examples/sprites"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Create a sprite with many colors (gradient)
	logger.Information("Step 1: Creating 128x128 RGB canvas with gradient...")
	createResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      128,
		"height":     128,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas failed: %w", err)
	}

	var createResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(createResp), &createResult); err != nil {
		return fmt.Errorf("failed to parse create result: %w", err)
	}
	spritePath := createResult.FilePath
	logger.Information("  Created: {SpritePath}", spritePath)

	// Step 2: Draw gradient with many colors
	logger.Information("")
	logger.Information("Step 2: Drawing gradient with ~16,000 unique colors...")

	// Draw horizontal gradient bars (red to yellow to green to cyan to blue)
	// We'll batch pixels to make this much faster
	colors := []struct{ r, g, b int }{
		{255, 0, 0},     // Red
		{255, 128, 0},   // Orange
		{255, 255, 0},   // Yellow
		{128, 255, 0},   // Yellow-Green
		{0, 255, 0},     // Green
		{0, 255, 128},   // Green-Cyan
		{0, 255, 255},   // Cyan
		{0, 128, 255},   // Cyan-Blue
		{0, 0, 255},     // Blue
		{128, 0, 255},   // Blue-Magenta
		{255, 0, 255},   // Magenta
		{255, 0, 128},   // Magenta-Red
	}

	barHeight := 128 / len(colors)

	// Batch all pixels into chunks to avoid thousands of individual calls
	batchSize := 1000
	allPixels := []map[string]any{}

	for i, startColor := range colors {
		endColor := colors[(i+1)%len(colors)]
		y := i * barHeight

		for x := 0; x < 128; x++ {
			t := float64(x) / 127.0
			r := int(float64(startColor.r)*(1-t) + float64(endColor.r)*t)
			g := int(float64(startColor.g)*(1-t) + float64(endColor.g)*t)
			b := int(float64(startColor.b)*(1-t) + float64(endColor.b)*t)

			color := fmt.Sprintf("#%02X%02X%02X", r, g, b)

			// Add all pixels in this column to the batch
			for dy := 0; dy < barHeight; dy++ {
				allPixels = append(allPixels, map[string]any{
					"x":     x,
					"y":     y + dy,
					"color": color,
				})
			}
		}
	}

	logger.Information("  Drawing {Count} pixels in batches...", len(allPixels))

	// Draw in batches
	for i := 0; i < len(allPixels); i += batchSize {
		end := i + batchSize
		if end > len(allPixels) {
			end = len(allPixels)
		}

		batch := allPixels[i:end]
		if _, err := callTool(ctx, session, "draw_pixels", map[string]any{
			"sprite_path":  spritePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"pixels":       batch,
		}); err != nil {
			return fmt.Errorf("draw_pixels failed: %w", err)
		}

		if (i/batchSize)%5 == 0 {
			logger.Information("    Progress: {Current}/{Total} pixels", end, len(allPixels))
		}
	}
	logger.Information("  Gradient drawn")

	// Export original
	originalPath := filepath.Join(outputDir, "quantization-original.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  spritePath,
		"output_path":  originalPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export original failed: %w", err)
	}
	logger.Information("  Exported original: {Path}", originalPath)

	// Step 3: Test each quantization algorithm
	algorithms := []struct {
		name        string
		targetColors int
		dither      bool
		description string
	}{
		{"median_cut", 16, false, "Median Cut (balanced, no dither)"},
		{"median_cut", 16, true, "Median Cut with Floyd-Steinberg dithering"},
		{"kmeans", 16, false, "K-means clustering (highest quality)"},
		{"octree", 16, false, "Octree quantization (fastest)"},
		{"median_cut", 8, false, "Median Cut to 8 colors"},
		{"median_cut", 4, true, "Median Cut to 4 colors with dithering"},
	}

	for i, algo := range algorithms {
		logger.Information("")
		logger.Information("Step {Step}: Testing {Description}...", i+3, algo.description)

		// Create a copy of the sprite for this algorithm
		copyResp, err := callTool(ctx, session, "create_canvas", map[string]any{
			"width":      128,
			"height":     128,
			"color_mode": "rgb",
		})
		if err != nil {
			return fmt.Errorf("create_canvas for copy failed: %w", err)
		}

		var copyResult struct {
			FilePath string `json:"file_path"`
		}
		if err := json.Unmarshal([]byte(copyResp), &copyResult); err != nil {
			return fmt.Errorf("failed to parse copy result: %w", err)
		}
		copyPath := copyResult.FilePath

		// Copy content from original (use import_image to copy)
		// First export original to temp location
		tempOriginal := filepath.Join(os.TempDir(), "temp-original.png")
		if _, err := callTool(ctx, session, "export_sprite", map[string]any{
			"sprite_path":  spritePath,
			"output_path":  tempOriginal,
			"format":       "png",
			"frame_number": 1,
		}); err != nil {
			return fmt.Errorf("export temp original failed: %w", err)
		}

		// Import into copy sprite
		if _, err := callTool(ctx, session, "import_image", map[string]any{
			"sprite_path":  copyPath,
			"image_path":   tempOriginal,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"position": map[string]any{
				"x": 0,
				"y": 0,
			},
		}); err != nil {
			return fmt.Errorf("import_image failed: %w", err)
		}
		os.Remove(tempOriginal)

		// Apply quantization
		quantizeResp, err := callTool(ctx, session, "quantize_palette", map[string]any{
			"sprite_path":            copyPath,
			"target_colors":          algo.targetColors,
			"algorithm":              algo.name,
			"dither":                 algo.dither,
			"preserve_transparency":  false,
			"convert_to_indexed":     true,
		})
		if err != nil {
			return fmt.Errorf("quantize_palette failed: %w", err)
		}

		var quantizeResult struct {
			Success         bool     `json:"success"`
			OriginalColors  int      `json:"original_colors"`
			QuantizedColors int      `json:"quantized_colors"`
			ColorMode       string   `json:"color_mode"`
			Palette         []string `json:"palette"`
			AlgorithmUsed   string   `json:"algorithm_used"`
		}
		if err := json.Unmarshal([]byte(quantizeResp), &quantizeResult); err != nil {
			return fmt.Errorf("failed to parse quantize result: %w", err)
		}

		logger.Information("  Quantized: {Original} → {Quantized} colors using {Algorithm}",
			quantizeResult.OriginalColors,
			quantizeResult.QuantizedColors,
			quantizeResult.AlgorithmUsed)
		logger.Information("  Color mode: {Mode}", quantizeResult.ColorMode)
		logger.Information("  Palette preview: {Colors}...", quantizeResult.Palette[:min(5, len(quantizeResult.Palette))])

		// Convert back to RGB for export to avoid color mode conversion issues
		convertScript := `
local spr = app.activeSprite
if spr.colorMode == ColorMode.INDEXED then
	app.command.ChangePixelFormat{format="rgb"}
	spr:saveAs(spr.filename)
end
`
		convertCtx, convertCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer convertCancel()
		if _, err := asepriteClient.ExecuteLua(convertCtx, convertScript, copyPath); err != nil {
			return fmt.Errorf("convert to RGB failed: %w", err)
		}

		// Export result
		ditherSuffix := ""
		if algo.dither {
			ditherSuffix = "-dither"
		}
		outputPath := filepath.Join(outputDir, fmt.Sprintf("quantization-%s-%dcolors%s.png",
			algo.name, algo.targetColors, ditherSuffix))
		if _, err := callTool(ctx, session, "export_sprite", map[string]any{
			"sprite_path":  copyPath,
			"output_path":  outputPath,
			"format":       "png",
			"frame_number": 0,
		}); err != nil {
			return fmt.Errorf("export quantized failed: %w", err)
		}
		logger.Information("  Exported: {Path}", outputPath)

		// Clean up copy
		os.Remove(copyPath)
	}

	// Step 9: Compare algorithms side-by-side
	logger.Information("")
	logger.Information("Step {Step}: Creating comparison sprite...", len(algorithms)+3)

	// Create a wide canvas to hold all comparisons
	compareResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      128 * 4,
		"height":     128,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create comparison canvas failed: %w", err)
	}

	var compareResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(compareResp), &compareResult); err != nil {
		return fmt.Errorf("failed to parse comparison result: %w", err)
	}
	comparePath := compareResult.FilePath

	// Import each algorithm result on separate layers at different x positions
	algorithmFiles := []struct {
		name       string
		sourcePath string
	}{
		{"Original", filepath.Join(outputDir, "quantization-original.png")},
		{"Median Cut", filepath.Join(outputDir, "quantization-median_cut-16colors.png")},
		{"K-means", filepath.Join(outputDir, "quantization-kmeans-16colors.png")},
		{"Octree", filepath.Join(outputDir, "quantization-octree-16colors.png")},
	}

	// Create first layer before deleting Layer 1 (can't delete last layer)
	if _, err := callTool(ctx, session, "add_layer", map[string]any{
		"sprite_path": comparePath,
		"layer_name":  "Comp_0",
	}); err != nil {
		return fmt.Errorf("add_layer failed: %w", err)
	}

	// Now delete Layer 1 so it doesn't cover everything
	if _, err := callTool(ctx, session, "delete_layer", map[string]any{
		"sprite_path": comparePath,
		"layer_name":  "Layer 1",
	}); err != nil {
		return fmt.Errorf("delete_layer failed: %w", err)
	}

	for i, file := range algorithmFiles {
		// Verify the source file exists
		if _, err := os.Stat(file.sourcePath); err != nil {
			return fmt.Errorf("source file %s does not exist: %w", file.sourcePath, err)
		}
		logger.Information("  Importing {Name} from {Path}", file.name, file.sourcePath)

		// Create a new layer for each image (except first, already created)
		layerName := fmt.Sprintf("Comp_%d", i)
		if i > 0 {
			if _, err := callTool(ctx, session, "add_layer", map[string]any{
				"sprite_path": comparePath,
				"layer_name":  layerName,
			}); err != nil {
				return fmt.Errorf("add_layer failed: %w", err)
			}
		}

		// Import image onto this layer at the appropriate x position
		if _, err := callTool(ctx, session, "import_image", map[string]any{
			"sprite_path":  comparePath,
			"image_path":   file.sourcePath,
			"layer_name":   layerName,
			"frame_number": 1,
			"position": map[string]any{
				"x": i * 128,
				"y": 0,
			},
		}); err != nil {
			return fmt.Errorf("import comparison image %s failed: %w", file.name, err)
		}

		logger.Information("  Added {Name} at position {X}", file.name, i*128)
	}

	// Flatten all layers before export so they're composited together
	logger.Information("  Flattening layers for: {Path}", comparePath)

	// Check if file exists
	if _, err := os.Stat(comparePath); os.IsNotExist(err) {
		return fmt.Errorf("comparison sprite file does not exist: %s", comparePath)
	}

	flattenScript := gen.FlattenLayers()
	flattenCtx, flattenCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer flattenCancel()
	if _, err := asepriteClient.ExecuteLua(flattenCtx, flattenScript, comparePath); err != nil {
		return fmt.Errorf("flatten layers failed: %w", err)
	}
	logger.Information("  Flattened successfully")

	// Export comparison
	comparisonPath := filepath.Join(outputDir, "quantization-comparison.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  comparePath,
		"output_path":  comparisonPath,
		"format":       "png",
		"frame_number": 1,
	}); err != nil {
		return fmt.Errorf("export comparison failed: %w", err)
	}
	logger.Information("  Exported comparison: {Path}", comparisonPath)
	logger.Information("  Layout: Original | Median Cut | K-means | Octree")

	// Clean up
	os.Remove(comparePath)

	logger.Information("")
	logger.Information("Example completed successfully!")
	logger.Information("")
	logger.Information("Output files created:")
	logger.Information("  - quantization-original.png (full color gradient)")
	logger.Information("  - quantization-median_cut-*.png (3 variants)")
	logger.Information("  - quantization-kmeans-16colors.png")
	logger.Information("  - quantization-octree-16colors.png")
	logger.Information("  - quantization-comparison.png (side-by-side)")

	return nil
}

func callTool(ctx context.Context, session *mcp.ClientSession, name string, args map[string]any) (string, error) {
	resp, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return "", fmt.Errorf("tool call failed: %w", err)
	}

	if resp.IsError {
		return "", fmt.Errorf("tool returned error")
	}

	if len(resp.Content) > 0 {
		if textContent, ok := resp.Content[0].(*mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in response")
}

func findServerBinary() string {
	candidates := []string{
		"../../bin/pixel-mcp",
		"../../bin/pixel-mcp.exe",
		"./bin/pixel-mcp",
		"./bin/pixel-mcp.exe",
		"pixel-mcp",
		"pixel-mcp.exe",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	if path, err := exec.LookPath("pixel-mcp"); err == nil {
		return path
	}

	return ""
}

func createLogger() core.Logger {
	sink := sinks.NewConsoleSink()
	logger := mtlog.New(
		mtlog.WithSink(sink),
		mtlog.WithMinimumLevel(core.InformationLevel),
	)
	return logger
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
