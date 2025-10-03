//go:build integration
// +build integration

package tools

import (
	"context"
	"testing"

	"github.com/willibrandon/aseprite-mcp-go/internal/testutil"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

func TestIntegration_DeleteLayer(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()

	// Create test sprite with multiple layers
	spritePath := testutil.TempSpritePath(t, "delete-layer-test.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(context.Background(), createScript, "")
	if err != nil {
		t.Fatalf("Failed to create sprite: %v", err)
	}

	// Add a second layer
	addLayerScript := gen.AddLayer("Layer 2")
	_, err = client.ExecuteLua(context.Background(), addLayerScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add layer: %v", err)
	}

	// Delete the second layer
	deleteScript := gen.DeleteLayer("Layer 2")
	output, err := client.ExecuteLua(context.Background(), deleteScript, spritePath)
	if err != nil {
		t.Fatalf("DeleteLayer failed: %v", err)
	}

	if output != "Layer deleted successfully\n" {
		t.Errorf("Unexpected output: %q", output)
	}

	// Verify layer was deleted
	infoScript := gen.GetSpriteInfo()
	output, err = client.ExecuteLua(context.Background(), infoScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to get sprite info: %v", err)
	}

	// Should only have 1 layer now
	if !contains(output, `"layer_count": 1`) {
		t.Errorf("Expected 1 layer, got output: %s", output)
	}

	t.Logf("✓ Layer deleted successfully")
}

func TestIntegration_DeleteLayer_LastLayer(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()

	// Create test sprite with one layer
	spritePath := testutil.TempSpritePath(t, "delete-last-layer-test.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(context.Background(), createScript, "")
	if err != nil {
		t.Fatalf("Failed to create sprite: %v", err)
	}

	// Try to delete the only layer (should fail)
	deleteScript := gen.DeleteLayer("Layer 1")
	_, err = client.ExecuteLua(context.Background(), deleteScript, spritePath)
	if err == nil {
		t.Error("Expected error when deleting last layer, got nil")
	}

	if !contains(err.Error(), "Cannot delete the last layer") {
		t.Errorf("Expected 'Cannot delete the last layer' error, got: %v", err)
	}

	t.Logf("✓ Correctly rejected deletion of last layer")
}

func TestIntegration_DeleteFrame(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()

	// Create test sprite with multiple frames
	spritePath := testutil.TempSpritePath(t, "delete-frame-test.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(context.Background(), createScript, "")
	if err != nil {
		t.Fatalf("Failed to create sprite: %v", err)
	}

	// Add a second frame
	addFrameScript := gen.AddFrame(100)
	_, err = client.ExecuteLua(context.Background(), addFrameScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to add frame: %v", err)
	}

	// Delete the second frame
	deleteScript := gen.DeleteFrame(2)
	output, err := client.ExecuteLua(context.Background(), deleteScript, spritePath)
	if err != nil {
		t.Fatalf("DeleteFrame failed: %v", err)
	}

	if output != "Frame deleted successfully\n" {
		t.Errorf("Unexpected output: %q", output)
	}

	// Verify frame was deleted
	infoScript := gen.GetSpriteInfo()
	output, err = client.ExecuteLua(context.Background(), infoScript, spritePath)
	if err != nil {
		t.Fatalf("Failed to get sprite info: %v", err)
	}

	// Should only have 1 frame now
	if !contains(output, `"frame_count": 1`) {
		t.Errorf("Expected 1 frame, got output: %s", output)
	}

	t.Logf("✓ Frame deleted successfully")
}

func TestIntegration_DeleteFrame_LastFrame(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := aseprite.NewLuaGenerator()

	// Create test sprite with one frame
	spritePath := testutil.TempSpritePath(t, "delete-last-frame-test.aseprite")
	createScript := gen.CreateCanvas(64, 64, aseprite.ColorModeRGB, spritePath)
	_, err := client.ExecuteLua(context.Background(), createScript, "")
	if err != nil {
		t.Fatalf("Failed to create sprite: %v", err)
	}

	// Try to delete the only frame (should fail)
	deleteScript := gen.DeleteFrame(1)
	_, err = client.ExecuteLua(context.Background(), deleteScript, spritePath)
	if err == nil {
		t.Error("Expected error when deleting last frame, got nil")
	}

	if !contains(err.Error(), "Cannot delete the last frame") {
		t.Errorf("Expected 'Cannot delete the last frame' error, got: %v", err)
	}

	t.Logf("✓ Correctly rejected deletion of last frame")
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
