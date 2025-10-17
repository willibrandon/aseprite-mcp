//go:build integration
// +build integration

package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

func TestIntegration_DownsampleImage(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	// Get path to test data (relative from pkg/tools/)
	testdataDir := filepath.Join("..", "..", "testdata")

	tests := []struct {
		name         string
		sourceFile   string
		targetWidth  int
		targetHeight int
	}{
		{
			name:         "downsample JPG to 128x194",
			sourceFile:   "Mona_Lisa.jpg",
			targetWidth:  128,
			targetHeight: 194,
		},
		{
			name:         "downsample Aseprite file to 64x97",
			sourceFile:   "Mona_Lisa.aseprite",
			targetWidth:  64,
			targetHeight: 97,
		},
		{
			name:         "downsample to 32x48",
			sourceFile:   "Mona_Lisa.jpg",
			targetWidth:  32,
			targetHeight: 48,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePath := filepath.Join(testdataDir, tt.sourceFile)
			outputPath := testutil.TempSpritePath(t, "downsampled.aseprite")

			// Generate and execute downsampling script
			script := gen.DownsampleImage(sourcePath, outputPath, tt.targetWidth, tt.targetHeight)
			ctx := context.Background()

			output, err := client.ExecuteLua(ctx, script, "")
			if err != nil {
				t.Fatalf("DownsampleImage() error = %v", err)
			}

			resultPath := strings.TrimSpace(output)

			// Verify output file exists
			if _, err := os.Stat(resultPath); os.IsNotExist(err) {
				t.Errorf("output file does not exist: %s", resultPath)
			} else {
				defer os.Remove(resultPath)
			}

			// Verify dimensions using get_sprite_info
			infoScript := gen.GetSpriteInfo()
			infoOutput, err := client.ExecuteLua(ctx, infoScript, resultPath)
			if err != nil {
				t.Fatalf("failed to get sprite info: %v", err)
			}

			var info GetSpriteInfoOutput
			if err := json.Unmarshal([]byte(infoOutput), &info); err != nil {
				t.Fatalf("failed to parse sprite info: %v", err)
			}

			if info.Width != tt.targetWidth {
				t.Errorf("output width: expected %d, got %d", tt.targetWidth, info.Width)
			}

			if info.Height != tt.targetHeight {
				t.Errorf("output height: expected %d, got %d", tt.targetHeight, info.Height)
			}

			t.Logf("✓ Downsampled to %dx%d successfully", info.Width, info.Height)
		})
	}
}

func TestIntegration_DownsampleImage_Quality(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	// Get path to test data
	testdataDir := filepath.Join("..", "..", "testdata")
	sourcePath := filepath.Join(testdataDir, "Mona_Lisa.jpg")
	outputPath := testutil.TempSpritePath(t, "mona-lisa-pixel-art.aseprite")
	defer os.Remove(outputPath)

	// Downsample Mona Lisa to pixel art size
	script := gen.DownsampleImage(sourcePath, outputPath, 128, 194)
	ctx := context.Background()

	output, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("failed to downsample image: %v", err)
	}

	resultPath := strings.TrimSpace(output)

	// Read a sample of pixels to verify quality
	getPixelsScript := gen.GetPixelsWithPagination("Layer 1", 1, 50, 50, 20, 20, 0, 400)
	pixelsOutput, err := client.ExecuteLua(ctx, getPixelsScript, resultPath)
	if err != nil {
		t.Fatalf("failed to get pixels: %v", err)
	}

	var pixels []PixelData
	if err := json.Unmarshal([]byte(pixelsOutput), &pixels); err != nil {
		t.Fatalf("failed to parse pixels: %v", err)
	}

	// Count unique colors to verify downsampling quality
	uniqueColors := make(map[string]bool)
	for _, p := range pixels {
		uniqueColors[p.Color] = true
	}

	// Expect significant color variation in the Mona Lisa
	if len(uniqueColors) < 10 {
		t.Errorf("expected at least 10 unique colors, got %d", len(uniqueColors))
	}

	t.Logf("✓ Found %d unique colors in %d pixel sample", len(uniqueColors), len(pixels))
}

func TestIntegration_FlipSprite(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "flip-test.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	ctx := context.Background()

	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("failed to create sprite: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a marker pixel to verify flip
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 10, Y: 20}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
	}
	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("failed to draw pixels: %v", err)
	}

	tests := []struct {
		name      string
		direction string
		target    string
	}{
		{"flip horizontal", "horizontal", "sprite"},
		{"flip vertical", "vertical", "sprite"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flipScript := gen.FlipSprite(tt.direction, tt.target)
			output, err := client.ExecuteLua(ctx, flipScript, spritePath)
			if err != nil {
				t.Fatalf("FlipSprite() error = %v", err)
			}

			if !strings.Contains(output, "flipped") {
				t.Errorf("expected success message, got: %s", output)
			}

			t.Logf("✓ Sprite flipped %s successfully", tt.direction)
		})
	}
}

func TestIntegration_RotateSprite(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "rotate-test.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	ctx := context.Background()

	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("failed to create sprite: %v", err)
	}
	defer os.Remove(spritePath)

	tests := []struct {
		name   string
		angle  int
		target string
	}{
		{"rotate 90 degrees", 90, "sprite"},
		{"rotate 180 degrees", 180, "sprite"},
		{"rotate 270 degrees", 270, "sprite"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rotateScript := gen.RotateSprite(tt.angle, tt.target)
			output, err := client.ExecuteLua(ctx, rotateScript, spritePath)
			if err != nil {
				t.Fatalf("RotateSprite() error = %v", err)
			}

			if !strings.Contains(output, "rotated") {
				t.Errorf("expected success message, got: %s", output)
			}

			t.Logf("✓ Sprite rotated %d degrees successfully", tt.angle)
		})
	}
}

