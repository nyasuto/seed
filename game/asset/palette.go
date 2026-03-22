package asset

import "image/color"

// Element colors — one per five-element (五行).
var (
	ColorWood  = color.RGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF} // green
	ColorFire  = color.RGBA{R: 0xE5, G: 0x39, B: 0x35, A: 0xFF} // red
	ColorEarth = color.RGBA{R: 0x8D, G: 0x6E, B: 0x63, A: 0xFF} // brown
	ColorMetal = color.RGBA{R: 0xFF, G: 0xB3, B: 0x00, A: 0xFF} // yellow
	ColorWater = color.RGBA{R: 0x1E, G: 0x88, B: 0xE5, A: 0xFF} // blue
)

// Terrain colors — mapped to CellType and special room types.
var (
	ColorWall         = color.RGBA{R: 0x61, G: 0x61, B: 0x61, A: 0xFF} // dark gray (Rock)
	ColorFloor        = color.RGBA{R: 0xBD, G: 0xBD, B: 0xBD, A: 0xFF} // light gray (RoomFloor)
	ColorCorridor     = color.RGBA{R: 0x9E, G: 0x9E, B: 0x9E, A: 0xFF} // gray (CorridorFloor)
	ColorDragonHole   = color.RGBA{R: 0xAB, G: 0x47, B: 0xBC, A: 0xFF} // purple (龍穴)
	ColorHardRock     = color.RGBA{R: 0x37, G: 0x37, B: 0x37, A: 0xFF} // very dark gray
	ColorWaterTerrain = color.RGBA{R: 0x0D, G: 0x47, B: 0xA1, A: 0xFF} // deep blue
	ColorEntrance     = color.RGBA{R: 0xA1, G: 0x88, B: 0x7F, A: 0xFF} // warm gray
)

// UI colors.
var (
	ColorUIBackground = color.RGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xFF}
	ColorUIText       = color.RGBA{R: 0xFA, G: 0xFA, B: 0xFA, A: 0xFF}
	ColorUIBorder     = color.RGBA{R: 0x75, G: 0x75, B: 0x75, A: 0xFF}
)
