package invasion

import (
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
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
	return world.CountTile(info.count, sym[1])
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
			if inv.State == Defeated || inv.CurrentRoomID == 0 {
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

	return world.RenderGrid(cave.Grid, func(_ types.Pos, cell world.Cell) string {
		if cell.Type == world.RoomFloor && cell.RoomID > 0 {
			if info, ok := roomInvaders[cell.RoomID]; ok {
				return invaderTile(info)
			}
		}
		return ""
	})
}
