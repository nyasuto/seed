package world

import "github.com/ponpoko/chaosseed-core/types"

// Corridor represents a passable path connecting two rooms in the cave.
type Corridor struct {
	// ID is a unique identifier for this corridor.
	ID int
	// FromRoomID is the ID of the room where the corridor starts.
	FromRoomID int
	// ToRoomID is the ID of the room where the corridor ends.
	ToRoomID int
	// Path is the ordered list of grid positions that form the corridor.
	Path []types.Pos
}
