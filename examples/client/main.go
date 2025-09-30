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