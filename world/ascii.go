package world

import (
	"strings"

	"github.com/ponpoko/chaosseed-core/types"
)

// roomIDChar returns a display character for a room ID.
// 1-9 → '1'-'9', 10-35 → 'A'-'Z'.
func roomIDChar(id int) byte {
	if id >= 1 && id <= 9 {
		return byte('0' + id)
	}
	return byte('A' + id - 10)
}

// RenderASCII returns an ASCII art representation of the cave.
// Each cell is rendered as a 2-character wide tile:
//
//	██ = Rock
//	.. = Corridor
//	[] = Room floor (no RoomID)
//	11 = Room floor with RoomID displayed (1-9 as digits, 10+ as A,B,C...)
//	>< = Entrance
func (c *Cave) RenderASCII() string {
	g := c.Grid
	var sb strings.Builder

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			cell, _ := g.At(types.Pos{X: x, Y: y})
			switch cell.Type {
			case Rock:
				sb.WriteString("██")
			case CorridorFloor:
				sb.WriteString("..")
			case RoomFloor:
				if cell.RoomID > 0 {
					ch := roomIDChar(cell.RoomID)
					sb.WriteByte(ch)
					sb.WriteByte(ch)
				} else {
					sb.WriteString("[]")
				}
			case Entrance:
				sb.WriteString("><")
			default:
				sb.WriteString("??")
			}
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}
