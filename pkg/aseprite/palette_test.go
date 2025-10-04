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

func TestSamplePixels_LargeImage(t *testing.T) {
	// Create a large image to test subsampling
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	for y := 0; y < 1000; y++ {
		for x := 0; x < 1000; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 0, 255})
		}
	}

	pixels := samplePixels(img, 5000)

	if len(pixels) == 0 {
		t.Error("samplePixels() returned no pixels")
	}

	if len(pixels) > 10000 {
		t.Errorf("samplePixels() returned too many pixels: %d", len(pixels))
	}
}

func TestSamplePixels_SmallImage(t *testing.T) {
	// Create a small image - all pixels should be sampled
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	pixels := samplePixels(img, 10000)

	if len(pixels) != 100 {
		t.Errorf("samplePixels() = %d pixels, want 100", len(pixels))
	}
}

func TestKmeansClustering_EdgeCases(t *testing.T) {
	// Test with exactly k colors
	pixels := []colorful.Color{
		{R: 1.0, G: 0.0, B: 0.0},
		{R: 0.0, G: 1.0, B: 0.0},
		{R: 0.0, G: 0.0, B: 1.0},
	}

	centers := kmeansClustering(pixels, 3, 10)

	if len(centers) != 3 {
		t.Errorf("kmeansClustering() returned %d centers, want 3", len(centers))
	}
}

func TestAssignPaletteRoles(t *testing.T) {
	palette := []PaletteColor{
		{Color: "#000000", Hue: 0, Saturation: 0, Lightness: 0},
		{Color: "#808080", Hue: 0, Saturation: 0, Lightness: 50},
		{Color: "#FFFFFF", Hue: 0, Saturation: 0, Lightness: 100},
		{Color: "#FF0000", Hue: 0, Saturation: 100, Lightness: 50},
	}

	// assignPaletteRoles modifies palette in place
	assignPaletteRoles(palette)

	// Check that roles are assigned (should be one of: dark_shadow, shadow, midtone, light, highlight)
	validRoles := map[string]bool{
		"dark_shadow": true,
		"shadow":      true,
		"midtone":     true,
		"light":       true,
		"highlight":   true,
	}

	for _, pc := range palette {
		if pc.Role == "" {
			t.Error("assignPaletteRoles() did not assign a role to all colors")
		}
		if !validRoles[pc.Role] {
			t.Errorf("assignPaletteRoles() assigned invalid role: %s", pc.Role)
		}
	}
}

func TestFindClosestPaletteColor_EmptyPalette(t *testing.T) {
	targetColor, _ := colorful.Hex("#FF0000")

	_, _, err := FindClosestPaletteColor(targetColor, []PaletteColor{})

	if err == nil {
		t.Error("FindClosestPaletteColor() should return error for empty palette")
	}
}

func TestFindClosestPaletteColor_InvalidHex(t *testing.T) {
	palette := []PaletteColor{
		{Color: "invalid"},
		{Color: "#FF0000"},
	}

	targetColor, _ := colorful.Hex("#00FF00")

	// Should skip invalid color and find the valid one
	gotColor, gotIdx, err := FindClosestPaletteColor(targetColor, palette)

	if err != nil {
		t.Errorf("FindClosestPaletteColor() unexpected error: %v", err)
	}

	if gotIdx != 1 {
		t.Errorf("FindClosestPaletteColor() index = %d, want 1", gotIdx)
	}

	if gotColor != "#FF0000" {
		t.Errorf("FindClosestPaletteColor() color = %s, want #FF0000", gotColor)
	}
}

func TestExtractPalette_EdgeCases(t *testing.T) {
	// Test with minimum colors
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	result, err := ExtractPalette(img, 2)

	if err != nil {
		t.Errorf("ExtractPalette() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("ExtractPalette() returned %d colors, want 2", len(result))
	}

	// Check that usage percentages sum to ~100%
	totalUsage := 0.0
	for _, pc := range result {
		totalUsage += pc.UsagePercent
	}

	if totalUsage < 99.9 || totalUsage > 100.1 {
		t.Errorf("ExtractPalette() usage percentages sum to %.2f, want ~100", totalUsage)
	}
}
