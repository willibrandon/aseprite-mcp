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

// FormatColorWithPalette formats a Color with optional palette snapping.
// If usePalette is true, wraps the color in a call to snapToPalette().
// Returns the color expression as a string.
func FormatColorWithPalette(c Color, usePalette bool) string {
	if !usePalette {
		return FormatColor(c)
	}
	return fmt.Sprintf("snapToPalette(%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
}

// GeneratePaletteSnapperHelper returns Lua code that defines a snapToPalette helper function.
// This function snaps an RGBA color to the nearest color in the sprite's active palette.
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
// The filename parameter should be the full path where the sprite should be saved.
func (g *LuaGenerator) CreateCanvas(width, height int, colorMode ColorMode, filename string) string {
	escapedFilename := EscapeString(filename)
	return fmt.Sprintf(`local spr = Sprite(%d, %d, %s)
spr:saveAs("%s")
print("%s")`, width, height, colorMode.ToLua(), escapedFilename, escapedFilename)
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

// DeleteLayer generates a Lua script to delete a layer.
// Returns an error if the layer is not found or if attempting to delete the last layer.
func (g *LuaGenerator) DeleteLayer(layerName string) string {
	escapedName := EscapeString(layerName)
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Check if this is the last layer
if #spr.layers == 1 then
	error("Cannot delete the last layer")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

if not layer then
	error("Layer not found: %s")
end

app.transaction(function()
	spr:deleteLayer(layer)
end)

spr:saveAs(spr.filename)
print("Layer deleted successfully")`, escapedName, escapedName)
}

// DeleteFrame generates a Lua script to delete a frame.
// Returns an error if the frame is not found or if attempting to delete the last frame.
func (g *LuaGenerator) DeleteFrame(frameNumber int) string {
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Check if this is the last frame
if #spr.frames == 1 then
	error("Cannot delete the last frame")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

app.transaction(function()
	spr:deleteFrame(frame)
end)

spr:saveAs(spr.filename)
print("Frame deleted successfully")`, frameNumber, frameNumber)
}

// DrawPixels generates a Lua script to draw multiple pixels.
func (g *LuaGenerator) DrawPixels(layerName string, frameNumber int, pixels []Pixel, usePalette bool) string {
	var sb strings.Builder

	escapedName := EscapeString(layerName)

	// Add palette snapper helper if needed
	if usePalette {
		sb.WriteString(GeneratePaletteSnapperHelper())
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

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
		sb.WriteString(fmt.Sprintf("\timg:putPixel(%d, %d, %s)\n", p.X, p.Y, FormatColorWithPalette(p.Color, usePalette)))
	}

	sb.WriteString(`end)

spr:saveAs(spr.filename)
print("Pixels drawn successfully")`)

	return sb.String()
}

// DrawLine generates a Lua script to draw a line.
func (g *LuaGenerator) DrawLine(layerName string, frameNumber int, x1, y1, x2, y2 int, color Color, thickness int, usePalette bool) string {
	var sb strings.Builder

	// Add palette snapper helper if needed
	if usePalette {
		sb.WriteString(GeneratePaletteSnapperHelper())
		sb.WriteString("\n")
	}

	escapedName := EscapeString(layerName)
	sb.WriteString(fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

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

	local brush = Brush(%d)

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
		FormatColorWithPalette(color, usePalette),
		FormatPoint(Point{X: x1, Y: y1}),
		FormatPoint(Point{X: x2, Y: y2})))

	return sb.String()
}

// DrawContour generates a Lua script to draw a polyline or polygon by connecting multiple points.
// If closed is true, connects the last point back to the first to form a polygon.
func (g *LuaGenerator) DrawContour(layerName string, frameNumber int, points []Point, color Color, thickness int, closed bool, usePalette bool) string {
	var sb strings.Builder

	// Add palette snapper helper if needed
	if usePalette {
		sb.WriteString(GeneratePaletteSnapperHelper())
		sb.WriteString("\n")
	}

	escapedName := EscapeString(layerName)
	sb.WriteString(fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

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

	local brush = Brush(%d)
	local color = %s

	-- Draw lines connecting each point`,
		escapedName, escapedName,
		frameNumber, frameNumber,
		thickness,
		FormatColorWithPalette(color, usePalette)))

	// Draw line segments between consecutive points
	for i := 0; i < len(points)-1; i++ {
		sb.WriteString(fmt.Sprintf(`
	app.useTool{
		tool = "line",
		color = color,
		brush = brush,
		points = {%s, %s}
	}`,
			FormatPoint(points[i]),
			FormatPoint(points[i+1])))
	}

	// If closed, connect last point back to first
	if closed && len(points) > 0 {
		sb.WriteString(fmt.Sprintf(`
	-- Close the contour
	app.useTool{
		tool = "line",
		color = color,
		brush = brush,
		points = {%s, %s}
	}`,
			FormatPoint(points[len(points)-1]),
			FormatPoint(points[0])))
	}

	sb.WriteString(`
end)

spr:saveAs(spr.filename)
print("Contour drawn successfully")`)

	return sb.String()
}

// DrawRectangle generates a Lua script to draw a rectangle.
func (g *LuaGenerator) DrawRectangle(layerName string, frameNumber int, x, y, width, height int, color Color, filled bool, usePalette bool) string {
	var sb strings.Builder

	// Add palette snapper helper if needed
	if usePalette {
		sb.WriteString(GeneratePaletteSnapperHelper())
		sb.WriteString("\n")
	}

	escapedName := EscapeString(layerName)
	tool := "rectangle"
	if filled {
		tool = "filled_rectangle"
	}

	sb.WriteString(fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

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

	app.useTool{
		tool = "%s",
		color = %s,
		points = {%s, %s}
	}
end)

spr:saveAs(spr.filename)
print("Rectangle drawn successfully")`,
		escapedName, escapedName,
		frameNumber, frameNumber,
		tool,
		FormatColorWithPalette(color, usePalette),
		FormatPoint(Point{X: x, Y: y}),
		FormatPoint(Point{X: x + width - 1, Y: y + height - 1})))

	return sb.String()
}

