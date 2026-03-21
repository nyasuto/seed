package invasion

import (
	"github.com/nyasuto/seed/core/world"
)

// Goal defines the interface for invader objectives.
// Each goal type determines where an invader wants to go and when the objective
// is considered achieved.
type Goal interface {
	// Type returns the GoalType classification of this goal.
	Type() GoalType

	// TargetRoomID returns the room ID that the invader should move toward.
	// It uses the cave layout, invader state, and exploration memory to determine
	// the best target. Returns 0 if no valid target is known.
	TargetRoomID(cave *world.Cave, invader *Invader, memory *ExplorationMemory) int

	// IsAchieved returns true if the invader has accomplished this goal.
	IsAchieved(cave *world.Cave, invader *Invader) bool
}
