// Package tools implements MCP tool handlers for Aseprite sprite manipulation.
//
// This file contains shared utilities used across all tool categories.
package tools

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
)

// wrapWithTiming wraps a tool handler with timing and request tracking.
//
// Adds request ID, operation timing, and enhanced logging to tool handlers.
// The wrapper:
//   - Generates a unique request ID for tracking
//   - Pushes request context to the context for logging
//   - Times the operation and logs duration
//   - Logs deadline warnings if approaching timeout
func wrapWithTiming[I any, O any](
	toolName string,
	logger core.Logger,
	handler func(context.Context, *mcp.CallToolRequest, I) (*mcp.CallToolResult, O, error),
) func(context.Context, *mcp.CallToolRequest, I) (*mcp.CallToolResult, O, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input I) (*mcp.CallToolResult, O, error) {
		// Generate request ID for tracking
		requestID := uuid.New().String()[:8] // Short ID for logs

		// Add request context for all logs within this operation
		ctx = mtlog.PushProperty(ctx, "RequestID", requestID)
		ctx = mtlog.PushProperty(ctx, "Tool", toolName)
		opLogger := logger.WithContext(ctx)

		// Log operation start with timing
		start := time.Now()
		opLogger.InfoContext(ctx, "Tool operation started")

		// Execute the actual handler
		result, output, err := handler(ctx, req, input)

		// Calculate duration
		duration := time.Since(start)

		// Log completion with timing
		if err != nil {
			opLogger.ErrorContext(ctx, "Tool operation failed after {Duration}", duration, "error", err)
		} else {
			opLogger.InfoContext(ctx, "Tool operation completed in {Duration}", duration)
		}

		return result, output, err
	}
}

// maybeWrapWithTiming conditionally wraps a handler with timing based on config.
//
// If enable is true, returns wrapWithTiming, otherwise returns the handler unchanged.
// This allows timing to be opt-in via configuration.
func maybeWrapWithTiming[I any, O any](
	toolName string,
	logger core.Logger,
	enable bool,
	handler func(context.Context, *mcp.CallToolRequest, I) (*mcp.CallToolResult, O, error),
) func(context.Context, *mcp.CallToolRequest, I) (*mcp.CallToolResult, O, error) {
	if enable {
		return wrapWithTiming(toolName, logger, handler)
	}
	return handler
}
