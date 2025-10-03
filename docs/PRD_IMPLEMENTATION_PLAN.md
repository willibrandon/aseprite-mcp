# PRD Implementation Plan

**Status:** Ready to Execute
**Created:** 2025-10-02
**Target Completion:** All missing PRD requirements

---

## Overview

This plan implements all missing requirements from `docs/PRD.md` following existing patterns and conventions. All tests will use real Aseprite instances.

**Total Tasks:** 30
**Missing P0 Requirements:** 3
**Missing P1 Requirements:** 18
**Missing P2 Requirements:** 5
**Documentation Tasks:** 4

---

## Implementation Strategy

### Core Principles
1. **Follow Existing Patterns**: Use established code patterns from current tools
2. **Real Aseprite Testing**: All tests require configured Aseprite executable
3. **Lua Script Generation**: All operations use Aseprite's Lua API via batch mode
4. **Go Idiomatic Docs**: Add comprehensive GoDoc comments to all public APIs
5. **Integration Tests**: Each tool gets integration test with real Aseprite

### File Organization Pattern
```
pkg/tools/
  ├── [category].go          # Tool implementations
  ├── [category]_test.go     # Unit tests (validation, Lua generation)
  └── [category]_integration_test.go  # Integration tests (real Aseprite)
```

---

## Phase 1: Missing P0 Requirements (Critical)

### Task 1.1: Delete Layer (REQ-CANVAS-004)
**File:** `pkg/tools/canvas.go`

