package aseprite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClient_CreateTempScript(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient("aseprite", tempDir, 30*time.Second)

	script := "print('hello')"
	scriptPath, cleanup, err := client.createTempScript(script)
	if err != nil {
		t.Fatalf("createTempScript() error = %v", err)
	}
	defer cleanup()

	// Verify file was created
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Errorf("script file was not created at %s", scriptPath)
	}

	// Verify content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read script file: %v", err)
	}

	if string(content) != script {
		t.Errorf("script content = %q, want %q", string(content), script)
	}

	// Verify cleanup removes file
	cleanup()
	if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
		t.Errorf("script file was not cleaned up")
	}
}

func TestClient_CleanupOldTempFiles(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient("aseprite", tempDir, 30*time.Second)

	// Create some test files
	oldFile := filepath.Join(tempDir, "script-old.lua")
	newFile := filepath.Join(tempDir, "script-new.lua")
	otherFile := filepath.Join(tempDir, "other.txt")

	// Create files
	for _, file := range []string{oldFile, newFile, otherFile} {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Make old file actually old
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Run cleanup with 1 hour max age
	if err := client.CleanupOldTempFiles(1 * time.Hour); err != nil {
		t.Fatalf("CleanupOldTempFiles() error = %v", err)
	}

	// Verify old file was removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("old script file was not removed")
	}

	// Verify new file still exists
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Errorf("new script file was incorrectly removed")
	}

	// Verify other file still exists
	if _, err := os.Stat(otherFile); os.IsNotExist(err) {
		t.Errorf("non-script file was incorrectly removed")
	}
}

func TestClient_ExecuteLua_MissingSprite(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient("aseprite", tempDir, 30*time.Second)

	ctx := context.Background()
	_, err := client.ExecuteLua(ctx, "print('test')", "/nonexistent/sprite.aseprite")

	if err == nil {
		t.Error("ExecuteLua() expected error for missing sprite, got nil")
	}
}

// Note: Additional tests that require actual Aseprite execution
// should be placed in integration tests
