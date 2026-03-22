package asset

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// TileSize is the pixel size of a single tile (width and height).
const TileSize = 32

// TilesetProvider supplies tile images for rendering the game map and entities.
type TilesetProvider interface {
	// GetTile returns a tile image for the given cell type and element.
	// The element is used to tint room/corridor tiles by their owning room's element.
	GetTile(cellType world.CellType, element types.Element) *ebiten.Image

	// GetBeastSprite returns a sprite for a beast of the given species and level.
	GetBeastSprite(species string, level int) *ebiten.Image

	// GetInvaderSprite returns a sprite for an invader of the given class.
	GetInvaderSprite(class string) *ebiten.Image
}
