package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateTempSprite creates a temporary .aseprite file for testing.
// Returns the file path and a cleanup function.
func CreateTempSprite(t *testing.T) (string, func()) {
	t.Helper()

	tempDir := t.TempDir()
	spritePath := filepath.Join(tempDir, "test.aseprite")

	// Create dummy sprite file (just needs to exist for most tests)
	if err := os.WriteFile(spritePath, []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to create temp sprite: %v", err)
	}

	cleanup := func() {
		os.Remove(spritePath)
	}

	return spritePath, cleanup
}

// CreateTestConfig creates a test configuration.
func CreateTestConfig(t *testing.T) (asepritePath, tempDir string) {
	t.Helper()

	tempDir = t.TempDir()

	// Create mock Aseprite
	mock, err := NewMockAseprite(tempDir)
	if err != nil {
		t.Fatalf("failed to create mock aseprite: %v", err)
	}

	return mock.Path(), tempDir
}