**Input Schema:**
```go
type DeleteLayerInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required,description=Path to sprite file"`
    LayerName  string `json:"layer_name" jsonschema:"required,description=Name of layer to delete"`
}
```

**Lua Generator:** `pkg/aseprite/lua.go`
```go
func (g *LuaGenerator) DeleteLayer(layerName string) string
```

**Error Cases:**
- Layer not found
- Attempting to delete last layer
- Sprite file not found

**Integration Test:** Verify layer deletion, verify error on last layer

---

### Task 1.2: Delete Frame (REQ-CANVAS-005)
**File:** `pkg/tools/canvas.go`

**Input Schema:**
```go
type DeleteFrameInput struct {
    SpritePath  string `json:"sprite_path" jsonschema:"required,description=Path to sprite file"`
    FrameNumber int    `json:"frame_number" jsonschema:"required,minimum=1,description=Frame to delete (1-based)"`
}
```

**Lua Generator:** `pkg/aseprite/lua.go`
```go
func (g *LuaGenerator) DeleteFrame(frameNumber int) string
```

**Error Cases:**
- Frame not found
- Attempting to delete last frame
- Invalid frame number

**Integration Test:** Verify frame deletion, verify error on last frame

---

### Task 1.3: Draw Contour (REQ-DRAW-006)
**File:** `pkg/tools/drawing.go`

**Input Schema:**
```go
type DrawContourInput struct {
    SpritePath  string            `json:"sprite_path" jsonschema:"required"`
    LayerName   string            `json:"layer_name" jsonschema:"required"`
    FrameNumber int               `json:"frame_number" jsonschema:"required,minimum=1"`
    Points      []aseprite.Point  `json:"points" jsonschema:"required,minItems=2"`
    Color       string            `json:"color" jsonschema:"required,pattern=^#?[0-9A-Fa-f]{6}([0-9A-Fa-f]{2})?$"`
    Thickness   int               `json:"thickness" jsonschema:"minimum=1,maximum=100,default=1"`
    Closed      bool              `json:"closed" jsonschema:"default=false,description=Connect last point to first"`
}
```

**Lua Generator:** `pkg/aseprite/lua.go`
```go
func (g *LuaGenerator) DrawContour(layerName string, frameNumber int, points []Point, color Color, thickness int, closed bool) string
```

**Implementation:** Use Aseprite's `app.useTool` with tool="line" for each segment

**Integration Test:** Draw open polyline, draw closed polygon, verify thickness

---

## Phase 2: Selection Tools (REQ-SEL-001 through REQ-SEL-008)

**New File:** `pkg/tools/selection.go`
**New File:** `pkg/tools/selection_test.go`
**New File:** `pkg/tools/selection_integration_test.go`

### Task 2.1: Select Rectangle (REQ-SEL-001)
```go
type SelectRectangleInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    X          int    `json:"x" jsonschema:"required"`
    Y          int    `json:"y" jsonschema:"required"`
    Width      int    `json:"width" jsonschema:"required,minimum=1"`
    Height     int    `json:"height" jsonschema:"required,minimum=1"`
    Mode       string `json:"mode" jsonschema:"enum=replace,enum=add,enum=subtract,enum=intersect,default=replace"`
}
```

**Lua:** Use `app.selection = Selection(Rectangle(...))` or modify existing selection

---

### Task 2.2: Select Ellipse (REQ-SEL-002)
```go
type SelectEllipseInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    X          int    `json:"x" jsonschema:"required"`
    Y          int    `json:"y" jsonschema:"required"`
    Width      int    `json:"width" jsonschema:"required,minimum=1"`
    Height     int    `json:"height" jsonschema:"required,minimum=1"`
    Mode       string `json:"mode" jsonschema:"enum=replace,enum=add,enum=subtract,enum=intersect,default=replace"`
}
```

**Lua:** Use Aseprite's ellipse selection with mode

---

### Task 2.3: Select All (REQ-SEL-003)
```go
type SelectAllInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
}
```

**Lua:** `app.command.SelectAll()`

---

### Task 2.4: Deselect (REQ-SEL-004)
```go
type DeselectInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
}
```

**Lua:** `app.command.DeselectMask()`

---

### Task 2.5: Move Selection (REQ-SEL-005)
```go
type MoveSelectionInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    DX         int    `json:"dx" jsonschema:"required,description=Horizontal offset in pixels"`
    DY         int    `json:"dy" jsonschema:"required,description=Vertical offset in pixels"`
}
```

**Lua:** Modify selection bounds programmatically

---

### Task 2.6: Cut Selection (REQ-SEL-006)
```go
type CutSelectionInput struct {
    SpritePath  string `json:"sprite_path" jsonschema:"required"`
    LayerName   string `json:"layer_name" jsonschema:"required"`
    FrameNumber int    `json:"frame_number" jsonschema:"required,minimum=1"`
}
```

**Lua:** `app.command.Cut()`

---

### Task 2.7: Copy Selection (REQ-SEL-007)
```go
type CopySelectionInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
}
```

**Lua:** `app.command.Copy()`

---

### Task 2.8: Paste Clipboard (REQ-SEL-008)
```go
type PasteClipboardInput struct {
    SpritePath  string `json:"sprite_path" jsonschema:"required"`
    LayerName   string `json:"layer_name" jsonschema:"required"`
    FrameNumber int    `json:"frame_number" jsonschema:"required,minimum=1"`
    X           *int   `json:"x" jsonschema:"description=Paste X position (optional)"`
    Y           *int   `json:"y" jsonschema:"description=Paste Y position (optional)"`
}
```

**Lua:** `app.command.Paste()`, set position if provided

---

## Phase 3: Palette Tools Completion

**File:** `pkg/tools/palette_tools.go` (extend existing)

### Task 3.1: Get Palette (REQ-PAL-001)
```go
type GetPaletteInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
}

type GetPaletteOutput struct {
    Colors []string `json:"colors" jsonschema:"description=Array of hex colors"`
    Size   int      `json:"size" jsonschema:"description=Number of colors"`
}
```

**Lua:** Iterate `app.activeSprite.palettes[1]`, output JSON array

---

### Task 3.2: Set Palette Color (REQ-PAL-003)
```go
type SetPaletteColorInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    Index      int    `json:"index" jsonschema:"required,minimum=0,maximum=255"`
    Color      string `json:"color" jsonschema:"required,pattern=^#?[0-9A-Fa-f]{6}([0-9A-Fa-f]{2})?$"`
}
```

**Lua:** `palette:setColor(index, Color(...))`

---

### Task 3.3: Add Palette Color (REQ-PAL-004)
```go
type AddPaletteColorInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    Color      string `json:"color" jsonschema:"required,pattern=^#?[0-9A-Fa-f]{6}([0-9A-Fa-f]{2})?$"`
}

