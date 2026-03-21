package senju

import "github.com/ponpoko/chaosseed-core/types"

// BeastState represents the current behavioral state of a beast.
type BeastState int

const (
	// Idle means the beast is resting and not performing any action.
	Idle BeastState = iota
	// Patrolling means the beast is moving around its assigned room.
	Patrolling
	// Chasing means the beast is pursuing an intruder.
	Chasing
	// Fighting means the beast is engaged in combat.
	Fighting
	// Recovering means the beast is healing after combat.
	Recovering
	// Stunned means the beast has been defeated and is temporarily incapacitated.
	// While stunned, the beast cannot act or participate in combat.
	Stunned
)

// String returns the name of the beast state.
func (s BeastState) String() string {
	switch s {
	case Idle:
		return "Idle"
	case Patrolling:
		return "Patrolling"
	case Chasing:
		return "Chasing"
	case Fighting:
		return "Fighting"
	case Recovering:
		return "Recovering"
	case Stunned:
		return "Stunned"
	default:
		return "Unknown"
	}
}

// Beast represents an individual beast instance placed in the cave.
type Beast struct {
	// ID is a unique identifier for this beast instance.
	ID int
	// SpeciesID is the ID of the species this beast belongs to.
	SpeciesID string
	// Name is the display name of this beast.
	Name string
	// Element is the elemental affinity, inherited from the species.
	Element types.Element
	// RoomID is the ID of the room this beast is currently placed in.
	// 0 means unassigned.
	RoomID int
	// Level is the current level of the beast (starts at 1).
	Level int
	// EXP is the current experience points toward the next level.
	EXP int
	// HP is the current hit points.
	HP int
	// MaxHP is the maximum hit points.
	MaxHP int
	// ATK is the base attack power.
	ATK int
	// DEF is the base defense power.
	DEF int
	// SPD is the base speed.
	SPD int
	// BornTick is the tick at which this beast was created.
	BornTick types.Tick
	// State is the current behavioral state.
	State BeastState
}

// NewBeast creates a new beast from the given species, initializing
// base stats from the species definition.
func NewBeast(id int, species *Species, tick types.Tick) *Beast {
	return &Beast{
		ID:        id,
		SpeciesID: species.ID,
		Name:      species.Name,
		Element:   species.Element,
		RoomID:    0,
		Level:     1,
		EXP:       0,
		HP:        species.BaseHP,
		MaxHP:     species.BaseHP,
		ATK:       species.BaseATK,
		DEF:       species.BaseDEF,
		SPD:       species.BaseSPD,
		BornTick:  tick,
		State:     Idle,
	}
}
