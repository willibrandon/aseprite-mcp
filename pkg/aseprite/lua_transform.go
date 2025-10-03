package aseprite

import (
	"fmt"
)

// FlipSprite generates a Lua script to flip a sprite, layer, or cel.
//
// Flips the image content either horizontally or vertically. The flip can be
// applied to the entire sprite, a single layer, or just the active cel.
//
// Parameters:
//   - direction: flip direction - "horizontal" (default) or "vertical"
//   - target: scope of the flip - "sprite" (default), "layer", or "cel"
//
// Target modes:
//   - "sprite": flips all layers and frames
//   - "layer": flips only the active layer across all frames
//   - "cel": flips only the active cel (single layer, single frame)
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the flip is complete.
//
// Prints "Sprite flipped [direction] successfully" on success.
// Returns an error if no sprite is active.
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
//
// Rotates the image content by 90, 180, or 270 degrees clockwise. The rotation
// can be applied to the entire sprite, a single layer, or just the active cel.
//
// Parameters:
//   - angle: rotation angle in degrees - must be 90, 180, or 270 (default: 90)
//   - target: scope of the rotation - "sprite" (default), "layer", or "cel"
//
// Target modes:
//   - "sprite": rotates all layers and frames (canvas dimensions change for 90/270)
//   - "layer": rotates only the active layer across all frames
//   - "cel": rotates only the active cel (single layer, single frame)
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the rotation is complete.
//
// Prints "Sprite rotated [angle] degrees successfully" on success.
// Returns an error if no sprite is active.
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
//
// Resizes the entire sprite (all layers and frames) by the specified scale factors
// using one of three scaling algorithms. This changes the canvas dimensions.
//
// Parameters:
//   - scaleX: horizontal scale factor (e.g., 2.0 = double width, 0.5 = half width)
//   - scaleY: vertical scale factor (e.g., 2.0 = double height, 0.5 = half height)
//   - algorithm: scaling algorithm - "nearest", "bilinear", or "rotsprite" (default: "nearest")
//
// Scaling algorithms:
//   - "nearest": nearest neighbor (fast, sharp, best for pixel art)
//   - "bilinear": bilinear interpolation (smooth, best for photos)
//   - "rotsprite": rotation sprite (high-quality pixel art upscaling)
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after scaling is complete.
//
// Prints JSON with new dimensions: {"success":true,"new_width":N,"new_height":N}
// Returns an error if no sprite is active.
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
//
// Trims the sprite canvas to the specified rectangular region, discarding all
// content outside the crop bounds. This affects all layers and frames.
//
// Parameters:
//   - x, y: top-left corner of the crop region (must be non-negative)
//   - width: crop width in pixels (must be positive)
//   - height: crop height in pixels (must be positive)
//
// The crop bounds are validated against the sprite dimensions:
//   - x + width must not exceed sprite.width
//   - y + height must not exceed sprite.height
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after cropping is complete.
//
// Prints "Sprite cropped successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - Crop position is negative
//   - Crop dimensions are not positive
//   - Crop bounds exceed sprite dimensions
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
	-- Select the crop region
	spr.selection = Selection(Rectangle(%d, %d, %d, %d))
	-- Crop to selection
	app.command.CropSprite()
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
//
// Changes the canvas size without scaling the pixel content. Content is anchored
// at the specified position, and empty space is added (or content is cropped) to
// reach the target dimensions.
//
// Parameters:
//   - width: new canvas width in pixels
//   - height: new canvas height in pixels
//   - anchor: content anchor position (see below for options)
//
// Anchor positions:
//   - "center" (default): centers content in new canvas
//   - "top_left": anchors content to top-left corner
//   - "top_right": anchors content to top-right corner
//   - "bottom_left": anchors content to bottom-left corner
//   - "bottom_right": anchors content to bottom-right corner
//
// If the new dimensions are smaller than the current size, content will be cropped
// according to the anchor position.
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the canvas is resized.
//
// Prints "Canvas resized successfully" on success.
// Returns an error if no sprite is active.
func (g *LuaGenerator) ResizeCanvas(width, height int, anchor string) string {
	// Calculate padding amounts based on anchor position
	var leftPad, topPad, rightPad, bottomPad string

	switch anchor {
	case "top_left":
		leftPad = "0"
		topPad = "0"
		rightPad = "newWidth - oldWidth"
		bottomPad = "newHeight - oldHeight"
	case "top_right":
		leftPad = "newWidth - oldWidth"
		topPad = "0"
		rightPad = "0"
		bottomPad = "newHeight - oldHeight"
	case "bottom_left":
		leftPad = "0"
		topPad = "newHeight - oldHeight"
		rightPad = "newWidth - oldWidth"
		bottomPad = "0"
	case "bottom_right":
		leftPad = "newWidth - oldWidth"
		topPad = "newHeight - oldHeight"
		rightPad = "0"
		bottomPad = "0"
	case "center":
		fallthrough
	default:
		leftPad = "math.floor((newWidth - oldWidth) / 2)"
		topPad = "math.floor((newHeight - oldHeight) / 2)"
		rightPad = "math.ceil((newWidth - oldWidth) / 2)"
		bottomPad = "math.ceil((newHeight - oldHeight) / 2)"
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
		right = %s,
		bottom = %s
	}
end)

spr:saveAs(spr.filename)
print("Canvas resized successfully")`, width, height, leftPad, topPad, rightPad, bottomPad)
}

// ApplyOutline generates a Lua script to apply an outline effect to a layer.
//
// Adds a colored outline border around non-transparent pixels in the specified
// cel. This is useful for adding emphasis, creating shadows, or separating sprites
// from backgrounds.
//
// Parameters:
//   - layerName: name of the layer to apply outline to (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to apply outline to
//   - color: outline color in RGBA format
//   - thickness: outline thickness in pixels (1 = single pixel border)
//
// The outline is added outside the existing pixel bounds, expanding the visual
// size of the content by the thickness amount on all sides.
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the outline is applied.
//
// Prints "Outline applied successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
//   - No cel exists at the specified layer/frame
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

// DownsampleImage generates a Lua script to downsample an image using box filter (area averaging).
//
// Reduces image resolution using high-quality box filtering (area averaging).
// This is ideal for converting high-resolution photos to pixel art by preserving
// color information and reducing aliasing artifacts.
//
// Parameters:
//   - sourcePath: absolute path to source image file (automatically escaped for Lua safety)
//   - outputPath: absolute path for output file (automatically escaped for Lua safety)
//   - targetWidth: target width in pixels
//   - targetHeight: target height in pixels
//
// Algorithm details:
//   - Uses box filter (area averaging) for high-quality downsampling
//   - Each target pixel averages all source pixels in its corresponding region
//   - Preserves color mode from source image
//   - Superior to simple nearest-neighbor for photo-to-pixel-art conversion
//
// The source image is opened, processed, and the result is saved to the output
// path. Both sprites are closed after the operation completes.
//
// Prints the output path on success.
// Returns an error if:
//   - Source image cannot be opened
//   - Source sprite has no layers or frames
//   - Source sprite has no cel in first frame
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
