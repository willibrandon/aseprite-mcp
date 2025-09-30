# Aseprite MCP Server Design Document

## Overview

An MCP (Model Context Protocol) server implementation in Go that provides programmatic access to Aseprite's pixel art and animation capabilities. The server exposes Aseprite's Lua scripting API through MCP tools, enabling AI assistants to create and manipulate sprites, animations, and pixel art.

## Architecture

### Technology Stack

- **Language**: Go 1.23+
- **MCP SDK**: Official `github.com/modelcontextprotocol/go-sdk`
- **Aseprite Integration**: Command-line invocation with Lua scripting via `--batch` and `--script` flags
- **Transport**: Stdio (primary), with optional HTTP/SSE support for web integrations

### Core Components

```
aseprite-mcp/
├── cmd/
│   └── aseprite-mcp/
│       └── main.go              # Entry point, server initialization
├── pkg/
│   ├── aseprite/
│   │   ├── client.go            # Aseprite command execution
│   │   ├── lua.go               # Lua script generation helpers
│   │   └── types.go             # Domain types (Color, Point, Rectangle, etc.)
│   ├── tools/
│   │   ├── canvas.go            # Canvas/sprite management tools
│   │   ├── drawing.go           # Drawing primitives tools
│   │   ├── animation.go         # Animation/frame/cel tools
│   │   ├── selection.go         # Selection manipulation tools
│   │   ├── palette.go           # Palette management tools
│   │   ├── transform.go         # Transform/filter tools
│   │   └── export.go            # Export/import tools
│   └── config/
│       └── config.go            # Configuration management
├── scripts/
│   └── templates/               # Lua script templates (optional)
├── examples/
│   ├── client.go                # Example MCP client usage
│   └── sprites/                 # Example output sprites
├── go.mod
├── go.sum
├── README.md
└── DESIGN.md                    # This document
```

## MCP Server Implementation

### Server Initialization

```go
type Config struct {
    AsepritePath string // Path to aseprite executable
    TempDir      string // Directory for temporary files
    Timeout      time.Duration
}

func NewServer(cfg *Config) *mcp.Server {
    opts := &mcp.ServerOptions{
        Instructions: "Aseprite MCP server for pixel art and sprite manipulation",
    }

    server := mcp.NewServer(
        &mcp.Implementation{
            Name:    "aseprite-mcp",
            Version: "v0.1.0",
        },
        opts,
    )

    // Register all tools
    registerCanvasTools(server, cfg)
    registerDrawingTools(server, cfg)
    registerAnimationTools(server, cfg)
    registerSelectionTools(server, cfg)
    registerPaletteTools(server, cfg)
    registerTransformTools(server, cfg)
    registerExportTools(server, cfg)

    return server
}
```

### Aseprite Client Abstraction

```go
type Client struct {
    execPath string
    tempDir  string
    timeout  time.Duration
}

// ExecuteLua runs a Lua script in Aseprite batch mode
func (c *Client) ExecuteLua(ctx context.Context, script string, filename string) (string, error) {
    // 1. Write script to temp file
    // 2. Build command: aseprite --batch [filename] --script <tmpfile>
    // 3. Execute with context timeout
    // 4. Clean up temp file
    // 5. Return stdout/stderr
}

// ExecuteCommand runs Aseprite CLI command
func (c *Client) ExecuteCommand(ctx context.Context, args []string) (string, error) {
    // Direct command execution for non-Lua operations
}
```

### Tool Categories

## 1. Canvas Management Tools

### `create_canvas`
- **Input**: width, height, color_mode ("rgb", "grayscale", "indexed")
- **Output**: filepath to created .aseprite file
- **Lua API**: `Sprite(width, height, colorMode)`

### `open_sprite`
- **Input**: filepath
- **Output**: sprite metadata (width, height, frames, layers)
- **Lua API**: `app.open(filename)`

### `add_layer`
- **Input**: sprite_path, layer_name, layer_type ("normal", "group", "tilemap")
- **Output**: success status
- **Lua API**: `sprite:newLayer()`, `layer.name = name`

