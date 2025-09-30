package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/willibrandon/aseprite-mcp-go/pkg/aseprite"
)

type ExportSpriteInput struct {
	SpritePath  string `json:"sprite_path" jsonschema:"required,description=Path to the Aseprite file"`
	OutputPath  string `json:"output_path" jsonschema:"required,description=Output file path"`
	FrameNumber int    `json:"frame_number" jsonschema:"minimum=0,description=Specific frame to export (0 = all frames)"`
}

type ExportSpriteOutput struct {
	ExportedPath string `json:"exported_path" jsonschema:"description=Path to the exported file"`
}

func RegisterExportTools(server *mcp.Server, client *aseprite.Client, gen *aseprite.LuaGenerator) {
	// export_sprite tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_sprite",
		Description: "Export sprite to common image formats (PNG, GIF, etc)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ExportSpriteInput) (*mcp.CallToolResult, *ExportSpriteOutput, error) {
		// Generate Lua script
		script := gen.ExportSprite(input.OutputPath, input.FrameNumber)

		// Execute
		_, err := client.ExecuteLua(ctx, script, input.SpritePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to export sprite: %w", err)
		}

		msg := fmt.Sprintf("Exported sprite to: %s", input.OutputPath)
		if input.FrameNumber > 0 {
			msg = fmt.Sprintf("Exported frame %d to: %s", input.FrameNumber, input.OutputPath)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: msg},
			},
		}, &ExportSpriteOutput{ExportedPath: input.OutputPath}, nil
	})
}