type AddPaletteColorOutput struct {
    ColorIndex int `json:"color_index" jsonschema:"description=Index of added color"`
}
```

**Lua:** `palette:resize(palette:size() + 1)`, set color at new index

---

### Task 3.4: Sort Palette (REQ-PAL-005)
```go
type SortPaletteInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    Method     string `json:"method" jsonschema:"enum=hue,enum=saturation,enum=brightness,enum=luminance,required"`
    Ascending  bool   `json:"ascending" jsonschema:"default=true"`
}
```

**Lua:** Extract colors, sort in Lua by method, rebuild palette

---

## Phase 4: Transform & Filter Tools

**File:** `pkg/tools/transform.go` (extend existing)
**File:** `pkg/tools/transform_integration_test.go` (extend existing)

### Task 4.1: Flip Sprite (REQ-XFORM-001)
```go
type FlipSpriteInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    Direction  string `json:"direction" jsonschema:"enum=horizontal,enum=vertical,required"`
    Target     string `json:"target" jsonschema:"enum=sprite,enum=layer,enum=cel,default=sprite"`
}
```

**Lua:** `app.command.Flip({ direction = "horizontal" })` with proper target

---

### Task 4.2: Rotate Sprite (REQ-XFORM-002)
```go
type RotateSpriteInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    Angle      string `json:"angle" jsonschema:"enum=90,enum=180,enum=270,required"`
    Target     string `json:"target" jsonschema:"enum=sprite,enum=layer,enum=cel,default=sprite"`
}
```

**Lua:** `app.command.Rotate({ angle = 90 })`

---

### Task 4.3: Scale Sprite (REQ-XFORM-003)
```go
type ScaleSpriteInput struct {
    SpritePath string  `json:"sprite_path" jsonschema:"required"`
    ScaleX     float64 `json:"scale_x" jsonschema:"required,minimum=0.01,maximum=100"`
    ScaleY     float64 `json:"scale_y" jsonschema:"required,minimum=0.01,maximum=100"`
    Algorithm  string  `json:"algorithm" jsonschema:"enum=nearest,enum=bilinear,enum=rotsprite,default=nearest"`
}

type ScaleSpriteOutput struct {
    Success   bool `json:"success"`
    NewWidth  int  `json:"new_width"`
    NewHeight int  `json:"new_height"`
}
```

**Lua:** `app.command.SpriteSize({ width = ..., height = ... })`

---

### Task 4.4: Crop Sprite (REQ-XFORM-004)
```go
type CropSpriteInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    X          int    `json:"x" jsonschema:"required"`
    Y          int    `json:"y" jsonschema:"required"`
    Width      int    `json:"width" jsonschema:"required,minimum=1"`
    Height     int    `json:"height" jsonschema:"required,minimum=1"`
}
```

**Lua:** `app.command.CropSprite({ bounds = Rectangle(...) })`

---

### Task 4.5: Resize Canvas (REQ-XFORM-005)
```go
type ResizeCanvasInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    Width      int    `json:"width" jsonschema:"required,minimum=1,maximum=65535"`
    Height     int    `json:"height" jsonschema:"required,minimum=1,maximum=65535"`
    Anchor     string `json:"anchor" jsonschema:"enum=center,enum=top_left,enum=top_right,enum=bottom_left,enum=bottom_right,default=center"`
}
```

**Lua:** `app.command.CanvasSize({ width = ..., height = ..., left = ..., top = ... })`

---

### Task 4.6: Apply Outline (REQ-XFORM-006)
```go
type ApplyOutlineInput struct {
    SpritePath  string `json:"sprite_path" jsonschema:"required"`
    LayerName   string `json:"layer_name" jsonschema:"required"`
    FrameNumber int    `json:"frame_number" jsonschema:"required,minimum=1"`
    Color       string `json:"color" jsonschema:"required,pattern=^#?[0-9A-Fa-f]{6}([0-9A-Fa-f]{2})?$"`
    Thickness   int    `json:"thickness" jsonschema:"required,minimum=1,maximum=10"`
}
```

**Lua:** `app.command.Outline({ color = Color(...), size = thickness })`

---

## Phase 5: Advanced Export Tools

**File:** `pkg/tools/export.go` (extend existing)

### Task 5.1: Export Spritesheet (REQ-EXPORT-002)
```go
type ExportSpritesheetInput struct {
    SpritePath  string `json:"sprite_path" jsonschema:"required"`
    OutputPath  string `json:"output_path" jsonschema:"required"`
    Layout      string `json:"layout" jsonschema:"enum=horizontal,enum=vertical,enum=rows,enum=columns,enum=packed,default=horizontal"`
    Padding     int    `json:"padding" jsonschema:"minimum=0,maximum=100,default=0"`
    IncludeJSON bool   `json:"include_json" jsonschema:"default=false"`
}

