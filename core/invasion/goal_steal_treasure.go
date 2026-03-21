package invasion

import (
	"github.com/nyasuto/seed/core/world"
)

// StealTreasureGoal directs an invader to reach a storage room (倉庫) and
// steal chi. The goal is achieved as soon as the invader arrives at a storage
// room, at which point it retreats with the plundered chi.
type StealTreasureGoal struct{}

// NewStealTreasureGoal creates a StealTreasureGoal.
func NewStealTreasureGoal() *StealTreasureGoal {
	return &StealTreasureGoal{}
}

// Type returns StealTreasure.
func (g *StealTreasureGoal) Type() GoalType {
	return StealTreasure
}

// TargetRoomID returns the ID of the nearest known storage room.
// If the invader's memory contains known treasure rooms, the closest one is
// returned. Otherwise, all rooms in the cave are scanned for the "storage"
// type. Returns 0 if no storage room exists.
func (g *StealTreasureGoal) TargetRoomID(cave *world.Cave, invader *Invader, memory *ExplorationMemory) int {
	// First check memory for known treasure rooms.
	if memory != nil && len(memory.KnownTreasureRooms) > 0 {
		bestID := 0
		bestDist := -1
		for _, roomID := range memory.KnownTreasureRooms {
			dist := roomDistance(cave, invader.CurrentRoomID, roomID)
			if bestID == 0 || dist < bestDist {
				bestID = roomID
				bestDist = dist
			}
		}
		if bestID != 0 {
			return bestID
		}
	}

	// Fallback: scan cave rooms for any storage room.
	bestID := 0
	bestDist := -1
	for _, room := range cave.Rooms {
		if room.TypeID == "storage" && room.ID != invader.CurrentRoomID {
			dist := roomDistance(cave, invader.CurrentRoomID, room.ID)
			if bestID == 0 || dist < bestDist {
				bestID = room.ID
				bestDist = dist
			}
		}
	}
	return bestID
}

// IsAchieved returns true when the invader is currently in a storage room.
func (g *StealTreasureGoal) IsAchieved(cave *world.Cave, invader *Invader) bool {
	for _, room := range cave.Rooms {
		if room.TypeID == "storage" && invader.CurrentRoomID == room.ID {
			return true
		}
	}
	return false
}
