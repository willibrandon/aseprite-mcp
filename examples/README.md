# Aseprite MCP Client Examples

This directory contains example clients demonstrating how to use the Aseprite MCP server.

## Example Client

The `client/` directory contains a complete example MCP client that demonstrates:

- Connecting to the Aseprite MCP server via stdio transport
- Creating a 64x64 RGB sprite
- Adding layers and frames
- Drawing animated content (growing circles)
- Filling areas with colors
- Retrieving sprite metadata
- Exporting to GIF and PNG

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
- `../sprites/animated-example.gif` - 4-frame animation with growing colored circles
- `../sprites/frame2-example.png` - Single frame export (frame 2)

## Example Output

```
Aseprite MCP Client Example
===========================

Starting server: D:\SRC\aseprite-mcp-go\bin\aseprite-mcp.exe
Connecting to server...
Connected!

Available tools:
  - create_canvas: Create a new Aseprite sprite
  - add_layer: Add a new layer to the sprite
  - add_frame: Add a new frame to the sprite timeline
  - get_sprite_info: Get metadata about a sprite
  - get_pixels: Read pixel data from a rectangular region
  - draw_pixels: Draw individual pixels on a layer
  - draw_line: Draw a line on a layer
  - draw_rectangle: Draw a rectangle on a layer
  - draw_circle: Draw a circle on a layer
  - fill_area: Fill an area with a color (paint bucket tool)
  - export_sprite: Export sprite to image file

Step 1: Creating 64x64 RGB canvas...
  Created: C:\Users\...\AppData\Local\Temp\aseprite-mcp\sprite-123456.aseprite

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