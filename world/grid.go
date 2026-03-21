package world

import (
	"errors"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// ErrOutOfBounds is returned when a position is outside the grid boundaries.
var ErrOutOfBounds = errors.New("position out of bounds")

// Grid represents a 2D grid of cells forming the cave map.
type Grid struct {
	Width  int
	Height int
	cells  [][]Cell
}

// NewGrid creates a new grid with the given dimensions.
// All cells are initialized to Rock with RoomID 0.
func NewGrid(w, h int) (*Grid, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("grid dimensions must be positive: got %dx%d", w, h)
	}
	cells := make([][]Cell, h)
	for y := range cells {
		cells[y] = make([]Cell, w)
	}
	return &Grid{
		Width:  w,
		Height: h,
		cells:  cells,
	}, nil
}

// InBounds reports whether the given position is within the grid boundaries.
func (g *Grid) InBounds(pos types.Pos) bool {
	return pos.X >= 0 && pos.X < g.Width && pos.Y >= 0 && pos.Y < g.Height
}

// At returns the cell at the given position.
func (g *Grid) At(pos types.Pos) (Cell, error) {
	if !g.InBounds(pos) {
		return Cell{}, fmt.Errorf("At(%d, %d): %w", pos.X, pos.Y, ErrOutOfBounds)
	}
	return g.cells[pos.Y][pos.X], nil
}

// Set sets the cell at the given position.
func (g *Grid) Set(pos types.Pos, cell Cell) error {
	if !g.InBounds(pos) {
		return fmt.Errorf("Set(%d, %d): %w", pos.X, pos.Y, ErrOutOfBounds)
	}
	g.cells[pos.Y][pos.X] = cell
	return nil
}
