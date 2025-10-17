# Testing Guide

## Prerequisites

ALL tests require a real Aseprite installation with configuration file.

Create `~/.config/pixel-mcp/config.json`:

**macOS:**
```json
{
  "aseprite_path": "/Applications/Aseprite.app/Contents/MacOS/aseprite",
  "temp_dir": "/tmp/pixel-mcp",
  "timeout": 30,
  "log_level": "info",
  "log_file": "",
  "enable_timing": false
}
```

**Linux:**
```json
{
  "aseprite_path": "/usr/bin/aseprite",
  "temp_dir": "/tmp/pixel-mcp",
  "timeout": 30,
  "log_level": "info",
  "log_file": "",
  "enable_timing": false
}
```

**Windows:**
```json
{
  "aseprite_path": "C:\\Program Files\\Aseprite\\aseprite.exe",
  "temp_dir": "C:\\Temp\\pixel-mcp",
  "timeout": 30,
  "log_level": "info",
  "log_file": "",
  "enable_timing": false
}
```

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

## Docker Testing

Run tests in Docker CI environment (includes Aseprite):
```bash
make docker-test-all
```

This runs both unit and integration tests in the CI container with Aseprite pre-built.

## Manual Testing

Test server manually:
```bash
go run ./cmd/pixel-mcp
```

Test server via Docker:
```bash
make docker-run-full
```

## Testing Philosophy

All tests use a real Aseprite executable to ensure:
- Tests accurately reflect Aseprite's actual behavior
- Changes in Aseprite's API are detected immediately
- Integration issues are discovered during development
- High confidence that the server works correctly in production

## Troubleshooting

**Error: "config file not found"**
- Create `~/.config/pixel-mcp/config.json` with valid `aseprite_path`

**Error: "aseprite executable not found"**
- Verify the path in config.json points to a real Aseprite binary
- On Windows, use double backslashes: `D:\\path\\to\\aseprite.exe`

**Tests timeout**
- Increase `timeout` value in config.json (value in seconds)
- Default is 30 seconds