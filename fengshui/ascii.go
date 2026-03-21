package fengshui

import (
	"strings"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
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
	g := cave.Grid

	// Build a set of dragon vein path positions for quick lookup.
	veinPaths := make(map[types.Pos]bool)
	for _, vein := range engine.Veins {
		for _, pos := range vein.Path {
			veinPaths[pos] = true
		}
	}

	var sb strings.Builder

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			pos := types.Pos{X: x, Y: y}
			cell, _ := g.At(pos)

			switch cell.Type {
			case world.RoomFloor:
				if cell.RoomID > 0 {
					sb.WriteString(chiTile(engine, cell.RoomID))
				} else {
					sb.WriteString("[]")
				}
			case world.Entrance:
				sb.WriteString("><")
			case world.CorridorFloor:
				if veinPaths[pos] {
					sb.WriteString("~~")
				} else {
					sb.WriteString("..")
				}
			case world.Rock:
				if veinPaths[pos] {
					sb.WriteString("~~")
				} else {
					sb.WriteString("██")
				}
			default:
				sb.WriteString("??")
			}
		}
		sb.WriteByte('\n')
	}

	return sb.String()
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