type ExportSpritesheetOutput struct {
    SpritesheetPath string  `json:"spritesheet_path"`
    MetadataPath    *string `json:"metadata_path,omitempty"`
    FrameCount      int     `json:"frame_count"`
}
```

**Lua:** `app.command.ExportSpriteSheet({ type = ..., textureFilename = ..., dataFilename = ... })`

---

### Task 5.2: Import Image as Layer (REQ-EXPORT-003)
```go
type ImportImageInput struct {
    SpritePath  string             `json:"sprite_path" jsonschema:"required"`
    ImagePath   string             `json:"image_path" jsonschema:"required"`
    LayerName   string             `json:"layer_name" jsonschema:"required"`
    FrameNumber int                `json:"frame_number" jsonschema:"required,minimum=1"`
    Position    *aseprite.Point    `json:"position,omitempty" jsonschema:"description=Placement position (optional)"`
}
```

**Lua:** Load image, create cel, draw image at position

---

### Task 5.3: Save As (REQ-EXPORT-004)
```go
type SaveAsInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    OutputPath string `json:"output_path" jsonschema:"required,description=New .aseprite file path"`
}
```

**Lua:** `app.activeSprite:saveAs(outputPath)`

---

### Task 5.4: Delete Tag (REQ-ANIM-005)
```go
type DeleteTagInput struct {
    SpritePath string `json:"sprite_path" jsonschema:"required"`
    TagName    string `json:"tag_name" jsonschema:"required"`
}
```

**File:** `pkg/tools/animation.go` (extend existing)

**Lua:** Find tag by name, `app.activeSprite:deleteTag(tag)`

---

## Phase 6: GoDoc Documentation

### Task 6.1: Document pkg/aseprite
**Files to document:**
- `client.go`: Client, ExecuteLua, ExecuteCommand, GetVersion
- `lua.go`: LuaGenerator, all generator methods, EscapeString
- `types.go`: Color, Point, Rectangle, Pixel, SpriteInfo, ColorMode
- `palette.go`: Palette extraction functions
- `image_analysis.go`: Analysis functions

**Pattern:**
```go
// Client executes Aseprite commands and Lua scripts in batch mode.
// It provides a high-level interface for sprite manipulation through
// Aseprite's command-line interface and Lua scripting API.
//
// All operations execute with a configurable timeout and automatic
// temporary file cleanup. The client is safe for concurrent use from
// multiple goroutines.
type Client struct {
    // ...
}

