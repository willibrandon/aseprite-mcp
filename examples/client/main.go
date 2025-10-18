// Package main demonstrates how to use the Aseprite MCP server as a client.
// This example creates a sprite, draws on it, and exports it.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/mtlog/sinks"
)

func main() {
	logger := createLogger()
	if err := run(logger); err != nil {
		logger.Fatal("Application error: {Error}", err)
	}
}

func run(logger core.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger.Information("Aseprite MCP Client Example")
	logger.Information("===========================")
	logger.Information("")

	// Start the MCP server as a subprocess
	serverPath := os.Getenv("ASEPRITE_MCP_PATH")
	if serverPath == "" {
		// Try to find in common locations
		serverPath = findServerBinary()
		if serverPath == "" {
			return fmt.Errorf("ASEPRITE_MCP_PATH not set and could not find pixel-mcp binary")
		}
	}

	// On Windows, add .exe extension if missing
	if filepath.Ext(serverPath) == "" {
		if _, err := os.Stat(serverPath + ".exe"); err == nil {
			serverPath = serverPath + ".exe"
		} else if _, err := os.Stat(serverPath); os.IsNotExist(err) {
			// Try adding .exe if file doesn't exist
			serverPath = serverPath + ".exe"
		}
	}

	logger.Information("Starting server: {ServerPath}", serverPath)
	cmd := exec.Command(serverPath)

	// Create MCP client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "aseprite-example-client",
		Version: "1.0.0",
	}, nil)

	// Connect to server via command transport
	logger.Information("Connecting to server...")
	transport := &mcp.CommandTransport{Command: cmd}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer session.Close()

	logger.Information("Connected!")
	logger.Information("")

	// List available tools
	logger.Information("Available tools:")
	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}
	for _, tool := range tools.Tools {
		logger.Information("  - {Name}: {Description}", tool.Name, tool.Description)
	}
	logger.Information("")

	// Example workflow: Create a sprite with animation
	if err := createAnimatedSprite(ctx, session, logger); err != nil {
		return err
	}

	logger.Information("")
	logger.Information("Example completed successfully!")
	return nil
}

