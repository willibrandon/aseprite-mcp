# Product Requirements Document: Aseprite MCP Server (Go)

**Version:** 1.0
**Status:** Draft
**Author:** Brandon Williams
**Last Updated:** 2025-09-29

---

## Executive Summary

The Aseprite MCP Server is a production-ready Model Context Protocol server that exposes Aseprite's pixel art and animation capabilities to AI assistants and other MCP clients. Built in Go using the official MCP SDK, it enables programmatic creation and manipulation of sprites, animations, and pixel art through a clean, type-safe API.

### Goals
- Enable AI assistants to create professional pixel art and animations
- Provide comprehensive coverage of Aseprite's Lua API (80%+ of common operations)
- Deliver production-quality reliability, performance, and error handling
- Support both standalone CLI and programmatic integration patterns

### Non-Goals
- HTTP/SSE transport support (future enhancement)
- Web-based UI or integrations
- Real-time collaborative editing
- Aseprite GUI automation (focus on batch/headless mode only)

---

## 1. Product Overview

### 1.1 Target Users

**Primary:**
- AI coding assistants (Claude, GPT, etc.) integrated via MCP
- Developers building pixel art generation tools
- Game developers automating sprite workflows

**Secondary:**
- Artists using AI-assisted pixel art creation
- Studios with automated asset pipelines

### 1.2 Value Proposition

- **For AI Assistants:** Direct access to professional pixel art tools without custom integrations
- **For Developers:** Type-safe Go API with comprehensive error handling and testing
- **For Artists:** AI-powered sprite generation and manipulation via natural language

### 1.3 Success Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| API Coverage | 80%+ of common Aseprite operations | Tool count vs. Lua API surface area |
| Reliability | 99%+ success rate on valid operations | Error rate in production usage |
| Performance | <1s response time for basic operations | P95 latency monitoring |
| Usability | AI can generate sprites without errors | Success rate on test prompts |
| Adoption | Integration with 2+ major AI assistants | Active installations |

---

## 2. Functional Requirements

### 2.1 Core Capabilities

#### 2.1.1 Canvas & Sprite Management

**REQ-CANVAS-001: Create New Sprite**
- **Priority:** P0 (Must Have)
- **Description:** Create a new Aseprite sprite with specified dimensions and color mode
- **Inputs:**
  - `width` (int, 1-65535): Canvas width in pixels
  - `height` (int, 1-65535): Canvas height in pixels
  - `color_mode` (enum: "rgb", "grayscale", "indexed"): Sprite color mode
- **Outputs:**
  - `file_path` (string): Absolute path to created .aseprite file
- **Error Cases:**
  - Invalid dimensions (out of range)
  - Disk write failure
  - Aseprite execution failure

**REQ-CANVAS-002: Add Layer**
- **Priority:** P0 (Must Have)
- **Description:** Add a new layer to an existing sprite
- **Inputs:**
  - `sprite_path` (string): Path to .aseprite file
  - `layer_name` (string): Name for the new layer
  - `layer_type` (enum: "normal", "group", "tilemap"): Layer type
- **Outputs:**
  - `success` (bool): Operation success status
  - `layer_index` (int): Index of created layer
- **Error Cases:**
  - Sprite file not found
  - Invalid layer name (empty, duplicate)
  - Aseprite execution failure

**REQ-CANVAS-003: Add Frame**
- **Priority:** P0 (Must Have)
- **Description:** Add a new frame to the sprite timeline
- **Inputs:**
  - `sprite_path` (string): Path to .aseprite file
  - `duration_ms` (int, 1-65535): Frame duration in milliseconds
  - `insert_after` (int, optional): Frame index to insert after
- **Outputs:**
  - `frame_number` (int): Index of created frame
- **Error Cases:**
  - Sprite file not found
  - Invalid frame index

**REQ-CANVAS-004: Delete Layer**
- **Priority:** P1 (Should Have)
- **Inputs:** `sprite_path`, `layer_name`
- **Outputs:** `success`
- **Error Cases:** Layer not found, last layer deletion attempt

**REQ-CANVAS-005: Delete Frame**
- **Priority:** P1 (Should Have)
- **Inputs:** `sprite_path`, `frame_number`
- **Outputs:** `success`
- **Error Cases:** Frame not found, last frame deletion attempt

**REQ-CANVAS-006: Get Sprite Info**
- **Priority:** P0 (Must Have)
- **Description:** Retrieve metadata about a sprite
- **Inputs:** `sprite_path`
- **Outputs:**
  - `width`, `height` (int): Sprite dimensions
  - `color_mode` (string): Color mode
  - `frame_count` (int): Number of frames
  - `layer_count` (int): Number of layers
  - `layers` ([]string): Layer names