// ExecuteLua executes a Lua script in Aseprite batch mode.
// If spritePath is non-empty, the sprite will be opened before running the script.
//
// The script is written to a temporary file with restricted permissions (0600)
// and automatically cleaned up after execution. Script execution respects the
// configured timeout.
//
// Returns the stdout output from Aseprite, or an error if execution fails.
func (c *Client) ExecuteLua(ctx context.Context, script string, spritePath string) (string, error)
```

---

### Task 6.2: Document pkg/tools
**Files to document:**
- All tool input/output structs
- All Register* functions
- Pattern: Describe purpose, parameters, errors, example usage

---

### Task 6.3: Document pkg/config
**Document:**
- Config struct and all fields
- Load, Validate, SetDefaults methods
- Configuration file format and location

---

### Task 6.4: Document pkg/server
**Document:**
- Server struct
- New, Run methods
- Integration with MCP protocol

---

## Phase 7: Documentation Updates

### Task 7.1: Update README.md

Add all new tools to the "Available Tools" section with descriptions:

#### Canvas & Layer Management (add to existing section)
```markdown
| `delete_layer` | Delete a layer from the sprite (cannot delete last layer) |
| `delete_frame` | Delete a frame from the sprite (cannot delete last frame) |
```

#### Drawing & Painting (add to existing section)
```markdown
| `draw_contour` | Draw a multi-point polyline or polygon with configurable thickness |
```

#### New Section: Selection Tools
```markdown
### Selection Tools
| Tool | Description |
|------|-------------|
| `select_rectangle` | Select a rectangular region with mode (replace/add/subtract/intersect) |
| `select_ellipse` | Select an elliptical region with mode (replace/add/subtract/intersect) |
| `select_all` | Select the entire canvas |
| `deselect` | Clear the current selection |
| `move_selection` | Move the current selection by offset |
| `cut_selection` | Cut selected pixels to clipboard |
| `copy_selection` | Copy selected pixels to clipboard |
| `paste_clipboard` | Paste clipboard content at specified position |
```

#### Palette Tools (expand existing section)
```markdown
| `get_palette` | Retrieve the current sprite palette as array of hex colors |
| `set_palette` | Set sprite's color palette to specified colors (supports 1-256 colors) |
| `set_palette_color` | Set a specific palette index to a color |
| `add_palette_color` | Add a new color to the palette |
| `sort_palette` | Sort palette by hue, saturation, brightness, or luminance |
```

#### New Section: Transform & Filter Tools
```markdown
### Transform & Filter Tools
| Tool | Description |
|------|-------------|
| `flip_sprite` | Flip sprite/layer/cel horizontally or vertically |
| `rotate_sprite` | Rotate sprite/layer/cel by 90, 180, or 270 degrees |
| `scale_sprite` | Scale sprite with nearest/bilinear/rotsprite algorithm |
| `crop_sprite` | Crop sprite to specified rectangle |
| `resize_canvas` | Resize canvas without scaling content (with anchor positioning) |
| `apply_outline` | Apply outline effect with configurable color and thickness |
```

#### Export & Import (expand existing section)
```markdown
### Inspection & Export
| Tool | Description |
|------|-------------|
| `get_pixels` | Read pixel data from a rectangular region (paginated, for verification) |
| `export_sprite` | Export sprite to PNG/GIF/JPG/BMP |
| `export_spritesheet` | Export animation frames as spritesheet with layout options |
| `import_image` | Import an image file as a new layer in the sprite |
| `save_as` | Save sprite to a new .aseprite file path |
```

#### Animation (add to existing section)
```markdown
| `delete_tag` | Delete an animation tag by name |
```

---

### Task 7.2: Update examples/README.md

Update the feature list to include all new capabilities:

```markdown
The `client/` directory contains a complete example MCP client that demonstrates:

