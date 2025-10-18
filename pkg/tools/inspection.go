package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/mtlog/core"
	"github.com/willibrandon/pixel-mcp/pkg/aseprite"
	"github.com/willibrandon/pixel-mcp/pkg/config"
)

// GetPixelsInput defines the input parameters for the get_pixels tool.
//
// Reads pixel data from a rectangular region. Supports pagination for large regions
// to avoid memory/bandwidth issues. Default page size is 1000 pixels, max 10000.
type GetPixelsInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`                                        // Path to the sprite file to read
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to read from"`                                           // Layer to read from
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to read from (1-based)"`                                    // 1-based frame index
	X           int    `json:"x" jsonschema:"X coordinate of top-left corner of region"`                                         // Top-left X coordinate
	Y           int    `json:"y" jsonschema:"Y coordinate of top-left corner of region"`                                         // Top-left Y coordinate
	Width       int    `json:"width" jsonschema:"Width of region to read"`                                                       // Region width in pixels
	Height      int    `json:"height" jsonschema:"Height of region to read"`                                                     // Region height in pixels
	Cursor      string `json:"cursor,omitempty" jsonschema:"Pagination cursor for fetching next page (optional)"`                // Cursor for next page (empty for first page)
	PageSize    int    `json:"page_size,omitempty" jsonschema:"Number of pixels to return per page (default: 1000, max: 10000)"` // Page size (default 1000, max 10000)
}

// PixelData represents a single pixel with coordinates and color.
//
// Used by get_pixels to return pixel information from sprite inspection.
type PixelData struct {
	X     int    `json:"x"`     // X coordinate in sprite space
	Y     int    `json:"y"`     // Y coordinate in sprite space
	Color string `json:"color"` // Color in hex format (#RRGGBBAA)
}

// GetPixelsOutput defines the output for the get_pixels tool.
//
// Returns pixel data with pagination support. NextCursor is empty when no more pages exist.
type GetPixelsOutput struct {
	Pixels      []PixelData `json:"pixels" jsonschema:"Array of pixels with coordinates and colors"`                           // Pixel data for current page
	NextCursor  string      `json:"next_cursor,omitempty" jsonschema:"Cursor for fetching next page (empty if no more pages)"` // Pagination cursor (empty if done)
	TotalPixels int         `json:"total_pixels" jsonschema:"Total number of pixels in the region"`                            // Total pixels in queried region
}

// RegisterInspectionTools registers all inspection tools with the MCP server.
//
// Registers the following tools:
//   - get_pixels: Read pixel data from rectangular regions with pagination support
//
// Inspection tools are read-only and do not modify sprite files.
func RegisterInspectionTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register get_pixels tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_pixels",
			Description: "Read pixel data from a rectangular region of a sprite. Returns an array of pixels with their coordinates and colors in hex format (#RRGGBBAA). Supports pagination for large regions using cursor and page_size parameters (default page size: 1000, max: 10000).",
		},
		maybeWrapWithTiming("get_pixels", logger, cfg.EnableTiming, func(ctx context.Context, req *mcp.CallToolRequest, input GetPixelsInput) (*mcp.CallToolResult, *GetPixelsOutput, error) {
			opLogger := logger.WithContext(ctx)
			opLogger.Debug("get_pixels tool called", "sprite_path", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "x", input.X, "y", input.Y, "width", input.Width, "height", input.Height, "cursor", input.Cursor, "page_size", input.PageSize)

			// Validate inputs
			if input.Width <= 0 || input.Height <= 0 {
				return nil, nil, fmt.Errorf("width and height must be positive, got width=%d height=%d", input.Width, input.Height)
			}

			if input.FrameNumber < 1 {
				return nil, nil, fmt.Errorf("frame_number must be >= 1, got %d", input.FrameNumber)
			}

			// Set default page size and validate
			pageSize := input.PageSize
			if pageSize <= 0 {
				pageSize = 1000 // default
			}
			if pageSize > 10000 {
				pageSize = 10000 // max
			}

			// Parse cursor to get offset
			offset := 0
			if input.Cursor != "" {
				var err error
				offset, err = strconv.Atoi(input.Cursor)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid cursor: %w", err)
				}
			}

			// Calculate total pixel count
			totalPixelCount := input.Width * input.Height

			// Generate Lua script with pagination
			script := gen.GetPixelsWithPagination(input.LayerName, input.FrameNumber, input.X, input.Y, input.Width, input.Height, offset, pageSize)

			// Execute Lua script with the sprite
			output, err := client.ExecuteLua(ctx, script, input.SpritePath)
			if err != nil {
				opLogger.Error("Failed to get pixels", "error", err)
				return nil, nil, fmt.Errorf("failed to get pixels: %w", err)
			}

			// Parse JSON output
			var pixels []PixelData
			if err := json.Unmarshal([]byte(output), &pixels); err != nil {
				opLogger.Error("Failed to parse pixel data", "error", err, "output", output)
				return nil, nil, fmt.Errorf("failed to parse pixel data: %w", err)
			}

			// Generate next cursor if there are more pixels
			nextCursor := ""
			nextOffset := offset + len(pixels)
			if nextOffset < totalPixelCount {
				nextCursor = strconv.Itoa(nextOffset)
			}

			opLogger.Information("Read pixels successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "total", totalPixelCount, "returned", len(pixels), "offset", offset)

			return nil, &GetPixelsOutput{
				Pixels:      pixels,
				NextCursor:  nextCursor,
				TotalPixels: totalPixelCount,
			}, nil
		}),
	)
}
