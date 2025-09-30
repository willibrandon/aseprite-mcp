// Copyright 2025 Brandon Williams. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "aseprite-mcp-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a fake aseprite executable
	fakeAseprite := filepath.Join(tempDir, "aseprite.exe")
	if err := os.WriteFile(fakeAseprite, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				AsepritePath: fakeAseprite,
				TempDir:      tempDir,
				Timeout:      30 * time.Second,
				LogLevel:     "info",
			},
			wantErr: false,
		},
		{
			name: "missing aseprite executable",
			config: &Config{
				AsepritePath: filepath.Join(tempDir, "nonexistent.exe"),
				TempDir:      tempDir,
				Timeout:      30 * time.Second,
				LogLevel:     "info",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				AsepritePath: fakeAseprite,
				TempDir:      tempDir,
				Timeout:      -1 * time.Second,
				LogLevel:     "info",
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				AsepritePath: fakeAseprite,
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

func TestConfig_LoadFromEnv(t *testing.T) {
	// Save original env vars
	origPath := os.Getenv("ASEPRITE_PATH")
	origTempDir := os.Getenv("ASEPRITE_TEMP_DIR")
	origTimeout := os.Getenv("ASEPRITE_TIMEOUT")
	origLogLevel := os.Getenv("ASEPRITE_LOG_LEVEL")

	// Restore env vars after test
	defer func() {
		os.Setenv("ASEPRITE_PATH", origPath)
		os.Setenv("ASEPRITE_TEMP_DIR", origTempDir)
		os.Setenv("ASEPRITE_TIMEOUT", origTimeout)
		os.Setenv("ASEPRITE_LOG_LEVEL", origLogLevel)
	}()

	// Set test env vars
	os.Setenv("ASEPRITE_PATH", "/test/aseprite")
	os.Setenv("ASEPRITE_TEMP_DIR", "/test/tmp")
	os.Setenv("ASEPRITE_TIMEOUT", "60")
	os.Setenv("ASEPRITE_LOG_LEVEL", "debug")

	cfg := &Config{}
	cfg.loadFromEnv()

	if cfg.AsepritePath != "/test/aseprite" {
		t.Errorf("AsepritePath = %v, want /test/aseprite", cfg.AsepritePath)
	}

	if cfg.TempDir != "/test/tmp" {
		t.Errorf("TempDir = %v, want /test/tmp", cfg.TempDir)
	}

	if cfg.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", cfg.Timeout)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
}

func TestGetCommonAsepritePaths(t *testing.T) {
	paths := getCommonAsepritePaths()

	if len(paths) == 0 {
		t.Error("getCommonAsepritePaths() returned no paths")
	}

	// Just verify we get some platform-specific paths
	// Don't check exact values as they vary by OS
	t.Logf("Common paths: %v", paths)
}