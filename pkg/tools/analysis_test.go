package tools

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// createAnalysisTestSession creates an MCP session with analysis tools registered
func createAnalysisTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pixel-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterAnalysisTools(server, client, gen, cfg, logger)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	_, err := server.Connect(context.Background(), serverTransport, nil)
	require.NoError(t, err)

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	session, err := mcpClient.Connect(context.Background(), clientTransport, nil)
	require.NoError(t, err)

	return server, session, client
}

// createTestReferenceImage creates a simple test image with known colors
func createTestReferenceImage(t *testing.T) string {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	imgPath := filepath.Join(cfg.TempDir, "test_reference.png")

	// Create a 100x100 image with a gradient
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Draw gradient from red to blue
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			r := uint8(255 - (x * 255 / 100))
			b := uint8(x * 255 / 100)
			img.Set(x, y, color.RGBA{R: r, G: 0, B: b, A: 255})
		}
	}

	// Save the image
	f, err := os.Create(imgPath)
	require.NoError(t, err)
	defer f.Close()

	err = png.Encode(f, img)
	require.NoError(t, err)

	return imgPath
}

func TestAnalyzeReference_ViaMCP(t *testing.T) {
	_, session, _ := createAnalysisTestSession(t)
	defer session.Close()

	// Create test reference image
	refPath := createTestReferenceImage(t)
	defer os.Remove(refPath)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "analyze_reference",
		Arguments: map[string]any{
			"reference_path":    refPath,
			"target_width":      32,
			"target_height":     32,
			"palette_size":      8,
			"brightness_levels": 5,
			"edge_threshold":    30,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output AnalyzeReferenceOutput
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	// Verify palette was extracted
	assert.NotNil(t, output.Palette, "Should have palette")
	assert.Equal(t, 8, len(output.Palette), "Should extract 8 colors")

	// Verify brightness map
	assert.NotNil(t, output.BrightnessMap, "Should have brightness map")
	assert.NotNil(t, output.BrightnessMap.Grid, "Should have brightness grid")
	assert.Equal(t, 32, len(output.BrightnessMap.Grid), "Brightness grid height should match target")
	if len(output.BrightnessMap.Grid) > 0 {
		assert.Equal(t, 32, len(output.BrightnessMap.Grid[0]), "Brightness grid width should match target")
	}
	assert.NotNil(t, output.BrightnessMap.Legend, "Should have brightness legend")

	// Verify edge map (downsampled to target resolution to reduce response size)
	assert.NotNil(t, output.EdgeMap, "Should have edge map")
	assert.NotNil(t, output.EdgeMap.Grid, "Should have edge grid")
	assert.Equal(t, 32, len(output.EdgeMap.Grid), "Edge grid height should match target")
	if len(output.EdgeMap.Grid) > 0 {
		assert.Equal(t, 32, len(output.EdgeMap.Grid[0]), "Edge grid width should match target")
	}

	// Verify composition analysis
	assert.NotNil(t, output.Composition, "Should have composition analysis")
	assert.NotNil(t, output.Composition.RuleOfThirds, "Should have rule of thirds")
	assert.Len(t, output.Composition.RuleOfThirds.HorizontalLines, 2, "Should have 2 horizontal third lines")
	assert.Len(t, output.Composition.RuleOfThirds.VerticalLines, 2, "Should have 2 vertical third lines")

	// Verify metadata
	assert.NotNil(t, output.Metadata, "Should have metadata")
	assert.Equal(t, 100, output.Metadata.SourceDimensions.Width)
	assert.Equal(t, 100, output.Metadata.SourceDimensions.Height)
	assert.Equal(t, 32, output.Metadata.TargetDimensions.Width)
	assert.Equal(t, 32, output.Metadata.TargetDimensions.Height)

	// Verify dithering zones were suggested
	assert.NotNil(t, output.DitheringZones, "Should have dithering zone suggestions")
}

func TestAnalyzeReferenceDefaults_ViaMCP(t *testing.T) {
	_, session, _ := createAnalysisTestSession(t)
	defer session.Close()

	// Create test reference image
	refPath := createTestReferenceImage(t)
	defer os.Remove(refPath)

	// Test with minimal params (should use defaults)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "analyze_reference",
		Arguments: map[string]any{
			"reference_path": refPath,
			"target_width":   16,
			"target_height":  16,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output AnalyzeReferenceOutput
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)

	// Verify defaults were used
	assert.Equal(t, 16, len(output.Palette), "Should use default palette size of 16")
	assert.NotNil(t, output.BrightnessMap, "Should have brightness map with defaults")
	assert.NotNil(t, output.EdgeMap, "Should have edge map with defaults")
}

// Unit tests for findColorsForBrightnessRange helper function
func TestFindColorsForBrightnessRange(t *testing.T) {
	tests := []struct {
		name      string
		palette   []aseprite.PaletteColor
		minLevel  int
		maxLevel  int
		numLevels int
		wantCount int
		checkFunc func(t *testing.T, colors []string)
	}{
		{
			name: "colors found in range",
			palette: []aseprite.PaletteColor{
				{Color: "#000000", Lightness: 0.0},
				{Color: "#808080", Lightness: 50.0},
				{Color: "#FFFFFF", Lightness: 100.0},
			},
			minLevel:  2,
			maxLevel:  2,
			numLevels: 5,
			wantCount: 1, // Should find middle color
			checkFunc: func(t *testing.T, colors []string) {
				assert.Contains(t, colors, "#808080")
			},
		},
		{
			name: "no colors in range - empty palette",
			palette: []aseprite.PaletteColor{
				{Color: "#000000", Lightness: 0.0},
				{Color: "#FFFFFF", Lightness: 100.0},
			},
			minLevel:  2,
			maxLevel:  2,
			numLevels: 5,
			wantCount: 2, // Should return darkest and lightest
			checkFunc: func(t *testing.T, colors []string) {
				assert.Contains(t, colors, "#000000", "Should include darkest color")
				assert.Contains(t, colors, "#FFFFFF", "Should include lightest color")
			},
		},
		{
			name:      "empty palette",
			palette:   []aseprite.PaletteColor{},
			minLevel:  0,
			maxLevel:  4,
			numLevels: 5,
			wantCount: 0,
			checkFunc: func(t *testing.T, colors []string) {
				assert.Empty(t, colors)
			},
		},
		{
			name: "single color palette - no match",
			palette: []aseprite.PaletteColor{
				{Color: "#FF0000", Lightness: 25.0},
			},
			minLevel:  4,
			maxLevel:  4,
			numLevels: 5,
			wantCount: 1, // Should return the only color
			checkFunc: func(t *testing.T, colors []string) {
				assert.Contains(t, colors, "#FF0000")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colors := findColorsForBrightnessRange(tt.palette, tt.minLevel, tt.maxLevel, tt.numLevels)
			assert.Len(t, colors, tt.wantCount)
			if tt.checkFunc != nil {
				tt.checkFunc(t, colors)
			}
		})
	}
}

// Unit tests for absDiff helper function
func TestAbsDiff(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		want float64
	}{
		{
			name: "positive difference",
			a:    100.0,
			b:    50.0,
			want: 50.0,
		},
		{
			name: "negative difference",
			a:    50.0,
			b:    100.0,
			want: 50.0,
		},
		{
			name: "wrap-around case (350 and 10 degrees)",
			a:    350.0,
			b:    10.0,
			want: 20.0, // 360 - 340 = 20
		},
		{
			name: "wrap-around case (10 and 350 degrees)",
			a:    10.0,
			b:    350.0,
			want: 20.0,
		},
		{
			name: "exact 180 degrees",
			a:    0.0,
			b:    180.0,
			want: 180.0,
		},
		{
			name: "greater than 180",
			a:    0.0,
			b:    270.0,
			want: 90.0, // 360 - 270 = 90
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := absDiff(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
