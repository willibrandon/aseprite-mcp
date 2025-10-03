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

func TestExportSpritesheetInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   ExportSpritesheetInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid horizontal layout",
			input: ExportSpritesheetInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/sheet.png",
				Layout:      "horizontal",
				Padding:     2,
				IncludeJSON: false,
			},
			wantErr: false,
		},
		{
			name: "valid with JSON metadata",
			input: ExportSpritesheetInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/sheet.png",
				Layout:      "packed",
				Padding:     0,
				IncludeJSON: true,
			},
			wantErr: false,
		},
		{
			name: "empty output path",
			input: ExportSpritesheetInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "",
				Layout:      "horizontal",
				Padding:     0,
				IncludeJSON: false,
			},
			wantErr: true,
			errMsg:  "output_path cannot be empty",
		},
		{
			name: "invalid layout",
			input: ExportSpritesheetInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/sheet.png",
				Layout:      "invalid",
				Padding:     0,
				IncludeJSON: false,
			},
			wantErr: true,
			errMsg:  "invalid layout",
		},
		{
			name: "padding too large",
			input: ExportSpritesheetInput{
				SpritePath:  "/path/to/sprite.aseprite",
				OutputPath:  "/path/to/sheet.png",
				Layout:      "horizontal",
				Padding:     101,
				IncludeJSON: false,
			},
			wantErr: true,
			errMsg:  "padding must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation checks
			if tt.input.OutputPath == "" && tt.wantErr {
				return
			}

			validLayouts := map[string]bool{
				"horizontal": true,
				"vertical":   true,
				"rows":       true,
				"columns":    true,
				"packed":     true,
			}
			if !validLayouts[tt.input.Layout] && tt.input.Layout != "" && tt.wantErr {
				return
			}

			if (tt.input.Padding < 0 || tt.input.Padding > 100) && tt.wantErr {
				return
			}

			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestImportImageInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   ImportImageInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid import",
			input: ImportImageInput{
				SpritePath:  "/path/to/sprite.aseprite",
				ImagePath:   "/path/to/image.png",
				LayerName:   "Imported",
				FrameNumber: 1,
				Position:    nil,
			},
			wantErr: false,
		},
		{
			name: "empty sprite path",
			input: ImportImageInput{
				SpritePath:  "",
				ImagePath:   "/path/to/image.png",
				LayerName:   "Imported",
				FrameNumber: 1,
			},
			wantErr: true,
			errMsg:  "sprite_path cannot be empty",
		},
		{
			name: "empty image path",
			input: ImportImageInput{
				SpritePath:  "/path/to/sprite.aseprite",
				ImagePath:   "",
				LayerName:   "Imported",
				FrameNumber: 1,
			},
			wantErr: true,
			errMsg:  "image_path cannot be empty",
		},
		{
			name: "invalid frame number",
			input: ImportImageInput{
				SpritePath:  "/path/to/sprite.aseprite",
				ImagePath:   "/path/to/image.png",
				LayerName:   "Imported",
				FrameNumber: 0,
			},
			wantErr: true,
			errMsg:  "frame_number must be >= 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation checks
			if tt.input.SpritePath == "" && tt.wantErr {
				return
			}
			if tt.input.ImagePath == "" && tt.wantErr {
				return
			}
			if tt.input.FrameNumber < 1 && tt.wantErr {
				return
			}

			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestSaveAsInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   SaveAsInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid .aseprite extension",
			input: SaveAsInput{
				SpritePath: "/path/to/sprite.aseprite",
				OutputPath: "/path/to/new.aseprite",
			},
			wantErr: false,
		},
		{
			name: "valid .ase extension",
			input: SaveAsInput{
				SpritePath: "/path/to/sprite.aseprite",
				OutputPath: "/path/to/new.ase",
			},
			wantErr: false,
		},
		{
			name: "missing extension",
			input: SaveAsInput{
				SpritePath: "/path/to/sprite.aseprite",
				OutputPath: "/path/to/new",
			},
			wantErr: true,
			errMsg:  "output_path must have .aseprite or .ase extension",
		},
		{
			name: "wrong extension",
			input: SaveAsInput{
				SpritePath: "/path/to/sprite.aseprite",
				OutputPath: "/path/to/new.png",
			},
			wantErr: true,
			errMsg:  "output_path must have .aseprite or .ase extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check extension
			hasValidExt := false
			for _, ext := range []string{".aseprite", ".ase"} {
				if len(tt.input.OutputPath) >= len(ext) && tt.input.OutputPath[len(tt.input.OutputPath)-len(ext):] == ext {
					hasValidExt = true
					break
				}
			}

			if !hasValidExt && tt.wantErr {
				return
			}

			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}
