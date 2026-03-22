package view

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
)

// EntityRenderer handles drawing beast and invader sprites on the map.
type EntityRenderer struct {
	mapView *MapView
}

// NewEntityRenderer creates an EntityRenderer that uses the given MapView
// for coordinate conversions.
func NewEntityRenderer(mv *MapView) *EntityRenderer {
	return &EntityRenderer{mapView: mv}
}

// RoomCenter returns the screen pixel coordinates of the center of a room.
// Returns (0, 0, false) if the room is not found.
func (er *EntityRenderer) RoomCenter(cave *world.Cave, roomID int) (px, py int, ok bool) {
	room := cave.RoomByID(roomID)
	if room == nil {
		return 0, 0, false
	}
	return er.RoomCenterFromRoom(room)
}

// RoomCenterFromRoom returns the screen pixel coordinates of the center of
// the given room. The sprite is centered within the room's tile area.
func (er *EntityRenderer) RoomCenterFromRoom(room *world.Room) (px, py int, ok bool) {
	// Calculate the center cell of the room.
	centerCellX := room.Pos.X + room.Width/2
	centerCellY := room.Pos.Y + room.Height/2

	// Convert to screen coordinates (top-left of the center cell).
	sx, sy := er.mapView.CellToScreen(centerCellX, centerCellY)

	// Offset to the center of the tile.
	px = sx + asset.TileSize/2
	py = sy + asset.TileSize/2
	return px, py, true
}

// DrawBeasts draws all beast sprites onto the screen. Each beast is drawn
// at the center of its assigned room. Beasts without a room (RoomID == 0)
// are skipped. If hiddenIDs is non-nil, beasts whose IDs are in the set
// are not drawn (used for blink effects).
func (er *EntityRenderer) DrawBeasts(screen *ebiten.Image, cave *world.Cave, beasts []*senju.Beast, provider asset.TilesetProvider, hiddenIDs ...map[int]bool) {
	var hidden map[int]bool
	if len(hiddenIDs) > 0 {
		hidden = hiddenIDs[0]
	}
	for _, beast := range beasts {
		if beast.RoomID == 0 {
			continue
		}
		if hidden != nil && hidden[beast.ID] {
			continue
		}
		px, py, ok := er.RoomCenter(cave, beast.RoomID)
		if !ok {
			continue
		}
		sprite := provider.GetBeastSprite(beast.SpeciesID, beast.Level)
		if sprite == nil {
			continue
		}
		er.drawSpriteAt(screen, sprite, px, py)
	}
}

// DrawInvaders draws all active invader sprites onto the screen. Each invader
// is drawn at the center of its current room. Invaders not in a room
// (CurrentRoomID == 0) or in a terminal state (Defeated/GoalAchieved) are skipped.
func (er *EntityRenderer) DrawInvaders(screen *ebiten.Image, cave *world.Cave, waves []*invasion.InvasionWave, provider asset.TilesetProvider) {
	for _, wave := range waves {
		if !wave.IsActive() {
			continue
		}
		for _, inv := range wave.Invaders {
			if inv.CurrentRoomID == 0 {
				continue
			}
			if inv.State == invasion.Defeated || inv.State == invasion.GoalAchieved {
				continue
			}
			px, py, ok := er.RoomCenter(cave, inv.CurrentRoomID)
			if !ok {
				continue
			}
			sprite := provider.GetInvaderSprite(inv.ClassID)
			if sprite == nil {
				continue
			}
			er.drawSpriteAt(screen, sprite, px, py)
		}
	}
}

// drawSpriteAt draws a sprite centered at the given screen pixel coordinates.
func (er *EntityRenderer) drawSpriteAt(screen, sprite *ebiten.Image, centerX, centerY int) {
	w, h := sprite.Bounds().Dx(), sprite.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(centerX-w/2), float64(centerY-h/2))
	screen.DrawImage(sprite, op)
}
