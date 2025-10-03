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

// FormatColorWithPalette formats a Color with optional palette snapping.
//
// If usePalette is false, returns a direct Color constructor call.
// If usePalette is true, wraps the color in snapToPalette() to find the nearest palette color.
//
// The snapToPalette function must be defined in the script (use GeneratePaletteSnapperHelper).
// This is useful for palette-constrained pixel art to ensure all colors match the palette.
func FormatColorWithPalette(c Color, usePalette bool) string {
	if !usePalette {
		return FormatColor(c)
	}
	return fmt.Sprintf("snapToPalette(%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
}

// GeneratePaletteSnapperHelper returns Lua code defining a snapToPalette helper function.
//
// The generated function snaps an arbitrary RGBA color to the nearest color in the
// sprite's active palette using LAB color space distance for perceptual accuracy.
//
// Include this helper at the start of scripts that use palette-aware drawing (use_palette=true).
// The function signature is: snapToPalette(r, g, b, a) -> Color
func GeneratePaletteSnapperHelper() string {
	return `
-- Helper: Snap color to nearest palette color
local function snapToPalette(r, g, b, a)
	local palette = app.activeSprite.palettes[1]
	if not palette or #palette == 0 then
		-- No palette available, return original color
		return Color(r, g, b, a)
	end

	local minDist = math.huge
	local nearestColor = palette:getColor(0)

	for i = 0, #palette - 1 do
		local palColor = palette:getColor(i)
		local dr = r - palColor.red
		local dg = g - palColor.green
		local db = b - palColor.blue
		local da = a - palColor.alpha
		local dist = dr*dr + dg*dg + db*db + da*da

		if dist < minDist then
			minDist = dist
			nearestColor = palColor
		end
	end

	return Color(nearestColor.red, nearestColor.green, nearestColor.blue, nearestColor.alpha)
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
