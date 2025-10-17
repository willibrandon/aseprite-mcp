//go:build integration
// +build integration

package tools

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// Integration tests for palette-aware drawing with real Aseprite.
// These tests verify that the use_palette flag correctly snaps colors to the sprite's palette.

func TestIntegration_DrawPixels_WithPaletteSnapping(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create an indexed sprite with a limited palette
	spritePath := testutil.TempSpritePath(t, "test-draw-pixels-palette.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set a simple 4-color palette (black, red, green, blue)
	setPaletteScript := gen.SetPalette([]string{"#000000", "#FF0000", "#00FF00", "#0000FF"})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Draw pixels with colors that should snap to palette
	// #FF8080 (light red) should snap to #FF0000 (pure red)
	// #80FF80 (light green) should snap to #00FF00 (pure green)
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 10, Y: 10}, Color: aseprite.Color{R: 255, G: 128, B: 128, A: 255}}, // Light red
		{Point: aseprite.Point{X: 11, Y: 10}, Color: aseprite.Color{R: 128, G: 255, B: 128, A: 255}}, // Light green
		{Point: aseprite.Point{X: 12, Y: 10}, Color: aseprite.Color{R: 128, G: 128, B: 255, A: 255}}, // Light blue
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels, true)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawPixels with palette) error = %v", err)
	}

	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew pixels with palette snapping successfully")
}

func TestIntegration_DrawLine_WithPaletteSnapping(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create sprite with indexed color mode
	spritePath := testutil.TempSpritePath(t, "test-draw-line-palette.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set a grayscale palette
	setPaletteScript := gen.SetPalette([]string{"#000000", "#808080", "#FFFFFF"})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Draw line with a color that should snap to palette
	// #A0A0A0 should snap to #808080 (middle gray)
	drawScript := gen.DrawLine("Layer 1", 1, 10, 10, 50, 50, aseprite.Color{R: 160, G: 160, B: 160, A: 255}, 2, true)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawLine with palette) error = %v", err)
	}

	if !strings.Contains(output, "Line drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew line with palette snapping successfully")
}

func TestIntegration_DrawRectangle_WithPaletteSnapping(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed sprite
	spritePath := testutil.TempSpritePath(t, "test-draw-rect-palette.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set Pico-8 inspired palette subset
	setPaletteScript := gen.SetPalette([]string{"#000000", "#1D2B53", "#7E2553", "#008751", "#AB5236"})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Draw rectangle with color that needs snapping
	drawScript := gen.DrawRectangle("Layer 1", 1, 20, 20, 40, 30, aseprite.Color{R: 100, G: 50, B: 70, A: 255}, true, true)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawRectangle with palette) error = %v", err)
	}

	if !strings.Contains(output, "Rectangle drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew rectangle with palette snapping successfully")
}

func TestIntegration_DrawCircle_WithPaletteSnapping(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed sprite
	spritePath := testutil.TempSpritePath(t, "test-draw-circle-palette.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set simple RGB palette
	setPaletteScript := gen.SetPalette([]string{"#FF0000", "#00FF00", "#0000FF", "#FFFF00", "#FF00FF", "#00FFFF"})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Draw circle with color that will snap to palette
	drawScript := gen.DrawCircle("Layer 1", 1, 50, 50, 20, aseprite.Color{R: 200, G: 200, B: 0, A: 255}, true, true)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawCircle with palette) error = %v", err)
	}

	if !strings.Contains(output, "Circle drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew circle with palette snapping successfully")
}

