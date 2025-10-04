package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlendColors(t *testing.T) {
	tests := []struct {
		name   string
		color1 string
		color2 string
		want   string
	}{
		{
			name:   "blend black and white",
			color1: "#000000FF",
			color2: "#FFFFFFFF",
			want:   "#7F7F7FFF", // RGB blended, alpha blended (FF+FF)/2=FF
		},
		{
			name:   "blend red and blue",
			color1: "#FF0000FF",
			color2: "#0000FFFF",
			want:   "#7F007FFF", // Purple, fully opaque
		},
		{
			name:   "blend with transparency",
			color1: "#FF000080",
			color2: "#00FF00FF",
			want:   "#7F7F00BF", // RGB blended, alpha blended (80+FF)/2=BF
		},
		{
			name:   "blend same color",
			color1: "#AB5236FF",
			color2: "#AB5236FF",
			want:   "#AB5236FF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blendColors(tt.color1, tt.color2)
			if got != tt.want {
				t.Errorf("blendColors(%q, %q) = %q, want %q",
					tt.color1, tt.color2, got, tt.want)
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name  string
		hex   string
		wantR uint8
		wantG uint8
		wantB uint8
		wantA uint8
	}{
		{
			name:  "full RGBA",
			hex:   "#FF00AABB",
			wantR: 255,
			wantG: 0,
			wantB: 170,
			wantA: 187,
		},
		{
			name:  "RGB without alpha",
			hex:   "#123456",
			wantR: 18,
			wantG: 52,
			wantB: 86,
			wantA: 255, // default alpha
		},
		{
			name:  "without # prefix",
			hex:   "ABCDEF00",
			wantR: 171,
			wantG: 205,
			wantB: 239,
			wantA: 0,
		},
		{
			name:  "black with full alpha",
			hex:   "#000000FF",
			wantR: 0,
			wantG: 0,
			wantB: 0,
			wantA: 255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, a := parseHexColor(tt.hex)
			if r != tt.wantR || g != tt.wantG || b != tt.wantB || a != tt.wantA {
				t.Errorf("parseHexColor(%q) = (%d, %d, %d, %d), want (%d, %d, %d, %d)",
					tt.hex, r, g, b, a, tt.wantR, tt.wantG, tt.wantB, tt.wantA)
			}
		})
	}
}

func TestIsTransparent(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{
			name:  "fully transparent",
			color: "#FF000000",
			want:  true,
		},
		{
			name:  "fully opaque",
			color: "#FF0000FF",
			want:  false,
		},
		{
			name:  "semi-transparent",
			color: "#FF000080",
			want:  false,
		},
		{
			name:  "short format",
			color: "#FFF",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTransparent(tt.color)
			if got != tt.want {
				t.Errorf("isTransparent(%q) = %v, want %v", tt.color, got, tt.want)
			}
		})
	}
}

