# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Aseprite MCP Server is a Model Context Protocol (MCP) server implementation in Go that exposes Aseprite's pixel art and animation capabilities to AI assistants. The server acts as a bridge between MCP clients (like Claude Desktop) and Aseprite's Lua scripting API, enabling programmatic sprite creation and manipulation.

**Key Technologies:**
- Go 1.23+
- Official MCP Go SDK (`github.com/modelcontextprotocol/go-sdk`)
- Aseprite Lua API (via batch mode execution)
- mtlog for structured logging

## Build Commands

```bash
# Build the server binary
make build                    # Outputs to bin/aseprite-mcp

# Run tests
make test                     # Unit tests only
go test -tags=integration ./... # Unit + integration tests (requires Aseprite)
make test-coverage            # Generate HTML coverage report

# Linting and formatting
make lint                     # Run go vet and go fmt

# Clean build artifacts
make clean

# Install to GOPATH
make install
```

## Development Commands

```bash
# Run a single test
go test -v -run TestName ./pkg/aseprite

# Run tests with race detector
go test -race ./...

# Run only unit tests (no integration)
go test ./...

# Update dependencies
go mod tidy

# View module dependencies
go mod graph
```

## Architecture

### High-Level Design

The server follows a three-layer architecture:

1. **MCP Layer** (`pkg/server/`): Handles MCP protocol communication, tool registration, and request/response marshaling
2. **Tool Layer** (`pkg/tools/`): Implements MCP tools for canvas, drawing, animation, selection, palette, transform, and export operations
3. **Aseprite Layer** (`pkg/aseprite/`): Abstracts Aseprite command execution and Lua script generation

### Key Components

**Configuration (`pkg/config/config.go`):**
- Loads config from environment variables and optional JSON file
- Auto-discovers Aseprite executable on Windows/macOS/Linux
- Validates paths, permissions, and settings
- Priority: env vars > config file > defaults

**Aseprite Client (`pkg/aseprite/client.go`):**
- Executes Aseprite commands via `exec.CommandContext`
- Manages Lua script lifecycle (create temp file → execute → cleanup)
- Handles timeouts using context.Context
- Cleans up old temp files

**Lua Generator (`pkg/aseprite/lua.go`):**
- Generates type-safe Lua scripts for Aseprite operations
- `EscapeString()` prevents injection attacks
- Helper formatters: `FormatColor()`, `FormatPoint()`, `FormatRectangle()`
- Script generators for canvas, drawing, export, etc.

**Domain Types (`pkg/aseprite/types.go`):**
- `Color`: RGBA with hex parsing (`FromHex()`, `ToHex()`)
- `Point`, `Rectangle`, `Pixel`: Geometric primitives
- `SpriteInfo`: Sprite metadata (width, height, layers, frames)
- `ColorMode`: RGB, Grayscale, Indexed with Lua conversion

### Tool Implementation Pattern

Each tool follows this structure:
1. Define input/output structs with JSON schema tags
2. Use `mcp.AddTool()` to register with server
3. Handler validates input → generates Lua script → executes → parses output
4. Wrap operations in `app.transaction()` for atomicity
5. Always save sprite after modifications

Example:
```go
type CreateCanvasInput struct {
    Width     int    `json:"width" jsonschema:"required,minimum=1,maximum=65535"`
    Height    int    `json:"height" jsonschema:"required,minimum=1,maximum=65535"`
    ColorMode string `json:"color_mode" jsonschema:"enum=rgb,enum=grayscale,enum=indexed"`
}

func registerCanvasTools(server *mcp.Server, client *Client, cfg *Config) {
    mcp.AddTool(server, &mcp.Tool{
        Name: "create_canvas",
        Description: "Create a new Aseprite sprite",
    }, func(ctx context.Context, req *mcp.CallToolRequest, input CreateCanvasInput) (*mcp.CallToolResult, *CreateCanvasOutput, error) {
        script := lua.CreateCanvas(input.Width, input.Height, input.ColorMode)
        output, err := client.ExecuteLua(ctx, script, "")
        // ... handle response
    })
}
```