func TestIntegration_ScaleSprite(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	tests := []struct {
		name      string
		scaleX    float64
		scaleY    float64
		algorithm string
		expectW   int
		expectH   int
	}{
		{"scale 2x nearest", 2.0, 2.0, "nearest", 128, 128},
		{"scale 0.5x bilinear", 0.5, 0.5, "bilinear", 32, 32},
		{"scale 1.5x rotsprite", 1.5, 1.5, "rotsprite", 96, 96},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh sprite for each test
			spritePath := testutil.TempSpritePath(t, "scale-test.aseprite")
			createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
			ctx := context.Background()

			_, err := client.ExecuteLua(ctx, createScript, "")
			if err != nil {
				t.Fatalf("failed to create sprite: %v", err)
			}
			defer os.Remove(spritePath)

			// Scale sprite
			scaleScript := gen.ScaleSprite(tt.scaleX, tt.scaleY, tt.algorithm)
			scaleOutput, err := client.ExecuteLua(ctx, scaleScript, spritePath)
			if err != nil {
				t.Fatalf("ScaleSprite() error = %v", err)
			}

			// Parse JSON output
			var result ScaleSpriteOutput
			if err := parseJSON(scaleOutput, &result); err != nil {
				t.Fatalf("failed to parse result: %v", err)
			}

			if !result.Success {
				t.Error("expected success=true")
			}

			if result.NewWidth != tt.expectW {
				t.Errorf("expected width=%d, got %d", tt.expectW, result.NewWidth)
			}

			if result.NewHeight != tt.expectH {
				t.Errorf("expected height=%d, got %d", tt.expectH, result.NewHeight)
			}

			t.Logf("✓ Sprite scaled to %dx%d using %s algorithm", result.NewWidth, result.NewHeight, tt.algorithm)
		})
	}
}

func TestIntegration_CropSprite(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "crop-test.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	ctx := context.Background()

	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("failed to create sprite: %v", err)
	}
	defer os.Remove(spritePath)

	// Crop to 50x50 region
	cropScript := gen.CropSprite(25, 25, 50, 50)
	cropOutput, err := client.ExecuteLua(ctx, cropScript, spritePath)
	if err != nil {
		t.Fatalf("CropSprite() error = %v", err)
	}

	if !strings.Contains(cropOutput, "cropped successfully") {
		t.Errorf("expected success message, got: %s", cropOutput)
	}

	// Verify new dimensions
	infoScript := gen.GetSpriteInfo()
	infoOutput, err := client.ExecuteLua(ctx, infoScript, spritePath)
	if err != nil {
		t.Fatalf("failed to get sprite info: %v", err)
	}

	var info GetSpriteInfoOutput
	if err := json.Unmarshal([]byte(infoOutput), &info); err != nil {
		t.Fatalf("failed to parse sprite info: %v", err)
	}

	if info.Width != 50 || info.Height != 50 {
		t.Errorf("expected 50x50, got %dx%d", info.Width, info.Height)
	}

	t.Logf("✓ Sprite cropped to %dx%d successfully", info.Width, info.Height)
}

func TestIntegration_ResizeCanvas(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	tests := []struct {
		name   string
		width  int
		height int
		anchor string
	}{
		{"resize center", 128, 128, "center"},
		{"resize top_left", 150, 100, "top_left"},
		{"resize bottom_right", 80, 120, "bottom_right"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh sprite for each test
			spritePath := testutil.TempSpritePath(t, "resize-canvas-test.aseprite")
			createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
			ctx := context.Background()

			_, err := client.ExecuteLua(ctx, createScript, "")
			if err != nil {
				t.Fatalf("failed to create sprite: %v", err)
			}
			defer os.Remove(spritePath)

			// Resize canvas
			resizeScript := gen.ResizeCanvas(tt.width, tt.height, tt.anchor)
			resizeOutput, err := client.ExecuteLua(ctx, resizeScript, spritePath)
			if err != nil {
				t.Fatalf("ResizeCanvas() error = %v", err)
			}

			if !strings.Contains(resizeOutput, "resized successfully") {
				t.Errorf("expected success message, got: %s", resizeOutput)
			}

			// Verify new dimensions
			infoScript := gen.GetSpriteInfo()
			infoOutput, err := client.ExecuteLua(ctx, infoScript, spritePath)
			if err != nil {
				t.Fatalf("failed to get sprite info: %v", err)
			}

			var info GetSpriteInfoOutput
			if err := json.Unmarshal([]byte(infoOutput), &info); err != nil {
				t.Fatalf("failed to parse sprite info: %v", err)
			}

			if info.Width != tt.width || info.Height != tt.height {
				t.Errorf("expected %dx%d, got %dx%d", tt.width, tt.height, info.Width, info.Height)
			}

			t.Logf("✓ Canvas resized to %dx%d with %s anchor", info.Width, info.Height, tt.anchor)
		})
	}
}

func TestIntegration_ApplyOutline(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "outline-test.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	ctx := context.Background()

	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("failed to create sprite: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a filled circle to apply outline to
	drawScript := gen.DrawCircle("Layer 1", 1, 32, 32, 10, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("failed to draw circle: %v", err)
	}

	// Apply outline
	outlineScript := gen.ApplyOutline("Layer 1", 1, aseprite.Color{R: 0, G: 0, B: 0, A: 255}, 2)
	outlineOutput, err := client.ExecuteLua(ctx, outlineScript, spritePath)
	if err != nil {
		t.Fatalf("ApplyOutline() error = %v", err)
	}

	if !strings.Contains(outlineOutput, "Outline applied successfully") {
		t.Errorf("expected success message, got: %s", outlineOutput)
	}

	t.Log("✓ Outline applied successfully")
}
