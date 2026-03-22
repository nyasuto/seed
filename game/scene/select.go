package scene

import (
	"image"
)

// ScenarioEntry represents a selectable scenario in the scenario selection screen.
type ScenarioEntry struct {
	ID          string
	Name        string
	Description string
	Difficulty  string
	Data        []byte // raw JSON bytes for creating a GameController
}

// ScenarioSelectScene displays a list of available scenarios for the player to choose.
type ScenarioSelectScene struct {
	entries  []ScenarioEntry
	btnRects []image.Rectangle
	backRect image.Rectangle
	onSelect func(entry ScenarioEntry)
	onBack   func()

	screenWidth  int
	screenHeight int

	// drawFunc is injected by the caller to handle ebiten-dependent rendering.
	drawFunc func(screen image.Image, ss *ScenarioSelectScene)
}

// NewScenarioSelectScene creates a ScenarioSelectScene. onSelect is called
// with the chosen ScenarioEntry. onBack returns to the title screen.
// drawFunc handles rendering (nil is safe — Draw becomes a no-op).
func NewScenarioSelectScene(screenWidth, screenHeight int, entries []ScenarioEntry, onSelect func(ScenarioEntry), onBack func(), drawFunc func(image.Image, *ScenarioSelectScene)) *ScenarioSelectScene {
	btnW := 400
	btnH := 48
	cx := (screenWidth - btnW) / 2
	baseY := 120

	rects := make([]image.Rectangle, len(entries))
	for i := range entries {
		y := baseY + i*(btnH+12)
		rects[i] = image.Rect(cx, y, cx+btnW, y+btnH)
	}

	backW := 120
	backH := 32
	backRect := image.Rect((screenWidth-backW)/2, screenHeight-80, (screenWidth+backW)/2, screenHeight-80+backH)

	return &ScenarioSelectScene{
		entries:      entries,
		btnRects:     rects,
		backRect:     backRect,
		onSelect:     onSelect,
		onBack:       onBack,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		drawFunc:     drawFunc,
	}
}

// Update is a no-op. Input is handled via HandleClick called from the host.
func (ss *ScenarioSelectScene) Update() error {
	return nil
}

// Draw renders the scenario selection screen via the injected drawFunc.
func (ss *ScenarioSelectScene) Draw(screen image.Image) {
	if ss.drawFunc != nil {
		ss.drawFunc(screen, ss)
	}
}

// HandleClick checks if the click at (px, py) hits a button and fires the callback.
// Returns true if a button was hit.
func (ss *ScenarioSelectScene) HandleClick(px, py int) bool {
	pt := image.Pt(px, py)
	for i, r := range ss.btnRects {
		if pt.In(r) && ss.onSelect != nil {
			ss.onSelect(ss.entries[i])
			return true
		}
	}
	if pt.In(ss.backRect) && ss.onBack != nil {
		ss.onBack()
		return true
	}
	return false
}

// OnEnter is called when transitioning to this scene.
func (ss *ScenarioSelectScene) OnEnter() {}

// OnExit is called when transitioning away from this scene.
func (ss *ScenarioSelectScene) OnExit() {}

// Entries returns the scenario entries.
func (ss *ScenarioSelectScene) Entries() []ScenarioEntry {
	return ss.entries
}

// ButtonRects returns the scenario button rectangles for testing/rendering.
func (ss *ScenarioSelectScene) ButtonRects() []image.Rectangle {
	return ss.btnRects
}

// BackRect returns the back button rectangle for testing/rendering.
func (ss *ScenarioSelectScene) BackRect() image.Rectangle {
	return ss.backRect
}

// ScreenWidth returns the screen width.
func (ss *ScenarioSelectScene) ScreenWidth() int {
	return ss.screenWidth
}

// ScreenHeight returns the screen height.
func (ss *ScenarioSelectScene) ScreenHeight() int {
	return ss.screenHeight
}
