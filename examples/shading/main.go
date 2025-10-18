// Package main demonstrates the apply_auto_shading tool with different styles and light directions.
// This example creates sprites with shapes and applies automatic shading based on geometry analysis.
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
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	logger.Information("Automatic Shading Example")
	logger.Information("=========================")
	logger.Information("")

	// Start MCP server
	serverPath := os.Getenv("ASEPRITE_MCP_PATH")
	if serverPath == "" {
		serverPath = findServerBinary()
		if serverPath == "" {
			return fmt.Errorf("ASEPRITE_MCP_PATH not set and could not find pixel-mcp binary")
		}
	}

	logger.Information("Starting server: {ServerPath}", serverPath)
	cmd := exec.Command(serverPath)

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "shading-example",
		Version: "1.0.0",
	}, nil)

	logger.Information("Connecting to server...")
	transport := &mcp.CommandTransport{Command: cmd}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer session.Close()

	logger.Information("Connected!")
	logger.Information("")

	// Create output directory
	outputDir := "examples/sprites"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Example 1: Cell shading style (hard-edged bands)
	logger.Information("Example 1: Cell shading style (2-3 color bands)")
	logger.Information("------------------------------------------------")
	if err := demonstrateCellShading(ctx, session, logger, outputDir); err != nil {
		return err
	}

	// Example 2: Smooth shading style (gradient with dithering)
	logger.Information("")
	logger.Information("Example 2: Smooth shading style (gradient)")
	logger.Information("-------------------------------------------")
	if err := demonstrateSmoothShading(ctx, session, logger, outputDir); err != nil {
		return err
	}

	// Example 3: Soft shading style (subtle gradient)
	logger.Information("")
	logger.Information("Example 3: Soft shading style (subtle)")
	logger.Information("---------------------------------------")
	if err := demonstrateSoftShading(ctx, session, logger, outputDir); err != nil {
		return err
	}

	// Example 4: Different light directions
	logger.Information("")
	logger.Information("Example 4: All 8 light directions")
	logger.Information("----------------------------------")
	if err := demonstrateLightDirections(ctx, session, logger, outputDir); err != nil {
		return err
	}

	// Example 5: Hue shifting comparison
	logger.Information("")
	logger.Information("Example 5: With vs without hue shifting")
	logger.Information("----------------------------------------")
	if err := demonstrateHueShifting(ctx, session, logger, outputDir); err != nil {
		return err
	}

	logger.Information("")
	logger.Information("Example completed successfully!")
	logger.Information("")
	logger.Information("Output files created:")
	logger.Information("  - shading-cell-*.png (3 examples with cell shading)")
	logger.Information("  - shading-smooth-*.png (3 examples with smooth shading)")
	logger.Information("  - shading-soft-*.png (1 example with soft shading)")
	logger.Information("  - shading-directions.png (comparison of 8 light directions)")
	logger.Information("  - shading-hueshift-*.png (comparison with/without hue shift)")

	return nil
}