func TestIntegration_FillArea_WithPaletteSnapping(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed sprite
	spritePath := testutil.TempSpritePath(t, "test-fill-palette.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set warm color palette
	setPaletteScript := gen.SetPalette([]string{"#FF6B6B", "#FFA07A", "#FFD93D", "#6BCB77", "#4D96FF"})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Fill with a color that will snap to palette
	fillScript := gen.FillArea("Layer 1", 1, 50, 50, aseprite.Color{R: 255, G: 150, B: 100, A: 255}, 0, true)
	output, err := client.ExecuteLua(ctx, fillScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(FillArea with palette) error = %v", err)
	}

	if !strings.Contains(output, "Area filled successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Filled area with palette snapping successfully")
}

func TestIntegration_MixedPaletteAndNonPalette_Drawing(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create RGB sprite (palette snapping should still work)
	spritePath := testutil.TempSpritePath(t, "test-mixed-palette.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set a custom palette (even in RGB mode, palette can be used)
	setPaletteScript := gen.SetPalette([]string{"#000000", "#FFFFFF"})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Draw without palette snapping (exact color)
	pixels1 := []aseprite.Pixel{
		{Point: aseprite.Point{X: 10, Y: 10}, Color: aseprite.Color{R: 128, G: 128, B: 128, A: 255}},
	}
	drawScript1 := gen.DrawPixels("Layer 1", 1, pixels1, false)
	_, err = client.ExecuteLua(ctx, drawScript1, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw without palette: %v", err)
	}

	// Draw with palette snapping (should snap to black or white)
	pixels2 := []aseprite.Pixel{
		{Point: aseprite.Point{X: 20, Y: 10}, Color: aseprite.Color{R: 128, G: 128, B: 128, A: 255}},
	}
	drawScript2 := gen.DrawPixels("Layer 1", 1, pixels2, true)
	output, err := client.ExecuteLua(ctx, drawScript2, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw with palette: %v", err)
	}

	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Mixed palette and non-palette drawing works correctly")
}

// TestIntegration_IndexedColorMode_DrawingVerification tests issue #1:
// Drawing tools should actually draw pixels in indexed color mode, not create transparent pixels.
func TestIntegration_IndexedColorMode_DrawingVerification(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed color mode sprite
	spritePath := testutil.TempSpritePath(t, "test-indexed-drawing-bug.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set a simple 4-color palette
	setPaletteScript := gen.SetPalette([]string{"#000000", "#FF0000", "#00FF00", "#0000FF"})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Test draw_pixels
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 10, Y: 10}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}}, // Red
		{Point: aseprite.Point{X: 11, Y: 10}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}}, // Green
		{Point: aseprite.Point{X: 12, Y: 10}, Color: aseprite.Color{R: 0, G: 0, B: 255, A: 255}}, // Blue
	}
	drawScript := gen.DrawPixels("Layer 1", 1, pixels, true)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawPixels) error = %v", err)
	}
	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify pixels were actually drawn (not transparent)
	getScript := gen.GetPixels("Layer 1", 1, 10, 10, 3, 1)
	output, err = client.ExecuteLua(ctx, getScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
	}

	var pixelData []PixelData
	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse pixel data: %v, output: %s", err, output)
	}

	// Check that pixels are NOT transparent
	for i, p := range pixelData {
		if strings.HasPrefix(strings.ToUpper(p.Color), "#00000000") || strings.HasPrefix(strings.ToUpper(p.Color), "#01000000") {
			t.Errorf("Pixel %d at (%d,%d) is transparent: %s (BUG: draw_pixels failed in indexed mode)", i, p.X, p.Y, p.Color)
		}
	}

	t.Logf("✓ draw_pixels in indexed mode: verified %d non-transparent pixels", len(pixelData))

	// Test draw_rectangle
	drawRectScript := gen.DrawRectangle("Layer 1", 1, 20, 20, 5, 5, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, true)
	output, err = client.ExecuteLua(ctx, drawRectScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawRectangle) error = %v", err)
	}
	if !strings.Contains(output, "Rectangle drawn successfully") {
		t.Errorf("Expected success message for rectangle, got: %s", output)
	}

	// Verify rectangle was drawn
	getRectScript := gen.GetPixels("Layer 1", 1, 20, 20, 5, 5)
	output, err = client.ExecuteLua(ctx, getRectScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels for rectangle) error = %v", err)
	}

	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse rectangle pixel data: %v", err)
	}

	transparentCount := 0
	for _, p := range pixelData {
		if strings.HasPrefix(strings.ToUpper(p.Color), "#00000000") || strings.HasPrefix(strings.ToUpper(p.Color), "#01000000") {
			transparentCount++
		}
	}
	if transparentCount > 0 {
		t.Errorf("BUG: draw_rectangle created %d transparent pixels out of %d in indexed mode", transparentCount, len(pixelData))
	}

	t.Logf("✓ draw_rectangle in indexed mode: verified %d non-transparent pixels", len(pixelData))

	// Test fill_area
	fillScript := gen.FillArea("Layer 1", 1, 40, 40, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, 0, true)
	output, err = client.ExecuteLua(ctx, fillScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(FillArea) error = %v", err)
	}
	if !strings.Contains(output, "Area filled successfully") {
		t.Errorf("Expected success message for fill, got: %s", output)
	}

	// Verify fill worked (check a sample region)
	getFillScript := gen.GetPixels("Layer 1", 1, 40, 40, 10, 10)
	output, err = client.ExecuteLua(ctx, getFillScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels for fill) error = %v", err)
	}

	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse fill pixel data: %v", err)
	}

	transparentCount = 0
	for _, p := range pixelData {
		if strings.HasPrefix(strings.ToUpper(p.Color), "#00000000") || strings.HasPrefix(strings.ToUpper(p.Color), "#01000000") {
			transparentCount++
		}
	}
	if transparentCount > 0 {
		t.Errorf("BUG: fill_area created/left %d transparent pixels in indexed mode", transparentCount)
	}

	t.Logf("✓ fill_area in indexed mode: verified non-transparent pixels")
}