func createAnimatedSprite(ctx context.Context, session *mcp.ClientSession, logger core.Logger) error {
	outputDir := "../sprites"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Create canvas
	logger.Information("Step 1: Creating 64x64 RGB canvas...")
	createResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      64,
		"height":     64,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas failed: %w", err)
	}

	var createResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(createResp), &createResult); err != nil {
		return fmt.Errorf("failed to parse create result: %w", err)
	}
	spritePath := createResult.FilePath
	logger.Information("  Created: {SpritePath}", spritePath)

	// Step 2: Fill background with blue on Layer 1
	logger.Information("")
	logger.Information("Step 2: Filling background with blue...")
	if _, err := callTool(ctx, session, "fill_area", map[string]any{
		"sprite_path":  spritePath,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            32,
		"y":            32,
		"color":        "#0066CC",
		"tolerance":    0,
	}); err != nil {
		return fmt.Errorf("fill_area failed: %w", err)
	}
	logger.Information("  Background filled")

	// Step 3: Add 3 more frames for animation
	logger.Information("")
	logger.Information("Step 3: Adding 3 animation frames...")
	for i := 0; i < 3; i++ {
		if _, err := callTool(ctx, session, "add_frame", map[string]any{
			"sprite_path": spritePath,
			"duration_ms": 100,
		}); err != nil {
			return fmt.Errorf("add_frame failed: %w", err)
		}
	}
	logger.Information("  Frames added (4 total)")

	// Step 4: Draw circles on each frame (growing animation)
	logger.Information("")
	logger.Information("Step 4: Drawing animated circles...")
	colors := []string{"#FF0000", "#00FF00", "#FFFF00", "#FF00FF"}
	for frame := 1; frame <= 4; frame++ {
		radius := 5 + frame*3
		if _, err := callTool(ctx, session, "draw_circle", map[string]any{
			"sprite_path":  spritePath,
			"layer_name":   "Layer 1",
			"frame_number": frame,
			"center_x":     32,
			"center_y":     32,
			"radius":       radius,
			"color":        colors[frame-1],
			"filled":       true,
		}); err != nil {
			return fmt.Errorf("draw_circle frame %d failed: %w", frame, err)
		}
		logger.Information("  Frame {Frame}: radius {Radius}, color {Color}", frame, radius, colors[frame-1])
	}

	// Step 5: Read pixels to verify drawing (read center 10x10 region from frame 2)
	logger.Information("")
	logger.Information("Step 5: Reading pixels from frame 2 to verify drawing...")
	pixelsResp, err := callTool(ctx, session, "get_pixels", map[string]any{
		"sprite_path":  spritePath,
		"layer_name":   "Layer 1",
		"frame_number": 2,
		"x":            27,
		"y":            27,
		"width":        10,
		"height":       10,
	})
	if err != nil {
		return fmt.Errorf("get_pixels failed: %w", err)
	}
	var pixelsResult struct {
		Pixels []struct {
			X     int    `json:"x"`
			Y     int    `json:"y"`
			Color string `json:"color"`
		} `json:"pixels"`
	}
	if err := json.Unmarshal([]byte(pixelsResp), &pixelsResult); err != nil {
		return fmt.Errorf("failed to parse pixels result: %w", err)
	}
	logger.Information("  Read {Count} pixels from center region", len(pixelsResult.Pixels))
	// Count green pixels (frame 2 has green circle)
	greenCount := 0
	for _, p := range pixelsResult.Pixels {
		if p.Color == "#00FF00FF" {
			greenCount++
		}
	}
	logger.Information("  Found {GreenCount} green pixels in the region", greenCount)

	// Step 6: Get sprite info
	logger.Information("")
	logger.Information("Step 6: Getting sprite metadata...")
	infoResp, err := callTool(ctx, session, "get_sprite_info", map[string]any{
		"sprite_path": spritePath,
	})
	if err != nil {
		return fmt.Errorf("get_sprite_info failed: %w", err)
	}
	logger.Information("  Info: {Info}", infoResp)

	// Step 7: Export as GIF
	logger.Information("")
	logger.Information("Step 7: Exporting as GIF...")
	gifPath := filepath.Join(outputDir, "animated-example.gif")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  spritePath,
		"output_path":  gifPath,
		"format":       "gif",
		"frame_number": 0, // 0 = all frames
	}); err != nil {
		return fmt.Errorf("export_sprite failed: %w", err)
	}
	logger.Information("  Exported: {GifPath}", gifPath)

	// Step 8: Export frame 2 as PNG
	logger.Information("")
	logger.Information("Step 8: Exporting frame 2 as PNG...")
	pngPath := filepath.Join(outputDir, "frame2-example.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  spritePath,
		"output_path":  pngPath,
		"format":       "png",
		"frame_number": 2,
	}); err != nil {
		return fmt.Errorf("export_sprite frame failed: %w", err)
	}
	logger.Information("  Exported: {PngPath}", pngPath)

	// Step 9: Demonstrate dithering (create a new sprite with gradient)
	logger.Information("")
	logger.Information("Step 9: Creating sprite for dithering comparison (Bayer vs Floyd-Steinberg)...")
	ditherResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      128,
		"height":     64,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for dither failed: %w", err)
	}
	var ditherResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(ditherResp), &ditherResult); err != nil {
		return fmt.Errorf("failed to parse dither canvas result: %w", err)
	}
	ditherSprite := ditherResult.FilePath
	logger.Information("  Created: {DitherSprite}", ditherSprite)

	// Apply Bayer 4x4 dithering to left half
	logger.Information("")
	logger.Information("Step 10: Applying Bayer 4x4 pattern (left half)...")
	if _, err := callTool(ctx, session, "draw_with_dither", map[string]any{
		"sprite_path":  ditherSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"region": map[string]any{
			"x":      0,
			"y":      0,
			"width":  64,
			"height": 64,
		},
		"color1":  "#001F3F",
		"color2":  "#7FDBFF",
		"pattern": "bayer_4x4",
		"density": 0.5,
	}); err != nil {
		return fmt.Errorf("draw_with_dither (bayer) failed: %w", err)
	}
	logger.Information("  Bayer 4x4 pattern applied (ordered dithering)")

	// Apply Floyd-Steinberg dithering to right half
	logger.Information("  Applying Floyd-Steinberg pattern (right half)...")
	if _, err := callTool(ctx, session, "draw_with_dither", map[string]any{
		"sprite_path":  ditherSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"region": map[string]any{
			"x":      64,
			"y":      0,
			"width":  64,
			"height": 64,
		},
		"color1":  "#001F3F",
		"color2":  "#7FDBFF",
		"pattern": "floyd_steinberg",
		"density": 0.5,
	}); err != nil {
		return fmt.Errorf("draw_with_dither (floyd-steinberg) failed: %w", err)
	}
	logger.Information("  Floyd-Steinberg pattern applied (error diffusion)")

	// Export dithered sprite
	logger.Information("")
	logger.Information("Step 11: Exporting dithering comparison...")
	ditherPngPath := filepath.Join(outputDir, "dithering-comparison.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  ditherSprite,
		"output_path":  ditherPngPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export_sprite dither failed: %w", err)
	}
	logger.Information("  Exported: {DitherPng}", ditherPngPath)
	logger.Information("  Left: Bayer 4x4 (ordered pattern) | Right: Floyd-Steinberg (error diffusion)")

	// Step 12: Analyze palette harmonies
	logger.Information("")
	logger.Information("Step 12: Analyzing palette harmonies from our colors...")
	harmonyResp, err := callTool(ctx, session, "analyze_palette_harmonies", map[string]any{
		"palette": []string{"#FF0000", "#00FF00", "#FFFF00", "#FF00FF", "#0066CC"},
	})
	if err != nil {
		return fmt.Errorf("analyze_palette_harmonies failed: %w", err)
	}
	var harmonyResult struct {
		Complementary []struct {
			Color1 string `json:"color1"`
			Color2 string `json:"color2"`
		} `json:"complementary"`
		Temperature struct {
			Dominant string `json:"dominant"`
		} `json:"temperature"`
	}
	if err := json.Unmarshal([]byte(harmonyResp), &harmonyResult); err != nil {
		return fmt.Errorf("failed to parse harmony result: %w", err)
	}
	logger.Information("  Dominant temperature: {Temp}", harmonyResult.Temperature.Dominant)
	if len(harmonyResult.Complementary) > 0 {
		logger.Information("  Found {Count} complementary pairs", len(harmonyResult.Complementary))
	}

	// Step 13: Create sprite with custom palette
	logger.Information("")
	logger.Information("Step 13: Creating sprite with limited palette...")
	paletteResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      32,
		"height":     32,
		"color_mode": "indexed",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for palette failed: %w", err)
	}
	var paletteResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(paletteResp), &paletteResult); err != nil {
		return fmt.Errorf("failed to parse palette canvas result: %w", err)
	}
	paletteSprite := paletteResult.FilePath
	logger.Information("  Created: {PaletteSprite}", paletteSprite)

	// Set a custom 8-color palette
	if _, err := callTool(ctx, session, "set_palette", map[string]any{
		"sprite_path": paletteSprite,
		"colors":      []string{"#000000", "#1D2B53", "#7E2553", "#008751", "#AB5236", "#5F574F", "#C2C3C7", "#FFF1E8"},
	}); err != nil {
		return fmt.Errorf("set_palette failed: %w", err)
	}
	logger.Information("  Palette set successfully (8 colors)")

	// Step 14: Apply palette-constrained shading
	logger.Information("")
	logger.Information("Step 14: Drawing shape with palette-constrained shading...")

	// Create a larger 64x64 sprite for better visibility
	shadingResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      64,
		"height":     64,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for shading failed: %w", err)
	}
	var shadingResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(shadingResp), &shadingResult); err != nil {
		return fmt.Errorf("failed to parse shading canvas result: %w", err)
	}
	shadingSprite := shadingResult.FilePath

	// Draw a larger base circle
	if _, err := callTool(ctx, session, "draw_circle", map[string]any{
		"sprite_path":  shadingSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"center_x":     32,
		"center_y":     32,
		"radius":       28,
		"color":        "#AB5236",
		"filled":       true,
	}); err != nil {
		return fmt.Errorf("draw_circle for shading failed: %w", err)
	}

	// Apply shading with wider palette range for more visible effect
	if _, err := callTool(ctx, session, "apply_shading", map[string]any{
		"sprite_path":  shadingSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"region": map[string]any{
			"x":      2,
			"y":      2,
			"width":  60,
			"height": 60,
		},
		"palette":         []string{"#1D2B53", "#7E2553", "#AB5236", "#FFB380", "#FFF1E8"},
		"light_direction": "top_left",
		"intensity":       0.9,
		"style":           "smooth",
	}); err != nil {
		return fmt.Errorf("apply_shading failed: %w", err)
	}
	logger.Information("  Shading applied successfully")

	// Export shaded sprite
	shadedPngPath := filepath.Join(outputDir, "shaded-sphere.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  shadingSprite,
		"output_path":  shadedPngPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export_sprite shaded failed: %w", err)
	}
	logger.Information("  Exported: {ShadedPng}", shadedPngPath)

	// Step 15: Demonstrate palette-aware drawing (use_palette flag)
	logger.Information("")
	logger.Information("Step 15: Demonstrating palette-aware drawing...")
	paletteDrawResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      96,
		"height":     32,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for palette draw failed: %w", err)
	}
	var paletteDrawResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(paletteDrawResp), &paletteDrawResult); err != nil {
		return fmt.Errorf("failed to parse palette draw canvas result: %w", err)
	}
	paletteDrawSprite := paletteDrawResult.FilePath
	logger.Information("  Created: {PaletteDrawSprite}", paletteDrawSprite)

	// Set a simple 4-color palette (black, red, green, blue)
	if _, err := callTool(ctx, session, "set_palette", map[string]any{
		"sprite_path": paletteDrawSprite,
		"colors":      []string{"#000000", "#FF0000", "#00FF00", "#0000FF"},
	}); err != nil {
		return fmt.Errorf("set_palette for draw demo failed: %w", err)
	}

	// Draw rectangles WITHOUT palette snapping (left side) - exact pastel colors
	if _, err := callTool(ctx, session, "draw_rectangle", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            4,
		"y":            4,
		"width":        12,
		"height":       8,
		"color":        "#FF8080", // Pastel red
		"filled":       true,
		"use_palette":  false,
	}); err != nil {
		return fmt.Errorf("draw_rectangle without palette (red) failed: %w", err)
	}

	if _, err := callTool(ctx, session, "draw_rectangle", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            20,
		"y":            4,
		"width":        12,
		"height":       8,
		"color":        "#80FF80", // Pastel green
		"filled":       true,
		"use_palette":  false,
	}); err != nil {
		return fmt.Errorf("draw_rectangle without palette (green) failed: %w", err)
	}

	if _, err := callTool(ctx, session, "draw_rectangle", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            36,
		"y":            4,
		"width":        12,
		"height":       8,
		"color":        "#8080FF", // Pastel blue
		"filled":       true,
		"use_palette":  false,
	}); err != nil {
		return fmt.Errorf("draw_rectangle without palette (blue) failed: %w", err)
	}

	// Draw rectangles WITH palette snapping (right side) - snaps to pure colors
	if _, err := callTool(ctx, session, "draw_rectangle", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            52,
		"y":            4,
		"width":        12,
		"height":       8,
		"color":        "#FF8080", // Pastel red → snaps to #FF0000 (pure red)
		"filled":       true,
		"use_palette":  true,
	}); err != nil {
		return fmt.Errorf("draw_rectangle with palette (red) failed: %w", err)
	}

	if _, err := callTool(ctx, session, "draw_rectangle", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            68,
		"y":            4,
		"width":        12,
		"height":       8,
		"color":        "#80FF80", // Pastel green → snaps to #00FF00 (pure green)
		"filled":       true,
		"use_palette":  true,
	}); err != nil {
		return fmt.Errorf("draw_rectangle with palette (green) failed: %w", err)
	}

	if _, err := callTool(ctx, session, "draw_rectangle", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            84,
		"y":            4,
		"width":        12,
		"height":       8,
		"color":        "#8080FF", // Pastel blue → snaps to #0000FF (pure blue)
		"filled":       true,
		"use_palette":  true,
	}); err != nil {
		return fmt.Errorf("draw_rectangle with palette (blue) failed: %w", err)
	}

	// Add labels with lines
	if _, err := callTool(ctx, session, "draw_line", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x1":           4,
		"y1":           20,
		"x2":           44,
		"y2":           20,
		"color":        "#FFFFFF",
		"thickness":    1,
		"use_palette":  false,
	}); err != nil {
		return fmt.Errorf("draw_line (left label) failed: %w", err)
	}

	if _, err := callTool(ctx, session, "draw_line", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x1":           52,
		"y1":           20,
		"x2":           92,
		"y2":           20,
		"color":        "#FFFFFF",
		"thickness":    1,
		"use_palette":  false,
	}); err != nil {
		return fmt.Errorf("draw_line (right label) failed: %w", err)
	}

	logger.Information("  Palette-aware drawing completed (left: pastel colors, right: snapped to pure colors)")

	// Export palette demo
	paletteDrawPngPath := filepath.Join(outputDir, "palette-drawing-comparison.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  paletteDrawSprite,
		"output_path":  paletteDrawPngPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export_sprite palette draw failed: %w", err)
	}
	logger.Information("  Exported: {PaletteDrawPng}", paletteDrawPngPath)

	// Step 16: Demonstrate antialiasing for smooth diagonal edges
	logger.Information("")
	logger.Information("Step 16: Demonstrating antialiasing suggestions...")

	// Create a sprite with jagged diagonal edges
	aaResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      64,
		"height":     64,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for antialiasing failed: %w", err)
	}
	var aaResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(aaResp), &aaResult); err != nil {
		return fmt.Errorf("failed to parse antialiasing canvas result: %w", err)
	}
	aaSprite := aaResult.FilePath
	logger.Information("  Created: {AASprite}", aaSprite)

	// Draw a jagged diagonal line (stair-step pattern)
	// Pattern:   ....####
	//            ...####.
	//            ..####..
	//            .####...
	//            ####....
	jaggedPixels := []map[string]any{}
	for i := 0; i < 5; i++ {
		for j := 0; j < 4; j++ {
			jaggedPixels = append(jaggedPixels, map[string]any{
				"x":     20 + i + j,
				"y":     10 + i,
				"color": "#FF00FFFF", // Magenta
			})
		}
	}

	if _, err := callTool(ctx, session, "draw_pixels", map[string]any{
		"sprite_path":  aaSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"pixels":       jaggedPixels,
		"use_palette":  false,
	}); err != nil {
		return fmt.Errorf("draw_pixels for jagged line failed: %w", err)
	}

	// Get antialiasing suggestions (without auto-apply first to see what it suggests)
	suggestResp, err := callTool(ctx, session, "suggest_antialiasing", map[string]any{
		"sprite_path":  aaSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"threshold":    128,
		"auto_apply":   false,
		"use_palette":  false,
	})
	if err != nil {
		return fmt.Errorf("suggest_antialiasing failed: %w", err)
	}

	var aaAnalysis struct {
		Suggestions []struct {
			X              int    `json:"x"`
			Y              int    `json:"y"`
			CurrentColor   string `json:"current_color"`
			NeighborColor  string `json:"neighbor_color"`
			SuggestedColor string `json:"suggested_color"`
			Direction      string `json:"direction"`
		} `json:"suggestions"`
		Applied    bool `json:"applied"`
		TotalEdges int  `json:"total_edges"`
	}
	if err := json.Unmarshal([]byte(suggestResp), &aaAnalysis); err != nil {
		return fmt.Errorf("failed to parse antialiasing result: %w", err)
	}

	logger.Information("  Found {EdgeCount} jagged edge positions", aaAnalysis.TotalEdges)
	if len(aaAnalysis.Suggestions) > 0 && len(aaAnalysis.Suggestions) <= 3 {
		for i, sug := range aaAnalysis.Suggestions {
			logger.Information("    - Suggestion {Index}: pos=({X},{Y}) direction={Direction}",
				i+1, sug.X, sug.Y, sug.Direction)
		}
	}

	// Export the jagged version
	jaggedPngPath := filepath.Join(outputDir, "antialiasing-before.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  aaSprite,
		"output_path":  jaggedPngPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export_sprite jagged failed: %w", err)
	}

	// Now apply antialiasing automatically to smooth the edges
	if _, err := callTool(ctx, session, "suggest_antialiasing", map[string]any{
		"sprite_path":  aaSprite,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"threshold":    128,
		"auto_apply":   true, // Apply smoothing automatically
		"use_palette":  false,
	}); err != nil {
		return fmt.Errorf("suggest_antialiasing with auto_apply failed: %w", err)
	}

	// Export the smoothed version
	smoothPngPath := filepath.Join(outputDir, "antialiasing-after.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  aaSprite,
		"output_path":  smoothPngPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export_sprite smooth failed: %w", err)
	}

	logger.Information("  Antialiasing applied: jagged diagonal smoothed")
	logger.Information("  Exported before: {JaggedPng}", jaggedPngPath)
	logger.Information("  Exported after: {SmoothPng}", smoothPngPath)

	// Step 17: Demonstrate layer and frame deletion
	logger.Information("")
	logger.Information("Step 17: Demonstrating layer and frame deletion...")

	// Create a test sprite with multiple layers and frames
	deleteResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      32,
		"height":     32,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for deletion demo failed: %w", err)
	}
	var deleteResult struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(deleteResp), &deleteResult); err != nil {
		return fmt.Errorf("failed to parse deletion canvas result: %w", err)
	}
	deleteSprite := deleteResult.FilePath
	logger.Information("  Created: {DeleteSprite}", deleteSprite)

	// Add two extra layers
	if _, err := callTool(ctx, session, "add_layer", map[string]any{
		"sprite_path": deleteSprite,
		"layer_name":  "Layer 2",
	}); err != nil {
		return fmt.Errorf("add_layer Layer 2 failed: %w", err)
	}

	if _, err := callTool(ctx, session, "add_layer", map[string]any{
		"sprite_path": deleteSprite,
		"layer_name":  "Layer 3",
	}); err != nil {
		return fmt.Errorf("add_layer Layer 3 failed: %w", err)
	}
	logger.Information("  Added 2 extra layers (3 total)")

	// Add two extra frames
	if _, err := callTool(ctx, session, "add_frame", map[string]any{
		"sprite_path": deleteSprite,
		"duration_ms": 100,
	}); err != nil {
		return fmt.Errorf("add_frame for deletion demo failed: %w", err)
	}

	if _, err := callTool(ctx, session, "add_frame", map[string]any{
		"sprite_path": deleteSprite,
		"duration_ms": 100,
	}); err != nil {
		return fmt.Errorf("add_frame for deletion demo failed: %w", err)
	}
	logger.Information("  Added 2 extra frames (3 total)")

	// Delete Layer 2
	if _, err := callTool(ctx, session, "delete_layer", map[string]any{
		"sprite_path": deleteSprite,
		"layer_name":  "Layer 2",
	}); err != nil {
		return fmt.Errorf("delete_layer failed: %w", err)
	}
	logger.Information("  Deleted Layer 2 (2 layers remaining)")

	// Delete frame 2
	if _, err := callTool(ctx, session, "delete_frame", map[string]any{
		"sprite_path":  deleteSprite,
		"frame_number": 2,
	}); err != nil {
		return fmt.Errorf("delete_frame failed: %w", err)
	}
	logger.Information("  Deleted frame 2 (2 frames remaining)")

	// Verify final state
	finalInfoResp, err := callTool(ctx, session, "get_sprite_info", map[string]any{
		"sprite_path": deleteSprite,
	})
	if err != nil {
		return fmt.Errorf("get_sprite_info after deletion failed: %w", err)
	}
	logger.Information("  Final state: {Info}", finalInfoResp)

	// Step 18: Demonstrate drawing polylines and polygons
	logger.Information("")
	logger.Information("Step 18: Demonstrating polylines and polygons...")

	// Draw a zigzag polyline (open contour) on frame 1
	logger.Information("  Drawing zigzag polyline on frame 1...")
	_, err = callTool(ctx, session, "draw_contour", map[string]any{
		"sprite_path":  spritePath,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"points": []map[string]any{
			{"x": 10, "y": 10},
			{"x": 30, "y": 30},
			{"x": 50, "y": 10},
			{"x": 70, "y": 30},
			{"x": 54, "y": 54},
		},
		"color":     "#FF0000",
		"thickness": 2,
		"closed":    false,
	})
	if err != nil {
		return fmt.Errorf("draw_contour (zigzag) failed: %w", err)
	}

	// Draw a triangle (closed polygon) on frame 2
	logger.Information("  Drawing triangle on frame 2...")
	_, err = callTool(ctx, session, "draw_contour", map[string]any{
		"sprite_path":  spritePath,
		"layer_name":   "Layer 1",
		"frame_number": 2,
		"points": []map[string]any{
			{"x": 32, "y": 10},
			{"x": 54, "y": 54},
			{"x": 10, "y": 54},
		},
		"color":     "#00FF00",
		"thickness": 3,
		"closed":    true,
	})
	if err != nil {
		return fmt.Errorf("draw_contour (triangle) failed: %w", err)
	}

	// Draw a star shape on frame 3 with palette snapping
	logger.Information("  Drawing star on frame 3 with palette snapping...")
	_, err = callTool(ctx, session, "draw_contour", map[string]any{
		"sprite_path":  spritePath,
		"layer_name":   "Layer 1",
		"frame_number": 3,
		"points": []map[string]any{
			{"x": 32, "y": 5},
			{"x": 36, "y": 22},
			{"x": 54, "y": 22},
			{"x": 40, "y": 33},
			{"x": 46, "y": 50},
			{"x": 32, "y": 40},
			{"x": 18, "y": 50},
			{"x": 24, "y": 33},
			{"x": 10, "y": 22},
			{"x": 28, "y": 22},
		},
		"color":       "#FFFF00", // Yellow
		"thickness":   1,
		"closed":      true,
		"use_palette": false, // No palette snapping for now
	})
	if err != nil {
		return fmt.Errorf("draw_contour (star) failed: %w", err)
	}
	logger.Information("  ✓ Drew polylines and polygons successfully")

	// Step 19: Demonstrate palette management tools
	logger.Information("")
	logger.Information("Step 19: Demonstrating palette management tools...")

	// Create an indexed color sprite for palette operations
	logger.Information("  Creating indexed color sprite for palette demo...")
	output, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      64,
		"height":     64,
		"color_mode": "indexed",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for palette failed: %w", err)
	}

	// Parse the output to get sprite path
	var paletteCreateOutput struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(output), &paletteCreateOutput); err != nil {
		return fmt.Errorf("failed to parse create_canvas output: %w", err)
	}
	paletteSpritePath := paletteCreateOutput.FilePath
	defer os.Remove(paletteSpritePath)

	// Set a custom palette
	logger.Information("  Setting custom 8-color palette...")
	_, err = callTool(ctx, session, "set_palette", map[string]any{
		"sprite_path": paletteSpritePath,
		"colors": []string{
			"#000000", // Black
			"#FFFFFF", // White
			"#FF0000", // Red
			"#00FF00", // Green
			"#0000FF", // Blue
			"#FFFF00", // Yellow
			"#FF00FF", // Magenta
			"#00FFFF", // Cyan
		},
	})
	if err != nil {
		return fmt.Errorf("set_palette failed: %w", err)
	}
	logger.Information("  ✓ Custom palette set successfully")

	// Get the palette back
	logger.Information("  Retrieving palette with get_palette...")
	paletteOutput, err := callTool(ctx, session, "get_palette", map[string]any{
		"sprite_path": paletteSpritePath,
	})
	if err != nil {
		return fmt.Errorf("get_palette failed: %w", err)
	}

	var getPaletteResult struct {
		Colors []string `json:"colors"`
		Size   int      `json:"size"`
	}
	if err := json.Unmarshal([]byte(paletteOutput), &getPaletteResult); err != nil {
		return fmt.Errorf("failed to parse get_palette output: %w", err)
	}
	logger.Information("  ✓ Retrieved palette with {Count} colors: {Colors}", getPaletteResult.Size, getPaletteResult.Colors)

	// Modify a specific palette color
	logger.Information("  Changing color at index 2 to orange (#FF8000)...")
	_, err = callTool(ctx, session, "set_palette_color", map[string]any{
		"sprite_path": paletteSpritePath,
		"index":       2,
		"color":       "#FF8000",
	})
	if err != nil {
		return fmt.Errorf("set_palette_color failed: %w", err)
	}
	logger.Information("  ✓ Palette color updated successfully")

	// Add a new color to the palette
	logger.Information("  Adding new color (brown #8B4513) to palette...")
	addColorOutput, err := callTool(ctx, session, "add_palette_color", map[string]any{
		"sprite_path": paletteSpritePath,
		"color":       "#8B4513",
	})
	if err != nil {
		return fmt.Errorf("add_palette_color failed: %w", err)
	}

	var addColorResult struct {
		ColorIndex int `json:"color_index"`
	}
	if err := json.Unmarshal([]byte(addColorOutput), &addColorResult); err != nil {
		return fmt.Errorf("failed to parse add_palette_color output: %w", err)
	}
	logger.Information("  ✓ Added color at index {Index}", addColorResult.ColorIndex)

	// Sort the palette by hue
	logger.Information("  Sorting palette by hue (ascending)...")
	_, err = callTool(ctx, session, "sort_palette", map[string]any{
		"sprite_path": paletteSpritePath,
		"method":      "hue",
		"ascending":   true,
	})
	if err != nil {
		return fmt.Errorf("sort_palette failed: %w", err)
	}
	logger.Information("  ✓ Palette sorted by hue")

	// Get the sorted palette
	logger.Information("  Retrieving sorted palette...")
	sortedPaletteOutput, err := callTool(ctx, session, "get_palette", map[string]any{
		"sprite_path": paletteSpritePath,
	})
	if err != nil {
		return fmt.Errorf("get_palette (after sort) failed: %w", err)
	}

	var getSortedPaletteResult struct {
		Colors []string `json:"colors"`
		Size   int      `json:"size"`
	}
	if err := json.Unmarshal([]byte(sortedPaletteOutput), &getSortedPaletteResult); err != nil {
		return fmt.Errorf("failed to parse sorted palette output: %w", err)
	}
	logger.Information("  ✓ Sorted palette has {Count} colors: {Colors}", getSortedPaletteResult.Size, getSortedPaletteResult.Colors)

	// Step 20: Demonstrate transform operations
	logger.Information("")
	logger.Information("Step 20: Demonstrating transform operations (flip, rotate, scale)...")

	// Create a sprite with asymmetric content for transforms
	logger.Information("  Creating sprite for transform demo...")
	output, err = callTool(ctx, session, "create_canvas", map[string]any{
		"width":      64,
		"height":     64,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for transform failed: %w", err)
	}

	// Parse the output to get sprite path
	var transformCreateOutput struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(output), &transformCreateOutput); err != nil {
		return fmt.Errorf("failed to parse create_canvas output: %w", err)
	}
	transformSpritePath := transformCreateOutput.FilePath
	defer os.Remove(transformSpritePath)

	// Draw an asymmetric shape (triangle pointing right)
	logger.Information("  Drawing triangle pointing right...")
	_, err = callTool(ctx, session, "draw_contour", map[string]any{
		"sprite_path":  transformSpritePath,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"points": []map[string]any{
			{"x": 20, "y": 10},
			{"x": 50, "y": 32},
			{"x": 20, "y": 54},
		},
		"color":     "#FF0000",
		"thickness": 2,
		"closed":    true,
	})
	if err != nil {
		return fmt.Errorf("draw_contour for transform failed: %w", err)
	}

	// Flip horizontally
	logger.Information("  Flipping sprite horizontally...")
	_, err = callTool(ctx, session, "flip_sprite", map[string]any{
		"sprite_path": transformSpritePath,
		"direction":   "horizontal",
		"target":      "sprite",
	})
	if err != nil {
		return fmt.Errorf("flip_sprite failed: %w", err)
	}
	logger.Information("  ✓ Triangle now points left")

	// Rotate 90 degrees
	logger.Information("  Rotating sprite 90 degrees...")
	_, err = callTool(ctx, session, "rotate_sprite", map[string]any{
		"sprite_path": transformSpritePath,
		"angle":       90,
		"target":      "sprite",
	})
	if err != nil {
		return fmt.Errorf("rotate_sprite failed: %w", err)
	}
	logger.Information("  ✓ Sprite rotated 90° clockwise")

	// Scale 2x with nearest neighbor
	logger.Information("  Scaling sprite 2x with nearest neighbor...")
	scaleOutput, err := callTool(ctx, session, "scale_sprite", map[string]any{
		"sprite_path": transformSpritePath,
		"scale_x":     2.0,
		"scale_y":     2.0,
		"algorithm":   "nearest",
	})
	if err != nil {
		return fmt.Errorf("scale_sprite failed: %w", err)
	}

	var scaleResult struct {
		Success   bool `json:"success"`
		NewWidth  int  `json:"new_width"`
		NewHeight int  `json:"new_height"`
	}
	if err := json.Unmarshal([]byte(scaleOutput), &scaleResult); err != nil {
		return fmt.Errorf("failed to parse scale result: %w", err)
	}
	logger.Information("  ✓ Scaled to {Width}x{Height}", scaleResult.NewWidth, scaleResult.NewHeight)

	// Export transformed sprite
	transformPngPath := filepath.Join(outputDir, "transform-demo.png")
	_, err = callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  transformSpritePath,
		"output_path":  transformPngPath,
		"format":       "png",
		"frame_number": 0,
	})
	if err != nil {
		return fmt.Errorf("export_sprite transform failed: %w", err)
	}
	logger.Information("  Exported: {TransformPng}", transformPngPath)

	// Step 21: Demonstrate selection and clipboard operations
	logger.Information("")
	logger.Information("Step 21: Demonstrating selection and clipboard operations...")

	// Create a new sprite for selection demo
	logger.Information("  Creating sprite for selection demo...")
	output, err = callTool(ctx, session, "create_canvas", map[string]any{
		"width":      100,
		"height":     100,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for selection failed: %w", err)
	}

	// Parse the output to get sprite path
	var selectionCreateOutput struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(output), &selectionCreateOutput); err != nil {
		return fmt.Errorf("failed to parse create_canvas output: %w", err)
	}
	selectionSpritePath := selectionCreateOutput.FilePath
	defer os.Remove(selectionSpritePath)

	// Draw some content to copy
	logger.Information("  Drawing red square (20x20 at 20,20)...")
	_, err = callTool(ctx, session, "draw_rectangle", map[string]any{
		"sprite_path":  selectionSpritePath,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            20,
		"y":            20,
		"width":        20,
		"height":       20,
		"color":        "#FF0000",
		"filled":       true,
	})
	if err != nil {
		return fmt.Errorf("draw_rectangle failed: %w", err)
	}

	// Note: Selection and clipboard operations persist across MCP tool calls using sprite custom properties (sprite.data)
	// and a hidden clipboard layer. This allows you to create a selection in one tool call, then copy/cut/paste in
	// subsequent calls. Selection bounds are automatically saved and restored, and clipboard content is stored in a
	// hidden __mcp_clipboard__ layer.

	// For demonstration, let's show how to use drawing tools to achieve copy/paste effect
	logger.Information("  Copying red square to position (60, 60) using draw_rectangle...")
	_, err = callTool(ctx, session, "draw_rectangle", map[string]any{
		"sprite_path":  selectionSpritePath,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"x":            60,
		"y":            60,
		"width":        20,
		"height":       20,
		"color":        "#FF0000",
		"filled":       true,
	})
	if err != nil {
		return fmt.Errorf("draw second rectangle failed: %w", err)
	}

	// Demonstrate circle drawing
	logger.Information("  Drawing blue circle (radius 15 at 30,80)...")
	_, err = callTool(ctx, session, "draw_circle", map[string]any{
		"sprite_path":  selectionSpritePath,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"center_x":     30,
		"center_y":     80,
		"radius":       15,
		"color":        "#0000FF",
		"filled":       true,
	})
	if err != nil {
		return fmt.Errorf("draw_circle failed: %w", err)
	}

	// Export selection demo result
	selectionOutputPath := filepath.Join(os.TempDir(), "selection-demo.png")
	logger.Information("  Exporting selection demo to: {OutputPath}", selectionOutputPath)
	_, err = callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  selectionSpritePath,
		"output_path":  selectionOutputPath,
		"format":       "png",
		"frame_number": 0, // 0 = all frames
	})
	if err != nil {
		return fmt.Errorf("export_sprite failed: %w", err)
	}

	logger.Information("  ✓ Drawing operations completed successfully")
	logger.Information("  ✓ Result saved to: {OutputPath}", selectionOutputPath)
	logger.Information("  Note: Selection and clipboard operations now persist across MCP tool calls using sprite.data and hidden clipboard layer")

	// Step 22: Demonstrate advanced export tools
	logger.Information("")
	logger.Information("Step 22: Demonstrating advanced export tools (spritesheet, import, save_as, delete_tag)...")

	// Create an animated sprite for spritesheet export
	logger.Information("  Creating 16x16 animated sprite with 4 frames...")
	output, err = callTool(ctx, session, "create_canvas", map[string]any{
		"width":      16,
		"height":     16,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for spritesheet failed: %w", err)
	}

	var sheetCreateOutput struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(output), &sheetCreateOutput); err != nil {
		return fmt.Errorf("failed to parse create_canvas output: %w", err)
	}
	sheetSpritePath := sheetCreateOutput.FilePath
	defer os.Remove(sheetSpritePath)

	// Add 3 more frames (total 4)
	for i := 0; i < 3; i++ {
		_, err = callTool(ctx, session, "add_frame", map[string]any{
			"sprite_path": sheetSpritePath,
			"duration_ms": 100,
		})
		if err != nil {
			return fmt.Errorf("add_frame failed: %w", err)
		}
	}

	// Draw different colored circles on each frame
	colors2 := []string{"#FF0000", "#00FF00", "#0000FF", "#FFFF00"}
	for i := 0; i < 4; i++ {
		_, err = callTool(ctx, session, "draw_circle", map[string]any{
			"sprite_path":  sheetSpritePath,
			"layer_name":   "Layer 1",
			"frame_number": i + 1,
			"center_x":     8,
			"center_y":     8,
			"radius":       4 + i,
			"color":        colors2[i],
			"filled":       true,
		})
		if err != nil {
			return fmt.Errorf("draw_circle frame %d failed: %w", i+1, err)
		}
	}

	// Create animation tags
	logger.Information("  Creating animation tags...")
	_, err = callTool(ctx, session, "create_tag", map[string]any{
		"sprite_path": sheetSpritePath,
		"tag_name":    "grow",
		"from_frame":  1,
		"to_frame":    4,
		"direction":   "forward",
	})
	if err != nil {
		return fmt.Errorf("create_tag failed: %w", err)
	}

	_, err = callTool(ctx, session, "create_tag", map[string]any{
		"sprite_path": sheetSpritePath,
		"tag_name":    "temp_tag",
		"from_frame":  1,
		"to_frame":    2,
		"direction":   "pingpong",
	})
	if err != nil {
		return fmt.Errorf("create second tag failed: %w", err)
	}

	// Export as spritesheet with JSON metadata
	sheetPath := filepath.Join(outputDir, "animation-spritesheet.png")
	logger.Information("  Exporting as horizontal spritesheet with JSON metadata...")
	sheetOutput, err := callTool(ctx, session, "export_spritesheet", map[string]any{
		"sprite_path":  sheetSpritePath,
		"output_path":  sheetPath,
		"layout":       "horizontal",
		"padding":      2,
		"include_json": true,
	})
	if err != nil {
		return fmt.Errorf("export_spritesheet failed: %w", err)
	}

	var sheetResult struct {
		SpritesheetPath string  `json:"spritesheet_path"`
		MetadataPath    *string `json:"metadata_path"`
		FrameCount      int     `json:"frame_count"`
	}
	if err := json.Unmarshal([]byte(sheetOutput), &sheetResult); err != nil {
		return fmt.Errorf("failed to parse export_spritesheet output: %w", err)
	}
	logger.Information("  ✓ Exported spritesheet: {Path} ({Count} frames)", sheetResult.SpritesheetPath, sheetResult.FrameCount)
	if sheetResult.MetadataPath != nil {
		logger.Information("  ✓ JSON metadata: {Path}", *sheetResult.MetadataPath)
	}

	// Delete the temporary tag
	logger.Information("  Deleting temporary tag 'temp_tag'...")
	_, err = callTool(ctx, session, "delete_tag", map[string]any{
		"sprite_path": sheetSpritePath,
		"tag_name":    "temp_tag",
	})
	if err != nil {
		return fmt.Errorf("delete_tag failed: %w", err)
	}
	logger.Information("  ✓ Tag deleted successfully")

	// Save the sprite to a new location
	savedSpritePath := filepath.Join(outputDir, "saved-animation.aseprite")
	logger.Information("  Saving sprite to new location...")
	saveOutput, err := callTool(ctx, session, "save_as", map[string]any{
		"sprite_path": sheetSpritePath,
		"output_path": savedSpritePath,
	})
	if err != nil {
		return fmt.Errorf("save_as failed: %w", err)
	}

	var saveResult struct {
		Success  bool   `json:"success"`
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(saveOutput), &saveResult); err != nil {
		return fmt.Errorf("failed to parse save_as output: %w", err)
	}
	logger.Information("  ✓ Sprite saved to: {Path}", saveResult.FilePath)

	// Import the spritesheet back into a new sprite
	logger.Information("  Creating new sprite to import spritesheet...")
	output, err = callTool(ctx, session, "create_canvas", map[string]any{
		"width":      64,
		"height":     32,
		"color_mode": "rgb",
	})
	if err != nil {
		return fmt.Errorf("create_canvas for import failed: %w", err)
	}

	var importCreateOutput struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(output), &importCreateOutput); err != nil {
		return fmt.Errorf("failed to parse create_canvas output: %w", err)
	}
	importSpritePath := importCreateOutput.FilePath
	defer os.Remove(importSpritePath)

	logger.Information("  Importing spritesheet as a layer...")
	_, err = callTool(ctx, session, "import_image", map[string]any{
		"sprite_path":  importSpritePath,
		"image_path":   sheetPath,
		"layer_name":   "Spritesheet",
		"frame_number": 1,
		"position": map[string]any{
			"x": 0,
			"y": 0,
		},
	})
	if err != nil {
		return fmt.Errorf("import_image failed: %w", err)
	}
	logger.Information("  ✓ Spritesheet imported as layer")

	// Export the result
	importResultPath := filepath.Join(outputDir, "imported-spritesheet.png")
	_, err = callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  importSpritePath,
		"output_path":  importResultPath,
		"format":       "png",
		"frame_number": 0,
	})
	if err != nil {
		return fmt.Errorf("export imported sprite failed: %w", err)
	}
	logger.Information("  ✓ Exported imported result: {Path}", importResultPath)

	return nil
}

func callTool(ctx context.Context, session *mcp.ClientSession, name string, args map[string]any) (string, error) {
	resp, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return "", fmt.Errorf("tool call failed: %w", err)
	}

	if resp.IsError {
		return "", fmt.Errorf("tool returned error")
	}

	// Extract text content from response
	if len(resp.Content) > 0 {
		if textContent, ok := resp.Content[0].(*mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in response")
}

func findServerBinary() string {
	// Check current directory
	candidates := []string{
		"./bin/pixel-mcp",
		"./bin/pixel-mcp.exe",
		"../../bin/pixel-mcp",
		"../../bin/pixel-mcp.exe",
		"pixel-mcp",
		"pixel-mcp.exe",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// Check PATH
	if path, err := exec.LookPath("pixel-mcp"); err == nil {
		return path
	}

	return ""
}

// createLogger creates a configured logger instance.
func createLogger() core.Logger {
	sink := sinks.NewConsoleSink()
	logger := mtlog.New(
		mtlog.WithSink(sink),
		mtlog.WithMinimumLevel(core.InformationLevel),
	)
	return logger
}
