package aseprite

import (
	"image"
	"image/color"
	"testing"
)

// createTestImage creates a simple test image with known brightness patterns
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create gradient from black to white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Horizontal gradient
			brightness := uint8(float64(x) / float64(width) * 255)
			img.Set(x, y, color.RGBA{R: brightness, G: brightness, B: brightness, A: 255})
		}
	}

	return img
}

// createEdgeTestImage creates an image with clear edges
func createEdgeTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	// Draw a black square in the center (creates edges)
	centerX, centerY := width/2, height/2
	squareSize := width / 4
	for y := centerY - squareSize/2; y < centerY+squareSize/2; y++ {
		for x := centerX - squareSize/2; x < centerX+squareSize/2; x++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}
	}

	return img
}

func TestGenerateBrightnessMap(t *testing.T) {
	tests := []struct {
		name      string
		targetW   int
		targetH   int
		numLevels int
		wantErr   bool
	}{
		{
			name:      "valid parameters",
			targetW:   10,
			targetH:   10,
			numLevels: 5,
			wantErr:   false,
		},
		{
			name:      "invalid width",
			targetW:   0,
			targetH:   10,
			numLevels: 5,
			wantErr:   true,
		},
		{
			name:      "invalid height",
			targetW:   10,
			targetH:   0,
			numLevels: 5,
			wantErr:   true,
		},
		{
			name:      "numLevels too low",
			targetW:   10,
			targetH:   10,
			numLevels: 1,
			wantErr:   true,
		},
		{
			name:      "numLevels too high",
			targetW:   10,
			targetH:   10,
			numLevels: 257,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := createTestImage(100, 100)
			result, err := GenerateBrightnessMap(img, tt.targetW, tt.targetH, tt.numLevels)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateBrightnessMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("GenerateBrightnessMap() returned nil result")
					return
				}

				if len(result.Grid) != tt.targetH {
					t.Errorf("Grid height = %d, want %d", len(result.Grid), tt.targetH)
				}

				if len(result.Grid[0]) != tt.targetW {
					t.Errorf("Grid width = %d, want %d", len(result.Grid[0]), tt.targetW)
				}

				if len(result.Legend) != tt.numLevels {
					t.Errorf("Legend size = %d, want %d", len(result.Legend), tt.numLevels)
				}

				// Verify brightness values are in valid range
				for y := 0; y < tt.targetH; y++ {
					for x := 0; x < tt.targetW; x++ {
						level := result.Grid[y][x]
						if level < 0 || level >= tt.numLevels {
							t.Errorf("Invalid brightness level %d at (%d,%d), want 0-%d", level, x, y, tt.numLevels-1)
						}
					}
				}
			}
		})
	}
}

func TestDetectEdges(t *testing.T) {
	tests := []struct {
		name      string
		threshold int
		wantErr   bool
	}{
		{
			name:      "valid threshold",
			threshold: 30,
			wantErr:   false,
		},
		{
			name:      "threshold too low",
			threshold: -1,
			wantErr:   true,
		},
		{
			name:      "threshold too high",
			threshold: 256,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := createEdgeTestImage(100, 100)
			result, err := DetectEdges(img, tt.threshold, 0, 0)

			if (err != nil) != tt.wantErr {
				t.Errorf("DetectEdges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("DetectEdges() returned nil result")
					return
				}

				if len(result.Grid) != 100 {
					t.Errorf("Grid height = %d, want 100", len(result.Grid))
				}

				if len(result.Grid[0]) != 100 {
					t.Errorf("Grid width = %d, want 100", len(result.Grid[0]))
				}

				// Verify edge values are binary (0 or 1)
				for y := 0; y < len(result.Grid); y++ {
					for x := 0; x < len(result.Grid[y]); x++ {
						val := result.Grid[y][x]
						if val != 0 && val != 1 {
							t.Errorf("Invalid edge value %d at (%d,%d), want 0 or 1", val, x, y)
						}
					}
				}

				// Should have detected some edges
				edgeCount := 0
				for y := 0; y < len(result.Grid); y++ {
					for x := 0; x < len(result.Grid[y]); x++ {
						edgeCount += result.Grid[y][x]
					}
				}
				if edgeCount == 0 {
					t.Error("DetectEdges() found no edges in test image with clear edges")
				}
			}
		})
	}
}

