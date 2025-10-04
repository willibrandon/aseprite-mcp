package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/aseprite-mcp-go/pkg/server"
	"github.com/willibrandon/mtlog"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/mtlog/sinks"
)

var (
	// Version information (set via ldflags during build)
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Parse command-line flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHealth  = flag.Bool("health", false, "Check health and exit")
		debugMode   = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("aseprite-mcp version %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nPlease create a config file at ~/.config/aseprite-mcp/config.json:\n")
		fmt.Fprintf(os.Stderr, "{\n")
		fmt.Fprintf(os.Stderr, "  \"aseprite_path\": \"/path/to/aseprite\",\n")
		fmt.Fprintf(os.Stderr, "  \"temp_dir\": \"/tmp/aseprite-mcp\",\n")
		fmt.Fprintf(os.Stderr, "  \"timeout\": 30,\n")
		fmt.Fprintf(os.Stderr, "  \"log_level\": \"info\"\n")
		fmt.Fprintf(os.Stderr, "}\n")
		os.Exit(1)
	}

	// Override log level if debug mode enabled
	if *debugMode {
		cfg.LogLevel = "debug"
	}

	// Initialize logger
	logger := createLogger(cfg.LogLevel, cfg.LogFile)
	defer closeLogger(logger) // Ensure logs are flushed on shutdown

	// Health check mode
	if *showHealth {
		exitCode := performHealthCheck(cfg, logger)
		closeLogger(logger) // Flush logs before exit
		os.Exit(exitCode)
	}

	// Log startup
	logger.Information("Starting Aseprite MCP Server version {Version} (built {BuildTime})", Version, BuildTime)
	logger.Debug("Configuration loaded: {@Config}", cfg)

	// Create server
	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create server: {Error}", err)
		closeLogger(logger) // Flush logs before exit
		os.Exit(1)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Run(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		logger.Information("Received shutdown signal: {Signal}", sig)
		cancel()
		// Give server time to shutdown gracefully
		time.Sleep(100 * time.Millisecond)
	case err := <-errChan:
		if err != nil {
			logger.Error("Server error: {Error}", err)
			closeLogger(logger) // Flush logs before exit
			os.Exit(1)
		}
	}

	logger.Information("Server stopped")
}

// closeLogger closes the logger and flushes any buffered logs.
func closeLogger(logger core.Logger) {
	// Type assert to access Close() method on concrete logger type
	if closer, ok := logger.(interface{ Close() error }); ok {
		_ = closer.Close() // Ignore close errors
	}
}

// createLogger creates a configured logger instance.
func createLogger(logLevel string, logFilePath string) core.Logger {
	// Create console sink with custom template and Literate theme
	// Template: [HH:mm:ss LEVEL] Context: Message (Properties when in debug mode)
	// Output to stderr since stdout is used for MCP communication
	template := "[${Timestamp:HH:mm:ss} ${Level:u3}] {SourceContext}: ${Message}"
	if logLevel == "debug" {
		// In debug mode, show enriched properties like machine name, process
		template = "[${Timestamp:HH:mm:ss} ${Level:u3}] {SourceContext} [{MachineName}/{ProcessName}:{ProcessId}]: ${Message}"
	}
	consoleSink, err := sinks.NewConsoleSinkWithTemplateAndTheme(
		template,
		sinks.LiterateTheme(),
	)
	if err != nil {
		// Fallback to simple console if template fails
		consoleSink = sinks.NewConsoleSink()
	}
	// Redirect output to stderr (stdout is used for MCP communication)
	consoleSink.SetOutput(os.Stderr)

	// Create logger with options based on log level
	var opts []mtlog.Option
	opts = append(opts, mtlog.WithSink(consoleSink))

	// Add file logging if path provided
	if logFilePath != "" {
		// Create rolling file sink for production use
		// Rolls daily or at 100MB, keeps 7 days of logs
		// BufferSize: 1 for near-immediate writes (values <=0 default to 64KB)
		fileSink, err := sinks.NewRollingFileSink(sinks.RollingFileOptions{
			FilePath:        logFilePath,
			MaxFileSize:     100 * 1024 * 1024, // 100MB
			RollingInterval: sinks.RollingIntervalDaily,
			RetainFileCount: 7, // Keep last 7 days
			BufferSize:      1, // Minimum buffer for near-immediate writes
		})
		if err != nil {
			// Log error but continue with console only
			fmt.Fprintf(os.Stderr, "Warning: Failed to create log file sink: %v\n", err)
		} else {
			opts = append(opts, mtlog.WithSink(fileSink))
		}
	}

	switch logLevel {
	case "debug":
		opts = append(opts, mtlog.WithMinimumLevel(core.DebugLevel))
	case "info":
		opts = append(opts, mtlog.WithMinimumLevel(core.InformationLevel))
	case "warn":
		opts = append(opts, mtlog.WithMinimumLevel(core.WarningLevel))
	case "error":
		opts = append(opts, mtlog.WithMinimumLevel(core.ErrorLevel))
	default:
		opts = append(opts, mtlog.WithMinimumLevel(core.InformationLevel))
	}

	// Add enrichers for better context and debugging
	opts = append(opts, mtlog.WithSourceContext("AsepriteMCP"))
	opts = append(opts, mtlog.WithMachineName()) // Add machine name
	opts = append(opts, mtlog.WithProcess())     // Add process ID and executable
	opts = append(opts, mtlog.WithTimestamp())   // Ensure timestamps are included

	// Add deadline awareness - warn when operations approach timeout
	// Since most operations have 30s timeout, warn at 5s remaining
	opts = append(opts, mtlog.WithContextDeadlineWarning(5*time.Second))

	// Create logger
	logger := mtlog.New(opts...)

	return logger
}

// performHealthCheck checks if Aseprite is accessible and returns exit code.
func performHealthCheck(cfg *config.Config, logger core.Logger) int {
	logger.Information("Performing health check...")

	// Check if Aseprite executable exists
	if _, err := os.Stat(cfg.AsepritePath); os.IsNotExist(err) {
		logger.Error("Health check failed: Aseprite executable not found at {Path}", cfg.AsepritePath)
		return 1
	}

	logger.Information("✓ Aseprite executable found at {Path}", cfg.AsepritePath)

	// Try to get Aseprite version
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("Health check failed: Could not create server - {Error}", err)
		return 1
	}

	// Get version through client
	version, err := srv.Client().GetVersion(ctx)
	if err != nil {
		logger.Error("Health check failed: Could not get Aseprite version - {Error}", err)
		return 1
	}

	logger.Information("✓ Aseprite is accessible (version {Version})", version)

	// Check temp directory
	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		logger.Error("Health check failed: Could not access temp directory {Path} - {Error}", cfg.TempDir, err)
		return 1
	}

	logger.Information("✓ Temp directory is accessible at {Path}", cfg.TempDir)

	logger.Information("Health check passed - all systems operational")
	return 0
}