- **Error Cases:** Sprite file not found, corrupted file

#### 2.1.2 Drawing Primitives

**REQ-DRAW-001: Draw Pixels**
- **Priority:** P0 (Must Have)
- **Description:** Draw individual pixels at specified coordinates
- **Inputs:**
  - `sprite_path` (string)
  - `layer_name` (string)
  - `frame_number` (int, 1-based)
  - `pixels` (array): `[{x: int, y: int, color: string}]`
    - `color` format: "#RRGGBB" or "#RRGGBBAA"
- **Outputs:** `pixels_drawn` (int)
- **Error Cases:**
  - Coordinates out of bounds
  - Invalid color format
  - Layer/frame not found
- **Performance:** Must handle 10,000+ pixels in <2s

**REQ-DRAW-002: Draw Line**
- **Priority:** P0 (Must Have)
- **Inputs:**
  - `sprite_path`, `layer_name`, `frame_number`
  - `x1`, `y1`, `x2`, `y2` (int): Start and end points
  - `color` (string): Hex color
  - `thickness` (int, 1-100): Line thickness in pixels
- **Outputs:** `success`
- **Error Cases:** Coordinates out of bounds, invalid thickness

**REQ-DRAW-003: Draw Rectangle**
- **Priority:** P0 (Must Have)
- **Inputs:**
  - `sprite_path`, `layer_name`, `frame_number`
  - `x`, `y`, `width`, `height` (int)
  - `color` (string)
  - `filled` (bool): Fill interior or outline only
- **Outputs:** `success`
- **Error Cases:** Invalid dimensions, coordinates out of bounds

**REQ-DRAW-004: Draw Circle**
- **Priority:** P0 (Must Have)
- **Inputs:**
  - `sprite_path`, `layer_name`, `frame_number`
  - `center_x`, `center_y`, `radius` (int)
  - `color` (string)
  - `filled` (bool)
- **Outputs:** `success`

**REQ-DRAW-005: Fill Area**
- **Priority:** P0 (Must Have)
- **Description:** Flood fill from a starting point (paint bucket tool)
- **Inputs:**
  - `sprite_path`, `layer_name`, `frame_number`
  - `x`, `y` (int): Starting point
  - `color` (string)
  - `tolerance` (int, 0-255, optional): Color matching tolerance
- **Outputs:** `success`

**REQ-DRAW-006: Draw Contour**
- **Priority:** P1 (Should Have)
- **Description:** Draw a multi-point polyline or polygon
- **Inputs:**
  - `sprite_path`, `layer_name`, `frame_number`
  - `points` (array): `[{x: int, y: int}]`
  - `color` (string)
  - `closed` (bool): Connect last point to first
  - `thickness` (int)
- **Outputs:** `success`

#### 2.1.3 Animation Tools

**REQ-ANIM-001: Set Frame Duration**
- **Priority:** P0 (Must Have)
- **Inputs:**
  - `sprite_path`
  - `frame_number` (int)
  - `duration_ms` (int, 1-65535)
- **Outputs:** `success`

**REQ-ANIM-002: Create Animation Tag**
- **Priority:** P0 (Must Have)
- **Description:** Define a named animation sequence
- **Inputs:**
  - `sprite_path`
  - `tag_name` (string): Name for the tag
  - `from_frame`, `to_frame` (int): Frame range (inclusive)
  - `direction` (enum: "forward", "reverse", "pingpong"): Playback direction
  - `repeat` (int, optional): Number of times to repeat (0 = infinite)
- **Outputs:** `success`
- **Error Cases:** Invalid frame range, duplicate tag name

**REQ-ANIM-003: Duplicate Frame**
- **Priority:** P1 (Should Have)
- **Inputs:**
  - `sprite_path`
  - `source_frame` (int)
  - `insert_after` (int, optional)
- **Outputs:** `new_frame_number` (int)

**REQ-ANIM-004: Link Cel**
- **Priority:** P1 (Should Have)
- **Description:** Create a linked cel (reuse image from another frame)
- **Inputs:**
  - `sprite_path`
  - `layer_name` (string)
  - `source_frame`, `target_frame` (int)
- **Outputs:** `success`
- **Error Cases:** Source cel doesn't exist

**REQ-ANIM-005: Delete Tag**
- **Priority:** P2 (Nice to Have)
- **Inputs:** `sprite_path`, `tag_name`
- **Outputs:** `success`

#### 2.1.4 Selection Tools

**REQ-SEL-001: Select Rectangle**
- **Priority:** P1 (Should Have)
- **Inputs:**
  - `sprite_path`
  - `x`, `y`, `width`, `height` (int)
  - `mode` (enum: "replace", "add", "subtract", "intersect")
- **Outputs:** `success`