### `add_frame`
- **Input**: sprite_path, duration_ms
- **Output**: frame number
- **Lua API**: `sprite:newFrame()`, `frame.duration = duration`

### `delete_layer`
- **Input**: sprite_path, layer_name
- **Output**: success status

### `delete_frame`
- **Input**: sprite_path, frame_number
- **Output**: success status

## 2. Drawing Tools

### `draw_pixels`
- **Input**: sprite_path, layer_name, frame_number, pixels: [{x, y, color}]
- **Output**: success status
- **Lua API**: `image:putPixel(x, y, color)`

### `draw_line`
- **Input**: sprite_path, layer, frame, x1, y1, x2, y2, color, thickness
- **Output**: success status
- **Lua API**: `app.useTool{tool="line", color=color, points={Point(x1,y1), Point(x2,y2)}}`

### `draw_rectangle`
- **Input**: sprite_path, layer, frame, x, y, width, height, color, filled
- **Output**: success status
- **Lua API**: `app.useTool{tool="rectangle", ...}`

### `draw_circle`
- **Input**: sprite_path, layer, frame, center_x, center_y, radius, color, filled
- **Output**: success status
- **Lua API**: `app.useTool{tool="ellipse", ...}`

### `fill_area`
- **Input**: sprite_path, layer, frame, x, y, color
- **Output**: success status
- **Lua API**: `app.useTool{tool="paint_bucket", ...}`

### `draw_contour`
- **Input**: sprite_path, layer, frame, points: [{x, y}], color, closed
- **Output**: success status
- **Lua API**: Multiple `useTool` calls with "line"

## 3. Animation Tools

### `set_frame_duration`
- **Input**: sprite_path, frame_number, duration_ms
- **Output**: success status
- **Lua API**: `frame.duration = duration`

### `create_tag`
- **Input**: sprite_path, tag_name, from_frame, to_frame, direction ("forward", "reverse", "pingpong")
- **Output**: success status
- **Lua API**: `sprite:newTag(from, to)`, `tag.name = name`, `tag.aniDir = AniDir.FORWARD`

### `duplicate_frame`
- **Input**: sprite_path, source_frame, insert_after
- **Output**: new frame number
- **Lua API**: `sprite:newFrame(frameNumber)`

### `link_cel`
- **Input**: sprite_path, layer_name, source_frame, target_frame
- **Output**: success status
- **Lua API**: `sprite:newCel(layer, frame, cel.image, cel.position)`

## 4. Selection Tools

### `select_rectangle`
- **Input**: sprite_path, x, y, width, height, mode ("replace", "add", "subtract", "intersect")
- **Output**: success status
- **Lua API**: `Selection()`, `selection:select(Rectangle(...))`

### `select_ellipse`
- **Input**: sprite_path, x, y, width, height, mode
- **Output**: success status

### `select_all`
- **Input**: sprite_path
- **Output**: success status
- **Lua API**: `selection:selectAll()`

### `deselect`
- **Input**: sprite_path
- **Output**: success status
- **Lua API**: `selection:deselect()`

### `transform_selection`
- **Input**: sprite_path, dx, dy (translation)
- **Output**: success status

### `cut_selection`
- **Input**: sprite_path, layer, frame
- **Output**: success status
- **Lua API**: `app.command.Cut()`

### `copy_selection`
- **Input**: sprite_path
- **Output**: success status
- **Lua API**: `app.command.Copy()`

### `paste_clipboard`
- **Input**: sprite_path, layer, frame, x, y
- **Output**: success status
- **Lua API**: `app.command.Paste()`

## 5. Palette Tools

### `create_palette`
- **Input**: sprite_path, colors: ["#RRGGBB"]
- **Output**: success status
- **Lua API**: `Palette()`, `palette:setColor(index, Color())`

### `get_palette`
- **Input**: sprite_path
- **Output**: array of colors
- **Lua API**: `sprite.palettes[1]`, iterate `palette:getColor(i)`

### `set_palette_color`
- **Input**: sprite_path, index, color
- **Output**: success status

### `add_palette_color`
- **Input**: sprite_path, color
- **Output**: color index

