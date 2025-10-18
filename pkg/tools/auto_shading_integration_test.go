//go:build integration
// +build integration

package tools

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// Integration tests for auto-shading tools with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools -run TestIntegration_AutoShading

func TestIntegration_ApplyAutoShading_CellStyle(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with a simple shape
	spritePath := testutil.TempSpritePath(t, "test-autoshading-cell.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a circle to shade
	drawScript := gen.DrawCircle("Layer 1", 1, 32, 32, 20, aseprite.Color{R: 255, G: 128, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw circle: %v", err)
	}

	// Export layer to PNG
	tempPNG := testutil.TempSpritePath(t, "layer.png")
	defer os.Remove(tempPNG)

	exportScript := gen.ExportSprite(tempPNG, 1)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to export layer: %v", err)
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

	// Apply auto-shading with cell style
	shadedImg, generatedColors, regionsShadedCount, err := aseprite.ApplyAutoShading(
		img,
		"top_left",
		0.6,
		"cell",
		true,
	)
	if err != nil {
		t.Fatalf("Auto-shading failed: %v", err)
	}

	if shadedImg == nil {
		t.Fatal("Shaded image should not be nil")
	}

	if len(generatedColors) == 0 {
		t.Error("Generated colors should not be empty")
	}

	if regionsShadedCount == 0 {
		t.Error("Regions shaded count should not be 0")
	}

	t.Logf("✓ Applied cell shading: %d colors generated, %d regions shaded", len(generatedColors), regionsShadedCount)
}

func TestIntegration_ApplyAutoShading_SmoothStyle(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-autoshading-smooth.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a rectangle
	drawScript := gen.DrawRectangle("Layer 1", 1, 16, 16, 32, 32, aseprite.Color{R: 0, G: 128, B: 255, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw rectangle: %v", err)
	}

	// Export and shade
	tempPNG := testutil.TempSpritePath(t, "rect.png")
	defer os.Remove(tempPNG)

	exportScript := gen.ExportSprite(tempPNG, 1)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to export: %v", err)
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

	// Apply auto-shading with smooth style
	shadedImg, generatedColors, regionsShadedCount, err := aseprite.ApplyAutoShading(
		img,
		"top",
		0.5,
		"smooth",
		false, // No hue shift
	)
	if err != nil {
		t.Fatalf("Auto-shading failed: %v", err)
	}

	if shadedImg == nil {
		t.Fatal("Shaded image should not be nil")
	}

	t.Logf("✓ Applied smooth shading: %d colors generated, %d regions shaded", len(generatedColors), regionsShadedCount)
}

func TestIntegration_ApplyAutoShading_SoftStyle(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 255, A: 255})
		}
	}

	// Apply auto-shading with soft style
	shadedImg, generatedColors, regionsShadedCount, err := aseprite.ApplyAutoShading(
		img,
		"bottom_right",
		0.3,
		"soft",
		true,
	)
	if err != nil {
		t.Fatalf("Auto-shading failed: %v", err)
	}

	if shadedImg == nil {
		t.Fatal("Shaded image should not be nil")
	}

	if len(generatedColors) == 0 {
		t.Error("Generated colors should not be empty")
	}

	t.Logf("✓ Applied soft shading: %d colors generated, %d regions shaded", len(generatedColors), regionsShadedCount)
}

func TestIntegration_ApplyAutoShading_AllLightDirections(t *testing.T) {
	directions := []string{
		"top_left", "top", "top_right",
		"left", "right",
		"bottom_left", "bottom", "bottom_right",
	}

	for _, dir := range directions {
		t.Run(dir, func(t *testing.T) {
			// Create a simple test image
			img := image.NewRGBA(image.Rect(0, 0, 32, 32))
			for y := 8; y < 24; y++ {
				for x := 8; x < 24; x++ {
					img.SetRGBA(x, y, color.RGBA{R: 100, G: 200, B: 100, A: 255})
				}
			}

			// Apply auto-shading
			_, generatedColors, _, err := aseprite.ApplyAutoShading(
				img,
				dir,
				0.5,
				"cell",
				true,
			)
			if err != nil {
				t.Fatalf("Auto-shading failed for direction %s: %v", dir, err)
			}

			if len(generatedColors) == 0 {
				t.Errorf("No colors generated for direction %s", dir)
			}

			t.Logf("✓ Shading from %s: %d colors", dir, len(generatedColors))
		})
	}
}

func TestIntegration_ApplyAutoShading_WithHueShift(t *testing.T) {
	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 128, B: 0, A: 255})
		}
	}

	// Test with hue shift enabled
	_, colorsWithShift, _, err := aseprite.ApplyAutoShading(img, "top_left", 0.6, "cell", true)
	if err != nil {
		t.Fatalf("Auto-shading with hue shift failed: %v", err)
	}

	// Test with hue shift disabled
	_, colorsWithoutShift, _, err := aseprite.ApplyAutoShading(img, "top_left", 0.6, "cell", false)
	if err != nil {
		t.Fatalf("Auto-shading without hue shift failed: %v", err)
	}

	if len(colorsWithShift) == 0 || len(colorsWithoutShift) == 0 {
		t.Fatal("Generated colors should not be empty")
	}

	t.Logf("✓ With hue shift: %d colors", len(colorsWithShift))
	t.Logf("✓ Without hue shift: %d colors", len(colorsWithoutShift))
}
