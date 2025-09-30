package aseprite

import (
	"strings"
	"testing"
)

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple string",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "string with quotes",
			input: `hello "world"`,
			want:  `hello \"world\"`,
		},
		{
			name:  "string with backslash",
			input: `hello\world`,
			want:  `hello\\world`,
		},
		{
			name:  "string with newline",
			input: "hello\nworld",
			want:  `hello\nworld`,
		},
		{
			name:  "string with tab",
			input: "hello\tworld",
			want:  `hello\tworld`,
		},
		{
			name:  "complex string with path and quotes",
			input: `C:\path\to\file "with quotes"`,
			want:  `C:\\path\\to\\file \"with quotes\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapeString(tt.input); got != tt.want {
				t.Errorf("EscapeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatColor(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		want  string
	}{
		{
			name:  "red",
			color: Color{R: 255, G: 0, B: 0, A: 255},
			want:  "Color(255, 0, 0, 255)",
		},
		{
			name:  "green with alpha",
			color: Color{R: 0, G: 255, B: 0, A: 128},
			want:  "Color(0, 255, 0, 128)",
		},
		{
			name:  "transparent black",
			color: Color{R: 0, G: 0, B: 0, A: 0},
			want:  "Color(0, 0, 0, 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatColor(tt.color); got != tt.want {
				t.Errorf("FormatColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatPoint(t *testing.T) {
	p := Point{X: 10, Y: 20}
	want := "Point(10, 20)"

	if got := FormatPoint(p); got != want {
		t.Errorf("FormatPoint() = %v, want %v", got, want)
	}
}

func TestFormatRectangle(t *testing.T) {
	r := Rectangle{X: 10, Y: 20, Width: 30, Height: 40}
	want := "Rectangle(10, 20, 30, 40)"

	if got := FormatRectangle(r); got != want {
		t.Errorf("FormatRectangle() = %v, want %v", got, want)
	}
}

func TestLuaGenerator_CreateCanvas(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.CreateCanvas(800, 600, ColorModeRGB)

	// Verify script contains expected elements
	if !strings.Contains(script, "Sprite(800, 600, ColorMode.RGB)") {
		t.Error("script missing Sprite constructor call")
	}

	if !strings.Contains(script, "spr:saveAs(filename)") {
		t.Error("script missing saveAs call")
	}

	if !strings.Contains(script, "print(filename)") {
		t.Error("script missing print statement")
	}
}

func TestLuaGenerator_AddLayer(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.AddLayer("My Layer")

	// Verify script contains expected elements
	if !strings.Contains(script, "spr:newLayer()") {
		t.Error("script missing newLayer call")
	}

	if !strings.Contains(script, `layer.name = "My Layer"`) {
		t.Error("script missing layer name assignment")
	}

	if !strings.Contains(script, "app.transaction(function()") {
		t.Error("script not wrapped in transaction")
	}
}

func TestLuaGenerator_DrawPixels(t *testing.T) {
	gen := NewLuaGenerator()

	pixels := []Pixel{
		{Point: Point{X: 0, Y: 0}, Color: Color{R: 255, G: 0, B: 0, A: 255}},
		{Point: Point{X: 1, Y: 1}, Color: Color{R: 0, G: 255, B: 0, A: 255}},
	}

	script := gen.DrawPixels("Layer 1", 1, pixels)

	// Verify script contains expected elements
	if !strings.Contains(script, `findLayerByName("Layer 1")`) {
		t.Error("script missing layer lookup")
	}

	if !strings.Contains(script, "img:putPixel(0, 0, Color(255, 0, 0, 255))") {
		t.Error("script missing first pixel")
	}

	if !strings.Contains(script, "img:putPixel(1, 1, Color(0, 255, 0, 255))") {
		t.Error("script missing second pixel")
	}
}

func TestLuaGenerator_ExportSprite(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("export all frames", func(t *testing.T) {
		script := gen.ExportSprite("output.png", 0)

		if !strings.Contains(script, `saveCopyAs("output.png")`) {
			t.Error("script missing saveCopyAs call")
		}
	})

	t.Run("export specific frame", func(t *testing.T) {
		script := gen.ExportSprite("output.png", 2)

		if !strings.Contains(script, "spr.frames[2]") {
			t.Error("script missing frame selection")
		}

		if !strings.Contains(script, "frame = frame") {
			t.Error("script missing frame parameter")
		}
	})
}