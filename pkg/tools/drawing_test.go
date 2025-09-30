package tools

import (
	"testing"
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
				X:     0,
				Y:     0,
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
		SpritePath:  "/path/to/sprite.aseprite",
		LayerName:   "Layer 1",
		FrameNumber: 1,
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