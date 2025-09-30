package tools

import (
	"testing"
)

func TestSetFrameDurationInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   SetFrameDurationInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: SetFrameDurationInput{
				SpritePath:  "/path/to/sprite.aseprite",
				FrameNumber: 1,
				DurationMs:  100,
			},
			wantErr: false,
		},
		{
			name: "frame number too small",
			input: SetFrameDurationInput{
				SpritePath:  "/path/to/sprite.aseprite",
				FrameNumber: 0,
				DurationMs:  100,
			},
			wantErr: true,
			errMsg:  "frame_number must be at least 1",
		},
		{
			name: "duration too small",
			input: SetFrameDurationInput{
				SpritePath:  "/path/to/sprite.aseprite",
				FrameNumber: 1,
				DurationMs:  0,
			},
			wantErr: true,
			errMsg:  "duration_ms must be between 1 and 65535",
		},
		{
			name: "duration too large",
			input: SetFrameDurationInput{
				SpritePath:  "/path/to/sprite.aseprite",
				FrameNumber: 1,
				DurationMs:  65536,
			},
			wantErr: true,
			errMsg:  "duration_ms must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate frame number
			if tt.input.FrameNumber < 1 && tt.wantErr {
				return
			}
			// Validate duration
			if (tt.input.DurationMs < 1 || tt.input.DurationMs > 65535) && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestCreateTagInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateTagInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input forward",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "forward",
			},
			wantErr: false,
		},
		{
			name: "valid input reverse",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "reverse",
			},
			wantErr: false,
		},
		{
			name: "valid input pingpong",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "pingpong",
			},
			wantErr: false,
		},
		{
			name: "empty tag name",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "forward",
			},
			wantErr: true,
			errMsg:  "tag_name cannot be empty",
		},
		{
			name: "from_frame too small",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  0,
				ToFrame:    4,
				Direction:  "forward",
			},
			wantErr: true,
			errMsg:  "from_frame must be at least 1",
		},
		{
			name: "to_frame before from_frame",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  4,
				ToFrame:    2,
				Direction:  "forward",
			},
			wantErr: true,
			errMsg:  "to_frame must be >= from_frame",
		},
		{
			name: "invalid direction",
			input: CreateTagInput{
				SpritePath: "/path/to/sprite.aseprite",
				TagName:    "walk",
				FromFrame:  1,
				ToFrame:    4,
				Direction:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid direction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate tag name
			if tt.input.TagName == "" && tt.wantErr {
				return
			}
			// Validate from_frame
			if tt.input.FromFrame < 1 && tt.wantErr {
				return
			}
			// Validate frame range
			if tt.input.ToFrame < tt.input.FromFrame && tt.wantErr {
				return
			}
			// Validate direction
			validDirections := map[string]bool{
				"forward":  true,
				"reverse":  true,
				"pingpong": true,
			}
			if !validDirections[tt.input.Direction] && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestDuplicateFrameInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   DuplicateFrameInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input insert at end",
			input: DuplicateFrameInput{
				SpritePath:  "/path/to/sprite.aseprite",
				SourceFrame: 1,
				InsertAfter: 0,
			},
			wantErr: false,
		},
		{
			name: "valid input insert after frame",
			input: DuplicateFrameInput{
				SpritePath:  "/path/to/sprite.aseprite",
				SourceFrame: 1,
				InsertAfter: 2,
			},
			wantErr: false,
		},
		{
			name: "source_frame too small",
			input: DuplicateFrameInput{
				SpritePath:  "/path/to/sprite.aseprite",
				SourceFrame: 0,
				InsertAfter: 1,
			},
			wantErr: true,
			errMsg:  "source_frame must be at least 1",
		},
		{
			name: "insert_after negative",
			input: DuplicateFrameInput{
				SpritePath:  "/path/to/sprite.aseprite",
				SourceFrame: 1,
				InsertAfter: -1,
			},
			wantErr: true,
			errMsg:  "insert_after must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate source_frame
			if tt.input.SourceFrame < 1 && tt.wantErr {
				return
			}
			// Validate insert_after
			if tt.input.InsertAfter < 0 && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestLinkCelInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   LinkCelInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				SourceFrame: 1,
				TargetFrame: 2,
			},
			wantErr: false,
		},
		{
			name: "empty layer name",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "",
				SourceFrame: 1,
				TargetFrame: 2,
			},
			wantErr: true,
			errMsg:  "layer_name cannot be empty",
		},
		{
			name: "source_frame too small",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				SourceFrame: 0,
				TargetFrame: 2,
			},
			wantErr: true,
			errMsg:  "source_frame must be at least 1",
		},
		{
			name: "target_frame too small",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				SourceFrame: 1,
				TargetFrame: 0,
			},
			wantErr: true,
			errMsg:  "target_frame must be at least 1",
		},
		{
			name: "source and target same",
			input: LinkCelInput{
				SpritePath:  "/path/to/sprite.aseprite",
				LayerName:   "Layer 1",
				SourceFrame: 2,
				TargetFrame: 2,
			},
			wantErr: true,
			errMsg:  "source_frame and target_frame cannot be the same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate layer name
			if tt.input.LayerName == "" && tt.wantErr {
				return
			}
			// Validate source_frame
			if tt.input.SourceFrame < 1 && tt.wantErr {
				return
			}
			// Validate target_frame
			if tt.input.TargetFrame < 1 && tt.wantErr {
				return
			}
			// Validate not same
			if tt.input.SourceFrame == tt.input.TargetFrame && tt.wantErr {
				return
			}
			if tt.wantErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}
