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
)

// Integration tests for palette tools with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools

func TestIntegration_SetPalette_BasicPalette(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-set-palette.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Set palette with 8 colors
	colors := []string{
		"#000000", // Black
		"#FF0000", // Red
		"#00FF00", // Green
		"#0000FF", // Blue
		"#FFFF00", // Yellow
		"#FF00FF", // Magenta
		"#00FFFF", // Cyan
		"#FFFFFF", // White
	}

	paletteScript := gen.SetPalette(colors)
	output, err := client.ExecuteLua(ctx, paletteScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(SetPalette) error = %v", err)
	}

	if !strings.Contains(output, "Palette set successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Set palette with %d colors successfully", len(colors))
}

func TestIntegration_SetPalette_LargePalette(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-set-large-palette.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Generate 32 colors (gradient from black to white)
	colors := make([]string, 32)
	for i := 0; i < 32; i++ {
		gray := uint8((i * 255) / 31)
		colors[i] = aseprite.Color{R: gray, G: gray, B: gray, A: 255}.ToHex()
	}

	paletteScript := gen.SetPalette(colors)
	output, err := client.ExecuteLua(ctx, paletteScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(SetPalette) error = %v", err)
	}

	if !strings.Contains(output, "Palette set successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Set large palette with %d colors successfully", len(colors))
}

func TestIntegration_ApplyShading_SmoothStyle(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas and draw a base shape
	spritePath := testutil.TempSpritePath(t, "test-shading-smooth.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a filled circle
	circleScript := gen.DrawCircle("Layer 1", 1, 32, 32, 20, aseprite.Color{R: 128, G: 128, B: 128, A: 255}, true)
	_, err = client.ExecuteLua(ctx, circleScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw circle: %v", err)
	}

	// Apply shading with 3-color palette (dark, mid, light)
	palette := []string{"#404040", "#808080", "#C0C0C0"}
	shadingScript := gen.ApplyShading("Layer 1", 1, 12, 12, 40, 40, palette, "top_left", 0.7, "smooth")
	output, err := client.ExecuteLua(ctx, shadingScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ApplyShading) error = %v", err)
	}

	if !strings.Contains(output, "Shading applied") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Applied smooth shading successfully")
}

func TestIntegration_ApplyShading_HardStyle(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas and draw a base shape
	spritePath := testutil.TempSpritePath(t, "test-shading-hard.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a filled rectangle
	rectScript := gen.DrawRectangle("Layer 1", 1, 10, 10, 44, 44, aseprite.Color{R: 255, G: 100, B: 100, A: 255}, true)
	_, err = client.ExecuteLua(ctx, rectScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw rectangle: %v", err)
	}

	// Apply hard shading with 4-color palette
	palette := []string{"#800000", "#C00000", "#FF6060", "#FFC0C0"}
	shadingScript := gen.ApplyShading("Layer 1", 1, 10, 10, 44, 44, palette, "top", 0.9, "hard")
	output, err := client.ExecuteLua(ctx, shadingScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ApplyShading) error = %v", err)
	}

	if !strings.Contains(output, "Shading applied") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Applied hard shading successfully")
}

func TestIntegration_ApplyShading_AllDirections(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	directions := []string{"top_left", "top", "top_right", "left", "right", "bottom_left", "bottom", "bottom_right"}
	palette := []string{"#000000", "#808080", "#FFFFFF"}

	for _, direction := range directions {
		t.Run(direction, func(t *testing.T) {
			// Create a canvas
			spritePath := testutil.TempSpritePath(t, "test-shading-"+direction+".aseprite")
			createScript := gen.CreateCanvas(32, 32, aseprite.ColorModeRGB, spritePath)
			_, err := client.ExecuteLua(ctx, createScript, "")
			if err != nil {
				t.Fatalf("Failed to create canvas: %v", err)
			}
			defer os.Remove(spritePath)

			// Draw filled rectangle
			rectScript := gen.DrawRectangle("Layer 1", 1, 5, 5, 22, 22, aseprite.Color{R: 128, G: 128, B: 128, A: 255}, true)
			_, err = client.ExecuteLua(ctx, rectScript, spritePath)
			if err != nil {
				t.Fatalf("Failed to draw rectangle: %v", err)
			}

			// Apply shading from this direction
			shadingScript := gen.ApplyShading("Layer 1", 1, 5, 5, 22, 22, palette, direction, 0.5, "smooth")
			output, err := client.ExecuteLua(ctx, shadingScript, spritePath)
			if err != nil {
				t.Fatalf("ExecuteLua(ApplyShading) error = %v", err)
			}

			if !strings.Contains(output, "Shading applied") {
				t.Errorf("Expected success message, got: %s", output)
			}

			t.Logf("✓ Applied shading from %s successfully", direction)
		})
	}
}

func TestIntegration_SetPalette_ThenApplyShading(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-palette-shading-workflow.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeIndexed, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Step 1: Set a limited palette (Lospec Pico-8 inspired)
	palette := []string{
		"#000000", "#1D2B53", "#7E2553", "#008751",
		"#AB5236", "#5F574F", "#C2C3C7", "#FFF1E8",
		"#FF004D", "#FFA300", "#FFEC27", "#00E436",
		"#29ADFF", "#83769C", "#FF77A8", "#FFCCAA",
	}

	paletteScript := gen.SetPalette(palette)
	_, err = client.ExecuteLua(ctx, paletteScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to set palette: %v", err)
	}

	// Step 2: Draw a circle with mid-tone
	circleScript := gen.DrawCircle("Layer 1", 1, 32, 32, 25, aseprite.Color{R: 195, G: 87, B: 54, A: 255}, true)
	_, err = client.ExecuteLua(ctx, circleScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw circle: %v", err)
	}

	// Step 3: Apply shading using palette subset
	shadingPalette := []string{"#7E2553", "#AB5236", "#FFA300"}
	shadingScript := gen.ApplyShading("Layer 1", 1, 7, 7, 50, 50, shadingPalette, "top_left", 0.6, "smooth")
	output, err := client.ExecuteLua(ctx, shadingScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(ApplyShading) error = %v", err)
	}

	if !strings.Contains(output, "Shading applied") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Complete palette workflow: set palette → draw → apply shading")
}
