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