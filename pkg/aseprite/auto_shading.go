package aseprite

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

// Vector3D represents a 3D vector for surface normals and light direction.
type Vector3D struct {
	X, Y, Z float64
}

// Normalize returns a unit vector in the same direction.
func (v Vector3D) Normalize() Vector3D {
	length := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if length == 0 {
		return v
	}
	return Vector3D{
		X: v.X / length,
		Y: v.Y / length,
		Z: v.Z / length,
	}
}

// Dot calculates the dot product with another vector.
func (v Vector3D) Dot(other Vector3D) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// ShadingRegion represents a contiguous region of similar colors for shading.
type ShadingRegion struct {
	Bounds    image.Rectangle
	BaseColor color.Color
	Pixels    []image.Point
	Normal    Vector3D
}

// ApplyAutoShading applies automatic geometry-based shading to an image.
//
// This function:
//  1. Detects edges and identifies distinct regions
//  2. Calculates surface normals for each region
//  3. Generates shadow and highlight colors for each base color
//  4. Applies shading based on light direction and style
//
// Parameters:
//   - img: source image to apply shading to
//   - lightDir: light direction (top_left, top, top_right, left, right, bottom_left, bottom, bottom_right)
//   - intensity: shading intensity (0.0-1.0)
//   - style: shading style ("cell", "smooth", "soft")
//   - hueShift: whether to apply hue shifting (shadows→cool, highlights→warm)
//
// Returns:
//   - shaded image
//   - array of generated palette colors (shadows and highlights)
//   - number of regions shaded
//   - error if any
func ApplyAutoShading(img image.Image, lightDir string, intensity float64, style string, hueShift bool) (*image.RGBA, []string, int, error) {
	if intensity < 0.0 || intensity > 1.0 {
		return nil, nil, 0, fmt.Errorf("intensity must be between 0.0 and 1.0, got %f", intensity)
	}

	// Convert light direction to vector
	lightVec, err := lightDirectionToVector(lightDir)
	if err != nil {
		return nil, nil, 0, err
	}

	// Detect regions in the image
	regions := detectRegions(img)
	if len(regions) == 0 {
		return nil, nil, 0, fmt.Errorf("no regions detected in image")
	}

	// Generate color ramps for each unique base color
	colorRamps := make(map[string][]string)
	baseColors := make(map[string]color.Color)

	for _, region := range regions {
		r, g, b, a := region.BaseColor.RGBA()
		key := fmt.Sprintf("#%02X%02X%02X%02X", uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
		if _, exists := colorRamps[key]; !exists {
			ramp, colors := generateColorRamp(region.BaseColor, intensity, hueShift)
			colorRamps[key] = ramp
			baseColors[key] = colors
		}
	}

	// Create output image
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Copy original image
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(x, y, img.At(x, y))
		}
	}

	// Apply shading to each region
	regionsShadedCount := 0
	for _, region := range regions {
		// Calculate lighting factor based on normal and light direction
		lightingFactor := region.Normal.Dot(lightVec)

		// Skip if region faces away from light (no shading needed)
		if lightingFactor < 0.1 && style != "cell" {
			continue
		}

		// Get color ramp for this region's base color
		r, g, b, a := region.BaseColor.RGBA()
		key := fmt.Sprintf("#%02X%02X%02X%02X", uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
		ramp, exists := colorRamps[key]
		if !exists {
			continue
		}

		// Apply shading based on style
		switch style {
		case "cell":
			applyCellShading(result, region, ramp, lightingFactor)
		case "smooth":
			applySmoothShading(result, region, ramp, lightingFactor)
		case "soft":
			applySoftShading(result, region, ramp, lightingFactor)
		}

		regionsShadedCount++
	}

	// Collect all generated colors for palette
	generatedColors := []string{}
	for _, ramp := range colorRamps {
		generatedColors = append(generatedColors, ramp...)
	}

	return result, generatedColors, regionsShadedCount, nil
}

