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

// RoomRenderInfo holds pre-computed rendering information for a room,
// derived from the GameState (Cave + RoomTypeRegistry).
type RoomRenderInfo struct {
	// Element is the room's five-element attribute, used for tile coloring.
	Element types.Element
	// IsDragonHole indicates whether this room is a dragon hole (core room).
	IsDragonHole bool
}

// BuildRoomRenderMap creates a map from room ID to RoomRenderInfo by
// inspecting the cave's rooms and the room type registry. This decouples
// rendering from direct GameState access.
func BuildRoomRenderMap(cave *world.Cave, registry *world.RoomTypeRegistry) map[int]RoomRenderInfo {
	m := make(map[int]RoomRenderInfo, len(cave.Rooms))
	for _, room := range cave.Rooms {
		rt, err := registry.Get(room.TypeID)
		if err != nil {
			continue
		}
		m[room.ID] = RoomRenderInfo{
			Element:      rt.Element,
			IsDragonHole: rt.BaseCoreHP > 0,
		}
	}
	return m
}

// Draw renders the entire cave tilemap onto the screen.
// It iterates over all cells in the cave grid and draws the appropriate tile
// using the provided TilesetProvider. Room cells are tinted by their owning
// room's element (looked up via roomInfo). Dragon hole rooms are drawn in purple.
func (mv *MapView) Draw(screen *ebiten.Image, cave *world.Cave, roomInfo map[int]RoomRenderInfo, provider asset.TilesetProvider) {
	for y := 0; y < cave.Grid.Height; y++ {
		for x := 0; x < cave.Grid.Width; x++ {
			cell, err := cave.Grid.At(types.Pos{X: x, Y: y})
			if err != nil {
				continue
			}

			element := types.Wood // default for non-room cells
			if cell.RoomID != 0 {
				if info, ok := roomInfo[cell.RoomID]; ok {
					if info.IsDragonHole {
						tile := provider.GetDragonHoleTile()
						if tile != nil {
							px, py := mv.CellToScreen(x, y)
							op := &ebiten.DrawImageOptions{}
							op.GeoM.Translate(float64(px), float64(py))
							screen.DrawImage(tile, op)
						}
						continue
					}
					element = info.Element
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
