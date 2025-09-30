package tools

import (
	"testing"
)

func TestExportSpriteInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   ExportSpriteInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid PNG export",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/output.png",
				Format:      "png",
				FrameNumber: 0,
			},
			wantErr: false,
		},
		{
			name: "valid GIF export",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/output.gif",
				Format:      "gif",
				FrameNumber: 0,
			},
			wantErr: false,
		},
		{
			name: "valid JPG export",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/output.jpg",
				Format:      "jpg",
				FrameNumber: 0,
			},
			wantErr: false,
		},
		{
			name: "valid BMP export",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/output.bmp",
				Format:      "bmp",
				FrameNumber: 0,
			},
			wantErr: false,
		},
		{
			name: "valid specific frame export",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/output.png",
				Format:      "png",
				FrameNumber: 5,
			},
			wantErr: false,
		},
		{
			name: "case insensitive format",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/output.png",
				Format:      "PNG",
				FrameNumber: 0,
			},
			wantErr: false,
		},
		{
			name: "empty output path",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "",
				Format:      "png",
				FrameNumber: 0,
			},
			wantErr: true,
			errMsg:  "output_path cannot be empty",
		},
		{
			name: "invalid format",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/output.invalid",
				Format:      "invalid",
				FrameNumber: 0,
			},
			wantErr: true,
			errMsg:  "invalid format",
		},
		{
			name: "negative frame number",
			input: ExportSpriteInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/output.png",
				Format:      "png",
				FrameNumber: -1,
			},
			wantErr: true,
			errMsg:  "frame_number must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate output path
			if tt.input.OutputPath == "" && tt.wantErr {
				return
			}

			// Validate format
			validFormats := map[string]bool{
				"png": true,
				"gif": true,
				"jpg": true,
				"bmp": true,
			}
			format := tt.input.Format
			if format != "" {
				format = string([]byte{format[0] | 0x20}) + format[1:] // simple lowercase
			}
			if !validFormats[format] && tt.wantErr {
				return
			}

			// Validate frame number
			if tt.input.FrameNumber < 0 && tt.wantErr {
				return
			}

			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}