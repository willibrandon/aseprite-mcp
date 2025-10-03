// Package config provides configuration management for the Aseprite MCP server.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the server configuration.
type Config struct {
	// AsepritePath is the path to the Aseprite executable.
	AsepritePath string `json:"aseprite_path"`

	// TempDir is the directory for temporary files.
	TempDir string `json:"temp_dir"`

	// Timeout is the maximum duration for operations.
	Timeout time.Duration `json:"timeout"`

	// LogLevel is the logging verbosity level.
	LogLevel string `json:"log_level"`
}

// Default configuration values.
const (
	DefaultTimeout  = 30 * time.Second
	DefaultLogLevel = "info"
)

// Load loads configuration from the default config file.
// The config file MUST contain an explicit aseprite_path.
func Load() (*Config, error) {
	cfg := &Config{
		Timeout:  DefaultTimeout,
		LogLevel: DefaultLogLevel,
	}

	// Try to load from config file
	if err := cfg.loadFromFile(); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s: %w\nPlease create config file with aseprite_path configured", getConfigFilePath(), err)
		}
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Set defaults for unset values
	if err := cfg.setDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// configJSON is a temporary struct for unmarshaling JSON with timeout as int (seconds)
type configJSON struct {
	AsepritePath string `json:"aseprite_path"`
	TempDir      string `json:"temp_dir"`
	Timeout      int    `json:"timeout"` // timeout in seconds
	LogLevel     string `json:"log_level"`
}

// loadFromFile loads configuration from the default config file location.
func (c *Config) loadFromFile() error {
	configPath := getConfigFilePath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Unmarshal into temporary struct
	var cj configJSON
	if err := json.Unmarshal(data, &cj); err != nil {
		return err
	}

	// Convert to Config with proper timeout conversion
	c.AsepritePath = cj.AsepritePath
	c.TempDir = cj.TempDir
	c.Timeout = time.Duration(cj.Timeout) * time.Second
	c.LogLevel = cj.LogLevel

	return nil
}

// setDefaults sets default values for unset configuration fields.
func (c *Config) setDefaults() error {
	// Aseprite path must be explicitly set in config file
	if c.AsepritePath == "" {
		return fmt.Errorf("aseprite_path must be explicitly configured in config file")
	}

	// Set default temp directory if not configured
	if c.TempDir == "" {
		c.TempDir = filepath.Join(os.TempDir(), "aseprite-mcp")
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(c.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Set default timeout if not set
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	// Set default log level if not set
	if c.LogLevel == "" {
		c.LogLevel = DefaultLogLevel
	}

	return nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Check if Aseprite executable exists
	if _, err := os.Stat(c.AsepritePath); os.IsNotExist(err) {
		return fmt.Errorf("aseprite executable not found at %s", c.AsepritePath)
	}

	// Check if temp directory is writable
	testFile := filepath.Join(c.TempDir, ".test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("temp directory is not writable: %w", err)
	}
	os.Remove(testFile)

	// Validate timeout
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", c.Timeout)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", c.LogLevel)
	}

	return nil
}

// getConfigFilePath is a function variable that returns the default config file path.
// Can be overridden in tests.
var getConfigFilePath = func() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "aseprite-mcp", "config.json")
}
