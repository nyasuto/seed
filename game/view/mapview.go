// Package view provides rendering components for the game GUI.
package view

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
)

// MapView handles rendering of Cave data as a tilemap and provides
// coordinate conversion between cell positions and screen positions.
type MapView struct {
	// OffsetX is the horizontal pixel offset of the map origin on screen.
	OffsetX int
	// OffsetY is the vertical pixel offset of the map origin on screen.
	OffsetY int
}

// NewMapView creates a MapView with the given pixel offset.
func NewMapView(offsetX, offsetY int) *MapView {
	return &MapView{
		OffsetX: offsetX,
		OffsetY: offsetY,
	}
}

// CellToScreen converts a cell grid position to the top-left screen pixel position.
func (mv *MapView) CellToScreen(cx, cy int) (px, py int) {
	px = cx*asset.TileSize + mv.OffsetX
	py = cy*asset.TileSize + mv.OffsetY
	return px, py
}

// ScreenToCell converts a screen pixel position to the corresponding cell grid position.
// Returns the cell coordinates and whether the position is valid (i.e., within grid bounds
// when checked against the provided grid dimensions).
func (mv *MapView) ScreenToCell(px, py int, gridWidth, gridHeight int) (cx, cy int, ok bool) {
	// Subtract offset to get map-relative pixel coordinates.
	rx := px - mv.OffsetX
	ry := py - mv.OffsetY

	// Negative pixel coordinates are outside the map.
	if rx < 0 || ry < 0 {
		return 0, 0, false
	}

	cx = rx / asset.TileSize
	cy = ry / asset.TileSize

	if cx < 0 || cx >= gridWidth || cy < 0 || cy >= gridHeight {
		return 0, 0, false
	}

	return cx, cy, true
}

// Draw renders the entire cave tilemap onto the screen.
// It iterates over all cells in the cave grid and draws the appropriate tile
// using the provided TilesetProvider. Room and corridor cells are tinted
// by their owning room's element, looked up via the RoomTypeRegistry.
func (mv *MapView) Draw(screen *ebiten.Image, cave *world.Cave, registry *world.RoomTypeRegistry, provider asset.TilesetProvider) {
	for y := 0; y < cave.Grid.Height; y++ {
		for x := 0; x < cave.Grid.Width; x++ {
			cell, err := cave.Grid.At(types.Pos{X: x, Y: y})
			if err != nil {
				continue
			}

			element := types.Wood // default for non-room cells
			if cell.RoomID != 0 {
				room := cave.RoomByID(cell.RoomID)
				if room != nil {
					rt, rtErr := registry.Get(room.TypeID)
					if rtErr == nil {
						element = rt.Element
					}
				}
			}

			tile := provider.GetTile(cell.Type, element)
			if tile == nil {
				continue
			}

			px, py := mv.CellToScreen(x, y)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(px), float64(py))
			screen.DrawImage(tile, op)
		}
	}
}
