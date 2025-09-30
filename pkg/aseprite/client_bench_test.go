//go:build integration
// +build integration

package aseprite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
)

// Benchmarks for Aseprite client operations.
// Run with: go test -tags=integration -bench=. -benchmem ./pkg/aseprite

func BenchmarkCreateCanvas_Small(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := filepath.Join(b.TempDir(), "bench-small.aseprite")
		script := gen.CreateCanvas(64, 64, ColorModeRGB, filename)
		_, err := client.ExecuteLua(ctx, script, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}
		os.Remove(filename)
	}
}

func BenchmarkCreateCanvas_Medium(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := filepath.Join(b.TempDir(), "bench-medium.aseprite")
		script := gen.CreateCanvas(320, 240, ColorModeRGB, filename)
		_, err := client.ExecuteLua(ctx, script, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}
		os.Remove(filename)
	}
}

func BenchmarkCreateCanvas_Large(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := filepath.Join(b.TempDir(), "bench-large.aseprite")
		script := gen.CreateCanvas(1920, 1080, ColorModeRGB, filename)
		_, err := client.ExecuteLua(ctx, script, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}
		os.Remove(filename)
	}
}

func BenchmarkDrawPixels_10(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup: create canvas once
	filename := filepath.Join(b.TempDir(), "bench-pixels-10.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	// Generate 10 pixels
	pixels := make([]Pixel, 10)
	for i := 0; i < 10; i++ {
		pixels[i] = Pixel{
			Point: Point{X: i, Y: i},
			Color: Color{R: 255, G: 0, B: 0, A: 255},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.DrawPixels("Layer 1", 1, pixels)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("DrawPixels failed: %v", err)
		}
	}
}

func BenchmarkDrawPixels_100(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup
	filename := filepath.Join(b.TempDir(), "bench-pixels-100.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	// Generate 100 pixels
	pixels := make([]Pixel, 100)
	for i := 0; i < 100; i++ {
		pixels[i] = Pixel{
			Point: Point{X: i % 100, Y: i / 100},
			Color: Color{R: uint8(i), G: 128, B: uint8(255 - i), A: 255},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.DrawPixels("Layer 1", 1, pixels)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("DrawPixels failed: %v", err)
		}
	}
}

func BenchmarkDrawPixels_1000(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup
	filename := filepath.Join(b.TempDir(), "bench-pixels-1000.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	// Generate 1000 pixels
	pixels := make([]Pixel, 1000)
	for i := 0; i < 1000; i++ {
		pixels[i] = Pixel{
			Point: Point{X: i % 100, Y: i / 100},
			Color: Color{R: uint8(i % 256), G: 128, B: uint8((255 - i) % 256), A: 255},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.DrawPixels("Layer 1", 1, pixels)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("DrawPixels failed: %v", err)
		}
	}
}

func BenchmarkDrawPixels_10000(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 60*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup
	filename := filepath.Join(b.TempDir(), "bench-pixels-10000.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	// Generate 10000 pixels (fill entire 100x100 canvas)
	pixels := make([]Pixel, 10000)
	for i := 0; i < 10000; i++ {
		pixels[i] = Pixel{
			Point: Point{X: i % 100, Y: i / 100},
			Color: Color{R: uint8(i % 256), G: uint8((i / 100) % 256), B: 128, A: 255},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.DrawPixels("Layer 1", 1, pixels)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("DrawPixels failed: %v", err)
		}
	}
}

func BenchmarkDrawLine(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup
	filename := filepath.Join(b.TempDir(), "bench-line.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.DrawLine("Layer 1", 1, 10, 10, 90, 90, Color{R: 255, G: 0, B: 0, A: 255}, 2)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("DrawLine failed: %v", err)
		}
	}
}

func BenchmarkDrawRectangle(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup
	filename := filepath.Join(b.TempDir(), "bench-rect.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.DrawRectangle("Layer 1", 1, 20, 20, 60, 60, Color{R: 0, G: 255, B: 0, A: 255}, true)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("DrawRectangle failed: %v", err)
		}
	}
}

func BenchmarkDrawCircle(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup
	filename := filepath.Join(b.TempDir(), "bench-circle.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.DrawCircle("Layer 1", 1, 50, 50, 30, Color{R: 0, G: 0, B: 255, A: 255}, true)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("DrawCircle failed: %v", err)
		}
	}
}

func BenchmarkFillArea(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup
	filename := filepath.Join(b.TempDir(), "bench-fill.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.FillArea("Layer 1", 1, 50, 50, Color{R: 255, G: 255, B: 0, A: 255}, 0)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("FillArea failed: %v", err)
		}
	}
}

