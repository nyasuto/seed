package view

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/game/asset"
)

// ButtonState represents the visual state of a button.
type ButtonState int

const (
	// ButtonNormal is the default button state.
	ButtonNormal ButtonState = iota
	// ButtonHover indicates the cursor is over the button.
	ButtonHover
	// ButtonActive indicates the button is currently active/selected.
	ButtonActive
)

// Button represents a clickable UI button with a label.
type Button struct {
	Rect  image.Rectangle
	Label string
}

// NewButton creates a Button at the given position with the given size and label.
func NewButton(x, y, w, h int, label string) Button {
	return Button{
		Rect:  image.Rect(x, y, x+w, y+h),
		Label: label,
	}
}

// ButtonFromRect creates a Button from an existing rectangle and label.
func ButtonFromRect(r image.Rectangle, label string) Button {
	return Button{Rect: r, Label: label}
}

// Contains reports whether the screen coordinate (px, py) is inside the button.
func (b Button) Contains(px, py int) bool {
	return image.Pt(px, py).In(b.Rect)
}

// Draw renders the button onto the screen with the given state.
func (b Button) Draw(screen *ebiten.Image, state ButtonState) {
	w := b.Rect.Dx()
	h := b.Rect.Dy()

	bg := ebiten.NewImage(w, h)
	bg.Fill(buttonBGColor(state))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.Rect.Min.X), float64(b.Rect.Min.Y))
	screen.DrawImage(bg, op)

	// Border.
	drawRectBorder(screen, b.Rect, buttonBorderColor(state))

	// Centered label.
	tw := TextWidth(b.Label)
	tx := b.Rect.Min.X + (w-tw)/2
	ty := b.Rect.Min.Y + (h-LineHeight)/2
	DrawText(screen, b.Label, tx, ty)
}

func buttonBGColor(state ButtonState) color.RGBA {
	switch state {
	case ButtonActive:
		return color.RGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF} // green highlight
	case ButtonHover:
		return color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xFF} // lighter gray
	default:
		return asset.ColorUIBackground
	}
}

func buttonBorderColor(state ButtonState) color.RGBA {
	switch state {
	case ButtonActive:
		return color.RGBA{R: 0x81, G: 0xC7, B: 0x84, A: 0xFF} // light green
	default:
		return asset.ColorUIBorder
	}
}

// drawRectBorder draws a 1-pixel border around the given rectangle.
func drawRectBorder(screen *ebiten.Image, r image.Rectangle, c color.RGBA) {
	w := r.Dx()
	h := r.Dy()

	hLine := ebiten.NewImage(w, 1)
	hLine.Fill(c)
	vLine := ebiten.NewImage(1, h)
	vLine.Fill(c)

	// Top.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	screen.DrawImage(hLine, op)

	// Bottom.
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(r.Min.X), float64(r.Max.Y-1))
	screen.DrawImage(hLine, op2)

	// Left.
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	screen.DrawImage(vLine, op3)

	// Right.
	op4 := &ebiten.DrawImageOptions{}
	op4.GeoM.Translate(float64(r.Max.X-1), float64(r.Min.Y))
	screen.DrawImage(vLine, op4)
}
