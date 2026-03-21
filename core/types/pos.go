package types

// Pos represents a 2D grid coordinate.
type Pos struct {
	X, Y int
}

// Add returns the sum of two positions.
func (p Pos) Add(other Pos) Pos {
	return Pos{X: p.X + other.X, Y: p.Y + other.Y}
}

// Sub returns the difference of two positions.
func (p Pos) Sub(other Pos) Pos {
	return Pos{X: p.X - other.X, Y: p.Y - other.Y}
}

// Distance returns the Manhattan distance between two positions.
func (p Pos) Distance(other Pos) int {
	dx := p.X - other.X
	if dx < 0 {
		dx = -dx
	}
	dy := p.Y - other.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// Neighbors returns the four orthogonally adjacent positions (N, S, E, W).
func (p Pos) Neighbors() [4]Pos {
	return [4]Pos{
		{p.X, p.Y - 1}, // North
		{p.X, p.Y + 1}, // South
		{p.X + 1, p.Y}, // East
		{p.X - 1, p.Y}, // West
	}
}
