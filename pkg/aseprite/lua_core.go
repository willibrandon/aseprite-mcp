package aseprite

import (
	"fmt"
	"strings"
)

// LuaGenerator provides utilities for generating Lua scripts for Aseprite batch operations.
//
// All generated scripts are designed to run in Aseprite's batch mode (--batch --script).
// Scripts include proper error handling, transactions for atomicity, and sprite saving.
//
// The generator is stateless and safe for concurrent use.
type LuaGenerator struct{}

// NewLuaGenerator creates a new Lua script generator.
//
// The generator is stateless and can be reused for multiple script generation operations.
func NewLuaGenerator() *LuaGenerator {
	return &LuaGenerator{}
}

// EscapeString escapes a string for safe use in Lua code.
//
// Handles special characters that could break Lua syntax or introduce injection vulnerabilities:
//   - Backslashes (\) are escaped to (\\)
//   - Double quotes (") are escaped to (\")
//   - Newlines (\n), carriage returns (\r), and tabs (\t) are escaped
//
// Always use this function when embedding user-provided strings in generated Lua code
// to prevent script injection attacks.
func EscapeString(s string) string {
	// Replace backslashes first
	s = strings.ReplaceAll(s, `\`, `\\`)

	// Replace quotes
	s = strings.ReplaceAll(s, `"`, `\"`)

	// Replace newlines
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)

	return s
}

// FormatColor formats a Color as a Lua Color constructor call.
//
// Returns a string like "Color(255, 0, 0, 255)" suitable for embedding in Lua scripts.
// The generated code creates an Aseprite Color object with RGBA values.
func FormatColor(c Color) string {
	return fmt.Sprintf("Color(%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
}

// FormatColorWithPalette formats a Color with optional palette snapping for img:putPixel.
//
// If usePalette is false, returns a direct Color constructor call.
// If usePalette is true, wraps the color in snapToPaletteForPixel() to find the nearest palette color.
//
// The snapToPaletteForPixel function must be defined in the script (use GeneratePaletteSnapperHelper).
// This is useful for palette-constrained pixel art to ensure all colors match the palette.
// Returns palette index in indexed mode, pixel color in other modes.
func FormatColorWithPalette(c Color, usePalette bool) string {
	if !usePalette {
		return FormatColor(c)
	}
	return fmt.Sprintf("snapToPaletteForPixel(%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
}

// FormatColorWithPaletteForTool formats a Color with optional palette snapping for app.useTool.
//
// If usePalette is false, returns a direct Color constructor call.
// If usePalette is true, wraps the color in snapToPaletteForTool() to find the nearest palette color.
//
// The snapToPaletteForTool function must be defined in the script (use GeneratePaletteSnapperHelper).
// Always returns a pixel color value suitable for app.useTool in all color modes.
func FormatColorWithPaletteForTool(c Color, usePalette bool) string {
	if !usePalette {
		return FormatColor(c)
	}
	return fmt.Sprintf("snapToPaletteForTool(%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
}

// GeneratePaletteSnapperHelper returns Lua code defining palette snapping helper functions.
//
// Generates two helper functions:
//  1. snapToPaletteForPixel(r, g, b, a) - for img:putPixel (returns index in indexed mode)
//  2. snapToPaletteForTool(r, g, b, a) - for app.useTool (returns pixel color)
//
// Both functions snap an arbitrary RGBA color to the nearest color in the sprite's active
// palette using Euclidean color space distance.
//
// Include this helper at the start of scripts that use palette-aware drawing (use_palette=true).
func GeneratePaletteSnapperHelper() string {
	return `
-- Helper: Find nearest palette index for given RGBA color
local function findNearestPaletteIndex(r, g, b, a)
	local spr = app.activeSprite
	local palette = spr.palettes[1]
	if not palette or #palette == 0 then
		return 0
	end

	local minDist = math.huge
	local nearestIndex = 0

	for i = 0, #palette - 1 do
		local palColor = palette:getColor(i)
		local dr = r - palColor.red
		local dg = g - palColor.green
		local db = b - palColor.blue
		local da = a - palColor.alpha
		local dist = dr*dr + dg*dg + db*db + da*da

		if dist < minDist then
			minDist = dist
			nearestIndex = i
		end
	end

	return nearestIndex
end

-- Helper: Snap color for img:putPixel (returns palette index in indexed mode)
local function snapToPaletteForPixel(r, g, b, a)
	local spr = app.activeSprite
	if spr.colorMode == ColorMode.INDEXED then
		-- In indexed mode, return the palette index directly
		return findNearestPaletteIndex(r, g, b, a)
	else
		-- In RGB/Grayscale, return pixel color
		local nearestIndex = findNearestPaletteIndex(r, g, b, a)
		local palette = spr.palettes[1]
		local nearestColor = palette:getColor(nearestIndex)
		return app.pixelColor.rgba(nearestColor.red, nearestColor.green, nearestColor.blue, nearestColor.alpha)
	end
end

-- Helper: Snap color for app.useTool (returns index in indexed mode, pixel color otherwise)
local function snapToPaletteForTool(r, g, b, a)
	local spr = app.activeSprite
	local nearestIndex = findNearestPaletteIndex(r, g, b, a)

	if spr.colorMode == ColorMode.INDEXED then
		-- In indexed mode, app.useTool expects palette index
		return nearestIndex
	else
		-- In RGB/Grayscale, app.useTool expects pixel color
		local palette = spr.palettes[1]
		local nearestColor = palette:getColor(nearestIndex)
		return app.pixelColor.rgba(nearestColor.red, nearestColor.green, nearestColor.blue, nearestColor.alpha)
	end
end

-- Default snapToPalette for backwards compatibility (uses ForPixel variant)
local function snapToPalette(r, g, b, a)
	return snapToPaletteForPixel(r, g, b, a)
end
`
}

// FormatPoint formats a Point as a Lua Point constructor call.
//
// Returns a string like "Point(10, 20)" suitable for embedding in Lua scripts.
// The generated code creates an Aseprite Point object with X, Y coordinates.
func FormatPoint(p Point) string {
	return fmt.Sprintf("Point(%d, %d)", p.X, p.Y)
}

// FormatRectangle formats a Rectangle as a Lua Rectangle constructor call.
//
// Returns a string like "Rectangle(10, 20, 30, 40)" suitable for embedding in Lua scripts.
// The generated code creates an Aseprite Rectangle object with X, Y, Width, Height.
func FormatRectangle(r Rectangle) string {
	return fmt.Sprintf("Rectangle(%d, %d, %d, %d)", r.X, r.Y, r.Width, r.Height)
}

// WrapInTransaction wraps Lua code in an app.transaction for atomicity.
//
// Aseprite transactions ensure that sprite modifications are atomic - either all
// changes succeed or all fail. This is important for undo/redo functionality.
//
// All mutation operations should be wrapped in transactions. The generated code
// has the form:
//
//	app.transaction(function()
//	  <your code here>
//	end)
func WrapInTransaction(code string) string {
	return fmt.Sprintf(`app.transaction(function()
%s
end)`, code)
}