## Testing Strategy

### Unit Tests
- Mock Aseprite using `internal/testutil/MockAseprite`
- Test Lua generation correctness (escaping, formatting)
- Test parameter validation and error handling
- Target: 80%+ coverage

### Integration Tests
- Use build tag `integration` to separate from unit tests
- Require real Aseprite installation
- Test actual script execution and sprite generation
- Run manually or in CI with Aseprite available

### Running Tests
```bash
# Unit only (fast, no Aseprite needed)
go test ./...

# Integration (requires Aseprite in PATH)
go test -tags=integration ./...

# Specific package
go test -v ./pkg/aseprite

# With coverage
go test -cover ./...
```

## Important Implementation Details

### Security & Safety
- **Path Validation**: All file paths must be validated to prevent traversal attacks
- **Lua Injection Prevention**: Use `EscapeString()` for all user-provided strings in Lua scripts
- **Resource Limits**: Max sprite size 65535x65535, max pixels/operation 100K
- **Temp File Cleanup**: Always use defer to clean up temp Lua scripts
- **Timeout Enforcement**: All operations use context.Context with configurable timeout

### Error Handling
- Configuration errors: Missing Aseprite, invalid paths → exit with clear message
- Validation errors: Invalid params → return JSON schema validation error
- Execution errors: Aseprite failures → include stderr in error message
- Timeout errors: Context deadline exceeded → return specific timeout error
- Always log errors with structured context (operation, duration, parameters)

### Performance Considerations
- Batch pixel operations into single transaction
- Clean up temp files periodically (not just on exit)
- Use context timeouts to prevent hanging operations
- Consider caching open sprites in future (not MVP)

### Aseprite Integration Quirks
- Batch mode requires `--batch` flag
- Scripts execute with `--script <file>`
- Sprite files must exist before opening with `--batch <sprite> --script`
- Frame numbers are 1-indexed in Lua
- Layer lookup uses `sprite:findLayerByName()` (case-sensitive)
- Always save sprite after modifications: `spr:saveAs(spr.filename)`

## Documentation References

- **Design**: `docs/DESIGN.md` - Full architecture and tool specifications
- **PRD**: `docs/PRD.md` - Requirements, success metrics, release plan
- **Implementation**: `docs/IMPLEMENTATION_GUIDE.md` - Step-by-step implementation chunks
- **Aseprite API**: https://github.com/aseprite/api - Official Lua API reference
- **MCP Spec**: https://modelcontextprotocol.io/ - Protocol specification

## Development Workflow

### Starting a New Tool
1. Define input/output structs in `pkg/tools/<category>.go`
2. Add JSON schema validation tags
3. Implement Lua generator in `pkg/aseprite/lua.go`
4. Add script generator tests in `pkg/aseprite/lua_test.go`
5. Register tool in server using `mcp.AddTool()`
6. Add integration test in `pkg/tools/<category>_test.go`

### Adding Tests
- Unit tests go next to source: `file.go` → `file_test.go`
- Integration tests use `//go:build integration` tag
- Use `t.TempDir()` for temporary files
- Use `testutil.CreateTestConfig()` for mock Aseprite setup

### Common Gotchas
- Windows path handling: Use `filepath.Join()` not string concat
- Lua escaping: Always use `EscapeString()` for user input
- Frame indexing: Aseprite uses 1-based, Go uses 0-based
- Temp directory: Create with `os.MkdirAll()`, clean with defer
- Context cancellation: Always pass context through execution chain

## Code Style

Follow standard Go conventions per global CLAUDE.md:
- Prefer early returns to reduce nesting
- Use table-driven tests
- Error handling: return early on errors
- Interfaces defined by consumers, not providers
- Functional options pattern for configurable APIs
- Comprehensive GoDoc on all exported symbols