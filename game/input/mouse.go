package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/game/view"
)

// MouseTracker tracks the mouse cursor position each frame and converts
// screen pixel coordinates to cell grid coordinates via the MapView.
type MouseTracker struct {
	mapView *view.MapView

	// gridWidth and gridHeight define the cave grid dimensions for bounds checking.
	gridWidth  int
	gridHeight int

	// CellX and CellY hold the hovered cell coordinates.
	// Valid is true when the cursor is over a valid cell.
	CellX int
	CellY int
	Valid bool
}

// NewMouseTracker creates a MouseTracker bound to the given MapView and grid dimensions.
func NewMouseTracker(mapView *view.MapView, gridWidth, gridHeight int) *MouseTracker {
	return &MouseTracker{
		mapView:    mapView,
		gridWidth:  gridWidth,
		gridHeight: gridHeight,
	}
}

// Update reads the current mouse cursor position and updates the hovered cell.
// Call this once per frame in Game.Update().
func (mt *MouseTracker) Update() {
	px, py := ebiten.CursorPosition()
	mt.CellX, mt.CellY, mt.Valid = mt.mapView.ScreenToCell(px, py, mt.gridWidth, mt.gridHeight)
}

// CursorCell returns the cell coordinates under the cursor and whether they are valid.
func (mt *MouseTracker) CursorCell() (cx, cy int, ok bool) {
	return mt.CellX, mt.CellY, mt.Valid
}