### `sort_palette`
- **Input**: sprite_path, method ("hue", "saturation", "brightness")
- **Output**: success status
- **Lua API**: `app.command.SortPalette{...}`

## 6. Transform & Filter Tools

### `flip_sprite`
- **Input**: sprite_path, direction ("horizontal", "vertical")
- **Output**: success status
- **Lua API**: `sprite:flip(FlipType.HORIZONTAL)`

### `rotate_sprite`
- **Input**: sprite_path, angle (90, 180, 270, or arbitrary)
- **Output**: success status
- **Lua API**: `sprite:rotate(angle)`

### `scale_sprite`
- **Input**: sprite_path, scale_x, scale_y, algorithm ("nearest", "bilinear")
- **Output**: success status
- **Lua API**: `sprite:resize(width, height)`

### `crop_sprite`
- **Input**: sprite_path, x, y, width, height
- **Output**: success status
- **Lua API**: `sprite:crop(Rectangle(...))`

### `apply_outline`
- **Input**: sprite_path, layer, frame, color, thickness
- **Output**: success status
- **Lua API**: `app.command.Outline{...}`

## 7. Export/Import Tools

### `export_sprite`
- **Input**: sprite_path, output_path, format ("png", "gif", "jpg"), frame_number (optional)
- **Output**: exported file path
- **Lua API**: `sprite:saveCopyAs(filename)` or CLI `--save-as`

### `export_spritesheet`
- **Input**: sprite_path, output_path, layout ("horizontal", "vertical", "packed"), padding
- **Output**: spritesheet path + JSON metadata
- **Lua API**: `app.command.ExportSpriteSheet{...}` or CLI `--sheet`

### `import_image`
- **Input**: sprite_path, image_path, layer_name, frame_number
- **Output**: success status
- **Lua API**: `Image{fromFile=path}`, `sprite:newCel(...)`

## Tool Registration Pattern

```go
type CanvasToolInput struct {
    Width     int    `json:"width" jsonschema:"required,minimum=1,maximum=65535,description=Canvas width in pixels"`
    Height    int    `json:"height" jsonschema:"required,minimum=1,maximum=65535,description=Canvas height in pixels"`
    ColorMode string `json:"color_mode" jsonschema:"enum=rgb,enum=grayscale,enum=indexed,description=Color mode for the sprite"`
}

type CanvasToolOutput struct {
    FilePath string `json:"file_path" jsonschema:"description=Path to the created Aseprite file"`
}

func registerCanvasTools(server *mcp.Server, cfg *Config) {
    client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)

    mcp.AddTool(
        server,
        &mcp.Tool{
            Name:        "create_canvas",
            Description: "Create a new Aseprite sprite with specified dimensions and color mode",
        },
        func(ctx context.Context, req *mcp.CallToolRequest, input CanvasToolInput) (*mcp.CallToolResult, *CanvasToolOutput, error) {
            script := lua.CreateCanvas(input.Width, input.Height, input.ColorMode)
            output, err := client.ExecuteLua(ctx, script, "")
            if err != nil {
                return nil, nil, err
            }
            return nil, &CanvasToolOutput{FilePath: output}, nil
        },
    )
}
```

## Lua Script Generation

### Template-based approach

```go
package lua

import (
    "fmt"
    "strings"
)

// CreateCanvas generates Lua script to create a new sprite
func CreateCanvas(width, height int, colorMode string) string {
    cm := map[string]string{
        "rgb":       "ColorMode.RGB",
        "grayscale": "ColorMode.GRAYSCALE",
        "indexed":   "ColorMode.INDEXED",
    }[colorMode]

    return fmt.Sprintf(`
local spr = Sprite(%d, %d, %s)
local filename = os.tmpname() .. ".aseprite"
spr:saveAs(filename)
print(filename)
`, width, height, cm)
}

// DrawPixels generates Lua for batch pixel drawing
func DrawPixels(layer, frame string, pixels []Pixel) string {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf(`
local spr = app.activeSprite
local layer = spr.layers["%s"]
local cel = layer:cel(%d)
if not cel then
    cel = spr:newCel(layer, %d)
