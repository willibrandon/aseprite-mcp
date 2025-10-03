# Aseprite MCP Client Examples

This directory contains example clients demonstrating how to use the Aseprite MCP server.

## Example Client

The `client/` directory contains a complete example MCP client that demonstrates:

- Connecting to the Aseprite MCP server via stdio transport
- Creating a 64x64 RGB sprite
- Adding and deleting layers and frames
- Drawing animated content (growing circles)
- Drawing polylines and polygons (zigzag, triangle, star)
- Filling areas with colors
- Reading pixels for verification (using get_pixels with pagination)
- Applying dithering patterns for professional gradients
- Analyzing palette harmonies (complementary, triadic, temperature)
- Setting custom limited palettes
- Applying palette-constrained shading with light direction
- Palette-aware drawing with automatic color snapping
- Detecting and smoothing jagged edges with antialiasing
- Drawing operations (demonstrating rectangle and circle tools as alternatives to copy/paste)
- Retrieving sprite metadata
- Exporting to GIF and PNG

**Note**: Selection tools (`select_rectangle`, `select_ellipse`, etc.) are demonstrated in integration tests. Selections are transient and don't persist across tool calls, so copy/paste workflows require combining operations in single Lua scripts.

## Running the Example

### Prerequisites

1. Build the Aseprite MCP server:
   ```bash
   cd ../..
   make build
   ```

2. Ensure you have Aseprite configured at `~/.config/aseprite-mcp/config.json`

### Run the Example

```bash
# From the examples/client directory
cd examples/client

# Run with server in ../../bin/
go run main.go

# Or set custom server path
ASEPRITE_MCP_PATH=/path/to/aseprite-mcp go run main.go
```

### Output

The example creates:
- `../sprites/animated-example.gif` - 4-frame animation with growing colored circles on blue background
- `../sprites/frame2-example.png` - Single frame export (frame 2, green circle)
- `../sprites/dithered-gradient.png` - Demonstration of Bayer 4x4 dithering pattern
- `../sprites/shaded-sphere.png` - 64x64 sphere with palette-constrained smooth shading from light to dark
- `../sprites/palette-drawing-comparison.png` - Side-by-side comparison: pastel colors (left) vs palette-snapped pure colors (right)
- `../sprites/antialiasing-before.png` - Jagged diagonal line (stair-step pattern)
- `../sprites/antialiasing-after.png` - Smoothed diagonal with intermediate colors applied
- `/tmp/selection-demo.png` - Drawing demo showing red squares and blue circle

## Example Output