**REQ-SEL-002: Select Ellipse**
- **Priority:** P1 (Should Have)
- **Inputs:** Same as REQ-SEL-001
- **Outputs:** `success`

**REQ-SEL-003: Select All**
- **Priority:** P1 (Should Have)
- **Inputs:** `sprite_path`
- **Outputs:** `success`

**REQ-SEL-004: Deselect**
- **Priority:** P1 (Should Have)
- **Inputs:** `sprite_path`
- **Outputs:** `success`

**REQ-SEL-005: Move Selection**
- **Priority:** P1 (Should Have)
- **Description:** Translate the current selection
- **Inputs:**
  - `sprite_path`
  - `dx`, `dy` (int): Translation offset in pixels
- **Outputs:** `success`

**REQ-SEL-006: Cut Selection**
- **Priority:** P1 (Should Have)
- **Inputs:** `sprite_path`, `layer_name`, `frame_number`
- **Outputs:** `success`

**REQ-SEL-007: Copy Selection**
- **Priority:** P1 (Should Have)
- **Inputs:** `sprite_path`
- **Outputs:** `success`

**REQ-SEL-008: Paste Clipboard**
- **Priority:** P1 (Should Have)
- **Inputs:**
  - `sprite_path`, `layer_name`, `frame_number`
  - `x`, `y` (int, optional): Paste position
- **Outputs:** `success`

#### 2.1.5 Palette Tools

**REQ-PAL-001: Get Palette**
- **Priority:** P1 (Should Have)
- **Description:** Retrieve the current color palette
- **Inputs:** `sprite_path`
- **Outputs:**
  - `colors` (array): `["#RRGGBB", ...]`
  - `size` (int): Number of colors
- **Error Cases:** Sprite file not found

**REQ-PAL-002: Set Palette**
- **Priority:** P1 (Should Have)
- **Inputs:**
  - `sprite_path`
  - `colors` (array): `["#RRGGBB", ...]` (max 256 colors)
- **Outputs:** `success`
- **Error Cases:** Invalid color format, too many colors

**REQ-PAL-003: Set Palette Color**
- **Priority:** P1 (Should Have)
- **Inputs:**
  - `sprite_path`
  - `index` (int, 0-255)
  - `color` (string)
- **Outputs:** `success`

**REQ-PAL-004: Add Palette Color**
- **Priority:** P2 (Nice to Have)
- **Inputs:** `sprite_path`, `color`
- **Outputs:** `color_index` (int)

**REQ-PAL-005: Sort Palette**
- **Priority:** P2 (Nice to Have)
- **Inputs:**
  - `sprite_path`
  - `method` (enum: "hue", "saturation", "brightness", "luminance")
  - `ascending` (bool)
- **Outputs:** `success`

#### 2.1.6 Transform & Filter Tools

**REQ-XFORM-001: Flip Sprite**
- **Priority:** P0 (Must Have)
- **Inputs:**
  - `sprite_path`
  - `direction` (enum: "horizontal", "vertical")
  - `target` (enum: "sprite", "layer", "cel"): What to flip
- **Outputs:** `success`

**REQ-XFORM-002: Rotate Sprite**
- **Priority:** P0 (Must Have)
- **Inputs:**
  - `sprite_path`
  - `angle` (enum: "90", "180", "270"): Rotation angle
  - `target` (enum: "sprite", "layer", "cel")
- **Outputs:** `success`

**REQ-XFORM-003: Scale Sprite**
- **Priority:** P0 (Must Have)
- **Inputs:**
  - `sprite_path`
  - `scale_x`, `scale_y` (float, 0.01-100.0): Scale factors
  - `algorithm` (enum: "nearest", "bilinear", "rotsprite"): Scaling algorithm
- **Outputs:** `success`, `new_width`, `new_height`
- **Error Cases:** Invalid scale factor (zero, negative)

**REQ-XFORM-004: Crop Sprite**
- **Priority:** P1 (Should Have)
- **Inputs:**
  - `sprite_path`
  - `x`, `y`, `width`, `height` (int): Crop rectangle
- **Outputs:** `success`
- **Error Cases:** Crop area outside bounds, zero dimensions

**REQ-XFORM-005: Resize Canvas**
- **Priority:** P1 (Should Have)
- **Description:** Change canvas size without scaling content
- **Inputs:**
  - `sprite_path`
  - `width`, `height` (int): New canvas dimensions
  - `anchor` (enum: "center", "top_left", "bottom_right", etc.): Content positioning
- **Outputs:** `success`

**REQ-XFORM-006: Apply Outline**
- **Priority:** P2 (Nice to Have)
- **Inputs:**
  - `sprite_path`, `layer_name`, `frame_number`
  - `color` (string)
  - `thickness` (int, 1-10)
- **Outputs:** `success`

