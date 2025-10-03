# Aseprite MCP Server

A Model Context Protocol (MCP) server that exposes Aseprite's pixel art and animation capabilities to AI assistants, enabling you to create and edit sprites using natural language.

## Features

- **Canvas & Layer Management:** RGB, Grayscale, and Indexed color modes with multi-layer support
- **Drawing Primitives:** Pixels, lines, rectangles, circles, flood fill with batch operations
- **Professional Pixel Art Tools:**
  - **Reference Analysis:** Extract palettes, brightness maps, edge detection, and composition guides from images
  - **Dithering:** Bayer matrix (2x2, 4x4, 8x8) and checkerboard patterns for smooth gradients and textures
  - **Palette Extraction:** K-means clustering in LAB color space for perceptually accurate color reduction
  - **Transform Operations:** Flip, rotate, scale, crop, resize canvas, and apply outlines
- **Animation Tools:** Frame durations, tags, frame duplication, linked cels
- **Inspection Tools:** Read pixel data with pagination for verification and analysis
- **Export Formats:** PNG, GIF, JPG, and BMP
- **Cross-platform:** Windows, macOS, Linux
- **MCP Integration:** Works with Claude Desktop and other MCP clients

## Requirements

- Go 1.23+
- Aseprite 1.3.0+ (1.3.10+ recommended)

## Quick Start

### 1. Install

```bash
# Clone the repository
git clone https://github.com/willibrandon/aseprite-mcp-go.git
cd aseprite-mcp-go

# Build the server
make build
```

### 2. Configure

Create `~/.config/aseprite-mcp/config.json`:

```json
{
  "aseprite_path": "/path/to/aseprite",
  "temp_dir": "/tmp/aseprite-mcp",
  "timeout": 30,
  "log_level": "info"
}
```

**Important**: The `aseprite_path` must be an absolute path to your Aseprite executable. No automatic discovery or environment variables are used.

### 3. Run

```bash
# Run the server (stdio transport)
./bin/aseprite-mcp

# Check health
./bin/aseprite-mcp --health

# Enable debug logging
./bin/aseprite-mcp --debug
```

## Usage with Claude Desktop

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "aseprite": {
      "command": "/absolute/path/to/aseprite-mcp"
    }
  }
}
```

Then use natural language to create sprites:
- "Create a 64x64 sprite and draw a red circle in the center"
- "Add a new layer called 'Background' and fill it with blue"
- "Export the sprite as a PNG"

## Available Tools

### Canvas & Layer Management
| Tool | Description |
|------|-------------|
| `create_canvas` | Create new sprite with specified dimensions and color mode |
| `add_layer` | Add a new layer to the sprite |
| `delete_layer` | Delete a layer from the sprite (cannot delete last layer) |
| `get_sprite_info` | Get sprite metadata (size, layers, frames) |

### Drawing & Painting
| Tool | Description |
|------|-------------|
| `draw_pixels` | Draw individual pixels (supports batch operations, optional palette snapping) |
| `draw_line` | Draw a line between two points (optional palette snapping) |
| `draw_contour` | Draw a polyline or polygon by connecting multiple points (optional palette snapping) |
| `draw_rectangle` | Draw a rectangle (filled or outline, optional palette snapping) |
| `draw_circle` | Draw a circle/ellipse (filled or outline, optional palette snapping) |
| `fill_area` | Flood fill from a point (paint bucket, optional palette snapping) |

### Selection & Clipboard
| Tool | Description |
|------|-------------|
| `select_rectangle` | Create a rectangular selection with mode (replace/add/subtract/intersect) |
| `select_ellipse` | Create an elliptical selection with mode (replace/add/subtract/intersect) |
| `select_all` | Select the entire canvas |
| `deselect` | Clear the current selection |
| `move_selection` | Move selection bounds by offset (does not move pixels) |
| `cut_selection` | Cut selected pixels to clipboard |
| `copy_selection` | Copy selected pixels to clipboard |
| `paste_clipboard` | Paste clipboard content at specified position |

**Important Note**: Selections are transient and exist only in memory during script execution. They do not persist when sprites are saved to disk. For workflows involving selections and clipboard operations (copy/paste), combine the operations in integration tests or custom Lua scripts rather than calling them as separate MCP tools.

### Professional Pixel Art
| Tool | Description |
|------|-------------|
| `analyze_reference` | Extract palette, brightness map, edges, and composition from reference images |
| `draw_with_dither` | Fill region with dithering patterns (15 patterns: Bayer, checkerboard, grass, water, stone, cloud, brick, etc.) |
| `downsample_image` | Downsample high-res images to pixel art dimensions using box filter |
| `get_palette` | Retrieve current sprite palette as array of hex colors with size |
| `set_palette` | Set sprite's color palette to specified colors (supports 1-256 colors) |
| `set_palette_color` | Set a specific palette index to a color (0-255) |
| `add_palette_color` | Add a new color to the palette (max 256 colors) |
| `sort_palette` | Sort palette by hue, saturation, brightness, or luminance (ascending/descending) |
| `apply_shading` | Apply palette-constrained shading based on light direction (smooth, hard, or pillow styles) |
| `analyze_palette_harmonies` | Analyze palette for complementary, triadic, analogous relationships and color temperature |
| `suggest_antialiasing` | Detect jagged diagonal edges and suggest intermediate colors for smooth curves (with optional auto-apply) |

### Transform & Filter
| Tool | Description |
|------|-------------|
| `flip_sprite` | Flip sprite, layer, or cel horizontally or vertically |
| `rotate_sprite` | Rotate sprite, layer, or cel by 90, 180, or 270 degrees |
| `scale_sprite` | Scale sprite with algorithm selection (nearest, bilinear, rotsprite) |
| `crop_sprite` | Crop sprite to rectangular region |
| `resize_canvas` | Resize canvas without scaling content (with anchor positioning) |
| `apply_outline` | Apply outline effect to layer with configurable color and thickness |

### Animation
| Tool | Description |
|------|-------------|
| `add_frame` | Add a new animation frame |
| `delete_frame` | Delete a frame from the sprite (cannot delete last frame) |
| `set_frame_duration` | Set the duration of an animation frame in milliseconds |
| `create_tag` | Create an animation tag with playback direction |
| `delete_tag` | Delete an animation tag by name |
| `duplicate_frame` | Duplicate an existing frame with all cels |
| `link_cel` | Create a linked cel that shares image data |

### Inspection & Export
| Tool | Description |
|------|-------------|
| `get_pixels` | Read pixel data from a rectangular region (paginated, for verification) |
| `export_sprite` | Export sprite to PNG/GIF/JPG/BMP |
| `export_spritesheet` | Export animation frames as spritesheet (horizontal/vertical/rows/columns/packed layout, optional JSON metadata) |
| `import_image` | Import external image file as a layer in the sprite |
| `save_as` | Save sprite to a new .aseprite file path |

## Examples

See [examples/](examples/) for a complete working example client that demonstrates all features including canvas creation, animation, dithering, palette management, shading, and antialiasing.

## Development

```bash
# Run tests
make test

# Run integration tests (requires configured Aseprite)
go test -tags=integration -v ./...

# Run benchmarks
go test -tags=integration -bench=. -benchmem ./pkg/aseprite ./pkg/tools

# Run linters
make lint

# Generate coverage report
make test-coverage
```

## Performance

All operations complete in under 100ms on typical hardware, with complete workflows finishing in 280-840ms. See [docs/BENCHMARKS.md](docs/BENCHMARKS.md) for detailed benchmark results.

## License

[MIT](LICENSE)