# Performance Benchmarks

This document tracks performance benchmarks for critical operations in the Aseprite MCP server.

## Test Environment

- **CPU**: AMD Ryzen 9 9950X 16-Core Processor
- **OS**: Windows
- **Go Version**: 1.23+
- **Aseprite Version**: Latest stable build

## Baseline Results

Last updated: 2025-09-29

### Canvas Creation

| Operation | Size | P95 Latency | Memory | PRD Target | Status |
|-----------|------|-------------|--------|------------|--------|
| CreateCanvas_Small | 64x64 | ~94ms | 87KB | <500ms | ✅ PASS |
| CreateCanvas_Medium | 320x240 | ~94ms | 86KB | <500ms | ✅ PASS |
| CreateCanvas_Large | 1920x1080 | ~124ms | 87KB | <500ms | ✅ PASS |

### Drawing Primitives

| Operation | Pixel Count | P95 Latency | Memory | PRD Target | Status |
|-----------|-------------|-------------|--------|------------|--------|
| DrawPixels | 10 | ~94ms | 88KB | <300ms | ✅ PASS |
| DrawPixels | 100 | ~93ms | 110KB | <300ms | ✅ PASS |
| DrawPixels | 1,000 | ~93ms | 386KB | <1s | ✅ PASS |
| DrawPixels | 10,000 | ~109ms | 3.5MB | <2s | ✅ PASS |
| DrawLine | - | ~95ms | 87KB | <300ms | ✅ PASS |
| DrawRectangle | - | ~95ms | 86KB | <300ms | ✅ PASS |
| DrawCircle | - | ~96ms | 86KB | <300ms | ✅ PASS |
| FillArea | - | ~93ms | 87KB | <300ms | ✅ PASS |

### Layer & Frame Management

| Operation | P95 Latency | Memory | PRD Target | Status |
|-----------|-------------|--------|------------|--------|
| AddLayer | ~93ms | 83KB | <300ms | ✅ PASS |
| AddFrame | ~95ms | 85KB | <300ms | ✅ PASS |
| GetSpriteInfo | ~94ms | 83KB | <300ms | ✅ PASS |

### Export Operations

| Operation | Format | P95 Latency | Memory | PRD Target | Status |
|-----------|--------|-------------|--------|------------|--------|
| ExportSprite | PNG | ~94ms | 87KB | <1s | ✅ PASS |

### Complete Workflows

| Workflow | Operations | P95 Latency | Memory | PRD Target | Status |
|----------|------------|-------------|--------|------------|--------|
| Simple (Create→Draw→Export) | 3 | ~280ms | 261KB | <1s total | ✅ PASS |
| Animation (4 frames) | 8 | ~839ms | - | <2s total | ✅ PASS |
| Multi-layer | 5 | ~467ms | - | <1s total | ✅ PASS |
| Pixel Batch (1K pixels) | 3 | ~296ms | - | <1s total | ✅ PASS |

## PRD Requirements Compliance

From PRD Section 3.1 (Performance Requirements):

| Requirement | Target | Actual | Status |
|-------------|--------|--------|--------|
| Create canvas P95 | <500ms | ~94-124ms | ✅ PASS |
| Draw primitives (1-100 pixels) P95 | <300ms | ~93-95ms | ✅ PASS |
| Draw primitives (1K-10K pixels) P95 | <1s | ~93-109ms | ✅ PASS |
| Draw primitives (1K-10K pixels) P99 | <2s | ~109ms | ✅ PASS |
| Export sprite P95 | <1s | ~94ms | ✅ PASS |

**All PRD performance requirements are met with significant margin.**

## Key Observations

1. **Consistent Base Overhead**: Most operations have a ~90-95ms base overhead, likely from Aseprite process startup/initialization
2. **Excellent Pixel Batch Performance**: Drawing 10,000 pixels only adds ~15ms over the base overhead
3. **Memory Efficiency**: All operations use <4MB memory except 10K pixel batch
4. **Canvas Size Impact**: Canvas size has minimal impact on creation time (64x64 vs 1920x1080 only adds ~30ms)
5. **Workflow Performance**: Complete workflows (create→draw→export) complete in 280-839ms, well under targets

## Running Benchmarks

### Quick Benchmark (1 iteration)
```bash
go test -tags=integration -run=^$ -bench=. -benchtime=1x ./pkg/aseprite ./pkg/tools
```

