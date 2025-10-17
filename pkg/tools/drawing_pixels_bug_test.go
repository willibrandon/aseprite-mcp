//go:build integration

package tools

import (
	"context"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// TestIntegration_DrawPixels_GetPixels_PaletteIndex0Bug tests the bug where
// get_pixels skips pixels using palette index 0 in indexed mode.
//
// BUG SCENARIO (reported by Claude Desktop):
// 1. Create indexed sprite with palette: [RED, GREEN, BLUE]
// 2. Draw RED pixels (index 0) at x=0-3, y=0-3
// 3. Draw GREEN pixels (index 1) at x=4-7, y=0-3
// 4. Call get_pixels on region (0, 0, 8, 4)
// 5. EXPECTED: Returns 32 pixels (16 red + 16 green)
// 6. ACTUAL: Returns only 16 pixels (0 red + 16 green) - palette index 0 missing!
func TestIntegration_DrawPixels_GetPixels_PaletteIndex0Bug(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed sprite
	spritePath := testutil.TempSpritePath(t, "test-pixels-palette-index0.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}

	// Set palette with RED at index 0, GREEN at index 1, BLUE at index 2
	setPaletteScript := gen.SetPalette([]string{
		"#FF0000FF", // Index 0: RED
		"#00FF00FF", // Index 1: GREEN
		"#0000FFFF", // Index 2: BLUE
	})
	_, err = client.ExecuteLua(ctx, setPaletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Draw RED pixels (palette index 0) at x=0-3, y=0-3 (16 pixels)
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

	// Draw GREEN pixels (palette index 1) at x=4-7, y=0-3 (16 pixels)
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
	otherCount := 0

	for _, p := range pixels {
		switch p.Color {
		case "#FF0000FF": // RED
			redCount++
		case "#00FF00FF": // GREEN
			greenCount++
		default:
			otherCount++
		}
	}

	// Log what we got for debugging
	t.Logf("Total pixels returned: %d (expected 32)", len(pixels))
	t.Logf("RED pixels (index 0): %d (expected 16)", redCount)
	t.Logf("GREEN pixels (index 1): %d (expected 16)", greenCount)
	t.Logf("Other colors: %d (expected 0)", otherCount)

	// Assertions
	if len(pixels) != 32 {
		t.Errorf("Expected 32 total pixels, got %d", len(pixels))
	}

	if redCount != 16 {
		t.Errorf("Expected 16 RED pixels (palette index 0), got %d - BUG: palette index 0 pixels are missing!", redCount)
	}

	if greenCount != 16 {
		t.Errorf("Expected 16 GREEN pixels (palette index 1), got %d", greenCount)
	}

	if otherCount != 0 {
		t.Errorf("Expected 0 pixels of other colors, got %d", otherCount)
	}

	// Verify specific pixel locations
	if redCount > 0 {
		// Check that red pixels are in correct positions (0-3, 0-3)
		redPositions := make(map[string]bool)
		for _, p := range pixels {
			if p.Color == "#FF0000FF" {
				if p.X < 0 || p.X > 3 || p.Y < 0 || p.Y > 3 {
					t.Errorf("RED pixel at (%d, %d) is outside expected region (0-3, 0-3)", p.X, p.Y)
				}
				redPositions[testutil.FormatPixelPos(p.X, p.Y)] = true
			}
		}
		t.Logf("RED pixels found at %d unique positions", len(redPositions))
	}

	if greenCount > 0 {
		// Check that green pixels are in correct positions (4-7, 0-3)
		greenPositions := make(map[string]bool)
		for _, p := range pixels {
			if p.Color == "#00FF00FF" {
				if p.X < 4 || p.X > 7 || p.Y < 0 || p.Y > 3 {
					t.Errorf("GREEN pixel at (%d, %d) is outside expected region (4-7, 0-3)", p.X, p.Y)
				}
				greenPositions[testutil.FormatPixelPos(p.X, p.Y)] = true
			}
		}
		t.Logf("GREEN pixels found at %d unique positions", len(greenPositions))
	}
}
