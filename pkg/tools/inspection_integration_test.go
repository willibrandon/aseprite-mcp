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

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

// Integration tests for get_pixels tool with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools

func TestIntegration_GetPixels_SinglePixel(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-getpixels-single.aseprite")
	createScript := gen.CreateCanvas(10, 10, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a single red pixel at (5, 5)
	pixels := []aseprite.Pixel{
		{
			Point: aseprite.Point{X: 5, Y: 5},
			Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255},
		},
	}
	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw pixel: %v", err)
	}

	// Get the pixel back
	getScript := gen.GetPixels("Layer 1", 1, 5, 5, 1, 1)
	output, err := client.ExecuteLua(ctx, getScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
	}

	// Parse output
	var pixelData []PixelData
	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse pixel data: %v, output: %s", err, output)
	}

	if len(pixelData) != 1 {
		t.Errorf("Expected 1 pixel, got %d", len(pixelData))
	}

	if len(pixelData) > 0 {
		p := pixelData[0]
		if p.X != 5 || p.Y != 5 {
			t.Errorf("Expected pixel at (5,5), got (%d,%d)", p.X, p.Y)
		}
		// Color should be #FF0000FF (red with full alpha)
		if !strings.Contains(strings.ToUpper(p.Color), "FF0000") {
			t.Errorf("Expected red color (#FF0000FF), got %s", p.Color)
		}
	}

	t.Logf("✓ Read single pixel successfully: %+v", pixelData[0])
}

func TestIntegration_GetPixels_MultiplePixels(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-getpixels-multiple.aseprite")
	createScript := gen.CreateCanvas(10, 10, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a 3x3 red rectangle
	drawScript := gen.DrawRectangle("Layer 1", 1, 2, 2, 3, 3, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw rectangle: %v", err)
	}

	// Get the 3x3 region
	getScript := gen.GetPixels("Layer 1", 1, 2, 2, 3, 3)
	output, err := client.ExecuteLua(ctx, getScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
	}

	// Parse output
	var pixelData []PixelData
	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse pixel data: %v, output: %s", err, output)
	}

	if len(pixelData) != 9 {
		t.Errorf("Expected 9 pixels (3x3), got %d", len(pixelData))
	}

	// Verify all pixels are red
	redCount := 0
	for _, p := range pixelData {
		if strings.Contains(strings.ToUpper(p.Color), "FF0000") {
			redCount++
		}
	}

	if redCount != 9 {
		t.Errorf("Expected 9 red pixels, got %d", redCount)
	}

	t.Logf("✓ Read 3x3 region successfully: %d red pixels", redCount)
}

func TestIntegration_GetPixels_EmptyRegion(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-getpixels-empty.aseprite")
	createScript := gen.CreateCanvas(10, 10, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Get pixels from an empty canvas (should return transparent pixels)
	getScript := gen.GetPixels("Layer 1", 1, 0, 0, 5, 5)
	output, err := client.ExecuteLua(ctx, getScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
	}

	// Parse output
	var pixelData []PixelData
	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse pixel data: %v, output: %s", err, output)
	}

	// Empty canvas may return 0 pixels (transparent) or 25 pixels (transparent pixels)
	// depending on how Aseprite handles empty cels
	t.Logf("✓ Read from empty region: %d pixels returned", len(pixelData))
}

func TestIntegration_GetPixels_PartialOverlap(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a 20x20 canvas
	spritePath := testutil.TempSpritePath(t, "test-getpixels-partial.aseprite")
	createScript := gen.CreateCanvas(20, 20, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a 5x5 green square at (10, 10)
	drawScript := gen.DrawRectangle("Layer 1", 1, 10, 10, 5, 5, aseprite.Color{R: 0, G: 255, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw rectangle: %v", err)
	}

	// Get a 10x10 region starting at (8, 8) - partially overlapping the green square
	getScript := gen.GetPixels("Layer 1", 1, 8, 8, 10, 10)
	output, err := client.ExecuteLua(ctx, getScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
	}

	// Parse output
	var pixelData []PixelData
	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse pixel data: %v, output: %s", err, output)
	}

	// Count green pixels (should be 25 from the 5x5 square)
	greenCount := 0
	for _, p := range pixelData {
		if strings.Contains(strings.ToUpper(p.Color), "00FF00") {
			greenCount++
		}
	}

	if greenCount != 25 {
		t.Errorf("Expected 25 green pixels, got %d", greenCount)
	}

	t.Logf("✓ Read partially overlapping region: %d green pixels out of %d total", greenCount, len(pixelData))
}

func TestIntegration_GetPixels_DifferentColorModes(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	tests := []struct {
		name      string
		colorMode aseprite.ColorMode
	}{
		{"RGB mode", aseprite.ColorModeRGB},
		{"Grayscale mode", aseprite.ColorModeGrayscale},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a canvas
			spritePath := testutil.TempSpritePath(t, "test-getpixels-"+tt.name+".aseprite")
			createScript := gen.CreateCanvas(10, 10, tt.colorMode, spritePath)
			_, err := client.ExecuteLua(ctx, createScript, "")
			if err != nil {
				t.Fatalf("Failed to create canvas: %v", err)
			}
			defer os.Remove(spritePath)

			// Draw a pixel
			pixels := []aseprite.Pixel{
				{
					Point: aseprite.Point{X: 5, Y: 5},
					Color: aseprite.Color{R: 255, G: 255, B: 255, A: 255},
				},
			}
			drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
			_, err = client.ExecuteLua(ctx, drawScript, spritePath)
			if err != nil {
				t.Fatalf("Failed to draw pixel: %v", err)
			}

			// Get the pixel
			getScript := gen.GetPixels("Layer 1", 1, 5, 5, 1, 1)
			output, err := client.ExecuteLua(ctx, getScript, spritePath)
			if err != nil {
				t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
			}

			// Parse output
			var pixelData []PixelData
			if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
				t.Fatalf("Failed to parse pixel data: %v", err)
			}

			if len(pixelData) != 1 {
				t.Errorf("Expected 1 pixel, got %d", len(pixelData))
			}

			t.Logf("✓ %s: Read pixel successfully: %+v", tt.name, pixelData[0])
		})
	}
}

func TestIntegration_GetPixels_LargeRegion(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a 100x100 canvas
	spritePath := testutil.TempSpritePath(t, "test-getpixels-large.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Fill the entire canvas with blue
	fillScript := gen.FillArea("Layer 1", 1, 50, 50, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, 0, false)
	_, err = client.ExecuteLua(ctx, fillScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to fill area: %v", err)
	}

	// Get a 50x50 region
	getScript := gen.GetPixels("Layer 1", 1, 25, 25, 50, 50)
	output, err := client.ExecuteLua(ctx, getScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(GetPixels) error = %v", err)
	}

	// Parse output
	var pixelData []PixelData
	if err := json.Unmarshal([]byte(output), &pixelData); err != nil {
		t.Fatalf("Failed to parse pixel data: %v", err)
	}

	if len(pixelData) != 2500 {
		t.Errorf("Expected 2500 pixels (50x50), got %d", len(pixelData))
	}

	// Count blue pixels
	blueCount := 0
	for _, p := range pixelData {
		if strings.Contains(strings.ToUpper(p.Color), "0000FF") {
			blueCount++
		}
	}

	if blueCount != 2500 {
		t.Errorf("Expected 2500 blue pixels, got %d", blueCount)
	}

	t.Logf("✓ Read large 50x50 region successfully: %d blue pixels", blueCount)
}