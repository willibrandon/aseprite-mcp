package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/mtlog/core"
)

// SuggestAntialiasingInput defines the input parameters for antialiasing suggestions.
type SuggestAntialiasingInput struct {
	SpritePath  string  `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string  `json:"layer_name" jsonschema:"Name of the layer to analyze"`
	FrameNumber int     `json:"frame_number" jsonschema:"Frame number to analyze (1-based)"`
	Region      *Region `json:"region,omitempty" jsonschema:"Region to analyze (defaults to entire sprite)"`
	Threshold   int     `json:"threshold,omitempty" jsonschema:"Edge detection sensitivity 0-255 (default: 128)"`
	AutoApply   bool    `json:"auto_apply,omitempty" jsonschema:"If true applies smoothing automatically (default: false)"`
	UsePalette  bool    `json:"use_palette,omitempty" jsonschema:"If true snaps intermediate colors to palette (default: false)"`
}

// EdgeSuggestion represents a suggested antialiasing pixel placement.
type EdgeSuggestion struct {
	X              int    `json:"x"`
	Y              int    `json:"y"`
	CurrentColor   string `json:"current_color"`
	NeighborColor  string `json:"neighbor_color"`
	SuggestedColor string `json:"suggested_color"`
	Direction      string `json:"direction"` // diagonal_ne, diagonal_nw, diagonal_se, diagonal_sw
}

// AntialiasingResult contains antialiasing analysis and suggestions.
type AntialiasingResult struct {
	Suggestions []EdgeSuggestion `json:"suggestions"`
	Applied     bool             `json:"applied"`
	TotalEdges  int              `json:"total_edges"`
}

// suggestAntialiasing analyzes pixel art for jagged edges and suggests smoothing.
func suggestAntialiasing(ctx context.Context, client *aseprite.Client, gen *aseprite.LuaGenerator, input SuggestAntialiasingInput) (*AntialiasingResult, error) {
	// Set defaults
	if input.Threshold == 0 {
		input.Threshold = 128
	}

	// Get sprite info to determine region if not specified
	var region Region
	if input.Region != nil {
		region = *input.Region
	} else {
		info, err := getSpriteInfoHelper(ctx, client, gen, input.SpritePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get sprite info: %w", err)
		}
		region = Region{X: 0, Y: 0, Width: info.Width, Height: info.Height}
	}

	// Read pixels from the region
	pixels, err := getPixelsHelperForAntialiasing(ctx, client, gen, input.SpritePath, input.LayerName, input.FrameNumber,
		region.X, region.Y, region.Width, region.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to read pixels: %w", err)
	}

	// Build pixel grid for edge detection
	pixelGrid := buildPixelGrid(pixels)

	// Detect jagged edges and generate suggestions
	suggestions := detectJaggedEdges(pixelGrid, region, input.Threshold, input.UsePalette)

	result := &AntialiasingResult{
		Suggestions: suggestions,
		Applied:     false,
		TotalEdges:  len(suggestions),
	}

	// Apply suggestions if requested
	if input.AutoApply && len(suggestions) > 0 {
		if err := applyAntialiasingHelper(ctx, client, gen, input.SpritePath, input.LayerName, input.FrameNumber, suggestions, input.UsePalette); err != nil {
			return nil, fmt.Errorf("failed to apply antialiasing: %w", err)
		}
		result.Applied = true
	}

	return result, nil
}

// buildPixelGrid creates a 2D map of colors for edge detection.
func buildPixelGrid(pixels []PixelData) map[int]map[int]string {
	grid := make(map[int]map[int]string)

	for _, p := range pixels {
		if grid[p.Y] == nil {
			grid[p.Y] = make(map[int]string)
		}
		grid[p.Y][p.X] = p.Color
	}

	return grid
}

// detectJaggedEdges identifies diagonal edges that would benefit from antialiasing.
// Note: threshold parameter reserved for future edge detection sensitivity tuning.
// Note: usePalette is used during application (applyAntialiasingHelper), not detection.
func detectJaggedEdges(grid map[int]map[int]string, region Region, threshold int, usePalette bool) []EdgeSuggestion {
	var suggestions []EdgeSuggestion
	_ = threshold  // Reserved for future use
	_ = usePalette // Used during application, not detection

	// Scan for jagged diagonal patterns
	for y := region.Y; y < region.Y+region.Height-1; y++ {
		for x := region.X; x < region.X+region.Width-1; x++ {
			current := getPixel(grid, x, y)
			if current == "" || isTransparent(current) {
				continue
			}

			// Check for diagonal stair-step patterns (4 directions)
			// Northeast diagonal: ..##
			//                     .##.
			if suggestion := checkDiagonalNE(grid, x, y, current); suggestion != nil {
				suggestions = append(suggestions, *suggestion)
			}

			// Northwest diagonal: ##..
			//                     .##.
			if suggestion := checkDiagonalNW(grid, x, y, current); suggestion != nil {
				suggestions = append(suggestions, *suggestion)
			}

			// Southeast diagonal: .##.
			//                     ..##
			if suggestion := checkDiagonalSE(grid, x, y, current); suggestion != nil {
				suggestions = append(suggestions, *suggestion)
			}

			// Southwest diagonal: .##.
			//                     ##..
			if suggestion := checkDiagonalSW(grid, x, y, current); suggestion != nil {
				suggestions = append(suggestions, *suggestion)
			}
		}
	}

	return suggestions
}

// checkDiagonalNE checks for northeast diagonal jagged edge.
func checkDiagonalNE(grid map[int]map[int]string, x, y int, current string) *EdgeSuggestion {
	// Pattern: current at (x,y), same color at (x+1,y)
	//          empty at (x,y+1), same color at (x+1,y+1)
	// Suggests filling (x,y+1) with intermediate color

	right := getPixel(grid, x+1, y)
	below := getPixel(grid, x, y+1)
	belowRight := getPixel(grid, x+1, y+1)

	if right == current && (below == "" || isTransparent(below)) && belowRight == current {
		// Blend colors to create smooth transition
		suggested := blendColors(current, below)

		return &EdgeSuggestion{
			X:              x,
			Y:              y + 1,
			CurrentColor:   below,
			NeighborColor:  current,
			SuggestedColor: suggested,
			Direction:      "diagonal_ne",
		}
	}

	return nil
}

// checkDiagonalNW checks for northwest diagonal jagged edge.
func checkDiagonalNW(grid map[int]map[int]string, x, y int, current string) *EdgeSuggestion {
	// Pattern: same color at (x-1,y), current at (x,y)
	//          same color at (x-1,y+1), empty at (x,y+1)
	// Suggests filling (x,y+1) with intermediate color

	if x == 0 {
		return nil
	}

	left := getPixel(grid, x-1, y)
	below := getPixel(grid, x, y+1)
	belowLeft := getPixel(grid, x-1, y+1)

	if left == current && (below == "" || isTransparent(below)) && belowLeft == current {
		suggested := blendColors(current, below)

		return &EdgeSuggestion{
			X:              x,
			Y:              y + 1,
			CurrentColor:   below,
			NeighborColor:  current,
			SuggestedColor: suggested,
			Direction:      "diagonal_nw",
		}
	}

	return nil
}

// checkDiagonalSE checks for southeast diagonal jagged edge.
func checkDiagonalSE(grid map[int]map[int]string, x, y int, current string) *EdgeSuggestion {
	// Pattern: empty at (x,y), current at (x+1,y)
	//          current at (x,y+1), current at (x+1,y+1)
	// Suggests filling (x,y) with intermediate color

	right := getPixel(grid, x+1, y)
	below := getPixel(grid, x, y+1)
	belowRight := getPixel(grid, x+1, y+1)

	if (current == "" || isTransparent(current)) && right != "" && !isTransparent(right) &&
		below == right && belowRight == right {
		suggested := blendColors(right, current)

		return &EdgeSuggestion{
			X:              x,
			Y:              y,
			CurrentColor:   current,
			NeighborColor:  right,
			SuggestedColor: suggested,
			Direction:      "diagonal_se",
		}
	}

	return nil
}

// checkDiagonalSW checks for southwest diagonal jagged edge.
func checkDiagonalSW(grid map[int]map[int]string, x, y int, current string) *EdgeSuggestion {
	// Pattern: current at (x-1,y), empty at (x,y)
	//          current at (x-1,y+1), current at (x,y+1)
	// Suggests filling (x,y) with intermediate color

	if x == 0 {
		return nil
	}

	left := getPixel(grid, x-1, y)
	below := getPixel(grid, x, y+1)
	belowLeft := getPixel(grid, x-1, y+1)

	if left != "" && !isTransparent(left) && (current == "" || isTransparent(current)) &&
		below == left && belowLeft == left {
		suggested := blendColors(left, current)

		return &EdgeSuggestion{
			X:              x,
			Y:              y,
			CurrentColor:   current,
			NeighborColor:  left,
			SuggestedColor: suggested,
			Direction:      "diagonal_sw",
		}
	}

	return nil
}

// getPixel safely retrieves a pixel color from the grid.
func getPixel(grid map[int]map[int]string, x, y int) string {
	if row, ok := grid[y]; ok {
		if color, ok := row[x]; ok {
			return color
		}
	}
	return ""
}

// isTransparent checks if a color is transparent (alpha channel is 00).
func isTransparent(color string) bool {
	if len(color) >= 9 {
		return color[7:9] == "00"
	}
	return false
}

// blendColors creates an intermediate color between two colors (50% blend).
func blendColors(color1, color2 string) string {
	// Parse hex colors
	r1, g1, b1, a1 := parseHexColor(color1)
	r2, g2, b2, a2 := parseHexColor(color2)

	// Blend 50/50
	r := uint8((int(r1) + int(r2)) / 2)
	g := uint8((int(g1) + int(g2)) / 2)
	b := uint8((int(b1) + int(b2)) / 2)
	a := uint8((int(a1) + int(a2)) / 2)

	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
}

// parseHexColor parses a hex color string to RGBA components.
func parseHexColor(hex string) (r, g, b, a uint8) {
	if len(hex) < 7 {
		return 0, 0, 0, 255
	}

	// Remove # if present
	if hex[0] == '#' {
		hex = hex[1:]
	}

	// Parse RGB (errors ignored as format is validated by caller)
	_, _ = fmt.Sscanf(hex[0:2], "%02x", &r)
	_, _ = fmt.Sscanf(hex[2:4], "%02x", &g)
	_, _ = fmt.Sscanf(hex[4:6], "%02x", &b)

	// Parse alpha if present
	if len(hex) >= 8 {
		_, _ = fmt.Sscanf(hex[6:8], "%02x", &a)
	} else {
		a = 255
	}

	return
}

// getPixelsHelperForAntialiasing is a helper to read pixels from a sprite.
func getPixelsHelperForAntialiasing(ctx context.Context, client *aseprite.Client, gen *aseprite.LuaGenerator, spritePath, layerName string, frameNumber, x, y, width, height int) ([]PixelData, error) {
	script := gen.GetPixels(layerName, frameNumber, x, y, width, height)

	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		return nil, fmt.Errorf("ExecuteLua(GetPixels) failed: %w", err)
	}

	var pixels []PixelData
	if err := json.Unmarshal([]byte(output), &pixels); err != nil {
		return nil, fmt.Errorf("failed to parse pixel data: %w", err)
	}

	return pixels, nil
}

// getSpriteInfoHelper is a helper to get sprite metadata.
func getSpriteInfoHelper(ctx context.Context, client *aseprite.Client, gen *aseprite.LuaGenerator, spritePath string) (*aseprite.SpriteInfo, error) {
	script := gen.GetSpriteInfo()

	output, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		return nil, fmt.Errorf("ExecuteLua(GetSpriteInfo) failed: %w", err)
	}

	var info aseprite.SpriteInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		return nil, fmt.Errorf("failed to parse sprite info: %w", err)
	}

	return &info, nil
}

// applyAntialiasingHelper applies the suggested antialiasing pixels.
func applyAntialiasingHelper(ctx context.Context, client *aseprite.Client, gen *aseprite.LuaGenerator, spritePath, layerName string, frameNumber int, suggestions []EdgeSuggestion, usePalette bool) error {
	// Convert suggestions to pixel inputs
	pixels := make([]aseprite.Pixel, len(suggestions))
	for i, sug := range suggestions {
		r, g, b, a := parseHexColor(sug.SuggestedColor)
		pixels[i] = aseprite.Pixel{
			Point: aseprite.Point{X: sug.X, Y: sug.Y},
			Color: aseprite.Color{R: r, G: g, B: b, A: a},
		}
	}

	script := gen.DrawPixels(layerName, frameNumber, pixels, usePalette)
	_, err := client.ExecuteLua(ctx, script, spritePath)
	if err != nil {
		return fmt.Errorf("ExecuteLua(DrawPixels) failed: %w", err)
	}

	return nil
}

// findClosestPaletteColor finds the nearest color in a palette (Euclidean distance in RGB space).
func findClosestPaletteColor(targetR, targetG, targetB uint8, palette []string) string {
	if len(palette) == 0 {
		return fmt.Sprintf("#%02X%02X%02X", targetR, targetG, targetB)
	}

	minDist := math.MaxFloat64
	closest := palette[0]

	for _, palColor := range palette {
		pr, pg, pb, _ := parseHexColor(palColor)

		dr := float64(targetR) - float64(pr)
		dg := float64(targetG) - float64(pg)
		db := float64(targetB) - float64(pb)

		dist := math.Sqrt(dr*dr + dg*dg + db*db)

		if dist < minDist {
			minDist = dist
			closest = palColor
		}
	}

	return closest
}

// RegisterAntialiasingTools registers all antialiasing tools with the MCP server.
func RegisterAntialiasingTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register suggest_antialiasing tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "suggest_antialiasing",
			Description: "Analyze pixel art for jagged diagonal edges and suggest intermediate colors to smooth them (antialiasing). Detects stair-step patterns on diagonals and calculates blended colors to create smoother curves. Use auto_apply to automatically apply suggestions or use_palette to constrain intermediate colors to the sprite's palette. Returns suggestions with positions, colors, and directions for manual review or automatic application.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input SuggestAntialiasingInput) (*mcp.CallToolResult, *AntialiasingResult, error) {
			logger.Debug("suggest_antialiasing tool called",
				"sprite", input.SpritePath,
				"layer", input.LayerName,
				"frame", input.FrameNumber,
				"threshold", input.Threshold,
				"auto_apply", input.AutoApply,
				"use_palette", input.UsePalette)

			// Validate inputs
			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be >= 1, got %d", input.FrameNumber)
			}

			if input.LayerName == "" {
				return nil, nil, fmt.Errorf("layer_name is required")
			}

			if input.Threshold < 0 || input.Threshold > 255 {
				return nil, nil, fmt.Errorf("threshold must be 0-255, got %d", input.Threshold)
			}

			// Execute antialiasing analysis
			result, err := suggestAntialiasing(ctx, client, gen, input)
			if err != nil {
				return nil, nil, fmt.Errorf("antialiasing analysis failed: %w", err)
			}

			return &mcp.CallToolResult{
				IsError: false,
			}, result, nil
		},
	)
}