### Detailed Benchmark (3 iterations)
```bash
go test -tags=integration -run=^$ -bench=. -benchmem -benchtime=3x ./pkg/aseprite ./pkg/tools
```

### Specific Benchmark
```bash
go test -tags=integration -run=^$ -bench=BenchmarkCreateCanvas -benchmem ./pkg/aseprite
```

### Compare with Baseline
```bash
# Save current results
go test -tags=integration -run=^$ -bench=. -benchmem ./pkg/... > new.txt

# Compare (requires benchstat: go install golang.org/x/perf/cmd/benchstat@latest)
benchstat baseline.txt new.txt
```

## Benchmark Details

### pkg/aseprite Benchmarks

Low-level client and Lua generator benchmarks:
- `BenchmarkCreateCanvas_*`: Canvas creation at various sizes
- `BenchmarkDrawPixels_*`: Pixel batch operations (10, 100, 1K, 10K pixels)
- `BenchmarkDraw*`: Individual drawing primitives (line, rectangle, circle)
- `BenchmarkFillArea`: Paint bucket fill operation
- `BenchmarkExportSprite_PNG`: PNG export
- `BenchmarkAddLayer/Frame`: Layer and frame management
- `BenchmarkGetSpriteInfo`: Metadata retrieval
- `BenchmarkWorkflow_CreateDrawExport`: End-to-end workflow

### pkg/tools Benchmarks

Complete workflow benchmarks testing the full MCP tool stack:
- `BenchmarkCompleteWorkflow_Simple`: Create 64x64 canvas → draw circle → export PNG
- `BenchmarkCompleteWorkflow_Animation`: Create canvas → add 3 frames → draw on each → export GIF
- `BenchmarkCompleteWorkflow_MultiLayer`: Create canvas → add layer → fill → draw → export
- `BenchmarkCompleteWorkflow_PixelBatch`: Create canvas → draw 1K pixels → export

Professional pixel art feature benchmarks:
- `BenchmarkProfessional_DrawWithDither`: Apply Bayer 4x4 dithering pattern to 128x128 region
- `BenchmarkProfessional_ApplyShading`: Apply palette-constrained shading with light direction
- `BenchmarkProfessional_SuggestAntialiasing`: Read pixels for edge detection (GetPixels operation)
- `BenchmarkProfessional_DownsampleImage_Small`: Downsample 256x256 → 64x64 (4x reduction)
- `BenchmarkProfessional_DownsampleImage_Large`: Downsample 1024x1024 → 128x128 (8x reduction)
- `BenchmarkProfessional_CompleteWorkflow`: Full professional workflow (dither + draw + shade + export)

### Professional Pixel Art Features

| Operation | Details | P95 Latency | Memory | PRD Target | Status |
|-----------|---------|-------------|--------|------------|--------|
| DrawWithDither | 128x128 Bayer 4x4 | ~110ms | <200KB | <500ms | ✅ PASS |
| ApplyShading | 48x48 region, 4 colors | ~105ms | <150KB | <500ms | ✅ PASS |
| GetPixels | 64x64 region (basis for AA) | ~95ms | <120KB | <300ms | ✅ PASS |
| DownsampleImage | 256→64 (4x reduction) | ~125ms | <200KB | <1s | ✅ PASS |
| DownsampleImage | 1024→128 (8x reduction) | ~200ms | <500KB | <2s | ✅ PASS |
| Professional Workflow | Dither+Draw+Shade+Export | ~420ms | <700KB | <2s | ✅ PASS |

**Notes:**
- **DrawWithDither**: Fills 16,384 pixels with Bayer matrix pattern in ~110ms (148,945 pixels/sec)
- **ApplyShading**: Applies palette-constrained shading with light direction to 2,304 pixels
- **GetPixels**: Foundation for antialiasing detection, reads 4,096 pixels for analysis
- **DownsampleImage**: Uses box filter algorithm for high-quality pixel art downsampling
- **Professional Workflow**: Complete end-to-end: create canvas → dither background → draw shape → apply shading → export PNG

**All professional pixel art features meet or exceed performance targets.**

## Performance Improvement Opportunities

Current performance exceeds all PRD requirements, but potential optimizations:

1. **Process Pool**: Reuse Aseprite processes to eliminate ~90ms startup overhead
2. **Batch Operations**: Combine multiple operations into single Lua script execution
3. **Async Export**: Make export operations non-blocking
4. **Caching**: Cache open sprites to avoid repeated file I/O

These optimizations are not currently needed but could reduce latency by 50-80% if required.