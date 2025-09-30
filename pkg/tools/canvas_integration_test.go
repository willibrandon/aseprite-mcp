//go:build integration
// +build integration

package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/sinks"
)

// Integration tests for create_canvas tool with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools

func TestIntegration_CreateCanvas_RGB(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithSink(sinks.NewMemorySink()))

	// Generate filename
	filename := testutil.TempSpritePath(t, "test-rgb.aseprite")

	logger.Information("Testing RGB canvas creation", "filename", filename)

	// Generate and execute script
	script := gen.CreateCanvas(800, 600, aseprite.ColorModeRGB, filename)
	ctx := context.Background()
	output, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("ExecuteLua() error = %v", err)
	}

	// Verify output path
	outputPath := strings.TrimSpace(output)
	if outputPath != filename {
		t.Errorf("Output path mismatch: got %q, want %q", outputPath, filename)
	}

	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Sprite file was not created at %s", filename)
	} else {
		defer os.Remove(filename)
		logger.Information("Created RGB sprite successfully", "path", filename)
		t.Logf("✓ Created RGB sprite at: %s", filename)
	}
}

func TestIntegration_CreateCanvas_Grayscale(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	filename := testutil.TempSpritePath(t, "test-grayscale.aseprite")

	script := gen.CreateCanvas(100, 100, aseprite.ColorModeGrayscale, filename)
	ctx := context.Background()
	_, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("ExecuteLua() error = %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Sprite file was not created")
	} else {
		defer os.Remove(filename)
		t.Logf("✓ Created grayscale sprite at: %s", filename)
	}
}

func TestIntegration_CreateCanvas_Indexed(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	filename := testutil.TempSpritePath(t, "test-indexed.aseprite")

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, filename)
	ctx := context.Background()
	_, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("ExecuteLua() error = %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Sprite file was not created")
	} else {
		defer os.Remove(filename)
		t.Logf("✓ Created indexed sprite at: %s", filename)
	}
}

func TestIntegration_CreateCanvas_MinimumSize(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	filename := testutil.TempSpritePath(t, "test-1x1.aseprite")

	script := gen.CreateCanvas(1, 1, aseprite.ColorModeRGB, filename)
	ctx := context.Background()
	_, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("ExecuteLua() error = %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Sprite file was not created")
	} else {
		defer os.Remove(filename)
		t.Logf("✓ Created 1x1 sprite at: %s", filename)
	}
}

func TestIntegration_CreateCanvas_LargeSize(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()

	filename := testutil.TempSpritePath(t, "test-large.aseprite")

	// Test a reasonably large size (not max to avoid timeout in CI)
	script := gen.CreateCanvas(2048, 2048, aseprite.ColorModeRGB, filename)
	ctx := context.Background()
	_, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("ExecuteLua() error = %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Sprite file was not created")
	} else {
		defer os.Remove(filename)
		t.Logf("✓ Created 2048x2048 sprite at: %s", filename)
	}
}

func TestIntegration_AddLayer(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// First create a canvas
	spritePath := testutil.TempSpritePath(t, "test-addlayer.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
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

func TestIntegration_AddMultipleLayers(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-multilayer.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add multiple layers
	layers := []string{"Background", "Midground", "Foreground"}
	for _, layerName := range layers {
		addLayerScript := gen.AddLayer(layerName)
		_, err := client.ExecuteLua(ctx, addLayerScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add layer %s: %v", layerName, err)
		}
	}

	t.Logf("✓ Added %d layers to sprite", len(layers))
}

func TestIntegration_AddFrame(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-addframe.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add a frame with 100ms duration
	addFrameScript := gen.AddFrame(100)
	output, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(AddFrame) error = %v", err)
	}

	// Parse frame count from output
	var frameCount int
	_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &frameCount)
	if err != nil {
		t.Fatalf("Failed to parse frame count: %v", err)
	}

	if frameCount < 2 {
		t.Errorf("Expected at least 2 frames, got %d", frameCount)
	}

	t.Logf("✓ Added frame (total frames: %d)", frameCount)
}

func TestIntegration_AddMultipleFrames(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-animation.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add multiple frames with different durations
	durations := []int{100, 150, 200, 100}
	for _, duration := range durations {
		addFrameScript := gen.AddFrame(duration)
		_, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame with duration %dms: %v", duration, err)
		}
	}

	t.Logf("✓ Created animation with %d frames", len(durations)+1) // +1 for initial frame
}

func TestIntegration_AddLayerSpecialCharacters(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-special-chars.aseprite")
	createScript := gen.CreateCanvas(50, 50, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add layer with special characters
	layerName := "Layer_1-Test (Copy)"
	addLayerScript := gen.AddLayer(layerName)
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add layer with special characters: %v", err)
	}

	t.Logf("✓ Added layer with special characters: %s", layerName)
}
