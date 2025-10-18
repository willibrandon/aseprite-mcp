//go:build integration
// +build integration

package aseprite

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
)

// Integration tests require real Aseprite installation.
// Config file must be present at ~/.config/pixel-mcp/config.json
// with aseprite_path pointing to real Aseprite executable.
//
// Run with: go test -tags=integration ./pkg/aseprite

func TestIntegration_GetVersion(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)

	ctx := context.Background()
	version, err := client.GetVersion(ctx)
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}

	if version == "" {
		t.Error("GetVersion() returned empty version")
	}

	// Verify version contains "Aseprite"
	if !strings.Contains(version, "Aseprite") {
		t.Errorf("GetVersion() = %q, expected to contain 'Aseprite'", version)
	}

	t.Logf("✓ Aseprite version: %s", version)
}

func TestIntegration_CreateCanvas(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()

	// Generate a temp file path for the sprite
	spritePath := testutil.TempSpritePath(t, "test-canvas.aseprite")

	// Generate script to create canvas with real Aseprite
	script := gen.CreateCanvas(100, 100, ColorModeRGB, spritePath)

	// Execute with real Aseprite
	ctx := context.Background()
	output, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("ExecuteLua() error = %v", err)
	}

	// Output should confirm the path
	outputPath := strings.TrimSpace(output)
	if outputPath != spritePath {
		t.Errorf("Output path mismatch: got %q, want %q", outputPath, spritePath)
	}

	// Verify file was created by real Aseprite
	if _, err := os.Stat(spritePath); os.IsNotExist(err) {
		t.Errorf("sprite file was not created at %s", spritePath)
	} else {
		// Clean up created sprite
		defer os.Remove(spritePath)
		t.Logf("✓ Created sprite at: %s", spritePath)
	}
}

func TestIntegration_AddLayer(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// First create a canvas
	spritePath := testutil.TempSpritePath(t, "test-addlayer.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Now add a layer
	addLayerScript := gen.AddLayer("Test Layer")
	output, err := client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(AddLayer) error = %v", err)
	}

	if !strings.Contains(output, "Layer added successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Added layer to sprite")
}

func TestIntegration_DrawPixels(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-drawpixels.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw some pixels
	pixels := []Pixel{
		{Point: Point{X: 10, Y: 10}, Color: NewColorRGB(255, 0, 0)}, // Red
		{Point: Point{X: 20, Y: 20}, Color: NewColorRGB(0, 255, 0)}, // Green
		{Point: Point{X: 30, Y: 30}, Color: NewColorRGB(0, 0, 255)}, // Blue
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DrawPixels) error = %v", err)
	}

	if !strings.Contains(output, "Pixels drawn successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Drew %d pixels", len(pixels))
}

func TestIntegration_ExportSprite(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-export.aseprite")
	createScript := gen.CreateCanvas(50, 50, ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Export to PNG
	exportPath := testutil.TempSpritePath(t, "test-export.png")
	exportScript := gen.ExportSprite(exportPath, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ExportSprite) error = %v", err)
	}

	// Verify PNG was created
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Errorf("exported PNG file was not created at %s", exportPath)
	} else {
		defer os.Remove(exportPath)
		t.Logf("✓ Exported sprite to: %s", exportPath)
	}
}