#### 2.1.7 Export & Import Tools

**REQ-EXPORT-001: Export Sprite**
- **Priority:** P0 (Must Have)
- **Description:** Export sprite to common image formats
- **Inputs:**
  - `sprite_path`
  - `output_path` (string): Output file path
  - `format` (enum: "png", "gif", "jpg", "bmp"): Export format
  - `frame_number` (int, optional): Specific frame (null = all frames)
  - `scale` (float, optional): Export scale factor
- **Outputs:**
  - `exported_path` (string): Actual output path
  - `file_size` (int): Size in bytes
- **Error Cases:** Unsupported format, write permission denied

**REQ-EXPORT-002: Export Spritesheet**
- **Priority:** P0 (Must Have)
- **Description:** Export animation frames as a spritesheet
- **Inputs:**
  - `sprite_path`
  - `output_path` (string)
  - `layout` (enum: "horizontal", "vertical", "rows", "columns", "packed")
  - `padding` (int, 0-100): Pixels between frames
  - `include_json` (bool): Generate metadata JSON
- **Outputs:**
  - `spritesheet_path` (string)
  - `metadata_path` (string, optional): JSON metadata file
  - `frame_count` (int)
- **Error Cases:** Too many frames for layout

**REQ-EXPORT-003: Import Image as Layer**
- **Priority:** P1 (Should Have)
- **Inputs:**
  - `sprite_path`
  - `image_path` (string): Path to image file
  - `layer_name` (string)
  - `frame_number` (int)
  - `position` (object, optional): `{x: int, y: int}` placement
- **Outputs:** `success`
- **Error Cases:** Image format not supported, dimension mismatch

**REQ-EXPORT-004: Save As**
- **Priority:** P1 (Should Have)
- **Inputs:**
  - `sprite_path`
  - `output_path` (string): New .aseprite file path
- **Outputs:** `success`

### 2.2 Cross-Cutting Requirements

#### 2.2.1 Configuration Management

**REQ-CONFIG-001: Environment Variables**
- `ASEPRITE_PATH`: Path to Aseprite executable (required - must be explicitly configured)
- `ASEPRITE_TEMP_DIR`: Temporary file directory
- `ASEPRITE_TIMEOUT`: Operation timeout in seconds (default: 30)
- `ASEPRITE_LOG_LEVEL`: Logging verbosity (debug, info, warn, error)

**REQ-CONFIG-002: Configuration File**
- Optional JSON config file: `~/.config/aseprite-mcp/config.json`
- Environment variables override config file
- Config file overrides defaults

**REQ-CONFIG-003: Aseprite Path Requirement**
- Aseprite executable path must be explicitly provided via `ASEPRITE_PATH` environment variable or config file
- No automatic discovery or PATH searching
- Common install locations for reference:
  - Windows: `C:\Program Files\Aseprite\Aseprite.exe`
  - macOS: `/Applications/Aseprite.app/Contents/MacOS/aseprite`
  - Linux: `/usr/bin/aseprite`, `/usr/local/bin/aseprite`

#### 2.2.2 Error Handling

**REQ-ERROR-001: Error Categories**
- Configuration errors (ERRCODE 1xxx): Missing binary, invalid paths
- Validation errors (ERRCODE 2xxx): Invalid parameters, out of range
- Execution errors (ERRCODE 3xxx): Aseprite command failures
- Timeout errors (ERRCODE 4xxx): Operation exceeded deadline
- File errors (ERRCODE 5xxx): Not found, permission denied

**REQ-ERROR-002: Error Response Format**
```json
{
  "error": {
    "code": "ERR_3001",
    "message": "Aseprite execution failed: layer not found",
    "details": {
      "stderr": "Error: layer 'Background' not found",
      "command": "aseprite --batch sprite.aseprite --script tmp.lua"
    }
  }
}
```

**REQ-ERROR-003: Error Recovery**
- Automatic retry on transient failures (max 3 attempts)
- Clean up temporary files on all error paths
- Log full context for debugging

#### 2.2.3 Logging

**REQ-LOG-001: Structured Logging**
- Use structured logging library (e.g., `mtlog`)
- Include: timestamp, level, operation, duration, error
- Redact sensitive paths in logs

**REQ-LOG-002: Log Levels**
- **DEBUG:** Lua script content, full command arguments
- **INFO:** Operation start/complete, file paths
- **WARN:** Retries, deprecated API usage
- **ERROR:** Execution failures, validation errors

**REQ-LOG-003: Performance Logging**
- Log operation duration for all tools
- Log P50, P95, P99 latencies on shutdown

#### 2.2.4 Security

**REQ-SEC-001: Path Validation**
- Validate all file paths are absolute or relative to working directory
- Prevent directory traversal attacks (`../`, etc.)
- Reject paths containing null bytes

