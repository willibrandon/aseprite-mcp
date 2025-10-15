package testutil

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"testing"
)

// TempSpriteDir returns a temporary directory for sprite files.
func TempSpriteDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// TempSpritePath returns a path for a temporary sprite file.
func TempSpritePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

// DecodeImage decodes an image from a reader and returns the image and format.
func DecodeImage(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}

// PixelData represents a pixel with coordinates and color for testing.
type PixelData struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

// ParsePixelData parses JSON pixel data returned by get_pixels.
func ParsePixelData(jsonData string) ([]PixelData, error) {
	var pixels []PixelData
	if err := json.Unmarshal([]byte(jsonData), &pixels); err != nil {
		return nil, fmt.Errorf("failed to parse pixel data: %w", err)
	}
	return pixels, nil
}

// FormatPixelPos formats a pixel position as a string for use as a map key.
func FormatPixelPos(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}
