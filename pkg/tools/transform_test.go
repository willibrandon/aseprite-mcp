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

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
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
