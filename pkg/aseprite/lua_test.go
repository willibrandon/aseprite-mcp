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

	script := gen.CreateCanvas(800, 600, ColorModeRGB, "test.aseprite")

	// Verify script contains expected elements
	if !strings.Contains(script, "Sprite(800, 600, ColorMode.RGB)") {
		t.Error("script missing Sprite constructor call")
	}

	if !strings.Contains(script, `spr:saveAs("test.aseprite")`) {
		t.Error("script missing saveAs call")
	}

	if !strings.Contains(script, `print("test.aseprite")`) {
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

	script := gen.DrawPixels("Layer 1", 1, pixels, false)

	// Verify script contains expected elements (using loop-based layer lookup)
	if !strings.Contains(script, `lyr.name == "Layer 1"`) {
		t.Error("script missing layer name check")
	}

	if !strings.Contains(script, "for i, lyr in ipairs(spr.layers)") {
		t.Error("script missing layer iteration")
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

		// Check that it creates a temporary sprite
		if !strings.Contains(script, "Sprite(spr.width, spr.height, spr.colorMode)") {
			t.Error("script missing temporary sprite creation")
		}

		// Check that it references the target frame
		if !strings.Contains(script, "spr.frames[2]") {
			t.Error("script missing target frame reference")
		}

		// Check that it flattens layers before export
		if !strings.Contains(script, "FlattenLayers") {
			t.Error("script missing flatten layers command")
		}

		// Check that it uses saveAs (not saveCopyAs which auto-numbers files)
		if !strings.Contains(script, `saveAs("output.png")`) {
			t.Error("script missing saveAs call")
		}

		// Check that it closes the temporary sprite
		if !strings.Contains(script, "close()") {
			t.Error("script missing sprite close")
		}
	})
}

func TestLuaGenerator_DeleteLayer(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.DeleteLayer("Background")

	// Verify script contains expected elements
	if !strings.Contains(script, `if #spr.layers == 1 then`) {
		t.Error("script missing last layer check")
	}

	if !strings.Contains(script, `if lyr.name == "Background"`) {
		t.Error("script missing layer name check")
	}

	if !strings.Contains(script, `spr:deleteLayer(layer)`) {
		t.Error("script missing deleteLayer call")
	}

	if !strings.Contains(script, "app.transaction(function()") {
		t.Error("script not wrapped in transaction")
	}
}

func TestLuaGenerator_DeleteFrame(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.DeleteFrame(2)

	// Verify script contains expected elements
	if !strings.Contains(script, `if #spr.frames == 1 then`) {
		t.Error("script missing last frame check")
	}

	if !strings.Contains(script, "spr.frames[2]") {
		t.Error("script missing frame reference")
	}

	if !strings.Contains(script, `spr:deleteFrame(frame)`) {
		t.Error("script missing deleteFrame call")
	}

	if !strings.Contains(script, "app.transaction(function()") {
		t.Error("script not wrapped in transaction")
	}
}

func TestLuaGenerator_DrawContour(t *testing.T) {
	gen := NewLuaGenerator()

	points := []Point{
		{X: 10, Y: 10},
		{X: 50, Y: 10},
		{X: 50, Y: 50},
		{X: 10, Y: 50},
	}
	color := Color{R: 255, G: 0, B: 0, A: 255}

	t.Run("open contour", func(t *testing.T) {
		script := gen.DrawContour("Layer 1", 1, points, color, 2, false, false)

		// Verify script contains expected elements
		if !strings.Contains(script, `lyr.name == "Layer 1"`) {
			t.Error("script missing layer name check")
		}

		if !strings.Contains(script, "Brush(2)") {
			t.Error("script missing brush thickness")
		}

		if !strings.Contains(script, "Color(255, 0, 0, 255)") {
			t.Error("script missing color")
		}

		if !strings.Contains(script, "Point(10, 10)") {
			t.Error("script missing first point")
		}

		if !strings.Contains(script, "Point(50, 50)") {
			t.Error("script missing last point")
		}

		if !strings.Contains(script, "app.transaction(function()") {
			t.Error("script not wrapped in transaction")
		}

		// Should NOT have closing line for open contour
		if strings.Contains(script, "-- Close the contour") {
			t.Error("open contour should not have closing line")
		}
	})

	t.Run("closed contour (polygon)", func(t *testing.T) {
		script := gen.DrawContour("Layer 1", 1, points, color, 2, true, false)

		// Should have closing line connecting last to first
		if !strings.Contains(script, "-- Close the contour") {
			t.Error("closed contour missing closing line comment")
		}

		// Verify it connects last point back to first
		if !strings.Contains(script, "Point(10, 50)") || !strings.Contains(script, "Point(10, 10)") {
			t.Error("closed contour should connect last point to first")
		}
	})

	t.Run("with palette", func(t *testing.T) {
		script := gen.DrawContour("Layer 1", 1, points, color, 2, false, true)

		// Should include palette snapper helper
		if !strings.Contains(script, "function snapToPalette") {
			t.Error("script with use_palette missing snapToPalette helper")
		}

		// The color variable should be assigned the result of snapToPalette (with RGBA values, not Color())
		if !strings.Contains(script, "local color = snapToPalette(255, 0, 0, 255)") {
			t.Error("script should use snapToPalette for color assignment")
		}
	})
}

