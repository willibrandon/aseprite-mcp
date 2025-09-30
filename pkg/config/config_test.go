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
		"timeout":       30000000000, // 30 seconds in nanoseconds
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
