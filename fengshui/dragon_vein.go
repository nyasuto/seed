package fengshui

import (
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// DragonVein represents a path through the cave that carries chi energy
// from an entrance into the interior. Each dragon vein has an elemental
// affinity and a flow rate that determines how much chi it supplies per tick.
type DragonVein struct {
	// ID is a unique identifier for this dragon vein.
	ID int
	// SourcePos is the grid position where this dragon vein originates
	// (typically a cave entrance).
	SourcePos types.Pos
	// Element is the elemental affinity of this dragon vein.
	Element types.Element
	// FlowRate is the base amount of chi supplied per tick to rooms
	// along this vein's path.
	FlowRate float64
	// Path is the ordered sequence of grid positions that this dragon
	// vein passes through, starting from SourcePos.
	Path []types.Pos
}

// RoomsOnPath returns the IDs of all rooms that lie on this dragon vein's path.
// Each room ID appears at most once, in the order first encountered along the path.
func (dv *DragonVein) RoomsOnPath(cave *world.Cave) []int {
	seen := make(map[int]bool)
	var result []int
	for _, pos := range dv.Path {
		cell, err := cave.Grid.At(pos)
		if err != nil {
			continue
		}
		if cell.RoomID != 0 && !seen[cell.RoomID] {
			seen[cell.RoomID] = true
			result = append(result, cell.RoomID)
		}
	}
	return result
}
