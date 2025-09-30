# Implementation Guide: Aseprite MCP Server (Go)

**Version:** 1.0
**Author:** Brandon Williams
**Last Updated:** 2025-09-29
**Target Audience:** AI Coding Collaborators (Claude Code, etc.)

---

## Overview

This guide provides a step-by-step implementation plan for building the Aseprite MCP Server. Each chunk is designed to be:
- **Self-contained**: Can be implemented and tested independently
- **Verifiable**: Includes manual and automated tests
- **Committable**: Ends with a git commit checkpoint

**Total Estimated Chunks:** 25
**Estimated Timeline:** 12-16 weeks

---

## Prerequisites

Before starting, ensure:
1. Go 1.23+ is installed
2. Aseprite is installed and accessible in PATH or known location
3. Git is installed
4. Working directory: `D:\SRC\aseprite-mcp-go`

---

## Implementation Sequence

### Phase 0: Project Setup (Chunks 1-3)

#### Chunk 1: Initialize Go Module and Project Structure

**Objective:** Create the basic project structure and initialize Go module.

**Tasks:**
1. Initialize Go module:
   ```bash
   cd D:\SRC\aseprite-mcp-go
   go mod init github.com/willibrandon/aseprite-mcp-go
   ```

2. Create directory structure:
   ```bash
   mkdir -p cmd/aseprite-mcp
   mkdir -p pkg/aseprite
   mkdir -p pkg/config
   mkdir -p pkg/server
   mkdir -p pkg/tools
   mkdir -p internal/testutil
   mkdir -p examples/client
   mkdir -p scripts
   ```

3. Create `.gitignore`:
   ```gitignore
   # Binaries
   /bin/
   /dist/
   *.exe
   *.exe~
   *.dll
   *.so
   *.dylib
   aseprite-mcp

   # Test files
   *.test
   *.out
   coverage.txt
   coverage.html

   # Temporary files
   *.tmp
   *.aseprite
   *.png
   *.gif
   /tmp/

   # IDE
   .vscode/
   .idea/
   *.swp
   *.swo
   *~

   # OS
   .DS_Store
   Thumbs.db

   # Config
   config.json
   .env
   ```

4. Create `README.md`:
   ```markdown
   # Aseprite MCP Server (Go)

   A Model Context Protocol (MCP) server that exposes Aseprite's pixel art and animation capabilities to AI assistants.

   ## Status

   🚧 Under active development - not ready for production use

   ## Requirements

   - Go 1.23+
   - Aseprite 1.3.0+

   ## Installation

   Coming soon...

   ## Usage

   Coming soon...

   ## License

   MIT
   ```

5. Create `LICENSE` (MIT):
   ```
   MIT License

   Copyright (c) 2025 Brandon Williams

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   SOFTWARE.
   ```

6. Create `Makefile`:
   ```makefile
   .PHONY: all build test lint clean install

   # Binary name
   BINARY_NAME=aseprite-mcp

   # Build variables
   VERSION?=$(shell git describe --tags --always --dirty)
   BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
   LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

   all: lint test build

   build:
   	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/aseprite-mcp

   test:
   	go test -v -race -cover ./...

   test-coverage:
   	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
   	go tool cover -html=coverage.txt -o coverage.html

   lint:
   	go vet ./...
   	go fmt ./...

   clean:
   	rm -rf bin/ dist/ coverage.txt coverage.html
   	go clean

   install:
   	go install $(LDFLAGS) ./cmd/aseprite-mcp

   # Development helpers
   run:
   	go run ./cmd/aseprite-mcp

   deps:
   	go mod download
   	go mod tidy
   ```

7. Create `CHANGELOG.md`:
   ```markdown
   # Changelog

   All notable changes to this project will be documented in this file.

   The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
   and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

   ## [Unreleased]

   ### Added
   - Initial project structure
   ```

**Verification:**
1. Run `go mod tidy` - should complete without errors
2. Verify all directories exist
3. Check `.gitignore` excludes test files

**Testing:**
```bash
# Manual verification
ls -la
cat go.mod
make clean  # Should run without errors

# All commands should complete successfully
```

**Git Commit:**
```bash
git init
git add .
git commit -m "feat: initialize project structure and Go module

- Create directory structure following Go standard layout
- Initialize go.mod with module path
- Add .gitignore, LICENSE, README, Makefile
- Add CHANGELOG.md for version tracking"
```

---

#### Chunk 2: Add Dependencies and Basic Types

**Objective:** Add required dependencies and define core domain types.

**Tasks:**

1. Add dependencies:
   ```bash
   go get github.com/modelcontextprotocol/go-sdk@latest
   go get github.com/willibrandon/mtlog@latest
   go mod tidy
   ```