// DrawCircle generates a Lua script to draw a circle (ellipse).
func (g *LuaGenerator) DrawCircle(layerName string, frameNumber int, centerX, centerY, radius int, color Color, filled bool, usePalette bool) string {
	var sb strings.Builder

	// Add palette snapper helper if needed
	if usePalette {
		sb.WriteString(GeneratePaletteSnapperHelper())
		sb.WriteString("\n")
	}

	escapedName := EscapeString(layerName)
	tool := "ellipse"
	if filled {
		tool = "filled_ellipse"
	}

	// Calculate bounding rectangle for circle
	x1 := centerX - radius
	y1 := centerY - radius
	x2 := centerX + radius
	y2 := centerY + radius

	sb.WriteString(fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

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

	app.useTool{
		tool = "%s",
		color = %s,
		points = {%s, %s}
	}
end)

spr:saveAs(spr.filename)
print("Circle drawn successfully")`,
		escapedName, escapedName,
		frameNumber, frameNumber,
		tool,
		FormatColorWithPalette(color, usePalette),
		FormatPoint(Point{X: x1, Y: y1}),
		FormatPoint(Point{X: x2, Y: y2})))

	return sb.String()
}

// FillArea generates a Lua script to flood fill an area (paint bucket).
func (g *LuaGenerator) FillArea(layerName string, frameNumber int, x, y int, color Color, tolerance int, usePalette bool) string {
	var sb strings.Builder

	// Add palette snapper helper if needed
	if usePalette {
		sb.WriteString(GeneratePaletteSnapperHelper())
		sb.WriteString("\n")
	}

	escapedName := EscapeString(layerName)
	sb.WriteString(fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

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

	app.useTool{
		tool = "paint_bucket",
		color = %s,
		points = {%s},
		contiguous = true,
		tolerance = %d
	}
end)

spr:saveAs(spr.filename)
print("Area filled successfully")`,
		escapedName, escapedName,
		frameNumber, frameNumber,
		FormatColorWithPalette(color, usePalette),
		FormatPoint(Point{X: x, Y: y}),
		tolerance))

	return sb.String()
}

// ExportSprite generates a Lua script to export a sprite.
func (g *LuaGenerator) ExportSprite(outputPath string, frameNumber int) string {
	escapedPath := EscapeString(outputPath)

	if frameNumber > 0 {
		// Export specific frame by creating a temporary sprite with just that frame
		return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

if #spr.frames < %d then
	error("Frame not found: %d")
end

-- Create a new sprite with the same dimensions and color mode
local tempSpr = Sprite(spr.width, spr.height, spr.colorMode)

-- Copy the specific frame
local targetFrame = spr.frames[%d]
for _, layer in ipairs(spr.layers) do
	local cel = layer:cel(targetFrame)
	if cel then
		local tempLayer = tempSpr.layers[1]
		if not tempLayer then
			tempLayer = tempSpr:newLayer()
		end
		tempSpr:newCel(tempLayer, 1, cel.image, cel.position)
	end
end

-- Flatten and export
app.command.FlattenLayers()
tempSpr:saveAs("%s")
tempSpr:close()

print("Exported successfully")`, frameNumber, frameNumber, frameNumber, escapedPath)
	}

	// Export all frames
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

spr:saveCopyAs("%s")
print("Exported successfully")`, escapedPath)
}

// SetFrameDuration generates a Lua script to set the duration of a frame.
func (g *LuaGenerator) SetFrameDuration(frameNumber int, durationMs int) string {
	durationSec := float64(durationMs) / 1000.0
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

app.transaction(function()
	frame.duration = %.3f
end)

spr:saveAs(spr.filename)
print("Frame duration set successfully")`, frameNumber, frameNumber, durationSec)
}

// CreateTag generates a Lua script to create an animation tag.
func (g *LuaGenerator) CreateTag(tagName string, fromFrame, toFrame int, direction string) string {
	escapedName := EscapeString(tagName)

	// Map direction to Aseprite AniDir enum
	var aniDir string
	switch direction {
	case "forward":
		aniDir = "AniDir.FORWARD"
	case "reverse":
		aniDir = "AniDir.REVERSE"
	case "pingpong":
		aniDir = "AniDir.PING_PONG"
	default:
		aniDir = "AniDir.FORWARD"
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

if #spr.frames < %d then
	error("Frame range exceeds sprite frames")
end

app.transaction(function()
	local tag = spr:newTag(%d, %d)
	tag.name = "%s"
	tag.aniDir = %s
end)

spr:saveAs(spr.filename)
print("Tag created successfully")`, toFrame, fromFrame, toFrame, escapedName, aniDir)
}

// SelectRectangle generates a Lua script to create a rectangular selection.
// Mode can be "replace", "add", "subtract", or "intersect".
func (g *LuaGenerator) SelectRectangle(x, y, width, height int, mode string) string {
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local rect = Rectangle(%d, %d, %d, %d)
local sel = Selection(rect)

if "%s" == "replace" then
	spr.selection = sel
else
	spr.selection:add(sel)
	if "%s" == "subtract" then
		spr.selection:subtract(sel)
	elseif "%s" == "intersect" then
		spr.selection:intersect(sel)
	end
end

-- Don't save - selections are transient and don't persist in .aseprite files
print("Rectangle selection created successfully")`, x, y, width, height, mode, mode, mode)
}

// SelectEllipse generates a Lua script to create an elliptical selection.
// Mode can be "replace", "add", "subtract", or "intersect".
func (g *LuaGenerator) SelectEllipse(x, y, width, height int, mode string) string {
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Create ellipse selection by using drawPixel for each point on ellipse
local sel = Selection()
local rx = %d / 2
local ry = %d / 2
local cx = %d + rx
local cy = %d + ry

-- Midpoint ellipse algorithm to create selection
for angle = 0, 360 do
	local rad = math.rad(angle)
	local ex = math.floor(cx + rx * math.cos(rad))
	local ey = math.floor(cy + ry * math.sin(rad))
	-- Fill from center to edge
	for fillx = math.floor(cx - rx), math.floor(cx + rx) do
		for filly = math.floor(cy - ry), math.floor(cy + ry) do
			local dx = (fillx - cx) / rx
			local dy = (filly - cy) / ry
			if dx * dx + dy * dy <= 1 then
				sel:add(Rectangle(fillx, filly, 1, 1))
			end
		end
	end
	break  -- Only need one pass to fill
end

if "%s" == "replace" then
	spr.selection = sel
elseif "%s" == "add" then
	spr.selection:add(sel)
elseif "%s" == "subtract" then
	spr.selection:subtract(sel)
elseif "%s" == "intersect" then
	spr.selection:intersect(sel)
end

-- Don't save - selections are transient and don't persist in .aseprite files
print("Ellipse selection created successfully")`, width, height, x, y, mode, mode, mode, mode)
}

