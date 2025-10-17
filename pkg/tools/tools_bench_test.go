//go:build integration
// +build integration

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
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

// Professional Pixel Art Feature Benchmarks

func BenchmarkProfessional_DrawWithDither(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Setup: create canvas once
	spritePath := filepath.Join(b.TempDir(), "bench-dither.aseprite")
	createScript := gen.CreateCanvas(128, 128, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("CreateCanvas failed: %v", err)
	}
	defer os.Remove(spritePath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ditherScript := gen.DrawWithDither("Layer 1", 1, 0, 0, 128, 128, "#0A1628", "#5B7FB0", "bayer_4x4", 0.5)
		_, err := client.ExecuteLua(ctx, ditherScript, spritePath)
		if err != nil {
			b.Fatalf("DrawWithDither failed: %v", err)
		}
	}
}

func BenchmarkProfessional_ApplyShading(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Setup: create canvas with a circle
	spritePath := filepath.Join(b.TempDir(), "bench-shading.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("CreateCanvas failed: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a circle to shade
	drawScript := gen.DrawCircle("Layer 1", 1, 32, 32, 24, aseprite.Color{R: 128, G: 128, B: 200, A: 255}, true, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		b.Fatalf("DrawCircle failed: %v", err)
	}

	palette := []string{"#1A2639", "#3D5A80", "#5B7FB0", "#8BA3C7"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shadingScript := gen.ApplyShading("Layer 1", 1, 8, 8, 48, 48, palette, "top_left", 0.7, "smooth")
		_, err := client.ExecuteLua(ctx, shadingScript, spritePath)
		if err != nil {
			b.Fatalf("ApplyShading failed: %v", err)
		}
	}
}

func BenchmarkProfessional_SuggestAntialiasing(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Setup: create canvas with jagged diagonal line once
	spritePath := filepath.Join(b.TempDir(), "bench-aa.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("CreateCanvas failed: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw jagged diagonal (stair-step pattern)
	jaggedPixels := make([]aseprite.Pixel, 0)
	for i := 0; i < 5; i++ {
		for j := 0; j < 4; j++ {
			jaggedPixels = append(jaggedPixels, aseprite.Pixel{
				Point: aseprite.Point{X: 20 + i + j, Y: 10 + i},
				Color: aseprite.Color{R: 255, G: 0, B: 255, A: 255},
			})
		}
	}
	drawScript := gen.DrawPixels("Layer 1", 1, jaggedPixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		b.Fatalf("DrawPixels failed: %v", err)
	}

	// For benchmarking, we'll use the Lua approach similar to integration tests
	// Since suggestAntialiasing is complex with pixel reading, we'll benchmark
	// the GetPixels operation which is the bulk of the work
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pixelScript := gen.GetPixels("Layer 1", 1, 0, 0, 64, 64)
		_, err := client.ExecuteLua(ctx, pixelScript, spritePath)
		if err != nil {
			b.Fatalf("GetPixels failed: %v", err)
		}
	}
}

func BenchmarkProfessional_DownsampleImage_Small(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a 256x256 test image
	sourcePath := filepath.Join(b.TempDir(), "bench-source-small.aseprite")
	createScript := gen.CreateCanvas(256, 256, aseprite.ColorModeRGB, sourcePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("CreateCanvas failed: %v", err)
	}
	defer os.Remove(sourcePath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(b.TempDir(), "bench-downsampled-small.aseprite")
		downsampleScript := gen.DownsampleImage(sourcePath, outputPath, 64, 64)
		_, err := client.ExecuteLua(ctx, downsampleScript, "")
		if err != nil {
			b.Fatalf("DownsampleImage failed: %v", err)
		}
		os.Remove(outputPath)
	}
}

func BenchmarkProfessional_DownsampleImage_Large(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 60*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a 1024x1024 test image
	sourcePath := filepath.Join(b.TempDir(), "bench-source-large.aseprite")
	createScript := gen.CreateCanvas(1024, 1024, aseprite.ColorModeRGB, sourcePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("CreateCanvas failed: %v", err)
	}
	defer os.Remove(sourcePath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(b.TempDir(), "bench-downsampled-large.aseprite")
		downsampleScript := gen.DownsampleImage(sourcePath, outputPath, 128, 128)
		_, err := client.ExecuteLua(ctx, downsampleScript, "")
		if err != nil {
			b.Fatalf("DownsampleImage failed: %v", err)
		}
		os.Remove(outputPath)
	}
}

// End-to-end professional workflow benchmark
func BenchmarkProfessional_CompleteWorkflow(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 60*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Step 1: Create 64x64 canvas
		spritePath := filepath.Join(b.TempDir(), "bench-pro-workflow.aseprite")
		createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
		_, err := client.ExecuteLua(ctx, createScript, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}

		// Step 2: Apply dithered gradient background
		ditherScript := gen.DrawWithDither("Layer 1", 1, 0, 0, 64, 64, "#0A1628", "#3D5A80", "bayer_4x4", 0.3)
		_, err = client.ExecuteLua(ctx, ditherScript, spritePath)
		if err != nil {
			b.Fatalf("DrawWithDither failed: %v", err)
		}

		// Step 3: Draw and shade a circle
		drawScript := gen.DrawCircle("Layer 1", 1, 32, 32, 20, aseprite.Color{R: 200, G: 100, B: 150, A: 255}, true, false)
		_, err = client.ExecuteLua(ctx, drawScript, spritePath)
		if err != nil {
			b.Fatalf("DrawCircle failed: %v", err)
		}

		palette := []string{"#3B1725", "#8B5580", "#D4A5C0", "#FEF9A7"}
		shadingScript := gen.ApplyShading("Layer 1", 1, 12, 12, 40, 40, palette, "top_left", 0.7, "smooth")
		_, err = client.ExecuteLua(ctx, shadingScript, spritePath)
		if err != nil {
			b.Fatalf("ApplyShading failed: %v", err)
		}

		// Step 4: Export
		exportPath := filepath.Join(b.TempDir(), "bench-pro-output.png")
		exportScript := gen.ExportSprite(exportPath, 0)
		_, err = client.ExecuteLua(ctx, exportScript, spritePath)
		if err != nil {
			b.Fatalf("ExportSprite failed: %v", err)
		}

		os.Remove(spritePath)
		os.Remove(exportPath)
	}
}
