package world

import "fmt"

// Cave represents an entire dungeon map, managing the grid, rooms, and corridors.
// It provides auto-incrementing IDs for rooms and corridors.
type Cave struct {
	Grid            *Grid
	Rooms           []*Room
	Corridors       []Corridor
	nextRoomID      int
	nextCorridorID  int
}

// NewCave creates a new Cave with a grid of the specified dimensions.
// Width and height must be positive.
func NewCave(w, h int) (*Cave, error) {
	grid, err := NewGrid(w, h)
	if err != nil {
		return nil, fmt.Errorf("creating cave: %w", err)
	}
	return &Cave{
		Grid:           grid,
		Rooms:          make([]*Room, 0),
		Corridors:      make([]Corridor, 0),
		nextRoomID:     1,
		nextCorridorID: 1,
	}, nil
}

// RoomByID returns the room with the given ID, or nil if not found.
func (c *Cave) RoomByID(id int) *Room {
	for _, r := range c.Rooms {
		if r.ID == id {
			return r
		}
	}
	return nil
}