- Connecting to the Aseprite MCP server via stdio transport
- Creating a 64x64 RGB sprite
- Adding layers and frames
- **Layer & Frame Management:** Deleting layers and frames
- Drawing animated content (growing circles)
- **Drawing Tools:** Lines, rectangles, circles, polylines/contours
- Filling areas with colors
- **Selection Tools:** Rectangle/ellipse selection, cut/copy/paste operations
- Reading pixels for verification (using get_pixels with pagination)
- Applying dithering patterns for professional gradients
- Analyzing palette harmonies (complementary, triadic, temperature)
- **Palette Management:** Getting, setting, sorting palettes
- Setting custom limited palettes
- Applying palette-constrained shading with light direction
- Palette-aware drawing with automatic color snapping
- **Transform Operations:** Flip, rotate, scale, crop sprites
- **Advanced Export:** Spritesheets, multi-format export, save as
- Detecting and smoothing jagged edges with antialiasing
- Retrieving sprite metadata
- Exporting to GIF and PNG
```

---

### Task 7.3: Update examples/client/main.go

Add example usage for key new tools in the `createAnimatedSprite` function:

```go
// After existing steps, add:

// Step 17: Demonstrate selection and transformation
logger.Information("")
logger.Information("Step 17: Demonstrating selection and transform tools...")

// Select a region
if _, err := callTool(ctx, session, "select_rectangle", map[string]any{
    "sprite_path": spritePath,
    "x":           10,
    "y":           10,
    "width":       20,
    "height":      20,
    "mode":        "replace",
}); err != nil {
    return fmt.Errorf("select_rectangle failed: %w", err)
}
logger.Information("  Selected 20x20 region")

// Flip the sprite horizontally
if _, err := callTool(ctx, session, "flip_sprite", map[string]any{
    "sprite_path": spritePath,
    "direction":   "horizontal",
    "target":      "sprite",
}); err != nil {
    return fmt.Errorf("flip_sprite failed: %w", err)
}
logger.Information("  Flipped sprite horizontally")

// Step 18: Demonstrate palette management
logger.Information("")
logger.Information("Step 18: Getting and sorting palette...")

paletteResp, err := callTool(ctx, session, "get_palette", map[string]any{
    "sprite_path": paletteSprite,
})
if err != nil {
    return fmt.Errorf("get_palette failed: %w", err)
}
var paletteData struct {
    Colors []string `json:"colors"`
    Size   int      `json:"size"`
}
if err := json.Unmarshal([]byte(paletteResp), &paletteData); err != nil {
    return fmt.Errorf("failed to parse palette: %w", err)
}
logger.Information("  Retrieved {Count} colors from palette", paletteData.Size)

// Sort palette by hue
if _, err := callTool(ctx, session, "sort_palette", map[string]any{
    "sprite_path": paletteSprite,
    "method":      "hue",
    "ascending":   true,
}); err != nil {
    return fmt.Errorf("sort_palette failed: %w", err)
}
logger.Information("  Sorted palette by hue")

// Step 19: Export spritesheet
logger.Information("")
logger.Information("Step 19: Exporting animation as spritesheet...")

sheetPath := filepath.Join(outputDir, "spritesheet.png")
if _, err := callTool(ctx, session, "export_spritesheet", map[string]any{
    "sprite_path":  spritePath,
    "output_path":  sheetPath,
    "layout":       "horizontal",
    "padding":      2,
    "include_json": true,
}); err != nil {
    return fmt.Errorf("export_spritesheet failed: %w", err)
}
logger.Information("  Exported spritesheet: {SheetPath}", sheetPath)
```

Expected output additions:
```
Step 17: Demonstrating selection and transform tools...
  Selected 20x20 region
  Flipped sprite horizontally

Step 18: Getting and sorting palette...
  Retrieved 8 colors from palette
  Sorted palette by hue

Step 19: Exporting animation as spritesheet...
  Exported spritesheet: ../sprites/spritesheet.png
```

Output files to add:
- `../sprites/spritesheet.png` - Horizontal spritesheet layout of all animation frames
- `../sprites/spritesheet.json` - Metadata for the spritesheet (if include_json=true)

---

### Task 7.4: Add Tool Usage Examples to README.md

Add a new "Tool Examples" section before "Development":

```markdown
## Tool Usage Examples

### Canvas Management
```bash
# Delete a layer (keeping at least one)
delete_layer --sprite=sprite.aseprite --layer="Background"

# Delete a frame (keeping at least one)
delete_frame --sprite=sprite.aseprite --frame=2
```

