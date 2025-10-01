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
			return fmt.Errorf("ASEPRITE_MCP_PATH not set and could not find aseprite-mcp binary")
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

	// Step 2: Add a background layer
	logger.Information("")
	logger.Information("Step 2: Adding 'Background' layer...")
	if _, err := callTool(ctx, session, "add_layer", map[string]any{
		"sprite_path": spritePath,
		"layer_name":  "Background",
	}); err != nil {
		return fmt.Errorf("add_layer failed: %w", err)
	}
	logger.Information("  Layer added")

	// Step 3: Fill background with blue
	logger.Information("")
	logger.Information("Step 3: Filling background with blue...")
	if _, err := callTool(ctx, session, "fill_area", map[string]any{
		"sprite_path":  spritePath,
		"layer_name":   "Background",
		"frame_number": 1,
		"x":            32,
		"y":            32,
		"color":        "#0066CC",
		"tolerance":    0,
	}); err != nil {
		return fmt.Errorf("fill_area failed: %w", err)
	}
	logger.Information("  Background filled")

	// Step 4: Add 3 more frames for animation
	logger.Information("")
	logger.Information("Step 4: Adding 3 animation frames...")
	for i := 0; i < 3; i++ {
		if _, err := callTool(ctx, session, "add_frame", map[string]any{
			"sprite_path": spritePath,
			"duration_ms": 100,
		}); err != nil {
			return fmt.Errorf("add_frame failed: %w", err)
		}
	}
	logger.Information("  Frames added (4 total)")

	// Step 5: Draw circles on each frame (growing animation)
	logger.Information("")
	logger.Information("Step 5: Drawing animated circles...")
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

	// Step 6: Read pixels to verify drawing (read center 10x10 region from frame 2)
	logger.Information("")
	logger.Information("Step 6: Reading pixels from frame 2 to verify drawing...")
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

	// Step 7: Get sprite info
	logger.Information("")
	logger.Information("Step 7: Getting sprite metadata...")
	infoResp, err := callTool(ctx, session, "get_sprite_info", map[string]any{
		"sprite_path": spritePath,
	})
	if err != nil {
		return fmt.Errorf("get_sprite_info failed: %w", err)
	}
	logger.Information("  Info: {Info}", infoResp)

	// Step 8: Export as GIF
	logger.Information("")
	logger.Information("Step 8: Exporting as GIF...")
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

	// Step 9: Export frame 2 as PNG
	logger.Information("")
	logger.Information("Step 9: Exporting frame 2 as PNG...")
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

	// Step 10: Demonstrate dithering (create a new sprite with gradient)
	logger.Information("")
	logger.Information("Step 10: Creating sprite with dithered gradient...")
	ditherResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      64,
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

	// Apply dithering with Bayer 4x4 pattern
	logger.Information("")
	logger.Information("Step 11: Applying Bayer 4x4 dithering pattern...")
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
		return fmt.Errorf("draw_with_dither failed: %w", err)
	}
	logger.Information("  Dithering applied successfully")

	// Export dithered sprite
	logger.Information("")
	logger.Information("Step 12: Exporting dithered gradient...")
	ditherPngPath := filepath.Join(outputDir, "dithered-gradient.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  ditherSprite,
		"output_path":  ditherPngPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export_sprite dither failed: %w", err)
	}
	logger.Information("  Exported: {DitherPng}", ditherPngPath)

	// Step 13: Analyze palette harmonies
	logger.Information("")
	logger.Information("Step 13: Analyzing palette harmonies from our colors...")
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

	// Step 14: Create sprite with custom palette
	logger.Information("")
	logger.Information("Step 14: Creating sprite with limited palette...")
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

	// Step 15: Apply palette-constrained shading
	logger.Information("")
	logger.Information("Step 15: Drawing shape with palette-constrained shading...")

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

	// Step 16: Demonstrate palette-aware drawing (use_palette flag)
	logger.Information("")
	logger.Information("Step 16: Demonstrating palette-aware drawing...")
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
		"./bin/aseprite-mcp",
		"./bin/aseprite-mcp.exe",
		"../../bin/aseprite-mcp",
		"../../bin/aseprite-mcp.exe",
		"aseprite-mcp",
		"aseprite-mcp.exe",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// Check PATH
	if path, err := exec.LookPath("aseprite-mcp"); err == nil {
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
