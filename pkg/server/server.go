package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/aseprite-mcp-go/pkg/tools"
	"github.com/willibrandon/mtlog/core"
)

// Server wraps the MCP server and provides Aseprite tool implementations.
type Server struct {
	mcp    *mcp.Server
	client *aseprite.Client
	gen    *aseprite.LuaGenerator
	config *config.Config
	logger core.Logger
}

// New creates a new Aseprite MCP server.
func New(cfg *config.Config, logger core.Logger) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create Aseprite client
	client := aseprite.NewClient(cfg.AsepritePath, cfg.TempDir, cfg.Timeout)

	// Create Lua generator
	gen := aseprite.NewLuaGenerator()

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "aseprite-mcp",
		Version: "0.1.0",
	}, nil)

	s := &Server{
		mcp:    mcpServer,
		client: client,
		gen:    gen,
		config: cfg,
		logger: logger,
	}

	// Register tools (will be implemented in next tasks)
	s.registerTools()

	return s, nil
}

// Run starts the MCP server with stdio transport.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Information("Starting Aseprite MCP server")
	s.logger.Debug("Configuration: {@Config}", s.config)

	transport := &mcp.StdioTransport{}

	if err := s.mcp.Run(ctx, transport); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// registerTools registers all MCP tools with the server.
func (s *Server) registerTools() {
	s.logger.Debug("Registering MCP tools")

	// Register canvas management tools (Tasks 8-10)
	tools.RegisterCanvasTools(s.mcp, s.client, s.gen, s.config, s.logger)

	// Register drawing tools (Tasks 11-13)
	tools.RegisterDrawingTools(s.mcp, s.client, s.gen, s.config, s.logger)

	// Register export tools (Task 14)
	tools.RegisterExportTools(s.mcp, s.client, s.gen, s.config, s.logger)
}

// Client returns the underlying Aseprite client for testing.
func (s *Server) Client() *aseprite.Client {
	return s.client
}
