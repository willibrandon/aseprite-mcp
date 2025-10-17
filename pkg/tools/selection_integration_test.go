//go:build integration
// +build integration

package tools

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// Integration tests for selection tools with real Aseprite.
// Run with: go test -tags=integration -v ./pkg/tools

func TestIntegration_SelectRectangle(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-select-rect.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Create a rectangular selection
	selectScript := gen.SelectRectangle(10, 10, 30, 40, "replace")
	output, err := client.ExecuteLua(ctx, selectScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(SelectRectangle) error = %v", err)
	}

	if !strings.Contains(output, "Rectangle selection created successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Created rectangular selection (10, 10, 30, 40)")
}

func TestIntegration_SelectionModes(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-select-modes.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Test each selection mode
	modes := []string{"replace", "add", "subtract", "intersect"}
	for _, mode := range modes {
		selectScript := gen.SelectRectangle(20, 20, 20, 20, mode)
		output, err := client.ExecuteLua(ctx, selectScript, spritePath)
		if err != nil {
			t.Errorf("ExecuteLua(SelectRectangle, mode=%s) error = %v", mode, err)
			continue
		}

		if !strings.Contains(output, "Rectangle selection created successfully") {
			t.Errorf("Mode %s failed, got: %s", mode, output)
		}

		t.Logf("✓ Selection mode %s works", mode)
	}
}

func TestIntegration_SelectAll_Deselect(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-select-all.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Select all
	selectAllScript := gen.SelectAll()
	output, err := client.ExecuteLua(ctx, selectAllScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(SelectAll) error = %v", err)
	}

	if !strings.Contains(output, "Select all completed successfully") {
		t.Errorf("Expected select all success, got: %s", output)
	}

	t.Logf("✓ Selected all pixels")

	// Deselect
	deselectScript := gen.Deselect()
	output, err = client.ExecuteLua(ctx, deselectScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(Deselect) error = %v", err)
	}

	if !strings.Contains(output, "Deselect completed successfully") {
		t.Errorf("Expected deselect success, got: %s", output)
	}

	t.Logf("✓ Deselected successfully")
}

func TestIntegration_MoveSelection(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-move-selection.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Combine select + move into a single Lua script
	// This is necessary because selections don't persist across separate Aseprite invocations
	combinedScript := `local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Create selection
local rect = Rectangle(10, 10, 20, 20)
local sel = Selection(rect)
spr.selection = sel

-- Move selection
if spr.selection.isEmpty then
	error("No active selection to move")
end

local bounds = spr.selection.bounds
local newRect = Rectangle(bounds.x + 15, bounds.y + -5, bounds.width, bounds.height)
local newSel = Selection(newRect)
spr.selection = newSel

spr:saveAs(spr.filename)
print("Selection moved successfully")`

	output, err := client.ExecuteLua(ctx, combinedScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(MoveSelection) error = %v", err)
	}

	if !strings.Contains(output, "Selection moved successfully") {
		t.Errorf("Expected move success, got: %s", output)
	}

	t.Logf("✓ Moved selection by (15, -5)")
}

func TestIntegration_CopyPasteWorkflow(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with some content
	spritePath := testutil.TempSpritePath(t, "test-copy-paste.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw pixels to have content to copy
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 20, Y: 20}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 21, Y: 20}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 20, Y: 21}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: aseprite.Point{X: 21, Y: 21}, Color: aseprite.Color{R: 255, G: 0, B: 0, A: 255}},
	}
	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw pixels: %v", err)
	}

	// Combine select + copy + paste into a single Lua script
	// This is necessary because clipboard state doesn't persist across separate Aseprite invocations
	combinedScript := `local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Create selection
local rect = Rectangle(20, 20, 2, 2)
local sel = Selection(rect)
spr.selection = sel

-- Copy selection
if spr.selection.isEmpty then
	error("No active selection to copy")
end
app.command.Copy()
print("✓ Copied selection")

-- Move to layer and frame for paste
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "Layer 1" then
		layer = lyr
		break
	end
end
if not layer then
	error("Layer not found")
end

local frame = spr.frames[1]
if not frame then
	error("Frame not found")
end

-- Paste at position (50, 50) using Paste command with x, y parameters
app.command.Paste { x = 50, y = 50 }
print("✓ Pasted at position (50, 50)")

spr:saveAs(spr.filename)
print("Copy-paste workflow completed successfully")`

	output, err := client.ExecuteLua(ctx, combinedScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(copy-paste workflow) error = %v", err)
	}

	if !strings.Contains(output, "Copy-paste workflow completed successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Copy-paste workflow completed")
}

func TestIntegration_CutWorkflow(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas with some content
	spritePath := testutil.TempSpritePath(t, "test-cut.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Draw some pixels
	pixels := []aseprite.Pixel{
		{Point: aseprite.Point{X: 30, Y: 30}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 31, Y: 30}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 30, Y: 31}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
		{Point: aseprite.Point{X: 31, Y: 31}, Color: aseprite.Color{R: 0, G: 255, B: 0, A: 255}},
	}
	drawScript := gen.DrawPixels("Layer 1", 1, pixels, false)
	_, err = client.ExecuteLua(ctx, drawScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to draw pixels: %v", err)
	}

	// Combine select + cut + paste into a single Lua script
	// This is necessary because clipboard state doesn't persist across separate Aseprite invocations
	combinedScript := `local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Create selection
local rect = Rectangle(30, 30, 2, 2)
local sel = Selection(rect)
spr.selection = sel

-- Find layer and set active
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "Layer 1" then
		layer = lyr
		break
	end
end
if not layer then
	error("Layer not found")
end

app.activeLayer = layer
app.activeFrame = spr.frames[1]

-- Cut selection (removes pixels and puts on clipboard)
if spr.selection.isEmpty then
	error("No active selection to cut")
end
app.command.Cut()
print("✓ Cut selection (pixels removed and on clipboard)")

-- Set paste position
local tempSel = Selection(Rectangle(60, 60, 1, 1))
spr.selection = tempSel

-- Paste at new location
app.command.Paste()
print("✓ Pasted cut pixels at new location (60, 60)")

spr:saveAs(spr.filename)
print("Cut workflow completed successfully")`

	output, err := client.ExecuteLua(ctx, combinedScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(cut workflow) error = %v", err)
	}

	if !strings.Contains(output, "Cut workflow completed successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Cut workflow completed")
}

func TestIntegration_SelectEllipse(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, 30*time.Second)
	gen := aseprite.NewLuaGenerator()
	ctx := context.Background()

	// Create a canvas
	spritePath := testutil.TempSpritePath(t, "test-select-ellipse.aseprite")
	createScript := gen.CreateCanvas(100, 100, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(ctx, createScript, "")
	if err != nil {
		t.Fatalf("Failed to create canvas: %v", err)
	}
	defer os.Remove(spritePath)

	// Create an elliptical selection
	selectScript := gen.SelectEllipse(20, 20, 40, 30, "replace")
	output, err := client.ExecuteLua(ctx, selectScript, spritePath)
	if err != nil {
		t.Fatalf("ExecuteLua(SelectEllipse) error = %v", err)
	}

	if !strings.Contains(output, "Ellipse selection created successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	t.Logf("✓ Created elliptical selection (40x30 at 20, 20)")
}
