package view

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
)

// TooltipInfo holds the text lines to display in a tooltip.
type TooltipInfo struct {
	Lines []string
}

// BuildTooltipInfo creates tooltip information for the cell at (cx, cy) in the cave.
// It shows the CellType, and for room cells also the room ID, element, and level.
func BuildTooltipInfo(cave *world.Cave, registry *world.RoomTypeRegistry, cx, cy int) TooltipInfo {
	cell, err := cave.Grid.At(types.Pos{X: cx, Y: cy})
	if err != nil {
		return TooltipInfo{Lines: []string{"(invalid)"}}
	}

	info := TooltipInfo{}
	info.Lines = append(info.Lines, fmt.Sprintf("(%d, %d) %s", cx, cy, cell.Type.String()))

	if cell.RoomID != 0 {
		room := cave.RoomByID(cell.RoomID)
		if room != nil {
			rt, rtErr := registry.Get(room.TypeID)
			if rtErr == nil {
				info.Lines = append(info.Lines, fmt.Sprintf("Room #%d  %s Lv%d", room.ID, rt.Element.String(), room.Level))
			} else {
				info.Lines = append(info.Lines, fmt.Sprintf("Room #%d  Lv%d", room.ID, room.Level))
			}
		}
	}

	return info
}

// Tooltip draws a tooltip near the given screen position.
type Tooltip struct{}

// Draw renders tooltip info near the cursor position on screen.
// The tooltip is drawn as a semi-transparent background box with text lines.
func (tt *Tooltip) Draw(screen *ebiten.Image, info TooltipInfo, screenX, screenY int) {
	if len(info.Lines) == 0 {
		return
	}

	const (
		lineHeight = 16
		paddingX   = 6
		paddingY   = 4
		offsetX    = 12 // offset from cursor
		offsetY    = -8
		charWidth  = 6 // approximate width of a debug-print character
	)

	// Calculate box dimensions.
	maxLen := 0
	for _, line := range info.Lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	boxW := maxLen*charWidth + paddingX*2
	boxH := len(info.Lines)*lineHeight + paddingY*2

	// Position the tooltip, clamping to screen bounds.
	tx := screenX + offsetX
	ty := screenY + offsetY

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if tx+boxW > sw {
		tx = screenX - boxW - 4
	}
	if ty+boxH > sh {
		ty = sh - boxH
	}
	if ty < 0 {
		ty = 0
	}

	// Draw background.
	bg := ebiten.NewImage(boxW, boxH)
	bg.Fill(color.RGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xDD})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(tx), float64(ty))
	screen.DrawImage(bg, op)

	// Draw border.
	drawRect(screen, tx, ty, boxW, boxH, asset.ColorUIBorder)

	// Draw text lines.
	for i, line := range info.Lines {
		ebitenutil.DebugPrintAt(screen, line, tx+paddingX, ty+paddingY+i*lineHeight)
	}
}

// drawRect draws a 1-pixel rectangle outline.
func drawRect(screen *ebiten.Image, x, y, w, h int, clr color.Color) {
	line := ebiten.NewImage(w, 1)
	line.Fill(clr)
	col := ebiten.NewImage(1, h)
	col.Fill(clr)

	op := &ebiten.DrawImageOptions{}
	// Top
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(line, op)
	// Bottom
	op.GeoM.Reset()
	op.GeoM.Translate(float64(x), float64(y+h-1))
	screen.DrawImage(line, op)
	// Left
	op.GeoM.Reset()
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(col, op)
	// Right
	op.GeoM.Reset()
	op.GeoM.Translate(float64(x+w-1), float64(y))
	screen.DrawImage(col, op)
}