**REQ-SEC-002: Lua Script Safety**
- Escape all user input in generated Lua scripts
- Use parameterized script generation (no string concatenation)
- Limit script execution time (timeout)

**REQ-SEC-003: Resource Limits**
- Max sprite dimensions: 65535x65535 (Aseprite limit)
- Max pixels per draw operation: 100,000
- Max file size: 1 GB
- Max concurrent operations: 10

**REQ-SEC-004: Temporary File Security**
- Create temp files with restricted permissions (0600)
- Clean up temp files on exit (defer cleanup)
- Use cryptographically random temp filenames

---

## 3. Non-Functional Requirements

### 3.1 Performance

**REQ-PERF-001: Latency Targets**
| Operation | P95 Latency | P99 Latency |
|-----------|-------------|-------------|
| Create canvas | 500ms | 1s |
| Draw primitives (1-100 pixels) | 300ms | 500ms |
| Draw primitives (1K-10K pixels) | 1s | 2s |
| Export sprite | 1s | 3s |
| Transform operations | 2s | 5s |

**REQ-PERF-002: Throughput**
- Support 10 concurrent operations per server instance
- Process 100+ operations per minute (single instance)

**REQ-PERF-003: Resource Usage**
- Memory: <100 MB baseline, <1 GB during operations
- CPU: <50% single core utilization during idle
- Disk: Clean up temp files within 1 minute of completion

### 3.2 Reliability

**REQ-REL-001: Availability**
- 99.9% uptime for stdio transport
- Graceful degradation on Aseprite failures

**REQ-REL-002: Data Integrity**
- All file writes must be atomic (write to temp, then rename)
- Verify sprite file integrity after operations
- Never corrupt existing sprites on failure

**REQ-REL-003: Fault Tolerance**
- Recover from Aseprite crashes without server restart
- Handle missing dependencies gracefully
- Provide clear error messages for fixable issues

### 3.3 Scalability

**REQ-SCALE-001: Horizontal Scaling**
- Stateless server design (no shared state between operations)
- Support running multiple server instances in parallel
- Safe for concurrent access to different sprite files

**REQ-SCALE-002: Large Sprites**
- Support sprites up to 65535x65535 pixels (Aseprite limit)
- Graceful degradation for operations on very large sprites
- Memory-efficient processing (streaming where possible)

### 3.4 Maintainability

**REQ-MAINT-001: Code Quality**
- 80%+ test coverage (unit + integration)
- All public APIs documented with GoDoc
- Follow Go standard project layout
- Pass `go vet`, `golangci-lint` with zero warnings

**REQ-MAINT-002: Versioning**
- Semantic versioning (MAJOR.MINOR.PATCH)
- Document breaking changes in CHANGELOG.md
- Maintain backward compatibility within major versions

**REQ-MAINT-003: Observability**
- Expose Prometheus metrics endpoint (optional)
- Health check endpoint
- Structured logging for all operations

### 3.5 Compatibility

**REQ-COMPAT-001: Aseprite Versions**
- Minimum supported: Aseprite v1.3.0
- Recommended: Aseprite v1.3.10+
- Detect version and warn on incompatibility

**REQ-COMPAT-002: Go Version**
- Minimum Go 1.23
- Use only stdlib and vetted third-party dependencies

**REQ-COMPAT-003: Operating Systems**
- Windows 10/11
- macOS 12+ (both Intel and Apple Silicon)
- Linux (Ubuntu 20.04+, other distros best-effort)

**REQ-COMPAT-004: MCP Protocol**
- MCP specification version: 2024-11-05
- Official Go SDK version: latest stable

---

## 4. Architecture & Design

### 4.1 System Architecture

```
┌─────────────────────────────────────────────────────┐
│                  MCP Client                         │
│          (Claude Desktop, Custom Clients)           │
└──────────────────┬──────────────────────────────────┘
                   │ MCP Protocol (Stdio)
                   │
┌──────────────────▼──────────────────────────────────┐
│              Aseprite MCP Server (Go)               │
│  ┌─────────────────────────────────────────────┐   │
│  │           MCP Server Core                   │   │
│  │  - Tool Registry                            │   │
│  │  - Request Validation                       │   │
│  │  - Session Management                       │   │
│  └──────────────────┬──────────────────────────┘   │
│                     │                               │
│  ┌──────────────────▼──────────────────────────┐   │
│  │          Tool Implementation                │   │
│  │  - Canvas Tools    - Animation Tools        │   │
│  │  - Drawing Tools   - Selection Tools        │   │
│  │  - Export Tools    - Transform Tools        │   │
│  └──────────────────┬──────────────────────────┘   │
│                     │                               │
│  ┌──────────────────▼──────────────────────────┐   │
│  │        Aseprite Client Abstraction          │   │
│  │  - Command Executor                         │   │
│  │  - Lua Script Generator                     │   │
│  │  - Temp File Manager                        │   │
│  └──────────────────┬──────────────────────────┘   │
└────────────────────┬────────────────────────────────┘
                     │ Process Execution
                     │
┌────────────────────▼────────────────────────────────┐
│            Aseprite (Batch Mode)                    │
│  - Lua Script Execution                             │
│  - CLI Commands                                     │
│  - Sprite File I/O                                  │
└─────────────────────────────────────────────────────┘
```

