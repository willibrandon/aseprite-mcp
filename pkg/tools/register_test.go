package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
)

// createTestServer creates a minimal test MCP server for registration tests
func createTestServer(t *testing.T) (*mcp.Server, *aseprite.Client, *aseprite.LuaGenerator) {
	t.Helper()

	cfg := testutil.LoadTestConfig(t)
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)
	gen := &aseprite.LuaGenerator{}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pixel-mcp-test",
		Version: "1.0.0",
	}, nil)

	return server, client, gen
}

func TestRegisterAnalysisTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterAnalysisTools(server, client, gen, cfg, logger)

	// Verify tools are registered by checking the server
	// (we can't easily enumerate tools without a session, but we can verify no panic)
	assert.NotNil(t, server)
}

func TestRegisterAnimationTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterAnimationTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterAntialiasingTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterAntialiasingTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterDitheringTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterDitheringTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterDrawingTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterDrawingTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterExportTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterExportTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterInspectionTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterInspectionTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterPaletteTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterPaletteTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterSelectionTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterSelectionTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterTransformTools(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterTransformTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

// Test with timing enabled to cover maybeWrapWithTiming branches
func TestRegisterCanvasTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterCanvasTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterDrawingTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterDrawingTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterAnalysisTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterAnalysisTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterAnimationTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterAnimationTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterAntialiasingTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterAntialiasingTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterDitheringTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterDitheringTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterExportTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterExportTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterInspectionTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterInspectionTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterPaletteTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterPaletteTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterSelectionTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterSelectionTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}

func TestRegisterTransformTools_WithTiming(t *testing.T) {
	server, client, gen := createTestServer(t)
	cfg := testutil.LoadTestConfig(t)
	cfg.EnableTiming = true
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	RegisterTransformTools(server, client, gen, cfg, logger)

	assert.NotNil(t, server)
}
