package scenario

import "github.com/nyasuto/seed/core/types"

// GameConstraints defines gameplay limits for a scenario.
// These constraints are enforced by the simulation to prevent
// exceeding design boundaries (e.g. maximum number of rooms or beasts).
type GameConstraints struct {
	// MaxRooms is the maximum number of rooms allowed in the cave.
	// A value of 0 means no limit.
	MaxRooms int
	// MaxBeasts is the maximum number of beasts allowed simultaneously.
	// A value of 0 means no limit.
	MaxBeasts int
	// MaxTicks is the maximum duration of the scenario in ticks.
	// A value of 0 means no time limit.
	MaxTicks types.Tick
	// ForbiddenRoomTypes lists room type IDs that cannot be built.
	ForbiddenRoomTypes []string
}
