package scenario

import "github.com/nyasuto/seed/core/types"

// RoomPlacement describes where to place a pre-built room at scenario start.
type RoomPlacement struct {
	// TypeID is the room type identifier (e.g. "dragon_hole", "wood_room").
	TypeID string
	// Pos is the top-left grid position of the room.
	Pos types.Pos
	// Level is the initial level of the room (1 if omitted).
	Level int
}

// DragonVeinPlacement describes a dragon vein to create at scenario start.
type DragonVeinPlacement struct {
	// SourcePos is the grid position where the dragon vein originates.
	SourcePos types.Pos
	// Element is the elemental affinity of the dragon vein.
	Element types.Element
	// FlowRate is the base chi flow rate per tick.
	FlowRate float64
}

// BeastPlacement describes an initial beast to place at scenario start.
type BeastPlacement struct {
	// SpeciesID is the species identifier for the beast.
	SpeciesID string
	// RoomIndex is the index into PrebuiltRooms where this beast is placed.
	// A value of -1 means the beast is unassigned to any room.
	RoomIndex int
}

// InitialState defines the starting state of a cave for a scenario.
type InitialState struct {
	// CaveWidth is the width of the cave grid.
	CaveWidth int
	// CaveHeight is the height of the cave grid.
	CaveHeight int
	// TerrainSeed is the RNG seed used for terrain generation.
	TerrainSeed int64
	// TerrainDensity controls how densely hard terrain is generated (0.0–1.0).
	TerrainDensity float64
	// PrebuiltRooms is the list of rooms placed at scenario start.
	PrebuiltRooms []RoomPlacement
	// DragonVeins is the list of dragon veins created at scenario start.
	DragonVeins []DragonVeinPlacement
	// StartingChi is the initial chi pool balance.
	StartingChi float64
	// StartingBeasts is the list of beasts placed at scenario start.
	StartingBeasts []BeastPlacement
}