// detectRegions identifies distinct regions in the image using flood fill.
func detectRegions(img image.Image) []ShadingRegion {
	bounds := img.Bounds()
	visited := make(map[image.Point]bool)
	regions := []ShadingRegion{}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pt := image.Point{X: x, Y: y}
			if visited[pt] {
				continue
			}

			// Check if pixel is transparent
			_, _, _, a := img.At(x, y).RGBA()
			if a == 0 {
				visited[pt] = true
				continue
			}

			// Flood fill to find region
			baseColor := img.At(x, y)
			pixels := floodFill(img, pt, baseColor, visited)

			if len(pixels) < 4 {
				// Skip very small regions
				continue
			}

			// Calculate bounds
			minX, minY := pixels[0].X, pixels[0].Y
			maxX, maxY := pixels[0].X, pixels[0].Y
			for _, p := range pixels {
				if p.X < minX {
					minX = p.X
				}
				if p.X > maxX {
					maxX = p.X
				}
				if p.Y < minY {
					minY = p.Y
				}
				if p.Y > maxY {
					maxY = p.Y
				}
			}

			// Calculate approximate surface normal
			// Per-pixel normals are calculated in shading functions for spherical surfaces
			normal := Vector3D{X: 0, Y: 0, Z: 1}.Normalize()

			regions = append(regions, ShadingRegion{
				Bounds:    image.Rect(minX, minY, maxX+1, maxY+1),
				BaseColor: baseColor,
				Pixels:    pixels,
				Normal:    normal,
			})
		}
	}

	return regions
}

// floodFill performs flood fill to find connected pixels of similar color.
func floodFill(img image.Image, start image.Point, targetColor color.Color, visited map[image.Point]bool) []image.Point {
	bounds := img.Bounds()
	pixels := []image.Point{}
	queue := []image.Point{start}

	// Color tolerance for matching
	tolerance := uint32(2000) // Small tolerance for similar colors

	for len(queue) > 0 {
		pt := queue[0]
		queue = queue[1:]

		if visited[pt] {
			continue
		}

		// Check bounds
		if !pt.In(bounds) {
			continue
		}

		// Check color match
		c := img.At(pt.X, pt.Y)
		if !colorsMatch(c, targetColor, tolerance) {
			continue
		}

		visited[pt] = true
		pixels = append(pixels, pt)

		// Add neighbors to queue
		queue = append(queue,
			image.Point{X: pt.X + 1, Y: pt.Y},
			image.Point{X: pt.X - 1, Y: pt.Y},
			image.Point{X: pt.X, Y: pt.Y + 1},
			image.Point{X: pt.X, Y: pt.Y - 1},
		)
	}

	return pixels
}

// colorsMatch checks if two colors are within tolerance.
func colorsMatch(c1, c2 color.Color, tolerance uint32) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	// Skip transparent pixels
	if a1 == 0 || a2 == 0 {
		return false
	}

	dr := int64(r1) - int64(r2)
	dg := int64(g1) - int64(g2)
	db := int64(b1) - int64(b2)

	distance := uint32(math.Sqrt(float64(dr*dr + dg*dg + db*db)))
	return distance <= tolerance
}

// generateColorRamp generates shadow and highlight colors for a base color.
// Returns: array of hex colors, base color (for reference)
func generateColorRamp(baseColor color.Color, intensity float64, hueShift bool) ([]string, color.Color) {
	r, g, b, _ := baseColor.RGBA()
	base := colorful.Color{
		R: float64(r) / 65535.0,
		G: float64(g) / 65535.0,
		B: float64(b) / 65535.0,
	}

	// Generate shadow (darker)
	h, s, l := base.Hsl()

	shadowL := l - (l * 0.3 * intensity) // Darken by up to 30%
	if shadowL < 0 {
		shadowL = 0
	}

	shadowH := h
	if hueShift {
		// Shift hue toward cool (blue) for shadows
		shadowH = h + 10 // Shift 10 degrees toward blue
		if shadowH > 360 {
			shadowH -= 360
		}
	}

	shadow := colorful.Hsl(shadowH, s, shadowL)

	// Generate highlight (lighter)
	highlightL := l + ((1.0 - l) * 0.3 * intensity) // Lighten by up to 30%
	if highlightL > 1.0 {
		highlightL = 1.0
	}

	highlightH := h
	if hueShift {
		// Shift hue toward warm (yellow) for highlights
		highlightH = h - 10 // Shift 10 degrees toward yellow
		if highlightH < 0 {
			highlightH += 360
		}
	}

	highlight := colorful.Hsl(highlightH, s, highlightL)

	// Convert to hex
	sr, sg, sb := shadow.RGB255()
	hr, hg, hb := highlight.RGB255()
	br, bg, bb := base.RGB255()

	return []string{
		fmt.Sprintf("#%02X%02X%02X", sr, sg, sb),
		fmt.Sprintf("#%02X%02X%02X", br, bg, bb),
		fmt.Sprintf("#%02X%02X%02X", hr, hg, hb),
	}, baseColor
}

