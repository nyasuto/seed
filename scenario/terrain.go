package scenario

import (
	"fmt"

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

// ApplyTerrain places the given terrain zones onto the cave grid.
// Each zone's cells are set to the zone's Type (HardRock or Water).
// Returns an error if any zone cell is out of bounds or already belongs
// to a room (RoomID != 0).
func ApplyTerrain(cave *world.Cave, zones []TerrainZone) error {
	for i, z := range zones {
		for dy := 0; dy < z.Height; dy++ {
			for dx := 0; dx < z.Width; dx++ {
				pos := types.Pos{X: z.Pos.X + dx, Y: z.Pos.Y + dy}
				if !cave.Grid.InBounds(pos) {
					return fmt.Errorf("terrain zone %d: position (%d,%d) out of bounds", i, pos.X, pos.Y)
				}
				existing, _ := cave.Grid.At(pos)
				if existing.RoomID != 0 {
					return fmt.Errorf("terrain zone %d: position (%d,%d) overlaps with room %d", i, pos.X, pos.Y, existing.RoomID)
				}
				if err := cave.Grid.Set(pos, world.Cell{Type: z.Type}); err != nil {
					return fmt.Errorf("terrain zone %d: %w", i, err)
				}
			}
		}
	}
	return nil
}
