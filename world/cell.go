package world

// CellType represents the type of a single grid cell.
type CellType int

const (
	// Rock is an unexcavated cell (default state).
	Rock CellType = iota
	// Corridor is a passable path between rooms.
	Corridor
	// RoomFloor is a cell that belongs to a room interior.
	RoomFloor
	// Entrance is a cell that serves as a room entrance/exit.
	Entrance
)

// String returns the name of the CellType.
func (c CellType) String() string {
	switch c {
	case Rock:
		return "Rock"
	case Corridor:
		return "Corridor"
	case RoomFloor:
		return "RoomFloor"
	case Entrance:
		return "Entrance"
	default:
		return "Unknown"
	}
}

// Cell represents a single tile in the cave grid.
type Cell struct {
	// Type is the kind of terrain this cell contains.
	Type CellType
	// RoomID is the ID of the room this cell belongs to.
	// Zero means the cell is not part of any room.
	RoomID int
}