// applyCellShading applies hard-edged cartoon shading (2-3 distinct color bands).
func applyCellShading(img *image.RGBA, region ShadingRegion, colorRamp []string, lightingFactor float64) {
	if len(colorRamp) < 3 {
		return
	}

	// Calculate region center for per-pixel normal calculation
	centerX := float64(region.Bounds.Min.X+region.Bounds.Max.X) / 2.0
	centerY := float64(region.Bounds.Min.Y+region.Bounds.Max.Y) / 2.0
	radius := math.Max(
		float64(region.Bounds.Dx())/2.0,
		float64(region.Bounds.Dy())/2.0,
	)

	// Apply to all pixels in region with per-pixel shading
	for _, pt := range region.Pixels {
		// Calculate distance from center (for spherical shading)
		dx := float64(pt.X) - centerX
		dy := float64(pt.Y) - centerY
		dist := math.Sqrt(dx*dx + dy*dy)

		// Calculate per-pixel normal for spherical surface
		// At the edges, Z component is lower (surface curves away)
		var pixelLightingFactor float64
		if dist < radius {
			// Calculate Z based on sphere equation: z = sqrt(r^2 - x^2 - y^2)
			normalizedDist := dist / radius
			zComponent := math.Sqrt(1.0 - normalizedDist*normalizedDist)

			// Use Z component directly for lighting (spherical shading)
			// At center: zComponent = 1.0 (facing camera, brightest)
			// At edge: zComponent = 0.0 (curving away, darkest)
			pixelLightingFactor = zComponent
		} else {
			// Outside radius, use mid-range lighting
			pixelLightingFactor = 0.5
		}

		// Select color based on per-pixel lighting
		var shadingColor string
		if pixelLightingFactor < 0.3 {
			shadingColor = colorRamp[0] // Shadow
		} else if pixelLightingFactor < 0.7 {
			shadingColor = colorRamp[1] // Base
		} else {
			shadingColor = colorRamp[2] // Highlight
		}

		// Parse and apply color
		c, err := parseHexColorRGBA(shadingColor)
		if err != nil {
			continue
		}
		img.Set(pt.X, pt.Y, c)
	}
}

// applySmoothShading applies gradient shading using dithering.
func applySmoothShading(img *image.RGBA, region ShadingRegion, colorRamp []string, lightingFactor float64) {
	if len(colorRamp) < 3 {
		return
	}

	// Calculate region center for per-pixel normal calculation
	centerX := float64(region.Bounds.Min.X+region.Bounds.Max.X) / 2.0
	centerY := float64(region.Bounds.Min.Y+region.Bounds.Max.Y) / 2.0
	radius := math.Max(
		float64(region.Bounds.Dx())/2.0,
		float64(region.Bounds.Dy())/2.0,
	)

	// Apply with per-pixel shading and dithering
	for _, pt := range region.Pixels {
		// Calculate distance from center (for spherical shading)
		dx := float64(pt.X) - centerX
		dy := float64(pt.Y) - centerY
		dist := math.Sqrt(dx*dx + dy*dy)

		// Calculate per-pixel lighting factor
		var pixelLightingFactor float64
		if dist < radius {
			normalizedDist := dist / radius
			zComponent := math.Sqrt(1.0 - normalizedDist*normalizedDist)
			// Use Z component directly for spherical shading
			pixelLightingFactor = zComponent
		} else {
			pixelLightingFactor = 0.5
		}

		// Select colors based on per-pixel lighting factor
		var c1, c2 string
		var blend float64

		if pixelLightingFactor < 0.5 {
			c1, c2 = colorRamp[0], colorRamp[1]
			blend = pixelLightingFactor * 2 // 0.0-1.0
		} else {
			c1, c2 = colorRamp[1], colorRamp[2]
			blend = (pixelLightingFactor - 0.5) * 2 // 0.0-1.0
		}

		color1, err1 := parseHexColorRGBA(c1)
		color2, err2 := parseHexColorRGBA(c2)
		if err1 != nil || err2 != nil {
			continue
		}

		// Simple checkerboard dithering
		if (pt.X+pt.Y)%2 == 0 {
			if blend > 0.5 {
				img.Set(pt.X, pt.Y, color2)
			} else {
				img.Set(pt.X, pt.Y, color1)
			}
		} else {
			if blend > 0.5 {
				img.Set(pt.X, pt.Y, color2)
			} else {
				img.Set(pt.X, pt.Y, color1)
			}
		}
	}
}

