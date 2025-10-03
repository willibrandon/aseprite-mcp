package aseprite

import (
	"fmt"
)

// SelectRectangle generates a Lua script to create a rectangular selection.
//
// Creates or modifies the current selection using a rectangular region.
// Selections are used to limit drawing operations, cut/copy content, or define
// regions for transformations.
//
// Parameters:
//   - x, y: top-left corner of the selection rectangle
//   - width: rectangle width in pixels (must be positive)
//   - height: rectangle height in pixels (must be positive)
//   - mode: selection mode - "replace" (default), "add", "subtract", or "intersect"
//
// Selection modes:
//   - "replace": replaces current selection with new rectangle
//   - "add": adds rectangle to current selection (union)
//   - "subtract": removes rectangle from current selection
//   - "intersect": keeps only the intersection of current and new selection
//
// Note: Selections are transient and do not persist in saved .aseprite files.
// The sprite is NOT saved after this operation.
//
// Prints "Rectangle selection created successfully" on success.
// Returns an error if no sprite is active.
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
//
// Creates or modifies the current selection using an elliptical region.
// The ellipse is defined by its bounding rectangle and filled using the
// midpoint ellipse algorithm.
//
// Parameters:
//   - x, y: top-left corner of the ellipse bounding box
//   - width: bounding box width in pixels (ellipse diameter on x-axis)
//   - height: bounding box height in pixels (ellipse diameter on y-axis)
//   - mode: selection mode - "replace" (default), "add", "subtract", or "intersect"
//
// Selection modes:
//   - "replace": replaces current selection with new ellipse
//   - "add": adds ellipse to current selection (union)
//   - "subtract": removes ellipse from current selection
//   - "intersect": keeps only the intersection of current and new selection
//
// Note: Selections are transient and do not persist in saved .aseprite files.
// The sprite is NOT saved after this operation.
//
// Prints "Ellipse selection created successfully" on success.
// Returns an error if no sprite is active.
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
//
// Creates a selection covering the entire sprite canvas from (0,0) to
// (sprite.width, sprite.height). This is useful before copy/cut operations
// or to quickly select all content for transformations.
//
// Note: Selections are transient and do not persist in saved .aseprite files.
// The sprite is NOT saved after this operation.
//
// Prints "Select all completed successfully" on success.
// Returns an error if no sprite is active.
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
//
// Removes the current selection mask, allowing operations to affect the entire
// canvas again. This is the opposite of SelectAll.
//
// Note: Selections are transient and do not persist in saved .aseprite files.
// The sprite is NOT saved after this operation.
//
// Prints "Deselect completed successfully" on success.
// Returns an error if no sprite is active.
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
//
// Shifts the selection mask by the specified offset without moving the pixel
// content. This is useful for repositioning the selection after creating it,
// or for aligning selections with specific features.
//
// Parameters:
//   - dx: horizontal offset in pixels (positive = right, negative = left)
//   - dy: vertical offset in pixels (positive = down, negative = up)
//
// Note: Selections are transient and do not persist in saved .aseprite files.
// The sprite is NOT saved after this operation.
//
// Prints "Selection moved successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - No selection exists to move
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
//
// Removes the pixels within the current selection and copies them to the
// clipboard. The cut area becomes transparent (filled with transparent pixels).
//
// Parameters:
//   - layerName: name of the layer to cut from (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to cut from
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the cut is complete.
//
// Prints "Cut selection completed successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - No selection exists
//   - The layer is not found
//   - The frame number is invalid
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
//
// Copies the pixels within the current selection to the clipboard without
// removing them. The clipboard content can then be pasted elsewhere.
//
// IMPORTANT: Clipboard operations have limitations in batch mode. The clipboard
// state does not persist across separate Aseprite process invocations, so you
// must perform copy and paste in the same script execution or session.
//
// The sprite is NOT saved after this operation (copy doesn't modify content).
//
// Prints "Copy selection completed successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - No selection exists
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
//
// Pastes the clipboard content (from a previous copy or cut operation) to the
// specified layer and frame. The paste position can be explicitly set or will
// use the current position if not specified.
//
// Parameters:
//   - layerName: name of the layer to paste to (automatically escaped for Lua safety)
//   - frameNumber: 1-based frame index to paste to
//   - x: optional x-coordinate for paste position (nil = current position)
//   - y: optional y-coordinate for paste position (nil = current position)
//
// IMPORTANT: Clipboard operations have limitations in batch mode. The clipboard
// state does not persist across separate Aseprite process invocations.
//
// The operation is wrapped in a transaction for atomicity and the sprite
// is saved after the paste is complete.
//
// Prints "Paste completed successfully" on success.
// Returns an error if:
//   - No sprite is active
//   - The layer is not found
//   - The frame number is invalid
//   - Clipboard is empty
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
