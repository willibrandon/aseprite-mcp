// Package config provides configuration management for the Aseprite MCP server.
//
// Configuration is loaded exclusively from a JSON file at ~/.config/pixel-mcp/config.json.
// No environment variables or auto-discovery mechanisms are used - all paths must be
// explicitly configured.
//
// Example config file:
//
//	{
//	  "aseprite_path": "/absolute/path/to/aseprite",
//	  "temp_dir": "/tmp/pixel-mcp",
//	  "timeout": 30,
//	  "log_level": "info",
//	  "log_file": "",
//	  "enable_timing": false
//	}
//
// The aseprite_path field is required and must point to a real Aseprite executable.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the Aseprite MCP server configuration.
//
// All fields must be explicitly set in the config file, except:
//   - TempDir defaults to /tmp/pixel-mcp if not specified
//   - Timeout defaults to 30 seconds if not specified
//   - LogLevel defaults to "info" if not specified
//   - LogFile defaults to empty (stderr only) if not specified
//   - EnableTiming defaults to false if not specified
//
// The AsepritePath is REQUIRED and must be an absolute path to a real executable.
type Config struct {
	// AsepritePath is the absolute path to the Aseprite executable.
	// REQUIRED. Must point to a real, executable file.
	AsepritePath string `json:"aseprite_path"`

	// TempDir is the directory for temporary Lua script files.
	// Defaults to /tmp/pixel-mcp if not specified.
	TempDir string `json:"temp_dir"`

	// Timeout is the maximum duration for Aseprite command execution.
	// Defaults to 30 seconds if not specified.
	Timeout time.Duration `json:"timeout"`

	// LogLevel is the logging verbosity level.
	// Valid values: "debug", "info", "warn", "error"
	// Defaults to "info" if not specified.
	LogLevel string `json:"log_level"`

	// LogFile is the optional path to a log file for persistent logging.
	// If empty, logs only go to stderr.
	// Defaults to empty string if not specified.
	LogFile string `json:"log_file"`

	// EnableTiming enables request tracking and operation timing for all tools.
	// When enabled, each operation gets a unique request ID and duration is logged.
	// Defaults to false if not specified.
	EnableTiming bool `json:"enable_timing"`
}

// Default configuration values applied when fields are not specified in the config file.
const (
	// DefaultTimeout is the default maximum duration for Aseprite operations (30 seconds)
	DefaultTimeout = 30 * time.Second

	// DefaultLogLevel is the default logging verbosity ("info")
	DefaultLogLevel = "info"
)

// Load loads configuration from the default config file at ~/.config/pixel-mcp/config.json.
//
// The config file MUST exist and MUST contain an explicit aseprite_path field.
// No environment variables or auto-discovery mechanisms are used.
//
// Returns an error if:
//   - Config file doesn't exist
//   - Config file is malformed JSON
//   - aseprite_path is not set
//   - aseprite_path doesn't point to a real executable
//   - Validation fails for any other field
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
	LogFile      string `json:"log_file"`
	EnableTiming bool   `json:"enable_timing"`
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
	c.LogFile = cj.LogFile
	c.EnableTiming = cj.EnableTiming

	return nil
}

// setDefaults sets default values for unset configuration fields.
//
// This method is called after loading from file to fill in any missing values.
// The aseprite_path field MUST be explicitly set and cannot be defaulted.
//
// Defaults applied:
//   - TempDir: /tmp/pixel-mcp (or OS temp dir + "pixel-mcp")
//   - Timeout: 30 seconds
//   - LogLevel: "info"
//
// Also creates the temp directory if it doesn't exist.
func (c *Config) setDefaults() error {
	// Aseprite path must be explicitly set in config file
	if c.AsepritePath == "" {
		return fmt.Errorf("aseprite_path must be explicitly configured in config file")
	}

	// Set default temp directory if not configured
	if c.TempDir == "" {
		c.TempDir = filepath.Join(os.TempDir(), "pixel-mcp")
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

// Validate checks if the configuration is valid and usable.
//
// Validation checks:
//   - Aseprite executable exists at the configured path
//   - Temp directory is writable
//   - Timeout is positive
//   - LogLevel is one of: debug, info, warn, error
//
// Returns an error if any validation check fails.
// This method is automatically called by Load() before returning the config.
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
	return filepath.Join(homeDir, ".config", "pixel-mcp", "config.json")
}
