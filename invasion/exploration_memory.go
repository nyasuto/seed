package invasion

import (
	"github.com/ponpoko/chaosseed-core/types"
)

// ExplorationMemory tracks an invader's knowledge of the cave layout.
// This is a minimal definition that will be expanded in a later task.
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
