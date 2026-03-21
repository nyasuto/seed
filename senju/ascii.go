package senju

import (
	"fmt"
	"strings"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// elementChar returns a single character representing the element.
//
//	Wood: W, Fire: F, Earth: E, Metal: M, Water: A
func elementChar(e types.Element) byte {
	switch e {
	case types.Wood:
		return 'W'
	case types.Fire:
		return 'F'
	case types.Earth:
		return 'E'
	case types.Metal:
		return 'M'
	case types.Water:
		return 'A'
	default:
		return '?'
	}
}

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

	g := cave.Grid
	var sb strings.Builder

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			cell, _ := g.At(types.Pos{X: x, Y: y})
			switch cell.Type {
			case world.RoomFloor:
				if cell.RoomID > 0 {
					if info, ok := roomBeasts[cell.RoomID]; ok {
						sb.WriteString(beastTile(info))
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

// beastTile returns a 2-character tile for a room's beast overlay.
func beastTile(info *beastOverlay) string {
	ch := elementChar(info.element)
	if info.count == 1 {
		return string([]byte{ch, ch})
	}
	if info.count <= 9 {
		return fmt.Sprintf("%d%c", info.count, ch)
	}
	// 10+ beasts: show "9+" as a cap
	return "9+"
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
	default:
		return "[?]"
	}
}

// behaviorOverlay holds pre-computed per-room beast behavior display info.
type behaviorOverlay struct {
	count   int
	element types.Element
	state   BeastState
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
			roomInfo[b.RoomID] = &behaviorOverlay{count: 1, element: b.Element, state: b.State}
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

	g := cave.Grid
	var sb strings.Builder

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			cell, _ := g.At(types.Pos{X: x, Y: y})
			switch cell.Type {
			case world.RoomFloor:
				if cell.RoomID > 0 {
					if invaderRooms[cell.RoomID] {
						sb.WriteString("??")
					} else if info, ok := roomInfo[cell.RoomID]; ok {
						sb.WriteString(behaviorTile(info))
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

// behaviorTile returns a 2-character tile based on the beast's state.
// For single beasts the state tag's inner character is doubled (e.g. "GG").
// For multiple beasts the count + state char (e.g. "2G").
func behaviorTile(info *behaviorOverlay) string {
	ch := stateChar(info.state)
	if info.count == 1 {
		return string([]byte{ch, ch})
	}
	if info.count <= 9 {
		return fmt.Sprintf("%d%c", info.count, ch)
	}
	return "9+"
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
	default:
		return '?'
	}
}

// roomIDChar returns a display character for a room ID.
// 1-9 → '1'-'9', 10-35 → 'A'-'Z'.
func roomIDChar(id int) byte {
	if id >= 1 && id <= 9 {
		return byte('0' + id)
	}
	return byte('A' + id - 10)
}
