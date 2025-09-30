package aseprite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Client executes Aseprite commands and Lua scripts.
type Client struct {
	execPath string
	tempDir  string
	timeout  time.Duration
}

// NewClient creates a new Aseprite client.
func NewClient(execPath, tempDir string, timeout time.Duration) *Client {
	return &Client{
		execPath: execPath,
		tempDir:  tempDir,
		timeout:  timeout,
	}
}

// ExecuteCommand runs an Aseprite command with the given arguments.
func (c *Client) ExecuteCommand(ctx context.Context, args []string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build command
	cmd := exec.CommandContext(ctx, c.execPath, args...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	// Check for errors
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("aseprite command timed out after %v", c.timeout)
		}

		// Include stderr in error message
		if stderr.Len() > 0 {
			return "", fmt.Errorf("aseprite command failed: %w\nstderr: %s\nstdout: %s", err, stderr.String(), stdout.String())
		}

		return "", fmt.Errorf("aseprite command failed: %w\nstdout: %s", err, stdout.String())
	}

	return stdout.String(), nil
}

// ExecuteLua executes a Lua script in Aseprite batch mode.
// If spritePath is non-empty, the sprite will be opened before running the script.
func (c *Client) ExecuteLua(ctx context.Context, script string, spritePath string) (string, error) {
	// Create temporary script file
	scriptPath, cleanup, err := c.createTempScript(script)
	if err != nil {
		return "", fmt.Errorf("failed to create temp script: %w", err)
	}
	defer cleanup()

	// Build arguments
	args := []string{"--batch"}

	// Add sprite path if specified
	if spritePath != "" {
		// Verify sprite exists
		if _, err := os.Stat(spritePath); os.IsNotExist(err) {
			return "", fmt.Errorf("sprite file not found: %s", spritePath)
		}
		args = append(args, spritePath)
	}

	// Add script argument
	args = append(args, "--script", scriptPath)

	// Execute command
	return c.ExecuteCommand(ctx, args)
}

// GetVersion retrieves the Aseprite version.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	output, err := c.ExecuteCommand(ctx, []string{"--version"})
	if err != nil {
		return "", err
	}

	// Parse version from output (format: "Aseprite 1.3.x")
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "", fmt.Errorf("failed to parse version from output: %s", output)
}

// createTempScript creates a temporary Lua script file.
// Returns the script path and a cleanup function.
func (c *Client) createTempScript(script string) (string, func(), error) {
	// Ensure temp directory exists
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create temp file with .lua extension
	tmpFile, err := os.CreateTemp(c.tempDir, "script-*.lua")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	scriptPath := tmpFile.Name()

	// Write script content
	if _, err := tmpFile.WriteString(script); err != nil {
		tmpFile.Close()
		os.Remove(scriptPath)
		return "", nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Close file
	if err := tmpFile.Close(); err != nil {
		os.Remove(scriptPath)
		return "", nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Remove(scriptPath)
	}

	return scriptPath, cleanup, nil
}

// CleanupOldTempFiles removes temporary files older than the specified duration.
func (c *Client) CleanupOldTempFiles(maxAge time.Duration) error {
	entries, err := os.ReadDir(c.tempDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, nothing to clean
		}
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		// Ignore directories
		if entry.IsDir() {
			continue
		}

		// Check if file matches our pattern (script-*.lua)
		if !strings.HasPrefix(entry.Name(), "script-") || !strings.HasSuffix(entry.Name(), ".lua") {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check age
		if now.Sub(info.ModTime()) > maxAge {
			filePath := filepath.Join(c.tempDir, entry.Name())
			os.Remove(filePath)
		}
	}

	return nil
}