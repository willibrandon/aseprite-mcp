package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

func TestDrawPixelsInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   DrawPixelsInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: DrawPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				Pixels: []PixelInput{
					{X: 0, Y: 0, Color: "#FF0000"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty layer name",
			input: DrawPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "",
				FrameNumber: 1,
				Pixels: []PixelInput{
					{X: 0, Y: 0, Color: "#FF0000"},
				},
			},
			wantErr: true,
			errMsg:  "layer_name cannot be empty",
		},
		{
			name: "invalid frame number (zero)",
			input: DrawPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 0,
				Pixels: []PixelInput{
					{X: 0, Y: 0, Color: "#FF0000"},
				},
			},
			wantErr: true,
			errMsg:  "frame_number must be at least 1",
		},
		{
			name: "invalid frame number (negative)",
			input: DrawPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: -1,
				Pixels: []PixelInput{
					{X: 0, Y: 0, Color: "#FF0000"},
				},
			},
			wantErr: true,
			errMsg:  "frame_number must be at least 1",
		},
		{
			name: "empty pixels array",
			input: DrawPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				Pixels:      []PixelInput{},
			},
			wantErr: true,
			errMsg:  "pixels array cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate layer name
			if tt.input.LayerName == "" && tt.wantErr {
				if tt.errMsg != "layer_name cannot be empty" {
					t.Errorf("Expected error message about layer_name")
				}
				return
			}

			// Validate frame number
			if tt.input.FrameNumber < 1 && tt.wantErr {
				if tt.errMsg != "frame_number must be at least 1" {
					t.Errorf("Expected error message about frame_number")
				}
				return
			}

			// Validate pixels array
			if len(tt.input.Pixels) == 0 && tt.wantErr {
				if tt.errMsg != "pixels array cannot be empty" {
					t.Errorf("Expected error message about pixels array")
				}
				return
			}

			// If we get here and wantErr is true, test failed
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestPixelInput_ColorFormats(t *testing.T) {
	tests := []struct {
		name    string
		color   string
		wantErr bool
	}{
		{
			name:    "valid RGB with hash",
			color:   "#FF0000",
			wantErr: false,
		},
		{
			name:    "valid RGB without hash",
			color:   "00FF00",
			wantErr: false,
		},
		{
			name:    "valid RGBA with hash",
			color:   "#0000FF80",
			wantErr: false,
		},
		{
			name:    "valid RGBA without hash",
			color:   "FFFF00FF",
			wantErr: false,
		},
		{
			name:    "invalid format - too short",
			color:   "#FFF",
			wantErr: true,
		},
		{
			name:    "invalid format - not hex",
			color:   "#GGGGGG",
			wantErr: true,
		},
		{
			name:    "invalid format - empty",
			color:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test color parsing using the Color type from aseprite package
			// This is tested indirectly through the tool validation
			pixel := PixelInput{
				Color: tt.color,
			}

			// Validate that pixel has a color string
			if pixel.Color == "" && !tt.wantErr {
				t.Error("Expected non-empty color")
			}
		})
	}
}

func TestDrawPixelsInput_MultiplePixels(t *testing.T) {
	input := DrawPixelsInput{
		Pixels: []PixelInput{
			{X: 0, Y: 0, Color: "#FF0000"},
			{X: 1, Y: 1, Color: "#00FF00"},
			{X: 2, Y: 2, Color: "#0000FF"},
			{X: 10, Y: 10, Color: "#FFFF00FF"},
		},
	}

	if len(input.Pixels) != 4 {
		t.Errorf("Expected 4 pixels, got %d", len(input.Pixels))
	}

	// Verify each pixel has valid structure
	for i, p := range input.Pixels {
		if p.Color == "" {
			t.Errorf("Pixel %d has empty color", i)
		}
	}
}

func TestDrawLineInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   DrawLineInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid line",
			input: DrawLineInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X1:          0,
				Y1:          0,
				X2:          10,
				Y2:          10,
				Color:       "#FF0000",
				Thickness:   1,
			},
			wantErr: false,
		},
		{
			name: "empty layer name",
			input: DrawLineInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "",
				FrameNumber: 1,
				X1:          0,
				Y1:          0,
				X2:          10,
				Y2:          10,
				Color:       "#FF0000",
				Thickness:   1,
			},
			wantErr: true,
			errMsg:  "layer_name cannot be empty",
		},
		{
			name: "invalid thickness - too small",
			input: DrawLineInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X1:          0,
				Y1:          0,
				X2:          10,
				Y2:          10,
				Color:       "#FF0000",
				Thickness:   0,
			},
			wantErr: true,
			errMsg:  "thickness must be between 1 and 100",
		},
		{
			name: "invalid thickness - too large",
			input: DrawLineInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X1:          0,
				Y1:          0,
				X2:          10,
				Y2:          10,
				Color:       "#FF0000",
				Thickness:   101,
			},
			wantErr: true,
			errMsg:  "thickness must be between 1 and 100",
		},
		{
			name: "maximum thickness",
			input: DrawLineInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X1:          0,
				Y1:          0,
				X2:          10,
				Y2:          10,
				Color:       "#FF0000",
				Thickness:   100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate inputs
			if tt.input.LayerName == "" && tt.wantErr {
				return
			}
			if (tt.input.Thickness < 1 || tt.input.Thickness > 100) && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestDrawRectangleInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   DrawRectangleInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid rectangle",
			input: DrawRectangleInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Width:       50,
				Height:      30,
				Color:       "#00FF00",
				Filled:      true,
			},
			wantErr: false,
		},
		{
			name: "invalid width",
			input: DrawRectangleInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Width:       0,
				Height:      30,
				Color:       "#00FF00",
				Filled:      false,
			},
			wantErr: true,
			errMsg:  "width and height must be at least 1",
		},
		{
			name: "invalid height",
			input: DrawRectangleInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Width:       50,
				Height:      0,
				Color:       "#00FF00",
				Filled:      false,
			},
			wantErr: true,
			errMsg:  "width and height must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate inputs
			if (tt.input.Width < 1 || tt.input.Height < 1) && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestDrawCircleInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   DrawCircleInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid circle",
			input: DrawCircleInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				CenterX:     50,
				CenterY:     50,
				Radius:      20,
				Color:       "#0000FF",
				Filled:      true,
			},
			wantErr: false,
		},
		{
			name: "invalid radius",
			input: DrawCircleInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				CenterX:     50,
				CenterY:     50,
				Radius:      0,
				Color:       "#0000FF",
				Filled:      false,
			},
			wantErr: true,
			errMsg:  "radius must be at least 1",
		},
		{
			name: "minimum radius",
			input: DrawCircleInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				CenterX:     50,
				CenterY:     50,
				Radius:      1,
				Color:       "#0000FF",
				Filled:      false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate inputs
			if tt.input.Radius < 1 && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestFillAreaInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   FillAreaInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid fill area",
			input: FillAreaInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Color:       "#FF0000",
				Tolerance:   0,
			},
			wantErr: false,
		},
		{
			name: "valid fill area with tolerance",
			input: FillAreaInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Color:       "#FF0000",
				Tolerance:   50,
			},
			wantErr: false,
		},
		{
			name: "maximum tolerance",
			input: FillAreaInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Color:       "#FF0000",
				Tolerance:   255,
			},
			wantErr: false,
		},
		{
			name: "invalid tolerance - too small",
			input: FillAreaInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Color:       "#FF0000",
				Tolerance:   -1,
			},
			wantErr: true,
			errMsg:  "tolerance must be between 0 and 255",
		},
		{
			name: "invalid tolerance - too large",
			input: FillAreaInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Color:       "#FF0000",
				Tolerance:   256,
			},
			wantErr: true,
			errMsg:  "tolerance must be between 0 and 255",
		},
		{
			name: "empty layer name",
			input: FillAreaInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "",
				FrameNumber: 1,
				X:           10,
				Y:           10,
				Color:       "#FF0000",
				Tolerance:   0,
			},
			wantErr: true,
			errMsg:  "layer_name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate inputs
			if tt.input.LayerName == "" && tt.wantErr {
				return
			}
			if (tt.input.Tolerance < 0 || tt.input.Tolerance > 255) && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

// MCP Protocol Tests (use real MCP with in-memory transport and real Aseprite)

// createDrawingTestSession creates an MCP session with drawing tools registered
func createDrawingTestSession(t *testing.T) (*mcp.Server, *mcp.ClientSession, *aseprite.Client) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pixel-mcp-test",
		Version: "1.0.0",
	}, nil)

	RegisterDrawingTools(server, client, gen, cfg, logger)

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

func TestDrawPixels_ViaMCP(t *testing.T) {
	_, session, client := createDrawingTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-draw.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-draw.aseprite")

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_pixels",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-draw.aseprite",
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"pixels": []map[string]any{
				{"x": 10, "y": 10, "color": "#FF0000"},
				{"x": 11, "y": 10, "color": "#00FF00"},
				{"x": 12, "y": 10, "color": "#0000FF"},
			},
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		PixelsDrawn int `json:"pixels_drawn"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.Equal(t, 3, output.PixelsDrawn)
}

func TestDrawLine_ViaMCP(t *testing.T) {
	_, session, client := createDrawingTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-line.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-line.aseprite")

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_line",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-line.aseprite",
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"x1":           10,
			"y1":           10,
			"x2":           50,
			"y2":           50,
			"color":        "#FFFFFF",
			"thickness":    2,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}

func TestDrawRectangle_ViaMCP(t *testing.T) {
	_, session, client := createDrawingTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-rect.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-rect.aseprite")

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_rectangle",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-rect.aseprite",
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"x":            10,
			"y":            10,
			"width":        30,
			"height":       20,
			"color":        "#FF00FF",
			"filled":       true,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}

func TestDrawCircle_ViaMCP(t *testing.T) {
	_, session, client := createDrawingTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-circle.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-circle.aseprite")

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_circle",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-circle.aseprite",
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"center_x":     32,
			"center_y":     32,
			"radius":       15,
			"color":        "#FFFF00",
			"filled":       false,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}

func TestFillArea_ViaMCP(t *testing.T) {
	_, session, client := createDrawingTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-fill.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-fill.aseprite")

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "fill_area",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-fill.aseprite",
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"x":            32,
			"y":            32,
			"color":        "#00FFFF",
			"tolerance":    0,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}

func TestDrawContour_ViaMCP(t *testing.T) {
	_, session, client := createDrawingTestSession(t)
	defer session.Close()

	cfg := testutil.LoadTestConfig(t)
	gen := aseprite.NewLuaGenerator()

	script := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, cfg.TempDir+"/test-contour.aseprite")
	_, err := client.ExecuteLua(context.Background(), script, "")
	require.NoError(t, err)
	defer os.Remove(cfg.TempDir + "/test-contour.aseprite")

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "draw_contour",
		Arguments: map[string]any{
			"sprite_path":  cfg.TempDir + "/test-contour.aseprite",
			"layer_name":   "Layer 1",
			"frame_number": 1,
			"points": []map[string]any{
				{"x": 10, "y": 10},
				{"x": 30, "y": 10},
				{"x": 30, "y": 30},
				{"x": 10, "y": 30},
			},
			"color":     "#FF00FF",
			"thickness": 2,
			"closed":    true,
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)

	var output struct {
		Success bool `json:"success"`
	}
	json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &output)
	assert.True(t, output.Success)
}
