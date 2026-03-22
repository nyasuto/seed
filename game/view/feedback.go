package view

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
	"github.com/nyasuto/seed/game/input"
)

const (
	// ErrorDuration is the number of frames an error message is displayed (~3s at 60 FPS).
	ErrorDuration = 180
	// errorFadeStart is the frame count at which fade-out begins.
	errorFadeStart = 60
)

// CellHighlight describes how a cell should be visually modified in the current mode.
type CellHighlight int

const (
	// CellNone means no special highlight.
	CellNone CellHighlight = iota
	// CellValid means the cell is a valid target — draw with a bright tint.
	CellValid
	// CellInvalid means the cell is an invalid target — draw darkened.
	CellInvalid
)

// FeedbackOverlay manages transient visual feedback: error messages,
// mode labels near the cursor, and cell highlights.
type FeedbackOverlay struct {
	errorMsg   string
	errorTimer int
}

// NewFeedbackOverlay creates a new FeedbackOverlay.
func NewFeedbackOverlay() *FeedbackOverlay {
	return &FeedbackOverlay{}
}

// ShowError sets an error message to display for ErrorDuration frames.
func (fo *FeedbackOverlay) ShowError(msg string) {
	fo.errorMsg = msg
	fo.errorTimer = ErrorDuration
}

// Update decrements the error timer each frame.
func (fo *FeedbackOverlay) Update() {
	if fo.errorTimer > 0 {
		fo.errorTimer--
		if fo.errorTimer == 0 {
			fo.errorMsg = ""
		}
	}
}

// ErrorMessage returns the current error message, or "" if none.
func (fo *FeedbackOverlay) ErrorMessage() string {
	return fo.errorMsg
}

// ErrorTimer returns the remaining frames for the current error display.
func (fo *FeedbackOverlay) ErrorTimer() int {
	return fo.errorTimer
}

// ErrorAlpha returns the opacity for the error text (1.0 → 0.0 fade-out).
func (fo *FeedbackOverlay) ErrorAlpha() float64 {
	if fo.errorTimer <= 0 {
		return 0
	}
	if fo.errorTimer > errorFadeStart {
		return 1.0
	}
	return float64(fo.errorTimer) / float64(errorFadeStart)
}

// DrawError renders the current error message in red at the given position,
// fading out over the last errorFadeStart frames.
func (fo *FeedbackOverlay) DrawError(screen *ebiten.Image, x, y int) {
	if fo.errorMsg == "" {
		return
	}
	red := color.RGBA{R: 0xFF, G: 0x44, B: 0x44, A: 0xFF}
	DrawColoredText(screen, fo.errorMsg, x, y, red, fo.ErrorAlpha())
}

// ModeLabel returns the Japanese display label for the given ActionMode.
// Returns "" for ModeNormal (no label shown).
func ModeLabel(mode input.ActionMode) string {
	switch mode {
	case input.ModeDigRoom:
		return "掘削モード"
	case input.ModeDigCorridor:
		return "通路モード"
	case input.ModeSummon:
		return "召喚モード"
	case input.ModeUpgrade:
		return "強化モード"
	default:
		return ""
	}
}

// DrawModeLabel renders the mode name near the given cursor position.
func DrawModeLabel(screen *ebiten.Image, mode input.ActionMode, cursorX, cursorY int) {
	label := ModeLabel(mode)
	if label == "" {
		return
	}
	// Offset slightly below and to the right of the cursor.
	DrawText(screen, label, cursorX+16, cursorY+16)
}

// CellHighlightFor determines how a cell should be visually modified given the
// current action mode and cell type.
func CellHighlightFor(mode input.ActionMode, cellType world.CellType) CellHighlight {
	switch mode {
	case input.ModeDigRoom:
		switch cellType {
		case world.Rock:
			return CellValid
		case world.HardRock, world.Water:
			return CellInvalid
		default:
			return CellNone
		}
	case input.ModeDigCorridor, input.ModeSummon, input.ModeUpgrade:
		switch cellType {
		case world.RoomFloor, world.Entrance:
			return CellValid
		default:
			return CellNone
		}
	default:
		return CellNone
	}
}

// DrawCellOverlay draws a semi-transparent overlay on a tile at the given
// screen position based on the highlight type.
func DrawCellOverlay(screen *ebiten.Image, px, py int, highlight CellHighlight) {
	if highlight == CellNone {
		return
	}

	tile := ebiten.NewImage(asset.TileSize, asset.TileSize)
	switch highlight {
	case CellValid:
		// Bright tint overlay.
		tile.Fill(color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x30})
	case CellInvalid:
		// Dark overlay.
		tile.Fill(color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x80})
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(px), float64(py))
	screen.DrawImage(tile, op)
}
