//go:build integration

package tools

import (
	"context"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// TestIntegration_DrawPixels_NewLayer_Bug tests the bug where DrawPixels
// fails to create a proper cel image when drawing on a newly created layer.
//
// ROOT CAUSE: DrawPixels calls spr:newCel() without providing an image,
// so cel.image is nil. Calling img:putPixel() on nil fails or does nothing.
func TestIntegration_DrawPixels_NewLayer_Bug(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed sprite
	spritePath := testutil.TempSpritePath(t, "test-pixels-new-layer.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}

	// Set palette
	setPaletteScript := gen.SetPalette([]string{
		"#FF0000FF", // Index 0: RED
		"#00FF00FF", // Index 1: GREEN
		"#0000FFFF", // Index 2: BLUE
	})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Add a NEW layer (this layer has NO cels yet)
	addLayerScript := gen.AddLayer("Test Layer")
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add layer: %v", err)
	}

	// Draw RED pixels at (0-3, 0-3) on the NEW layer (this will trigger cel creation)
	redPixels := make([]aseprite.Pixel, 0, 16)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			redPixels = append(redPixels, aseprite.Pixel{
				Point: aseprite.Point{X: x, Y: y},
				Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}, // RED
			})
		}
	}
	drawRedScript := gen.DrawPixels("Test Layer", 1, redPixels, true)
	_, err = client.ExecuteLua(ctx, drawRedScript, spritePath)
	if err != nil {
		// Before the fix, this would fail with "attempt to index a nil value (local 'img')"
		t.Fatalf("Failed to draw red pixels on new layer: %v - BUG: cel.image is nil!", err)
	}

	// Draw GREEN pixels at (4-7, 0-3) to verify cel image is working properly
	greenPixels := make([]aseprite.Pixel, 0, 16)
	for y := 0; y < 4; y++ {
		for x := 4; x < 8; x++ {
			greenPixels = append(greenPixels, aseprite.Pixel{
				Point: aseprite.Point{X: x, Y: y},
				Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}, // GREEN
			})
		}
	}
	drawGreenScript := gen.DrawPixels("Test Layer", 1, greenPixels, true)
	_, err = client.ExecuteLua(ctx, drawGreenScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw green pixels: %v", err)
	}

	// Read back pixels to verify both colors were drawn correctly
	getPixelsScript := gen.GetPixels("Test Layer", 1, 0, 0, 8, 4)
	output, err := client.ExecuteLua(ctx, getPixelsScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to get pixels: %v", err)
	}

	// Parse pixel data
	pixels, err := testutil.ParsePixelData(output)
	if err != nil {
		t.Fatalf("Failed to parse pixel data: %v", err)
	}

	// Count pixels by color
	colorCounts := make(map[string]int)
	for _, p := range pixels {
		colorCounts[p.Color]++
	}

	redCount := colorCounts["#FF0000FF"]
	greenCount := colorCounts["#00FF00FF"]

	// Log results
	t.Logf("Total pixels returned: %d (expected 32 for 8x4 region)", len(pixels))
	t.Logf("RED pixels: %d (expected 16)", redCount)
	t.Logf("GREEN pixels: %d (expected 16)", greenCount)
	t.Logf("All color counts: %v", colorCounts)

	// Verify the fix worked: DrawPixels successfully created cel image and drew both colors
	if len(pixels) != 32 {
		t.Errorf("Expected 32 total pixels, got %d", len(pixels))
	}

	if redCount != 16 {
		t.Errorf("Expected 16 RED pixels, got %d", redCount)
	}

	if greenCount != 16 {
		t.Errorf("Expected 16 GREEN pixels, got %d", greenCount)
	}
}
