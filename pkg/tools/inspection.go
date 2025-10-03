package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
	"github.com/willibrandon/aseprite-mcp-go/pkg/config"
	"github.com/willibrandon/mtlog/core"
)

// GetPixelsInput defines the input parameters for the get_pixels tool.
type GetPixelsInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"Path to the Aseprite sprite file"`
	LayerName   string `json:"layer_name" jsonschema:"Name of the layer to read from"`
	FrameNumber int    `json:"frame_number" jsonschema:"Frame number to read from (1-based)"`
	X           int    `json:"x" jsonschema:"X coordinate of top-left corner of region"`
	Y           int    `json:"y" jsonschema:"Y coordinate of top-left corner of region"`
	Width       int    `json:"width" jsonschema:"Width of region to read"`
	Height      int    `json:"height" jsonschema:"Height of region to read"`
	Cursor      string `json:"cursor,omitempty" jsonschema:"Pagination cursor for fetching next page (optional)"`
	PageSize    int    `json:"page_size,omitempty" jsonschema:"Number of pixels to return per page (default: 1000, max: 10000)"`
}

// PixelData represents a single pixel with coordinates and color.
type PixelData struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

// GetPixelsOutput defines the output for the get_pixels tool.
type GetPixelsOutput struct {
	Pixels      []PixelData `json:"pixels" jsonschema:"Array of pixels with coordinates and colors"`
	NextCursor  string      `json:"next_cursor,omitempty" jsonschema:"Cursor for fetching next page (empty if no more pages)"`
	TotalPixels int         `json:"total_pixels" jsonschema:"Total number of pixels in the region"`
}

// RegisterInspectionTools registers all inspection tools with the MCP server.
func RegisterInspectionTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator, cfg *config.Config, logger core.Logger) {
	// Register get_pixels tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_pixels",
			Description: "Read pixel data from a rectangular region of a sprite. Returns an array of pixels with their coordinates and colors in hex format (#RRGGBBAA). Supports pagination for large regions using cursor and page_size parameters (default page size: 1000, max: 10000).",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input GetPixelsInput) (*mcp.CallToolResult, *GetPixelsOutput, error) {
			logger.Debug("get_pixels tool called", "sprite_path", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "x", input.X, "y", input.Y, "width", input.Width, "height", input.Height, "cursor", input.Cursor, "page_size", input.PageSize)

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
				logger.Error("Failed to get pixels", "error", err)
				return nil, nil, fmt.Errorf("failed to get pixels: %w", err)
			}

			// Parse JSON output
			var pixels []PixelData
			if err := json.Unmarshal([]byte(output), &pixels); err != nil {
				logger.Error("Failed to parse pixel data", "error", err, "output", output)
				return nil, nil, fmt.Errorf("failed to parse pixel data: %w", err)
			}

			// Generate next cursor if there are more pixels
			nextCursor := ""
			nextOffset := offset + len(pixels)
			if nextOffset < totalPixelCount {
				nextCursor = strconv.Itoa(nextOffset)
			}

			logger.Information("Read pixels successfully", "sprite", input.SpritePath, "layer", input.LayerName, "frame", input.FrameNumber, "total", totalPixelCount, "returned", len(pixels), "offset", offset)

			return nil, &GetPixelsOutput{
				Pixels:      pixels,
				NextCursor:  nextCursor,
				TotalPixels: totalPixelCount,
			}, nil
		},
	)
}
