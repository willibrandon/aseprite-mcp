# Aseprite MCP Server

A Model Context Protocol (MCP) server that exposes Aseprite's pixel art and animation capabilities to AI assistants.

## Features

- **Canvas Management**: Create sprites, add layers and frames, get sprite info
- **Drawing Primitives**: Draw pixels, lines, rectangles, circles, and flood fill
- **Export**: Export sprites to PNG, GIF, and other formats
- **Type-Safe**: Uses Go's type system and JSON schema validation
- **Stdio Transport**: Works with Claude Desktop and other MCP clients

## Requirements

- Go 1.23+
- Aseprite 1.3.0+

## Installation

### From Source

```bash
git clone https://github.com/willibrandon/aseprite-mcp-go.git
cd aseprite-mcp-go
go build -o bin/aseprite-mcp.exe ./cmd/aseprite-mcp
```

### Configuration

The server auto-discovers Aseprite on your system. You can override with environment variables:

```bash
# Windows
set ASEPRITE_PATH=C:\Program Files\Aseprite\Aseprite.exe
set ASEPRITE_TEMP_DIR=C:\Temp\aseprite-mcp
set ASEPRITE_TIMEOUT=60
set ASEPRITE_LOG_LEVEL=debug

# Linux/macOS
export ASEPRITE_PATH=/usr/local/bin/aseprite
export ASEPRITE_TEMP_DIR=/tmp/aseprite-mcp
export ASEPRITE_TIMEOUT=60
export ASEPRITE_LOG_LEVEL=debug
```

Or create a config file at `~/.config/aseprite-mcp/config.json`:

```json
{
  "aseprite_path": "/path/to/aseprite",
  "temp_dir": "/tmp/aseprite-mcp",
  "timeout": 60,
  "log_level": "info"
}
```

## Usage

### With Claude Desktop

Add to your Claude Desktop configuration (`%APPDATA%\Claude\claude_desktop_config.json` on Windows):

```json
{
  "mcpServers": {
    "aseprite": {
      "command": "D:\\path\\to\\aseprite-mcp.exe",
      "env": {
        "ASEPRITE_PATH": "C:\\Program Files\\Aseprite\\Aseprite.exe"
      }
    }
  }
}
```

### CLI Usage

```bash
# Check version
aseprite-mcp --version

# Health check (verify Aseprite is accessible)
aseprite-mcp --health

# Run server (stdio transport)
aseprite-mcp
```

## Available Tools

### Canvas Management

- `create_canvas` - Create a new sprite with specified dimensions and color mode
- `add_layer` - Add a new layer to an existing sprite
- `add_frame` - Add a new frame to the timeline
- `get_sprite_info` - Get metadata about a sprite (dimensions, frames, layers)

### Drawing

- `draw_pixels` - Draw individual pixels at specified coordinates
- `draw_line` - Draw a line between two points
- `draw_rectangle` - Draw a rectangle (filled or outline)
- `draw_circle` - Draw a circle (filled or outline)
- `fill_area` - Flood fill an area (paint bucket tool)

### Export

- `export_sprite` - Export sprite to PNG, GIF, or other formats

## Example Usage with AI

```
Create a 64x64 pixel sprite in RGB mode, add a layer called "Background",
draw a red rectangle from (10,10) to (54,54), then export it as sprite.png
```

The AI assistant will use the MCP tools to:
1. Call `create_canvas` with width=64, height=64, color_mode="rgb"
2. Call `add_layer` with layer_name="Background"
3. Call `draw_rectangle` with the specified coordinates and color
4. Call `export_sprite` to save as PNG

## Development

### Building

```bash
make build          # Build binary
make test           # Run tests
make test-coverage  # Generate coverage report
make lint           # Run linters
```

### Testing

```bash
# Unit tests (no Aseprite required)
go test ./...

# Integration tests (requires Aseprite)
go test -tags=integration ./...
```

### Project Structure

```
aseprite-mcp-go/
├── cmd/aseprite-mcp/    # Main entry point
├── pkg/
│   ├── aseprite/        # Aseprite client and Lua generation
│   ├── config/          # Configuration management
│   └── tools/           # MCP tool implementations
├── internal/testutil/   # Test utilities
└── docs/                # Documentation
```

## Documentation

- [Design Document](docs/DESIGN.md) - Architecture and implementation details
- [Product Requirements](docs/PRD.md) - Full requirements and specifications
- [Implementation Guide](docs/IMPLEMENTATION_GUIDE.md) - Step-by-step development guide
- [CLAUDE.md](CLAUDE.md) - Guide for AI coding assistants

## License

MIT