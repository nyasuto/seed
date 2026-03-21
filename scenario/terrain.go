package scenario

import (
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// TerrainZone describes a rectangular zone of impassable terrain.
type TerrainZone struct {
	// Pos is the top-left grid position of the zone.
	Pos types.Pos
	// Width is the horizontal extent of the zone in cells.
	Width int
	// Height is the vertical extent of the zone in cells.
	Height int
	// Type is the cell type for the zone (HardRock or Water).
	Type world.CellType
}

// TerrainGenerator generates random impassable terrain zones for a cave.
type TerrainGenerator struct{}

// GenerateTerrain produces a slice of TerrainZone representing impassable
// regions (HardRock or Water) scattered across the given area.
// density controls how much of the area is covered (0.0–1.0).
// The result is deterministic for a given RNG state.
func (tg *TerrainGenerator) GenerateTerrain(width, height int, density float64, rng types.RNG) []TerrainZone {
	if width <= 0 || height <= 0 || density <= 0 {
		return nil
	}
	if density > 1.0 {
		density = 1.0
	}

	totalArea := width * height
	targetArea := int(float64(totalArea) * density)
	coveredArea := 0

	var zones []TerrainZone

	for coveredArea < targetArea {
		// Random zone dimensions: 1–3 cells wide/tall
		zw := 1 + rng.Intn(3)
		zh := 1 + rng.Intn(3)

		// Random position within grid bounds
		maxX := width - zw
		maxY := height - zh
		if maxX < 0 || maxY < 0 {
			// Zone too large for the grid; try a 1x1 fallback
			zw = 1
			zh = 1
			maxX = width - 1
			maxY = height - 1
		}

		x := rng.Intn(maxX + 1)
		y := rng.Intn(maxY + 1)

		// Randomly choose HardRock or Water
		cellType := world.HardRock
		if rng.Intn(2) == 0 {
			cellType = world.Water
		}

		zones = append(zones, TerrainZone{
			Pos:    types.Pos{X: x, Y: y},
			Width:  zw,
			Height: zh,
			Type:   cellType,
		})

		coveredArea += zw * zh
	}

	return zones
}
