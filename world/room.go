package world

import "github.com/ponpoko/chaosseed-core/types"

// RoomEntrance represents an entrance point of a room.
type RoomEntrance struct {
	// Pos is the grid position of the entrance.
	Pos types.Pos
	// Dir is the direction the entrance faces (outward from the room).
	Dir types.Direction
}

// Room represents a placed room instance in the cave.
type Room struct {
	// ID is a unique identifier for this room instance.
	ID int
	// TypeID is the ID of the RoomType this room is based on.
	TypeID string
	// Pos is the top-left corner position of the room on the grid.
	Pos types.Pos
	// Width is the horizontal size of the room.
	Width int
	// Height is the vertical size of the room.
	Height int
	// Level is the current level of the room (starts at 1).
	Level int
	// Entrances is the list of entrance points for this room.
	Entrances []RoomEntrance
}

// Contains reports whether the given position is inside the room bounds.
func (r *Room) Contains(pos types.Pos) bool {
	return pos.X >= r.Pos.X && pos.X < r.Pos.X+r.Width &&
		pos.Y >= r.Pos.Y && pos.Y < r.Pos.Y+r.Height
}
