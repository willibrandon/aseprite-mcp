package aseprite

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Color represents an RGBA color value.
type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

var hexColorPattern = regexp.MustCompile(`^#?([A-Fa-f0-9]{6}|[A-Fa-f0-9]{8})$`)

// NewColor creates a new Color with the specified RGBA values.
func NewColor(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// NewColorRGB creates a new opaque Color with the specified RGB values.
func NewColorRGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, A: 255}
}

// FromHex parses a hex color string in the format "#RRGGBB" or "#RRGGBBAA".
// The "#" prefix is optional.
func (c *Color) FromHex(hex string) error {
	hex = strings.TrimPrefix(hex, "#")

	if !hexColorPattern.MatchString("#" + hex) {
		return fmt.Errorf("invalid hex color format: %q (expected #RRGGBB or #RRGGBBAA)", hex)
	}

	// Parse RGB
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)

	c.R = uint8(r)
	c.G = uint8(g)
	c.B = uint8(b)

	// Parse alpha if present
	if len(hex) == 8 {
		a, _ := strconv.ParseUint(hex[6:8], 16, 8)
		c.A = uint8(a)
	} else {
		c.A = 255 // Opaque by default
	}

	return nil
}

// ToHex converts the color to a hex string in the format "#RRGGBBAA".
func (c Color) ToHex() string {
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}

// ToHexRGB converts the color to a hex string in the format "#RRGGBB" (ignoring alpha).
func (c Color) ToHexRGB() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// Point represents a 2D coordinate.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Rectangle represents a rectangular region.
type Rectangle struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Pixel represents a single pixel with color and position.
type Pixel struct {
	Point
	Color Color `json:"color"`
}

// SpriteInfo contains metadata about a sprite.
type SpriteInfo struct {
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	ColorMode  string   `json:"color_mode"`
	FrameCount int      `json:"frame_count"`
	LayerCount int      `json:"layer_count"`
	Layers     []string `json:"layers"`
}

// ColorMode represents the color mode of a sprite.
type ColorMode string

const (
	ColorModeRGB       ColorMode = "rgb"
	ColorModeGrayscale ColorMode = "grayscale"
	ColorModeIndexed   ColorMode = "indexed"
)

// String returns the string representation of the color mode.
func (cm ColorMode) String() string {
	return string(cm)
}

// ToLua returns the Lua constant for the color mode.
func (cm ColorMode) ToLua() string {
	switch cm {
	case ColorModeRGB:
		return "ColorMode.RGB"
	case ColorModeGrayscale:
		return "ColorMode.GRAYSCALE"
	case ColorModeIndexed:
		return "ColorMode.INDEXED"
	default:
		return "ColorMode.RGB"
	}
}