func TestLuaGenerator_SelectRectangle(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("replace mode", func(t *testing.T) {
		script := gen.SelectRectangle(10, 20, 30, 40, "replace")

		if !strings.Contains(script, "Rectangle(10, 20, 30, 40)") {
			t.Error("script missing rectangle coordinates")
		}

		if !strings.Contains(script, `if "replace" == "replace"`) {
			t.Error("script missing replace mode check")
		}

		if !strings.Contains(script, "Rectangle selection created successfully") {
			t.Error("script missing success message")
		}
	})

	t.Run("add mode", func(t *testing.T) {
		script := gen.SelectRectangle(5, 5, 10, 10, "add")

		if !strings.Contains(script, `if "add" == "replace"`) {
			t.Error("script missing mode check")
		}
	})
}

func TestLuaGenerator_SelectEllipse(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.SelectEllipse(10, 10, 50, 30, "replace")

	// Check that it uses width/height for radii calculation
	if !strings.Contains(script, "local rx = 50 / 2") {
		t.Error("script missing rx calculation")
	}

	if !strings.Contains(script, "local ry = 30 / 2") {
		t.Error("script missing ry calculation")
	}

	if !strings.Contains(script, "Ellipse selection created successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_SelectAll(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.SelectAll()

	if !strings.Contains(script, "Rectangle(0, 0, spr.width, spr.height)") {
		t.Error("script missing rectangle covering entire sprite")
	}

	if !strings.Contains(script, "spr.selection = sel") {
		t.Error("script missing selection assignment")
	}

	if !strings.Contains(script, "Select all completed successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_Deselect(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.Deselect()

	if !strings.Contains(script, "app.command.DeselectMask()") {
		t.Error("script missing DeselectMask command")
	}

	if !strings.Contains(script, "Deselect completed successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_MoveSelection(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.MoveSelection(15, -10)

	if !strings.Contains(script, "if spr.selection.isEmpty then") {
		t.Error("script missing empty selection check")
	}

	if !strings.Contains(script, "bounds.x + 15") {
		t.Error("script missing dx offset")
	}

	if !strings.Contains(script, "bounds.y + -10") {
		t.Error("script missing dy offset")
	}

	if !strings.Contains(script, "Selection moved successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_CutSelection(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.CutSelection("Layer 1", 2)

	if !strings.Contains(script, "if spr.selection.isEmpty then") {
		t.Error("script missing empty selection check")
	}

	if !strings.Contains(script, `lyr.name == "Layer 1"`) {
		t.Error("script missing layer name check")
	}

	if !strings.Contains(script, "spr.frames[2]") {
		t.Error("script missing frame reference")
	}

	if !strings.Contains(script, "app.command.Cut()") {
		t.Error("script missing Cut command")
	}

	if !strings.Contains(script, "Cut selection completed successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_CopySelection(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.CopySelection()

	if !strings.Contains(script, "if spr.selection.isEmpty then") {
		t.Error("script missing empty selection check")
	}

	if !strings.Contains(script, "app.command.Copy()") {
		t.Error("script missing Copy command")
	}

	if !strings.Contains(script, "Copy selection completed successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_PasteClipboard(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("without position", func(t *testing.T) {
		script := gen.PasteClipboard("Layer 1", 1, nil, nil)

		if !strings.Contains(script, `lyr.name == "Layer 1"`) {
			t.Error("script missing layer name check")
		}

		if !strings.Contains(script, "app.command.Paste()") {
			t.Error("script missing Paste command")
		}

		if strings.Contains(script, "Set paste position") {
			t.Error("script should not set position when nil")
		}

		if !strings.Contains(script, "Paste completed successfully") {
			t.Error("script missing success message")
		}
	})

	t.Run("with position", func(t *testing.T) {
		x, y := 25, 35
		script := gen.PasteClipboard("Layer 1", 1, &x, &y)

		if !strings.Contains(script, "app.command.Paste { x = 25, y = 35 }") {
			t.Error("script missing Paste command with position parameters")
		}

		if !strings.Contains(script, "Paste completed successfully") {
			t.Error("script missing success message")
		}
	})
}

func TestLuaGenerator_GetPalette(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.GetPalette()

	if !strings.Contains(script, "local palette = spr.palettes[1]") {
		t.Error("script missing palette retrieval")
	}

	if !strings.Contains(script, "palette:getColor(i)") {
		t.Error("script missing getColor call")
	}

	if !strings.Contains(script, `string.format("#%02X%02X%02X"`) {
		t.Error("script missing hex color formatting")
	}

	if !strings.Contains(script, `"colors":`) {
		t.Error("script missing JSON colors field")
	}

	if !strings.Contains(script, `"size":`) {
		t.Error("script missing JSON size field")
	}
}

func TestLuaGenerator_SetPaletteColor(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.SetPaletteColor(5, "#FF0000")

	if !strings.Contains(script, "local palette = spr.palettes[1]") {
		t.Error("script missing palette retrieval")
	}

	if !strings.Contains(script, "if 5 < 0 or 5 >= #palette then") {
		t.Error("script missing index validation")
	}

	if !strings.Contains(script, "palette:setColor(5, Color{r=255, g=0, b=0, a=255})") {
		t.Error("script missing setColor call with correct values")
	}

	if !strings.Contains(script, "Palette color set successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_AddPaletteColor(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.AddPaletteColor("#00FF00")

	if !strings.Contains(script, "local palette = spr.palettes[1]") {
		t.Error("script missing palette retrieval")
	}

	if !strings.Contains(script, "if #palette >= 256 then") {
		t.Error("script missing maximum size check")
	}

	if !strings.Contains(script, "palette:resize(#palette + 1)") {
		t.Error("script missing palette resize")
	}

	if !strings.Contains(script, "Color{r=0, g=255, b=0, a=255}") {
		t.Error("script missing color with correct values")
	}

	if !strings.Contains(script, `"color_index":`) {
		t.Error("script missing JSON color_index field")
	}
}

func TestLuaGenerator_SortPalette(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("sort by hue ascending", func(t *testing.T) {
		script := gen.SortPalette("hue", true)

		if !strings.Contains(script, "function rgbToHSL(r, g, b)") {
			t.Error("script missing RGB to HSL conversion function")
		}

		if !strings.Contains(script, "local h, s, l = rgbToHSL") {
			t.Error("script missing HSL calculation")
		}

		if !strings.Contains(script, "return a.h < b.h") {
			t.Error("script missing hue comparison for ascending sort")
		}

		if !strings.Contains(script, "Palette sorted by hue successfully") {
			t.Error("script missing success message")
		}
	})

	t.Run("sort by saturation descending", func(t *testing.T) {
		script := gen.SortPalette("saturation", false)

		if !strings.Contains(script, "return a.s > b.s") {
			t.Error("script missing saturation comparison for descending sort")
		}

		if !strings.Contains(script, "Palette sorted by saturation successfully") {
			t.Error("script missing success message")
		}
	})

	t.Run("sort by brightness ascending", func(t *testing.T) {
		script := gen.SortPalette("brightness", true)

		if !strings.Contains(script, "return a.l < b.l") {
			t.Error("script missing lightness comparison for ascending sort")
		}

		if !strings.Contains(script, "Palette sorted by brightness successfully") {
			t.Error("script missing success message")
		}
	})

	t.Run("sort by luminance", func(t *testing.T) {
		script := gen.SortPalette("luminance", true)

		if !strings.Contains(script, "return a.l < b.l") {
			t.Error("script missing lightness comparison")
		}

		if !strings.Contains(script, "Palette sorted by luminance successfully") {
			t.Error("script missing success message")
		}
	})
}

func TestLuaGenerator_FlipSprite(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("flip horizontal sprite", func(t *testing.T) {
		script := gen.FlipSprite("horizontal", "sprite")

		if !strings.Contains(script, `orientation = "horizontal"`) {
			t.Error("script missing horizontal orientation")
		}

		if !strings.Contains(script, `target = 'sprite'`) {
			t.Error("script missing sprite target")
		}

		if !strings.Contains(script, "app.command.Flip") {
			t.Error("script missing Flip command")
		}

		if !strings.Contains(script, "Sprite flipped horizontal successfully") {
			t.Error("script missing success message")
		}
	})

	t.Run("flip vertical layer", func(t *testing.T) {
		script := gen.FlipSprite("vertical", "layer")

		if !strings.Contains(script, `orientation = "vertical"`) {
			t.Error("script missing vertical orientation")
		}

		if !strings.Contains(script, `target = 'layer'`) {
			t.Error("script missing layer target")
		}
	})

	t.Run("invalid direction defaults to horizontal", func(t *testing.T) {
		script := gen.FlipSprite("invalid", "sprite")

		if !strings.Contains(script, `orientation = "horizontal"`) {
			t.Error("invalid direction should default to horizontal")
		}
	})
}

func TestLuaGenerator_RotateSprite(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("rotate 90 degrees", func(t *testing.T) {
		script := gen.RotateSprite(90, "sprite")

		if !strings.Contains(script, "angle = 90") {
			t.Error("script missing 90 degree angle")
		}

		if !strings.Contains(script, "app.command.Rotate") {
			t.Error("script missing Rotate command")
		}

		if !strings.Contains(script, "Sprite rotated 90 degrees successfully") {
			t.Error("script missing success message")
		}
	})

	t.Run("rotate 180 degrees", func(t *testing.T) {
		script := gen.RotateSprite(180, "sprite")

		if !strings.Contains(script, "angle = 180") {
			t.Error("script missing 180 degree angle")
		}
	})

	t.Run("rotate 270 degrees with cel target", func(t *testing.T) {
		script := gen.RotateSprite(270, "cel")

		if !strings.Contains(script, "angle = 270") {
			t.Error("script missing 270 degree angle")
		}

		if !strings.Contains(script, `target = 'cel'`) {
			t.Error("script missing cel target")
		}
	})

	t.Run("invalid angle defaults to 90", func(t *testing.T) {
		script := gen.RotateSprite(45, "sprite")

		if !strings.Contains(script, "angle = 90") {
			t.Error("invalid angle should default to 90")
		}
	})
}

func TestLuaGenerator_ScaleSprite(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("scale with nearest neighbor", func(t *testing.T) {
		script := gen.ScaleSprite(2.0, 2.0, "nearest")

		if !strings.Contains(script, "oldWidth * 2.000") {
			t.Error("script missing scale factor calculation")
		}

		if !strings.Contains(script, `method = "nearest"`) {
			t.Error("script missing nearest algorithm")
		}

		if !strings.Contains(script, "app.command.SpriteSize") {
			t.Error("script missing SpriteSize command")
		}

		if !strings.Contains(script, `"success":true`) {
			t.Error("script missing JSON output")
		}
	})

	t.Run("scale with bilinear", func(t *testing.T) {
		script := gen.ScaleSprite(0.5, 0.5, "bilinear")

		if !strings.Contains(script, `method = "bilinear"`) {
			t.Error("script missing bilinear algorithm")
		}
	})

	t.Run("scale with rotsprite", func(t *testing.T) {
		script := gen.ScaleSprite(1.5, 1.5, "rotsprite")

		if !strings.Contains(script, `method = "rotsprite"`) {
			t.Error("script missing rotsprite algorithm")
		}
	})

	t.Run("invalid algorithm defaults to nearest", func(t *testing.T) {
		script := gen.ScaleSprite(2.0, 2.0, "invalid")

		if !strings.Contains(script, `method = "nearest"`) {
			t.Error("invalid algorithm should default to nearest")
		}
	})
}

func TestLuaGenerator_CropSprite(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.CropSprite(10, 20, 100, 80)

	// Check validation
	if !strings.Contains(script, "if 10 < 0 or 20 < 0") {
		t.Error("script missing position validation")
	}

	if !strings.Contains(script, "if 100 <= 0 or 80 <= 0") {
		t.Error("script missing dimension validation")
	}

	// Check bounds checking
	if !strings.Contains(script, "if 10 + 100 > spr.width or 20 + 80 > spr.height") {
		t.Error("script missing bounds validation")
	}

	// Check command
	if !strings.Contains(script, "app.command.CropSprite") {
		t.Error("script missing CropSprite command")
	}

	if !strings.Contains(script, "bounds = Rectangle(10, 20, 100, 80)") {
		t.Error("script missing Rectangle bounds")
	}

	if !strings.Contains(script, "Sprite cropped successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_ResizeCanvas(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("resize with center anchor", func(t *testing.T) {
		script := gen.ResizeCanvas(200, 150, "center")

		if !strings.Contains(script, "local newWidth = 200") {
			t.Error("script missing new width")
		}

		if !strings.Contains(script, "local newHeight = 150") {
			t.Error("script missing new height")
		}

		if !strings.Contains(script, "left = math.floor((newWidth - oldWidth) / 2)") {
			t.Error("script missing center left offset")
		}

		if !strings.Contains(script, "top = math.floor((newHeight - oldHeight) / 2)") {
			t.Error("script missing center top offset")
		}

		if !strings.Contains(script, "app.command.CanvasSize") {
			t.Error("script missing CanvasSize command")
		}
	})

	t.Run("resize with top_left anchor", func(t *testing.T) {
		script := gen.ResizeCanvas(200, 150, "top_left")

		if !strings.Contains(script, "left = 0") {
			t.Error("script missing top_left left offset")
		}

		if !strings.Contains(script, "top = 0") {
			t.Error("script missing top_left top offset")
		}
	})

	t.Run("resize with bottom_right anchor", func(t *testing.T) {
		script := gen.ResizeCanvas(200, 150, "bottom_right")

		if !strings.Contains(script, "left = newWidth - oldWidth") {
			t.Error("script missing bottom_right left offset")
		}

		if !strings.Contains(script, "top = newHeight - oldHeight") {
			t.Error("script missing bottom_right top offset")
		}
	})

	t.Run("invalid anchor defaults to center", func(t *testing.T) {
		script := gen.ResizeCanvas(200, 150, "invalid")

		if !strings.Contains(script, "math.floor((newWidth - oldWidth) / 2)") {
			t.Error("invalid anchor should default to center")
		}
	})
}

func TestLuaGenerator_ApplyOutline(t *testing.T) {
	gen := NewLuaGenerator()

	color := Color{R: 255, G: 0, B: 0, A: 255}
	script := gen.ApplyOutline("Layer 1", 1, color, 2)

	// Check layer lookup
	if !strings.Contains(script, `for i, lyr in ipairs(spr.layers)`) {
		t.Error("script missing layer iteration")
	}

	if !strings.Contains(script, `if lyr.name == "Layer 1"`) {
		t.Error("script missing layer name comparison")
	}

	// Check frame lookup
	if !strings.Contains(script, "spr.frames[1]") {
		t.Error("script missing frame lookup")
	}

	// Check cel existence check
	if !strings.Contains(script, "if not cel then") {
		t.Error("script missing cel existence check")
	}

	// Check Outline command
	if !strings.Contains(script, "app.command.Outline") {
		t.Error("script missing Outline command")
	}

	if !strings.Contains(script, "color = Color(255, 0, 0, 255)") {
		t.Error("script missing color parameter")
	}

	if !strings.Contains(script, "size = 2") {
		t.Error("script missing size parameter")
	}

	if !strings.Contains(script, "Outline applied successfully") {
		t.Error("script missing success message")
	}
}

func TestLuaGenerator_ExportSpritesheet(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("export with horizontal layout", func(t *testing.T) {
		script := gen.ExportSpritesheet("/tmp/output.png", "horizontal", 2, false)

		if !strings.Contains(script, `type = "horizontal"`) {
			t.Error("script missing horizontal layout")
		}

		if !strings.Contains(script, `textureFilename = outputPath`) {
			t.Error("script missing texture filename")
		}

		if !strings.Contains(script, "borderPadding = 2") {
			t.Error("script missing padding")
		}

		if !strings.Contains(script, "app.command.ExportSpriteSheet") {
			t.Error("script missing ExportSpriteSheet command")
		}
	})

	t.Run("export with JSON metadata", func(t *testing.T) {
		script := gen.ExportSpritesheet("/tmp/output.png", "packed", 0, true)

		if !strings.Contains(script, `dataFormat = "json"`) {
			t.Error("script missing JSON data format")
		}

		if !strings.Contains(script, `metadata_path`) {
			t.Error("script missing metadata path in output")
		}
	})

	t.Run("invalid layout defaults to horizontal", func(t *testing.T) {
		script := gen.ExportSpritesheet("/tmp/output.png", "invalid", 0, false)

		if !strings.Contains(script, `type = "horizontal"`) {
			t.Error("invalid layout should default to horizontal")
		}
	})
}

func TestLuaGenerator_ImportImage(t *testing.T) {
	gen := NewLuaGenerator()

	t.Run("import with default position", func(t *testing.T) {
		script := gen.ImportImage("/tmp/image.png", "Imported Layer", 1, nil, nil)

		if !strings.Contains(script, `Image{ fromFile = "/tmp/image.png" }`) {
			t.Error("script missing image load")
		}

		if !strings.Contains(script, `if lyr.name == "Imported Layer"`) {
			t.Error("script missing layer name check")
		}

		if !strings.Contains(script, "Point(0, 0)") {
			t.Error("script missing default position")
		}

		if !strings.Contains(script, "spr:newCel(layer, frame)") {
			t.Error("script missing cel creation")
		}
	})

	t.Run("import with custom position", func(t *testing.T) {
		x := 50
		y := 100
		script := gen.ImportImage("/tmp/image.png", "Layer 1", 2, &x, &y)

		if !strings.Contains(script, "Point(50, 100)") {
			t.Error("script missing custom position")
		}

		if !strings.Contains(script, "spr.frames[2]") {
			t.Error("script missing frame number")
		}
	})
}

func TestLuaGenerator_SaveAs(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.SaveAs("/tmp/new-sprite.aseprite")

	if !strings.Contains(script, `local newPath = "/tmp/new-sprite.aseprite"`) {
		t.Error("script missing escaped path")
	}

	if !strings.Contains(script, "spr:saveAs(newPath)") {
		t.Error("script missing saveAs call")
	}

	if !strings.Contains(script, `"success":true`) {
		t.Error("script missing success flag")
	}

	if !strings.Contains(script, `"file_path"`) {
		t.Error("script missing file_path in output")
	}
}

func TestLuaGenerator_DeleteTag(t *testing.T) {
	gen := NewLuaGenerator()

	script := gen.DeleteTag("walk")

	// Check tag iteration
	if !strings.Contains(script, "for i, t in ipairs(spr.tags)") {
		t.Error("script missing tag iteration")
	}

	if !strings.Contains(script, `if t.name == "walk"`) {
		t.Error("script missing tag name check")
	}

	// Check error handling
	if !strings.Contains(script, `error("Tag not found: walk")`) {
		t.Error("script missing tag not found error")
	}

	// Check deletion
	if !strings.Contains(script, "spr:deleteTag(tag)") {
		t.Error("script missing tag deletion")
	}

	if !strings.Contains(script, "app.transaction") {
		t.Error("script missing transaction wrapper")
	}

	if !strings.Contains(script, "Tag deleted successfully") {
		t.Error("script missing success message")
	}
}
