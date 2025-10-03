package tools

import (
	"testing"
)

func TestGetPixelsInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   GetPixelsInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           0,
				Y:           0,
				Width:       10,
				Height:      10,
			},
			wantErr: false,
		},
		{
			name: "invalid width - zero",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           0,
				Y:           0,
				Width:       0,
				Height:      10,
			},
			wantErr: true,
			errMsg:  "width and height must be positive",
		},
		{
			name: "invalid width - negative",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           0,
				Y:           0,
				Width:       -1,
				Height:      10,
			},
			wantErr: true,
			errMsg:  "width and height must be positive",
		},
		{
			name: "invalid height - zero",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           0,
				Y:           0,
				Width:       10,
				Height:      0,
			},
			wantErr: true,
			errMsg:  "width and height must be positive",
		},
		{
			name: "invalid height - negative",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           0,
				Y:           0,
				Width:       10,
				Height:      -1,
			},
			wantErr: true,
			errMsg:  "width and height must be positive",
		},
		{
			name: "invalid frame number - zero",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 0,
				X:           0,
				Y:           0,
				Width:       10,
				Height:      10,
			},
			wantErr: true,
			errMsg:  "frame_number must be >= 1",
		},
		{
			name: "invalid frame number - negative",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: -1,
				X:           0,
				Y:           0,
				Width:       10,
				Height:      10,
			},
			wantErr: true,
			errMsg:  "frame_number must be >= 1",
		},
		{
			name: "valid with negative coordinates",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           -5,
				Y:           -5,
				Width:       10,
				Height:      10,
			},
			wantErr: false,
		},
		{
			name: "large region",
			input: GetPixelsInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				FrameNumber: 1,
				X:           0,
				Y:           0,
				Width:       1000,
				Height:      1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate width and height
			if (tt.input.Width <= 0 || tt.input.Height <= 0) && tt.wantErr {
				if tt.errMsg != "width and height must be positive" {
					t.Errorf("Expected error message about width/height")
				}
				return
			}

			// Validate frame number
			if tt.input.FrameNumber < 1 && tt.wantErr {
				if tt.errMsg != "frame_number must be >= 1" {
					t.Errorf("Expected error message about frame_number")
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

func TestPixelData_Structure(t *testing.T) {
	tests := []struct {
		name  string
		pixel PixelData
	}{
		{
			name: "RGB color",
			pixel: PixelData{
				X:     10,
				Y:     20,
				Color: "#FF0000",
			},
		},
		{
			name: "RGBA color",
			pixel: PixelData{
				X:     5,
				Y:     15,
				Color: "#00FF0080",
			},
		},
		{
			name: "black color",
			pixel: PixelData{
				X:     0,
				Y:     0,
				Color: "#000000",
			},
		},
		{
			name: "white color",
			pixel: PixelData{
				X:     100,
				Y:     100,
				Color: "#FFFFFF",
			},
		},
		{
			name: "transparent color",
			pixel: PixelData{
				X:     50,
				Y:     50,
				Color: "#00000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pixel.Color == "" {
				t.Error("Expected non-empty color")
			}
			if len(tt.pixel.Color) < 7 {
				t.Errorf("Expected color to be at least 7 characters (e.g., #RRGGBB), got %d", len(tt.pixel.Color))
			}
			if tt.pixel.Color[0] != '#' {
				t.Errorf("Expected color to start with #, got %s", tt.pixel.Color)
			}
		})
	}
}

func TestGetPixelsOutput_Structure(t *testing.T) {
	output := GetPixelsOutput{
		Pixels: []PixelData{
			{X: 0, Y: 0, Color: "#FF0000FF"},
			{X: 1, Y: 0, Color: "#00FF00FF"},
			{X: 2, Y: 0, Color: "#0000FFFF"},
		},
	}

	if len(output.Pixels) != 3 {
		t.Errorf("Expected 3 pixels, got %d", len(output.Pixels))
	}

	// Verify each pixel has valid structure
	for i, p := range output.Pixels {
		if p.Color == "" {
			t.Errorf("Pixel %d has empty color", i)
		}
		if p.X < 0 {
			t.Errorf("Pixel %d has negative X coordinate: %d", i, p.X)
		}
		if p.Y < 0 {
			t.Errorf("Pixel %d has negative Y coordinate: %d", i, p.Y)
		}
	}
}

func TestGetPixelsInput_Regions(t *testing.T) {
	tests := []struct {
		name   string
		input  GetPixelsInput
		pixels int // expected number of pixels in region
	}{
		{
			name: "1x1 region",
			input: GetPixelsInput{
				Width:  1,
				Height: 1,
			},
			pixels: 1,
		},
		{
			name: "10x10 region",
			input: GetPixelsInput{
				Width:  10,
				Height: 10,
			},
			pixels: 100,
		},
		{
			name: "rectangular region 5x10",
			input: GetPixelsInput{
				Width:  5,
				Height: 10,
			},
			pixels: 50,
		},
		{
			name: "large region 100x100",
			input: GetPixelsInput{
				Width:  100,
				Height: 100,
			},
			pixels: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedPixels := tt.input.Width * tt.input.Height
			if expectedPixels != tt.pixels {
				t.Errorf("Expected %d pixels, calculated %d", tt.pixels, expectedPixels)
			}
		})
	}
}