func demonstrateCellShading(ctx context.Context, session *mcp.ClientSession, logger core.Logger, outputDir string) error {
	// Create sprites with different shapes
	shapes := []struct {
		name  string
		color string
		draw  func(context.Context, *mcp.ClientSession, string) error
	}{
		{
			name:  "sphere",
			color: "#FF6B6B",
			draw: func(ctx context.Context, session *mcp.ClientSession, spritePath string) error {
				_, err := callTool(ctx, session, "draw_circle", map[string]any{
					"sprite_path":  spritePath,
					"layer_name":   "Layer 1",
					"frame_number": 1,
					"center_x":     32,
					"center_y":     32,
					"radius":       28,
					"color":        "#FF6B6B",
					"filled":       true,
				})
				return err
			},
		},
		{
			name:  "cube",
			color: "#4ECDC4",
			draw: func(ctx context.Context, session *mcp.ClientSession, spritePath string) error {
				_, err := callTool(ctx, session, "draw_rectangle", map[string]any{
					"sprite_path":  spritePath,
					"layer_name":   "Layer 1",
					"frame_number": 1,
					"x":            8,
					"y":            8,
					"width":        48,
					"height":       48,
					"color":        "#4ECDC4",
					"filled":       true,
				})
				return err
			},
		},
		{
			name:  "pill",
			color: "#FFE66D",
			draw: func(ctx context.Context, session *mcp.ClientSession, spritePath string) error {
				// Draw a capsule/pill shape (rectangle + circles on ends)
				if _, err := callTool(ctx, session, "draw_rectangle", map[string]any{
					"sprite_path":  spritePath,
					"layer_name":   "Layer 1",
					"frame_number": 1,
					"x":            20,
					"y":            16,
					"width":        24,
					"height":       32,
					"color":        "#FFE66D",
					"filled":       true,
				}); err != nil {
					return err
				}
				if _, err := callTool(ctx, session, "draw_circle", map[string]any{
					"sprite_path":  spritePath,
					"layer_name":   "Layer 1",
					"frame_number": 1,
					"center_x":     32,
					"center_y":     16,
					"radius":       12,
					"color":        "#FFE66D",
					"filled":       true,
				}); err != nil {
					return err
				}
				_, err := callTool(ctx, session, "draw_circle", map[string]any{
					"sprite_path":  spritePath,
					"layer_name":   "Layer 1",
					"frame_number": 1,
					"center_x":     32,
					"center_y":     48,
					"radius":       12,
					"color":        "#FFE66D",
					"filled":       true,
				})
				return err
			},
		},
	}

	for _, shape := range shapes {
		// Create canvas
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

		// Draw shape
		logger.Information("  Drawing {Shape}...", shape.name)
		if err := shape.draw(ctx, session, spritePath); err != nil {
			return fmt.Errorf("draw shape failed: %w", err)
		}

		// Apply cell shading
		shadingResp, err := callTool(ctx, session, "apply_auto_shading", map[string]any{
			"sprite_path":     spritePath,
			"layer_name":      "Layer 1",
			"frame_number":    1,
			"light_direction": "top_left",
			"intensity":       0.6,
			"style":           "cell",
			"hue_shift":       true,
		})
		if err != nil {
			return fmt.Errorf("apply_auto_shading failed: %w", err)
		}

		var shadingResult struct {
			Success       bool     `json:"success"`
			ColorsAdded   int      `json:"colors_added"`
			Palette       []string `json:"palette"`
			RegionsShaded int      `json:"regions_shaded"`
		}
		if err := json.Unmarshal([]byte(shadingResp), &shadingResult); err != nil {
			return fmt.Errorf("failed to parse shading result: %w", err)
		}

		logger.Information("    Shaded {Shape}: {Colors} colors added, {Regions} region(s)",
			shape.name, shadingResult.ColorsAdded, shadingResult.RegionsShaded)

		// Export
		outputPath := filepath.Join(outputDir, fmt.Sprintf("shading-cell-%s.png", shape.name))
		if _, err := callTool(ctx, session, "export_sprite", map[string]any{
			"sprite_path":  spritePath,
			"output_path":  outputPath,
			"format":       "png",
			"frame_number": 0,
		}); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		logger.Information("    Exported: {Path}", outputPath)

		os.Remove(spritePath)
	}

	return nil
}

func demonstrateSmoothShading(ctx context.Context, session *mcp.ClientSession, logger core.Logger, outputDir string) error {
	// Create sprites with different intensities
	intensities := []struct {
		value float64
		label string
	}{
		{0.3, "subtle"},
		{0.6, "medium"},
		{0.9, "strong"},
	}

	for _, intensity := range intensities {
		// Create canvas
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

		// Draw sphere
		logger.Information("  Drawing sphere with {Intensity} intensity...", intensity.label)
		if _, err := callTool(ctx, session, "draw_circle", map[string]any{
			"sprite_path":  spritePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"center_x":     32,
			"center_y":     32,
			"radius":       28,
			"color":        "#A78BFA",
			"filled":       true,
		}); err != nil {
			return fmt.Errorf("draw_circle failed: %w", err)
		}

		// Apply smooth shading
		shadingResp, err := callTool(ctx, session, "apply_auto_shading", map[string]any{
			"sprite_path":     spritePath,
			"layer_name":      "Layer 1",
			"frame_number":    1,
			"light_direction": "top_left",
			"intensity":       intensity.value,
			"style":           "smooth",
			"hue_shift":       true,
		})
		if err != nil {
			return fmt.Errorf("apply_auto_shading failed: %w", err)
		}

		var shadingResult struct {
			Success       bool `json:"success"`
			ColorsAdded   int  `json:"colors_added"`
			RegionsShaded int  `json:"regions_shaded"`
		}
		if err := json.Unmarshal([]byte(shadingResp), &shadingResult); err != nil {
			return fmt.Errorf("failed to parse shading result: %w", err)
		}

		logger.Information("    Shaded: {Colors} colors added", shadingResult.ColorsAdded)

		// Export
		outputPath := filepath.Join(outputDir, fmt.Sprintf("shading-smooth-%s.png", intensity.label))
		if _, err := callTool(ctx, session, "export_sprite", map[string]any{
			"sprite_path":  spritePath,
			"output_path":  outputPath,
			"format":       "png",
			"frame_number": 0,
		}); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		logger.Information("    Exported: {Path}", outputPath)

		os.Remove(spritePath)
	}

	return nil
}

