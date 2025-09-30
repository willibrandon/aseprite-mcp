package testutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// MockAseprite represents a mock Aseprite executable for testing.
type MockAseprite struct {
	execPath  string
	responses map[string]string
}

// NewMockAseprite creates a new mock Aseprite executable.
func NewMockAseprite(dir string) (*MockAseprite, error) {
	// Create mock executable script
	execPath := filepath.Join(dir, "aseprite.exe")
	if err := createMockScript(execPath); err != nil {
		return nil, err
	}

	return &MockAseprite{
		execPath:  execPath,
		responses: make(map[string]string),
	}, nil
}

// Path returns the path to the mock executable.
func (m *MockAseprite) Path() string {
	return m.execPath
}

// SetResponse sets the mock response for a specific script pattern.
func (m *MockAseprite) SetResponse(pattern, response string) {
	m.responses[pattern] = response
}

// createMockScript creates a Windows batch script that acts as mock Aseprite.
func createMockScript(path string) error {
	script := `@echo off
if "%1"=="--version" (
	echo Aseprite 1.3.0
	exit /b 0
)
if "%1"=="--batch" (
	REM Mock batch mode - just echo success
	echo Mock Aseprite batch execution
	exit /b 0
)
echo Aseprite 1.3.0
exit /b 0
`

	// Write script
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to create mock script: %w", err)
	}

	return nil
}