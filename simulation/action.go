package simulation

import (
	"github.com/ponpoko/chaosseed-core/types"
)

// PlayerAction is the interface that all player actions must implement.
// Each tick, a player (or AI) may submit one or more actions to the
// simulation engine for validation and execution.
type PlayerAction interface {
	// ActionType returns a string identifier for this action type.
	ActionType() string
}

// ActionResult captures the outcome of applying a player action.
type ActionResult struct {
	// Success indicates whether the action was executed successfully.
	Success bool
	// Cost is the chi cost that was spent (0 if no cost or failed).
	Cost float64
	// Description is a human-readable summary of what happened.
	Description string
}

// DigRoomAction requests digging a new room in the cave.
type DigRoomAction struct {
	// RoomTypeID is the type of room to build (e.g. "dragon_lair").
	RoomTypeID string
	// Pos is the top-left position where the room will be placed.
	Pos types.Pos
	// Width is the room width in cells.
	Width int
	// Height is the room height in cells.
	Height int
}

// ActionType returns the action type identifier.
func (a DigRoomAction) ActionType() string { return "dig_room" }

// DigCorridorAction requests connecting two rooms with a corridor.
type DigCorridorAction struct {
	// FromRoomID is the source room ID.
	FromRoomID int
	// ToRoomID is the destination room ID.
	ToRoomID int
}

// ActionType returns the action type identifier.
func (a DigCorridorAction) ActionType() string { return "dig_corridor" }

// PlaceBeastAction requests placing an existing beast into a room.
type PlaceBeastAction struct {
	// SpeciesID is the species of beast to place.
	SpeciesID string
	// RoomID is the target room ID.
	RoomID int
}

// ActionType returns the action type identifier.
func (a PlaceBeastAction) ActionType() string { return "place_beast" }

// UpgradeRoomAction requests upgrading a room to the next level.
type UpgradeRoomAction struct {
	// RoomID is the room to upgrade.
	RoomID int
}

// ActionType returns the action type identifier.
func (a UpgradeRoomAction) ActionType() string { return "upgrade_room" }

// SummonBeastAction requests summoning a new beast of the given element.
type SummonBeastAction struct {
	// Element is the element of the beast to summon.
	Element types.Element
}

// ActionType returns the action type identifier.
func (a SummonBeastAction) ActionType() string { return "summon_beast" }

// EvolveBeastAction requests evolving a beast along its evolution path.
type EvolveBeastAction struct {
	// BeastID is the ID of the beast to evolve.
	BeastID int
}

// ActionType returns the action type identifier.
func (a EvolveBeastAction) ActionType() string { return "evolve_beast" }

// NoAction represents a deliberate choice to do nothing this tick.
type NoAction struct{}

// ActionType returns the action type identifier.
func (a NoAction) ActionType() string { return "no_action" }
