//go:build integration
// +build integration

package tools

import (
	"context"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

// TestIntegration_ExportSprite_IndexedPalette tests that exporting an indexed
// sprite to PNG preserves the custom palette.
//
// This test captures the bug where export_sprite creates a temp sprite without
// copying the palette, resulting in a black/incorrect PNG export.
func TestIntegration_ExportSprite_IndexedPalette(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed sprite
	spritePath := testutil.TempSpritePath(t, "test-export-indexed-palette.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}

	// Set custom palette with distinct colors
	setPaletteScript := gen.SetPalette([]string{
		"#FF0000FF", // Index 0: RED
		"#00FF00FF", // Index 1: GREEN
		"#0000FFFF", // Index 2: BLUE
		"#FFFF00FF", // Index 3: YELLOW
	})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Draw rectangles with different palette colors
	// RED rectangle (palette index 0) at top-left
	drawRedScript := gen.DrawRectangle("Layer 1", 1, 0, 0, 16, 16, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, true)
	_, err = client.ExecuteLua(ctx, drawRedScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw RED rectangle: %v", err)
	}

	// GREEN rectangle (palette index 1) at top-right
	drawGreenScript := gen.DrawRectangle("Layer 1", 1, 48, 0, 16, 16, aseprite.Color{R: 0, G: 255, B: 0, A: 255}, true, true)
	_, err = client.ExecuteLua(ctx, drawGreenScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw GREEN rectangle: %v", err)
	}

	// BLUE rectangle (palette index 2) at bottom-left
	drawBlueScript := gen.DrawRectangle("Layer 1", 1, 0, 48, 16, 16, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, true, true)
	_, err = client.ExecuteLua(ctx, drawBlueScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw BLUE rectangle: %v", err)
	}

	// YELLOW rectangle (palette index 3) at bottom-right
	drawYellowScript := gen.DrawRectangle("Layer 1", 1, 48, 48, 16, 16, aseprite.Color{R: 255, G: 255, B: 0, A: 255}, true, true)
	_, err = client.ExecuteLua(ctx, drawYellowScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw YELLOW rectangle: %v", err)
	}

	// Export to PNG
	pngPath := testutil.TempSpritePath(t, "test-export-indexed-palette.png")
	exportScript := gen.ExportSprite(pngPath, 1)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to export sprite: %v", err)
	}

	// Read the PNG back
	pngFile, err := os.Open(pngPath)
	if err != nil {
		t.Fatalf("Failed to open exported PNG: %v", err)
	}
	defer pngFile.Close()

	img, err := png.Decode(pngFile)
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	// Verify the PNG has the correct colors
	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("PNG dimensions incorrect: got %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
	}

	// Sample pixels from each colored rectangle and verify colors
	testCases := []struct {
		name     string
		x, y     int
		expected aseprite.Color
	}{
		{"RED rectangle", 8, 8, aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{"GREEN rectangle", 56, 8, aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{"BLUE rectangle", 8, 56, aseprite.Color{R: 0, G: 0, B: 255, A: 255}},
		{"YELLOW rectangle", 56, 56, aseprite.Color{R: 255, G: 255, B: 0, A: 255}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := img.At(tc.x, tc.y)
			r, g, b, a := c.RGBA()

			// Convert from 16-bit to 8-bit color values
			actualR := uint8(r >> 8)
			actualG := uint8(g >> 8)
			actualB := uint8(b >> 8)
			actualA := uint8(a >> 8)

			// Allow slight tolerance for PNG compression
			tolerance := uint8(2)

			if !colorMatch(actualR, tc.expected.R, tolerance) ||
				!colorMatch(actualG, tc.expected.G, tolerance) ||
				!colorMatch(actualB, tc.expected.B, tolerance) {
				t.Errorf("BUG CONFIRMED: Pixel at (%d,%d) has wrong color\n"+
					"  Expected: RGBA(%d,%d,%d,%d)\n"+
					"  Got:      RGBA(%d,%d,%d,%d)\n"+
					"  This indicates the palette was not copied during PNG export!",
					tc.x, tc.y,
					tc.expected.R, tc.expected.G, tc.expected.B, tc.expected.A,
					actualR, actualG, actualB, actualA)
			}
		})
	}
}

// colorMatch checks if two color values match within a tolerance
func colorMatch(a, b, tolerance uint8) bool {
	diff := int(a) - int(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= int(tolerance)
}
