package invasion

import (
	"github.com/ponpoko/chaosseed-core/types"
)

// InvaderState represents the current behavior state of an invader.
type InvaderState int

const (
	// Advancing means the invader is moving toward its goal.
	Advancing InvaderState = iota
	// Fighting means the invader is engaged in combat.
	Fighting
	// Retreating means the invader is fleeing toward the exit.
	Retreating
	// Defeated means the invader has been eliminated.
	Defeated
	// GoalAchieved means the invader completed its objective.
	GoalAchieved
)

// Invader represents a single invading entity inside the cave.
// This is a minimal definition that will be expanded in a later task.
type Invader struct {
	// ID is the unique identifier for this invader.
	ID int
	// ClassID references the InvaderClass template.
	ClassID string
	// Name is the display name.
	Name string
	// Element is the five-element attribute.
	Element types.Element
	// Level is the invader's current level.
	Level int
	// HP is the current hit points.
	HP int
	// MaxHP is the maximum hit points.
	MaxHP int
	// ATK is the attack power.
	ATK int
	// DEF is the defense power.
	DEF int
	// SPD is the speed stat.
	SPD int
	// CurrentRoomID is the room the invader is currently in.
	CurrentRoomID int
	// Goal is the invader's objective.
	Goal Goal
	// Memory tracks the invader's exploration knowledge.
	Memory *ExplorationMemory
	// State is the current behavior state.
	State InvaderState
	// SlowTicks counts ticks where the invader is slowed.
	SlowTicks int
	// EntryTick is the tick when the invader entered the cave.
	EntryTick types.Tick
	// StayTicks counts how many ticks the invader has stayed in the current room.
	StayTicks int
}
