package view

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	// CharWidth is the approximate width of a single character in the debug font.
	CharWidth = 6
	// LineHeight is the vertical spacing between lines in the debug font.
	LineHeight = 16
)

// DrawText renders a string at the given screen position using Ebitengine's
// built-in debug font (white text on transparent background).
func DrawText(screen *ebiten.Image, str string, x, y int) {
	ebitenutil.DebugPrintAt(screen, str, x, y)
}

// DrawColoredText renders a string at the given screen position with color
// and alpha tinting. It draws white debug text to a temporary image and
// composites it with the specified color and alpha onto screen.
func DrawColoredText(screen *ebiten.Image, str string, x, y int, c color.RGBA, alpha float64) {
	if str == "" || alpha <= 0 {
		return
	}
	w := TextWidth(str) + CharWidth // small padding
	h := LineHeight
	tmp := ebiten.NewImage(w, h)
	ebitenutil.DebugPrintAt(tmp, str, 0, 0)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	// Apply color tint: debug font is white (1,1,1), scale to target color.
	op.ColorScale.Scale(float32(c.R)/255.0, float32(c.G)/255.0, float32(c.B)/255.0, float32(alpha))
	screen.DrawImage(tmp, op)
}

// TextWidth returns the approximate pixel width of a string rendered
// with the debug font.
func TextWidth(str string) int {
	return len(str) * CharWidth
}
