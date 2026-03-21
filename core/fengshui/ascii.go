package fengshui

import (
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// RenderChiOverlay returns an ASCII representation of the cave with chi fill
// levels overlaid on room cells and dragon vein paths marked.
//
// Room cell display varies by chi fill ratio:
//
//	0%:     __
//	1-33%:  ░░
//	34-66%: ▒▒
//	67-99%: ▓▓
//	100%:   ██
//
// Dragon vein paths on non-room cells are shown as ~~.
// Other cells follow the same rendering as Cave.RenderASCII():
//
//	██ = Rock, .. = Corridor, >< = Entrance
func RenderChiOverlay(cave *world.Cave, engine *ChiFlowEngine) string {
	// Build a set of dragon vein path positions for quick lookup.
	veinPaths := make(map[types.Pos]bool)
	for _, vein := range engine.Veins {
		for _, pos := range vein.Path {
			veinPaths[pos] = true
		}
	}

	return world.RenderGrid(cave.Grid, func(pos types.Pos, cell world.Cell) string {
		switch cell.Type {
		case world.RoomFloor:
			if cell.RoomID > 0 {
				return chiTile(engine, cell.RoomID)
			}
		case world.CorridorFloor, world.Rock:
			if veinPaths[pos] {
				return "~~"
			}
		}
		return ""
	})
}

// chiTile returns a 2-character tile representing the chi fill ratio for a room.
func chiTile(engine *ChiFlowEngine, roomID int) string {
	rc, ok := engine.RoomChi[roomID]
	if !ok {
		return "__"
	}
	ratio := rc.Ratio()
	switch {
	case ratio <= 0:
		return "__"
	case ratio < 0.34:
		return "░░"
	case ratio < 0.67:
		return "▒▒"
	case ratio < 1.0:
		return "▓▓"
	default:
		return "██"
	}
}
