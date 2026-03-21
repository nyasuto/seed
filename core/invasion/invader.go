package invasion

import (
	"github.com/nyasuto/seed/core/types"
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

// String returns the name of the invader state.
func (s InvaderState) String() string {
	switch s {
	case Advancing:
		return "Advancing"
	case Fighting:
		return "Fighting"
	case Retreating:
		return "Retreating"
	case Defeated:
		return "Defeated"
	case GoalAchieved:
		return "GoalAchieved"
	default:
		return "Unknown"
	}
}

// Invader represents a single invading entity inside the cave.
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

// scaleStatByLevel applies level scaling to a base stat.
// Stats grow by 10% per level above 1.
func scaleStatByLevel(base int, level int) int {
	if level <= 1 {
		return base
	}
	return base + base*(level-1)/10
}

// NewInvader creates a new Invader from an InvaderClass template with level scaling.
// Stats are calculated as: base + base*(level-1)/10 (integer arithmetic).
// The invader starts in the Advancing state with a fresh ExplorationMemory.
func NewInvader(id int, class InvaderClass, level int, goal Goal, entryRoomID int, tick types.Tick) *Invader {
	if level < 1 {
		level = 1
	}
	hp := scaleStatByLevel(class.BaseHP, level)
	return &Invader{
		ID:            id,
		ClassID:       class.ID,
		Name:          class.Name,
		Element:       class.Element,
		Level:         level,
		HP:            hp,
		MaxHP:         hp,
		ATK:           scaleStatByLevel(class.BaseATK, level),
		DEF:           scaleStatByLevel(class.BaseDEF, level),
		SPD:           scaleStatByLevel(class.BaseSPD, level),
		CurrentRoomID: entryRoomID,
		Goal:          goal,
		Memory: &ExplorationMemory{
			VisitedRooms:       make(map[int]types.Tick),
			KnownBeastRooms:    make(map[int]bool),
			KnownCoreRoom:      0,
			KnownTreasureRooms: nil,
		},
		State:     Advancing,
		SlowTicks: 0,
		EntryTick: tick,
		StayTicks: 0,
	}
}

// NewExplorationMemory creates a fresh ExplorationMemory with initialized maps.
func NewExplorationMemory() *ExplorationMemory {
	return &ExplorationMemory{
		VisitedRooms:       make(map[int]types.Tick),
		KnownBeastRooms:    make(map[int]bool),
		KnownCoreRoom:      0,
		KnownTreasureRooms: nil,
	}
}
