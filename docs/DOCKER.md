# Docker Deployment Guide

This guide explains how to build and run the aseprite-mcp server using Docker.

## Available Docker Images

### 1. Lightweight Image (`Dockerfile`)
- **Size**: ~300MB
- **Use case**: When you have Aseprite installed locally
- **Requires**: Volume mount of Aseprite binary

### 2. Complete Image (`Dockerfile.with-aseprite`)
- **Size**: ~2GB
- **Use case**: Self-contained deployment with Aseprite included
- **Requires**: Nothing (Aseprite built-in)

## Quick Start

### Option 1: Lightweight Image (Recommended for Development)

```bash
# Build the image
make docker-build

# Run with your local Aseprite (macOS example)
docker run --rm -i \
  -v /Applications/Aseprite.app/Contents/MacOS/aseprite:/usr/local/bin/aseprite:ro \
  aseprite-mcp:latest

# Linux example
docker run --rm -i \
  -v /usr/bin/aseprite:/usr/local/bin/aseprite:ro \
  aseprite-mcp:latest
```

### Option 2: Complete Image (Self-Contained)

```bash
# Build the image (takes ~15-20 minutes first time)
make docker-build-full

# Run (no volume mounts needed)
make docker-run-full
```

## Using Docker Compose

The `docker-compose.yml` file provides convenient service definitions:

```bash
# Run lightweight version (edit docker-compose.yml to match your Aseprite path)
docker-compose run aseprite-mcp

# Run complete version with Aseprite included
docker-compose run aseprite-mcp-full
```

## Configuration

### Environment Variables

All Docker images support these environment variables:

- `ASEPRITE_PATH`: Path to Aseprite executable (default: `/usr/local/bin/aseprite`)
- `TEMP_DIR`: Temporary directory for sprite files (default: `/tmp/aseprite-mcp`)
- `TIMEOUT`: Command timeout in seconds (default: `30`)
- `LOG_LEVEL`: Logging verbosity (`debug`, `info`, `warn`, `error`)

Example:

```bash
docker run --rm -i \
  -e ASEPRITE_PATH=/custom/path/aseprite \
  -e LOG_LEVEL=debug \
  aseprite-mcp:latest
```

### Custom Configuration File

Mount a custom configuration file:

```bash
docker run --rm -i \
  -v $(PWD)/config.json:/root/.config/aseprite-mcp/config.json:ro \
  aseprite-mcp:latest
```

## Building Images

### Build Lightweight Image

```bash
# Using Make
make docker-build

# Using Docker directly
docker build -t aseprite-mcp:latest -f Dockerfile .
```

### Build Complete Image

```bash
# Using Make
make docker-build-full

# Using Docker directly
docker build -t aseprite-mcp:full -f Dockerfile.with-aseprite .
```

## Integrating with MCP Clients

### Claude Desktop Configuration

Add to your Claude configuration file (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "aseprite": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-v",
        "/Applications/Aseprite.app/Contents/MacOS/aseprite:/usr/local/bin/aseprite:ro",
        "aseprite-mcp:latest"
      ]
    }
  }
}
```

### Using Complete Image (No Volume Mounts)

```json
{
  "mcpServers": {
    "aseprite": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "aseprite-mcp:full"
      ]
    }
  }
}
```

## Publishing to GitHub Container Registry

```bash
# Tag the image
docker tag aseprite-mcp:latest ghcr.io/willibrandon/aseprite-mcp:latest
docker tag aseprite-mcp:latest ghcr.io/willibrandon/aseprite-mcp:v0.1.0

# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Push
docker push ghcr.io/willibrandon/aseprite-mcp:latest
docker push ghcr.io/willibrandon/aseprite-mcp:v0.1.0
```

## Troubleshooting

### Aseprite Not Found

If you see "aseprite: not found" errors:

1. **Lightweight image**: Verify your volume mount path is correct
2. **Complete image**: Image should work out of the box

Test Aseprite availability:

```bash
docker run --rm -i aseprite-mcp:latest /usr/local/bin/aseprite --version
```

### Permission Errors

Ensure temp directory is writable:

```bash
docker run --rm -i \
  -v $(PWD)/tmp:/tmp/aseprite-mcp \
  aseprite-mcp:latest
```

### Platform Issues

Force specific platform for compatibility:

```bash
docker build --platform linux/amd64 -t aseprite-mcp:latest -f Dockerfile .
```

## Image Sizes

- **Lightweight** (`aseprite-mcp:latest`): ~300MB
- **Complete** (`aseprite-mcp:full`): ~2GB
- **CI Image** (`aseprite-mcp-ci:latest`): ~3GB (includes build tools)

## Security Considerations

- Images run as root by default (MCP servers typically run in isolated environments)
- Aseprite binary volume mounts use `:ro` (read-only) for security
- No network access required for MCP protocol (uses stdio)
- Temp directories are ephemeral by default
