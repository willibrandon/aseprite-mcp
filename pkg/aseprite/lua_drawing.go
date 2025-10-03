package aseprite

import (
	"fmt"
	"strings"
)

// DrawPixels generates a Lua script to draw multiple pixels.
//
// Draws a batch of pixels to the specified layer and frame. This is the most
// efficient way to draw multiple pixels as they are all committed in a single transaction.
//
// Parameters:
//   - layerName: name of the target layer (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to draw on
//   - pixels: slice of Pixel structs containing {X, Y, Color} data
//   - usePalette: if true, snaps each pixel color to nearest palette color
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after pixels are drawn. If the specified layer has no cel at the
// target frame, a new cel is created automatically.
//
// Prints "Pixels drawn successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
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
//
// Draws a straight line between two points using Aseprite's line tool.
// The line is anti-aliased and uses the specified brush thickness.
//
// Parameters:
//   - layerName: name of the target layer (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to draw on
//   - x1, y1: starting point coordinates
//   - x2, y2: ending point coordinates
//   - color: line color in RGBA format
//   - thickness: brush size in pixels (1 = single pixel line)
//   - usePalette: if true, snaps color to nearest palette color
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the line is drawn.
//
// Prints "Line drawn successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
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
//
// Draws a series of connected line segments forming either an open polyline or
// closed polygon. This is useful for creating complex shapes like hand-drawn paths,
// borders, or traced outlines.
//
// Parameters:
//   - layerName: name of the target layer (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to draw on
//   - points: ordered slice of Point structs defining the path vertices
//   - color: line color in RGBA format
//   - thickness: brush size in pixels (1 = single pixel line)
//   - closed: if true, connects the last point back to the first to form a polygon
//   - usePalette: if true, snaps color to nearest palette color
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the contour is drawn. Each segment is drawn as a separate line
// to ensure proper vertex connection.
//
// Prints "Contour drawn successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
//   - The points slice is empty
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
//
// Draws either a filled or outlined rectangle using Aseprite's rectangle tools.
// The rectangle is defined by its top-left corner position and dimensions.
//
// Parameters:
//   - layerName: name of the target layer (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to draw on
//   - x, y: top-left corner coordinates
//   - width: rectangle width in pixels (must be positive)
//   - height: rectangle height in pixels (must be positive)
//   - color: fill/stroke color in RGBA format
//   - filled: if true, uses filled_rectangle tool; if false, draws outline only
//   - usePalette: if true, snaps color to nearest palette color
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the rectangle is drawn.
//
// Prints "Rectangle drawn successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
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
//
// Draws either a filled or outlined circle using Aseprite's ellipse tools.
// The circle is defined by its center point and radius. Note that this actually
// draws a perfect circle, not an arbitrary ellipse.
//
// Parameters:
//   - layerName: name of the target layer (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to draw on
//   - centerX, centerY: center point coordinates
//   - radius: circle radius in pixels (must be positive)
//   - color: fill/stroke color in RGBA format
//   - filled: if true, uses filled_ellipse tool; if false, draws outline only
//   - usePalette: if true, snaps color to nearest palette color
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the circle is drawn.
//
// Prints "Circle drawn successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
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
//
// Performs a flood fill operation starting from the specified point, replacing
// all contiguous pixels of similar color with the target color. The tolerance
// parameter controls how similar colors must be to be considered the same.
//
// Parameters:
//   - layerName: name of the target layer (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to draw on
//   - x, y: starting point coordinates for the flood fill
//   - color: fill color in RGBA format
//   - tolerance: color similarity threshold (0-255, where 0 = exact match only)
//   - usePalette: if true, snaps color to nearest palette color
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the fill is complete. The fill is contiguous, meaning it only
// affects connected pixels, not all pixels of the same color.
//
// Prints "Area filled successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
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

// DrawWithDither generates a Lua script to fill a region with a dithering pattern.
//
// Applies a dithering pattern to create texture, gradients, or retro aesthetic effects.
// Supports 15 different patterns including Bayer matrices and texture patterns.
//
// Parameters:
//   - layerName: name of the target layer (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to draw on
//   - x, y: top-left corner of the dither region
//   - width: region width in pixels
//   - height: region height in pixels
//   - color1: first color as hex string (e.g., "#FF0000FF")
//   - color2: second color as hex string (alternates with color1 based on pattern)
//   - pattern: dithering pattern name (see below for available patterns)
//   - density: pattern density/threshold (0.0-1.0, where 0.5 = balanced mix)
//
// Available patterns:
//   - "bayer_2x2", "bayer_4x4", "bayer_8x8" - Ordered dithering matrices
//   - "checkerboard" - Simple alternating pattern
//   - "grass", "water", "stone", "cloud" - Organic texture patterns
//   - "brick" - Masonry pattern
//   - "dots", "diagonal", "cross" - Geometric patterns
//   - "noise" - Pseudo-random pattern
//   - "horizontal_lines", "vertical_lines" - Directional line patterns
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the dither is applied.
//
// Prints "Dithering applied successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
//   - The pattern name is not recognized
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