2. Create `pkg/aseprite/types.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   // Package aseprite provides types and utilities for interacting with Aseprite.
   package aseprite

   import (
   	"fmt"
   	"regexp"
   	"strconv"
   	"strings"
   )

   // Color represents an RGBA color value.
   type Color struct {
   	R uint8 `json:"r"`
   	G uint8 `json:"g"`
   	B uint8 `json:"b"`
   	A uint8 `json:"a"`
   }

   var hexColorPattern = regexp.MustCompile(`^#?([A-Fa-f0-9]{6}|[A-Fa-f0-9]{8})$`)

   // NewColor creates a new Color with the specified RGBA values.
   func NewColor(r, g, b, a uint8) Color {
   	return Color{R: r, G: g, B: b, A: a}
   }

   // NewColorRGB creates a new opaque Color with the specified RGB values.
   func NewColorRGB(r, g, b uint8) Color {
   	return Color{R: r, G: g, B: b, A: 255}
   }

   // FromHex parses a hex color string in the format "#RRGGBB" or "#RRGGBBAA".
   // The "#" prefix is optional.
   func (c *Color) FromHex(hex string) error {
   	hex = strings.TrimPrefix(hex, "#")

   	if !hexColorPattern.MatchString("#" + hex) {
   		return fmt.Errorf("invalid hex color format: %q (expected #RRGGBB or #RRGGBBAA)", hex)
   	}

   	// Parse RGB
   	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
   	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
   	b, _ := strconv.ParseUint(hex[4:6], 16, 8)

   	c.R = uint8(r)
   	c.G = uint8(g)
   	c.B = uint8(b)

   	// Parse alpha if present
   	if len(hex) == 8 {
   		a, _ := strconv.ParseUint(hex[6:8], 16, 8)
   		c.A = uint8(a)
   	} else {
   		c.A = 255 // Opaque by default
   	}

   	return nil
   }

   // ToHex converts the color to a hex string in the format "#RRGGBBAA".
   func (c Color) ToHex() string {
   	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
   }

   // ToHexRGB converts the color to a hex string in the format "#RRGGBB" (ignoring alpha).
   func (c Color) ToHexRGB() string {
   	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
   }

   // Point represents a 2D coordinate.
   type Point struct {
   	X int `json:"x"`
   	Y int `json:"y"`
   }

   // Rectangle represents a rectangular region.
   type Rectangle struct {
   	X      int `json:"x"`
   	Y      int `json:"y"`
   	Width  int `json:"width"`
   	Height int `json:"height"`
   }

   // Pixel represents a single pixel with color and position.
   type Pixel struct {
   	Point
   	Color Color `json:"color"`
   }

   // SpriteInfo contains metadata about a sprite.
   type SpriteInfo struct {
   	Width      int      `json:"width"`
   	Height     int      `json:"height"`
   	ColorMode  string   `json:"color_mode"`
   	FrameCount int      `json:"frame_count"`
   	LayerCount int      `json:"layer_count"`
   	Layers     []string `json:"layers"`
   }

   // ColorMode represents the color mode of a sprite.
   type ColorMode string

   const (
   	ColorModeRGB       ColorMode = "rgb"
   	ColorModeGrayscale ColorMode = "grayscale"
   	ColorModeIndexed   ColorMode = "indexed"
   )

   // String returns the string representation of the color mode.
   func (cm ColorMode) String() string {
   	return string(cm)
   }

   // ToLua returns the Lua constant for the color mode.
   func (cm ColorMode) ToLua() string {
   	switch cm {
   	case ColorModeRGB:
   		return "ColorMode.RGB"
   	case ColorModeGrayscale:
   		return "ColorMode.GRAYSCALE"
   	case ColorModeIndexed:
   		return "ColorMode.INDEXED"
   	default:
   		return "ColorMode.RGB"
   	}
   }
   ```

3. Create `pkg/aseprite/types_test.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   package aseprite

   import (
   	"testing"
   )

   func TestColor_FromHex(t *testing.T) {
   	tests := []struct {
   		name    string
   		hex     string
   		want    Color
   		wantErr bool
   	}{
   		{
   			name: "RGB with hash",
   			hex:  "#FF0000",
   			want: Color{R: 255, G: 0, B: 0, A: 255},
   		},
   		{
   			name: "RGB without hash",
   			hex:  "00FF00",
   			want: Color{R: 0, G: 255, B: 0, A: 255},
   		},
   		{
   			name: "RGBA with hash",
   			hex:  "#0000FF80",
   			want: Color{R: 0, G: 0, B: 255, A: 128},
   		},
   		{
   			name: "RGBA without hash",
   			hex:  "FFFF00FF",
   			want: Color{R: 255, G: 255, B: 0, A: 255},
   		},
   		{
   			name:    "invalid format",
   			hex:     "invalid",
   			wantErr: true,
   		},
   		{
   			name:    "too short",
   			hex:     "#FFF",
   			wantErr: true,
   		},
   		{
   			name:    "too long",
   			hex:     "#FFFFFFFFF",
   			wantErr: true,
   		},
   	}

   	for _, tt := range tests {
   		t.Run(tt.name, func(t *testing.T) {
   			var c Color
   			err := c.FromHex(tt.hex)

   			if (err != nil) != tt.wantErr {
   				t.Errorf("FromHex() error = %v, wantErr %v", err, tt.wantErr)
   				return
   			}

   			if !tt.wantErr && c != tt.want {
   				t.Errorf("FromHex() = %+v, want %+v", c, tt.want)
   			}
   		})
   	}
   }

   func TestColor_ToHex(t *testing.T) {
   	tests := []struct {
   		name  string
   		color Color
   		want  string
   	}{
   		{
   			name:  "red",
   			color: Color{R: 255, G: 0, B: 0, A: 255},
   			want:  "#FF0000FF",
   		},
   		{
   			name:  "green",
   			color: Color{R: 0, G: 255, B: 0, A: 255},
   			want:  "#00FF00FF",
   		},
   		{
   			name:  "blue with alpha",
   			color: Color{R: 0, G: 0, B: 255, A: 128},
   			want:  "#0000FF80",
   		},
   		{
   			name:  "black transparent",
   			color: Color{R: 0, G: 0, B: 0, A: 0},
   			want:  "#00000000",
   		},
   	}

   	for _, tt := range tests {
   		t.Run(tt.name, func(t *testing.T) {
   			if got := tt.color.ToHex(); got != tt.want {
   				t.Errorf("ToHex() = %v, want %v", got, tt.want)
   			}
   		})
   	}
   }

   func TestColorMode_ToLua(t *testing.T) {
   	tests := []struct {
   		name string
   		cm   ColorMode
   		want string
   	}{
   		{
   			name: "RGB",
   			cm:   ColorModeRGB,
   			want: "ColorMode.RGB",
   		},
   		{
   			name: "Grayscale",
   			cm:   ColorModeGrayscale,
   			want: "ColorMode.GRAYSCALE",
   		},
   		{
   			name: "Indexed",
   			cm:   ColorModeIndexed,
   			want: "ColorMode.INDEXED",
   		},
   		{
   			name: "Unknown defaults to RGB",
   			cm:   ColorMode("unknown"),
   			want: "ColorMode.RGB",
   		},
   	}

   	for _, tt := range tests {
   		t.Run(tt.name, func(t *testing.T) {
   			if got := tt.cm.ToLua(); got != tt.want {
   				t.Errorf("ToLua() = %v, want %v", got, tt.want)
   			}
   		})
   	}
   }

   func TestNewColor(t *testing.T) {
   	c := NewColor(255, 128, 64, 32)

   	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 32 {
   		t.Errorf("NewColor() = %+v, want R:255 G:128 B:64 A:32", c)
   	}
   }

   func TestNewColorRGB(t *testing.T) {
   	c := NewColorRGB(255, 128, 64)

   	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 255 {
   		t.Errorf("NewColorRGB() = %+v, want R:255 G:128 B:64 A:255", c)
   	}
   }
   ```

**Verification:**
1. Run `go mod tidy` - should download dependencies
2. Run `go build ./...` - should compile without errors
3. Check `go.mod` contains correct dependencies

**Testing:**
```bash
# Run tests
go test -v ./pkg/aseprite

# Check test coverage
go test -cover ./pkg/aseprite

# Should see:
# - All Color tests passing
# - Coverage > 90%
```

**Git Commit:**
```bash
git add .
git commit -m "feat(aseprite): add core domain types and color utilities

