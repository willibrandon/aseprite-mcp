# Testing Guide

## Prerequisites

ALL tests require a real Aseprite installation with configuration file.

Create `~/.config/aseprite-mcp/config.json`:
```json
{
  "aseprite_path": "/absolute/path/to/aseprite",
  "temp_dir": "/tmp/aseprite-mcp",
  "timeout": 30,
  "log_level": "info"
}
```

Example paths:
- Windows: `D:\\SRC\\aseprite\\build\\bin\\aseprite.exe`
- macOS: `/Applications/Aseprite.app/Contents/MacOS/aseprite`
- Linux: `/usr/bin/aseprite`

## Unit Tests

Run unit tests (requires config file with real Aseprite):
```bash
go test ./...
```

Run with verbose output:
```bash
go test -v ./...
```

## Integration Tests

Run integration tests (requires config file with real Aseprite):
```bash
go test -tags=integration ./...
```

Run integration tests with verbose output:
```bash
go test -tags=integration -v ./pkg/aseprite
```

## Coverage

Generate coverage report:
```bash
make test-coverage
open coverage.html
```

## Test Structure

- **Unit tests**: `*_test.go` files (no build tags)
  - Test pure Go logic, Lua script generation, string escaping
  - Do NOT mock Aseprite - use real executable

- **Integration tests**: `integration_test.go` files with `//go:build integration`
  - Test actual Aseprite execution
  - Create real sprites, draw pixels, export images
  - Verify file I/O with real Aseprite binary

## Manual Testing

Test server manually:
```bash
go run ./cmd/aseprite-mcp
```

## Testing Philosophy

**No Mocks, No Fakes, No Stubs**

All tests use real Aseprite executable. This ensures:
- Tests reflect actual Aseprite behavior
- Changes in Aseprite API are caught immediately
- Integration issues are discovered during development
- Confidence that the server works with real Aseprite

## Troubleshooting

**Error: "config file not found"**
- Create `~/.config/aseprite-mcp/config.json` with valid `aseprite_path`

**Error: "aseprite executable not found"**
- Verify the path in config.json points to a real Aseprite binary
- On Windows, use double backslashes: `D:\\path\\to\\aseprite.exe`

**Tests timeout**
- Increase `timeout` value in config.json (value in nanoseconds)
- Default is 30000000000 (30 seconds)