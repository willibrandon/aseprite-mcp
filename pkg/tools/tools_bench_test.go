//go:build integration
// +build integration

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

// Benchmarks for complete tool workflows (client + generator).
// These test the full stack that MCP tools use.
// Run with: go test -tags=integration -bench=. -benchmem ./pkg/tools

func BenchmarkCompleteWorkflow_Simple(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create 64x64 canvas
		spritePath := filepath.Join(b.TempDir(), "bench-simple.aseprite")
		createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
		_, err := client.ExecuteLua(ctx, createScript, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}

		// Draw circle
		drawScript := gen.DrawCircle("Layer 1", 1, 32, 32, 20, aseprite.Color{R: 255, G: 0, B: 0, A: 255}, true, false)
		_, err = client.ExecuteLua(ctx, drawScript, spritePath)
		if err != nil {
			b.Fatalf("DrawCircle failed: %v", err)
		}

		// Export PNG
		exportPath := filepath.Join(b.TempDir(), "bench-output.png")
		exportScript := gen.ExportSprite(exportPath, 0)
		_, err = client.ExecuteLua(ctx, exportScript, spritePath)
		if err != nil {
			b.Fatalf("ExportSprite failed: %v", err)
		}

		os.Remove(spritePath)
		os.Remove(exportPath)
	}
}

func BenchmarkCompleteWorkflow_Animation(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create canvas
		spritePath := filepath.Join(b.TempDir(), "bench-anim.aseprite")
		createScript := gen.CreateCanvas(48, 48, aseprite.ColorModeRGB, spritePath)
		_, err := client.ExecuteLua(ctx, createScript, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}

		// Add 3 frames
		for j := 0; j < 3; j++ {
			frameScript := gen.AddFrame(100)
			_, err = client.ExecuteLua(ctx, frameScript, spritePath)
			if err != nil {
				b.Fatalf("AddFrame failed: %v", err)
			}
		}

		// Draw on each frame
		for frame := 1; frame <= 4; frame++ {
			drawScript := gen.DrawCircle("Layer 1", frame, 24, 24, 5+frame*2, aseprite.Color{R: 0, G: 255, B: 0, A: 255}, true, false)
			_, err = client.ExecuteLua(ctx, drawScript, spritePath)
			if err != nil {
				b.Fatalf("DrawCircle failed: %v", err)
			}
		}

		// Export GIF
		exportPath := filepath.Join(b.TempDir(), "bench-anim.gif")
		exportScript := gen.ExportSprite(exportPath, 0)
		_, err = client.ExecuteLua(ctx, exportScript, spritePath)
		if err != nil {
			b.Fatalf("ExportSprite failed: %v", err)
		}

		os.Remove(spritePath)
		os.Remove(exportPath)
	}
}

func BenchmarkCompleteWorkflow_MultiLayer(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create canvas
		spritePath := filepath.Join(b.TempDir(), "bench-layers.aseprite")
		createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
		_, err := client.ExecuteLua(ctx, createScript, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}

		// Add background layer
		layerScript := gen.AddLayer("Background")
		_, err = client.ExecuteLua(ctx, layerScript, spritePath)
		if err != nil {
			b.Fatalf("AddLayer failed: %v", err)
		}

		// Fill background
		fillScript := gen.FillArea("Background", 1, 50, 50, aseprite.Color{R: 50, G: 50, B: 150, A: 255}, 0, false)
		_, err = client.ExecuteLua(ctx, fillScript, spritePath)
		if err != nil {
			b.Fatalf("FillArea failed: %v", err)
		}

		// Draw on Layer 1
		rectScript := gen.DrawRectangle("Layer 1", 1, 30, 30, 40, 40, aseprite.Color{R: 255, G: 255, B: 0, A: 255}, true, false)
		_, err = client.ExecuteLua(ctx, rectScript, spritePath)
		if err != nil {
			b.Fatalf("DrawRectangle failed: %v", err)
		}

		// Export
		exportPath := filepath.Join(b.TempDir(), "bench-layers.png")
		exportScript := gen.ExportSprite(exportPath, 0)
		_, err = client.ExecuteLua(ctx, exportScript, spritePath)
		if err != nil {
			b.Fatalf("ExportSprite failed: %v", err)
		}

		os.Remove(spritePath)
		os.Remove(exportPath)
	}
}

func BenchmarkCompleteWorkflow_PixelBatch(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Generate 1000 pixels once
	pixels := make([]aseprite.Pixel, 1000)
	for i := 0; i < 1000; i++ {
		pixels[i] = aseprite.Pixel{
			Point: aseprite.Point{X: i % 100, Y: i / 100},
			Color: aseprite.Color{R: uint8(i % 256), G: 128, B: uint8((255 - i) % 256), A: 255},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create canvas
		spritePath := filepath.Join(b.TempDir(), "bench-pixels.aseprite")
		createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
		_, err := client.ExecuteLua(ctx, createScript, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}

		// Draw 1000 pixels
		drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
		_, err = client.ExecuteLua(ctx, drawScript, spritePath)
		if err != nil {
			b.Fatalf("DrawPixels failed: %v", err)
		}

		// Export
		exportPath := filepath.Join(b.TempDir(), "bench-pixels.png")
		exportScript := gen.ExportSprite(exportPath, 0)
		_, err = client.ExecuteLua(ctx, exportScript, spritePath)
		if err != nil {
			b.Fatalf("ExportSprite failed: %v", err)
		}

		os.Remove(spritePath)
		os.Remove(exportPath)
	}
}