// SelectAll generates a Lua script to select the entire canvas.
func (g *LuaGenerator) SelectAll() string {
	return `local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Create selection covering entire sprite
local rect = Rectangle(0, 0, spr.width, spr.height)
local sel = Selection(rect)
spr.selection = sel

-- Don't save - selections are transient and don't persist in .aseprite files
print("Select all completed successfully")`
}

// Deselect generates a Lua script to clear the current selection.
func (g *LuaGenerator) Deselect() string {
	return `local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

app.command.DeselectMask()

-- Don't save - selections are transient and don't persist in .aseprite files
print("Deselect completed successfully")`
}

// MoveSelection generates a Lua script to translate the selection bounds.
func (g *LuaGenerator) MoveSelection(dx, dy int) string {
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

if spr.selection.isEmpty then
	error("No active selection to move")
end

local bounds = spr.selection.bounds
local newSel = Selection(Rectangle(bounds.x + %d, bounds.y + %d, bounds.width, bounds.height))
spr.selection = newSel

-- Don't save - selections are transient and don't persist in .aseprite files
print("Selection moved successfully")`, dx, dy)
}

// CutSelection generates a Lua script to cut the selected pixels to clipboard.
func (g *LuaGenerator) CutSelection(layerName string, frameNumber int) string {
	escapedName := EscapeString(layerName)
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

if spr.selection.isEmpty then
	error("No active selection to cut")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

if not layer then
	error("Layer not found: %s")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

app.activeLayer = layer
app.activeFrame = frame

app.command.Cut()

spr:saveAs(spr.filename)
print("Cut selection completed successfully")`, escapedName, escapedName, frameNumber, frameNumber)
}

// CopySelection generates a Lua script to copy the selected pixels to clipboard.
// Note: Copy command may have limitations in batch mode - clipboard state doesn't
// persist across separate Aseprite process invocations.
func (g *LuaGenerator) CopySelection() string {
	return `local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

if spr.selection.isEmpty then
	error("No active selection to copy")
end

-- Ensure we have an active layer and frame for the Copy command
if #spr.layers > 0 then
	app.activeLayer = spr.layers[1]
end
if #spr.frames > 0 then
	app.activeFrame = spr.frames[1]
end

app.command.Copy()

print("Copy selection completed successfully")`
}

// PasteClipboard generates a Lua script to paste clipboard content.
// If x and y are nil, pastes at current position.
func (g *LuaGenerator) PasteClipboard(layerName string, frameNumber int, x, y *int) string {
	escapedName := EscapeString(layerName)

	pasteCommand := "app.command.Paste()"
	if x != nil && y != nil {
		pasteCommand = fmt.Sprintf("app.command.Paste { x = %d, y = %d }", *x, *y)
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

if not layer then
	error("Layer not found: %s")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

app.activeLayer = layer
app.activeFrame = frame

%s

spr:saveAs(spr.filename)
print("Paste completed successfully")`, escapedName, escapedName, frameNumber, frameNumber, pasteCommand)
}

// DuplicateFrame generates a Lua script to duplicate a frame.
func (g *LuaGenerator) DuplicateFrame(sourceFrame int, insertAfter int) string {
	if insertAfter == 0 {
		// Insert at end
		return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local srcFrame = spr.frames[%d]
if not srcFrame then
	error("Source frame not found: %d")
end

local newFrame
app.transaction(function()
	newFrame = spr:newFrame(#spr.frames + 1)
	newFrame.duration = srcFrame.duration

	-- Copy cels from source frame to new frame
	for _, layer in ipairs(spr.layers) do
		local srcCel = layer:cel(srcFrame)
		if srcCel then
			local newCel = spr:newCel(layer, newFrame, srcCel.image, srcCel.position)
		end
	end
end)

spr:saveAs(spr.filename)
print(#spr.frames)`, sourceFrame, sourceFrame)
	}

	// Insert after specific frame
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local srcFrame = spr.frames[%d]
if not srcFrame then
	error("Source frame not found: %d")
end

if #spr.frames < %d then
	error("Insert position exceeds sprite frames")
end

local newFrame
app.transaction(function()
	newFrame = spr:newFrame(%d + 1)
	newFrame.duration = srcFrame.duration

	-- Copy cels from source frame to new frame
	for _, layer in ipairs(spr.layers) do
		local srcCel = layer:cel(srcFrame)
		if srcCel then
			local newCel = spr:newCel(layer, newFrame, srcCel.image, srcCel.position)
		end
	end
end)

spr:saveAs(spr.filename)
print(%d + 1)`, sourceFrame, sourceFrame, insertAfter, insertAfter, insertAfter)
}

// DownsampleImage generates a Lua script to downsample an image using box filter (area averaging).
func (g *LuaGenerator) DownsampleImage(sourcePath, outputPath string, targetWidth, targetHeight int) string {
	escapedSource := EscapeString(sourcePath)
	escapedOutput := EscapeString(outputPath)

	return fmt.Sprintf(`-- Load source image
local srcSprite = app.open("%s")
if not srcSprite then
	error("Failed to open source image: %s")
end

local srcWidth = srcSprite.width
local srcHeight = srcSprite.height
local targetWidth = %d
local targetHeight = %d

-- Get the first layer and frame from source
local srcLayer = srcSprite.layers[1]
if not srcLayer then
	error("Source sprite has no layers")
end

local srcFrame = srcSprite.frames[1]
if not srcFrame then
	error("Source sprite has no frames")
end

local srcCel = srcLayer:cel(srcFrame)
if not srcCel then
	error("Source sprite has no cel in first frame")
end

local srcImage = srcCel.image

-- Create target sprite with same color mode
local targetSprite = Sprite(targetWidth, targetHeight, srcSprite.colorMode)
local targetLayer = targetSprite.layers[1]
local targetFrame = targetSprite.frames[1]
local targetCel = targetSprite:newCel(targetLayer, targetFrame)
local targetImage = Image(targetWidth, targetHeight, srcSprite.colorMode)

-- Calculate scaling factors
local scaleX = srcWidth / targetWidth
local scaleY = srcHeight / targetHeight

