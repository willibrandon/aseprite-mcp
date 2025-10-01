package aseprite

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/lucasb-eyer/go-colorful"
)

// PaletteColor represents a single color in a palette with metadata.
type PaletteColor struct {
	Color        string  `json:"color"`         // Hex color #RRGGBB
	Hue          float64 `json:"hue"`           // 0-360 degrees
	Saturation   float64 `json:"saturation"`    // 0-100%
	Lightness    float64 `json:"lightness"`     // 0-100%
	UsagePercent float64 `json:"usage_percent"` // Percentage of pixels using this color
	Role         string  `json:"role"`          // "dark_shadow", "midtone", "highlight", etc.
}

// ExtractPalette extracts a limited color palette from an image using k-means clustering.
// Colors are extracted in LAB color space (perceptually uniform) and sorted by hue then lightness.
func ExtractPalette(img image.Image, numColors int) ([]PaletteColor, error) {
	if numColors < 2 || numColors > 256 {
		return nil, fmt.Errorf("numColors must be between 2 and 256, got %d", numColors)
	}

	// Step 1: Sample pixels (subsample if image is large)
	pixels := samplePixels(img, 10000)
	if len(pixels) == 0 {
		return nil, fmt.Errorf("no pixels to sample from image")
	}

	// Step 2: Convert RGB to LAB color space
	labPixels := make([]colorful.Color, len(pixels))
	for i, p := range pixels {
		r, g, b, _ := p.RGBA()
		rgb := colorful.Color{
			R: float64(r) / 65535.0,
			G: float64(g) / 65535.0,
			B: float64(b) / 65535.0,
		}
		labPixels[i] = rgb
	}

	// Step 3: Run k-means clustering
	centroids := kmeansClustering(labPixels, numColors, 100)

	// Step 4: Convert centroids back to RGB hex and calculate metadata
	palette := make([]PaletteColor, len(centroids))
	for i, centroid := range centroids {
		h, s, l := centroid.Hsl()
		r, g, b := centroid.RGB255()

		palette[i] = PaletteColor{
			Color:      fmt.Sprintf("#%02X%02X%02X", r, g, b),
			Hue:        h,
			Saturation: s * 100, // Convert to percentage
			Lightness:  l * 100, // Convert to percentage
		}
	}

	// Step 5: Calculate usage percentages
	calculateUsagePercentages(palette, labPixels, centroids)

	// Step 6: Sort by hue, then lightness (Lospec standard)
	sort.Slice(palette, func(i, j int) bool {
		// If hues are very close (within 5 degrees), sort by lightness
		if math.Abs(palette[i].Hue-palette[j].Hue) < 5 {
			return palette[i].Lightness < palette[j].Lightness
		}
		return palette[i].Hue < palette[j].Hue
	})

	// Step 7: Assign roles based on lightness
	assignPaletteRoles(palette)

	return palette, nil
}

// samplePixels samples up to maxSamples pixels from the image.
// If the image has fewer pixels, all pixels are sampled.
func samplePixels(img image.Image, maxSamples int) []color.Color {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	totalPixels := width * height

	if totalPixels <= maxSamples {
		// Sample all pixels
		pixels := make([]color.Color, 0, totalPixels)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				pixels = append(pixels, img.At(x, y))
			}
		}
		return pixels
	}

	// Subsample pixels uniformly
	step := int(math.Sqrt(float64(totalPixels) / float64(maxSamples)))
	if step < 1 {
		step = 1
	}

	pixels := make([]color.Color, 0, maxSamples)
	for y := 0; y < height; y += step {
		for x := 0; x < width; x += step {
			pixels = append(pixels, img.At(x, y))
		}
	}

	return pixels
}