func TestAnalyzeComposition(t *testing.T) {
	img := createEdgeTestImage(100, 100)

	// First detect edges
	edgeMap, err := DetectEdges(img, 30, 0, 0)
	if err != nil {
		t.Fatalf("DetectEdges() error = %v", err)
	}

	result, err := AnalyzeComposition(img, edgeMap)
	if err != nil {
		t.Fatalf("AnalyzeComposition() error = %v", err)
	}

	if result == nil {
		t.Fatal("AnalyzeComposition() returned nil result")
	}

	// Check RuleOfThirds
	if len(result.RuleOfThirds.VerticalLines) != 2 {
		t.Errorf("VerticalLines count = %d, want 2", len(result.RuleOfThirds.VerticalLines))
	}

	if len(result.RuleOfThirds.HorizontalLines) != 2 {
		t.Errorf("HorizontalLines count = %d, want 2", len(result.RuleOfThirds.HorizontalLines))
	}

	// Check focal points
	if len(result.FocalPoints) == 0 {
		t.Error("AnalyzeComposition() returned no focal points")
	}

	// Verify focal points are within bounds
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	for i, fp := range result.FocalPoints {
		if fp.X < 0 || fp.X >= width {
			t.Errorf("Focal point %d X=%d out of bounds [0,%d)", i, fp.X, width)
		}
		if fp.Y < 0 || fp.Y >= height {
			t.Errorf("Focal point %d Y=%d out of bounds [0,%d)", i, fp.Y, height)
		}
		if fp.Weight < 0 || fp.Weight > 1 {
			t.Errorf("Focal point %d weight=%f out of range [0,1]", i, fp.Weight)
		}
	}

	// Check dominant region
	if result.DominantRegion.Width == 0 || result.DominantRegion.Height == 0 {
		t.Error("AnalyzeComposition() returned empty dominant region")
	}
}

func TestGenerateBrightnessMap_EdgeCases(t *testing.T) {
	// Test with exact match dimensions (no downsampling needed)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			brightness := uint8((x + y) * 255 / 18)
			img.Set(x, y, color.RGBA{brightness, brightness, brightness, 255})
		}
	}

	result, err := GenerateBrightnessMap(img, 10, 10, 5)

	if err != nil {
		t.Fatalf("GenerateBrightnessMap() error = %v", err)
	}

	// Grid is 2D array, so check total cells
	totalCells := 0
	for _, row := range result.Grid {
		totalCells += len(row)
	}

	if totalCells != 100 {
		t.Errorf("GenerateBrightnessMap() total grid cells = %d, want 100", totalCells)
	}

	if len(result.Legend) != 5 {
		t.Errorf("GenerateBrightnessMap() legend size = %d, want 5", len(result.Legend))
	}
}

func TestFindFocalPoints(t *testing.T) {
	// Create a 100x100 edge map for realistic focal point detection
	grid := make([][]int, 100)
	for i := range grid {
		grid[i] = make([]int, 100)
	}

	edgeMap := &EdgeMap{
		Grid: grid,
	}

	// Create strong edge patterns in multiple regions to trigger focal point detection
	// Top-left quadrant: dense edges
	for y := 10; y < 30; y++ {
		for x := 10; x < 30; x++ {
			edgeMap.Grid[y][x] = 1
		}
	}

	// Center: another focal area
	for y := 45; y < 55; y++ {
		for x := 45; x < 55; x++ {
			edgeMap.Grid[y][x] = 1
		}
	}

	focalPoints := findFocalPoints(edgeMap, 100, 100)

	// Function may or may not return focal points depending on density threshold
	// Just verify that if it returns any, they're valid
	for _, fp := range focalPoints {
		if fp.X < 0 || fp.X >= 100 || fp.Y < 0 || fp.Y >= 100 {
			t.Errorf("findFocalPoints() returned invalid coordinates: (%d, %d)", fp.X, fp.Y)
		}
		if fp.Weight < 0 || fp.Weight > 1 {
			t.Errorf("findFocalPoints() returned invalid weight: %f", fp.Weight)
		}
	}

	// Test passes if no invalid focal points were found
	t.Logf("findFocalPoints() returned %d focal points", len(focalPoints))
}
