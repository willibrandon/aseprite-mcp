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
	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
)

// createAnalysisTestSession creates an MCP session with analysis tools registered
func createAnalysisTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "aseprite-mcp-test",
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

	// Verify edge map (note: edge detection is done on source image for full detail)
	assert.NotNil(t, output.EdgeMap, "Should have edge map")
	assert.NotNil(t, output.EdgeMap.Grid, "Should have edge grid")
	assert.Equal(t, 100, len(output.EdgeMap.Grid), "Edge grid height should match source")
	if len(output.EdgeMap.Grid) > 0 {
		assert.Equal(t, 100, len(output.EdgeMap.Grid[0]), "Edge grid width should match source")
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
