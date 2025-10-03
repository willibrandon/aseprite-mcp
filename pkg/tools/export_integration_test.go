//go:build integration
// +build integration

package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

// Integration tests for export_sprite tool with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools

func TestIntegration_ExportSprite_PNG(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with some content
	spritePath := testutil.TempSpritePath(t, "test-export-png.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a circle for visual content
	drawScript := gen.DrawCircle("Layer 1", 1, 50, 50, 30, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw circle: %v", err)
	}

	// Export to PNG
	outputPath := filepath.Join(t.TempDir(), "output.png")
	exportScript := gen.ExportSprite(outputPath, 0)
	output, err := client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ExportSprite) error = %v", err)
	}

	if !strings.Contains(output, "Exported successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify file was created
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("Exported file is empty")
	}

	t.Logf("✓ Exported to PNG successfully (%d bytes)", fileInfo.Size())
}

func TestIntegration_ExportSprite_GIF(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-export-gif.aseprite")
	createScript := gen.CreateCanvas(50, 50, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add multiple frames for animation
	for i := 0; i < 3; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err = client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame %d: %v", i+1, err)
		}
	}

	// Export to GIF
	outputPath := filepath.Join(t.TempDir(), "animation.gif")
	exportScript := gen.ExportSprite(outputPath, 0)
	output, err := client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ExportSprite) error = %v", err)
	}

	if !strings.Contains(output, "Exported successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify file was created
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("Exported file is empty")
	}

	t.Logf("✓ Exported to GIF successfully (%d bytes)", fileInfo.Size())
}

func TestIntegration_ExportSprite_SpecificFrame(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-export-frame.aseprite")
	createScript := gen.CreateCanvas(50, 50, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add a second frame
	addFrameScript := gen.AddFrame(100)
	_, err = client.ExecuteLua(ctx, addFrameScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add frame: %v", err)
	}

	// Draw different content on frame 2
	drawScript := gen.DrawRectangle("Layer 1", 2, 10, 10, 30, 30, aseprite.Color{R: 0, G: 255, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw rectangle: %v", err)
	}

	// Export only frame 2
	outputPath := filepath.Join(t.TempDir(), "frame2.png")
	exportScript := gen.ExportSprite(outputPath, 2)
	output, err := client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ExportSprite) error = %v", err)
	}

	if !strings.Contains(output, "Exported successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify file was created
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("Exported file is empty")
	}

	t.Logf("✓ Exported specific frame successfully (%d bytes)", fileInfo.Size())
}

func TestIntegration_ExportSprite_JPG(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with content
	spritePath := testutil.TempSpritePath(t, "test-export-jpg.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Fill with color
	fillScript := gen.FillArea("Layer 1", 1, 50, 50, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, 0, false)
	_, err = client.ExecuteLua(ctx, fillScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to fill area: %v", err)
	}

	// Export to JPG
	outputPath := filepath.Join(t.TempDir(), "output.jpg")
	exportScript := gen.ExportSprite(outputPath, 0)
	output, err := client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ExportSprite) error = %v", err)
	}

	if !strings.Contains(output, "Exported successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify file was created
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("Exported file is empty")
	}

	t.Logf("✓ Exported to JPG successfully (%d bytes)", fileInfo.Size())
}

