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

// Integration tests for animation tools with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools

func TestIntegration_SetFrameDuration(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithSink(sinks.NewMemorySink()))
	ctx := context.Background()

	// Create test sprite with 3 frames
	spritePath := testutil.TempSpritePath(t, "test-frame-duration.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add 2 more frames
	for i := 0; i < 2; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame: %v", err)
		}
	}

	logger.Information("Testing set frame duration", "sprite", spritePath)

	// Test setting frame 1 duration
	script := gen.SetFrameDuration(1, 200)
	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(SetFrameDuration) error = %v", err)
	}

	if !strings.Contains(output, "Frame duration set successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Set frame 1 duration to 200ms")

	// Test setting frame 2 duration
	script = gen.SetFrameDuration(2, 150)
	output, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(SetFrameDuration) error = %v", err)
	}

	if !strings.Contains(output, "Frame duration set successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Set frame 2 duration to 150ms")
}

func TestIntegration_SetFrameDuration_InvalidFrame(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "test-invalid-frame.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Try to set duration for non-existent frame
	script := gen.SetFrameDuration(99, 100)
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err == nil {
		t.Error("Expected error for invalid frame number, got nil")
	}

	t.Logf("✓ Correctly rejected invalid frame number")
}

func TestIntegration_CreateTag_Forward(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithSink(sinks.NewMemorySink()))
	ctx := context.Background()

	// Create test sprite with 5 frames
	spritePath := testutil.TempSpritePath(t, "test-tag-forward.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add 4 more frames
	for i := 0; i < 4; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame: %v", err)
		}
	}

	logger.Information("Testing create tag with forward direction", "sprite", spritePath)

	// Create forward tag
	script := gen.CreateTag("walk", 1, 3, "forward")
	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(CreateTag) error = %v", err)
	}

	if !strings.Contains(output, "Tag created successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Created forward tag 'walk' (frames 1-3)")
}

func TestIntegration_CreateTag_Reverse(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create test sprite with 5 frames
	spritePath := testutil.TempSpritePath(t, "test-tag-reverse.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add 4 more frames
	for i := 0; i < 4; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame: %v", err)
		}
	}

	// Create reverse tag
	script := gen.CreateTag("reverse_walk", 2, 4, "reverse")
	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(CreateTag) error = %v", err)
	}

	if !strings.Contains(output, "Tag created successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Created reverse tag 'reverse_walk' (frames 2-4)")
}

func TestIntegration_CreateTag_Pingpong(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create test sprite with 5 frames
	spritePath := testutil.TempSpritePath(t, "test-tag-pingpong.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add 4 more frames
	for i := 0; i < 4; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame: %v", err)
		}
	}

	// Create pingpong tag
	script := gen.CreateTag("idle", 1, 5, "pingpong")
	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(CreateTag) error = %v", err)
	}

	if !strings.Contains(output, "Tag created successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Created pingpong tag 'idle' (frames 1-5)")
}

func TestIntegration_CreateTag_InvalidRange(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create test sprite with 3 frames
	spritePath := testutil.TempSpritePath(t, "test-tag-invalid.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	for i := 0; i < 2; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame: %v", err)
		}
	}

	// Try to create tag with range exceeding sprite frames
	script := gen.CreateTag("overflow", 1, 10, "forward")
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err == nil {
		t.Error("Expected error for invalid frame range, got nil")
	}

	t.Logf("✓ Correctly rejected invalid frame range")
}

func TestIntegration_DuplicateFrame_AtEnd(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithSink(sinks.NewMemorySink()))
	ctx := context.Background()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "test-duplicate-end.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add a layer and draw content
	addLayerScript := gen.AddLayer("Layer 1")
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add layer: %v", err)
	}

	pixels := []aseprite.Pixel{
		{
			Point: aseprite.Point{X: 10, Y: 10},
			Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255},
		},
		{
			Point: aseprite.Point{X: 20, Y: 20},
			Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255},
		},
	}
	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw pixels: %v", err)
	}

	logger.Information("Testing duplicate frame at end", "sprite", spritePath)

	// Duplicate frame 1 at end (insertAfter = 0)
	script := gen.DuplicateFrame(1, 0)
	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DuplicateFrame) error = %v", err)
	}

	// Parse new frame number
	var newFrameNumber int
	_, parseErr := fmt.Sscanf(strings.TrimSpace(output), "%d", &newFrameNumber)
	if parseErr != nil {
		t.Fatalf("Failed to parse frame number from output %q: %v", output, parseErr)
	}

	if newFrameNumber < 2 {
		t.Errorf("Expected frame number >= 2, got %d", newFrameNumber)
	}

	t.Logf("✓ Duplicated frame 1 at end (new frame: %d)", newFrameNumber)
}

