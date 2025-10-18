//go:build integration
// +build integration

package tools

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// Integration tests for quantization tools with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools -run TestIntegration_Quantize

func TestIntegration_QuantizePalette_MedianCut(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with multiple colors
	spritePath := testutil.TempSpritePath(t, "test-quantize-mediancut.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw rectangles with different colors
	colors := []aseprite.Color{
		{R: 255, G: 0, B: 0, A: 255},     // Red
		{R: 0, G: 255, B: 0, A: 255},     // Green
		{R: 0, G: 0, B: 255, A: 255},     // Blue
		{R: 255, G: 255, B: 0, A: 255},   // Yellow
		{R: 255, G: 0, B: 255, A: 255},   // Magenta
		{R: 0, G: 255, B: 255, A: 255},   // Cyan
		{R: 255, G: 128, B: 0, A: 255},   // Orange
		{R: 128, G: 0, B: 255, A: 255},   // Purple
		{R: 255, G: 0, B: 128, A: 255},   // Pink
		{R: 128, G: 255, B: 0, A: 255},   // Lime
		{R: 0, G: 128, B: 255, A: 255},   // Sky Blue
		{R: 255, G: 128, B: 128, A: 255}, // Light Red
	}

	for i, color := range colors {
		x := (i % 4) * 16
		y := (i / 4) * 21
		drawScript := gen.DrawRectangle("Layer 1", 1, x, y, 16, 21, color, false)
		_, err := client.ExecuteLua(ctx, drawScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to draw rectangle: %v", err)
		}
	}

	// Export sprite to PNG for quantization
	tempPNG := testutil.TempSpritePath(t, "sprite.png")
	defer os.Remove(tempPNG)

	exportScript := gen.ExportSprite(tempPNG, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to export sprite: %v", err)
	}

	// Load PNG and perform quantization
	imgFile, err := os.Open(tempPNG)
	if err != nil {
		t.Fatalf("Failed to open PNG: %v", err)
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	// Quantize to 8 colors using median_cut
	palette, originalColors, err := aseprite.QuantizePalette(img, 8, "median_cut", true)
	if err != nil {
		t.Fatalf("Quantization failed: %v", err)
	}

	if originalColors == 0 {
		t.Error("Original colors should not be 0")
	}

	if len(palette) == 0 {
		t.Fatal("Palette should not be empty")
	}

	t.Logf("✓ Quantized from %d to %d colors using median_cut", originalColors, len(palette))

	// Apply quantized palette to sprite
	applyScript := gen.ApplyQuantizedPalette(palette, originalColors, "median_cut", false)
	output, err := client.ExecuteLua(ctx, applyScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to apply quantized palette: %v", err)
	}

	// Parse JSON output
	var result struct {
		Success         bool     `json:"success"`
		OriginalColors  int      `json:"original_colors"`
		QuantizedColors int      `json:"quantized_colors"`
		ColorMode       string   `json:"color_mode"`
		Palette         []string `json:"palette"`
		AlgorithmUsed   string   `json:"algorithm_used"`
	}

	// Extract JSON from output
	startIdx := strings.Index(output, "{")
	if startIdx == -1 {
		t.Fatalf("No JSON in output: %s", output)
	}
	jsonStr := output[startIdx:]

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if !result.Success {
		t.Error("Expected success=true")
	}

	if result.ColorMode != "indexed" {
		t.Errorf("Expected color_mode=indexed, got %s", result.ColorMode)
	}

	t.Logf("✓ Applied quantized palette successfully: %d colors", result.QuantizedColors)
}

func TestIntegration_QuantizePalette_Kmeans(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with gradient
	spritePath := testutil.TempSpritePath(t, "test-quantize-kmeans.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw gradient using multiple colors
	for y := 0; y < 64; y++ {
		gray := uint8((y * 255) / 63)
		color := aseprite.Color{R: gray, G: gray, B: gray, A: 255}
		drawScript := gen.DrawLine("Layer 1", 1, 0, y, 63, y, color, 1, false)
		_, err := client.ExecuteLua(ctx, drawScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to draw line: %v", err)
		}
	}

	// Export and quantize
	tempPNG := testutil.TempSpritePath(t, "gradient.png")
	defer os.Remove(tempPNG)

	exportScript := gen.ExportSprite(tempPNG, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to export sprite: %v", err)
	}

	imgFile, err := os.Open(tempPNG)
	if err != nil {
		t.Fatalf("Failed to open PNG: %v", err)
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	// Quantize to 4 colors using kmeans
	palette, originalColors, err := aseprite.QuantizePalette(img, 4, "kmeans", true)
	if err != nil {
		t.Fatalf("Quantization failed: %v", err)
	}

	if len(palette) == 0 {
		t.Fatal("Palette should not be empty")
	}

	t.Logf("✓ Quantized gradient from %d to %d colors using kmeans", originalColors, len(palette))
}

func TestIntegration_QuantizePalette_Octree(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-quantize-octree.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw circles with various colors
	colors := []aseprite.Color{
		{R: 255, G: 0, B: 0, A: 255},     // Red
		{R: 0, G: 255, B: 0, A: 255},     // Green
		{R: 0, G: 0, B: 255, A: 255},     // Blue
		{R: 255, G: 255, B: 0, A: 255},   // Yellow
		{R: 255, G: 0, B: 255, A: 255},   // Magenta
		{R: 0, G: 255, B: 255, A: 255},   // Cyan
	}
	for i, color := range colors {
		x := (i % 3) * 21 + 10
		y := (i / 3) * 32 + 16
		drawScript := gen.DrawCircle("Layer 1", 1, x, y, 8, color, false)
		_, err := client.ExecuteLua(ctx, drawScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to draw circle: %v", err)
		}
	}

	// Export and quantize
	tempPNG := testutil.TempSpritePath(t, "circles.png")
	defer os.Remove(tempPNG)

	exportScript := gen.ExportSprite(tempPNG, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to export sprite: %v", err)
	}

	imgFile, err := os.Open(tempPNG)
	if err != nil {
		t.Fatalf("Failed to open PNG: %v", err)
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	// Quantize to 8 colors using octree
	palette, originalColors, err := aseprite.QuantizePalette(img, 8, "octree", true)
	if err != nil {
		t.Fatalf("Quantization failed: %v", err)
	}

	if len(palette) == 0 {
		t.Fatal("Palette should not be empty")
	}

	t.Logf("✓ Quantized from %d to %d colors using octree", originalColors, len(palette))
}

func TestIntegration_QuantizePalette_WithDithering(t *testing.T) {
	// Create a simple gradient image
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			gray := uint8((x + y) * 4)
			img.SetRGBA(x, y, color.RGBA{R: gray, G: gray, B: gray, A: 255})
		}
	}

	// Quantize with dithering
	palette, originalColors, err := aseprite.QuantizePalette(img, 4, "median_cut", true)
	if err != nil {
		t.Fatalf("Quantization with dithering failed: %v", err)
	}

	if len(palette) == 0 {
		t.Fatal("Palette should not be empty")
	}

	t.Logf("✓ Quantized from %d colors with dithering to %d colors", originalColors, len(palette))
}

func TestIntegration_QuantizePalette_WithDitheringAndReplacement(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with gradient
	spritePath := testutil.TempSpritePath(t, "test-quantize-dither-replace.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw gradient (should produce many colors)
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			gray := uint8((x + y) * 2)
			color := aseprite.Color{R: gray, G: gray, B: gray, A: 255}
			pixels := []aseprite.Pixel{{Point: aseprite.Point{X: x, Y: y}, Color: color}}
			drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
			_, err := client.ExecuteLua(ctx, drawScript, spritePath)
			if err != nil {
				t.Fatalf("Failed to draw pixel: %v", err)
			}
		}
	}

	// Export sprite to PNG
	tempPNG := testutil.TempSpritePath(t, "gradient.png")
	defer os.Remove(tempPNG)

	exportScript := gen.ExportSprite(tempPNG, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to export sprite: %v", err)
	}

	// Load PNG
	imgFile, err := os.Open(tempPNG)
	if err != nil {
		t.Fatalf("Failed to open PNG: %v", err)
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	// Quantize with dithering enabled
	palette, originalColors, err := aseprite.QuantizePalette(img, 8, "median_cut", true)
	if err != nil {
		t.Fatalf("Quantization failed: %v", err)
	}

	if len(palette) == 0 {
		t.Fatal("Palette should not be empty")
	}

	t.Logf("✓ Quantized from %d to %d colors with dithering", originalColors, len(palette))

	// Convert palette to color.Color slice (testing the color parsing)
	paletteColors := make([]color.Color, len(palette))
	for i, hexColor := range palette {
		// Parse hex color
		var r, g, b uint8
		if _, err := parseHexColorValues(hexColor, &r, &g, &b); err != nil {
			t.Fatalf("Failed to parse palette color %s: %v", hexColor, err)
		}
		paletteColors[i] = color.RGBA{R: r, G: g, B: b, A: 255}
	}

	// Apply dithering
	ditheredImg := aseprite.RemapPixelsWithDithering(img, paletteColors, true)
	if ditheredImg == nil {
		t.Fatal("Dithered image should not be nil")
	}

	// Save dithered image
	ditheredPNG := testutil.TempSpritePath(t, "dithered.png")
	defer os.Remove(ditheredPNG)

	ditheredFile, err := os.Create(ditheredPNG)
	if err != nil {
		t.Fatalf("Failed to create dithered PNG: %v", err)
	}
	defer ditheredFile.Close()

	if err := png.Encode(ditheredFile, ditheredImg); err != nil {
		t.Fatalf("Failed to encode dithered PNG: %v", err)
	}

	// Replace sprite content with dithered image
	replaceScript := gen.ReplaceWithImage(ditheredPNG)
	_, err = client.ExecuteLua(ctx, replaceScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to replace sprite with dithered image: %v", err)
	}

	t.Logf("✓ Successfully replaced sprite content with dithered image")

	// Verify sprite was modified
	infoScript := gen.GetSpriteInfo()
	output, err := client.ExecuteLua(ctx, infoScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to get sprite info: %v", err)
	}

	if !strings.Contains(output, "\"width\": 64") {
		t.Error("Sprite dimensions should be preserved after replacement")
	}

	t.Logf("✓ Verified sprite was successfully updated with dithered content")
}

// Helper function to parse hex color values
func parseHexColorValues(hexColor string, r, g, b *uint8) (int, error) {
	hexColor = strings.TrimPrefix(hexColor, "#")

	var rVal, gVal, bVal int
	if len(hexColor) == 6 {
		n, err := parseColor(hexColor, &rVal, &gVal, &bVal)
		if err != nil {
			return 0, err
		}
		*r = uint8(rVal)
		*g = uint8(gVal)
		*b = uint8(bVal)
		return n, nil
	}
	return 0, nil
}

func parseColor(hex string, r, g, b *int) (int, error) {
	n, err := parseHex(hex[:2], r)
	if err != nil {
		return 0, err
	}
	n2, err := parseHex(hex[2:4], g)
	if err != nil {
		return 0, err
	}
	n3, err := parseHex(hex[4:6], b)
	if err != nil {
		return 0, err
	}
	return n + n2 + n3, nil
}

func parseHex(s string, val *int) (int, error) {
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		var v int
		if '0' <= c && c <= '9' {
			v = int(c - '0')
		} else if 'a' <= c && c <= 'f' {
			v = int(c - 'a' + 10)
		} else if 'A' <= c && c <= 'F' {
			v = int(c - 'A' + 10)
		} else {
			return 0, nil
		}
		n = n*16 + v
	}
	*val = n
	return 1, nil
}
