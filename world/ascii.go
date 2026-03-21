package world

import (
	"fmt"
	"strings"

	"github.com/ponpoko/chaosseed-core/types"
)

// RoomIDChar returns a display character for a room ID.
// 1-9 → '1'-'9', 10-35 → 'A'-'Z'.
func RoomIDChar(id int) byte {
	if id >= 1 && id <= 9 {
		return byte('0' + id)
	}
	return byte('A' + id - 10)
}

// CountTile returns a 2-character tile for a count and display character.
// If count is 1, the character is doubled (e.g. "WW").
// If count is 2-9, the count is prefixed (e.g. "3W").
// If count is 10+, "9+" is returned.
func CountTile(count int, ch byte) string {
	if count == 1 {
		return string([]byte{ch, ch})
	}
	if count <= 9 {
		return fmt.Sprintf("%d%c", count, ch)
	}
	return "9+"
}

// DefaultTile returns the standard 2-character tile for a cell.
//
//	Rock: "██", CorridorFloor: "..", Entrance: "><",
//	RoomFloor with ID: doubled RoomIDChar, RoomFloor without ID: "[]",
//	default: "??"
func DefaultTile(cell Cell) string {
	switch cell.Type {
	case Rock:
		return "██"
	case CorridorFloor:
		return ".."
	case RoomFloor:
		if cell.RoomID > 0 {
			ch := RoomIDChar(cell.RoomID)
			return string([]byte{ch, ch})
		}
		return "[]"
	case Entrance:
		return "><"
	default:
		return "??"
	}
}

// TileFunc is called for each cell during grid rendering.
// It receives the position and cell, and returns a 2-character tile string.
// If it returns "", the DefaultTile is used.
type TileFunc func(pos types.Pos, cell Cell) string

// RenderGrid iterates over the grid and builds an ASCII string.
// For each cell, it calls tileFn if non-nil. If tileFn returns "",
// DefaultTile is used as a fallback.
func RenderGrid(g *Grid, tileFn TileFunc) string {
	var sb strings.Builder
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			pos := types.Pos{X: x, Y: y}
			cell, _ := g.At(pos)
			tile := ""
			if tileFn != nil {
				tile = tileFn(pos, cell)
			}
			if tile == "" {
				tile = DefaultTile(cell)
			}
			sb.WriteString(tile)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
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
	return RenderGrid(c.Grid, nil)
}
