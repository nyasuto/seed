package view

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/game/asset"
	"github.com/nyasuto/seed/game/input"
)

const (
	actionBarHeight  = 36
	actionBarPadding = 4
	actionBtnWidth   = 72
	actionBtnHeight  = 28
	actionBtnGap     = 4
	tickBtnWidth     = 48
	tickSectionGap   = 16
)

// TickMode represents the current tick advancement mode for display.
type TickMode int

const (
	// TickManual means ticks advance only on manual input.
	TickManual TickMode = iota
	// TickFastForward means ticks advance automatically.
	TickFastForward
	// TickPaused means tick advancement is paused.
	TickPaused
)

// ActionBar renders the bottom action bar with action buttons and tick controls.
type ActionBar struct {
	actionButtons []Button
	tickButtons   []Button
	barY          int
	initialized   bool
}

// actionDef describes a single action button.
type actionDef struct {
	label string
	mode  input.ActionMode
}

var actionDefs = []actionDef{
	{"Dig D", input.ModeDigRoom},
	{"Path C", input.ModeDigCorridor},
	{"Call S", input.ModeSummon},
	{"Up U", input.ModeUpgrade},
	{"Wait W", input.ModeNormal},
}

// tickDef describes a single tick control button.
type tickDef struct {
	label string
	mode  TickMode
}

var tickDefs = []tickDef{
	{"\u25b6 Spc", TickManual},
	{"\u25b6\u25b6 F", TickFastForward},
	{"\u23f8 Esc", TickPaused},
}

// NewActionBar creates an ActionBar positioned at the bottom of a screen
// with the given height.
func NewActionBar(screenHeight int) *ActionBar {
	ab := &ActionBar{
		barY: screenHeight - actionBarHeight,
	}
	ab.initButtons()
	return ab
}

func (ab *ActionBar) initButtons() {
	x := actionBarPadding
	y := ab.barY + actionBarPadding

	// Action buttons.
	ab.actionButtons = make([]Button, len(actionDefs))
	for i := range actionDefs {
		ab.actionButtons[i] = NewButton(x, y, actionBtnWidth, actionBtnHeight, actionDefs[i].label)
		x += actionBtnWidth + actionBtnGap
	}

	x += tickSectionGap

	// Tick control buttons.
	ab.tickButtons = make([]Button, len(tickDefs))
	for i := range tickDefs {
		ab.tickButtons[i] = NewButton(x, y, tickBtnWidth, actionBtnHeight, tickDefs[i].label)
		x += tickBtnWidth + actionBtnGap
	}

	ab.initialized = true
}

// Draw renders the action bar onto the screen.
func (ab *ActionBar) Draw(screen *ebiten.Image, activeMode input.ActionMode, tickMode TickMode) {
	sw := screen.Bounds().Dx()

	// Background.
	bg := ebiten.NewImage(sw, actionBarHeight)
	bg.Fill(asset.ColorUIBackground)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(ab.barY))
	screen.DrawImage(bg, op)

	// Top border.
	border := ebiten.NewImage(sw, 1)
	border.Fill(asset.ColorUIBorder)
	bop := &ebiten.DrawImageOptions{}
	bop.GeoM.Translate(0, float64(ab.barY))
	screen.DrawImage(border, bop)

	// Action buttons.
	for i, btn := range ab.actionButtons {
		state := ButtonNormal
		if actionDefs[i].mode == activeMode {
			state = ButtonActive
		}
		btn.Draw(screen, state)
	}

	// Tick control buttons.
	for i, btn := range ab.tickButtons {
		state := ButtonNormal
		if tickDefs[i].mode == tickMode {
			state = ButtonActive
		}
		btn.Draw(screen, state)
	}
}

// HandleClick checks if the click at (px, py) hits any button and returns
// the result. Returns actionHit=true with the corresponding ActionMode if
// an action button was clicked. Returns tickHit=true with the corresponding
// TickMode if a tick button was clicked.
func (ab *ActionBar) HandleClick(px, py int) (actionHit bool, mode input.ActionMode, tickHit bool, tmode TickMode) {
	for i, btn := range ab.actionButtons {
		if btn.Contains(px, py) {
			return true, actionDefs[i].mode, false, 0
		}
	}
	for i, btn := range ab.tickButtons {
		if btn.Contains(px, py) {
			return false, 0, true, tickDefs[i].mode
		}
	}
	return false, 0, false, 0
}

// BarY returns the Y coordinate of the action bar top edge.
func (ab *ActionBar) BarY() int {
	return ab.barY
}
