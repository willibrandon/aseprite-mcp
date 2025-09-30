# Aseprite MCP Server (Go)

A Model Context Protocol (MCP) server that exposes Aseprite's pixel art and animation capabilities to AI assistants and other MCP clients.

## Status

**MVP Feature Complete** - Core functionality implemented, testing in progress

### What Works

- Canvas creation (RGB, Grayscale, Indexed color modes)
- Layer and frame management
- Drawing primitives (pixels, lines, rectangles, circles, fill)
- Sprite export (PNG, GIF, JPG, BMP)
- Sprite metadata retrieval
- Full integration test suite
- Performance benchmarks (all PRD targets exceeded)

### What's Next

- Example client implementation
- Documentation and usage guides
- CI/CD pipeline setup

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
| `draw_pixels` | Draw individual pixels (supports batch operations) |
| `draw_line` | Draw a line between two points |
| `draw_rectangle` | Draw a rectangle (filled or outline) |
| `draw_circle` | Draw a circle/ellipse (filled or outline) |
| `fill_area` | Flood fill from a point (paint bucket) |
| `export_sprite` | Export sprite to PNG/GIF/JPG/BMP |

See [docs/API.md](docs/API.md) for detailed API documentation (coming soon).

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

All operations meet or exceed PRD performance targets:

- Canvas creation: ~94ms (target: <500ms)
- Drawing primitives: ~93-96ms (target: <300ms)
- 10K pixel batch: ~109ms (target: <2s)
- Complete workflows: ~280-840ms

See [docs/BENCHMARKS.md](docs/BENCHMARKS.md) for detailed results.

## Documentation

- [BENCHMARKS.md](docs/BENCHMARKS.md) - Performance benchmark results

## License

MIT