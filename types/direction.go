package types

// Direction represents a cardinal direction on the grid.
type Direction int

const (
	// North points toward decreasing Y.
	North Direction = iota
	// South points toward increasing Y.
	South
	// East points toward increasing X.
	East
	// West points toward decreasing X.
	West
)

// Opposite returns the direction opposite to d.
func (d Direction) Opposite() Direction {
	switch d {
	case North:
		return South
	case South:
		return North
	case East:
		return West
	case West:
		return East
	default:
		return d
	}
}

// Delta returns the unit movement vector for d as a Pos.
func (d Direction) Delta() Pos {
	switch d {
	case North:
		return Pos{0, -1}
	case South:
		return Pos{0, 1}
	case East:
		return Pos{1, 0}
	case West:
		return Pos{-1, 0}
	default:
		return Pos{0, 0}
	}
}

// String returns the name of the direction.
func (d Direction) String() string {
	switch d {
	case North:
		return "North"
	case South:
		return "South"
	case East:
		return "East"
	case West:
		return "West"
	default:
		return "Unknown"
	}
}
