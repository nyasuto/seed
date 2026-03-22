package input

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// CellConverter converts screen pixel coordinates to cell grid coordinates.
type CellConverter interface {
	ScreenToCell(px, py, gridWidth, gridHeight int) (cx, cy int, ok bool)
}

// MouseTracker tracks the mouse cursor position each frame and converts
// screen pixel coordinates to cell grid coordinates via a CellConverter.
type MouseTracker struct {
	converter CellConverter

	// gridWidth and gridHeight define the cave grid dimensions for bounds checking.
	gridWidth  int
	gridHeight int

	// CellX and CellY hold the hovered cell coordinates.
	// Valid is true when the cursor is over a valid cell.
	CellX int
	CellY int
	Valid bool
}

// NewMouseTracker creates a MouseTracker bound to the given CellConverter and grid dimensions.
func NewMouseTracker(converter CellConverter, gridWidth, gridHeight int) *MouseTracker {
	return &MouseTracker{
		converter:  converter,
		gridWidth:  gridWidth,
		gridHeight: gridHeight,
	}
}

// Update reads the current mouse cursor position and updates the hovered cell.
// Call this once per frame in Game.Update().
func (mt *MouseTracker) Update() {
	px, py := ebiten.CursorPosition()
	mt.CellX, mt.CellY, mt.Valid = mt.converter.ScreenToCell(px, py, mt.gridWidth, mt.gridHeight)
}

// CursorCell returns the cell coordinates under the cursor and whether they are valid.
func (mt *MouseTracker) CursorCell() (cx, cy int, ok bool) {
	return mt.CellX, mt.CellY, mt.Valid
}
