package asset

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// PlaceholderProvider generates colored rectangles as placeholder tiles.
// It satisfies the TilesetProvider interface.
type PlaceholderProvider struct {
	// tiles caches generated tile images keyed by (CellType, Element).
	tiles map[tileKey]*ebiten.Image
	// beastSprites caches beast placeholder sprites.
	beastSprites map[string]*ebiten.Image
	// invaderSprites caches invader placeholder sprites.
	invaderSprites map[string]*ebiten.Image
}

type tileKey struct {
	cell    world.CellType
	element types.Element
}

// NewPlaceholderProvider creates a PlaceholderProvider with pre-generated tiles
// for all CellType and Element combinations.
func NewPlaceholderProvider() *PlaceholderProvider {
	p := &PlaceholderProvider{
		tiles:          make(map[tileKey]*ebiten.Image),
		beastSprites:   make(map[string]*ebiten.Image),
		invaderSprites: make(map[string]*ebiten.Image),
	}
	p.generateAllTiles()
	return p
}

// GetTile returns a placeholder tile for the given cell type and element.
func (p *PlaceholderProvider) GetTile(cellType world.CellType, element types.Element) *ebiten.Image {
	key := tileKey{cell: cellType, element: element}
	if img, ok := p.tiles[key]; ok {
		return img
	}
	// Fallback: generate on demand (shouldn't happen after init).
	img := p.generateTile(cellType, element)
	p.tiles[key] = img
	return img
}

// GetBeastSprite returns a placeholder sprite for a beast.
func (p *PlaceholderProvider) GetBeastSprite(species string, level int) *ebiten.Image {
	if img, ok := p.beastSprites[species]; ok {
		return img
	}
	img := generateColoredRect(color.RGBA{R: 0x00, G: 0xE6, B: 0x76, A: 0xFF})
	drawCenteredChar(img, 'B')
	p.beastSprites[species] = img
	return img
}

// GetInvaderSprite returns a placeholder sprite for an invader.
func (p *PlaceholderProvider) GetInvaderSprite(class string) *ebiten.Image {
	if img, ok := p.invaderSprites[class]; ok {
		return img
	}
	img := generateColoredRect(color.RGBA{R: 0xFF, G: 0x17, B: 0x44, A: 0xFF})
	drawCenteredChar(img, 'I')
	p.invaderSprites[class] = img
	return img
}

// generateAllTiles pre-generates tiles for every CellType × Element combination.
func (p *PlaceholderProvider) generateAllTiles() {
	cellTypes := []world.CellType{
		world.Rock,
		world.CorridorFloor,
		world.RoomFloor,
		world.Entrance,
		world.HardRock,
		world.Water,
	}
	for _, ct := range cellTypes {
		for e := types.Wood; e <= types.Water; e++ {
			key := tileKey{cell: ct, element: e}
			p.tiles[key] = p.generateTile(ct, e)
		}
	}
}

// generateTile creates a single 32×32 placeholder tile.
func (p *PlaceholderProvider) generateTile(ct world.CellType, element types.Element) *ebiten.Image {
	switch ct {
	case world.HardRock:
		img := generateColoredRect(ColorHardRock)
		drawCrossPattern(img, color.RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xFF})
		return img
	case world.Water:
		img := generateColoredRect(ColorWaterTerrain)
		drawWavePattern(img, color.RGBA{R: 0x42, G: 0xA5, B: 0xF5, A: 0xFF})
		return img
	case world.Rock:
		return generateColoredRect(ColorWall)
	case world.CorridorFloor:
		return generateColoredRect(ColorCorridor)
	case world.Entrance:
		return generateColoredRect(ColorEntrance)
	case world.RoomFloor:
		return generateColoredRect(elementColor(element))
	default:
		return generateColoredRect(ColorWall)
	}
}

// elementColor returns the palette color for an element.
func elementColor(e types.Element) color.RGBA {
	switch e {
	case types.Wood:
		return ColorWood
	case types.Fire:
		return ColorFire
	case types.Earth:
		return ColorEarth
	case types.Metal:
		return ColorMetal
	case types.Water:
		return ColorWater
	default:
		return ColorFloor
	}
}

// generateColoredRect creates a TileSize×TileSize image filled with the given color.
func generateColoredRect(c color.Color) *ebiten.Image {
	img := ebiten.NewImage(TileSize, TileSize)
	img.Fill(c)
	return img
}

// drawCrossPattern draws an ✕ pattern on the tile.
func drawCrossPattern(img *ebiten.Image, c color.Color) {
	pix := image.NewRGBA(image.Rect(0, 0, TileSize, TileSize))
	r, g, b, a := c.RGBA()
	cr := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	for i := 2; i < TileSize-2; i++ {
		pix.SetRGBA(i, i, cr)
		pix.SetRGBA(i+1, i, cr)
		pix.SetRGBA(TileSize-1-i, i, cr)
		pix.SetRGBA(TileSize-2-i, i, cr)
	}
	overlay := ebiten.NewImageFromImage(pix)
	img.DrawImage(overlay, nil)
}

// drawWavePattern draws a ～ wave pattern on the tile.
func drawWavePattern(img *ebiten.Image, c color.Color) {
	pix := image.NewRGBA(image.Rect(0, 0, TileSize, TileSize))
	r, g, b, a := c.RGBA()
	cr := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	// Draw two horizontal wave lines.
	for _, baseY := range []int{10, 22} {
		for x := 0; x < TileSize; x++ {
			// Simple sine-like wave: offset by ±2 pixels.
			dy := 0
			quarter := TileSize / 4
			switch {
			case x%quarter < quarter/2:
				dy = -2 + (x%quarter)*4/(quarter)
			default:
				dy = 2 - (x%quarter-quarter/2)*4/(quarter)
			}
			y := baseY + dy
			if y >= 0 && y < TileSize {
				pix.SetRGBA(x, y, cr)
			}
			if y+1 >= 0 && y+1 < TileSize {
				pix.SetRGBA(x, y+1, cr)
			}
		}
	}
	overlay := ebiten.NewImageFromImage(pix)
	img.DrawImage(overlay, nil)
}

// drawCenteredChar draws a simple character marker at the center of the tile.
// Uses a minimal 5×7 bitmap font for 'B' and 'I'.
func drawCenteredChar(img *ebiten.Image, ch byte) {
	var bitmap [7]uint8
	switch ch {
	case 'B':
		bitmap = [7]uint8{
			0b11110,
			0b10001,
			0b10001,
			0b11110,
			0b10001,
			0b10001,
			0b11110,
		}
	case 'I':
		bitmap = [7]uint8{
			0b11111,
			0b00100,
			0b00100,
			0b00100,
			0b00100,
			0b00100,
			0b11111,
		}
	default:
		return
	}

	pix := image.NewRGBA(image.Rect(0, 0, TileSize, TileSize))
	white := color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	ox := (TileSize - 5) / 2
	oy := (TileSize - 7) / 2
	for row := 0; row < 7; row++ {
		for col := 0; col < 5; col++ {
			if bitmap[row]&(1<<uint(4-col)) != 0 {
				pix.SetRGBA(ox+col, oy+row, white)
			}
		}
	}
	overlay := ebiten.NewImageFromImage(pix)
	img.DrawImage(overlay, nil)
}

// Compile-time check that PlaceholderProvider satisfies TilesetProvider.
var _ TilesetProvider = (*PlaceholderProvider)(nil)