- Add Color type with hex parsing and formatting
- Add Point, Rectangle, Pixel types
- Add SpriteInfo and ColorMode types
- Implement comprehensive unit tests
- Add MCP SDK and mtlog dependencies"
```

---

#### Chunk 3: Configuration Management

**Objective:** Implement configuration loading from environment variables and files.

**Tasks:**

1. Create `pkg/config/config.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   // Package config provides configuration management for the Aseprite MCP server.
   package config

   import (
   	"encoding/json"
   	"fmt"
   	"os"
   	"os/exec"
   	"path/filepath"
   	"runtime"
   	"time"
   )

   // Config holds the server configuration.
   type Config struct {
   	// AsepritePath is the path to the Aseprite executable.
   	AsepritePath string `json:"aseprite_path"`

   	// TempDir is the directory for temporary files.
   	TempDir string `json:"temp_dir"`

   	// Timeout is the maximum duration for operations.
   	Timeout time.Duration `json:"timeout"`

   	// LogLevel is the logging verbosity level.
   	LogLevel string `json:"log_level"`
   }

   // Default configuration values.
   const (
   	DefaultTimeout  = 30 * time.Second
   	DefaultLogLevel = "info"
   )

   // Load loads configuration from environment variables and optional config file.
   // Environment variables take precedence over config file values.
   func Load() (*Config, error) {
   	cfg := &Config{
   		Timeout:  DefaultTimeout,
   		LogLevel: DefaultLogLevel,
   	}

   	// Try to load from config file
   	if err := cfg.loadFromFile(); err != nil && !os.IsNotExist(err) {
   		return nil, fmt.Errorf("failed to load config file: %w", err)
   	}

   	// Override with environment variables
   	cfg.loadFromEnv()

   	// Set defaults for unset values
   	if err := cfg.setDefaults(); err != nil {
   		return nil, fmt.Errorf("failed to set defaults: %w", err)
   	}

   	// Validate configuration
   	if err := cfg.Validate(); err != nil {
   		return nil, fmt.Errorf("invalid configuration: %w", err)
   	}

   	return cfg, nil
   }

   // loadFromFile loads configuration from the default config file location.
   func (c *Config) loadFromFile() error {
   	configPath := getConfigFilePath()

   	data, err := os.ReadFile(configPath)
   	if err != nil {
   		return err
   	}

   	return json.Unmarshal(data, c)
   }

   // loadFromEnv overrides configuration with environment variables.
   func (c *Config) loadFromEnv() {
   	if path := os.Getenv("ASEPRITE_PATH"); path != "" {
   		c.AsepritePath = path
   	}

   	if dir := os.Getenv("ASEPRITE_TEMP_DIR"); dir != "" {
   		c.TempDir = dir
   	}

   	if timeout := os.Getenv("ASEPRITE_TIMEOUT"); timeout != "" {
   		if d, err := time.ParseDuration(timeout + "s"); err == nil {
   			c.Timeout = d
   		}
   	}

   	if level := os.Getenv("ASEPRITE_LOG_LEVEL"); level != "" {
   		c.LogLevel = level
   	}
   }

   // setDefaults sets default values for unset configuration fields.
   func (c *Config) setDefaults() error {
   	// Set default Aseprite path if not configured
   	if c.AsepritePath == "" {
   		path, err := findAsepritePath()
   		if err != nil {
   			return fmt.Errorf("could not find Aseprite executable: %w", err)
   		}
   		c.AsepritePath = path
   	}

   	// Set default temp directory if not configured
   	if c.TempDir == "" {
   		c.TempDir = filepath.Join(os.TempDir(), "aseprite-mcp")
   	}

   	// Ensure temp directory exists
   	if err := os.MkdirAll(c.TempDir, 0755); err != nil {
   		return fmt.Errorf("failed to create temp directory: %w", err)
   	}

   	return nil
   }

   // Validate checks if the configuration is valid.
   func (c *Config) Validate() error {
   	// Check if Aseprite executable exists
   	if _, err := os.Stat(c.AsepritePath); os.IsNotExist(err) {
   		return fmt.Errorf("aseprite executable not found at %s", c.AsepritePath)
   	}

   	// Check if temp directory is writable
   	testFile := filepath.Join(c.TempDir, ".test")
   	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
   		return fmt.Errorf("temp directory is not writable: %w", err)
   	}
   	os.Remove(testFile)

   	// Validate timeout
   	if c.Timeout <= 0 {
   		return fmt.Errorf("timeout must be positive, got %v", c.Timeout)
   	}

   	// Validate log level
   	validLevels := map[string]bool{
   		"debug": true,
   		"info":  true,
   		"warn":  true,
   		"error": true,
   	}
   	if !validLevels[c.LogLevel] {
   		return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", c.LogLevel)
   	}

   	return nil
   }

   // getConfigFilePath returns the default config file path.
   func getConfigFilePath() string {
   	homeDir, _ := os.UserHomeDir()
   	return filepath.Join(homeDir, ".config", "aseprite-mcp", "config.json")
   }

   // findAsepritePath attempts to locate the Aseprite executable.
   func findAsepritePath() (string, error) {
   	// Try common installation locations first
   	commonPaths := getCommonAsepritePaths()
   	for _, path := range commonPaths {
   		if _, err := os.Stat(path); err == nil {
   			return path, nil
   		}
   	}

   	// Search in PATH
   	path, err := exec.LookPath("aseprite")
   	if err != nil {
   		return "", fmt.Errorf("aseprite not found in PATH or common locations")
   	}

   	return path, nil
   }

   // getCommonAsepritePaths returns common installation paths by OS.
   func getCommonAsepritePaths() []string {
   	switch runtime.GOOS {
   	case "windows":
   		return []string{
   			`C:\Program Files\Aseprite\Aseprite.exe`,
   			`C:\Program Files (x86)\Aseprite\Aseprite.exe`,
   		}
   	case "darwin":
   		return []string{
   			"/Applications/Aseprite.app/Contents/MacOS/aseprite",
   			filepath.Join(os.Getenv("HOME"), "Applications", "Aseprite.app", "Contents", "MacOS", "aseprite"),
   		}
   	case "linux":
   		return []string{
   			"/usr/bin/aseprite",
   			"/usr/local/bin/aseprite",
   			filepath.Join(os.Getenv("HOME"), ".local", "bin", "aseprite"),
   		}
   	default:
   		return nil
   	}
   }
   ```

2. Create `pkg/config/config_test.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   package config

   import (
   	"os"
   	"path/filepath"
   	"testing"
   	"time"
   )

   func TestConfig_Validate(t *testing.T) {
   	// Create a temporary directory for testing
   	tempDir, err := os.MkdirTemp("", "aseprite-mcp-test-*")
   	if err != nil {
   		t.Fatal(err)
   	}
   	defer os.RemoveAll(tempDir)

   	// Create a fake aseprite executable
   	fakeAseprite := filepath.Join(tempDir, "aseprite")
   	if err := os.WriteFile(fakeAseprite, []byte("#!/bin/sh\n"), 0755); err != nil {
   		t.Fatal(err)
   	}

   	tests := []struct {
   		name    string
   		config  *Config
   		wantErr bool
   	}{
   		{
   			name: "valid config",
   			config: &Config{
   				AsepritePath: fakeAseprite,
   				TempDir:      tempDir,
   				Timeout:      30 * time.Second,
   				LogLevel:     "info",
   			},
   			wantErr: false,
   		},
   		{
   			name: "missing aseprite executable",
   			config: &Config{
   				AsepritePath: "/nonexistent/aseprite",
   				TempDir:      tempDir,
   				Timeout:      30 * time.Second,
   				LogLevel:     "info",
   			},
   			wantErr: true,
   		},
   		{
   			name: "invalid timeout",
   			config: &Config{
   				AsepritePath: fakeAseprite,
   				TempDir:      tempDir,
   				Timeout:      -1 * time.Second,
   				LogLevel:     "info",
   			},
   			wantErr: true,
   		},
   		{
   			name: "invalid log level",
   			config: &Config{
   				AsepritePath: fakeAseprite,
   				TempDir:      tempDir,
   				Timeout:      30 * time.Second,
   				LogLevel:     "invalid",
   			},
   			wantErr: true,
   		},
   	}

   	for _, tt := range tests {
   		t.Run(tt.name, func(t *testing.T) {
   			err := tt.config.Validate()
   			if (err != nil) != tt.wantErr {
   				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
   			}
   		})
   	}
   }

   func TestConfig_LoadFromEnv(t *testing.T) {
   	// Save original env vars
   	origPath := os.Getenv("ASEPRITE_PATH")
   	origTempDir := os.Getenv("ASEPRITE_TEMP_DIR")
   	origTimeout := os.Getenv("ASEPRITE_TIMEOUT")
   	origLogLevel := os.Getenv("ASEPRITE_LOG_LEVEL")

   	// Restore env vars after test
   	defer func() {
   		os.Setenv("ASEPRITE_PATH", origPath)
   		os.Setenv("ASEPRITE_TEMP_DIR", origTempDir)
   		os.Setenv("ASEPRITE_TIMEOUT", origTimeout)
   		os.Setenv("ASEPRITE_LOG_LEVEL", origLogLevel)
   	}()

   	// Set test env vars
   	os.Setenv("ASEPRITE_PATH", "/test/aseprite")
   	os.Setenv("ASEPRITE_TEMP_DIR", "/test/tmp")
   	os.Setenv("ASEPRITE_TIMEOUT", "60")
   	os.Setenv("ASEPRITE_LOG_LEVEL", "debug")

   	cfg := &Config{}
   	cfg.loadFromEnv()

   	if cfg.AsepritePath != "/test/aseprite" {
   		t.Errorf("AsepritePath = %v, want /test/aseprite", cfg.AsepritePath)
   	}

   	if cfg.TempDir != "/test/tmp" {
   		t.Errorf("TempDir = %v, want /test/tmp", cfg.TempDir)
   	}

   	if cfg.Timeout != 60*time.Second {
   		t.Errorf("Timeout = %v, want 60s", cfg.Timeout)
   	}

   	if cfg.LogLevel != "debug" {
   		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
   	}
   }

   func TestGetCommonAsepritePaths(t *testing.T) {
   	paths := getCommonAsepritePaths()

   	if len(paths) == 0 {
   		t.Error("getCommonAsepritePaths() returned no paths")
   	}

   	// Just verify we get some platform-specific paths
   	// Don't check exact values as they vary by OS
   	t.Logf("Common paths: %v", paths)
   }
   ```

**Verification:**
1. Code compiles without errors
2. Tests pass with >80% coverage

**Testing:**
```bash
# Run tests
go test -v ./pkg/config

