package invasion

import (
	"fmt"
	"strings"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// invaderOverlay holds pre-computed per-room invader display info.
type invaderOverlay struct {
	count int
	state InvaderState
}

// stateSymbol returns a 2-character symbol for an invader state.
//
//	Advancing: >>, Fighting: XX, Retreating: <<, GoalAchieved: $$
func stateSymbol(state InvaderState) string {
	switch state {
	case Advancing:
		return ">>"
	case Fighting:
		return "XX"
	case Retreating:
		return "<<"
	case GoalAchieved:
		return "$$"
	default:
		return "??"
	}
}

// invaderTile returns a 2-character tile for a room's invader overlay.
// For a single invader, the state symbol is used directly.
// For multiple invaders, the count is prefixed to the state's second character
// (e.g. "3>" for 3 advancing invaders). Counts above 9 are capped at "9+".
func invaderTile(info *invaderOverlay) string {
	if info.count == 1 {
		return stateSymbol(info.state)
	}
	sym := stateSymbol(info.state)
	if info.count <= 9 {
		return fmt.Sprintf("%d%c", info.count, sym[1])
	}
	return "9+"
}

// RenderInvasionOverlay returns an ASCII representation of the cave with
// invader positions overlaid on room cells.
//
// Invader states are rendered as:
//
//	>>  = Advancing (moving toward goal)
//	XX  = Fighting (engaged in combat)
//	<<  = Retreating (fleeing toward exit)
//	$$  = GoalAchieved (completed objective)
//
// Multiple invaders in the same room show count + state character (e.g. "3>>").
// Defeated invaders are excluded from the overlay.
// Rooms with no invaders display standard room IDs.
// Other cell types follow Cave.RenderASCII() conventions.
func RenderInvasionOverlay(cave *world.Cave, waves []*InvasionWave) string {
	// Build per-room invader info from all active waves.
	roomInvaders := make(map[int]*invaderOverlay)
	for _, w := range waves {
		for _, inv := range w.Invaders {
			if inv.State == Defeated {
				continue
			}
			if inv.CurrentRoomID == 0 {
				continue
			}
			info, ok := roomInvaders[inv.CurrentRoomID]
			if !ok {
				roomInvaders[inv.CurrentRoomID] = &invaderOverlay{count: 1, state: inv.State}
			} else {
				info.count++
			}
		}
	}

	g := cave.Grid
	var sb strings.Builder

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			cell, _ := g.At(types.Pos{X: x, Y: y})
			switch cell.Type {
			case world.RoomFloor:
				if cell.RoomID > 0 {
					if info, ok := roomInvaders[cell.RoomID]; ok {
						sb.WriteString(invaderTile(info))
					} else {
						ch := roomIDChar(cell.RoomID)
						sb.WriteByte(ch)
						sb.WriteByte(ch)
					}
				} else {
					sb.WriteString("[]")
				}
			case world.Entrance:
				sb.WriteString("><")
			case world.CorridorFloor:
				sb.WriteString("..")
			case world.Rock:
				sb.WriteString("██")
			default:
				sb.WriteString("??")
			}
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}

// roomIDChar returns a display character for a room ID.
// 1-9 → '1'-'9', 10-35 → 'A'-'Z'.
func roomIDChar(id int) byte {
	if id >= 1 && id <= 9 {
		return byte('0' + id)
	}
	return byte('A' + id - 10)
}
