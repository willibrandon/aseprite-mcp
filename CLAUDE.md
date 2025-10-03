# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

An MCP (Model Context Protocol) server implementation in Go that exposes Aseprite's pixel art and animation capabilities through programmatic access. The server integrates with Aseprite's command-line interface and Lua scripting API to enable AI assistants to create and manipulate sprites.

## Build Commands

```bash
# Build the server binary
make build
# or: go build -o bin/aseprite-mcp ./cmd/aseprite-mcp

# Run tests
make test
# or: go test -v -race -cover ./...

# Run integration tests (requires real Aseprite)
go test -tags=integration ./...

# Generate test coverage report
make test-coverage

# Run linters
make lint

# Clean build artifacts
make clean

# Install binary
make install
```

## Configuration Requirements

**CRITICAL**: This project requires explicit configuration - no automatic discovery or environment variables.

All operations require a configuration file at `~/.config/aseprite-mcp/config.json`:
```json
{
  "aseprite_path": "/absolute/path/to/aseprite",
  "temp_dir": "/tmp/aseprite-mcp",
  "timeout": 30,
  "log_level": "info"
}
```

The `aseprite_path` MUST be an absolute path to a real Aseprite executable. The project does NOT:
- Auto-discover Aseprite installations
- Use environment variables for Aseprite path
- Search system PATH for Aseprite

## Testing Philosophy

**ALL tests use real Aseprite** - no mocks, stubs, or fakes for Aseprite integration.

- Unit tests require config file with real Aseprite path
- Integration tests (tagged with `//go:build integration`) also use real Aseprite
- Test utilities are in `internal/testutil/`
- Use `testutil.LoadTestConfig(t)` to load test configuration
- See `docs/TESTING.md` for detailed testing requirements

## Architecture

### High-Level Structure

The server acts as a bridge between MCP clients (AI assistants) and Aseprite:

```
MCP Client → MCP Server (Go) → Lua Script Generation → Aseprite CLI (--batch --script)
```

### Package Organization

- `cmd/aseprite-mcp/` - Server entry point and initialization
- `pkg/aseprite/` - Core Aseprite integration
  - `client.go` - Command execution and Lua script execution
  - `lua.go` - Lua script generation utilities (includes palette and dithering generators)
  - `types.go` - Domain types (Color, Point, Rectangle, Pixel, etc.)
  - `palette.go` - K-means palette extraction and color analysis
  - `image_analysis.go` - Brightness maps, Sobel edge detection, composition analysis
- `pkg/config/` - Configuration management (file-based only)
- `pkg/server/` - MCP server implementation
- `pkg/tools/` - MCP tool implementations organized by category:
  - `canvas.go` - Sprite/layer/frame management (create_sprite, add_layer, add_frame, delete_layer with protection, delete_frame with protection)
  - `drawing.go` - Drawing primitives (pixels, lines, rectangles, circles, fill, contours for polylines/polygons)
  - `selection.go` - Selection and clipboard operations (8 tools)
  - `animation.go` - Animation and timeline operations
  - `inspection.go` - Pixel data inspection and reading
  - `analysis.go` - Reference image analysis (palette extraction, edge detection, composition)
  - `dithering.go` - Dithering patterns for gradients and textures (15 patterns)
  - `palette_tools.go` - Palette management (set_palette, apply_shading, analyze_palette_harmonies)
  - `transform.go` - Transform operations (downsampling)
  - `export.go` - Export and import operations
- `internal/testutil/` - Testing utilities (no mocks)

### Core Workflow

1. MCP client sends tool request with parameters
2. Server validates parameters via JSON schema
3. Appropriate Lua script is generated from template
4. Script written to temp file with restricted permissions (0600)
5. Aseprite executed in batch mode: `aseprite --batch [sprite] --script [temp.lua]`
6. Output parsed and returned to client
7. Temp files cleaned up (always, even on error)

### Key Design Constraints

- **Stateless**: Each operation is independent, no shared state between requests
- **Batch Mode Only**: All Aseprite operations run in headless `--batch` mode
- **Lua-based**: Operations use Aseprite's Lua API, not GUI automation
- **File-centric**: Operations work with sprite files on disk, not in-memory state
- **Security**: Lua script injection prevention via proper escaping (see `EscapeString`)

## Implementation Status

