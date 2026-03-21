package senju

import (
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// beastOverlay holds pre-computed per-room beast display info.
type beastOverlay struct {
	count   int
	element types.Element
}

// RenderBeastOverlay returns an ASCII representation of the cave with beast
// placements overlaid on room cells.
//
// For rooms containing exactly one beast, the room cells show the beast's
// element character doubled (e.g. "WW" for Wood). For rooms with two or more
// beasts, cells show the count followed by the element character of the first
// beast (e.g. "2F"). Rooms with no beasts are rendered with the standard room
// ID display. Other cell types follow Cave.RenderASCII() conventions.
func RenderBeastOverlay(cave *world.Cave, beasts []*Beast) string {
	// Build per-room beast info.
	roomBeasts := make(map[int]*beastOverlay)
	for _, b := range beasts {
		if b.RoomID == 0 {
			continue
		}
		info, ok := roomBeasts[b.RoomID]
		if !ok {
			roomBeasts[b.RoomID] = &beastOverlay{count: 1, element: b.Element}
		} else {
			info.count++
		}
	}

	return world.RenderGrid(cave.Grid, func(_ types.Pos, cell world.Cell) string {
		if cell.Type == world.RoomFloor && cell.RoomID > 0 {
			if info, ok := roomBeasts[cell.RoomID]; ok {
				return world.CountTile(info.count, info.element.Char())
			}
		}
		return ""
	})
}

// stateTag returns a short display string representing the beast's behavior state.
//
//	Guard: [G], Patrol: [P], Chase: [!], Flee: [←], Recovering: [+], other: [?]
func stateTag(state BeastState) string {
	switch state {
	case Idle:
		return "[G]"
	case Patrolling:
		return "[P]"
	case Chasing:
		return "[!]"
	case Fighting:
		return "[!]"
	case Recovering:
		return "[+]"
	case Stunned:
		return "[X]"
	default:
		return "[?]"
	}
}

// behaviorOverlay holds pre-computed per-room beast behavior display info.
type behaviorOverlay struct {
	count int
	state BeastState
}

// RenderBehaviorOverlay returns an ASCII representation of the cave with beast
// behavior states overlaid. Each room with beasts shows the state tag of the
// first beast (e.g. "[G]" for Guard). Invader positions are shown as "??" as a
// placeholder for Phase 4.
func RenderBehaviorOverlay(cave *world.Cave, beasts []*Beast, invaderPositions map[int][]int) string {
	// Build per-room beast info.
	roomInfo := make(map[int]*behaviorOverlay)
	for _, b := range beasts {
		if b.RoomID == 0 {
			continue
		}
		info, ok := roomInfo[b.RoomID]
		if !ok {
			roomInfo[b.RoomID] = &behaviorOverlay{count: 1, state: b.State}
		} else {
			info.count++
		}
	}

	// Build invader room set.
	invaderRooms := make(map[int]bool)
	for roomID, ids := range invaderPositions {
		if len(ids) > 0 {
			invaderRooms[roomID] = true
		}
	}

	return world.RenderGrid(cave.Grid, func(_ types.Pos, cell world.Cell) string {
		if cell.Type == world.RoomFloor && cell.RoomID > 0 {
			if invaderRooms[cell.RoomID] {
				return "??"
			}
			if info, ok := roomInfo[cell.RoomID]; ok {
				return world.CountTile(info.count, stateChar(info.state))
			}
		}
		return ""
	})
}

// stateChar returns a single character for a beast state display in tiles.
func stateChar(state BeastState) byte {
	switch state {
	case Idle:
		return 'G'
	case Patrolling:
		return 'P'
	case Chasing, Fighting:
		return '!'
	case Recovering:
		return '+'
	case Stunned:
		return 'X'
	default:
		return '?'
	}
}
