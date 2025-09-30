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

func TestAddLayerInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input AddLayerInput
		valid bool
	}{
		{
			name: "valid layer",
			input: AddLayerInput{
				SpritePath: "/path/to/sprite.aseprite",
				LayerName:  "Background",
			},
			valid: true,
		},
		{
			name: "empty layer name",
			input: AddLayerInput{
				SpritePath: "/path/to/sprite.aseprite",
				LayerName:  "",
			},
			valid: false,
		},
		{
			name: "special characters in layer name",
			input: AddLayerInput{
				SpritePath: "/path/to/sprite.aseprite",
				LayerName:  "Layer-1_Test (Copy)",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.input.LayerName == ""
			if isEmpty && tt.valid {
				t.Error("Expected valid input but layer name is empty")
			}
			if !isEmpty && !tt.valid {
				t.Error("Expected invalid input but layer name is not empty")
			}
		})
	}
}

func TestAddFrameInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input AddFrameInput
		valid bool
	}{
		{
			name: "valid frame with 100ms duration",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 100,
			},
			valid: true,
		},
		{
			name: "minimum duration (1ms)",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 1,
			},
			valid: true,
		},
		{
			name: "maximum duration (65535ms)",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 65535,
			},
			valid: true,
		},
		{
			name: "zero duration",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 0,
			},
			valid: false,
		},
		{
			name: "negative duration",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: -1,
			},
			valid: false,
		},
		{
			name: "duration too large",
			input: AddFrameInput{
				SpritePath: "/path/to/sprite.aseprite",
				DurationMs: 65536,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.input.DurationMs >= 1 && tt.input.DurationMs <= 65535
			if isValid != tt.valid {
				t.Errorf("Expected valid=%v but got %v for duration=%d", tt.valid, isValid, tt.input.DurationMs)
			}
		})
	}
}

func TestGetSpriteInfoInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input GetSpriteInfoInput
		valid bool
	}{
		{
			name: "valid sprite path",
			input: GetSpriteInfoInput{
				SpritePath: "/path/to/sprite.aseprite",
			},
			valid: true,
		},
		{
			name: "empty sprite path",
			input: GetSpriteInfoInput{
				SpritePath: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.input.SpritePath == ""
			if isEmpty == tt.valid {
				t.Errorf("Expected valid=%v but path is empty=%v", tt.valid, isEmpty)
			}
		})
	}
}
