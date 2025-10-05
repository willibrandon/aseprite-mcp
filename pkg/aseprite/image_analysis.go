package aseprite

import (
	"fmt"
	"image"
	"math"

	"github.com/nfnt/resize"
)

// BrightnessMap represents quantized brightness levels across an image.
type BrightnessMap struct {
	Grid   [][]int           `json:"grid"`   // 2D array of brightness levels
	Legend map[string]string `json:"legend"` // Maps level number to description
}

// EdgeMap represents detected edges in an image.
type EdgeMap struct {
	Grid       [][]int    `json:"grid"`        // 2D array where 1 = edge, 0 = no edge
	MajorEdges []EdgeLine `json:"major_edges"` // Significant edge contours
}

// EdgeLine represents a detected edge line.
type EdgeLine struct {
	From     Point   `json:"from"`     // Starting point
	To       Point   `json:"to"`       // Ending point
	Strength float64 `json:"strength"` // Edge strength 0-100
}

// Point is defined in types.go

// GenerateBrightnessMap creates a quantized brightness map from an image.
// The image is downsampled to targetW x targetH and brightness is quantized into numLevels.
func GenerateBrightnessMap(img image.Image, targetW, targetH, numLevels int) (*BrightnessMap, error) {
	if targetW <= 0 || targetH <= 0 {
		return nil, fmt.Errorf("target dimensions must be positive, got %dx%d", targetW, targetH)
	}
	if numLevels < 2 || numLevels > 256 {
		return nil, fmt.Errorf("numLevels must be between 2 and 256, got %d", numLevels)
	}

	// Downsample image to target dimensions
	resized := resize.Resize(uint(targetW), uint(targetH), img, resize.Bilinear)

	// Create brightness grid
	grid := make([][]int, targetH)
	for y := 0; y < targetH; y++ {
		grid[y] = make([]int, targetW)
		for x := 0; x < targetW; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()

			// Convert to 0-255 range
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Calculate grayscale using Rec. 709 luma coefficients
			// These coefficients account for human eye sensitivity to different colors
			gray := 0.2126*float64(r8) + 0.7152*float64(g8) + 0.0722*float64(b8)

			// Quantize to brightness level
			level := int(gray / 256.0 * float64(numLevels))
			if level >= numLevels {
				level = numLevels - 1
			}

			grid[y][x] = level
		}
	}

	// Create legend
	legend := make(map[string]string)
	for i := 0; i < numLevels; i++ {
		ratio := float64(i) / float64(numLevels-1)
		var desc string
		if ratio < 0.2 {
			desc = "darkest"
		} else if ratio < 0.4 {
			desc = "dark"
		} else if ratio < 0.6 {
			desc = "mid"
		} else if ratio < 0.8 {
			desc = "light"
		} else {
			desc = "lightest"
		}
		legend[fmt.Sprintf("%d", i)] = desc
	}

	return &BrightnessMap{
		Grid:   grid,
		Legend: legend,
	}, nil
}