// TestIntegration_DrawWithDither_RGBMode tests issue #2:
// draw_with_dither should create a dithering pattern, not transparent pixels.
func TestIntegration_DrawWithDither_RGBMode(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create sprite in RGB mode (as specified in issue #2)
	spritePath := testutil.TempSpritePath(t, "test-dither-transparency-bug.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Call draw_with_dither with two valid colors and bayer_4x4 pattern
	color1 := "#FF0000" // Red
	color2 := "#0000FF" // Blue
	drawScript := gen.DrawWithDither("Layer 1", 1, 10, 10, 20, 20, color1, color2, "bayer_4x4", 0.5)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawWithDither) error = %v", err)
	}

	if !strings.Contains(output, "Dithering applied successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Check pixels with get_pixels to verify they are NOT transparent
	getScript := gen.GetPixels("Layer 1", 1, 10, 10, 20, 20)
	output, err = client.ExecuteLua(ctx, getScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
	}

	var pixelData []PixelData
	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse pixel data: %v, output: %s", err, output)
	}

	// Count transparent pixels (the bug)
	transparentCount := 0
	redCount := 0
	blueCount := 0
	otherCount := 0

	for _, p := range pixelData {
		colorUpper := strings.ToUpper(p.Color)
		if colorUpper == "#00000000" || colorUpper == "#01000000" {
			transparentCount++
			t.Logf("BUG: Found transparent pixel at (%d,%d): %s", p.X, p.Y, p.Color)
		} else if colorUpper == "#FF0000FF" {
			redCount++
		} else if colorUpper == "#0000FFFF" {
			blueCount++
		} else {
			otherCount++
			t.Logf("Unexpected color at (%d,%d): %s (expected red or blue)", p.X, p.Y, p.Color)
		}
	}

	// Expected: Dithering pattern mixing color1 and color2
	// Actual (bug): Transparent pixels appear in the dithered region
	if transparentCount > 0 {
		t.Errorf("BUG: draw_with_dither created %d transparent pixels out of %d total (issue #2)", transparentCount, len(pixelData))
	}

	// Verify we have a dithering pattern (mix of both colors)
	if redCount == 0 && blueCount == 0 {
		t.Errorf("No dither pattern found: red=%d, blue=%d, transparent=%d, other=%d", redCount, blueCount, transparentCount, otherCount)
	}

	t.Logf("✓ draw_with_dither pattern: red=%d, blue=%d, transparent=%d, other=%d", redCount, blueCount, transparentCount, otherCount)
}

// TestIntegration_DrawWithDither_IndexedMode tests issue #2 in indexed color mode:
// draw_with_dither may create transparent pixels in indexed mode like issue #1.
func TestIntegration_DrawWithDither_IndexedMode(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create sprite in indexed color mode
	spritePath := testutil.TempSpritePath(t, "test-dither-indexed.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set a palette with our test colors
	setPaletteScript := gen.SetPalette([]string{"#000000", "#FF0000", "#0000FF", "#FFFFFF"})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Call draw_with_dither with two valid colors from the palette
	color1 := "#FF0000" // Red
	color2 := "#0000FF" // Blue
	drawScript := gen.DrawWithDither("Layer 1", 1, 10, 10, 20, 20, color1, color2, "bayer_4x4", 0.5)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawWithDither) error = %v", err)
	}

	if !strings.Contains(output, "Dithering applied successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Check pixels with get_pixels to verify they are NOT transparent
	getScript := gen.GetPixels("Layer 1", 1, 10, 10, 20, 20)
	output, err = client.ExecuteLua(ctx, getScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
	}

	var pixelData []PixelData
	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse pixel data: %v, output: %s", err, output)
	}

	// Count transparent pixels (the bug)
	transparentCount := 0
	redCount := 0
	blueCount := 0
	otherCount := 0

	for _, p := range pixelData {
		colorUpper := strings.ToUpper(p.Color)
		if colorUpper == "#00000000" || colorUpper == "#01000000" {
			transparentCount++
			t.Logf("BUG: Found transparent pixel at (%d,%d): %s", p.X, p.Y, p.Color)
		} else if colorUpper == "#FF0000FF" {
			redCount++
		} else if colorUpper == "#0000FFFF" {
			blueCount++
		} else {
			otherCount++
			t.Logf("Unexpected color at (%d,%d): %s (expected red or blue)", p.X, p.Y, p.Color)
		}
	}

	// Expected: Dithering pattern mixing color1 and color2
	// Actual (bug): Transparent pixels appear in the dithered region (issue #2)
	if transparentCount > 0 {
		t.Errorf("BUG: draw_with_dither created %d transparent pixels in indexed mode (issue #2)", transparentCount)
	}

	// Verify we have a dithering pattern (mix of both colors)
	if redCount == 0 && blueCount == 0 {
		t.Errorf("No dither pattern found: red=%d, blue=%d, transparent=%d, other=%d", redCount, blueCount, transparentCount, otherCount)
	}

	t.Logf("✓ draw_with_dither in indexed mode: red=%d, blue=%d, transparent=%d, other=%d", redCount, blueCount, transparentCount, otherCount)
}