func demonstrateSoftShading(ctx context.Context, session *mcp.ClientSession, logger core.Logger, outputDir string) error {
	// Create canvas
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

	// Draw sphere
	logger.Information("  Drawing sphere with soft shading...")
	if _, err := callTool(ctx, session, "draw_circle", map[string]any{
		"sprite_path":  spritePath,
		"layer_name":   "Layer 1",
		"frame_number": 1,
		"center_x":     32,
		"center_y":     32,
		"radius":       28,
		"color":        "#FB923C",
		"filled":       true,
	}); err != nil {
		return fmt.Errorf("draw_circle failed: %w", err)
	}

	// Apply soft shading (low intensity)
	shadingResp, err := callTool(ctx, session, "apply_auto_shading", map[string]any{
		"sprite_path":     spritePath,
		"layer_name":      "Layer 1",
		"frame_number":    1,
		"light_direction": "top_right",
		"intensity":       0.3,
		"style":           "soft",
		"hue_shift":       true,
	})
	if err != nil {
		return fmt.Errorf("apply_auto_shading failed: %w", err)
	}

	var shadingResult struct {
		Success       bool `json:"success"`
		ColorsAdded   int  `json:"colors_added"`
		RegionsShaded int  `json:"regions_shaded"`
	}
	if err := json.Unmarshal([]byte(shadingResp), &shadingResult); err != nil {
		return fmt.Errorf("failed to parse shading result: %w", err)
	}

	logger.Information("    Shaded: {Colors} colors added", shadingResult.ColorsAdded)

	// Export
	outputPath := filepath.Join(outputDir, "shading-soft-sphere.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  spritePath,
		"output_path":  outputPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}
	logger.Information("    Exported: {Path}", outputPath)

	os.Remove(spritePath)

	return nil
}

