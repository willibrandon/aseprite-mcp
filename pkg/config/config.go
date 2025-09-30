package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

// Load loads configuration from environment variables and optional config file.
// Environment variables take precedence over config file values.
func Load() (*Config, error) {
	cfg := &Config{
		Timeout:  DefaultTimeout,
		LogLevel: DefaultLogLevel,
	}

	// Try to load from config file
	if err := cfg.loadFromFile(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Override with environment variables
	cfg.loadFromEnv()

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

// loadFromFile loads configuration from the default config file location.
func (c *Config) loadFromFile() error {
	configPath := getConfigFilePath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

// loadFromEnv overrides configuration with environment variables.
func (c *Config) loadFromEnv() {
	if path := os.Getenv("ASEPRITE_PATH"); path != "" {
		c.AsepritePath = path
	}

	if dir := os.Getenv("ASEPRITE_TEMP_DIR"); dir != "" {
		c.TempDir = dir
	}

	if timeout := os.Getenv("ASEPRITE_TIMEOUT"); timeout != "" {
		if seconds, err := strconv.Atoi(timeout); err == nil {
			c.Timeout = time.Duration(seconds) * time.Second
		}
	}

	if level := os.Getenv("ASEPRITE_LOG_LEVEL"); level != "" {
		c.LogLevel = level
	}
}

// setDefaults sets default values for unset configuration fields.
func (c *Config) setDefaults() error {
	// Set default Aseprite path if not configured
	if c.AsepritePath == "" {
		path, err := findAsepritePath()
		if err != nil {
			return fmt.Errorf("could not find Aseprite executable: %w", err)
		}
		c.AsepritePath = path
	}

	// Set default temp directory if not configured
	if c.TempDir == "" {
		c.TempDir = filepath.Join(os.TempDir(), "aseprite-mcp")
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(c.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
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

// getConfigFilePath returns the default config file path.
func getConfigFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "aseprite-mcp", "config.json")
}

// findAsepritePath attempts to locate the Aseprite executable.
func findAsepritePath() (string, error) {
	// Try common installation locations first
	commonPaths := getCommonAsepritePaths()
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Search in PATH
	path, err := exec.LookPath("aseprite")
	if err != nil {
		return "", fmt.Errorf("aseprite not found in PATH or common locations")
	}

	return path, nil
}

// getCommonAsepritePaths returns common installation paths by OS.
func getCommonAsepritePaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`C:\Program Files\Aseprite\Aseprite.exe`,
			`C:\Program Files (x86)\Aseprite\Aseprite.exe`,
		}
	case "darwin":
		return []string{
			"/Applications/Aseprite.app/Contents/MacOS/aseprite",
			filepath.Join(os.Getenv("HOME"), "Applications", "Aseprite.app", "Contents", "MacOS", "aseprite"),
		}
	case "linux":
		return []string{
			"/usr/bin/aseprite",
			"/usr/local/bin/aseprite",
			filepath.Join(os.Getenv("HOME"), ".local", "bin", "aseprite"),
		}
	default:
		return nil
	}
}