### 4.2 Component Design

#### 4.2.1 Directory Structure

```
aseprite-mcp-go/
├── cmd/
│   └── aseprite-mcp/
│       └── main.go                  # Entry point
├── pkg/
│   ├── aseprite/
│   │   ├── client.go               # Aseprite command executor
│   │   ├── lua.go                  # Lua script generation
│   │   ├── types.go                # Domain types
│   │   ├── client_test.go
│   │   └── lua_test.go
│   ├── tools/
│   │   ├── canvas.go               # Canvas management tools
│   │   ├── drawing.go              # Drawing primitive tools
│   │   ├── animation.go            # Animation tools
│   │   ├── selection.go            # Selection tools
│   │   ├── palette.go              # Palette tools
│   │   ├── transform.go            # Transform/filter tools
│   │   ├── export.go               # Export/import tools
│   │   └── *_test.go               # Test files
│   ├── config/
│   │   ├── config.go               # Configuration management
│   │   └── config_test.go
│   └── server/
│       ├── server.go               # MCP server setup
│       └── server_test.go
├── internal/
│   └── testutil/                   # Test helpers
│       ├── mock_aseprite.go        # Mock Aseprite for testing
│       └── fixtures/               # Test sprite files
├── scripts/
│   ├── install.sh                  # Installation script
│   └── release.sh                  # Release automation
├── docs/
│   ├── PRD.md                      # This document
│   ├── DESIGN.md                   # Architecture design
│   ├── API.md                      # Tool API reference
│   └── CONTRIBUTING.md             # Contribution guidelines
├── examples/
│   ├── client/
│   │   └── main.go                 # Example MCP client
│   └── sprites/                    # Example outputs
├── .github/
│   └── workflows/
│       ├── test.yml                # CI: tests
│       ├── lint.yml                # CI: linting
│       └── release.yml             # CI: releases
├── go.mod
├── go.sum
├── Makefile                        # Build automation
├── README.md                       # Project overview
├── LICENSE                         # MIT license
└── CHANGELOG.md                    # Version history
```

#### 4.2.2 Core Interfaces

```go
// Client interface for Aseprite execution
type Client interface {
    ExecuteLua(ctx context.Context, script string, spritePath string) (string, error)
    ExecuteCommand(ctx context.Context, args []string) (string, error)
    GetVersion(ctx context.Context) (string, error)
}

// LuaGenerator interface for script generation
type LuaGenerator interface {
    CreateCanvas(width, height int, colorMode string) string
    DrawPixels(layer string, frame int, pixels []Pixel) string
    ExportSprite(outputPath string, format string, frame int) string
    // ... more script generators
}

// Config interface for configuration management
type Config interface {
    AsepritePath() string
    TempDir() string
    Timeout() time.Duration
    Validate() error
}
```

### 4.3 Data Models

#### 4.3.1 Core Types

```go
// Color represents an RGBA color
type Color struct {
    R uint8 `json:"r"`
    G uint8 `json:"g"`
    B uint8 `json:"b"`
    A uint8 `json:"a"`
}

// FromHex parses "#RRGGBB" or "#RRGGBBAA"
func (c *Color) FromHex(hex string) error

// ToHex formats as "#RRGGBBAA"
func (c Color) ToHex() string

// Point represents a 2D coordinate
type Point struct {
    X int `json:"x"`
    Y int `json:"y"`
}

// Rectangle represents a rectangular region
type Rectangle struct {
    X      int `json:"x"`
    Y      int `json:"y"`
    Width  int `json:"width"`
    Height int `json:"height"`
}

// Pixel represents a single pixel with color
type Pixel struct {
    Point
    Color Color `json:"color"`
}

// SpriteInfo contains sprite metadata
type SpriteInfo struct {
    Width      int      `json:"width"`
    Height     int      `json:"height"`
    ColorMode  string   `json:"color_mode"`
    FrameCount int      `json:"frame_count"`
    LayerCount int      `json:"layer_count"`
    Layers     []string `json:"layers"`
}
```

### 4.4 Tool Implementation Pattern