# Check coverage
go test -cover ./pkg/config

# Test with environment variables
ASEPRITE_PATH=/test/path ASEPRITE_LOG_LEVEL=debug go test -v ./pkg/config
```

**Git Commit:**
```bash
git add .
git commit -m "feat(config): implement configuration management

- Add Config struct with env var and file support
- Implement Aseprite path discovery for Windows/macOS/Linux
- Add validation for paths, timeout, and log levels
- Implement comprehensive unit tests with >80% coverage"
```

---

### Phase 1: Core Aseprite Integration (Chunks 4-8)

#### Chunk 4: Aseprite Client - Command Execution

**Objective:** Implement the Aseprite command executor with timeout and error handling.

**Tasks:**

1. Create `pkg/aseprite/client.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   package aseprite

   import (
   	"bytes"
   	"context"
   	"fmt"
   	"os"
   	"os/exec"
   	"path/filepath"
   	"strings"
   	"time"
   )

   // Client executes Aseprite commands and Lua scripts.
   type Client struct {
   	execPath string
   	tempDir  string
   	timeout  time.Duration
   }

   // NewClient creates a new Aseprite client.
   func NewClient(execPath, tempDir string, timeout time.Duration) *Client {
   	return &Client{
   		execPath: execPath,
   		tempDir:  tempDir,
   		timeout:  timeout,
   	}
   }

   // ExecuteCommand runs an Aseprite command with the given arguments.
   func (c *Client) ExecuteCommand(ctx context.Context, args []string) (string, error) {
   	// Create context with timeout
   	ctx, cancel := context.WithTimeout(ctx, c.timeout)
   	defer cancel()

   	// Build command
   	cmd := exec.CommandContext(ctx, c.execPath, args...)

   	// Capture stdout and stderr
   	var stdout, stderr bytes.Buffer
   	cmd.Stdout = &stdout
   	cmd.Stderr = &stderr

   	// Execute command
   	err := cmd.Run()

   	// Check for errors
   	if err != nil {
   		// Check if it was a timeout
   		if ctx.Err() == context.DeadlineExceeded {
   			return "", fmt.Errorf("aseprite command timed out after %v", c.timeout)
   		}

   		// Include stderr in error message
   		if stderr.Len() > 0 {
   			return "", fmt.Errorf("aseprite command failed: %w\nstderr: %s", err, stderr.String())
   		}

   		return "", fmt.Errorf("aseprite command failed: %w", err)
   	}

   	return stdout.String(), nil
   }

   // ExecuteLua executes a Lua script in Aseprite batch mode.
   // If spritePath is non-empty, the sprite will be opened before running the script.
   func (c *Client) ExecuteLua(ctx context.Context, script string, spritePath string) (string, error) {
   	// Create temporary script file
   	scriptPath, cleanup, err := c.createTempScript(script)
   	if err != nil {
   		return "", fmt.Errorf("failed to create temp script: %w", err)
   	}
   	defer cleanup()

   	// Build arguments
   	args := []string{"--batch"}

   	// Add sprite path if specified
   	if spritePath != "" {
   		// Verify sprite exists
   		if _, err := os.Stat(spritePath); os.IsNotExist(err) {
   			return "", fmt.Errorf("sprite file not found: %s", spritePath)
   		}
   		args = append(args, spritePath)
   	}

   	// Add script argument
   	args = append(args, "--script", scriptPath)

   	// Execute command
   	return c.ExecuteCommand(ctx, args)
   }

   // GetVersion retrieves the Aseprite version.
   func (c *Client) GetVersion(ctx context.Context) (string, error) {
   	output, err := c.ExecuteCommand(ctx, []string{"--version"})
   	if err != nil {
   		return "", err
   	}

   	// Parse version from output (format: "Aseprite 1.3.x")
   	lines := strings.Split(output, "\n")
   	if len(lines) > 0 {
   		return strings.TrimSpace(lines[0]), nil
   	}

   	return "", fmt.Errorf("failed to parse version from output: %s", output)
   }

   // createTempScript creates a temporary Lua script file.
   // Returns the script path and a cleanup function.
   func (c *Client) createTempScript(script string) (string, func(), error) {
   	// Ensure temp directory exists
   	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
   		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
   	}

   	// Create temp file with .lua extension
   	tmpFile, err := os.CreateTemp(c.tempDir, "script-*.lua")
   	if err != nil {
   		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
   	}

   	scriptPath := tmpFile.Name()

   	// Write script content
   	if _, err := tmpFile.WriteString(script); err != nil {
   		tmpFile.Close()
   		os.Remove(scriptPath)
   		return "", nil, fmt.Errorf("failed to write script: %w", err)
   	}

   	// Close file
   	if err := tmpFile.Close(); err != nil {
   		os.Remove(scriptPath)
   		return "", nil, fmt.Errorf("failed to close temp file: %w", err)
   	}

   	// Return cleanup function
   	cleanup := func() {
   		os.Remove(scriptPath)
   	}

   	return scriptPath, cleanup, nil
   }

   // CleanupOldTempFiles removes temporary files older than the specified duration.
   func (c *Client) CleanupOldTempFiles(maxAge time.Duration) error {
   	entries, err := os.ReadDir(c.tempDir)
   	if err != nil {
   		if os.IsNotExist(err) {
   			return nil // Directory doesn't exist, nothing to clean
   		}
   		return fmt.Errorf("failed to read temp directory: %w", err)
   	}

   	now := time.Now()
   	for _, entry := range entries {
   		// Skip directories
   		if entry.IsDir() {
   			continue
   		}

   		// Check if file matches our pattern (script-*.lua)
   		if !strings.HasPrefix(entry.Name(), "script-") || !strings.HasSuffix(entry.Name(), ".lua") {
   			continue
   		}

   		// Get file info
   		info, err := entry.Info()
   		if err != nil {
   			continue
   		}

   		// Check age
   		if now.Sub(info.ModTime()) > maxAge {
   			filePath := filepath.Join(c.tempDir, entry.Name())
   			os.Remove(filePath)
   		}
   	}

   	return nil
   }
   ```

2. Create `pkg/aseprite/client_test.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   package aseprite

   import (
   	"context"
   	"os"
   	"path/filepath"
   	"testing"
   	"time"
   )

   func TestClient_CreateTempScript(t *testing.T) {
   	tempDir := t.TempDir()
   	client := NewClient("aseprite", tempDir, 30*time.Second)

   	script := "print('hello')"
   	scriptPath, cleanup, err := client.createTempScript(script)
   	if err != nil {
   		t.Fatalf("createTempScript() error = %v", err)
   	}
   	defer cleanup()

   	// Verify file was created
   	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
   		t.Errorf("script file was not created at %s", scriptPath)
   	}

   	// Verify content
   	content, err := os.ReadFile(scriptPath)
   	if err != nil {
   		t.Fatalf("failed to read script file: %v", err)
   	}

   	if string(content) != script {
   		t.Errorf("script content = %q, want %q", string(content), script)
   	}

   	// Verify cleanup removes file
   	cleanup()
   	if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
   		t.Errorf("script file was not cleaned up")
   	}
   }

   func TestClient_CleanupOldTempFiles(t *testing.T) {
   	tempDir := t.TempDir()
   	client := NewClient("aseprite", tempDir, 30*time.Second)

   	// Create some test files
   	oldFile := filepath.Join(tempDir, "script-old.lua")
   	newFile := filepath.Join(tempDir, "script-new.lua")
   	otherFile := filepath.Join(tempDir, "other.txt")

   	// Create files
   	for _, file := range []string{oldFile, newFile, otherFile} {
   		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
   			t.Fatal(err)
   		}
   	}

   	// Make old file actually old
   	oldTime := time.Now().Add(-2 * time.Hour)
   	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
   		t.Fatal(err)
   	}

   	// Run cleanup with 1 hour max age
   	if err := client.CleanupOldTempFiles(1 * time.Hour); err != nil {
   		t.Fatalf("CleanupOldTempFiles() error = %v", err)
   	}

   	// Verify old file was removed
   	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
   		t.Errorf("old script file was not removed")
   	}

   	// Verify new file still exists
   	if _, err := os.Stat(newFile); os.IsNotExist(err) {
   		t.Errorf("new script file was incorrectly removed")
   	}

   	// Verify other file still exists
   	if _, err := os.Stat(otherFile); os.IsNotExist(err) {
   		t.Errorf("non-script file was incorrectly removed")
   	}
   }

   func TestClient_ExecuteLua_MissingSprite(t *testing.T) {
   	tempDir := t.TempDir()
   	client := NewClient("aseprite", tempDir, 30*time.Second)

   	ctx := context.Background()
   	_, err := client.ExecuteLua(ctx, "print('test')", "/nonexistent/sprite.aseprite")

   	if err == nil {
   		t.Error("ExecuteLua() expected error for missing sprite, got nil")
   	}
   }

   // Note: Additional tests that require actual Aseprite execution
   // should be placed in integration tests
   ```

**Verification:**
1. Code compiles without errors
2. Unit tests pass
3. Temp file creation and cleanup work correctly

**Testing:**
```bash
# Run unit tests
go test -v ./pkg/aseprite

