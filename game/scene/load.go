package scene

import (
	"image"
	"time"
)

// LoadEntry holds metadata about a save file for display in the load screen.
type LoadEntry struct {
	Path       string
	Filename   string
	SavedAt    time.Time
	ScenarioID string
}

// LoadScene displays a list of save files for the player to load.
type LoadScene struct {
	entries  []LoadEntry
	btnRects []image.Rectangle
	backRect image.Rectangle
	onSelect func(entry LoadEntry)
	onBack   func()

	screenWidth  int
	screenHeight int

	drawFunc func(screen image.Image, ls *LoadScene)
}

// NewLoadScene creates a LoadScene. onSelect is called with the chosen
// LoadEntry. onBack returns to the title screen. drawFunc handles
// rendering (nil is safe — Draw becomes a no-op).
func NewLoadScene(screenWidth, screenHeight int, entries []LoadEntry, onSelect func(LoadEntry), onBack func(), drawFunc func(image.Image, *LoadScene)) *LoadScene {
	btnW := 500
	btnH := 40
	cx := (screenWidth - btnW) / 2
	baseY := 120

	// Limit visible entries to avoid overflowing screen.
	maxVisible := (screenHeight - baseY - 100) / (btnH + 8)
	visible := entries
	if len(visible) > maxVisible {
		visible = visible[:maxVisible]
	}

	rects := make([]image.Rectangle, len(visible))
	for i := range visible {
		y := baseY + i*(btnH+8)
		rects[i] = image.Rect(cx, y, cx+btnW, y+btnH)
	}

	backW := 120
	backH := 32
	backRect := image.Rect((screenWidth-backW)/2, screenHeight-80, (screenWidth+backW)/2, screenHeight-80+backH)

	return &LoadScene{
		entries:      visible,
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
func (ls *LoadScene) Update() error {
	return nil
}

// Draw renders the load screen via the injected drawFunc.
func (ls *LoadScene) Draw(screen image.Image) {
	if ls.drawFunc != nil {
		ls.drawFunc(screen, ls)
	}
}

// HandleClick checks if the click at (px, py) hits a button and fires the callback.
// Returns true if a button was hit.
func (ls *LoadScene) HandleClick(px, py int) bool {
	pt := image.Pt(px, py)
	for i, r := range ls.btnRects {
		if pt.In(r) && ls.onSelect != nil {
			ls.onSelect(ls.entries[i])
			return true
		}
	}
	if pt.In(ls.backRect) && ls.onBack != nil {
		ls.onBack()
		return true
	}
	return false
}

// OnEnter is called when transitioning to this scene.
func (ls *LoadScene) OnEnter() {}

// OnExit is called when transitioning away from this scene.
func (ls *LoadScene) OnExit() {}

// Entries returns the load entries.
func (ls *LoadScene) Entries() []LoadEntry {
	return ls.entries
}

// ButtonRects returns the save entry button rectangles for testing/rendering.
func (ls *LoadScene) ButtonRects() []image.Rectangle {
	return ls.btnRects
}

// BackRect returns the back button rectangle for testing/rendering.
func (ls *LoadScene) BackRect() image.Rectangle {
	return ls.backRect
}

// ScreenWidth returns the screen width.
func (ls *LoadScene) ScreenWidth() int {
	return ls.screenWidth
}

// ScreenHeight returns the screen height.
func (ls *LoadScene) ScreenHeight() int {
	return ls.screenHeight
}

// FormatEntryLabel returns a display label for a LoadEntry.
func FormatEntryLabel(e LoadEntry) string {
	ts := e.SavedAt.Format("2006-01-02 15:04:05")
	if e.ScenarioID != "" {
		return ts + "  [" + e.ScenarioID + "]"
	}
	return ts + "  " + e.Filename
}
