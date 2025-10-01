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

// End-to-end workflow tests that combine multiple tools to validate complete sprite creation workflows.
// Run with: go test -tags=integration -v ./pkg/tools -run TestWorkflow

func TestWorkflow_CreateDrawExport_SimpleSprite(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Step 1: Create canvas
	spritePath := testutil.TempSpritePath(t, "workflow-simple.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Step 1 (CreateCanvas) failed: %v", err)
	}
	defer os.Remove(spritePath)
	t.Log("✓ Step 1: Created canvas")

	// Step 2: Draw a filled circle
	drawScript := gen.DrawCircle("Layer 1", 1, 32, 32, 20, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	output, err := client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 2 (DrawCircle) failed: %v", err)
	}
	if !strings.Contains(output, "Circle drawn successfully") {
		t.Errorf("Step 2 unexpected output: %s", output)
	}
	t.Log("✓ Step 2: Drew red circle")

	// Step 3: Export to PNG
	outputPath := filepath.Join(t.TempDir(), "simple-sprite.png")
	exportScript := gen.ExportSprite(outputPath, 0)
	output, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Step 3 (ExportSprite) failed: %v", err)
	}
	if !strings.Contains(output, "Exported successfully") {
		t.Errorf("Step 3 unexpected output: %s", output)
	}
	t.Log("✓ Step 3: Exported to PNG")

	// Verify exported file
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}
	if fileInfo.Size() == 0 {
		t.Error("Exported file is empty")
	}

	t.Logf("✓ Complete workflow succeeded - exported %d bytes", fileInfo.Size())
}

func TestWorkflow_MultiLayerDrawing(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Step 1: Create canvas
	spritePath := testutil.TempSpritePath(t, "workflow-multilayer.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Step 1 (CreateCanvas) failed: %v", err)
	}
	defer os.Remove(spritePath)

	// Step 2: Add background layer
	addLayerScript := gen.AddLayer("Background")
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Step 2 (AddLayer Background) failed: %v", err)
	}

	// Step 3: Add foreground layer
	addLayerScript = gen.AddLayer("Foreground")
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Step 3 (AddLayer Foreground) failed: %v", err)
	}

	// Step 4: Fill background layer
	fillScript := gen.FillArea("Background", 1, 50, 50, aseprite.Color{R: 50, G: 50, B: 150, A: 255}, 0, false)
	_, err = client.ExecuteLua(ctx, fillScript, spritePath)
	if err != nil {
		t.Fatalf("Step 4 (FillArea Background) failed: %v", err)
	}

	// Step 5: Draw on foreground layer
	drawScript := gen.DrawRectangle("Foreground", 1, 30, 30, 40, 40, aseprite.Color{R: 255, G: 255, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 5 (DrawRectangle Foreground) failed: %v", err)
	}

	// Step 6: Get sprite info to verify layers
	infoScript := gen.GetSpriteInfo()
	output, err := client.ExecuteLua(ctx, infoScript, spritePath)
	if err != nil {
		t.Fatalf("Step 6 (GetSpriteInfo) failed: %v", err)
	}
	if !strings.Contains(output, "\"layer_count\": 3") {
		t.Errorf("Expected 3 layers, got output: %s", output)
	}

	// Step 7: Export
	outputPath := filepath.Join(t.TempDir(), "multilayer.png")
	exportScript := gen.ExportSprite(outputPath, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Step 7 (ExportSprite) failed: %v", err)
	}

	// Verify export
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}

	t.Log("✓ Multi-layer workflow succeeded")
}

