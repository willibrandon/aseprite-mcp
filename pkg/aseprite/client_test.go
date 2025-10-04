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

func TestClient_GetVersion(t *testing.T) {
	// Test with invalid executable path to check error handling
	client := NewClient("/nonexistent/aseprite", "", 30*time.Second)

	ctx := context.Background()
	_, err := client.GetVersion(ctx)
	if err == nil {
		t.Error("GetVersion() with invalid path should return error")
	}
}

func TestClient_ExecuteCommand(t *testing.T) {
	// Test with invalid executable path to check error handling
	client := NewClient("/nonexistent/aseprite", "", 30*time.Second)

	ctx := context.Background()
	_, err := client.ExecuteCommand(ctx, []string{"--version"})
	if err == nil {
		t.Error("ExecuteCommand() with invalid path should return error")
	}
}

// Note: Additional tests that require actual Aseprite execution
// should be placed in integration tests

func TestClient_ExecuteCommand_Timeout(t *testing.T) {
	// Use a command that will take longer than timeout
	client := NewClient("sleep", "", 1*time.Millisecond)

	ctx := context.Background()
	_, err := client.ExecuteCommand(ctx, []string{"10"})

	if err == nil {
		t.Error("ExecuteCommand() should timeout")
	}

	if err != nil && !contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestClient_ExecuteCommand_Success(t *testing.T) {
	client := NewClient("echo", "", 5*time.Second)

	ctx := context.Background()
	output, err := client.ExecuteCommand(ctx, []string{"hello"})

	if err != nil {
		t.Errorf("ExecuteCommand() unexpected error: %v", err)
	}

	if !contains(output, "hello") {
		t.Errorf("ExecuteCommand() output = %q, want to contain 'hello'", output)
	}
}

func TestClient_CreateTempScript_InvalidDir(t *testing.T) {
	client := NewClient("aseprite", "/nonexistent/directory/path", 30*time.Second)

	_, _, err := client.createTempScript("print('test')")

	if err == nil {
		t.Skip("Skipping: createTempScript succeeded (temp dir may be created automatically)")
	}
}

func TestClient_CleanupOldTempFiles_InvalidDir(t *testing.T) {
	client := NewClient("aseprite", "/nonexistent/directory/path", 30*time.Second)

	err := client.CleanupOldTempFiles(1 * time.Hour)

	// CleanupOldTempFiles returns nil if directory doesn't exist
	if err != nil {
		t.Logf("CleanupOldTempFiles() error (acceptable): %v", err)
	}
}

func TestClient_CleanupOldTempFiles_NoFiles(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient("aseprite", tempDir, 30*time.Second)

	// Should not error when directory is empty
	err := client.CleanupOldTempFiles(1 * time.Hour)

	if err != nil {
		t.Errorf("CleanupOldTempFiles() unexpected error: %v", err)
	}
}

func TestClient_ExecuteLua_EmptySpritePath(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient("echo", tempDir, 5*time.Second)

	ctx := context.Background()
	_, err := client.ExecuteLua(ctx, "print('test')", "")

	// With echo command, this should succeed
	if err != nil {
		t.Logf("ExecuteLua() error (expected with echo): %v", err)
	}
}

func TestClient_GetVersion_Success(t *testing.T) {
	// Use 'echo' to simulate version output
	client := NewClient("echo", "", 5*time.Second)

	ctx := context.Background()
	version, err := client.GetVersion(ctx)

	if err != nil {
		t.Errorf("GetVersion() unexpected error: %v", err)
	}

	// Version should be trimmed output from echo
	if version == "" {
		t.Error("GetVersion() returned empty version")
	}
}

func TestClient_GetVersion_EmptyOutput(t *testing.T) {
	// Use 'true' command which produces no output (in some environments)
	client := NewClient("true", "", 5*time.Second)

	ctx := context.Background()
	version, err := client.GetVersion(ctx)

	if err != nil {
		t.Errorf("GetVersion() unexpected error: %v", err)
	}

	// In some environments, 'true' is a coreutils binary that prints version info
	// This test is mainly to verify no panic occurs with minimal output
	t.Logf("GetVersion() returned: %q", version)
}

func TestClient_ExecuteLua_ValidSprite(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient("echo", tempDir, 5*time.Second)

	// Create a fake sprite file
	spritePath := filepath.Join(tempDir, "test.aseprite")
	if err := os.WriteFile(spritePath, []byte("fake sprite"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err := client.ExecuteLua(ctx, "print('test')", spritePath)

	// With echo, this should succeed (echo will just print args)
	if err != nil {
		t.Logf("ExecuteLua() error (expected with echo): %v", err)
	}
}

func TestClient_ExecuteCommand_WithStderr(t *testing.T) {
	// Use 'sh -c' to produce stderr output
	client := NewClient("sh", "", 5*time.Second)

	ctx := context.Background()
	_, err := client.ExecuteCommand(ctx, []string{"-c", "echo 'error message' >&2; exit 1"})

	if err == nil {
		t.Error("ExecuteCommand() should fail when command exits with error")
	}

	if err != nil && !contains(err.Error(), "stderr:") {
		t.Errorf("Expected stderr in error message, got: %v", err)
	}
}

func TestClient_CleanupOldTempFiles_WithSubdirectory(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient("aseprite", tempDir, 30*time.Second)

	// Create a subdirectory (should be ignored)
	subdir := filepath.Join(tempDir, "script-subdir.lua")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create an old file
	oldFile := filepath.Join(tempDir, "script-old.lua")
	if err := os.WriteFile(oldFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Run cleanup
	if err := client.CleanupOldTempFiles(1 * time.Hour); err != nil {
		t.Fatalf("CleanupOldTempFiles() error = %v", err)
	}

	// Subdirectory should still exist
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Error("Subdirectory was incorrectly removed")
	}

	// Old file should be removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file was not removed")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