// DetectEdges applies Sobel edge detection to an image.
// Returns a binary edge map where 1 = edge detected, 0 = no edge.
// If targetW and targetH are > 0, the image is downsampled before edge detection.
func DetectEdges(img image.Image, threshold int, targetW, targetH int) (*EdgeMap, error) {
	if threshold < 0 || threshold > 255 {
		return nil, fmt.Errorf("threshold must be between 0 and 255, got %d", threshold)
	}

	// Downsample image to target dimensions if specified
	processImg := img
	if targetW > 0 && targetH > 0 {
		processImg = resize.Resize(uint(targetW), uint(targetH), img, resize.Bilinear)
	}

	bounds := processImg.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Convert to grayscale
	gray := image.NewGray(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray.Set(x, y, processImg.At(x, y))
		}
	}

	// Apply Sobel operator
	edgeMap := make([][]int, height)
	gradientMagnitudes := make([][]float64, height)

	for y := 0; y < height; y++ {
		edgeMap[y] = make([]int, width)
		gradientMagnitudes[y] = make([]float64, width)
	}

	// Sobel kernels
	// Gx detects vertical edges (horizontal gradient)
	// Gy detects horizontal edges (vertical gradient)
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			// Get 3x3 neighborhood
			nw := int(gray.GrayAt(x-1, y-1).Y)
			n := int(gray.GrayAt(x, y-1).Y)
			ne := int(gray.GrayAt(x+1, y-1).Y)
			w := int(gray.GrayAt(x-1, y).Y)
			e := int(gray.GrayAt(x+1, y).Y)
			sw := int(gray.GrayAt(x-1, y+1).Y)
			s := int(gray.GrayAt(x, y+1).Y)
			se := int(gray.GrayAt(x+1, y+1).Y)

			// Compute Gx (horizontal gradient)
			gx := -nw + ne - 2*w + 2*e - sw + se

			// Compute Gy (vertical gradient)
			gy := -nw - 2*n - ne + sw + 2*s + se

			// Gradient magnitude
			magnitude := math.Sqrt(float64(gx*gx + gy*gy))
			gradientMagnitudes[y][x] = magnitude

			// Threshold to create binary edge map
			if magnitude > float64(threshold) {
				edgeMap[y][x] = 1
			} else {
				edgeMap[y][x] = 0
			}
		}
	}

	// Find major edges (contours with high strength)
	majorEdges := findMajorEdges(edgeMap, gradientMagnitudes, width, height)

	return &EdgeMap{
		Grid:       edgeMap,
		MajorEdges: majorEdges,
	}, nil
}

// findMajorEdges identifies significant edge contours from the edge map.
// Returns simplified edge lines representing major features.
func findMajorEdges(edgeMap [][]int, magnitudes [][]float64, width, height int) []EdgeLine {
	majorEdges := make([]EdgeLine, 0)

	// Simple approach: scan for continuous edge segments
	// This is a basic implementation - could be enhanced with line detection algorithms
	minLength := 5 // Minimum edge length to be considered "major"

	// Scan horizontally
	for y := 1; y < height-1; y++ {
		startX := -1
		for x := 1; x < width-1; x++ {
			if edgeMap[y][x] == 1 {
				if startX == -1 {
					startX = x
				}
			} else {
				if startX != -1 && x-startX >= minLength {
					// Calculate average strength
					strength := 0.0
					for i := startX; i < x; i++ {
						strength += magnitudes[y][i]
					}
					strength = strength / float64(x-startX)
					strength = (strength / 255.0) * 100.0 // Normalize to 0-100

					majorEdges = append(majorEdges, EdgeLine{
						From:     Point{X: startX, Y: y},
						To:       Point{X: x - 1, Y: y},
						Strength: strength,
					})
				}
				startX = -1
			}
		}
	}

	// Scan vertically
	for x := 1; x < width-1; x++ {
		startY := -1
		for y := 1; y < height-1; y++ {
			if edgeMap[y][x] == 1 {
				if startY == -1 {
					startY = y
				}
			} else {
				if startY != -1 && y-startY >= minLength {
					// Calculate average strength
					strength := 0.0
					for i := startY; i < y; i++ {
						strength += magnitudes[i][x]
					}
					strength = strength / float64(y-startY)
					strength = (strength / 255.0) * 100.0 // Normalize to 0-100

					majorEdges = append(majorEdges, EdgeLine{
						From:     Point{X: x, Y: startY},
						To:       Point{X: x, Y: y - 1},
						Strength: strength,
					})
				}
				startY = -1
			}
		}
	}

	return majorEdges
}

// AnalyzeComposition performs basic composition analysis on an image.
type Composition struct {
	FocalPoints    []FocalPoint      `json:"focal_points"`
	RuleOfThirds   RuleOfThirdsGuide `json:"rule_of_thirds"`
	DominantRegion Region            `json:"dominant_region"`
}

// FocalPoint represents a point of visual interest.
type FocalPoint struct {
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Weight float64 `json:"weight"` // 0.0-1.0, higher = more prominent
}

// RuleOfThirdsGuide represents the rule of thirds grid.
type RuleOfThirdsGuide struct {
	VerticalLines   []int          `json:"vertical_lines"`   // X coordinates at 1/3 and 2/3
	HorizontalLines []int          `json:"horizontal_lines"` // Y coordinates at 1/3 and 2/3
	Intersections   []Intersection `json:"intersections"`    // Four power points
}

