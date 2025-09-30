//go:build integration
// +build integration

package aseprite

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

// These tests require a real Aseprite installation.
// Run with: go test -tags=integration ./pkg/aseprite

func TestIntegration_GetVersion(t *testing.T) {
	asepritePath, err := exec.LookPath("aseprite")
	if err != nil {
		t.Skip("Aseprite not found in PATH")
	}

	tempDir := t.TempDir()
	client := NewClient(asepritePath, tempDir, 30*time.Second)

	ctx := context.Background()
	version, err := client.GetVersion(ctx)
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}

	if version == "" {
		t.Error("GetVersion() returned empty version")
	}

	t.Logf("Aseprite version: %s", version)
}

func TestIntegration_CreateCanvas(t *testing.T) {
	asepritePath, err := exec.LookPath("aseprite")
	if err != nil {
		t.Skip("Aseprite not found in PATH")
	}

	tempDir := t.TempDir()
	client := NewClient(asepritePath, tempDir, 30*time.Second)
	gen := NewLuaGenerator()

	script := gen.CreateCanvas(100, 100, ColorModeRGB)

	ctx := context.Background()
	output, err := client.ExecuteLua(ctx, script, "")
	if err != nil {
		t.Fatalf("ExecuteLua() error = %v", err)
	}

	t.Logf("Created sprite at: %s", output)
}