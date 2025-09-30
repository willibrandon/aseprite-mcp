//go:build integration
// +build integration

package tools

import (
	"context"
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