//go:build integration

package tools

import (
	"context"
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

// TestIntegration_DrawRectangle_GetPixels_PaletteIndex0Bug tests the EXACT scenario
// reported by Claude Desktop where get_pixels skips palette index 0 pixels after
// drawing rectangles.
//
// EXACT REPRODUCTION:
// 1. Create indexed sprite with palette: [RED, GREEN, BLUE]
// 2. Draw filled RED rectangle at (0,0) to (3,3) - 16 pixels using palette index 0
// 3. Draw filled GREEN rectangle at (4,0) to (7,3) - 16 pixels using palette index 1
// 4. Call get_pixels on region (0,0) to (7,3) - 32 pixels total
// 5. BUG: Only returns 16 GREEN pixels, missing all 16 RED pixels (palette index 0)
func TestIntegration_DrawRectangle_GetPixels_PaletteIndex0Bug(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create indexed sprite
	spritePath := testutil.TempSpritePath(t, "test-rectangle-palette-index0.aseprite")
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

	// Draw filled RED rectangle at (0,0) to (3,3) using palette index 0
	drawRedScript := gen.DrawRectangle(
		"Layer 1",
		1,
		0, 0,  // x, y
		4, 4,  // width, height
		aseprite.Color{R: 255, G: 0, B: 0, A: 255}, // RED
		true,  // filled
		true,  // use_palette (snap to nearest palette color)
	)
	_, err = client.ExecuteLua(ctx, drawRedScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw RED rectangle: %v", err)
	}

	// Draw filled GREEN rectangle at (4,0) to (7,3) using palette index 1
	drawGreenScript := gen.DrawRectangle(
		"Layer 1",
		1,
		4, 0,  // x, y
		4, 4,  // width, height
		aseprite.Color{R: 0, G: 255, B: 0, A: 255}, // GREEN
		true,  // filled
		true,  // use_palette
	)
	_, err = client.ExecuteLua(ctx, drawGreenScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw GREEN rectangle: %v", err)
	}

	// Read back ALL pixels from region (0,0) to (7,3) - should be 32 pixels total
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
	colorCounts := make(map[string]int)
	for _, p := range pixels {
		colorCounts[p.Color]++
	}

	redCount := colorCounts["#FF0000FF"]
	greenCount := colorCounts["#00FF00FF"]

	// Log what we got
	t.Logf("Total pixels returned: %d (expected 32)", len(pixels))
	t.Logf("RED pixels (palette index 0): %d (expected 16)", redCount)
	t.Logf("GREEN pixels (palette index 1): %d (expected 16)", greenCount)
	t.Logf("All color counts: %v", colorCounts)

	// This test should FAIL if the bug exists
	if len(pixels) != 32 {
		t.Errorf("BUG CONFIRMED: Expected 32 total pixels, got %d", len(pixels))
	}

	if redCount != 16 {
		t.Errorf("BUG CONFIRMED: Expected 16 RED pixels (palette index 0), got %d - palette index 0 pixels are missing!", redCount)
	}

	if greenCount != 16 {
		t.Errorf("Expected 16 GREEN pixels (palette index 1), got %d", greenCount)
	}
}