Core functionality implemented and tested:
- Canvas creation and management (RGB, Grayscale, Indexed)
- Layer and frame operations (add, delete with last-layer/frame protection)
- Drawing primitives (pixels, lines, rectangles, circles, fill) with optional palette-aware color snapping
- Advanced drawing: Contour tool for drawing polylines and closed polygons with points arrays
- **Selection and Clipboard Tools (8 tools):**
  - Rectangular and elliptical selections with 4 modes (replace/add/subtract/intersect)
  - Select all, deselect, move selection
  - Cut, copy, paste operations
  - **Important limitation**: Selections are transient and do NOT persist in .aseprite files - they only exist in memory during script execution
- Animation tools (frame duration, tags, duplication, linked cels)
- Inspection tools (pixel data reading with pagination for verification and analysis)
- **Professional Pixel Art Tools:**
  - Reference image analysis (k-means palette extraction, brightness maps, Sobel edge detection)
  - Composition analysis (rule of thirds, focal points)
  - Dithering with 15 patterns: Bayer matrices (2x2, 4x4, 8x8), checkerboard, and texture patterns (grass, water, stone, cloud, brick, dots, diagonal, cross, noise, horizontal_lines, vertical_lines)
  - Image downsampling with box filter for pixel art conversion
  - **Palette Management Tools (5 tools):**
    - `get_palette`: Retrieve current palette as hex color array with size
    - `set_palette`: Set entire palette with 1-256 colors
    - `set_palette_color`: Modify specific palette index (0-255)
    - `add_palette_color`: Add new color to palette (returns index)
    - `sort_palette`: Sort by hue/saturation/brightness/luminance (ascending/descending)
  - Palette-constrained shading (smooth, hard, pillow styles with 8 light directions)
  - Palette harmony analysis (complementary, triadic, analogous relationships, color temperature)
  - Palette-aware drawing: All drawing tools support `use_palette` flag to snap arbitrary colors to nearest palette color
  - Antialiasing suggestions: Detect jagged diagonal edges and suggest intermediate colors to smooth curves (suggest_antialiasing tool with auto-apply option)
- Sprite export (PNG, GIF, JPG, BMP)
- Metadata retrieval
- Example client implementation (examples/client/main.go)
- Integration test suite with real Aseprite
- Performance benchmarks (all targets exceeded)
- Cross-platform CI with GitHub Actions

## Code Style Guidelines

Follow Go standard library conventions:
- Use `gofmt` for formatting (enforced in `make lint`)
- Pass `go vet` and `golangci-lint` with zero warnings
- GoDoc comments on all exported types and functions
- Table-driven tests with descriptive names
- Error wrapping with `fmt.Errorf("%w", err)` for context

Lua script generation:
- Always use `EscapeString()` for user input in generated scripts
- Wrap mutations in `app.transaction()` for atomicity
- Include error checks in Lua: `if not spr then error("...") end`
- Save sprite after mutations: `spr:saveAs(spr.filename)`
- **Exception**: Do NOT save after selection operations - selections don't persist in .aseprite files

## Common Pitfalls

1. **Don't auto-discover Aseprite** - Always require explicit path in config file
2. **Don't mock Aseprite** - All tests must use real executable
3. **Don't forget Lua escaping** - User input MUST be escaped via `EscapeString()`
4. **Don't leave temp files** - Always defer cleanup in Client methods
5. **Don't skip transactions** - Wrap Lua mutations in `app.transaction()`
6. **Don't assume sprite is open** - Check `app.activeSprite` in Lua scripts
7. **Don't hardcode paths** - Use `filepath` package for cross-platform path handling
8. **Don't delete last layer/frame** - Aseprite requires at least one layer and one frame; validate before deletion
9. **Don't save after selections** - Selections are transient and lost on save; combine selection workflows in single scripts
10. **Don't expect clipboard persistence** - Copy/Paste operations must be combined in single Lua scripts (clipboard doesn't persist across processes)

## Dependencies

Core dependencies (see `go.mod` for versions):
- `github.com/modelcontextprotocol/go-sdk` - Official MCP SDK
- `github.com/willibrandon/mtlog` - Structured logging (Serilog-style for Go)

Professional pixel art tools:
- `github.com/lucasb-eyer/go-colorful` - Color space conversions (RGB, HSL, LAB)
- `github.com/nfnt/resize` - Image scaling with box filter
- `gonum.org/v1/gonum` - K-means clustering for palette extraction

Aseprite requirement:
- Minimum version: 1.3.0
- Recommended: 1.3.10+

## Documentation References

- `docs/BENCHMARKS.md` - Performance benchmark results
- `README.md` - Quick start guide and tool reference