end
local img = cel.image
app.transaction(function()
`, layer, frame, frame))

    for _, p := range pixels {
        sb.WriteString(fmt.Sprintf(
            "    img:putPixel(%d, %d, Color(%d, %d, %d, %d))\n",
            p.X, p.Y, p.R, p.G, p.B, p.A,
        ))
    }

    sb.WriteString("end)\nspr:saveAs(spr.filename)\n")
    return sb.String()
}
```

## Configuration

### Configuration File
- Required config file: `~/.config/aseprite-mcp/config.json`
- Fields:
  - `aseprite_path` (required): Full absolute path to Aseprite executable
  - `temp_dir` (optional): Directory for temporary files (default: system temp)
  - `timeout` (optional): Timeout for operations in seconds (default: 30)
  - `log_level` (optional): Logging verbosity (default: info)

### Configuration File (optional)
```json
{
  "aseprite_path": "/usr/local/bin/aseprite",
  "temp_dir": "/tmp/aseprite-mcp",
  "timeout": 60,
  "log_level": "info"
}
```

## Error Handling

### Error Types
1. **Configuration errors**: Missing Aseprite binary, invalid paths
2. **Execution errors**: Aseprite command failures, timeouts
3. **Validation errors**: Invalid tool parameters (caught by JSON schema)
4. **File errors**: Missing sprites, permission issues

### Error Response Pattern
```go
func handleAsepriteError(err error, stderr string) error {
    if err == context.DeadlineExceeded {
        return fmt.Errorf("aseprite operation timed out")
    }
    if stderr != "" {
        return fmt.Errorf("aseprite error: %s", stderr)
    }
    return fmt.Errorf("aseprite execution failed: %w", err)
}
```

## Testing Strategy

### Unit Tests
- Lua script generation correctness
- Parameter validation
- Error handling

### Integration Tests
- Real Aseprite executable required (no mocks)
- Test configuration must provide explicit path to Aseprite
- Test full tool workflows with actual Aseprite
- Verify generated sprites

### Example Test
```go
func TestCreateCanvas(t *testing.T) {
    // Load test config with real Aseprite path
    cfg := &Config{
        AsepritePath: "/absolute/path/to/aseprite",  // Must be real executable
        TempDir:      t.TempDir(),
    }
    server := NewServer(cfg)

    // Test via in-memory transport
    clientTransport, serverTransport := mcp.NewInMemoryTransports()
    // ... test tool invocation with real Aseprite
}
```

## Future Enhancements

### Phase 1 (MVP)
- Basic canvas, drawing, and export tools
- Stdio transport only
- Manual sprite path management

### Phase 2
- Animation tools (tags, cel linking)
- Selection and transform tools
- Resource exposure for sprite previews

### Phase 3
- Palette tools
- Advanced filters (outline, shade)
- HTTP/SSE transport support
- Sprite state management (track open sprites)

### Phase 4
- Text rendering tool
- Tilemap support
- Plugin script execution
- Aseprite extension integration

## Performance Considerations

1. **Temp File Management**: Clean up Lua scripts after execution
2. **Batch Operations**: Combine multiple pixel draws into single transaction
3. **Caching**: Consider caching open sprites for repeated operations
4. **Concurrency**: Use context for timeouts, avoid blocking on long operations
5. **Resource Limits**: Implement max sprite size, max pixels per operation

## Security Considerations

1. **Path Validation**: Sanitize all file paths to prevent directory traversal
2. **Script Injection**: Escape all user input in Lua scripts
3. **Resource Limits**: Prevent DoS via large sprites or infinite loops
4. **Temp File Cleanup**: Ensure temp files are always cleaned up
5. **Command Injection**: Use proper argument passing, not shell execution

## References

- [Aseprite Lua API](https://github.com/aseprite/api)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://modelcontextprotocol.io/)
- [Aseprite CLI Documentation](https://www.aseprite.org/docs/cli/)

## Success Metrics

1. **Coverage**: Expose 80%+ of commonly-used Aseprite Lua API
2. **Usability**: AI can generate simple sprites without errors
3. **Performance**: Sub-second response for basic operations
4. **Reliability**: 99%+ success rate on valid operations
5. **Adoption**: Integrate with major AI coding assistants