# Run with coverage
go test -cover ./pkg/aseprite

# Manual test with real Aseprite (if available)
# This will be tested more thoroughly in integration tests
```

**Git Commit:**
```bash
git add .
git commit -m "feat(aseprite): implement command executor with timeout

- Add Client struct for Aseprite command execution
- Implement ExecuteCommand with context timeout
- Implement ExecuteLua for Lua script execution
- Add temp file management with cleanup
- Add GetVersion for version detection
- Implement comprehensive unit tests"
```

---

#### Chunk 5: Lua Script Generation - Utilities

**Objective:** Implement Lua script generation utilities and escaping functions.

**Tasks:**

1. Create `pkg/aseprite/lua.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   package aseprite

   import (
   	"fmt"
   	"strings"
   )

   // LuaGenerator provides utilities for generating Lua scripts for Aseprite.
   type LuaGenerator struct{}

   // NewLuaGenerator creates a new Lua script generator.
   func NewLuaGenerator() *LuaGenerator {
   	return &LuaGenerator{}
   }

   // EscapeString escapes a string for use in Lua code.
   // It handles quotes, newlines, and other special characters.
   func EscapeString(s string) string {
   	// Replace backslashes first
   	s = strings.ReplaceAll(s, `\`, `\\`)

   	// Replace quotes
   	s = strings.ReplaceAll(s, `"`, `\"`)

   	// Replace newlines
   	s = strings.ReplaceAll(s, "\n", `\n`)
   	s = strings.ReplaceAll(s, "\r", `\r`)
   	s = strings.ReplaceAll(s, "\t", `\t`)

   	return s
   }

   // FormatColor formats a Color as a Lua Color constructor call.
   func FormatColor(c Color) string {
   	return fmt.Sprintf("Color(%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
   }

   // FormatPoint formats a Point as a Lua Point constructor call.
   func FormatPoint(p Point) string {
   	return fmt.Sprintf("Point(%d, %d)", p.X, p.Y)
   }

   // FormatRectangle formats a Rectangle as a Lua Rectangle constructor call.
   func FormatRectangle(r Rectangle) string {
   	return fmt.Sprintf("Rectangle(%d, %d, %d, %d)", r.X, r.Y, r.Width, r.Height)
   }

   // WrapInTransaction wraps Lua code in an app.transaction for atomicity.
   func WrapInTransaction(code string) string {
   	return fmt.Sprintf(`app.transaction(function()
   %s
   end)`, code)
   }

   // CreateCanvas generates a Lua script to create a new sprite.
   func (g *LuaGenerator) CreateCanvas(width, height int, colorMode ColorMode) string {
   	return fmt.Sprintf(`local spr = Sprite(%d, %d, %s)
   local filename = os.tmpname() .. ".aseprite"
   spr:saveAs(filename)
   print(filename)`, width, height, colorMode.ToLua())
   }

   // GetSpriteInfo generates a Lua script to retrieve sprite metadata.
   func (g *LuaGenerator) GetSpriteInfo() string {
   	return `local spr = app.activeSprite
   if not spr then
   	error("No active sprite")
   end

   -- Collect layer names
   local layers = {}
   for i, layer in ipairs(spr.layers) do
   	table.insert(layers, layer.name)
   end

   -- Format as JSON-like output
   local output = string.format([[{
   	"width": %d,
   	"height": %d,
   	"color_mode": "%s",
   	"frame_count": %d,
   	"layer_count": %d,
   	"layers": ["%s"]
   }]],
   	spr.width,
   	spr.height,
   	tostring(spr.colorMode),
   	#spr.frames,
   	#spr.layers,
   	table.concat(layers, '","')
   )

   print(output)`
   }

   // AddLayer generates a Lua script to add a new layer.
   func (g *LuaGenerator) AddLayer(layerName string) string {
   	escapedName := EscapeString(layerName)
   	return fmt.Sprintf(`local spr = app.activeSprite
   if not spr then
   	error("No active sprite")
   end

   app.transaction(function()
   	local layer = spr:newLayer()
   	layer.name = "%s"
   end)

   spr:saveAs(spr.filename)
   print("Layer added successfully")`, escapedName)
   }

   // AddFrame generates a Lua script to add a new frame.
   func (g *LuaGenerator) AddFrame(durationMs int) string {
   	durationSec := float64(durationMs) / 1000.0
   	return fmt.Sprintf(`local spr = app.activeSprite
   if not spr then
   	error("No active sprite")
   end

   app.transaction(function()
   	local frame = spr:newFrame()
   	frame.duration = %.3f
   end)

   spr:saveAs(spr.filename)
   print(#spr.frames)`, durationSec)
   }

   // DrawPixels generates a Lua script to draw multiple pixels.
   func (g *LuaGenerator) DrawPixels(layerName string, frameNumber int, pixels []Pixel) string {
   	var sb strings.Builder

   	escapedName := EscapeString(layerName)

   	sb.WriteString(fmt.Sprintf(`local spr = app.activeSprite
   if not spr then
   	error("No active sprite")
   end

   local layer = spr:findLayerByName("%s")
   if not layer then
   	error("Layer not found: %s")
   end

   local frame = spr.frames[%d]
   if not frame then
   	error("Frame not found: %d")
   end

   app.transaction(function()
   	local cel = layer:cel(frame)
   	if not cel then
   		cel = spr:newCel(layer, frame)
   	end

   	local img = cel.image
   `, escapedName, escapedName, frameNumber, frameNumber))

   	// Add pixel drawing commands
   	for _, p := range pixels {
   		sb.WriteString(fmt.Sprintf("\timg:putPixel(%d, %d, %s)\n", p.X, p.Y, FormatColor(p.Color)))
   	}

   	sb.WriteString(`end)

   spr:saveAs(spr.filename)
   print("Pixels drawn successfully")`)

   	return sb.String()
   }

   // DrawLine generates a Lua script to draw a line.
   func (g *LuaGenerator) DrawLine(layerName string, frameNumber int, x1, y1, x2, y2 int, color Color, thickness int) string {
   	escapedName := EscapeString(layerName)
   	return fmt.Sprintf(`local spr = app.activeSprite
   if not spr then
   	error("No active sprite")
   end

   local layer = spr:findLayerByName("%s")
   if not layer then
   	error("Layer not found: %s")
   end

   local frame = spr.frames[%d]
   if not frame then
   	error("Frame not found: %d")
   end

   app.transaction(function()
   	app.activeLayer = layer
   	app.activeFrame = frame

   	local brush = Brush()
   	brush.size = %d

   	app.useTool{
   		tool = "line",
   		color = %s,
   		brush = brush,
   		points = {%s, %s}
   	}
   end)

   spr:saveAs(spr.filename)
   print("Line drawn successfully")`,
   		escapedName, escapedName,
   		frameNumber, frameNumber,
   		thickness,
   		FormatColor(color),
   		FormatPoint(Point{X: x1, Y: y1}),
   		FormatPoint(Point{X: x2, Y: y2}))
   }

   // ExportSprite generates a Lua script to export a sprite.
   func (g *LuaGenerator) ExportSprite(outputPath string, frameNumber int) string {
   	escapedPath := EscapeString(outputPath)

   	if frameNumber > 0 {
   		// Export specific frame
   		return fmt.Sprintf(`local spr = app.activeSprite
   if not spr then
   	error("No active sprite")
   end

   local frame = spr.frames[%d]
   if not frame then
   	error("Frame not found: %d")
   end

   spr:saveCopyAs{
   	filename = "%s",
   	frame = frame
   }

   print("Exported successfully")`, frameNumber, frameNumber, escapedPath)
   	}

   	// Export all frames
   	return fmt.Sprintf(`local spr = app.activeSprite
   if not spr then
   	error("No active sprite")
   end

   spr:saveCopyAs("%s")
   print("Exported successfully")`, escapedPath)
   }
   ```

2. Create `pkg/aseprite/lua_test.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   package aseprite

   import (
   	"strings"
   	"testing"
   )

   func TestEscapeString(t *testing.T) {
   	tests := []struct {
   		name  string
   		input string
   		want  string
   	}{
   		{
   			name:  "simple string",
   			input: "hello",
   			want:  "hello",
   		},
   		{
   			name:  "string with quotes",
   			input: `hello "world"`,
   			want:  `hello \"world\"`,
   		},
   		{
   			name:  "string with backslash",
   			input: `hello\world`,
   			want:  `hello\\world`,
   		},
   		{
   			name:  "string with newline",
   			input: "hello\nworld",
   			want:  `hello\nworld`,
   		},
   		{
   			name:  "string with tab",
   			input: "hello\tworld",
   			want:  `hello\tworld`,
   		},
   		{
   			name:  "complex string",
   			input: `C:\path\to\file "with quotes"\nand newlines`,
   			want:  `C:\\path\\to\\file \"with quotes\"\nand newlines`,
   		},
   	}

   	for _, tt := range tests {
   		t.Run(tt.name, func(t *testing.T) {
   			if got := EscapeString(tt.input); got != tt.want {
   				t.Errorf("EscapeString() = %q, want %q", got, tt.want)
   			}
   		})
   	}
   }

   func TestFormatColor(t *testing.T) {
   	tests := []struct {
   		name  string
   		color Color
   		want  string
   	}{
   		{
   			name:  "red",
   			color: Color{R: 255, G: 0, B: 0, A: 255},
   			want:  "Color(255, 0, 0, 255)",
   		},
   		{
   			name:  "green with alpha",
   			color: Color{R: 0, G: 255, B: 0, A: 128},
   			want:  "Color(0, 255, 0, 128)",
   		},
   		{
   			name:  "transparent black",
   			color: Color{R: 0, G: 0, B: 0, A: 0},
   			want:  "Color(0, 0, 0, 0)",
   		},
   	}

   	for _, tt := range tests {
   		t.Run(tt.name, func(t *testing.T) {
   			if got := FormatColor(tt.color); got != tt.want {
   				t.Errorf("FormatColor() = %v, want %v", got, tt.want)
   			}
   		})
   	}
   }

   func TestFormatPoint(t *testing.T) {
   	p := Point{X: 10, Y: 20}
   	want := "Point(10, 20)"

   	if got := FormatPoint(p); got != want {
   		t.Errorf("FormatPoint() = %v, want %v", got, want)
   	}
   }

   func TestFormatRectangle(t *testing.T) {
   	r := Rectangle{X: 10, Y: 20, Width: 30, Height: 40}
   	want := "Rectangle(10, 20, 30, 40)"

   	if got := FormatRectangle(r); got != want {
   		t.Errorf("FormatRectangle() = %v, want %v", got, want)
   	}
   }

   func TestLuaGenerator_CreateCanvas(t *testing.T) {
   	gen := NewLuaGenerator()

   	script := gen.CreateCanvas(800, 600, ColorModeRGB)

   	// Verify script contains expected elements
   	if !strings.Contains(script, "Sprite(800, 600, ColorMode.RGB)") {
   		t.Error("script missing Sprite constructor call")
   	}

   	if !strings.Contains(script, "spr:saveAs(filename)") {
   		t.Error("script missing saveAs call")
   	}

   	if !strings.Contains(script, "print(filename)") {
   		t.Error("script missing print statement")
   	}
   }

   func TestLuaGenerator_AddLayer(t *testing.T) {
   	gen := NewLuaGenerator()

   	script := gen.AddLayer("My Layer")

   	// Verify script contains expected elements
   	if !strings.Contains(script, "spr:newLayer()") {
   		t.Error("script missing newLayer call")
   	}

   	if !strings.Contains(script, `layer.name = "My Layer"`) {
   		t.Error("script missing layer name assignment")
   	}

   	if !strings.Contains(script, "app.transaction(function()") {
   		t.Error("script not wrapped in transaction")
   	}
   }

   func TestLuaGenerator_DrawPixels(t *testing.T) {
   	gen := NewLuaGenerator()

   	pixels := []Pixel{
   		{Point: Point{X: 0, Y: 0}, Color: Color{R: 255, G: 0, B: 0, A: 255}},
   		{Point: Point{X: 1, Y: 1}, Color: Color{R: 0, G: 255, B: 0, A: 255}},
   	}

   	script := gen.DrawPixels("Layer 1", 1, pixels)

   	// Verify script contains expected elements
   	if !strings.Contains(script, `findLayerByName("Layer 1")`) {
   		t.Error("script missing layer lookup")
   	}

   	if !strings.Contains(script, "img:putPixel(0, 0, Color(255, 0, 0, 255))") {
   		t.Error("script missing first pixel")
   	}

   	if !strings.Contains(script, "img:putPixel(1, 1, Color(0, 255, 0, 255))") {
   		t.Error("script missing second pixel")
   	}
   }

   func TestLuaGenerator_ExportSprite(t *testing.T) {
   	gen := NewLuaGenerator()

   	t.Run("export all frames", func(t *testing.T) {
   		script := gen.ExportSprite("output.png", 0)

   		if !strings.Contains(script, `saveCopyAs("output.png")`) {
   			t.Error("script missing saveCopyAs call")
   		}
   	})

   	t.Run("export specific frame", func(t *testing.T) {
   		script := gen.ExportSprite("output.png", 2)

   		if !strings.Contains(script, "spr.frames[2]") {
   			t.Error("script missing frame selection")
   		}

   		if !strings.Contains(script, "frame = frame") {
   			t.Error("script missing frame parameter")
   		}
   	})
   }
   ```

**Verification:**
1. All Lua generation tests pass
2. String escaping handles special characters correctly
3. Generated Lua scripts have proper syntax

**Testing:**
```bash
# Run tests
go test -v ./pkg/aseprite

# Check coverage
go test -cover ./pkg/aseprite

# Should see >90% coverage for lua.go
```

**Git Commit:**
```bash
git add .
git commit -m "feat(aseprite): implement Lua script generation utilities

- Add EscapeString for safe string interpolation
- Add color, point, rectangle formatters
- Implement CreateCanvas, AddLayer, AddFrame generators
- Implement DrawPixels, DrawLine generators
- Implement ExportSprite generator
- Add comprehensive unit tests with >90% coverage"
```

---

#### Chunk 6: Integration Test Infrastructure

**Objective:** Create mock Aseprite for testing and integration test utilities.

**Tasks:**

1. Create `internal/testutil/mock_aseprite.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   // Package testutil provides testing utilities for the Aseprite MCP server.
   package testutil

   import (
   	"fmt"
   	"os"
   	"path/filepath"
   	"strings"
   )

   // MockAseprite represents a mock Aseprite executable for testing.
   type MockAseprite struct {
   	execPath string
   	responses map[string]string
   }

   // NewMockAseprite creates a new mock Aseprite executable.
   func NewMockAseprite(dir string) (*MockAseprite, error) {
   	// Create mock executable script
   	execPath := filepath.Join(dir, "aseprite")
   	if err := createMockScript(execPath); err != nil {
   		return nil, err
   	}

   	return &MockAseprite{
   		execPath: execPath,
   		responses: make(map[string]string),
   	}, nil
   }

   // Path returns the path to the mock executable.
   func (m *MockAseprite) Path() string {
   	return m.execPath
   }

   // SetResponse sets the mock response for a specific script pattern.
   func (m *MockAseprite) SetResponse(pattern, response string) {
   	m.responses[pattern] = response
   }

   // createMockScript creates a platform-specific mock script.
   func createMockScript(path string) error {
   	var script string

   	// Determine script type based on platform
   	if strings.HasSuffix(path, ".bat") || strings.HasSuffix(path, ".cmd") {
   		// Windows batch script
   		script = `@echo off
   echo Aseprite 1.3.0
   exit /b 0
   `
   	} else {
   		// Unix shell script
   		script = `#!/bin/sh
   echo "Aseprite 1.3.0"
   exit 0
   `
   	}

   	// Write script
   	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
   		return fmt.Errorf("failed to create mock script: %w", err)
   	}

   	return nil
   }
   ```

2. Create `internal/testutil/fixtures.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   package testutil

   import (
   	"os"
   	"path/filepath"
   	"testing"
   )

   // CreateTempSprite creates a temporary .aseprite file for testing.
   // Returns the file path and a cleanup function.
   func CreateTempSprite(t *testing.T) (string, func()) {
   	t.Helper()

   	tempDir := t.TempDir()
   	spritePath := filepath.Join(tempDir, "test.aseprite")

   	// Create dummy sprite file (just needs to exist for most tests)
   	if err := os.WriteFile(spritePath, []byte("dummy"), 0644); err != nil {
   		t.Fatalf("failed to create temp sprite: %v", err)
   	}

   	cleanup := func() {
   		os.Remove(spritePath)
   	}

   	return spritePath, cleanup
   }

   // CreateTestConfig creates a test configuration.
   func CreateTestConfig(t *testing.T) (asepritePath, tempDir string) {
   	t.Helper()

   	tempDir = t.TempDir()

   	// Create mock Aseprite
   	mock, err := NewMockAseprite(tempDir)
   	if err != nil {
   		t.Fatalf("failed to create mock aseprite: %v", err)
   	}

   	return mock.Path(), tempDir
   }
   ```

3. Create `pkg/aseprite/integration_test.go`:
   ```go
   // Copyright 2025 Brandon Williams. All rights reserved.
   // Use of this source code is governed by an MIT-style
   // license that can be found in the LICENSE file.

   //go:build integration
   // +build integration

   package aseprite

   import (
   	"context"
   	"os"
   	"os/exec"
   	"path/filepath"
   	"testing"
   	"time"
   )

   // These tests require a real Aseprite installation.
   // Run with: go test -tags=integration ./pkg/aseprite

   func TestIntegration_CreateCanvas(t *testing.T) {
   	// Skip if Aseprite not available
   	asepritePath, err := exec.LookPath("aseprite")
   	if err != nil {
   		t.Skip("Aseprite not found in PATH")
   	}

   	tempDir := t.TempDir()
   	client := NewClient(asepritePath, tempDir, 30*time.Second)
   	gen := NewLuaGenerator()

   	// Generate script to create canvas
   	script := gen.CreateCanvas(100, 100, ColorModeRGB)

   	// Execute
   	ctx := context.Background()
   	output, err := client.ExecuteLua(ctx, script, "")
   	if err != nil {
   		t.Fatalf("ExecuteLua() error = %v", err)
   	}

   	// Output should be the path to the created sprite
   	spritePath := output

   	// Verify file was created
   	if _, err := os.Stat(spritePath); os.IsNotExist(err) {
   		t.Errorf("sprite file was not created at %s", spritePath)
   	}

   	t.Logf("Created sprite at: %s", spritePath)
   }

   func TestIntegration_GetVersion(t *testing.T) {
   	asepritePath, err := exec.LookPath("aseprite")
   	if err != nil {
   		t.Skip("Aseprite not found in PATH")
   	}

   	tempDir := t.TempDir()
   	client := NewClient(asepritePath, tempDir, 30*time.Second)

   	ctx := context.Background()
   	version, err := client.GetVersion(ctx)
   	if err != nil {
   		t.Fatalf("GetVersion() error = %v", err)
   	}

   	if version == "" {
   		t.Error("GetVersion() returned empty version")
   	}

   	t.Logf("Aseprite version: %s", version)
   }
   ```

4. Add integration test documentation to `docs/TESTING.md`:
   ```markdown
   # Testing Guide

   ## Unit Tests

   Run unit tests:
   ```bash
   go test ./...
   ```

   ## Integration Tests

   Integration tests require Aseprite to be installed.

   Run integration tests:
   ```bash
   go test -tags=integration ./...
   ```

   ## Coverage

   Generate coverage report:
   ```bash
   make test-coverage
   open coverage.html
   ```

   ## Manual Testing

   Test with real Aseprite:
   ```bash
   export ASEPRITE_PATH=/path/to/aseprite
   go run ./cmd/aseprite-mcp
   ```
   ```

**Verification:**
1. Mock Aseprite script is created correctly
2. Unit tests pass without real Aseprite
3. Integration tests can be run with real Aseprite (if available)

**Testing:**
```bash
# Run unit tests (no Aseprite required)
go test ./...

# Run integration tests (requires Aseprite)
go test -tags=integration ./...

# Check overall coverage
make test-coverage
```

**Git Commit:**
```bash
git add .
git commit -m "test: add integration test infrastructure

- Create MockAseprite for unit testing
- Add fixture utilities for test data
- Implement integration tests for real Aseprite
- Add TESTING.md documentation
- Use build tags to separate unit and integration tests"
```

---

**Continue to next comment for Chunks 7-10...**