func BenchmarkExportSprite_PNG(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup: create canvas with content
	filename := filepath.Join(b.TempDir(), "bench-export.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	// Draw something
	drawScript := gen.DrawCircle("Layer 1", 1, 50, 50, 30, Color{R: 255, G: 0, B: 0, A: 255}, true)
	_, err = client.ExecuteLua(ctx, drawScript, filename)
	if err != nil {
		b.Fatalf("Setup draw failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(b.TempDir(), "export-bench.png")
		script := gen.ExportSprite(outputPath, 0)
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("ExportSprite failed: %v", err)
		}
		os.Remove(outputPath)
	}
}

func BenchmarkAddLayer(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: create fresh canvas for each iteration
		filename := filepath.Join(b.TempDir(), "bench-layer.aseprite")
		createScript := gen.CreateCanvas(64, 64, ColorModeRGB, filename)
		_, err := client.ExecuteLua(ctx, createScript, "")
		if err != nil {
			b.Fatalf("Setup failed: %v", err)
		}
		b.StartTimer()

		script := gen.AddLayer("Test Layer")
		_, err = client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("AddLayer failed: %v", err)
		}

		b.StopTimer()
		os.Remove(filename)
	}
}

func BenchmarkAddFrame(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: create fresh canvas for each iteration
		filename := filepath.Join(b.TempDir(), "bench-frame.aseprite")
		createScript := gen.CreateCanvas(64, 64, ColorModeRGB, filename)
		_, err := client.ExecuteLua(ctx, createScript, "")
		if err != nil {
			b.Fatalf("Setup failed: %v", err)
		}
		b.StartTimer()

		script := gen.AddFrame(100)
		_, err = client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("AddFrame failed: %v", err)
		}

		b.StopTimer()
		os.Remove(filename)
	}
}

func BenchmarkGetSpriteInfo(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	// Setup: create canvas once
	filename := filepath.Join(b.TempDir(), "bench-info.aseprite")
	createScript := gen.CreateCanvas(100, 100, ColorModeRGB, filename)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	defer os.Remove(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := gen.GetSpriteInfo()
		_, err := client.ExecuteLua(ctx, script, filename)
		if err != nil {
			b.Fatalf("GetSpriteInfo failed: %v", err)
		}
	}
}

// End-to-end workflow benchmarks

func BenchmarkWorkflow_CreateDrawExport(b *testing.B) {
	cfg := testutil.LoadTestConfigTB(b)
	client := NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := NewLuaGenerator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create canvas
		spritePath := filepath.Join(b.TempDir(), "bench-workflow.aseprite")
		createScript := gen.CreateCanvas(64, 64, ColorModeRGB, spritePath)
		_, err := client.ExecuteLua(ctx, createScript, "")
		if err != nil {
			b.Fatalf("CreateCanvas failed: %v", err)
		}

		// Draw circle
		drawScript := gen.DrawCircle("Layer 1", 1, 32, 32, 20, Color{R: 255, G: 0, B: 0, A: 255}, true)
		_, err = client.ExecuteLua(ctx, drawScript, spritePath)
		if err != nil {
			b.Fatalf("DrawCircle failed: %v", err)
		}

		// Export
		outputPath := filepath.Join(b.TempDir(), "bench-output.png")
		exportScript := gen.ExportSprite(outputPath, 0)
		_, err = client.ExecuteLua(ctx, exportScript, spritePath)
		if err != nil {
			b.Fatalf("ExportSprite failed: %v", err)
		}

		// Cleanup
		os.Remove(spritePath)
		os.Remove(outputPath)
	}
}
