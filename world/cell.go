package world

// CellType represents the type of a single grid cell.
type CellType int

const (
	// Rock is an unexcavated cell (default state).
	Rock CellType = iota
	// CorridorFloor is a passable path between rooms.
	CorridorFloor
	// RoomFloor is a cell that belongs to a room interior.
	RoomFloor
	// Entrance is a cell that serves as a room entrance/exit.
	Entrance
	// HardRock is an indestructible rock cell that cannot be excavated.
	HardRock
	// Water is an underground water vein cell that cannot be excavated.
	Water
)

// String returns the name of the CellType.
func (c CellType) String() string {
	switch c {
	case Rock:
		return "Rock"
	case CorridorFloor:
		return "CorridorFloor"
	case RoomFloor:
		return "RoomFloor"
	case Entrance:
		return "Entrance"
	case HardRock:
		return "HardRock"
	case Water:
		return "Water"
	default:
		return "Unknown"
	}
}

// IsImpassable reports whether the cell type cannot be excavated or traversed.
func (c CellType) IsImpassable() bool {
	return c == HardRock || c == Water
}

// Cell represents a single tile in the cave grid.
type Cell struct {
	// Type is the kind of terrain this cell contains.
	Type CellType
	// RoomID is the ID of the room this cell belongs to.
	// Zero means the cell is not part of any room.
	RoomID int
}
