# Aseprite MCP Server

> **AI-Powered Pixel Art Creation** - Connect Aseprite to AI assistants like Claude for natural language sprite creation and manipulation.

A Model Context Protocol (MCP) server that exposes Aseprite's pixel art and animation capabilities to AI assistants, enabling you to create and edit sprites using natural language.

## Why Use This?

**For Pixel Artists & Game Designers:**
- Rapidly prototype sprite concepts by describing them to your AI assistant
- Automate repetitive tasks like creating animation frames or color variations
- Generate placeholder art for game development
- Experiment with pixel art ideas without manual pixel pushing

**For Game Developers:**
- Integrate sprite generation into your game development workflow
- Automate asset creation pipelines
- Quickly iterate on visual concepts during prototyping
- Generate test assets and mockups

**For AI Assistant Users:**
- Use Claude Desktop or other MCP clients to create pixel art through conversation
- No need to learn Aseprite's UI - just describe what you want
- Perfect for non-artists who need simple sprites and animations

## Features

- **Canvas & Layer Management:** RGB, Grayscale, and Indexed color modes with multi-layer support
- **Drawing Primitives:** Pixels, lines, rectangles, circles, flood fill with batch operations
- **Professional Pixel Art Tools:**
  - **Reference Analysis:** Extract palettes, brightness maps, edge detection, and composition guides from images
  - **Dithering:** Bayer matrix (2x2, 4x4, 8x8) and checkerboard patterns for smooth gradients and textures
  - **Palette Extraction:** K-means clustering in LAB color space for perceptually accurate color reduction
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
| `get_sprite_info` | Get sprite metadata (size, layers, frames) |

### Drawing & Painting
| Tool | Description |
|------|-------------|
| `draw_pixels` | Draw individual pixels (supports batch operations) |
| `draw_line` | Draw a line between two points |
| `draw_rectangle` | Draw a rectangle (filled or outline) |
| `draw_circle` | Draw a circle/ellipse (filled or outline) |
| `fill_area` | Flood fill from a point (paint bucket) |

### Professional Pixel Art
| Tool | Description |
|------|-------------|
| `analyze_reference` | Extract palette, brightness map, edges, and composition from reference images |
| `draw_with_dither` | Fill region with dithering patterns (15 patterns: Bayer, checkerboard, grass, water, stone, cloud, brick, etc.) |
| `downsample_image` | Downsample high-res images to pixel art dimensions using box filter |
| `set_palette` | Set sprite's color palette to specified colors (supports 1-256 colors) |
| `apply_shading` | Apply palette-constrained shading based on light direction (smooth, hard, or pillow styles) |
| `analyze_palette_harmonies` | Analyze palette for complementary, triadic, analogous relationships and color temperature |

### Animation
| Tool | Description |
|------|-------------|
| `add_frame` | Add a new animation frame |
| `set_frame_duration` | Set the duration of an animation frame in milliseconds |
| `create_tag` | Create an animation tag with playback direction |
| `duplicate_frame` | Duplicate an existing frame with all cels |
| `link_cel` | Create a linked cel that shares image data |

### Inspection & Export
| Tool | Description |
|------|-------------|
| `get_pixels` | Read pixel data from a rectangular region (paginated, for verification) |
| `export_sprite` | Export sprite to PNG/GIF/JPG/BMP |

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

- Canvas creation: ~94ms
- Drawing primitives: ~93-96ms
- 10K pixel batch: ~109ms
- Complete workflows: ~280-840ms

See [docs/BENCHMARKS.md](docs/BENCHMARKS.md) for detailed benchmark results.

## License

MIT