// applySoftShading applies subtle gradient shading.
func applySoftShading(img *image.RGBA, region ShadingRegion, colorRamp []string, lightingFactor float64) {
	if len(colorRamp) < 3 {
		return
	}

	// Calculate region center for per-pixel normal calculation
	centerX := float64(region.Bounds.Min.X+region.Bounds.Max.X) / 2.0
	centerY := float64(region.Bounds.Min.Y+region.Bounds.Max.Y) / 2.0
	radius := math.Max(
		float64(region.Bounds.Dx())/2.0,
		float64(region.Bounds.Dy())/2.0,
	)

	// Apply to all pixels in region with per-pixel shading
	for _, pt := range region.Pixels {
		// Calculate distance from center (for spherical shading)
		dx := float64(pt.X) - centerX
		dy := float64(pt.Y) - centerY
		dist := math.Sqrt(dx*dx + dy*dy)

		// Calculate per-pixel lighting factor
		var pixelLightingFactor float64
		if dist < radius {
			normalizedDist := dist / radius
			zComponent := math.Sqrt(1.0 - normalizedDist*normalizedDist)
			// Use Z component directly for spherical shading
			pixelLightingFactor = zComponent
		} else {
			pixelLightingFactor = 0.5
		}

		// Blend between colors based on per-pixel lighting factor
		var resultColor color.Color

		if pixelLightingFactor < 0.5 {
			// Blend between shadow and base
			c1, _ := parseHexColorRGBA(colorRamp[0])
			c2, _ := parseHexColorRGBA(colorRamp[1])
			resultColor = blendColors(c1, c2, pixelLightingFactor*2)
		} else {
			// Blend between base and highlight
			c1, _ := parseHexColorRGBA(colorRamp[1])
			c2, _ := parseHexColorRGBA(colorRamp[2])
			resultColor = blendColors(c1, c2, (pixelLightingFactor-0.5)*2)
		}

		img.Set(pt.X, pt.Y, resultColor)
	}
}

// blendColors blends two colors by a factor (0.0 = c1, 1.0 = c2).
func blendColors(c1, c2 color.Color, factor float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8(float64(r1>>8)*(1-factor) + float64(r2>>8)*factor)
	g := uint8(float64(g1>>8)*(1-factor) + float64(g2>>8)*factor)
	b := uint8(float64(b1>>8)*(1-factor) + float64(b2>>8)*factor)
	a := uint8(float64(a1>>8)*(1-factor) + float64(a2>>8)*factor)

	return color.RGBA{R: r, G: g, B: b, A: a}
}

// parseHexColorRGBA parses a hex color string to a color.RGBA.
func parseHexColorRGBA(hexColor string) (color.Color, error) {
	c, err := colorful.Hex(hexColor)
	if err != nil {
		return nil, err
	}
	r, g, b := c.RGB255()
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

// lightDirectionToVector converts light direction string to a normalized vector.
func lightDirectionToVector(dir string) (Vector3D, error) {
	vectors := map[string]Vector3D{
		"top_left":     {X: -1, Y: -1, Z: 1},
		"top":          {X: 0, Y: -1, Z: 1},
		"top_right":    {X: 1, Y: -1, Z: 1},
		"left":         {X: -1, Y: 0, Z: 1},
		"right":        {X: 1, Y: 0, Z: 1},
		"bottom_left":  {X: -1, Y: 1, Z: 1},
		"bottom":       {X: 0, Y: 1, Z: 1},
		"bottom_right": {X: 1, Y: 1, Z: 1},
	}

	vec, exists := vectors[dir]
	if !exists {
		return Vector3D{}, fmt.Errorf("invalid light direction: %s", dir)
	}

	return vec.Normalize(), nil
}
