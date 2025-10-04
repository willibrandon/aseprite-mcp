package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
)

// Test input/output types for handler testing
type testInput struct {
	Value string
}

type testOutput struct {
	Result string
}

func TestWrapWithTiming_Success(t *testing.T) {
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	// Create a simple handler that succeeds
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input testInput) (*mcp.CallToolResult, *testOutput, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "success",
				},
			},
		}, &testOutput{Result: input.Value + "_processed"}, nil
	}

	// Wrap it with timing
	wrapped := wrapWithTiming("test_tool", logger, handler)

	// Execute wrapped handler
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := testInput{Value: "test"}

	result, output, err := wrapped(ctx, req, input)

	if err != nil {
		t.Errorf("wrapped handler returned error: %v", err)
	}

	if result == nil {
		t.Error("wrapped handler returned nil result")
	}

	if output == nil || output.Result != "test_processed" {
		t.Errorf("wrapped handler returned wrong output: %+v", output)
	}
}

func TestWrapWithTiming_Error(t *testing.T) {
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	// Create a handler that fails
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input testInput) (*mcp.CallToolResult, *testOutput, error) {
		return nil, nil, errors.New("test error")
	}

	// Wrap it with timing
	wrapped := wrapWithTiming("test_tool", logger, handler)

	// Execute wrapped handler
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := testInput{Value: "test"}

	result, output, err := wrapped(ctx, req, input)

	if err == nil {
		t.Error("wrapped handler should return error")
	}

	if err.Error() != "test error" {
		t.Errorf("wrapped handler returned wrong error: %v", err)
	}

	if result != nil {
		t.Error("wrapped handler should return nil result on error")
	}

	if output != nil {
		t.Error("wrapped handler should return nil output on error")
	}
}

func TestMaybeWrapWithTiming_Enabled(t *testing.T) {
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	// Create a simple handler
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input testInput) (*mcp.CallToolResult, *testOutput, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "success",
				},
			},
		}, &testOutput{Result: "wrapped"}, nil
	}

	// Wrap with timing enabled
	wrapped := maybeWrapWithTiming("test_tool", logger, true, handler)

	// Execute - should be wrapped with timing
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := testInput{Value: "test"}

	result, output, err := wrapped(ctx, req, input)

	if err != nil {
		t.Errorf("wrapped handler returned error: %v", err)
	}

	if result == nil {
		t.Error("wrapped handler returned nil result")
	}

	if output == nil || output.Result != "wrapped" {
		t.Errorf("wrapped handler returned wrong output: %+v", output)
	}
}

func TestMaybeWrapWithTiming_Disabled(t *testing.T) {
	logger := mtlog.New(mtlog.WithMinimumLevel(core.ErrorLevel))

	callCount := 0
	// Create a simple handler
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input testInput) (*mcp.CallToolResult, *testOutput, error) {
		callCount++
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "success",
				},
			},
		}, &testOutput{Result: "direct"}, nil
	}

	// Wrap with timing disabled
	wrapped := maybeWrapWithTiming("test_tool", logger, false, handler)

	// Execute - should be direct call without timing wrapper
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := testInput{Value: "test"}

	result, output, err := wrapped(ctx, req, input)

	if err != nil {
		t.Errorf("handler returned error: %v", err)
	}

	if result == nil {
		t.Error("handler returned nil result")
	}

	if output == nil || output.Result != "direct" {
		t.Errorf("handler returned wrong output: %+v", output)
	}

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}
}
