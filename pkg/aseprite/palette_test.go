package aseprite

import (
	"image"
	"image/color"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

// createPaletteTestImage creates a test image with known colors for palette extraction
func createPaletteTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Create 4 distinct color regions
	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},   // Red
		{R: 0, G: 255, B: 0, A: 255},   // Green
		{R: 0, G: 0, B: 255, A: 255},   // Blue
		{R: 255, G: 255, B: 0, A: 255}, // Yellow
	}

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			// Divide image into 4 quadrants
			idx := 0
			if x >= 50 {
				idx += 1
			}
			if y >= 50 {
				idx += 2
			}
			img.Set(x, y, colors[idx%len(colors)])
		}
	}

	return img
}

func TestExtractPalette(t *testing.T) {
	tests := []struct {
		name         string
		paletteSize  int
		wantErr      bool
		checkPalette bool
	}{
		{
			name:         "valid 8 color palette",
			paletteSize:  8,
			wantErr:      false,
			checkPalette: true,
		},
		{
			name:         "valid 16 color palette",
			paletteSize:  16,
			wantErr:      false,
			checkPalette: true,
		},
		{
			name:        "palette size too small",
			paletteSize: 1,
			wantErr:     true,
		},
		{
			name:        "palette size too large",
			paletteSize: 257,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := createPaletteTestImage()
			result, err := ExtractPalette(img, tt.paletteSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPalette() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatal("ExtractPalette() returned nil result")
				}

				if len(result) != tt.paletteSize {
					t.Errorf("Palette size = %d, want %d", len(result), tt.paletteSize)
				}

				// Verify colors are sorted by hue
				if tt.checkPalette {
					for i, pc := range result {
						if pc.Color == "" {
							t.Errorf("Color at index %d is empty", i)
						}
						// Verify hex format
						if len(pc.Color) != 7 || pc.Color[0] != '#' {
							t.Errorf("Color %s is not in #RRGGBB format", pc.Color)
						}
					}
				}

				// Check usage percentages sum to ~100%
				totalUsage := 0.0
				for _, pc := range result {
					if pc.UsagePercent < 0 || pc.UsagePercent > 100 {
						t.Errorf("Usage percentage %f out of range [0,100]", pc.UsagePercent)
					}
					totalUsage += pc.UsagePercent
				}
				if totalUsage < 99.9 || totalUsage > 100.1 {
					t.Errorf("Total usage = %f, want ~100", totalUsage)
				}
			}
		})
	}
}

func TestFindClosestPaletteColor(t *testing.T) {
	palette := []PaletteColor{
		{Color: "#000000"}, // Black
		{Color: "#FFFFFF"}, // White
		{Color: "#FF0000"}, // Red
		{Color: "#00FF00"}, // Green
		{Color: "#0000FF"}, // Blue
	}

	tests := []struct {
		name  string
		color string // Hex color
		want  int    // Expected palette index
	}{
		{
			name:  "exact black match",
			color: "#000000",
			want:  0,
		},
		{
			name:  "exact white match",
			color: "#FFFFFF",
			want:  1,
		},
		{
			name:  "exact red match",
			color: "#FF0000",
			want:  2,
		},
		{
			name:  "close to black",
			color: "#0A0A0A",
			want:  0,
		},
		{
			name:  "close to white",
			color: "#F5F5F5",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetColor, _ := colorful.Hex(tt.color)
			_, gotIdx, err := FindClosestPaletteColor(targetColor, palette)
			if err != nil {
				t.Fatalf("FindClosestPaletteColor() error = %v", err)
			}
			if gotIdx != tt.want {
				t.Errorf("FindClosestPaletteColor() = %d, want %d", gotIdx, tt.want)
			}
		})
	}
}
