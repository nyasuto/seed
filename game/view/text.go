package view

import (
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

// TextWidth returns the approximate pixel width of a string rendered
// with the debug font.
func TextWidth(str string) int {
	return len(str) * CharWidth
}
