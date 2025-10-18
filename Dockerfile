# Multi-stage build for pixel-mcp server

# Stage 1: Build the Go binary
FROM golang:1.24.1-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pixel-mcp ./cmd/pixel-mcp

# Stage 2: Runtime image with Aseprite
FROM ubuntu:22.04

LABEL org.opencontainers.image.source=https://github.com/willibrandon/pixel-mcp
LABEL org.opencontainers.image.description="MCP server for Aseprite pixel art and animation capabilities"
LABEL org.opencontainers.image.licenses=MIT
LABEL io.modelcontextprotocol.server.name="io.github.willibrandon/pixel-mcp"

# Install runtime dependencies for Aseprite
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    libx11-6 \
    libxcursor1 \
    libxi6 \
    libxrandr2 \
    libgl1 \
    libfontconfig1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the Go binary from builder
COPY --from=builder /build/pixel-mcp /usr/local/bin/pixel-mcp

# Create necessary directories
RUN mkdir -p /tmp/pixel-mcp /root/.config/pixel-mcp

# Note: Aseprite binary must be provided via:
# 1. Build arg: --build-arg ASEPRITE_SOURCE=<path-to-aseprite>
# 2. Volume mount: -v /path/to/aseprite:/usr/local/bin/aseprite
# 3. Or copy from CI image (see Dockerfile.with-aseprite)

# Default environment variables (can be overridden)
ENV ASEPRITE_PATH=/usr/local/bin/aseprite
ENV TEMP_DIR=/tmp/pixel-mcp
ENV TIMEOUT=30
ENV LOG_LEVEL=info

# Generate default config file at runtime if not provided
RUN printf '{\n  "aseprite_path": "%s",\n  "temp_dir": "%s",\n  "timeout": %d,\n  "log_level": "%s"\n}\n' \
    "$ASEPRITE_PATH" "$TEMP_DIR" "$TIMEOUT" "$LOG_LEVEL" \
    > /root/.config/pixel-mcp/config.json

# MCP servers communicate over stdio
ENTRYPOINT ["/usr/local/bin/pixel-mcp"]
