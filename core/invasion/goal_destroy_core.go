package invasion

import (
	"github.com/nyasuto/seed/core/world"
)

// DefaultDestroyCoreTicks is the number of ticks an invader must stay in the
// core room to destroy it.
const DefaultDestroyCoreTicks = 5

// DestroyCoreGoal directs an invader to reach the dragon vein core room
// (龍穴) and stay there for a set number of ticks to destroy it.
type DestroyCoreGoal struct {
	// RequiredStayTicks is how many ticks the invader must remain in the
	// core room to achieve the goal.
	RequiredStayTicks int
}

// NewDestroyCoreGoal creates a DestroyCoreGoal with the default stay duration.
func NewDestroyCoreGoal() *DestroyCoreGoal {
	return &DestroyCoreGoal{
		RequiredStayTicks: DefaultDestroyCoreTicks,
	}
}

// Type returns DestroyCore.
func (g *DestroyCoreGoal) Type() GoalType {
	return DestroyCore
}

// TargetRoomID returns the ID of the dragon vein core room.
// If the invader already knows where the core room is (via memory), that room
// ID is returned directly. Otherwise, all rooms in the cave are scanned for
// the "dragon_hole" type. Returns 0 if no core room exists.
func (g *DestroyCoreGoal) TargetRoomID(cave *world.Cave, invader *Invader, memory *ExplorationMemory) int {
	if memory != nil && memory.KnownCoreRoom != 0 {
		return memory.KnownCoreRoom
	}

	// Scan the cave for the core room.
	for _, room := range cave.Rooms {
		if room.TypeID == "dragon_hole" {
			return room.ID
		}
	}
	return 0
}

// IsAchieved returns true when the invader is in the core room and has stayed
// there for at least RequiredStayTicks ticks.
func (g *DestroyCoreGoal) IsAchieved(cave *world.Cave, invader *Invader) bool {
	for _, room := range cave.Rooms {
		if room.TypeID == "dragon_hole" && invader.CurrentRoomID == room.ID {
			return invader.StayTicks >= g.RequiredStayTicks
		}
	}
	return false
}