### Drawing
```bash
# Draw a polygon/polyline
draw_contour --sprite=sprite.aseprite --layer="Layer 1" --frame=1 \
  --points='[{"x":10,"y":10},{"x":20,"y":5},{"x":30,"y":10}]' \
  --color="#FF0000" --thickness=2 --closed=true
```

### Selection Tools
```bash
# Select rectangular region
select_rectangle --sprite=sprite.aseprite --x=10 --y=10 --width=20 --height=20 --mode=replace

# Copy and paste selection
copy_selection --sprite=sprite.aseprite
paste_clipboard --sprite=sprite.aseprite --layer="Layer 1" --frame=1 --x=50 --y=50
```

### Transform Operations
```bash
# Flip sprite horizontally
flip_sprite --sprite=sprite.aseprite --direction=horizontal --target=sprite

# Rotate 90 degrees
rotate_sprite --sprite=sprite.aseprite --angle=90 --target=sprite

# Scale sprite 2x with rotsprite algorithm
scale_sprite --sprite=sprite.aseprite --scale_x=2.0 --scale_y=2.0 --algorithm=rotsprite
```

### Palette Management
```bash
# Get current palette
get_palette --sprite=sprite.aseprite

# Set specific palette color
set_palette_color --sprite=sprite.aseprite --index=0 --color="#FF0000"

# Sort palette by hue
sort_palette --sprite=sprite.aseprite --method=hue --ascending=true
```

### Advanced Export
```bash
# Export spritesheet with metadata
export_spritesheet --sprite=sprite.aseprite --output=sheet.png \
  --layout=horizontal --padding=2 --include_json=true

# Import image as layer
import_image --sprite=sprite.aseprite --image=reference.png \
  --layer="Reference" --frame=1 --position='{"x":0,"y":0}'
```
```

---

## Testing Strategy

### Integration Test Pattern
```go
//go:build integration
// +build integration

package tools

import (
    "context"
    "testing"
    "github.com/willibrandon/aseprite-mcp-go/internal/testutil"
)

func TestIntegration_DeleteLayer(t *testing.T) {
    cfg := testutil.LoadTestConfig(t)
    client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
    gen := aseprite.NewLuaGenerator()

    // Create test sprite with multiple layers
    // ...

    // Test successful deletion
    // ...

    // Test error: delete last layer
    // ...
}
```

### Test Coverage Requirements
- Unit tests: 100% of validation logic
- Integration tests: All tools with real Aseprite
- Error path tests: All error cases from PRD
- Performance tests: Large sprites, bulk operations

---

## Execution Order

### Week 1: Critical P0 + Selection Tools
1. Delete Layer & Frame (P0) - 1 day
2. Draw Contour (P0) - 1 day
3. Selection Tools (8 tools) - 3 days

### Week 2: Palette + Transform Tools
4. Palette Tools (4 tools) - 2 days
5. Transform Tools (6 tools) - 3 days

### Week 3: Export + Documentation
6. Export Tools (3 tools) - 1 day
7. Delete Tag - 0.5 day
8. GoDoc all packages - 2 days
9. README updates - 0.5 day

### Week 4: Testing & Polish
10. Integration test coverage - 2 days
11. Performance benchmarks - 1 day
12. Final review & fixes - 2 days

---

## Success Criteria

- ✅ All 30 PRD requirements implemented
- ✅ All tools have integration tests with real Aseprite
- ✅ All public APIs have GoDoc comments
- ✅ README documents all tools
- ✅ 90%+ test coverage maintained
- ✅ All tests pass: `go test -tags=integration ./...`
- ✅ All lints pass: `make lint`
- ✅ Performance benchmarks meet targets

---

## Notes

- Use existing patterns from `pkg/tools/*.go` files
- Follow Lua generation patterns from `pkg/aseprite/lua.go`
- All temporary files use restricted permissions (0600)
- All operations wrapped in `app.transaction()` for atomicity
- Error messages must be actionable and clear
- Sprite files saved after mutations: `spr:saveAs(spr.filename)`
