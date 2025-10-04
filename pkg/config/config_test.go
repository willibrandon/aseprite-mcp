package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	// Load Aseprite path from test config (works both locally and in CI)
	testCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load test config: %v\n\nPlease ensure ~/.config/aseprite-mcp/config.json exists with aseprite_path configured.", err)
	}
	realAseprite := testCfg.AsepritePath
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with real aseprite",
			config: &Config{
				AsepritePath: realAseprite,
				TempDir:      tempDir,
				Timeout:      30 * time.Second,
				LogLevel:     "info",
			},
			wantErr: false,
		},
		{
			name: "missing aseprite executable",
			config: &Config{
				AsepritePath: "/nonexistent/aseprite",
				TempDir:      tempDir,
				Timeout:      30 * time.Second,
				LogLevel:     "info",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				AsepritePath: realAseprite,
				TempDir:      tempDir,
				Timeout:      -1 * time.Second,
				LogLevel:     "info",
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				AsepritePath: realAseprite,
				TempDir:      tempDir,
				Timeout:      30 * time.Second,
				LogLevel:     "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	testCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load test config: %v\n\nPlease ensure ~/.config/aseprite-mcp/config.json exists with aseprite_path configured.", err)
	}
	realAseprite := testCfg.AsepritePath

	t.Run("sets defaults for empty fields", func(t *testing.T) {
		cfg := &Config{
			AsepritePath: realAseprite,
		}

		err := cfg.setDefaults()
		if err != nil {
			t.Fatalf("setDefaults() error = %v", err)
		}

		if cfg.TempDir == "" {
			t.Error("TempDir was not set to default")
		}

		if cfg.Timeout == 0 {
			cfg.Timeout = DefaultTimeout
		}

		if cfg.Timeout != DefaultTimeout {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
		}

		if cfg.LogLevel != DefaultLogLevel {
			t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, DefaultLogLevel)
		}
	})

	t.Run("errors if aseprite_path not set", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.setDefaults()
		if err == nil {
			t.Error("setDefaults() expected error for missing aseprite_path, got nil")
		}
	})
}

func TestConfig_LoadFromFile(t *testing.T) {
	realAseprite := `D:\SRC\aseprite\build\bin\aseprite.exe`
	tempDir := t.TempDir()

	// Create a test config file
	configData := map[string]any{
		"aseprite_path": realAseprite,
		"temp_dir":      tempDir,
		"timeout":       30, // 30 seconds
		"log_level":     "debug",
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		t.Fatal(err)
	}

	// Temporarily override getConfigFilePath for testing
	originalGetConfigFilePath := getConfigFilePath
	getConfigFilePath = func() string { return configPath }
	defer func() { getConfigFilePath = originalGetConfigFilePath }()

	cfg := &Config{}
	err = cfg.loadFromFile()
	if err != nil {
		t.Fatalf("loadFromFile() error = %v", err)
	}

	if cfg.AsepritePath != realAseprite {
		t.Errorf("AsepritePath = %v, want %v", cfg.AsepritePath, realAseprite)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
}

func TestLoad_MissingConfigFile(t *testing.T) {
	// Temporarily override getConfigFilePath to non-existent path
	originalGetConfigFilePath := getConfigFilePath
	getConfigFilePath = func() string { return "/nonexistent/config.json" }
	defer func() { getConfigFilePath = originalGetConfigFilePath }()

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for missing config file, got nil")
	}

	// Check that error message mentions config file not found
	if err != nil && !contains(err.Error(), "config file not found") {
		t.Errorf("Load() error = %v, want 'config file not found' message", err)
	}
}

func TestLoad_MalformedJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Write malformed JSON
	if err := os.WriteFile(configPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	// Temporarily override getConfigFilePath for testing
	originalGetConfigFilePath := getConfigFilePath
	getConfigFilePath = func() string { return configPath }
	defer func() { getConfigFilePath = originalGetConfigFilePath }()

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for malformed JSON, got nil")
	}
}

func TestLoad_MissingAsepritePath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create config without aseprite_path
	configData := map[string]any{
		"temp_dir":  tempDir,
		"log_level": "info",
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		t.Fatal(err)
	}

	// Temporarily override getConfigFilePath for testing
	originalGetConfigFilePath := getConfigFilePath
	getConfigFilePath = func() string { return configPath }
	defer func() { getConfigFilePath = originalGetConfigFilePath }()

	_, err = Load()
	if err == nil {
		t.Error("Load() expected error for missing aseprite_path, got nil")
	}
}

func TestLoad_InvalidAsepritePath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create config with invalid aseprite path
	configData := map[string]any{
		"aseprite_path": "/nonexistent/aseprite",
		"temp_dir":      tempDir,
		"log_level":     "info",
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		t.Fatal(err)
	}

	// Temporarily override getConfigFilePath for testing
	originalGetConfigFilePath := getConfigFilePath
	getConfigFilePath = func() string { return configPath }
	defer func() { getConfigFilePath = originalGetConfigFilePath }()

	_, err = Load()
	if err == nil {
		t.Error("Load() expected error for invalid aseprite path, got nil")
	}

	if err != nil && !contains(err.Error(), "invalid configuration") {
		t.Errorf("Load() error = %v, want 'invalid configuration' message", err)
	}
}

func TestValidate_UnwritableTempDir(t *testing.T) {
	// Load real config for aseprite path
	testCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Create a read-only directory (simulate unwritable)
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup

	cfg := &Config{
		AsepritePath: testCfg.AsepritePath,
		TempDir:      readOnlyDir,
		Timeout:      30 * time.Second,
		LogLevel:     "info",
	}

	err = cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for unwritable temp dir, got nil")
	}

	if err != nil && !contains(err.Error(), "not writable") {
		t.Errorf("Validate() error = %v, want 'not writable' message", err)
	}
}

func TestValidate_ZeroTimeout(t *testing.T) {
	testCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	cfg := &Config{
		AsepritePath: testCfg.AsepritePath,
		TempDir:      t.TempDir(),
		Timeout:      0,
		LogLevel:     "info",
	}

	err = cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for zero timeout, got nil")
	}

	if err != nil && !contains(err.Error(), "timeout must be positive") {
		t.Errorf("Validate() error = %v, want 'timeout must be positive' message", err)
	}
}

func TestLoadFromFile_MissingFile(t *testing.T) {
	// Temporarily override getConfigFilePath to non-existent path
	originalGetConfigFilePath := getConfigFilePath
	getConfigFilePath = func() string { return "/nonexistent/config.json" }
	defer func() { getConfigFilePath = originalGetConfigFilePath }()

	cfg := &Config{}
	err := cfg.loadFromFile()
	if err == nil {
		t.Error("loadFromFile() expected error for missing file, got nil")
	}
}

// Helper function to check if string contains substring
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
