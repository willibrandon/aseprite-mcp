package testutil

import (
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