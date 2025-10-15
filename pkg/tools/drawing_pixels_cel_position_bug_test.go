//go:build integration

package tools

import (
	"context"
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

// TestIntegration_DrawPixels_CelPosition_Bug tests the bug where DrawPixels
// creates cels with incorrect positions, causing get_pixels to miss pixels.
//
// BUG SCENARIO:
// When drawing pixels sequentially with DrawPixels, if each call creates a new
// cel, the cel position might be set incorrectly based on the first pixel drawn.
// This causes get_pixels to skip pixels that are "outside" the cel's bounds.
//
// ROOT CAUSE: DrawPixels doesn't properly initialize the cel image, causing
// Aseprite to auto-position the cel based on the first pixel, rather than
// positioning it at (0,0).
func TestIntegration_DrawPixels_CelPosition_Bug(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed sprite
	spritePath := testutil.TempSpritePath(t, "test-pixels-cel-position.aseprite")
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

	// CRITICAL: Draw pixels in SEPARATE calls to trigger cel positioning bug
	// First call draws at x=4-7 (this might set cel.position.x = 4)
	greenPixels := make([]aseprite.Pixel, 0, 16)
	for y := 0; y < 4; y++ {
		for x := 4; x < 8; x++ {
			greenPixels = append(greenPixels, aseprite.Pixel{
				Point: aseprite.Point{X: x, Y: y},
				Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}, // GREEN
			})
		}
	}
	drawGreenScript := gen.DrawPixels("Layer 1", 1, greenPixels, true)
	_, err = client.ExecuteLua(ctx, drawGreenScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw green pixels: %v", err)
	}

	// Second call draws at x=0-3 (but cel is already positioned at x=4!)
	redPixels := make([]aseprite.Pixel, 0, 16)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			redPixels = append(redPixels, aseprite.Pixel{
				Point: aseprite.Point{X: x, Y: y},
				Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}, // RED
			})
		}
	}
	drawRedScript := gen.DrawPixels("Layer 1", 1, redPixels, true)
	_, err = client.ExecuteLua(ctx, drawRedScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw red pixels: %v", err)
	}

	// Read back ALL pixels from region (0, 0, 8, 4) - should be 32 pixels total
	getPixelsScript := gen.GetPixels("Layer 1", 1, 0, 0, 8, 4)
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
	redCount := 0
	greenCount := 0

	for _, p := range pixels {
		switch p.Color {
		case "#FF0000FF": // RED
			redCount++
		case "#00FF00FF": // GREEN
			greenCount++
		}
	}

	// Log what we got for debugging
	t.Logf("Total pixels returned: %d (expected 32)", len(pixels))
	t.Logf("RED pixels (index 0): %d (expected 16)", redCount)
	t.Logf("GREEN pixels (index 1): %d (expected 16)", greenCount)

	// This test captures the bug:
	// If DrawPixels doesn't create a full-sprite-sized image with cel at (0,0),
	// then pixels drawn in the second call will be lost or misplaced.

	if len(pixels) != 32 {
		t.Errorf("BUG CONFIRMED: Expected 32 total pixels, got %d - pixels are being lost!", len(pixels))
	}

	if redCount != 16 {
		t.Errorf("BUG CONFIRMED: Expected 16 RED pixels, got %d - RED pixels are missing!", redCount)
		t.Logf("This happens because DrawPixels doesn't create a properly sized cel image at position (0,0)")
	}

	if greenCount != 16 {
		t.Errorf("Expected 16 GREEN pixels, got %d", greenCount)
	}
}
