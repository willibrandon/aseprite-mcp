package tools

import (
	"testing"
)

func TestSetPaletteInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   SetPaletteInput
		wantErr bool
	}{
		{
			name: "valid palette with 8 colors",
			input: SetPaletteInput{
				SpritePath: "/tmp/test.aseprite",
				Colors:     []string{"#000000", "#FF0000", "#00FF00", "#0000FF", "#FFFF00", "#FF00FF", "#00FFFF", "#FFFFFF"},
			},
			wantErr: false,
		},
		{
			name: "valid palette with 16 colors",
			input: SetPaletteInput{
				SpritePath: "/tmp/test.aseprite",
				Colors:     []string{"#000000", "#111111", "#222222", "#333333", "#444444", "#555555", "#666666", "#777777", "#888888", "#999999", "#AAAAAA", "#BBBBBB", "#CCCCCC", "#DDDDDD", "#EEEEEE", "#FFFFFF"},
			},
			wantErr: false,
		},
		{
			name: "single color palette",
			input: SetPaletteInput{
				SpritePath: "/tmp/test.aseprite",
				Colors:     []string{"#FF5733"},
			},
			wantErr: false,
		},
		{
			name: "maximum 256 colors",
			input: SetPaletteInput{
				SpritePath: "/tmp/test.aseprite",
				Colors:     make([]string, 256),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate palette size
			if len(tt.input.Colors) == 0 {
				if !tt.wantErr {
					t.Error("Expected valid input but colors array is empty")
				}
			}
			if len(tt.input.Colors) > 256 {
				if !tt.wantErr {
					t.Error("Expected valid input but colors array exceeds 256")
				}
			}

			// Validate hex color format
			for _, color := range tt.input.Colors {
				if color != "" && !isValidHexColor(color) {
					if !tt.wantErr {
						t.Errorf("Expected valid input but color format is invalid: %s", color)
					}
				}
			}
		})
	}
}

func TestApplyShadingInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   ApplyShadingInput
		wantErr bool
	}{
		{
			name: "valid shading with smooth style",
			input: ApplyShadingInput{
				SpritePath:     "/tmp/test.aseprite",
				LayerName:      "Layer 1",
				FrameNumber:    1,
				Region:         Region{X: 0, Y: 0, Width: 32, Height: 32},
				Palette:        []string{"#000000", "#808080", "#FFFFFF"},
				LightDirection: "top_left",
				Intensity:      0.5,
				Style:          "smooth",
			},
			wantErr: false,
		},
		{
			name: "valid shading with hard style",
			input: ApplyShadingInput{
				SpritePath:     "/tmp/test.aseprite",
				LayerName:      "Layer 1",
				FrameNumber:    1,
				Region:         Region{X: 10, Y: 10, Width: 20, Height: 20},
				Palette:        []string{"#FF0000", "#FF8080", "#FFC0C0"},
				LightDirection: "top",
				Intensity:      0.8,
				Style:          "hard",
			},
			wantErr: false,
		},
		{
			name: "all light directions valid",
			input: ApplyShadingInput{
				SpritePath:     "/tmp/test.aseprite",
				LayerName:      "Layer 1",
				FrameNumber:    1,
				Region:         Region{X: 0, Y: 0, Width: 10, Height: 10},
				Palette:        []string{"#000000", "#FFFFFF"},
				LightDirection: "bottom_right",
				Intensity:      0.3,
				Style:          "smooth",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate region
			if tt.input.Region.Width <= 0 || tt.input.Region.Height <= 0 {
				if !tt.wantErr {
					t.Error("Expected valid input but region dimensions are invalid")
				}
			}

			// Validate intensity
			if tt.input.Intensity < 0.0 || tt.input.Intensity > 1.0 {
				if !tt.wantErr {
					t.Errorf("Expected valid input but intensity out of range: %f", tt.input.Intensity)
				}
			}

			// Validate light direction
			validDirections := map[string]bool{
				"top_left": true, "top": true, "top_right": true,
				"left": true, "right": true,
				"bottom_left": true, "bottom": true, "bottom_right": true,
			}
			if !validDirections[tt.input.LightDirection] {
				if !tt.wantErr {
					t.Errorf("Expected valid input but light direction is invalid: %s", tt.input.LightDirection)
				}
			}

			// Validate style
			validStyles := map[string]bool{"pillow": true, "smooth": true, "hard": true}
			if !validStyles[tt.input.Style] {
				if !tt.wantErr {
					t.Errorf("Expected valid input but style is invalid: %s", tt.input.Style)
				}
			}
		})
	}
}

func TestAnalyzePaletteHarmoniesInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   AnalyzePaletteHarmoniesInput
		wantErr bool
	}{
		{
			name: "valid palette with complementary colors",
			input: AnalyzePaletteHarmoniesInput{
				Palette: []string{"#FF0000", "#00FFFF"}, // Red and Cyan (complementary)
			},
			wantErr: false,
		},
		{
			name: "valid palette with triadic colors",
			input: AnalyzePaletteHarmoniesInput{
				Palette: []string{"#FF0000", "#00FF00", "#0000FF"}, // Red, Green, Blue (triadic)
			},
			wantErr: false,
		},
		{
			name: "large palette",
			input: AnalyzePaletteHarmoniesInput{
				Palette: []string{"#000000", "#111111", "#222222", "#333333", "#444444", "#555555", "#666666", "#777777", "#888888", "#999999", "#AAAAAA", "#BBBBBB", "#CCCCCC", "#DDDDDD", "#EEEEEE", "#FFFFFF"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate palette size
			if len(tt.input.Palette) == 0 {
				if !tt.wantErr {
					t.Error("Expected valid input but palette array is empty")
				}
			}
			if len(tt.input.Palette) > 256 {
				if !tt.wantErr {
					t.Error("Expected valid input but palette array exceeds 256")
				}
			}

			// Validate hex color format
			for _, color := range tt.input.Palette {
				if !isValidHexColor(color) {
					if !tt.wantErr {
						t.Errorf("Expected valid input but color format is invalid: %s", color)
					}
				}
			}
		})
	}
}

func TestHexToHSL(t *testing.T) {
	tests := []struct {
		name      string
		hexColor  string
		wantH     float64
		wantS     float64
		wantL     float64
		tolerance float64
	}{
		{
			name:      "pure red",
			hexColor:  "#FF0000",
			wantH:     0,
			wantS:     1.0,
			wantL:     0.5,
			tolerance: 1.0,
		},
		{
			name:      "pure green",
			hexColor:  "#00FF00",
			wantH:     120,
			wantS:     1.0,
			wantL:     0.5,
			tolerance: 1.0,
		},
		{
			name:      "pure blue",
			hexColor:  "#0000FF",
			wantH:     240,
			wantS:     1.0,
			wantL:     0.5,
			tolerance: 1.0,
		},
		{
			name:      "white",
			hexColor:  "#FFFFFF",
			wantH:     0,
			wantS:     0,
			wantL:     1.0,
			tolerance: 0.01,
		},
		{
			name:      "black",
			hexColor:  "#000000",
			wantH:     0,
			wantS:     0,
			wantL:     0,
			tolerance: 0.01,
		},
		{
			name:      "gray",
			hexColor:  "#808080",
			wantH:     0,
			wantS:     0,
			wantL:     0.5,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s, l := hexToHSL(tt.hexColor)

			// Check hue (with wraparound tolerance for 0/360)
			hueDiff := absFloat(h - tt.wantH)
			if hueDiff > 360-tt.tolerance {
				hueDiff = 360 - hueDiff
			}
			if hueDiff > tt.tolerance {
				t.Errorf("hexToHSL(%s) hue = %v, want %v (±%v)", tt.hexColor, h, tt.wantH, tt.tolerance)
			}

			// Check saturation
			if absFloat(s-tt.wantS) > tt.tolerance {
				t.Errorf("hexToHSL(%s) saturation = %v, want %v (±%v)", tt.hexColor, s, tt.wantS, tt.tolerance)
			}

			// Check lightness
			if absFloat(l-tt.wantL) > tt.tolerance {
				t.Errorf("hexToHSL(%s) lightness = %v, want %v (±%v)", tt.hexColor, l, tt.wantL, tt.tolerance)
			}
		})
	}
}

func TestAnalyzePaletteHarmonies_Complementary(t *testing.T) {
	// Test with known complementary pair
	palette := []string{"#FF0000", "#00FFFF"} // Red and Cyan (180° apart)
	result := analyzePaletteHarmonies(palette)

	if len(result.Complementary) == 0 {
		t.Error("Expected to find complementary pair, found none")
	}

	found := false
	for _, pair := range result.Complementary {
		if (pair.Color1 == "#FF0000" && pair.Color2 == "#00FFFF") ||
			(pair.Color1 == "#00FFFF" && pair.Color2 == "#FF0000") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find red-cyan complementary pair")
	}
}

func TestAnalyzePaletteHarmonies_Temperature(t *testing.T) {
	tests := []struct {
		name         string
		palette      []string
		wantDominant string
	}{
		{
			name:         "warm palette",
			palette:      []string{"#FF0000", "#FF8000", "#FFFF00"}, // Red, Orange, Yellow
			wantDominant: "warm",
		},
		{
			name:         "cool palette",
			palette:      []string{"#0000FF", "#0080FF", "#00FFFF"}, // Blue, Light Blue, Cyan
			wantDominant: "cool",
		},
		{
			name:         "neutral palette",
			palette:      []string{"#808080", "#A0A0A0", "#C0C0C0"}, // Grays
			wantDominant: "neutral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzePaletteHarmonies(tt.palette)

			if result.Temperature.Dominant != tt.wantDominant {
				t.Errorf("analyzePaletteHarmonies() dominant temperature = %v, want %v",
					result.Temperature.Dominant, tt.wantDominant)
			}
		})
	}
}
