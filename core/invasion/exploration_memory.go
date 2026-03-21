package invasion

import (
	"slices"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// ExplorationMemory tracks an invader's knowledge of the cave layout.
// As invaders explore, they record visited rooms and discover special rooms
// (beast rooms, core room, treasure rooms) to inform pathfinding decisions.
type ExplorationMemory struct {
	// VisitedRooms maps room IDs to the tick when they were first visited.
	VisitedRooms map[int]types.Tick
	// KnownBeastRooms tracks room IDs known to contain beasts.
	KnownBeastRooms map[int]bool
	// KnownCoreRoom is the room ID of the known dragon vein core (0 if unknown).
	KnownCoreRoom int
	// KnownTreasureRooms tracks room IDs known to be storage rooms.
	KnownTreasureRooms []int
}

// Visit records that the invader has visited the given room at the given tick.
// Only the first visit tick is recorded; subsequent visits do not update the tick.
// Visit also inspects the room and cave to discover special rooms (core, treasure, beasts).
func (m *ExplorationMemory) Visit(roomID int, tick types.Tick, cave *world.Cave, rooms []*world.Room) {
	if _, visited := m.VisitedRooms[roomID]; !visited {
		m.VisitedRooms[roomID] = tick
	}

	// Discover room properties by inspecting the room.
	var room *world.Room
	for _, r := range rooms {
		if r.ID == roomID {
			room = r
			break
		}
	}
	if room == nil {
		return
	}

	// Discover core room (dragon_hole).
	if room.TypeID == "dragon_hole" && m.KnownCoreRoom == 0 {
		m.KnownCoreRoom = roomID
	}

	// Discover treasure room (storage).
	if room.TypeID == "storage" {
		if !m.hasTreasureRoom(roomID) {
			m.KnownTreasureRooms = append(m.KnownTreasureRooms, roomID)
		}
	}

	// Discover beast presence.
	m.KnownBeastRooms[roomID] = room.BeastCount() > 0
}

// HasVisited reports whether the invader has visited the given room.
func (m *ExplorationMemory) HasVisited(roomID int) bool {
	_, ok := m.VisitedRooms[roomID]
	return ok
}

// VisitedCount returns the number of rooms the invader has visited.
func (m *ExplorationMemory) VisitedCount() int {
	return len(m.VisitedRooms)
}

// RecordBeastRoom records that a room is known to contain (or not contain) beasts.
func (m *ExplorationMemory) RecordBeastRoom(roomID int, hasBeasts bool) {
	m.KnownBeastRooms[roomID] = hasBeasts
}

// RecordCoreRoom records the discovery of the core room (dragon_hole).
func (m *ExplorationMemory) RecordCoreRoom(roomID int) {
	m.KnownCoreRoom = roomID
}

// RecordTreasureRoom records the discovery of a treasure room (storage).
func (m *ExplorationMemory) RecordTreasureRoom(roomID int) {
	if !m.hasTreasureRoom(roomID) {
		m.KnownTreasureRooms = append(m.KnownTreasureRooms, roomID)
	}
}

// hasTreasureRoom checks if a room ID is already in KnownTreasureRooms.
func (m *ExplorationMemory) hasTreasureRoom(roomID int) bool {
	return slices.Contains(m.KnownTreasureRooms, roomID)
}
