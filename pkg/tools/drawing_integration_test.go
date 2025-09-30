//go:build integration
// +build integration

package tools

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

// Integration tests for draw_pixels tool with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools

func TestIntegration_DrawPixels_SinglePixel(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-single.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a single red pixel at (10, 10)
	pixels := []aseprite.Pixel{
		{
			Point: aseprite.Point{X: 10, Y: 10},
			Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255},
		},
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawPixels) error = %v", err)
	}

	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew single pixel successfully")
}

func TestIntegration_DrawPixels_MultiplePixels(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-multiple.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw multiple pixels in different colors
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 0, Y: 0}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},   // Red
		{Point: aseprite.Point{X: 1, Y: 0}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},   // Green
		{Point: aseprite.Point{X: 2, Y: 0}, Color: aseprite.Color{R: 0, G: 0, B: 255, A: 255}},   // Blue
		{Point: aseprite.Point{X: 3, Y: 0}, Color: aseprite.Color{R: 255, G: 255, B: 0, A: 255}}, // Yellow
		{Point: aseprite.Point{X: 4, Y: 0}, Color: aseprite.Color{R: 255, G: 0, B: 255, A: 255}}, // Magenta
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawPixels) error = %v", err)
	}

	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew %d pixels successfully", len(pixels))
}

func TestIntegration_DrawPixels_LargeBatch(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-batch.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw 100 pixels in a diagonal line
	pixels := make([]aseprite.Pixel, 100)
	for i := 0; i < 100; i++ {
		pixels[i] = aseprite.Pixel{
			Point: aseprite.Point{X: i, Y: i},
			Color: aseprite.Color{R: uint8(i * 2), G: 100, B: uint8(255 - i*2), A: 255},
		}
	}

	start := time.Now()
	drawScript := gen.DrawPixels("Layer 1", 1, pixels)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("ExecuteLua(DrawPixels) error = %v", err)
	}

	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Performance check: 100 pixels should complete in <2s (per PRD)
	if duration > 2*time.Second {
		t.Errorf("Drawing 100 pixels took %v, expected <2s", duration)
	}

	t.Logf("✓ Drew %d pixels in %v", len(pixels), duration)
}

func TestIntegration_DrawPixels_WithAlpha(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-alpha.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw pixels with varying alpha values
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 10, Y: 10}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}}, // Opaque
		{Point: aseprite.Point{X: 11, Y: 10}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 192}}, // 75% opacity
		{Point: aseprite.Point{X: 12, Y: 10}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 128}}, // 50% opacity
		{Point: aseprite.Point{X: 13, Y: 10}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 64}},  // 25% opacity
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawPixels) error = %v", err)
	}

	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew pixels with varying alpha values")
}

func TestIntegration_DrawPixels_OnCustomLayer(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-custom-layer.aseprite")
	createScript := gen.CreateCanvas(50, 50, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add a custom layer
	layerName := "Drawing Layer"
	addLayerScript := gen.AddLayer(layerName)
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add layer: %v", err)
	}

	// Draw pixels on the custom layer
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 5, Y: 5}, Color: aseprite.Color{R: 0, G: 255, B: 255, A: 255}},
		{Point: aseprite.Point{X: 6, Y: 6}, Color: aseprite.Color{R: 0, G: 255, B: 255, A: 255}},
	}

	drawScript := gen.DrawPixels(layerName, 1, pixels)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawPixels) error = %v", err)
	}

	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew pixels on custom layer '%s'", layerName)
}

func TestIntegration_DrawLine(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-line.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a diagonal line
	drawScript := gen.DrawLine("Layer 1", 1, 10, 10, 50, 50, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, 2)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawLine) error = %v", err)
	}

	if !strings.Contains(output, "Line drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew line successfully")
}

func TestIntegration_DrawRectangle_Outline(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-rect-outline.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw rectangle outline
	drawScript := gen.DrawRectangle("Layer 1", 1, 10, 10, 40, 30, aseprite.Color{R: 0, G: 255, B: 0, A: 255}, false)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawRectangle) error = %v", err)
	}

	if !strings.Contains(output, "Rectangle drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew rectangle outline successfully")
}

func TestIntegration_DrawRectangle_Filled(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-rect-filled.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw filled rectangle
	drawScript := gen.DrawRectangle("Layer 1", 1, 20, 20, 30, 20, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, true)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawRectangle) error = %v", err)
	}

	if !strings.Contains(output, "Rectangle drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew filled rectangle successfully")
}

func TestIntegration_DrawCircle_Outline(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-circle-outline.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw circle outline
	drawScript := gen.DrawCircle("Layer 1", 1, 50, 50, 20, aseprite.Color{R: 255, G: 255, B: 0, A: 255}, false)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawCircle) error = %v", err)
	}

	if !strings.Contains(output, "Circle drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew circle outline successfully")
}

func TestIntegration_DrawCircle_Filled(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-draw-circle-filled.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw filled circle
	drawScript := gen.DrawCircle("Layer 1", 1, 30, 30, 15, aseprite.Color{R: 255, G: 0, B: 255, A: 255}, true)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawCircle) error = %v", err)
	}

	if !strings.Contains(output, "Circle drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew filled circle successfully")
}