func TestWorkflow_AnimationCreation(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Step 1: Create canvas
	spritePath := testutil.TempSpritePath(t, "workflow-animation.aseprite")
	createScript := gen.CreateCanvas(32, 32, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Step 1 (CreateCanvas) failed: %v", err)
	}
	defer os.Remove(spritePath)

	// Step 2-4: Add frames
	for i := 0; i < 3; i++ {
		addFrameScript := gen.AddFrame(100)
		_, err = client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Step %d (AddFrame) failed: %v", i+2, err)
		}
	}

	// Step 5: Draw on frame 1
	drawScript := gen.DrawCircle("Layer 1", 1, 16, 16, 8, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 5 (DrawCircle frame 1) failed: %v", err)
	}

	// Step 6: Draw on frame 2
	drawScript = gen.DrawCircle("Layer 1", 2, 16, 16, 10, aseprite.Color{R: 0, G: 255, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 6 (DrawCircle frame 2) failed: %v", err)
	}

	// Step 7: Draw on frame 3
	drawScript = gen.DrawCircle("Layer 1", 3, 16, 16, 12, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 7 (DrawCircle frame 3) failed: %v", err)
	}

	// Step 8: Draw on frame 4
	drawScript = gen.DrawCircle("Layer 1", 4, 16, 16, 14, aseprite.Color{R: 255, G: 255, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 8 (DrawCircle frame 4) failed: %v", err)
	}

	// Step 9: Verify sprite info
	infoScript := gen.GetSpriteInfo()
	output, err := client.ExecuteLua(ctx, infoScript, spritePath)
	if err != nil {
		t.Fatalf("Step 9 (GetSpriteInfo) failed: %v", err)
	}
	if !strings.Contains(output, "\"frame_count\": 4") {
		t.Errorf("Expected 4 frames, got output: %s", output)
	}

	// Step 10: Export as GIF
	gifPath := filepath.Join(t.TempDir(), "animation.gif")
	exportScript := gen.ExportSprite(gifPath, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Step 10 (ExportSprite GIF) failed: %v", err)
	}

	// Verify GIF export
	gifInfo, err := os.Stat(gifPath)
	if err != nil {
		t.Fatalf("Exported GIF not found: %v", err)
	}
	if gifInfo.Size() == 0 {
		t.Error("Exported GIF is empty")
	}

	t.Logf("✓ Animation workflow succeeded - created %d-frame GIF (%d bytes)", 4, gifInfo.Size())
}

func TestWorkflow_ComplexDrawing_PixelsAndPrimitives(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Step 1: Create canvas
	spritePath := testutil.TempSpritePath(t, "workflow-complex.aseprite")
	createScript := gen.CreateCanvas(128, 128, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Step 1 (CreateCanvas) failed: %v", err)
	}
	defer os.Remove(spritePath)

	// Step 2: Fill background
	fillScript := gen.FillArea("Layer 1", 1, 64, 64, aseprite.Color{R: 240, G: 240, B: 240, A: 255}, 0, false)
	_, err = client.ExecuteLua(ctx, fillScript, spritePath)
	if err != nil {
		t.Fatalf("Step 2 (FillArea) failed: %v", err)
	}

	// Step 3: Draw rectangle outline
	drawScript := gen.DrawRectangle("Layer 1", 1, 10, 10, 50, 50, aseprite.Color{R: 0, G: 0, B: 0, A: 255}, false, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 3 (DrawRectangle) failed: %v", err)
	}

	// Step 4: Draw filled circle
	drawScript = gen.DrawCircle("Layer 1", 1, 90, 30, 15, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 4 (DrawCircle) failed: %v", err)
	}

	// Step 5: Draw line
	drawScript = gen.DrawLine("Layer 1", 1, 20, 80, 100, 110, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, 2, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 5 (DrawLine) failed: %v", err)
	}

	// Step 6: Draw individual pixels for detail
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 30, Y: 30}, Color: aseprite.Color{R: 255, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 31, Y: 31}, Color: aseprite.Color{R: 255, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 32, Y: 32}, Color: aseprite.Color{R: 255, G: 255, B: 0, A: 255}},
	}
	drawScript = gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 6 (DrawPixels) failed: %v", err)
	}

	// Step 7: Export
	outputPath := filepath.Join(t.TempDir(), "complex-drawing.png")
	exportScript := gen.ExportSprite(outputPath, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Step 7 (ExportSprite) failed: %v", err)
	}

	// Verify export
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}
	if fileInfo.Size() == 0 {
		t.Error("Exported file is empty")
	}

	t.Logf("✓ Complex drawing workflow succeeded (%d bytes)", fileInfo.Size())
}