func TestIntegration_ExportSprite_SpecificFrameWithContent(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a 16x16 canvas
	spritePath := testutil.TempSpritePath(t, "test-export-content.aseprite")
	createScript := gen.CreateCanvas(16, 16, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a filled red rectangle covering the entire canvas
	drawScript := gen.DrawRectangle("Layer 1", 1, 0, 0, 16, 16, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw rectangle: %v", err)
	}

	// Export frame 1 to PNG
	outputPath := filepath.Join(t.TempDir(), "red_square.png")
	exportScript := gen.ExportSprite(outputPath, 1)
	output, err := client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ExportSprite) error = %v", err)
	}

	if !strings.Contains(output, "Exported successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify file was created
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}

	// A 16x16 red PNG should be at least 95 bytes
	// Blank/nearly empty PNGs are typically 88-92 bytes
	minExpectedSize := int64(95)
	if fileInfo.Size() < minExpectedSize {
		t.Errorf("Exported PNG is too small (%d bytes), expected at least %d bytes. This indicates the export produced a blank image.", fileInfo.Size(), minExpectedSize)
	}

	// Verify the PNG can be decoded and has correct dimensions
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open exported PNG: %v", err)
	}
	defer file.Close()

	img, format, err := testutil.DecodeImage(file)
	if err != nil {
		t.Fatalf("Failed to decode exported PNG: %v", err)
	}

	if format != "png" {
		t.Errorf("Expected PNG format, got: %s", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 16 || bounds.Dy() != 16 {
		t.Errorf("Expected 16x16 dimensions, got: %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify at least some pixels are red (not all transparent/black)
	redPixels := 0
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// RGBA returns values in range 0-65535, so red would be (65535, 0, 0, 65535)
			if r > 50000 && g < 10000 && b < 10000 && a > 50000 {
				redPixels++
			}
		}
	}

	if redPixels < 200 { // Expect most of the 256 pixels to be red
		t.Errorf("Expected red pixels in exported image, found only %d red pixels out of 256", redPixels)
	}

	t.Logf("✓ Exported frame with content successfully (%d bytes, %d red pixels)", fileInfo.Size(), redPixels)
}

func TestIntegration_ExportSpritesheet_Horizontal(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with multiple frames
	spritePath := testutil.TempSpritePath(t, "test-spritesheet-h.aseprite")
	createScript := gen.CreateCanvas(32, 32, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add 2 more frames (total 3 frames)
	for i := 0; i < 2; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err = client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame %d: %v", i+1, err)
		}
	}

	// Draw different content on each frame
	colors := []aseprite.Color{
		{R: 255, G: 0, B: 0, A: 255}, // Red
		{R: 0, G: 255, B: 0, A: 255}, // Green
		{R: 0, G: 0, B: 255, A: 255}, // Blue
	}
	for i := 0; i < 3; i++ {
		drawScript := gen.DrawCircle("Layer 1", i+1, 16, 16, 10, colors[i], true, false)
		_, err = client.ExecuteLua(ctx, drawScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to draw on frame %d: %v", i+1, err)
		}
	}

	// Export spritesheet with horizontal layout
	outputPath := filepath.Join(t.TempDir(), "spritesheet-h.png")
	exportScript := gen.ExportSpritesheet(outputPath, "horizontal", 2, false)
	output, err := client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ExportSpritesheet) error = %v", err)
	}

	// Parse JSON output
	var result ExportSpritesheetOutput
	if err := parseJSON(output, &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.FrameCount != 3 {
		t.Errorf("Expected 3 frames, got %d", result.FrameCount)
	}

	if result.SpritesheetPath != outputPath {
		t.Errorf("Expected path %s, got %s", outputPath, result.SpritesheetPath)
	}

	// Verify file was created
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Spritesheet file not found: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("Spritesheet file is empty")
	}

	t.Logf("✓ Exported horizontal spritesheet successfully (%d frames, %d bytes)", result.FrameCount, fileInfo.Size())
}