// Intersection represents a rule-of-thirds intersection point.
type Intersection struct {
	X             int  `json:"x"`
	Y             int  `json:"y"`
	HasFocalPoint bool `json:"has_focal_point"`
}

// Region represents a rectangular area.
type Region struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// AnalyzeComposition analyzes image composition and returns guide data.
func AnalyzeComposition(img image.Image, edgeMap *EdgeMap) (*Composition, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Calculate rule of thirds lines
	verticalLines := []int{
		width / 3,
		(width * 2) / 3,
	}
	horizontalLines := []int{
		height / 3,
		(height * 2) / 3,
	}

	// Calculate intersections
	intersections := make([]Intersection, 0, 4)
	for _, vl := range verticalLines {
		for _, hl := range horizontalLines {
			intersections = append(intersections, Intersection{
				X:             vl,
				Y:             hl,
				HasFocalPoint: false,
			})
		}
	}

	// Find focal points using edge density
	focalPoints := findFocalPoints(edgeMap, width, height)

	// Mark intersections near focal points
	for i := range intersections {
		for _, fp := range focalPoints {
			dist := math.Sqrt(math.Pow(float64(intersections[i].X-fp.X), 2) +
				math.Pow(float64(intersections[i].Y-fp.Y), 2))
			if dist < float64(width)/6 { // Within 1/6 of image width
				intersections[i].HasFocalPoint = true
				break
			}
		}
	}

	return &Composition{
		FocalPoints: focalPoints,
		RuleOfThirds: RuleOfThirdsGuide{
			VerticalLines:   verticalLines,
			HorizontalLines: horizontalLines,
			Intersections:   intersections,
		},
		DominantRegion: Region{
			X:      width / 4,
			Y:      height / 4,
			Width:  width / 2,
			Height: height / 2,
		},
	}, nil
}

// findFocalPoints identifies areas of high visual interest based on edge density.
func findFocalPoints(edgeMap *EdgeMap, width, height int) []FocalPoint {
	focalPoints := make([]FocalPoint, 0)

	// Divide image into grid and calculate edge density
	gridSize := 16 // 16x16 grid
	cellWidth := width / gridSize
	cellHeight := height / gridSize

	type cell struct {
		x, y    int
		density float64
	}

	cells := make([]cell, 0)

	for gy := 0; gy < gridSize; gy++ {
		for gx := 0; gx < gridSize; gx++ {
			edgeCount := 0
			totalPixels := 0

			// Count edges in this cell
			for y := gy * cellHeight; y < (gy+1)*cellHeight && y < height; y++ {
				for x := gx * cellWidth; x < (gx+1)*cellWidth && x < width; x++ {
					if y < len(edgeMap.Grid) && x < len(edgeMap.Grid[y]) {
						if edgeMap.Grid[y][x] == 1 {
							edgeCount++
						}
						totalPixels++
					}
				}
			}

			if totalPixels > 0 {
				density := float64(edgeCount) / float64(totalPixels)
				cells = append(cells, cell{
					x:       gx*cellWidth + cellWidth/2,
					y:       gy*cellHeight + cellHeight/2,
					density: density,
				})
			}
		}
	}

	// Sort by density and take top focal points
	if len(cells) > 0 {
		// Find max density
		maxDensity := 0.0
		for _, c := range cells {
			if c.density > maxDensity {
				maxDensity = c.density
			}
		}

		// Take cells with >50% of max density as focal points
		for _, c := range cells {
			if c.density > maxDensity*0.5 {
				focalPoints = append(focalPoints, FocalPoint{
					X:      c.x,
					Y:      c.y,
					Weight: c.density / maxDensity,
				})
			}
		}

		// Limit to top 3 focal points
		if len(focalPoints) > 3 {
			// Sort by weight descending
			sortedFP := make([]FocalPoint, len(focalPoints))
			copy(sortedFP, focalPoints)
			for i := 0; i < len(sortedFP)-1; i++ {
				for j := i + 1; j < len(sortedFP); j++ {
					if sortedFP[j].Weight > sortedFP[i].Weight {
						sortedFP[i], sortedFP[j] = sortedFP[j], sortedFP[i]
					}
				}
			}
			focalPoints = sortedFP[:3]
		}
	}

	return focalPoints
}