func TestIntegration_DuplicateFrame_AfterSpecific(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create test sprite with 3 frames
	spritePath := testutil.TempSpritePath(t, "test-duplicate-after.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add 2 more frames
	for i := 0; i < 2; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame: %v", err)
		}
	}

	// Duplicate frame 1 after frame 1
	script := gen.DuplicateFrame(1, 1)
	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(DuplicateFrame) error = %v", err)
	}

	// Parse new frame number
	var newFrameNumber int
	_, parseErr := fmt.Sscanf(strings.TrimSpace(output), "%d", &newFrameNumber)
	if parseErr != nil {
		t.Fatalf("Failed to parse frame number from output %q: %v", output, parseErr)
	}

	if newFrameNumber != 2 {
		t.Errorf("Expected new frame at position 2, got %d", newFrameNumber)
	}

	t.Logf("✓ Duplicated frame 1 after frame 1 (new frame: %d)", newFrameNumber)
}

func TestIntegration_DuplicateFrame_InvalidSource(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "test-duplicate-invalid.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Try to duplicate non-existent frame
	script := gen.DuplicateFrame(99, 0)
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err == nil {
		t.Error("Expected error for invalid source frame, got nil")
	}

	t.Logf("✓ Correctly rejected invalid source frame")
}

func TestIntegration_LinkCel(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithSink(sinks.NewMemorySink()))
	ctx := context.Background()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "test-link-cel.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add a layer
	addLayerScript := gen.AddLayer("Layer 1")
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add layer: %v", err)
	}

	// Add 2 more frames
	for i := 0; i < 2; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err := client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Failed to add frame: %v", err)
		}
	}

	// Draw content on frame 1
	pixels := []aseprite.Pixel{
		{
			Point: aseprite.Point{X: 15, Y: 15},
			Color: aseprite.Color{R: 255, G: 0, B: 255, A: 255},
		},
		{
			Point: aseprite.Point{X: 25, Y: 25},
			Color: aseprite.Color{R: 0, G: 255, B: 255, A: 255},
		},
	}
	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw pixels: %v", err)
	}

	logger.Information("Testing link cel", "sprite", spritePath)

	// Link frame 1 cel to frame 2
	script := gen.LinkCel("Layer 1", 1, 2)
	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(LinkCel) error = %v", err)
	}

	if !strings.Contains(output, "Cel linked successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Linked frame 1 cel to frame 2 on Layer 1")

	// Link frame 1 cel to frame 3
	script = gen.LinkCel("Layer 1", 1, 3)
	output, err = client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(LinkCel) error = %v", err)
	}

	if !strings.Contains(output, "Cel linked successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Linked frame 1 cel to frame 3 on Layer 1")
}

func TestIntegration_LinkCel_InvalidLayer(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "test-link-invalid-layer.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add a frame
	addFrameScript := gen.AddFrame(100)
	_, err = client.ExecuteLua(ctx, addFrameScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add frame: %v", err)
	}

	// Try to link cel on non-existent layer
	script := gen.LinkCel("NonExistentLayer", 1, 2)
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err == nil {
		t.Error("Expected error for invalid layer name, got nil")
	}

	t.Logf("✓ Correctly rejected invalid layer name")
}

func TestIntegration_LinkCel_InvalidFrame(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create test sprite
	spritePath := testutil.TempSpritePath(t, "test-link-invalid-frame.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Add a layer
	addLayerScript := gen.AddLayer("Layer 1")
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add layer: %v", err)
	}

	// Try to link cel with invalid source frame
	script := gen.LinkCel("Layer 1", 99, 1)
	_, err = client.ExecuteLua(ctx, script, spritePath)
	if err == nil {
		t.Error("Expected error for invalid source frame, got nil")
	}

	t.Logf("✓ Correctly rejected invalid source frame")
}
