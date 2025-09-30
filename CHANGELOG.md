# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure and Go module initialization
- MCP server with stdio transport
- Configuration management (file-based at ~/.config/aseprite-mcp/config.json)
- Aseprite client with Lua script execution and timeout handling
- Lua script generation utilities with proper escaping
- Canvas management tools (create_canvas, add_layer, add_frame, get_sprite_info)
- Drawing primitive tools (draw_pixels, draw_line, draw_rectangle, draw_circle, fill_area)
- Export tool (export_sprite) supporting PNG, GIF, JPG, BMP formats
- Integration test suite with real Aseprite testing
- End-to-end workflow tests
- Performance benchmarks for all critical operations
- Benchmark documentation showing all PRD targets exceeded
- Build system with cross-platform Makefile
- Health check and version flags
- Structured logging with mtlog (message template logging)