```go
// Example: Create Canvas Tool
type CreateCanvasInput struct {
    Width     int    `json:"width" jsonschema:"required,minimum=1,maximum=65535,description=Canvas width in pixels"`
    Height    int    `json:"height" jsonschema:"required,minimum=1,maximum=65535,description=Canvas height in pixels"`
    ColorMode string `json:"color_mode" jsonschema:"enum=rgb,enum=grayscale,enum=indexed,description=Color mode"`
}

type CreateCanvasOutput struct {
    FilePath string `json:"file_path" jsonschema:"description=Path to created sprite file"`
}

func registerCanvasTools(server *mcp.Server, client aseprite.Client, cfg config.Config) {
    mcp.AddTool(
        server,
        &mcp.Tool{
            Name:        "create_canvas",
            Description: "Create a new Aseprite sprite with specified dimensions and color mode",
        },
        func(ctx context.Context, req *mcp.CallToolRequest, input CreateCanvasInput) (
            *mcp.CallToolResult,
            *CreateCanvasOutput,
            error,
        ) {
            // Generate Lua script
            script := lua.CreateCanvas(input.Width, input.Height, input.ColorMode)

            // Execute
            output, err := client.ExecuteLua(ctx, script, "")
            if err != nil {
                return nil, nil, fmt.Errorf("failed to create canvas: %w", err)
            }

            // Parse output (file path)
            filePath := strings.TrimSpace(output)

            return nil, &CreateCanvasOutput{FilePath: filePath}, nil
        },
    )
}
```

---

## 5. Testing Strategy

### 5.1 Unit Tests

**REQ-TEST-001: Lua Script Generation**
- Test all Lua generator functions
- Verify correct escaping of user input
- Validate generated Lua syntax
- Coverage target: 90%+

**REQ-TEST-002: Input Validation**
- Test JSON schema validation for all tools
- Verify boundary conditions (min/max values)
- Test invalid input rejection
- Coverage target: 100% of validation paths

**REQ-TEST-003: Configuration**
- Test config loading from file, env, defaults
- Verify precedence order
- Test validation errors

### 5.2 Integration Tests

**REQ-TEST-004: Mock Aseprite**
- Create mock Aseprite executable for CI
- Mock responses for all script patterns
- Verify command-line argument passing

**REQ-TEST-005: End-to-End Workflows**
- Create sprite → draw → export workflow
- Animation creation workflow
- Selection + transform workflow
- Run against real Aseprite in CI (optional)

**REQ-TEST-006: Error Handling**
- Test all error paths
- Verify error messages are actionable
- Test timeout handling
- Test cleanup on failure

### 5.3 Performance Tests

**REQ-TEST-007: Benchmarks**
- Benchmark all critical operations
- Track performance regressions in CI
- Test with large sprites (10000x10000)
- Test bulk operations (10K+ pixels)

### 5.4 Compatibility Tests

**REQ-TEST-008: Aseprite Versions**
- Test against Aseprite 1.3.0, 1.3.5, 1.3.10 (latest stable)
- Document version-specific behavior
- Automated version detection tests

**REQ-TEST-009: Operating Systems**
- Run full test suite on Windows, macOS, Linux
- CI matrix: [windows-latest, macos-latest, ubuntu-latest]

---

## 6. Deployment & Operations

### 6.1 Installation

**REQ-DEPLOY-001: Binary Distribution**
- Pre-built binaries for Windows, macOS (Intel + ARM), Linux (amd64, arm64)
- GitHub Releases with checksums
- Archive format: `.tar.gz` (Unix), `.zip` (Windows)

**REQ-DEPLOY-002: Installation Script**
```bash
curl -fsSL https://raw.githubusercontent.com/user/aseprite-mcp-go/main/scripts/install.sh | sh
```
- Detect OS and architecture
- Download appropriate binary
- Install to `~/.local/bin` or `/usr/local/bin`
- Set executable permissions

**REQ-DEPLOY-003: MCP Client Configuration**
Example Claude Desktop config:
```json
{
  "mcpServers": {
    "aseprite": {
      "command": "/path/to/aseprite-mcp",
      "env": {
        "ASEPRITE_PATH": "/Applications/Aseprite.app/Contents/MacOS/aseprite"
      }
    }
  }
}
```

### 6.2 Monitoring

**REQ-OPS-001: Health Checks**
- `--health` flag: verify Aseprite availability and version
- Exit codes: 0 (healthy), 1 (unhealthy)

**REQ-OPS-002: Metrics (Optional)**
- Operation count by tool
- Success/failure rates
- Latency percentiles
- Resource usage (memory, CPU)

**REQ-OPS-003: Diagnostics**
- `--debug` flag: verbose logging
- `--version` flag: show version info
- Log file location: `~/.config/aseprite-mcp/logs/`

