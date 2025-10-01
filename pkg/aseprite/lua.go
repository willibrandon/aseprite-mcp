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
		// Export specific frame by temporarily setting the active frame and using saveCopyAs
		// Note: app.command.SaveFileCopyAs produces blank PNGs, so we use this approach instead
		return fmt.Sprintf(`local spr = app.activeSprite
if not spr then
	error("No active sprite")
end

if #spr.frames < %d then
	error("Frame not found: %d")
end

-- Save the current active frame
local originalFrame = app.activeFrame

-- Set the target frame as active
app.activeFrame = spr.frames[%d]

-- Export using saveCopyAs (which respects the active frame)
spr:saveCopyAs("%s")

-- Restore the original active frame
app.activeFrame = originalFrame

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