-- Downsample using box filter (area averaging)
for ty = 0, targetHeight - 1 do
	for tx = 0, targetWidth - 1 do
		-- Calculate source region bounds
		local sx1 = math.floor(tx * scaleX)
		local sy1 = math.floor(ty * scaleY)
		local sx2 = math.floor((tx + 1) * scaleX)
		local sy2 = math.floor((ty + 1) * scaleY)

		-- Clamp to source image bounds
		sx2 = math.min(sx2, srcWidth)
		sy2 = math.min(sy2, srcHeight)

		-- Average all pixels in the source region
		local sumR, sumG, sumB, sumA = 0, 0, 0, 0
		local count = 0

		for sy = sy1, sy2 - 1 do
			for sx = sx1, sx2 - 1 do
				local pixel = srcImage:getPixel(sx, sy)
				sumR = sumR + app.pixelColor.rgbaR(pixel)
				sumG = sumG + app.pixelColor.rgbaG(pixel)
				sumB = sumB + app.pixelColor.rgbaB(pixel)
				sumA = sumA + app.pixelColor.rgbaA(pixel)
				count = count + 1
			end
		end

		-- Calculate average color
		local avgR = math.floor(sumR / count + 0.5)
		local avgG = math.floor(sumG / count + 0.5)
		local avgB = math.floor(sumB / count + 0.5)
		local avgA = math.floor(sumA / count + 0.5)

		-- Set target pixel
		local color = app.pixelColor.rgba(avgR, avgG, avgB, avgA)
		targetImage:drawPixel(tx, ty, color)
	end
end

-- Assign image to cel
targetCel.image = targetImage

-- Save target sprite
targetSprite:saveAs("%s")

-- Close sprites
targetSprite:close()
srcSprite:close()

-- Output the result path
print("%s")`,
		escapedSource, escapedSource,
		targetWidth, targetHeight,
		escapedOutput, escapedOutput)
}

// SetPalette generates a Lua script to set the sprite's palette to the specified colors.
func (g *LuaGenerator) SetPalette(colors []string) string {
	if len(colors) == 0 {
		return `error("No colors provided for palette")`
	}

	// Build palette color list
	colorList := "{\n"
	for i, hexColor := range colors {
		// Parse hex color #RRGGBB
		hexColor = strings.TrimPrefix(hexColor, "#")
		if len(hexColor) != 6 && len(hexColor) != 8 {
			continue
		}

		var r, g, b, a int
		// Parse hex color components (errors ignored as format is validated above)
		_, _ = fmt.Sscanf(hexColor[:2], "%x", &r)
		_, _ = fmt.Sscanf(hexColor[2:4], "%x", &g)
		_, _ = fmt.Sscanf(hexColor[4:6], "%x", &b)
		if len(hexColor) == 8 {
			_, _ = fmt.Sscanf(hexColor[6:8], "%x", &a)
		} else {
			a = 255
		}

		colorList += fmt.Sprintf("\t\tColor{r=%d, g=%d, b=%d, a=%d}", r, g, b, a)
		if i < len(colors)-1 {
			colorList += ","
		}
		colorList += "\n"
	}
	colorList += "\t}"

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Get or create palette
local palette = spr.palettes[1]
if not palette then
	error("No palette found")
end

-- Resize palette to match color count
palette:resize(%d)

-- Set palette colors
local colors = %s

for i, color in ipairs(colors) do
	palette:setColor(i - 1, color)  -- Palette is 0-indexed
end

spr:saveAs(spr.filename)
print("Palette set successfully")`, len(colors), colorList)
}

// GetPalette generates a Lua script to retrieve the sprite's current palette.
// Returns a JSON object with colors array and size.
func (g *LuaGenerator) GetPalette() string {
	return `local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Get palette
local palette = spr.palettes[1]
if not palette then
	error("No palette found")
end

-- Extract colors as hex strings
local colors = {}
for i = 0, #palette - 1 do
	local color = palette:getColor(i)
	local hex = string.format("#%02X%02X%02X", color.red, color.green, color.blue)
	table.insert(colors, hex)
end

-- Format as JSON
local colorList = '["' .. table.concat(colors, '","') .. '"]'
local output = string.format('{"colors":%s,"size":%d}', colorList, #palette)

print(output)`
}