```
Aseprite MCP Client Example
===========================

Starting server: ../../bin/aseprite-mcp
Connecting to server...
Connected!

Available tools:
  - create_canvas: Create a new Aseprite sprite
  - add_layer: Add a new layer to the sprite
  - delete_layer: Delete a layer from the sprite
  - add_frame: Add a new frame to the sprite timeline
  - delete_frame: Delete a frame from the sprite
  - get_sprite_info: Get metadata about a sprite
  - get_pixels: Read pixel data from a rectangular region (with pagination)
  - draw_pixels: Draw individual pixels on a layer
  - draw_line: Draw a line on a layer
  - draw_contour: Draw polylines and polygons by connecting points
  - draw_rectangle: Draw a rectangle on a layer
  - draw_circle: Draw a circle on a layer
  - fill_area: Fill an area with a color (paint bucket tool)
  - draw_with_dither: Fill region with dithering patterns for gradients/textures
  - analyze_reference: Extract palette, edges, and composition from reference images
  - downsample_image: Convert high-res images to pixel art dimensions
  - set_palette: Set sprite's color palette
  - apply_shading: Apply palette-constrained shading with light direction
  - analyze_palette_harmonies: Analyze color relationships and temperature
  - select_rectangle: Create rectangular selection with mode (replace/add/subtract/intersect)
  - select_ellipse: Create elliptical selection with mode
  - select_all: Select entire canvas
  - deselect: Clear selection
  - move_selection: Move selection bounds by offset
  - cut_selection: Cut selected pixels to clipboard
  - copy_selection: Copy selected pixels to clipboard
  - paste_clipboard: Paste clipboard content at position
  - export_sprite: Export sprite to image file

Step 1: Creating 64x64 RGB canvas...
  Created: /tmp/aseprite-mcp/sprite-123456.aseprite

Step 2: Adding 'Background' layer...
  Layer added

Step 3: Filling background with blue...
  Background filled

Step 4: Adding 3 animation frames...
  Frames added (4 total)

Step 5: Drawing animated circles...
  Frame 1: radius 8, color #FF0000
  Frame 2: radius 11, color #00FF00
  Frame 3: radius 14, color #FFFF00
  Frame 4: radius 17, color #FF00FF

Step 6: Reading pixels from frame 2 to verify drawing...
  Read 100 pixels from center region
  Found 78 green pixels in the region

Step 7: Getting sprite metadata...
  Info: {"width":64,"height":64,"color_mode":"RGB","frame_count":4,"layer_count":2,"layers":["Background","Layer 1"]}

Step 8: Exporting as GIF...
  Exported: ../sprites/animated-example.gif

Step 9: Exporting frame 2 as PNG...
  Exported: ../sprites/frame2-example.png

Step 10: Creating sprite with dithered gradient...
  Created: /tmp/aseprite-mcp/sprite-789012.aseprite

Step 11: Applying Bayer 4x4 dithering pattern...
  Dithering applied successfully

Step 12: Exporting dithered gradient...
  Exported: ../sprites/dithered-gradient.png

Step 13: Analyzing palette harmonies from our colors...
  Dominant temperature: warm
  Found 2 complementary pairs

Step 14: Creating sprite with limited palette...
  Created: /tmp/aseprite-mcp/sprite-456789.aseprite
  Palette set successfully (8 colors)

Step 15: Drawing shape with palette-constrained shading...
  Shading applied successfully
  Exported: ../sprites/shaded-sphere.png

Step 16: Demonstrating palette-aware drawing...
  Created: /tmp/aseprite-mcp/sprite-...
  Palette-aware drawing completed (left: pastel colors, right: snapped to pure colors)
  Exported: ../sprites/palette-drawing-comparison.png

Step 17: Demonstrating antialiasing suggestions...
  Created: /tmp/aseprite-mcp/sprite-...
  Found 4 jagged edge positions
    - Suggestion 1: pos=(24,11) direction=diagonal_nw
    - Suggestion 2: pos=(25,12) direction=diagonal_nw
    - Suggestion 3: pos=(26,13) direction=diagonal_nw
  Antialiasing applied: jagged diagonal smoothed
  Exported before: ../sprites/antialiasing-before.png
  Exported after: ../sprites/antialiasing-after.png

Step 17: Demonstrating layer and frame deletion...
  Created: /tmp/aseprite-mcp/sprite-...
  Added 2 extra layers (3 total)
  Added 2 extra frames (3 total)
  Deleted Layer 2 (2 layers remaining)
  Deleted frame 2 (2 frames remaining)
  Final state: {"width":32,"height":32,"color_mode":"RGB","frame_count":2,"layer_count":2,"layers":["Layer 1","Layer 3"]}

Step 18: Demonstrating polylines and polygons...
  Drawing zigzag polyline on frame 1...
  Drawing triangle on frame 2...
  Drawing star on frame 3 with palette snapping...
  ✓ Drew polylines and polygons successfully

Step 19: Demonstrating selection and clipboard operations...
  Creating sprite for selection demo...
  Drawing red square (20x20 at 20,20)...
  Copying red square to position (60, 60) using draw_rectangle...
  Drawing blue circle (radius 15 at 30,80)...
  Exporting selection demo to: /tmp/selection-demo.png
  ✓ Drawing operations completed successfully
  ✓ Result saved to: /tmp/selection-demo.png
  Note: Selection tools work within single Lua scripts but don't persist across tool calls

Example completed successfully!
```

## Building Your Own Client

To create your own MCP client for Aseprite:

1. Import the MCP SDK:
   ```go
   import "github.com/modelcontextprotocol/go-sdk/mcp"
   ```

2. Start the server as a subprocess:
   ```go
   cmd := exec.Command("/path/to/aseprite-mcp")
   clientTransport, serverTransport := mcp.NewStdioServerTransport(cmd)
   ```

3. Create and connect the client:
   ```go
   client := mcp.NewClient(&mcp.Implementation{
       Name:    "my-client",
       Version: "1.0.0",
   })
   client.Connect(ctx, clientTransport)
   serverTransport.Start()
   ```

4. Call tools:
   ```go
   args, _ := json.Marshal(map[string]interface{}{
       "width": 64,
       "height": 64,
       "color_mode": "rgb",
   })
   resp, _ := client.CallTool(ctx, &mcp.CallToolRequest{
       Name: "create_canvas",
       Arguments: args,
   })
   ```

See `client/main.go` for a complete working example.