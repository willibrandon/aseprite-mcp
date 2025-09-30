// Package testutil provides testing utilities for the Aseprite MCP server.
package testutil

import (
	"testing"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
)

// LoadTestConfig loads the test configuration from the standard config file.
// Tests MUST have a valid config file with real Aseprite path configured.
func LoadTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load test config: %v\n\nPlease ensure ~/.config/aseprite-mcp/config.json exists with aseprite_path configured.\nExample config:\n{\n  \"aseprite_path\": \"D:\\\\SRC\\\\aseprite\\\\build\\\\bin\\\\aseprite.exe\"\n}", err)
	}

	return cfg
}

// CreateTestConfigWithPath creates a test config with explicit Aseprite path.
// Use this when you need to override the default config for specific tests.
func CreateTestConfigWithPath(t *testing.T, asepritePath string) *config.Config {
	t.Helper()

	return &config.Config{
		AsepritePath: asepritePath,
		TempDir:      t.TempDir(),
		Timeout:      30 * time.Second,
		LogLevel:     "info",
	}
}