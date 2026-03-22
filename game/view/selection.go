package view

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/game/asset"
)

const (
	selPanelPad    = 8
	selBtnHeight   = 24
	selBtnGap      = 4
	selMinWidth    = 240
	selBtnMinWidth = 200
)

// SelectionItem represents a single selectable entry in a SelectionPanel.
type SelectionItem struct {
	// Label is the displayed text.
	Label string
	// Color is an optional accent color. If zero-value, the default UI color is used.
	Color color.RGBA
}

// SelectionPanel is a generic modal selection UI that displays a title
// and a list of labeled buttons. It blocks other input while visible.
type SelectionPanel struct {
	title   string
	items   []SelectionItem
	buttons []Button
	x, y    int
	width   int
	height  int
}

// NewSelectionPanel creates a SelectionPanel centered at (centerX, centerY)
// with the given title and items.
func NewSelectionPanel(centerX, centerY int, title string, items []SelectionItem) *SelectionPanel {
	btnWidth := selBtnMinWidth
	// Ensure buttons are wide enough for the longest label.
	for _, item := range items {
		tw := TextWidth(item.Label) + selPanelPad*2
		if tw > btnWidth {
			btnWidth = tw
		}
	}

	panelWidth := btnWidth + selPanelPad*2
	if panelWidth < selMinWidth {
		panelWidth = selMinWidth
	}

	// Panel height: padding + title line + gap + (btnHeight+gap)*n + padding.
	panelHeight := selPanelPad + LineHeight + selBtnGap +
		len(items)*(selBtnHeight+selBtnGap) + selPanelPad

	px := centerX - panelWidth/2
	py := centerY - panelHeight/2

	bx := px + (panelWidth-btnWidth)/2
	by := py + selPanelPad + LineHeight + selBtnGap

	buttons := make([]Button, len(items))
	for i := range items {
		buttons[i] = NewButton(bx, by, btnWidth, selBtnHeight, items[i].Label)
		by += selBtnHeight + selBtnGap
	}

	return &SelectionPanel{
		title:   title,
		items:   items,
		buttons: buttons,
		x:       px,
		y:       py,
		width:   panelWidth,
		height:  panelHeight,
	}
}

// HandleClick checks if a click at (px, py) hits any item button.
// Returns the index of the clicked item and true, or (-1, false) if
// no item was hit.
func (sp *SelectionPanel) HandleClick(px, py int) (int, bool) {
	for i, btn := range sp.buttons {
		if btn.Contains(px, py) {
			return i, true
		}
	}
	return -1, false
}

// Contains reports whether the screen position (px, py) is within the panel bounds.
// Use this to determine if clicks should be blocked from reaching the game layer.
func (sp *SelectionPanel) Contains(px, py int) bool {
	r := image.Rect(sp.x, sp.y, sp.x+sp.width, sp.y+sp.height)
	return image.Pt(px, py).In(r)
}

// Draw renders the selection panel onto the screen.
func (sp *SelectionPanel) Draw(screen *ebiten.Image) {
	// Panel background.
	bg := ebiten.NewImage(sp.width, sp.height)
	bg.Fill(color.RGBA{R: 0x1A, G: 0x1A, B: 0x2E, A: 0xE0})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(sp.x), float64(sp.y))
	screen.DrawImage(bg, op)

	// Border.
	drawRectBorder(screen, image.Rect(sp.x, sp.y, sp.x+sp.width, sp.y+sp.height), asset.ColorUIBorder)

	// Title.
	DrawText(screen, sp.title, sp.x+selPanelPad, sp.y+selPanelPad)

	// Item buttons.
	for i, btn := range sp.buttons {
		if sp.items[i].Color != (color.RGBA{}) {
			drawElementButton(screen, btn, sp.items[i].Color)
		} else {
			btn.Draw(screen, ButtonNormal)
		}
	}
}

// Bounds returns the panel rectangle.
func (sp *SelectionPanel) Bounds() image.Rectangle {
	return image.Rect(sp.x, sp.y, sp.x+sp.width, sp.y+sp.height)
}

// ItemCount returns the number of selectable items.
func (sp *SelectionPanel) ItemCount() int {
	return len(sp.items)
}
