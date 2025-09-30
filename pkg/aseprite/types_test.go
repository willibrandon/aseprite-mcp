// Copyright 2025 Brandon Williams. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package aseprite

import (
	"testing"
)

func TestColor_FromHex(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		want    Color
		wantErr bool
	}{
		{
			name: "RGB with hash",
			hex:  "#FF0000",
			want: Color{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name: "RGB without hash",
			hex:  "00FF00",
			want: Color{R: 0, G: 255, B: 0, A: 255},
		},
		{
			name: "RGBA with hash",
			hex:  "#0000FF80",
			want: Color{R: 0, G: 0, B: 255, A: 128},
		},
		{
			name: "RGBA without hash",
			hex:  "FFFF00FF",
			want: Color{R: 255, G: 255, B: 0, A: 255},
		},
		{
			name:    "invalid format",
			hex:     "invalid",
			wantErr: true,
		},
		{
			name:    "too short",
			hex:     "#FFF",
			wantErr: true,
		},
		{
			name:    "too long",
			hex:     "#FFFFFFFFF",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Color
			err := c.FromHex(tt.hex)

			if (err != nil) != tt.wantErr {
				t.Errorf("FromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && c != tt.want {
				t.Errorf("FromHex() = %+v, want %+v", c, tt.want)
			}
		})
	}
}

func TestColor_ToHex(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		want  string
	}{
		{
			name:  "red",
			color: Color{R: 255, G: 0, B: 0, A: 255},
			want:  "#FF0000FF",
		},
		{
			name:  "green",
			color: Color{R: 0, G: 255, B: 0, A: 255},
			want:  "#00FF00FF",
		},
		{
			name:  "blue with alpha",
			color: Color{R: 0, G: 0, B: 255, A: 128},
			want:  "#0000FF80",
		},
		{
			name:  "black transparent",
			color: Color{R: 0, G: 0, B: 0, A: 0},
			want:  "#00000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.color.ToHex(); got != tt.want {
				t.Errorf("ToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorMode_ToLua(t *testing.T) {
	tests := []struct {
		name string
		cm   ColorMode
		want string
	}{
		{
			name: "RGB",
			cm:   ColorModeRGB,
			want: "ColorMode.RGB",
		},
		{
			name: "Grayscale",
			cm:   ColorModeGrayscale,
			want: "ColorMode.GRAYSCALE",
		},
		{
			name: "Indexed",
			cm:   ColorModeIndexed,
			want: "ColorMode.INDEXED",
		},
		{
			name: "Unknown defaults to RGB",
			cm:   ColorMode("unknown"),
			want: "ColorMode.RGB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cm.ToLua(); got != tt.want {
				t.Errorf("ToLua() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewColor(t *testing.T) {
	c := NewColor(255, 128, 64, 32)

	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 32 {
		t.Errorf("NewColor() = %+v, want R:255 G:128 B:64 A:32", c)
	}
}

func TestNewColorRGB(t *testing.T) {
	c := NewColorRGB(255, 128, 64)

	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 255 {
		t.Errorf("NewColorRGB() = %+v, want R:255 G:128 B:64 A:255", c)
	}
}