// kmeansClustering performs k-means clustering on LAB color values.
// Returns k centroids representing the cluster centers.
func kmeansClustering(pixels []colorful.Color, k int, maxIterations int) []colorful.Color {
	if len(pixels) == 0 || k <= 0 {
		return nil
	}

	// Ensure k doesn't exceed number of pixels
	if k > len(pixels) {
		k = len(pixels)
	}

	// Initialize centroids randomly from pixel set
	centroids := make([]colorful.Color, k)
	indices := rand.Perm(len(pixels))
	for i := 0; i < k; i++ {
		centroids[i] = pixels[indices[i]]
	}

	// Cluster assignments
	assignments := make([]int, len(pixels))

	// Iterate until convergence or max iterations
	for iter := 0; iter < maxIterations; iter++ {
		changed := false

		// Assign each pixel to nearest centroid
		for i, pixel := range pixels {
			minDist := math.MaxFloat64
			minIdx := 0

			for j, centroid := range centroids {
				dist := colorDistanceLab(pixel, centroid)
				if dist < minDist {
					minDist = dist
					minIdx = j
				}
			}

			if assignments[i] != minIdx {
				assignments[i] = minIdx
				changed = true
			}
		}

		// If no assignments changed, we've converged
		if !changed {
			break
		}

		// Recalculate centroids as mean of assigned pixels
		for j := 0; j < k; j++ {
			var sumL, sumA, sumB float64
			count := 0

			for i, pixel := range pixels {
				if assignments[i] == j {
					l, a, b := pixel.Lab()
					sumL += l
					sumA += a
					sumB += b
					count++
				}
			}

			if count > 0 {
				l := sumL / float64(count)
				a := sumA / float64(count)
				b := sumB / float64(count)
				centroids[j] = colorful.Lab(l, a, b)
			}
		}
	}

	return centroids
}

// colorDistanceLab calculates Euclidean distance between two colors in LAB space.
func colorDistanceLab(c1, c2 colorful.Color) float64 {
	l1, a1, b1 := c1.Lab()
	l2, a2, b2 := c2.Lab()

	dl := l1 - l2
	da := a1 - a2
	db := b1 - b2

	return math.Sqrt(dl*dl + da*da + db*db)
}

// calculateUsagePercentages calculates what percentage of pixels use each palette color.
func calculateUsagePercentages(palette []PaletteColor, pixels []colorful.Color, centroids []colorful.Color) {
	counts := make([]int, len(centroids))

	// Count assignments to each centroid
	for _, pixel := range pixels {
		minDist := math.MaxFloat64
		minIdx := 0

		for j, centroid := range centroids {
			dist := colorDistanceLab(pixel, centroid)
			if dist < minDist {
				minDist = dist
				minIdx = j
			}
		}

		counts[minIdx]++
	}

	// Calculate percentages
	total := len(pixels)
	for i := range palette {
		if i < len(counts) {
			palette[i].UsagePercent = float64(counts[i]) * 100.0 / float64(total)
		}
	}
}

// assignPaletteRoles assigns semantic roles to palette colors based on lightness.
func assignPaletteRoles(palette []PaletteColor) {
	if len(palette) == 0 {
		return
	}

	// Create a copy sorted by lightness for role assignment
	byLightness := make([]PaletteColor, len(palette))
	copy(byLightness, palette)
	sort.Slice(byLightness, func(i, j int) bool {
		return byLightness[i].Lightness < byLightness[j].Lightness
	})

	// Assign roles based on position in lightness order
	n := len(palette)
	for i := range palette {
		// Find this color in the lightness-sorted list
		lightnessRank := 0
		for j := range byLightness {
			if byLightness[j].Color == palette[i].Color {
				lightnessRank = j
				break
			}
		}

		// Assign role based on lightness rank
		ratio := float64(lightnessRank) / float64(n-1)
		if ratio < 0.2 {
			palette[i].Role = "dark_shadow"
		} else if ratio < 0.4 {
			palette[i].Role = "shadow"
		} else if ratio < 0.6 {
			palette[i].Role = "midtone"
		} else if ratio < 0.8 {
			palette[i].Role = "light"
		} else {
			palette[i].Role = "highlight"
		}
	}
}

// FindClosestPaletteColor finds the palette color closest to the given color.
func FindClosestPaletteColor(targetColor colorful.Color, palette []PaletteColor) (string, int, error) {
	if len(palette) == 0 {
		return "", -1, fmt.Errorf("palette is empty")
	}

	minDist := math.MaxFloat64
	minIdx := 0

	for i, pColor := range palette {
		// Parse palette color
		c, err := colorful.Hex(pColor.Color)
		if err != nil {
			continue
		}

		dist := colorDistanceLab(targetColor, c)
		if dist < minDist {
			minDist = dist
			minIdx = i
		}
	}

	return palette[minIdx].Color, minIdx, nil
}
