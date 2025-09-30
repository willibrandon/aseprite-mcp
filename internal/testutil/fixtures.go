package testutil

import (
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
