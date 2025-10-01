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
func (g *LuaGenerator) DrawPixels(layerName string, frameNumber int, pixels []Pixel) string {
	var sb strings.Builder

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
		FormatColor(color),
		FormatPoint(Point{X: x1, Y: y1}),
		FormatPoint(Point{X: x2, Y: y2}))
}

// DrawRectangle generates a Lua script to draw a rectangle.
func (g *LuaGenerator) DrawRectangle(layerName string, frameNumber int, x, y, width, height int, color Color, filled bool) string {
	escapedName := EscapeString(layerName)
	tool := "rectangle"
	if filled {
		tool = "filled_rectangle"
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
		FormatColor(color),
		FormatPoint(Point{X: x, Y: y}),
		FormatPoint(Point{X: x + width - 1, Y: y + height - 1}))
}

// DrawCircle generates a Lua script to draw a circle (ellipse).
func (g *LuaGenerator) DrawCircle(layerName string, frameNumber int, centerX, centerY, radius int, color Color, filled bool) string {
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
		FormatColor(color),
		FormatPoint(Point{X: x1, Y: y1}),
		FormatPoint(Point{X: x2, Y: y2}))
}

// FillArea generates a Lua script to flood fill an area (paint bucket).
func (g *LuaGenerator) FillArea(layerName string, frameNumber int, x, y int, color Color, tolerance int) string {
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
		FormatColor(color),
		FormatPoint(Point{X: x, Y: y}),
		tolerance)
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
