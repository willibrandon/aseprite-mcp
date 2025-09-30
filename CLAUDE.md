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
  - `lua.go` - Lua script generation utilities
  - `types.go` - Domain types (Color, Point, Rectangle, Pixel, etc.)
- `pkg/config/` - Configuration management (file-based only)
- `pkg/server/` - MCP server implementation
- `pkg/tools/` - MCP tool implementations organized by category:
  - `canvas.go` - Sprite/layer/frame management
  - `drawing.go` - Drawing primitives (pixels, lines, shapes)
  - `animation.go` - Animation and timeline operations
  - `selection.go` - Selection manipulation
  - `palette.go` - Color palette operations
  - `transform.go` - Transform and filter operations
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

**MVP Feature Complete** - Core functionality implemented and tested.

Completed features:
- Canvas creation and management (RGB, Grayscale, Indexed)
- Layer and frame operations
- Drawing primitives (pixels, lines, rectangles, circles, fill)
- Sprite export (PNG, GIF, JPG, BMP)
- Metadata retrieval
- Integration test suite with real Aseprite
- Performance benchmarks (all targets exceeded)

Remaining work:
- Example client implementation
- CI/CD pipeline setup

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

## Common Pitfalls

1. **Don't auto-discover Aseprite** - Always require explicit path in config file
2. **Don't mock Aseprite** - All tests must use real executable
3. **Don't forget Lua escaping** - User input MUST be escaped via `EscapeString()`
4. **Don't leave temp files** - Always defer cleanup in Client methods
5. **Don't skip transactions** - Wrap Lua mutations in `app.transaction()`
6. **Don't assume sprite is open** - Check `app.activeSprite` in Lua scripts
7. **Don't hardcode paths** - Use `filepath` package for cross-platform path handling

## Dependencies

Core dependencies (see `go.mod` for versions):
- `github.com/modelcontextprotocol/go-sdk` - Official MCP SDK
- `github.com/willibrandon/mtlog` - Structured logging (Serilog-style for Go)

Aseprite requirement:
- Minimum version: 1.3.0
- Recommended: 1.3.10+

## Documentation References

- `docs/BENCHMARKS.md` - Performance benchmark results
- `README.md` - Quick start guide and tool reference