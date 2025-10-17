//go:build integration
// +build integration

package tools

import (
	"context"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// TestExportMultiLayerCompositing tests that export properly composites all visible layers.
// This test reproduces the bug where only the topmost layer is visible in exported images.
func TestExportMultiLayerCompositing(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a test sprite with multiple layers and distinct colors
	spritePath := filepath.Join(cfg.TempDir, "multilayer_export_test.aseprite")
	script := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add layer 1 (background) - fill with red
	script = gen.AddLayer("Background")
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("Failed to add Background layer: %v", err)
	}

	script = gen.DrawRectangle("Background", 1, 0, 0, 100, 100, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw red background: %v", err)
	}

	// Add layer 2 (middle) - draw blue circle in center
	script = gen.AddLayer("Middle")
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("Failed to add Middle layer: %v", err)
	}

	script = gen.DrawCircle("Middle", 1, 50, 50, 20, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw blue circle: %v", err)
	}

	// Add layer 3 (foreground) - draw small green square
	script = gen.AddLayer("Foreground")
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("Failed to add Foreground layer: %v", err)
	}

	script = gen.DrawRectangle("Foreground", 1, 10, 10, 20, 20, aseprite.Color{R: 0, G: 255, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw green rectangle: %v", err)
	}

	// Export to PNG
	outputPath := filepath.Join(cfg.TempDir, "multilayer_export_test.png")
	script = gen.ExportSprite(outputPath, 1)
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("Failed to export sprite: %v", err)
	}
	defer os.Remove(outputPath)

	// Read the exported PNG
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open exported PNG: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	// Verify pixels contain all three colors (red background, blue circle, green square)
	bounds := img.Bounds()

	// Check for red background (should be visible in corners)
	r, g, b, _ := img.At(5, 5).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red background at (5,5), got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Check for green square (at position 15, 15)
	r, g, b, _ = img.At(15, 15).RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 {
		t.Errorf("Expected green square at (15,15), got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Check for blue circle (at center position 50, 50)
	r, g, b, _ = img.At(50, 50).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("Expected blue circle at (50,50), got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Check red background is visible in bottom-right corner (no overlap)
	r, g, b, _ = img.At(bounds.Max.X-5, bounds.Max.Y-5).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red background at bottom-right, got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	t.Log("Multi-layer export test passed - all layers properly composited")
}