### 6.3 Upgrades

**REQ-OPS-004: Backward Compatibility**
- Maintain tool API compatibility within major versions
- Deprecation warnings for 2 minor versions before removal
- Migration guide for breaking changes

**REQ-OPS-005: Rollback**
- Support running multiple versions side-by-side
- No database migrations required (stateless)

---

## 7. Documentation

### 7.1 User Documentation

**REQ-DOC-001: README.md**
- Quick start guide
- Installation instructions for all platforms
- Basic usage examples
- Link to full documentation

**REQ-DOC-002: API Reference (docs/API.md)**
- Complete tool reference
- Input/output schemas for each tool
- Example requests and responses
- Error code reference

**REQ-DOC-003: Examples**
- Working code examples in `examples/`
- Common workflows (sprite creation, animation, export)
- Integration with popular AI assistants

### 7.2 Developer Documentation

**REQ-DOC-004: Architecture (docs/DESIGN.md)**
- System architecture diagrams
- Component design
- Data flow
- Extension points

**REQ-DOC-005: Contributing Guide**
- Development setup instructions
- Coding standards
- Testing requirements
- PR process

**REQ-DOC-006: GoDoc**
- All exported types, functions, and packages documented
- Code examples in documentation
- Link to godoc.org

---

## 8. Release Plan

### 8.1 Milestones

#### Phase 1: MVP (v0.1.0) - Target: Week 4
**Scope:**
- Canvas management (create, add layer/frame)
- Basic drawing (pixels, line, rectangle, circle, fill)
- Export (single frame PNG)
- Stdio transport only
- Configuration via environment variables

**Exit Criteria:**
- All P0 requirements implemented
- 70%+ test coverage
- Runs on Windows, macOS, Linux
- Example client works

#### Phase 2: Animation & Selection (v0.2.0) - Target: Week 8
**Scope:**
- Frame duration, tags, linked cels
- Selection tools (rectangle, ellipse, all, deselect)
- Cut/copy/paste
- Export GIF and spritesheet

**Exit Criteria:**
- All P0 + P1 animation requirements
- All P1 selection requirements
- 80%+ test coverage
- Performance benchmarks meet targets

#### Phase 3: Advanced Features (v0.3.0) - Target: Week 12
**Scope:**
- Palette management
- Transform operations (flip, rotate, scale, crop)
- Apply outline filter
- Configuration file support

**Exit Criteria:**
- All P1 + P2 requirements
- 85%+ test coverage
- Documented integration with 2+ AI assistants

#### Phase 4: Production Release (v1.0.0) - Target: Week 16
**Scope:**
- Performance optimization
- Comprehensive error handling
- Production hardening
- Full documentation

**Exit Criteria:**
- All functional requirements met
- All non-functional requirements met
- 90%+ test coverage
- User documentation complete
- Release automation in place

### 8.2 Success Metrics Tracking

| Metric | Week 4 | Week 8 | Week 12 | Week 16 |
|--------|--------|--------|---------|---------|
| API Coverage | 40% | 60% | 80% | 95% |
| Test Coverage | 70% | 80% | 85% | 90% |
| P95 Latency | <2s | <1s | <800ms | <500ms |
| Active Users | 5 | 20 | 50 | 100 |

---

## 9. Risk Management

### 9.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Aseprite Lua API changes | Medium | High | Pin to specific Aseprite versions; version detection |
| Performance issues with large sprites | Medium | Medium | Implement streaming; set resource limits |
| Platform-specific bugs | Low | Medium | Comprehensive OS testing; CI matrix |
| MCP protocol changes | Low | High | Track MCP SDK releases; maintain compatibility layer |

### 9.2 Operational Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Aseprite not installed | High | High | Clear error messages; installation guide |
| Incompatible Aseprite version | Medium | Medium | Version detection; compatibility warnings |
| Disk space exhaustion | Low | Medium | Temp file cleanup; disk space checks |
| Security vulnerabilities | Low | High | Regular dependency updates; security audits |

---

## 10. Appendices

### Appendix A: Glossary

- **MCP:** Model Context Protocol - standardized protocol for AI assistant integrations
- **Cel:** A single image in a layer at a specific frame
- **Spritesheet:** Grid layout of animation frames in a single image
- **Batch mode:** Aseprite's headless execution mode for scripting
- **Linked cel:** Cel that references another cel's image data

### Appendix B: References

- [Aseprite Lua API](https://github.com/aseprite/api)
- [MCP Specification](https://modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Aseprite CLI Documentation](https://www.aseprite.org/docs/cli/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

### Appendix C: Change Log

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-09-29 | Initial PRD |

---

**Document Status:** Draft
**Next Review:** After architecture approval
**Approval Required From:** Technical Lead, Product Owner