func TestIntegration_ExportSpritesheet_WithJSON(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with 2 frames
	spritePath := testutil.TempSpritePath(t, "test-spritesheet-json.aseprite")
	createScript := gen.CreateCanvas(16, 16, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	addFrameScript := gen.AddFrame(100)
	_, err = client.ExecuteLua(ctx, addFrameScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add frame: %v", err)
	}

	// Export spritesheet with JSON metadata
	outputPath := filepath.Join(t.TempDir(), "spritesheet-with-json.png")
	exportScript := gen.ExportSpritesheet(outputPath, "packed", 0, true)
	output, err := client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ExportSpritesheet) error = %v", err)
	}

	// Parse JSON output
	var result ExportSpritesheetOutput
	if err := parseJSON(output, &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.MetadataPath == nil {
		t.Error("Expected metadata path, got nil")
	} else {
		// Verify JSON file was created
		jsonInfo, err := os.Stat(*result.MetadataPath)
		if err != nil {
			t.Errorf("JSON metadata file not found: %v", err)
		} else if jsonInfo.Size() == 0 {
			t.Error("JSON metadata file is empty")
		} else {
			t.Logf("✓ JSON metadata created (%d bytes)", jsonInfo.Size())
		}
	}

	t.Logf("✓ Exported spritesheet with JSON metadata successfully (%d frames)", result.FrameCount)
}

func TestIntegration_ImportImage(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-import.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// First export a small image to use for import
	drawScript := gen.DrawRectangle("Layer 1", 1, 10, 10, 20, 20, aseprite.Color{R: 255, G: 255, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw rectangle: %v", err)
	}

	imagePath := filepath.Join(t.TempDir(), "import-source.png")
	exportScript := gen.ExportSprite(imagePath, 1)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to export source image: %v", err)
	}

	// Now import that image back as a new layer
	importScript := gen.ImportImage(imagePath, "Imported Layer", 1, nil, nil)
	output, err := client.ExecuteLua(ctx, importScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ImportImage) error = %v", err)
	}

	if !strings.Contains(output, "Image imported successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify the layer was created by checking sprite info
	infoScript := gen.GetSpriteInfo()
	infoOutput, err := client.ExecuteLua(ctx, infoScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to get sprite info: %v", err)
	}

	if !strings.Contains(infoOutput, "Imported Layer") {
		t.Error("Imported layer not found in sprite")
	}

	t.Logf("✓ Imported image successfully")
}

func TestIntegration_ImportImage_WithPosition(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-import-pos.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Create an image to import (use test data)
	testdataDir := filepath.Join("..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "Mona_Lisa.jpg")

	// Import with custom position
	x := 25
	y := 25
	importScript := gen.ImportImage(imagePath, "Reference", 1, &x, &y)
	output, err := client.ExecuteLua(ctx, importScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ImportImage) error = %v", err)
	}

	if !strings.Contains(output, "Image imported successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Imported image with custom position successfully")
}

func TestIntegration_SaveAs(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with some content
	spritePath := testutil.TempSpritePath(t, "test-save-original.aseprite")
	createScript := gen.CreateCanvas(32, 32, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw something
	drawScript := gen.DrawCircle("Layer 1", 1, 16, 16, 10, aseprite.Color{R: 0, G: 255, B: 255, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw: %v", err)
	}

	// Save as new file
	newPath := filepath.Join(t.TempDir(), "saved-copy.aseprite")
	saveAsScript := gen.SaveAs(newPath)
	output, err := client.ExecuteLua(ctx, saveAsScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(SaveAs) error = %v", err)
	}

	// Parse JSON output
	var result SaveAsOutput
	if err := parseJSON(output, &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if !result.Success {
		t.Error("Expected success=true")
	}

	if result.FilePath != newPath {
		t.Errorf("Expected file path %s, got %s", newPath, result.FilePath)
	}

	// Verify new file was created
	fileInfo, err := os.Stat(newPath)
	if err != nil {
		t.Fatalf("Saved file not found: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("Saved file is empty")
	}

	// Verify we can open the new file
	infoScript := gen.GetSpriteInfo()
	infoOutput, err := client.ExecuteLua(ctx, infoScript, newPath)
	if err != nil {
		t.Fatalf("Failed to open saved file: %v", err)
	}

	if !strings.Contains(infoOutput, "32") {
		t.Error("Saved file doesn't have expected dimensions")
	}

	t.Logf("✓ Saved sprite to new file successfully (%d bytes)", fileInfo.Size())
}
