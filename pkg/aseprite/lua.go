package aseprite

import (
	"fmt"
	"strings"
)

// LuaGenerator provides utilities for generating Lua scripts for Aseprite.
type LuaGenerator struct{}

// NewLuaGenerator creates a new Lua script generator.
func NewLuaGenerator() *LuaGenerator {
	return &LuaGenerator{}
}

// EscapeString escapes a string for use in Lua code.
// It handles quotes, newlines, and other special characters.
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
func FormatColor(c Color) string {
	return fmt.Sprintf("Color(%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
}

// FormatPoint formats a Point as a Lua Point constructor call.
func FormatPoint(p Point) string {
	return fmt.Sprintf("Point(%d, %d)", p.X, p.Y)
}

// FormatRectangle formats a Rectangle as a Lua Rectangle constructor call.
func FormatRectangle(r Rectangle) string {
	return fmt.Sprintf("Rectangle(%d, %d, %d, %d)", r.X, r.Y, r.Width, r.Height)
}

// WrapInTransaction wraps Lua code in an app.transaction for atomicity.
func WrapInTransaction(code string) string {
	return fmt.Sprintf(`app.transaction(function()
%s
end)`, code)
}

// CreateCanvas generates a Lua script to create a new sprite.
func (g *LuaGenerator) CreateCanvas(width, height int, colorMode ColorMode) string {
	return fmt.Sprintf(`local spr = Sprite(%d, %d, %s)
local filename = os.tmpname() .. ".aseprite"
spr:saveAs(filename)
print(filename)`, width, height, colorMode.ToLua())
}

// GetSpriteInfo generates a Lua script to retrieve sprite metadata.
func (g *LuaGenerator) GetSpriteInfo() string {
	return `local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Collect layer names
local layers = {}
for i, layer in ipairs(spr.layers) do
	table.insert(layers, layer.name)
end

-- Format as JSON-like output
local output = string.format([[{
	"width": %d,
	"height": %d,
	"color_mode": "%s",
	"frame_count": %d,
	"layer_count": %d,
	"layers": ["%s"]
}]],
	spr.width,
	spr.height,
	tostring(spr.colorMode),
	#spr.frames,
	#spr.layers,
	table.concat(layers, '","')
)

print(output)`
}

// AddLayer generates a Lua script to add a new layer.
func (g *LuaGenerator) AddLayer(layerName string) string {
	escapedName := EscapeString(layerName)
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

app.transaction(function()
	local layer = spr:newLayer()
	layer.name = "%s"
end)

spr:saveAs(spr.filename)
print("Layer added successfully")`, escapedName)
}

// AddFrame generates a Lua script to add a new frame.
func (g *LuaGenerator) AddFrame(durationMs int) string {
	durationSec := float64(durationMs) / 1000.0
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

app.transaction(function()
	local frame = spr:newFrame()
	frame.duration = %.3f
end)

spr:saveAs(spr.filename)
print(#spr.frames)`, durationSec)
}

// DrawPixels generates a Lua script to draw multiple pixels.
func (g *LuaGenerator) DrawPixels(layerName string, frameNumber int, pixels []Pixel) string {
	var sb strings.Builder

	escapedName := EscapeString(layerName)

	sb.WriteString(fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local layer = spr:findLayerByName("%s")
if not layer then
	error("Layer not found: %s")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

app.transaction(function()
	local cel = layer:cel(frame)
	if not cel then
		cel = spr:newCel(layer, frame)
	end

	local img = cel.image
`, escapedName, escapedName, frameNumber, frameNumber))

	// Add pixel drawing commands
	for _, p := range pixels {
		sb.WriteString(fmt.Sprintf("\timg:putPixel(%d, %d, %s)\n", p.X, p.Y, FormatColor(p.Color)))
	}

	sb.WriteString(`end)

spr:saveAs(spr.filename)
print("Pixels drawn successfully")`)

	return sb.String()
}

// DrawLine generates a Lua script to draw a line.
func (g *LuaGenerator) DrawLine(layerName string, frameNumber int, x1, y1, x2, y2 int, color Color, thickness int) string {
	escapedName := EscapeString(layerName)
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local layer = spr:findLayerByName("%s")
if not layer then
	error("Layer not found: %s")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

app.transaction(function()
	app.activeLayer = layer
	app.activeFrame = frame

	local brush = Brush()
	brush.size = %d

	app.useTool{
		tool = "line",
		color = %s,
		brush = brush,
		points = {%s, %s}
	}
end)

spr:saveAs(spr.filename)
print("Line drawn successfully")`,
		escapedName, escapedName,
		frameNumber, frameNumber,
		thickness,
		FormatColor(color),
		FormatPoint(Point{X: x1, Y: y1}),
		FormatPoint(Point{X: x2, Y: y2}))
}

// ExportSprite generates a Lua script to export a sprite.
func (g *LuaGenerator) ExportSprite(outputPath string, frameNumber int) string {
	escapedPath := EscapeString(outputPath)

	if frameNumber > 0 {
		// Export specific frame
		return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

spr:saveCopyAs{
	filename = "%s",
	frame = frame
}

print("Exported successfully")`, frameNumber, frameNumber, escapedPath)
	}

	// Export all frames
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

spr:saveCopyAs("%s")
print("Exported successfully")`, escapedPath)
}