package view

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/game/asset"
)

const (
	elemPanelWidth  = 240
	elemPanelHeight = 160
	elemBtnWidth    = 200
	elemBtnHeight   = 24
	elemBtnGap      = 4
	elemPanelPad    = 8
)

// ElementPanel displays a list of element buttons for the player to choose from.
type ElementPanel struct {
	buttons  []Button
	elements []types.Element
	x, y     int
}

// NewElementPanel creates an ElementPanel centered at the given screen position.
func NewElementPanel(centerX, centerY int) *ElementPanel {
	px := centerX - elemPanelWidth/2
	py := centerY - elemPanelHeight/2

	elems := []types.Element{types.Wood, types.Fire, types.Earth, types.Metal, types.Water}
	buttons := make([]Button, len(elems))

	bx := px + (elemPanelWidth-elemBtnWidth)/2
	by := py + elemPanelPad + LineHeight + elemBtnGap // space for title

	for i, elem := range elems {
		buttons[i] = NewButton(bx, by, elemBtnWidth, elemBtnHeight, elem.String())
		by += elemBtnHeight + elemBtnGap
		_ = elem
	}

	return &ElementPanel{
		buttons:  buttons,
		elements: elems,
		x:        px,
		y:        py,
	}
}

// Draw renders the element selection panel.
func (ep *ElementPanel) Draw(screen *ebiten.Image) {
	// Panel background.
	bg := ebiten.NewImage(elemPanelWidth, elemPanelHeight)
	bg.Fill(color.RGBA{R: 0x1A, G: 0x1A, B: 0x2E, A: 0xE0})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ep.x), float64(ep.y))
	screen.DrawImage(bg, op)

	// Border.
	drawRectBorder(screen, image.Rect(ep.x, ep.y, ep.x+elemPanelWidth, ep.y+elemPanelHeight), asset.ColorUIBorder)

	// Title.
	DrawText(screen, "Select Element", ep.x+elemPanelPad, ep.y+elemPanelPad)

	// Element buttons with element-specific colors.
	for i, btn := range ep.buttons {
		elemColor := asset.ElementColor(ep.elements[i])
		drawElementButton(screen, btn, elemColor)
	}
}

// HandleClick checks if a click hits an element button and returns the
// selected element and whether a selection was made.
func (ep *ElementPanel) HandleClick(px, py int) (types.Element, bool) {
	for i, btn := range ep.buttons {
		if btn.Contains(px, py) {
			return ep.elements[i], true
		}
	}
	return 0, false
}

// drawElementButton renders a button with an element-specific background color.
func drawElementButton(screen *ebiten.Image, btn Button, elemColor color.RGBA) {
	w := btn.Rect.Dx()
	h := btn.Rect.Dy()

	bg := ebiten.NewImage(w, h)
	dimmed := color.RGBA{R: elemColor.R / 2, G: elemColor.G / 2, B: elemColor.B / 2, A: 0xFF}
	bg.Fill(dimmed)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(btn.Rect.Min.X), float64(btn.Rect.Min.Y))
	screen.DrawImage(bg, op)

	drawRectBorder(screen, btn.Rect, elemColor)

	tw := TextWidth(btn.Label)
	tx := btn.Rect.Min.X + (w-tw)/2
	ty := btn.Rect.Min.Y + (h-LineHeight)/2
	DrawText(screen, btn.Label, tx, ty)
}
