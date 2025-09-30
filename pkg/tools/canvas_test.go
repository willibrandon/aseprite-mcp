package tools

import (
	"testing"
)

func TestCreateCanvasInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input CreateCanvasInput
		valid bool
	}{
		{
			name: "valid RGB canvas",
			input: CreateCanvasInput{
				Width:     800,
				Height:    600,
				ColorMode: "rgb",
			},
			valid: true,
		},
		{
			name: "valid grayscale canvas",
			input: CreateCanvasInput{
				Width:     100,
				Height:    100,
				ColorMode: "grayscale",
			},
			valid: true,
		},
		{
			name: "valid indexed canvas",
			input: CreateCanvasInput{
				Width:     320,
				Height:    240,
				ColorMode: "indexed",
			},
			valid: true,
		},
		{
			name: "minimum dimensions",
			input: CreateCanvasInput{
				Width:     1,
				Height:    1,
				ColorMode: "rgb",
			},
			valid: true,
		},
		{
			name: "maximum dimensions",
			input: CreateCanvasInput{
				Width:     65535,
				Height:    65535,
				ColorMode: "rgb",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation that struct can be created
			if tt.input.Width < 1 || tt.input.Width > 65535 {
				if tt.valid {
					t.Error("Expected valid input but width is out of range")
				}
			}
			if tt.input.Height < 1 || tt.input.Height > 65535 {
				if tt.valid {
					t.Error("Expected valid input but height is out of range")
				}
			}
			validModes := map[string]bool{"rgb": true, "grayscale": true, "indexed": true}
			if !validModes[tt.input.ColorMode] {
				if tt.valid {
					t.Error("Expected valid input but color mode is invalid")
				}
			}
		})
	}
}

func TestGenerateTimestamp(t *testing.T) {
	ts1 := generateTimestamp()
	ts2 := generateTimestamp()

	// Timestamps should be positive
	if ts1 <= 0 {
		t.Error("generateTimestamp() returned non-positive value")
	}

	// Second timestamp should be >= first (monotonic)
	if ts2 < ts1 {
		t.Error("generateTimestamp() not monotonic")
	}
}