func TestBuildPixelGrid(t *testing.T) {
	pixels := []PixelData{
		{X: 0, Y: 0, Color: "#FF0000FF"},
		{X: 1, Y: 0, Color: "#00FF00FF"},
		{X: 0, Y: 1, Color: "#0000FFFF"},
		{X: 1, Y: 1, Color: "#FFFFFFFF"},
	}

	grid := buildPixelGrid(pixels)

	tests := []struct {
		x, y int
		want string
	}{
		{0, 0, "#FF0000FF"},
		{1, 0, "#00FF00FF"},
		{0, 1, "#0000FFFF"},
		{1, 1, "#FFFFFFFF"},
		{2, 2, ""}, // out of bounds
	}

	for _, tt := range tests {
		got := getPixel(grid, tt.x, tt.y)
		if got != tt.want {
			t.Errorf("getPixel(grid, %d, %d) = %q, want %q", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestDetectJaggedEdges(t *testing.T) {
	// Create a simple test grid with a diagonal edge
	// Pattern:   ..##
	//            .##.
	//            ##..
	grid := map[int]map[int]string{
		0: {2: "#FF0000FF", 3: "#FF0000FF"},
		1: {1: "#FF0000FF", 2: "#FF0000FF"},
		2: {0: "#FF0000FF", 1: "#FF0000FF"},
	}

	region := Region{X: 0, Y: 0, Width: 4, Height: 3}
	suggestions := detectJaggedEdges(grid, region, 128, false)

	// Should detect diagonal edges and suggest intermediate pixels
	if len(suggestions) == 0 {
		t.Error("Expected to detect jagged edges, got no suggestions")
	}

	// Verify suggestions have required fields
	for i, sug := range suggestions {
		if sug.SuggestedColor == "" {
			t.Errorf("Suggestion %d missing suggested color", i)
		}
		if sug.Direction == "" {
			t.Errorf("Suggestion %d missing direction", i)
		}
	}
}

func TestCheckDiagonalNE(t *testing.T) {
	// Pattern: ##  (current at 0,0 and 1,0)
	//          .#  (empty at 0,1, current at 1,1)
	// Should suggest filling (0,1)
	grid := map[int]map[int]string{
		0: {0: "#FF0000FF", 1: "#FF0000FF"},
		1: {1: "#FF0000FF"},
	}

	current := "#FF0000FF"
	suggestion := checkDiagonalNE(grid, 0, 0, current)

	if suggestion == nil {
		t.Fatal("Expected suggestion for NE diagonal, got nil")
	}

	if suggestion.X != 0 || suggestion.Y != 1 {
		t.Errorf("Expected suggestion at (0, 1), got (%d, %d)", suggestion.X, suggestion.Y)
	}

	if suggestion.Direction != "diagonal_ne" {
		t.Errorf("Expected direction 'diagonal_ne', got %q", suggestion.Direction)
	}

	if suggestion.SuggestedColor == "" {
		t.Error("Suggested color should not be empty")
	}
}

func TestCheckDiagonalSE(t *testing.T) {
	// Pattern: .#  (empty at 0,0, current at 1,0)
	//          ##  (current at 0,1 and 1,1)
	// Should suggest filling (0,0)
	grid := map[int]map[int]string{
		0: {1: "#00FF00FF"},
		1: {0: "#00FF00FF", 1: "#00FF00FF"},
	}

	current := "" // empty pixel
	suggestion := checkDiagonalSE(grid, 0, 0, current)

	if suggestion == nil {
		t.Fatal("Expected suggestion for SE diagonal, got nil")
	}

	if suggestion.X != 0 || suggestion.Y != 0 {
		t.Errorf("Expected suggestion at (0, 0), got (%d, %d)", suggestion.X, suggestion.Y)
	}

	if suggestion.Direction != "diagonal_se" {
		t.Errorf("Expected direction 'diagonal_se', got %q", suggestion.Direction)
	}
}

func TestFindClosestPaletteColor(t *testing.T) {
	palette := []string{"#000000", "#FF0000", "#00FF00", "#0000FF", "#FFFFFF"}

	tests := []struct {
		name    string
		r, g, b uint8
		want    string
	}{
		{
			name: "exact match black",
			r:    0, g: 0, b: 0,
			want: "#000000",
		},
		{
			name: "exact match red",
			r:    255, g: 0, b: 0,
			want: "#FF0000",
		},
		{
			name: "close to red",
			r:    200, g: 50, b: 50,
			want: "#FF0000",
		},
		{
			name: "close to white",
			r:    240, g: 240, b: 240,
			want: "#FFFFFF",
		},
		{
			name: "close to green",
			r:    50, g: 200, b: 50,
			want: "#00FF00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findClosestPaletteColor(tt.r, tt.g, tt.b, palette)
			if got != tt.want {
				t.Errorf("findClosestPaletteColor(%d, %d, %d) = %q, want %q",
					tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}

func TestFindClosestPaletteColor_EmptyPalette(t *testing.T) {
	palette := []string{}
	got := findClosestPaletteColor(128, 64, 32, palette)
	want := "#804020"
	if got != want {
		t.Errorf("findClosestPaletteColor with empty palette = %q, want %q", got, want)
	}
}

// Unit tests for checkDiagonalSW helper function
func TestCheckDiagonalSW(t *testing.T) {
	tests := []struct {
		name    string
		grid    map[int]map[int]string
		x       int
		y       int
		current string
		want    *EdgeSuggestion
	}{
		{
			name:    "x is 0 - edge case",
			grid:    map[int]map[int]string{},
			x:       0,
			y:       5,
			current: "",
			want:    nil,
		},
		{
			name: "diagonal SW pattern detected",
			grid: map[int]map[int]string{
				0: {0: "#FF0000FF"},                 // row 0: (0,0) has color, (1,0) is empty
				1: {0: "#FF0000FF", 1: "#FF0000FF"}, // row 1: (0,1) and (1,1) have color
			},
			x:       1,
			y:       0,
			current: "",
			want: &EdgeSuggestion{
				X:              1,
				Y:              0,
				CurrentColor:   "",
				NeighborColor:  "#FF0000FF",
				SuggestedColor: "#FF000080",
				Direction:      "diagonal_sw",
			},
		},
		{
			name: "no pattern - left is empty",
			grid: map[int]map[int]string{
				0: {},
				1: {1: "#FF0000FF"},
			},
			x:       1,
			y:       0,
			current: "",
			want:    nil,
		},
		{
			name: "no pattern - below doesn't match",
			grid: map[int]map[int]string{
				0: {0: "#FF0000FF", 1: "#FF0000FF"},
				1: {1: "#00FF00FF"}, // Different color
			},
			x:       1,
			y:       0,
			current: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkDiagonalSW(tt.grid, tt.x, tt.y, tt.current)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.X, got.X)
				assert.Equal(t, tt.want.Y, got.Y)
				assert.Equal(t, tt.want.Direction, got.Direction)
				assert.NotEmpty(t, got.SuggestedColor)
			}
		})
	}
}

// Unit tests for checkDiagonalNW helper function
func TestCheckDiagonalNW(t *testing.T) {
	tests := []struct {
		name    string
		grid    map[int]map[int]string
		x       int
		y       int
		current string
		want    *EdgeSuggestion
	}{
		{
			name:    "x is 0 - edge case",
			grid:    map[int]map[int]string{},
			x:       0,
			y:       5,
			current: "#FF0000FF",
			want:    nil,
		},
		{
			name: "diagonal NW pattern detected",
			grid: map[int]map[int]string{
				0: {0: "#FF0000FF", 1: "#FF0000FF"}, // row 0: (0,0) and (1,0) have color
				1: {0: "#FF0000FF"},                 // row 1: (0,1) has color, (1,1) is empty
			},
			x:       1,
			y:       0,
			current: "#FF0000FF",
			want: &EdgeSuggestion{
				X:              1,
				Y:              1,
				CurrentColor:   "",
				NeighborColor:  "#FF0000FF",
				SuggestedColor: "#FF000080",
				Direction:      "diagonal_nw",
			},
		},
		{
			name: "no pattern - left doesn't match current",
			grid: map[int]map[int]string{
				0: {0: "#00FF00FF", 1: "#FF0000FF"}, // Different colors
				1: {0: "#FF0000FF"},
			},
			x:       1,
			y:       0,
			current: "#FF0000FF",
			want:    nil,
		},
		{
			name: "no pattern - belowLeft doesn't match current",
			grid: map[int]map[int]string{
				0: {0: "#FF0000FF", 1: "#FF0000FF"},
				1: {0: "#00FF00FF"}, // Different color
			},
			x:       1,
			y:       0,
			current: "#FF0000FF",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkDiagonalNW(tt.grid, tt.x, tt.y, tt.current)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.X, got.X)
				assert.Equal(t, tt.want.Y, got.Y)
				assert.Equal(t, tt.want.Direction, got.Direction)
				assert.NotEmpty(t, got.SuggestedColor)
			}
		})
	}
}