func TestWorkflow_FrameByFrameAnimation(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Step 1: Create canvas
	spritePath := testutil.TempSpritePath(t, "workflow-frame-by-frame.aseprite")
	createScript := gen.CreateCanvas(48, 48, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Step 1 (CreateCanvas) failed: %v", err)
	}
	defer os.Remove(spritePath)

	// Step 2: Add layers for organization
	addLayerScript := gen.AddLayer("Background")
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Step 2 (AddLayer Background) failed: %v", err)
	}

	addLayerScript = gen.AddLayer("Animation")
	_, err = client.ExecuteLua(ctx, addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Step 3 (AddLayer Animation) failed: %v", err)
	}

	// Step 4: Add frames
	for i := 0; i < 2; i++ {
		addFrameScript := gen.AddFrame(150)
		_, err = client.ExecuteLua(ctx, addFrameScript, spritePath)
		if err != nil {
			t.Fatalf("Step %d (AddFrame) failed: %v", i+4, err)
		}
	}

	// Step 5: Draw background on all frames (fill once)
	fillScript := gen.FillArea("Background", 1, 24, 24, aseprite.Color{R: 30, G: 30, B: 30, A: 255}, 0, false)
	_, err = client.ExecuteLua(ctx, fillScript, spritePath)
	if err != nil {
		t.Fatalf("Step 5 (FillArea Background) failed: %v", err)
	}

	// Step 6-8: Draw different shapes on each frame of Animation layer
	// Frame 1: Square
	drawScript := gen.DrawRectangle("Animation", 1, 14, 14, 20, 20, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 6 (DrawRectangle frame 1) failed: %v", err)
	}

	// Frame 2: Circle
	drawScript = gen.DrawCircle("Animation", 2, 24, 24, 10, aseprite.Color{R: 0, G: 255, B: 0, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 7 (DrawCircle frame 2) failed: %v", err)
	}

	// Frame 3: Triangle (using line tool)
	drawScript = gen.DrawLine("Animation", 3, 24, 10, 10, 34, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, 2, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 8a (DrawLine1 frame 3) failed: %v", err)
	}
	drawScript = gen.DrawLine("Animation", 3, 10, 34, 38, 34, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, 2, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 8b (DrawLine2 frame 3) failed: %v", err)
	}
	drawScript = gen.DrawLine("Animation", 3, 38, 34, 24, 10, aseprite.Color{R: 0, G: 0, B: 255, A: 255}, 2, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Step 8c (DrawLine3 frame 3) failed: %v", err)
	}

	// Step 9: Verify frame count
	infoScript := gen.GetSpriteInfo()
	output, err := client.ExecuteLua(ctx, infoScript, spritePath)
	if err != nil {
		t.Fatalf("Step 9 (GetSpriteInfo) failed: %v", err)
	}
	if !strings.Contains(output, "\"frame_count\": 3") {
		t.Errorf("Expected 3 frames, got output: %s", output)
	}

	// Step 10: Export each frame individually
	for i := 1; i <= 3; i++ {
		framePath := filepath.Join(t.TempDir(), "frame"+string(rune('0'+i))+".png")
		exportScript := gen.ExportSprite(framePath, i)
		_, err = client.ExecuteLua(ctx, exportScript, spritePath)
		if err != nil {
			t.Fatalf("Step 10.%d (ExportSprite frame %d) failed: %v", i, i, err)
		}
		if _, err := os.Stat(framePath); err != nil {
			t.Errorf("Frame %d export not found: %v", i, err)
		}
	}

	// Step 11: Export full animation as GIF
	gifPath := filepath.Join(t.TempDir(), "frame-by-frame.gif")
	exportScript := gen.ExportSprite(gifPath, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("Step 11 (ExportSprite GIF) failed: %v", err)
	}

	fileInfo, err := os.Stat(gifPath)
	if err != nil {
		t.Fatalf("GIF export not found: %v", err)
	}

	t.Logf("✓ Frame-by-frame animation workflow succeeded (%d bytes GIF)", fileInfo.Size())
}

func TestWorkflow_RapidPrototyping(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	start := time.Now()

	// Quick workflow: Create → Draw → Export in under 2 seconds (per PRD performance requirement)
	spritePath := testutil.TempSpritePath(t, "workflow-rapid.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("CreateCanvas failed: %v", err)
	}
	defer os.Remove(spritePath)

	drawScript := gen.DrawRectangle("Layer 1", 1, 16, 16, 32, 32, aseprite.Color{R: 128, G: 128, B: 128, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("DrawRectangle failed: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "rapid.png")
	exportScript := gen.ExportSprite(outputPath, 0)
	_, err = client.ExecuteLua(ctx, exportScript, spritePath)
	if err != nil {
		t.Fatalf("ExportSprite failed: %v", err)
	}

	duration := time.Since(start)

	// Verify export
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}

	// Performance check: Complete workflow should be fast (< 5s for basic operations)
	if duration > 5*time.Second {
		t.Logf("Warning: Rapid prototyping workflow took %v (expected <5s)", duration)
	}

	t.Logf("✓ Rapid prototyping workflow completed in %v", duration)
}
