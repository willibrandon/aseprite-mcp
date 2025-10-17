package server

import (
	"testing"

	"github.com/willibrandon/pixel-mcp/internal/testutil"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/sinks"
)

func TestNew(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithSink(sinks.NewMemorySink()))

	server, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if server == nil {
		t.Fatal("New() returned nil server")
	}

	if server.client == nil {
		t.Error("server.client is nil")
	}

	if server.gen == nil {
		t.Error("server.gen is nil")
	}

	if server.config != cfg {
		t.Error("server.config does not match provided config")
	}

	if server.logger == nil {
		t.Error("server.logger is nil")
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	logger := mtlog.New(mtlog.WithSink(sinks.NewMemorySink()))

	tests := []struct {
		name          string
		asepritePath  string
		wantErrSubstr string
	}{
		{
			name:          "empty aseprite path",
			asepritePath:  "",
			wantErrSubstr: "aseprite executable not found",
		},
		{
			name:          "nonexistent aseprite path",
			asepritePath:  "D:\\nonexistent\\aseprite.exe",
			wantErrSubstr: "aseprite executable not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testutil.CreateTestConfigWithPath(t, tt.asepritePath)

			_, err := New(cfg, logger)
			if err == nil {
				t.Fatal("New() expected error, got nil")
			}

			if tt.wantErrSubstr != "" {
				errMsg := err.Error()
				if !contains(errMsg, tt.wantErrSubstr) {
					t.Errorf("New() error = %v, want substring %q", err, tt.wantErrSubstr)
				}
			}
		})
	}
}

func TestServer_Client(t *testing.T) {
	cfg := testutil.LoadTestConfig(t)
	logger := mtlog.New(mtlog.WithSink(sinks.NewMemorySink()))

	server, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	client := server.Client()
	if client == nil {
		t.Error("Client() returned nil")
	}
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