// SetPaletteColor generates a Lua script to set a specific palette color at an index.
func (g *LuaGenerator) SetPaletteColor(index int, hexColor string) string {
	// Parse hex color #RRGGBB
	hexColor = strings.TrimPrefix(hexColor, "#")

	var red, green, blue, alpha int
	// Parse hex color components (errors ignored as format is validated by caller)
	_, _ = fmt.Sscanf(hexColor[:2], "%x", &red)
	_, _ = fmt.Sscanf(hexColor[2:4], "%x", &green)
	_, _ = fmt.Sscanf(hexColor[4:6], "%x", &blue)
	if len(hexColor) == 8 {
		_, _ = fmt.Sscanf(hexColor[6:8], "%x", &alpha)
	} else {
		alpha = 255
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Get palette
local palette = spr.palettes[1]
if not palette then
	error("No palette found")
end

-- Validate index
if %d < 0 or %d >= #palette then
	error(string.format("Palette index %%d out of range (palette has %%d colors)", %d, #palette))
end

-- Set color at index
palette:setColor(%d, Color{r=%d, g=%d, b=%d, a=%d})

spr:saveAs(spr.filename)
print("Palette color set successfully")`, index, index, index, index, red, green, blue, alpha)
}

// AddPaletteColor generates a Lua script to add a new color to the palette.
// Returns the index of the newly added color.
func (g *LuaGenerator) AddPaletteColor(hexColor string) string {
	// Parse hex color #RRGGBB
	hexColor = strings.TrimPrefix(hexColor, "#")

	var red, green, blue, alpha int
	// Parse hex color components (errors ignored as format is validated by caller)
	_, _ = fmt.Sscanf(hexColor[:2], "%x", &red)
	_, _ = fmt.Sscanf(hexColor[2:4], "%x", &green)
	_, _ = fmt.Sscanf(hexColor[4:6], "%x", &blue)
	if len(hexColor) == 8 {
		_, _ = fmt.Sscanf(hexColor[6:8], "%x", &alpha)
	} else {
		alpha = 255
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Get palette
local palette = spr.palettes[1]
if not palette then
	error("No palette found")
end

-- Check if palette is at maximum size
if #palette >= 256 then
	error("Palette is already at maximum size (256 colors)")
end

-- Add new color to palette
local newIndex = #palette
palette:resize(#palette + 1)
palette:setColor(newIndex, Color{r=%d, g=%d, b=%d, a=%d})

spr:saveAs(spr.filename)

-- Output JSON with color_index
local output = string.format('{"color_index":%%d}', newIndex)
print(output)`, red, green, blue, alpha)
}

// SortPalette generates a Lua script to sort the palette by a specified method.
// Methods: "hue", "saturation", "brightness", "luminance"
func (g *LuaGenerator) SortPalette(method string, ascending bool) string {
	// Helper function to convert RGB to HSL in Lua
	rgbToHSL := `-- Convert RGB to HSL
local function rgbToHSL(r, g, b)
	r, g, b = r / 255, g / 255, b / 255
	local max = math.max(r, g, b)
	local min = math.min(r, g, b)
	local delta = max - min

	local h, s, l = 0, 0, (max + min) / 2

	if delta ~= 0 then
		-- Calculate saturation
		if l < 0.5 then
			s = delta / (max + min)
		else
			s = delta / (2.0 - max - min)
		end

		-- Calculate hue
		if max == r then
			h = ((g - b) / delta)
			if g < b then
				h = h + 6.0
			end
		elseif max == g then
			h = ((b - r) / delta) + 2.0
		elseif max == b then
			h = ((r - g) / delta) + 4.0
		end

		h = h * 60.0
	end

	return h, s, l
end`

	// Sort function based on method
	var sortKey string
	switch method {
	case "hue":
		sortKey = "h"
	case "saturation":
		sortKey = "s"
	case "brightness", "luminance":
		sortKey = "l"
	default:
		sortKey = "h"
	}

	sortDirection := ">"
	if ascending {
		sortDirection = "<"
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Get palette
local palette = spr.palettes[1]
if not palette then
	error("No palette found")
end

%s

-- Extract colors with HSL values
local colors = {}
for i = 0, #palette - 1 do
	local color = palette:getColor(i)
	local h, s, l = rgbToHSL(color.red, color.green, color.blue)
	table.insert(colors, {
		index = i,
		color = color,
		h = h,
		s = s,
		l = l
	})
end

-- Sort colors by %s
table.sort(colors, function(a, b)
	return a.%s %s b.%s
end)

-- Apply sorted colors back to palette
for i, entry in ipairs(colors) do
	palette:setColor(i - 1, entry.color)
end

spr:saveAs(spr.filename)
print("Palette sorted by %s successfully")`, rgbToHSL, method, sortKey, sortDirection, sortKey, method)
}

// DrawWithDither generates a Lua script to fill a region with a dithering pattern.
func (g *LuaGenerator) DrawWithDither(layerName string, frameNumber int, x, y, width, height int, color1, color2 string, pattern string, density float64) string {
	escapedLayerName := EscapeString(layerName)

	// Parse hex colors
	c1 := parseHexColor(color1)
	c2 := parseHexColor(color2)

	// Get dithering matrix based on pattern
	var matrixCode string
	switch pattern {
	case "bayer_2x2":
		matrixCode = `local matrix = {{0, 2}, {3, 1}}
local matrixSize = 2`
	case "bayer_4x4":
		matrixCode = `local matrix = {
	{ 0,  8,  2, 10},
	{12,  4, 14,  6},
	{ 3, 11,  1,  9},
	{15,  7, 13,  5}
}
local matrixSize = 4`
	case "bayer_8x8":
		matrixCode = `local matrix = {
	{ 0, 32,  8, 40,  2, 34, 10, 42},
	{48, 16, 56, 24, 50, 18, 58, 26},
	{12, 44,  4, 36, 14, 46,  6, 38},
	{60, 28, 52, 20, 62, 30, 54, 22},
	{ 3, 35, 11, 43,  1, 33,  9, 41},
	{51, 19, 59, 27, 49, 17, 57, 25},
	{15, 47,  7, 39, 13, 45,  5, 37},
	{63, 31, 55, 23, 61, 29, 53, 21}
}
local matrixSize = 8`
	case "checkerboard":
		matrixCode = `local matrix = {{0, 1}, {1, 0}}
local matrixSize = 2`
	case "grass":
		matrixCode = `local matrix = {
	{1, 0, 1, 0, 1, 0},
	{0, 1, 1, 0, 0, 1},
	{1, 1, 0, 1, 0, 0},
	{0, 1, 0, 1, 1, 0},
	{1, 0, 0, 0, 1, 1},
	{0, 0, 1, 1, 0, 1}
}
local matrixSize = 6`
	case "water":
		matrixCode = `local matrix = {
	{0, 0, 1, 1, 0, 0},
	{0, 1, 1, 1, 1, 0},
	{1, 1, 0, 0, 1, 1},
	{1, 0, 0, 0, 0, 1},
	{0, 1, 1, 1, 1, 0},
	{0, 0, 1, 1, 0, 0}
}
local matrixSize = 6`
	case "stone":
		matrixCode = `local matrix = {
	{0, 0, 0, 1, 1, 0},
	{0, 1, 0, 0, 1, 1},
	{0, 0, 1, 1, 0, 0},
	{1, 1, 0, 0, 0, 1},
	{1, 0, 0, 1, 1, 0},
	{0, 1, 1, 0, 0, 0}
}
local matrixSize = 6`
	case "cloud":
		matrixCode = `local matrix = {
	{0, 0, 0, 0, 1, 1},
	{0, 0, 0, 1, 1, 1},
	{0, 0, 1, 1, 1, 0},
	{0, 1, 1, 1, 0, 0},
	{1, 1, 1, 0, 0, 0},
	{1, 1, 0, 0, 0, 0}
}
local matrixSize = 6`
	case "brick":
		matrixCode = `local matrix = {
	{0, 0, 0, 0, 1, 0, 0, 0},
	{0, 0, 0, 0, 1, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 1, 1},
	{0, 0, 1, 0, 0, 0, 0, 1},
	{0, 0, 1, 0, 0, 0, 0, 1},
	{1, 1, 1, 1, 1, 1, 1, 1},
	{0, 0, 0, 0, 1, 0, 0, 0},
	{0, 0, 0, 0, 1, 0, 0, 0}
}
local matrixSize = 8`
	case "dots":
		matrixCode = `local matrix = {
	{1, 0, 0, 0},
	{0, 0, 0, 0},
	{0, 0, 1, 0},
	{0, 0, 0, 0}
}
local matrixSize = 4`
	case "diagonal":
		matrixCode = `local matrix = {
	{1, 0, 0, 0},
	{0, 1, 0, 0},
	{0, 0, 1, 0},
	{0, 0, 0, 1}
}
local matrixSize = 4`
	case "cross":
		matrixCode = `local matrix = {
	{0, 1, 0},
	{1, 1, 1},
	{0, 1, 0}
}
local matrixSize = 3`
	case "noise":
		matrixCode = `local matrix = {
	{1, 0, 1, 0, 0, 1},
	{0, 1, 0, 1, 1, 0},
	{1, 0, 0, 1, 0, 1},
	{0, 1, 1, 0, 1, 0},
	{0, 0, 1, 0, 1, 1},
	{1, 1, 0, 1, 0, 0}
}
local matrixSize = 6`
	case "horizontal_lines":
		matrixCode = `local matrix = {
	{1, 1, 1, 1},
	{0, 0, 0, 0},
	{1, 1, 1, 1},
	{0, 0, 0, 0}
}
local matrixSize = 4`
	case "vertical_lines":
		matrixCode = `local matrix = {
	{1, 0, 1, 0},
	{1, 0, 1, 0},
	{1, 0, 1, 0},
	{1, 0, 1, 0}
}
local matrixSize = 4`
	default:
		return fmt.Sprintf(`error("Unknown dithering pattern: %s")`, pattern)
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer
local layer = nil
for _, l in ipairs(spr.layers) do
	if l.name == "%s" then
		layer = l
		break
	end
end

if not layer then
	error("Layer not found: %s")
end

-- Get frame
local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

-- Get or create cel
local cel = layer:cel(frame)
if not cel then
	cel = spr:newCel(layer, frame)
end

-- Create or get image
local img = cel.image
if not img then
	img = Image(spr.width, spr.height, spr.colorMode)
	cel.image = img
end

-- Define colors
local color1 = app.pixelColor.rgba(%d, %d, %d, %d)
local color2 = app.pixelColor.rgba(%d, %d, %d, %d)

-- Dithering matrix
%s

-- Dithering threshold (based on density)
local threshold = %f * (matrixSize * matrixSize)

-- Apply dithering pattern
app.transaction(function()
	for py = 0, %d - 1 do
		for px = 0, %d - 1 do
			local mx = (px %% matrixSize) + 1
			local my = (py %% matrixSize) + 1
			local matrixValue = matrix[my][mx]

			local useColor2 = matrixValue < threshold
			local finalColor = useColor2 and color2 or color1

			img:drawPixel(%d + px, %d + py, finalColor)
		end
	end
end)

spr:saveAs(spr.filename)
print("Dithering applied successfully")`,
		escapedLayerName, escapedLayerName,
		frameNumber, frameNumber,
		c1.R, c1.G, c1.B, c1.A,
		c2.R, c2.G, c2.B, c2.A,
		matrixCode,
		density,
		height, width,
		x, y)
}

// parseHexColor parses a hex color string and returns RGBA components.
func parseHexColor(hexColor string) Color {
	hexColor = strings.TrimPrefix(hexColor, "#")
	if len(hexColor) != 6 && len(hexColor) != 8 {
		return Color{R: 0, G: 0, B: 0, A: 255}
	}

	var r, g, b, a int
	// Parse hex color components (errors ignored as format is validated above)
	_, _ = fmt.Sscanf(hexColor[:2], "%x", &r)
	_, _ = fmt.Sscanf(hexColor[2:4], "%x", &g)
	_, _ = fmt.Sscanf(hexColor[4:6], "%x", &b)
	if len(hexColor) == 8 {
		_, _ = fmt.Sscanf(hexColor[6:8], "%x", &a)
	} else {
		a = 255
	}

	return Color{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

// LinkCel generates a Lua script to create a linked cel.
func (g *LuaGenerator) LinkCel(layerName string, sourceFrame, targetFrame int) string {
	escapedName := EscapeString(layerName)
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

if not layer then
	error("Layer not found: %s")
end

local srcFrame = spr.frames[%d]
if not srcFrame then
	error("Source frame not found: %d")
end

local tgtFrame = spr.frames[%d]
if not tgtFrame then
	error("Target frame not found: %d")
end

local srcCel = layer:cel(srcFrame)
if not srcCel then
	error("Source cel not found in frame %d")
end

app.transaction(function()
	-- Create linked cel by copying with same image reference
	spr:newCel(layer, tgtFrame, srcCel.image, srcCel.position)
end)

spr:saveAs(spr.filename)
print("Cel linked successfully")`,
		escapedName, escapedName,
		sourceFrame, sourceFrame,
		targetFrame, targetFrame,
		sourceFrame)
}

// ApplyShading generates a Lua script to apply palette-constrained shading based on light direction.
func (g *LuaGenerator) ApplyShading(layerName string, frameNumber int, x, y, width, height int, palette []string, lightDirection string, intensity float64, style string) string {
	escapedLayerName := EscapeString(layerName)

	// Build palette color list
	paletteColors := "{\n"
	for i, hexColor := range palette {
		c := parseHexColor(hexColor)
		paletteColors += fmt.Sprintf("\t\t{r=%d, g=%d, b=%d, a=%d}", c.R, c.G, c.B, c.A)
		if i < len(palette)-1 {
			paletteColors += ","
		}
		paletteColors += "\n"
	}
	paletteColors += "\t}"

	// Map light direction to offset vectors
	var dx, dy int
	switch lightDirection {
	case "top_left":
		dx, dy = -1, -1
	case "top":
		dx, dy = 0, -1
	case "top_right":
		dx, dy = 1, -1
	case "left":
		dx, dy = -1, 0
	case "right":
		dx, dy = 1, 0
	case "bottom_left":
		dx, dy = -1, 1
	case "bottom":
		dx, dy = 0, 1
	case "bottom_right":
		dx, dy = 1, 1
	default:
		dx, dy = -1, -1 // Default to top_left
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer
local layer = nil
for _, l in ipairs(spr.layers) do
	if l.name == "%s" then
		layer = l
		break
	end
end

if not layer then
	error("Layer not found: %s")
end

-- Get frame
local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

-- Get or create cel
local cel = layer:cel(frame)
if not cel then
	error("No cel found in frame %d")
end

local img = cel.image

-- Palette colors (ordered darkest to lightest)
local palette = %s

-- Light direction offset
local lightDx = %d
local lightDy = %d
local intensity = %f
local style = "%s"

-- Helper: Find nearest palette color
local function findNearestPaletteColor(r, g, b, a)
	local minDist = math.huge
	local nearestColor = palette[1]

	for _, palColor in ipairs(palette) do
		local dr = r - palColor.r
		local dg = g - palColor.g
		local db = b - palColor.b
		local da = a - palColor.a
		local dist = dr*dr + dg*dg + db*db + da*da

		if dist < minDist then
			minDist = dist
			nearestColor = palColor
		end
	end

	return app.pixelColor.rgba(nearestColor.r, nearestColor.g, nearestColor.b, nearestColor.a)
end

-- Helper: Calculate shading factor based on position and light direction
local function calculateShadingFactor(px, py, regionX, regionY, regionW, regionH)
	-- Normalize position to 0-1
	local normX = (px - regionX) / regionW
	local normY = (py - regionY) / regionH

	-- Calculate distance from light source direction
	local lightFactor = 0.5
	if lightDx < 0 then
		lightFactor = lightFactor + (1 - normX) * 0.5
	elseif lightDx > 0 then
		lightFactor = lightFactor + normX * 0.5
	end

	if lightDy < 0 then
		lightFactor = lightFactor + (1 - normY) * 0.5
	elseif lightDy > 0 then
		lightFactor = lightFactor + normY * 0.5
	end

	-- Normalize to 0-1 range
	lightFactor = math.max(0, math.min(1, lightFactor))

	return lightFactor
end

-- Helper: Apply shading to color
local function applyShading(pixelValue, shadingFactor)
	local r = app.pixelColor.rgbaR(pixelValue)
	local g = app.pixelColor.rgbaG(pixelValue)
	local b = app.pixelColor.rgbaB(pixelValue)
	local a = app.pixelColor.rgbaA(pixelValue)

	if a == 0 then
		return pixelValue  -- Skip transparent pixels
	end

	-- Apply shading based on style
	local shadedR, shadedG, shadedB

	if style == "hard" then
		-- Hard shading: quantize to palette steps
		local paletteIndex = math.floor(shadingFactor * (#palette - 1)) + 1
		paletteIndex = math.max(1, math.min(#palette, paletteIndex))
		return app.pixelColor.rgba(palette[paletteIndex].r, palette[paletteIndex].g, palette[paletteIndex].b, a)
	elseif style == "smooth" then
		-- Smooth shading: blend current color toward palette extremes
		local targetBrightness = shadingFactor
		local currentBrightness = (0.299 * r + 0.587 * g + 0.114 * b) / 255

		-- Blend toward target brightness
		local blend = intensity * 0.5
		shadedR = math.floor(r * (1 - blend) + r * (targetBrightness / math.max(0.01, currentBrightness)) * blend)
		shadedG = math.floor(g * (1 - blend) + g * (targetBrightness / math.max(0.01, currentBrightness)) * blend)
		shadedB = math.floor(b * (1 - blend) + b * (targetBrightness / math.max(0.01, currentBrightness)) * blend)
	else  -- pillow (avoid - but included for completeness)
		-- Pillow shading: center highlight regardless of light direction
		local centerX = 0.5
		local centerY = 0.5
		local distFromCenter = math.sqrt((shadingFactor - centerX)^2 + (shadingFactor - centerY)^2)
		local pillow = 1 - distFromCenter

		shadedR = math.floor(r * (1 + pillow * intensity))
		shadedG = math.floor(g * (1 + pillow * intensity))
		shadedB = math.floor(b * (1 + pillow * intensity))
	end

	-- Clamp values
	shadedR = math.max(0, math.min(255, shadedR))
	shadedG = math.max(0, math.min(255, shadedG))
	shadedB = math.max(0, math.min(255, shadedB))

	-- Find nearest palette color
	return findNearestPaletteColor(shadedR, shadedG, shadedB, a)
end

-- Apply shading to region
local affectedPixels = 0

app.transaction(function()
	for py = %d, %d do
		for px = %d, %d do
			-- Adjust coordinates relative to cel position
			local imgX = px - cel.position.x
			local imgY = py - cel.position.y

			-- Check if coordinates are within image bounds
			if imgX >= 0 and imgX < img.width and imgY >= 0 and imgY < img.height then
				local pixelValue = img:getPixel(imgX, imgY)
				local alpha = app.pixelColor.rgbaA(pixelValue)

				if alpha > 0 then
					local shadingFactor = calculateShadingFactor(px, py, %d, %d, %d, %d)
					local shadedColor = applyShading(pixelValue, shadingFactor)
					img:drawPixel(imgX, imgY, shadedColor)
					affectedPixels = affectedPixels + 1
				end
			end
		end
	end
end)

spr:saveAs(spr.filename)
print(string.format("Shading applied to %%d pixels", affectedPixels))`,
		escapedLayerName, escapedLayerName,
		frameNumber, frameNumber, frameNumber,
		paletteColors,
		dx, dy, intensity, style,
		y, y+height-1, x, x+width-1,
		x, y, width, height)
}

// GetPixels generates a Lua script to read pixel data from a rectangular region.
func (g *LuaGenerator) GetPixels(layerName string, frameNumber int, x, y, width, height int) string {
	return g.GetPixelsWithPagination(layerName, frameNumber, x, y, width, height, 0, 0)
}

func (g *LuaGenerator) GetPixelsWithPagination(layerName string, frameNumber int, x, y, width, height int, offset int, limit int) string {
	escapedName := EscapeString(layerName)
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

if not layer then
	error("Layer not found: %s")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

local cel = layer:cel(frame)
if not cel then
	error("No cel found in frame %d")
end

local img = cel.image
local offset = %d
local limit = %d

-- Calculate pagination bounds
local pixelIndex = 0
local startIdx = offset
local endIdx = (limit > 0) and (offset + limit - 1) or -1

local pixels = {}

-- Read pixels from the specified region
for py = %d, %d do
	for px = %d, %d do
		-- Check if we're within the pagination range
		if endIdx < 0 or (pixelIndex >= startIdx and pixelIndex <= endIdx) then
			-- Adjust coordinates relative to cel position
			local imgX = px - cel.position.x
			local imgY = py - cel.position.y

			-- Check if coordinates are within image bounds
			if imgX >= 0 and imgX < img.width and imgY >= 0 and imgY < img.height then
				local pixelValue = img:getPixel(imgX, imgY)

				-- Convert pixel value to RGBA
				local r = app.pixelColor.rgbaR(pixelValue)
				local g = app.pixelColor.rgbaG(pixelValue)
				local b = app.pixelColor.rgbaB(pixelValue)
				local a = app.pixelColor.rgbaA(pixelValue)

				-- Store as hex color
				local hexColor = string.format("#%%02X%%02X%%02X%%02X", r, g, b, a)
				table.insert(pixels, string.format('{"x":%%d,"y":%%d,"color":"%%s"}', px, py, hexColor))
			end
		end

		pixelIndex = pixelIndex + 1

		-- Early exit if we've collected enough pixels
		if limit > 0 and #pixels >= limit then
			break
		end
	end

	-- Early exit from outer loop
	if limit > 0 and #pixels >= limit then
		break
	end
end

-- Output as JSON array
print("[" .. table.concat(pixels, ",") .. "]")`,
		escapedName, escapedName,
		frameNumber, frameNumber, frameNumber,
		offset, limit,
		y, y+height-1, x, x+width-1)
}

// FlipSprite generates a Lua script to flip a sprite, layer, or cel.
func (g *LuaGenerator) FlipSprite(direction, target string) string {
	// Validate direction
	if direction != "horizontal" && direction != "vertical" {
		direction = "horizontal"
	}

	// Map target to Aseprite's target enum
	targetParam := ""
	switch target {
	case "sprite":
		targetParam = "target = 'sprite'"
	case "layer":
		targetParam = "target = 'layer'"
	case "cel":
		targetParam = "target = 'cel'"
	default:
		targetParam = "target = 'sprite'"
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

app.transaction(function()
	app.command.Flip{
		orientation = "%s",
		%s
	}
end)

spr:saveAs(spr.filename)
print("Sprite flipped %s successfully")`, direction, targetParam, direction)
}

// RotateSprite generates a Lua script to rotate a sprite, layer, or cel.
func (g *LuaGenerator) RotateSprite(angle int, target string) string {
	// Validate angle (must be 90, 180, or 270)
	if angle != 90 && angle != 180 && angle != 270 {
		angle = 90
	}

	// Map target to Aseprite's target enum
	targetParam := ""
	switch target {
	case "sprite":
		targetParam = "target = 'sprite'"
	case "layer":
		targetParam = "target = 'layer'"
	case "cel":
		targetParam = "target = 'cel'"
	default:
		targetParam = "target = 'sprite'"
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

app.transaction(function()
	app.command.Rotate{
		angle = %d,
		%s
	}
end)

spr:saveAs(spr.filename)
print("Sprite rotated %d degrees successfully")`, angle, targetParam, angle)
}

// ScaleSprite generates a Lua script to scale a sprite with a specified algorithm.
func (g *LuaGenerator) ScaleSprite(scaleX, scaleY float64, algorithm string) string {
	// Validate algorithm
	validAlgorithms := map[string]bool{
		"nearest":   true,
		"bilinear":  true,
		"rotsprite": true,
	}
	if !validAlgorithms[algorithm] {
		algorithm = "nearest"
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local oldWidth = spr.width
local oldHeight = spr.height
local newWidth = math.floor(oldWidth * %.3f)
local newHeight = math.floor(oldHeight * %.3f)

app.transaction(function()
	app.command.SpriteSize{
		width = newWidth,
		height = newHeight,
		method = "%s"
	}
end)

spr:saveAs(spr.filename)

-- Output JSON with new dimensions
local output = string.format('{"success":true,"new_width":%%d,"new_height":%%d}', spr.width, spr.height)
print(output)`, scaleX, scaleY, algorithm)
}

// CropSprite generates a Lua script to crop a sprite to a rectangular region.
func (g *LuaGenerator) CropSprite(x, y, width, height int) string {
	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Validate crop bounds
if %d < 0 or %d < 0 then
	error("Crop position must be non-negative")
end

if %d <= 0 or %d <= 0 then
	error("Crop dimensions must be positive")
end

if %d + %d > spr.width or %d + %d > spr.height then
	error(string.format("Crop bounds exceed sprite dimensions (sprite: %%dx%%d, crop: %%d,%%d,%%dx%%d)",
		spr.width, spr.height, %d, %d, %d, %d))
end

app.transaction(function()
	app.command.CropSprite{
		bounds = Rectangle(%d, %d, %d, %d)
	}
end)

spr:saveAs(spr.filename)
print("Sprite cropped successfully")`,
		x, y,
		width, height,
		x, width, y, height,
		x, y, width, height,
		x, y, width, height)
}

// ResizeCanvas generates a Lua script to resize the canvas without scaling content.
func (g *LuaGenerator) ResizeCanvas(width, height int, anchor string) string {
	// Calculate anchor offset based on anchor position
	var leftOffset, topOffset string

	switch anchor {
	case "top_left":
		leftOffset = "0"
		topOffset = "0"
	case "top_right":
		leftOffset = "newWidth - oldWidth"
		topOffset = "0"
	case "bottom_left":
		leftOffset = "0"
		topOffset = "newHeight - oldHeight"
	case "bottom_right":
		leftOffset = "newWidth - oldWidth"
		topOffset = "newHeight - oldHeight"
	case "center":
		fallthrough
	default:
		leftOffset = "math.floor((newWidth - oldWidth) / 2)"
		topOffset = "math.floor((newHeight - oldHeight) / 2)"
	}

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

local oldWidth = spr.width
local oldHeight = spr.height
local newWidth = %d
local newHeight = %d

app.transaction(function()
	app.command.CanvasSize{
		left = %s,
		top = %s,
		right = 0,
		bottom = 0,
		width = newWidth,
		height = newHeight
	}
end)

spr:saveAs(spr.filename)
print("Canvas resized successfully")`, width, height, leftOffset, topOffset)
}

// ApplyOutline generates a Lua script to apply an outline effect to a layer.
func (g *LuaGenerator) ApplyOutline(layerName string, frameNumber int, color Color, thickness int) string {
	escapedName := EscapeString(layerName)

	return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

-- Find layer by name
local layer = nil
for i, lyr in ipairs(spr.layers) do
	if lyr.name == "%s" then
		layer = lyr
		break
	end
end

if not layer then
	error("Layer not found: %s")
end

local frame = spr.frames[%d]
if not frame then
	error("Frame not found: %d")
end

-- Set active layer and frame
app.activeLayer = layer
app.activeFrame = frame

-- Check if cel exists
local cel = layer:cel(frame)
if not cel then
	error("No cel found at layer '%s' frame %d")
end

app.transaction(function()
	app.command.Outline{
		color = %s,
		size = %d,
		bgColor = Color(0, 0, 0, 0)
	}
end)

spr:saveAs(spr.filename)
print("Outline applied successfully")`,
		escapedName, escapedName,
		frameNumber, frameNumber,
		escapedName, frameNumber,
		FormatColor(color), thickness)
}
