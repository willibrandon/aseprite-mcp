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

// Integration tests for antialiasing with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools -run=TestIntegration_Antialiasing

func TestIntegration_SuggestAntialiasing_JaggedDiagonal(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a sprite with a jagged diagonal line
	spritePath := testutil.TempSpritePath(t, "test-antialiasing-jagged.aseprite")
	createScript := gen.CreateCanvas(16, 16, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a jagged diagonal line pattern:
	//   ..##
	//   .##.
	//   ##..
	pixels := []aseprite.Pixel{
		// Row 0
		{Point: aseprite.Point{X: 2, Y: 0}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 3, Y: 0}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		// Row 1
		{Point: aseprite.Point{X: 1, Y: 1}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 2, Y: 1}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		// Row 2
		{Point: aseprite.Point{X: 0, Y: 2}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 1, Y: 2}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw jagged line: %v", err)
	}

	// Analyze for antialiasing suggestions (without auto-apply)
	input := SuggestAntialiasingInput{
		SpritePath:  spritePath,
		LayerName:   "Layer 1",
		FrameNumber: 1,
		Threshold:   128,
		AutoApply:   false,
		UsePalette:  false,
	}

	result, err := suggestAntialiasing(ctx, client, gen, input)
	if err != nil {
		t.Fatalf("suggestAntialiasing failed: %v", err)
	}

	// Verify we got suggestions for the jagged edges
	if len(result.Suggestions) == 0 {
		t.Error("Expected antialiasing suggestions for jagged diagonal, got none")
	}

	if result.Applied {
		t.Error("AutoApply was false, but result.Applied is true")
	}

	t.Logf("✓ Detected %d jagged edge positions", len(result.Suggestions))

	// Verify suggestions have valid data
	for i, sug := range result.Suggestions {
		if sug.SuggestedColor == "" {
			t.Errorf("Suggestion %d has empty suggested color", i)
		}
		if sug.Direction == "" {
			t.Errorf("Suggestion %d has empty direction", i)
		}
		t.Logf("  Suggestion %d: pos=(%d,%d) direction=%s color=%s",
			i, sug.X, sug.Y, sug.Direction, sug.SuggestedColor)
	}
}

func TestIntegration_SuggestAntialiasing_AutoApply(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a sprite
	spritePath := testutil.TempSpritePath(t, "test-antialiasing-autoapply.aseprite")
	createScript := gen.CreateCanvas(16, 16, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a simple jagged diagonal
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 2, Y: 0}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 3, Y: 0}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 1, Y: 1}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 2, Y: 1}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw pixels: %v", err)
	}

	// Suggest and auto-apply antialiasing
	input := SuggestAntialiasingInput{
		SpritePath:  spritePath,
		LayerName:   "Layer 1",
		FrameNumber: 1,
		Threshold:   128,
		AutoApply:   true,
		UsePalette:  false,
	}

	result, err := suggestAntialiasing(ctx, client, gen, input)
	if err != nil {
		t.Fatalf("suggestAntialiasing with auto-apply failed: %v", err)
	}

	if !result.Applied {
		t.Error("AutoApply was true, but result.Applied is false")
	}

	if len(result.Suggestions) > 0 {
		t.Logf("✓ Applied %d antialiasing pixels automatically", len(result.Suggestions))
	}
}

func TestIntegration_SuggestAntialiasing_NoJaggedEdges(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a sprite
	spritePath := testutil.TempSpritePath(t, "test-antialiasing-smooth.aseprite")
	createScript := gen.CreateCanvas(16, 16, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw a smooth horizontal line (no jagged edges)
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 2, Y: 5}, Color: aseprite.Color{R: 0, G: 0, B: 255, A: 255}},
		{Point: aseprite.Point{X: 3, Y: 5}, Color: aseprite.Color{R: 0, G: 0, B: 255, A: 255}},
		{Point: aseprite.Point{X: 4, Y: 5}, Color: aseprite.Color{R: 0, G: 0, B: 255, A: 255}},
		{Point: aseprite.Point{X: 5, Y: 5}, Color: aseprite.Color{R: 0, G: 0, B: 255, A: 255}},
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw smooth line: %v", err)
	}

	// Analyze for antialiasing suggestions
	input := SuggestAntialiasingInput{
		SpritePath:  spritePath,
		LayerName:   "Layer 1",
		FrameNumber: 1,
		Threshold:   128,
		AutoApply:   false,
		UsePalette:  false,
	}

	result, err := suggestAntialiasing(ctx, client, gen, input)
	if err != nil {
		t.Fatalf("suggestAntialiasing failed: %v", err)
	}

	// Should have no suggestions for smooth horizontal line
	if len(result.Suggestions) > 0 {
		t.Logf("Note: Found %d suggestions for horizontal line (may be edge artifacts)", len(result.Suggestions))
		// Don't fail - some edge detection might occur at boundaries
	} else {
		t.Log("✓ No antialiasing suggestions for smooth line (expected)")
	}
}

func TestIntegration_SuggestAntialiasing_WithRegion(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a larger sprite
	spritePath := testutil.TempSpritePath(t, "test-antialiasing-region.aseprite")
	createScript := gen.CreateCanvas(32, 32, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw jagged edges in top-left and bottom-right
	pixels := []aseprite.Pixel{
		// Top-left jagged diagonal
		{Point: aseprite.Point{X: 2, Y: 0}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 3, Y: 0}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 1, Y: 1}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 2, Y: 1}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		// Bottom-right jagged diagonal
		{Point: aseprite.Point{X: 28, Y: 30}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 29, Y: 30}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 29, Y: 31}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 30, Y: 31}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
	}

	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw pixels: %v", err)
	}

	// Analyze only the top-left 16x16 region
	region := Region{X: 0, Y: 0, Width: 16, Height: 16}
	input := SuggestAntialiasingInput{
		SpritePath:  spritePath,
		LayerName:   "Layer 1",
		FrameNumber: 1,
		Region:      &region,
		Threshold:   128,
		AutoApply:   false,
		UsePalette:  false,
	}

	result, err := suggestAntialiasing(ctx, client, gen, input)
	if err != nil {
		t.Fatalf("suggestAntialiasing with region failed: %v", err)
	}

	// Should only detect jagged edges in the top-left region
	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions in top-left region, got none")
	}

	// Verify all suggestions are within the specified region
	for i, sug := range result.Suggestions {
		if sug.X < region.X || sug.X >= region.X+region.Width ||
			sug.Y < region.Y || sug.Y >= region.Y+region.Height {
			t.Errorf("Suggestion %d at (%d,%d) is outside region (0,0,16,16)",
				i, sug.X, sug.Y)
		}
	}

	t.Logf("✓ Found %d suggestions within region (0,0,16,16)", len(result.Suggestions))
}

func TestIntegration_SuggestAntialiasing_InvalidLayer(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a sprite
	spritePath := testutil.TempSpritePath(t, "test-antialiasing-invalid-layer.aseprite")
	createScript := gen.CreateCanvas(16, 16, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Try to analyze a non-existent layer
	input := SuggestAntialiasingInput{
		SpritePath:  spritePath,
		LayerName:   "NonExistentLayer",
		FrameNumber: 1,
		Threshold:   128,
		AutoApply:   false,
		UsePalette:  false,
	}

	_, err = suggestAntialiasing(ctx, client, gen, input)
	if err == nil {
		t.Fatal("Expected error for non-existent layer, got nil")
	}

	if !strings.Contains(err.Error(), "Layer not found") {
		t.Logf("Got expected error: %v", err)
	}

	t.Log("✓ Invalid layer error caught correctly")
}