func demonstrateLightDirections(ctx context.Context, session *mcp.ClientSession, logger core.Logger, outputDir string) error {
	// Create wide canvas for comparison
	createResp, err := callTool(ctx, session, "create_canvas", map[string]any{
		"width":      64 * 8,
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
	comparisonPath := createResult.FilePath

	directions := []string{
		"top_left", "top", "top_right",
		"left", "right",
		"bottom_left", "bottom", "bottom_right",
	}

	logger.Information("  Creating comparison with 8 light directions...")

	for i, direction := range directions {
		// Create individual sprite for each direction
		dirCreateResp, err := callTool(ctx, session, "create_canvas", map[string]any{
			"width":      64,
			"height":     64,
			"color_mode": "rgb",
		})
		if err != nil {
			return fmt.Errorf("create_canvas for direction failed: %w", err)
		}

		var dirCreateResult struct {
			FilePath string `json:"file_path"`
		}
		if err := json.Unmarshal([]byte(dirCreateResp), &dirCreateResult); err != nil {
			return fmt.Errorf("failed to parse create result: %w", err)
		}
		dirSpritePath := dirCreateResult.FilePath

		// Draw sphere
		if _, err := callTool(ctx, session, "draw_circle", map[string]any{
			"sprite_path":  dirSpritePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"center_x":     32,
			"center_y":     32,
			"radius":       28,
			"color":        "#34D399",
			"filled":       true,
		}); err != nil {
			return fmt.Errorf("draw_circle failed: %w", err)
		}

		// Apply shading from this direction
		if _, err := callTool(ctx, session, "apply_auto_shading", map[string]any{
			"sprite_path":     dirSpritePath,
			"layer_name":      "Layer 1",
			"frame_number":    1,
			"light_direction": direction,
			"intensity":       0.7,
			"style":           "cell",
			"hue_shift":       true,
		}); err != nil {
			return fmt.Errorf("apply_auto_shading failed: %w", err)
		}

		// Export to temp location
		tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("temp-dir-%s.png", direction))
		if _, err := callTool(ctx, session, "export_sprite", map[string]any{
			"sprite_path":  dirSpritePath,
			"output_path":  tempPath,
			"format":       "png",
			"frame_number": 0,
		}); err != nil {
			return fmt.Errorf("export temp failed: %w", err)
		}

		// Import into comparison sprite
		if _, err := callTool(ctx, session, "import_image", map[string]any{
			"sprite_path":  comparisonPath,
			"image_path":   tempPath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"position": map[string]any{
				"x": i * 64,
				"y": 0,
			},
		}); err != nil {
			return fmt.Errorf("import_image failed: %w", err)
		}

		// Clean up
		os.Remove(tempPath)
		os.Remove(dirSpritePath)

		logger.Information("    Added {Direction}", direction)
	}

	// Export comparison
	outputPath := filepath.Join(outputDir, "shading-directions.png")
	if _, err := callTool(ctx, session, "export_sprite", map[string]any{
		"sprite_path":  comparisonPath,
		"output_path":  outputPath,
		"format":       "png",
		"frame_number": 0,
	}); err != nil {
		return fmt.Errorf("export comparison failed: %w", err)
	}
	logger.Information("  Exported: {Path}", outputPath)
	logger.Information("  Layout: top_left | top | top_right | left | right | bottom_left | bottom | bottom_right")

	os.Remove(comparisonPath)

	return nil
}

func demonstrateHueShifting(ctx context.Context, session *mcp.ClientSession, logger core.Logger, outputDir string) error {
	logger.Information("  Comparing with and without hue shifting...")

	for _, withHueShift := range []bool{true, false} {
		// Create canvas
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

		// Draw sphere
		if _, err := callTool(ctx, session, "draw_circle", map[string]any{
			"sprite_path":  spritePath,
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"center_x":     32,
			"center_y":     32,
			"radius":       28,
			"color":        "#F59E0B",
			"filled":       true,
		}); err != nil {
			return fmt.Errorf("draw_circle failed: %w", err)
		}

		// Apply shading
		shadingResp, err := callTool(ctx, session, "apply_auto_shading", map[string]any{
			"sprite_path":     spritePath,
			"layer_name":      "Layer 1",
			"frame_number":    1,
			"light_direction": "top_left",
			"intensity":       0.7,
			"style":           "smooth",
			"hue_shift":       withHueShift,
		})
		if err != nil {
			return fmt.Errorf("apply_auto_shading failed: %w", err)
		}

		var shadingResult struct {
			Success     bool     `json:"success"`
			ColorsAdded int      `json:"colors_added"`
			Palette     []string `json:"palette"`
		}
		if err := json.Unmarshal([]byte(shadingResp), &shadingResult); err != nil {
			return fmt.Errorf("failed to parse shading result: %w", err)
		}

		label := "without"
		if withHueShift {
			label = "with"
		}
		logger.Information("    {Label} hue shift: {Colors} colors", label, shadingResult.ColorsAdded)
		logger.Information("      Palette: {Colors}", shadingResult.Palette)

		// Export
		suffix := "no-hueshift"
		if withHueShift {
			suffix = "with-hueshift"
		}
		outputPath := filepath.Join(outputDir, fmt.Sprintf("shading-hueshift-%s.png", suffix))
		if _, err := callTool(ctx, session, "export_sprite", map[string]any{
			"sprite_path":  spritePath,
			"output_path":  outputPath,
			"format":       "png",
			"frame_number": 0,
		}); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		logger.Information("      Exported: {Path}", outputPath)

		os.Remove(spritePath)
	}

	logger.Information("  Note: Hue shift makes shadows cooler and highlights warmer for more natural appearance")

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

	if len(resp.Content) > 0 {
		if textContent, ok := resp.Content[0].(*mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in response")
}

func findServerBinary() string {
	candidates := []string{
		"../../bin/pixel-mcp",
		"../../bin/pixel-mcp.exe",
		"./bin/pixel-mcp",
		"./bin/pixel-mcp.exe",
		"pixel-mcp",
		"pixel-mcp.exe",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	if path, err := exec.LookPath("pixel-mcp"); err == nil {
		return path
	}

	return ""
}

func createLogger() core.Logger {
	sink := sinks.NewConsoleSink()
	logger := mtlog.New(
		mtlog.WithSink(sink),
		mtlog.WithMinimumLevel(core.InformationLevel),
	)
	return logger
}
