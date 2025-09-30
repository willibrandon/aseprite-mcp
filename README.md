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

- Canvas creation with RGB, Grayscale, and Indexed color modes
- Layer and frame management for complex compositions
- Drawing primitives: pixels, lines, rectangles, circles, flood fill
- Animation tools: frame durations, tags, frame duplication, linked cels
- Sprite export to PNG, GIF, JPG, and BMP formats
- Sprite metadata retrieval
- Cross-platform support: Windows, macOS, Linux
- Batch pixel operations for performance
- Integration with Claude Desktop and other MCP clients

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

| Tool | Description |
|------|-------------|
| `create_canvas` | Create new sprite with specified dimensions and color mode |
| `add_layer` | Add a new layer to the sprite |
| `add_frame` | Add a new animation frame |
| `get_sprite_info` | Get sprite metadata (size, layers, frames) |
| `get_pixels` | Read pixel data from a rectangular region (returns colors and coordinates) |
| `draw_pixels` | Draw individual pixels (supports batch operations) |
| `draw_line` | Draw a line between two points |
| `draw_rectangle` | Draw a rectangle (filled or outline) |
| `draw_circle` | Draw a circle/ellipse (filled or outline) |
| `fill_area` | Flood fill from a point (paint bucket) |
| `export_sprite` | Export sprite to PNG/GIF/JPG/BMP |
| `set_frame_duration` | Set the duration of an animation frame in milliseconds |
| `create_tag` | Create an animation tag with playback direction |
| `duplicate_frame` | Duplicate an existing frame with all cels |
| `link_cel` | Create a linked cel that shares image data |

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