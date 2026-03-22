package scene

import (
	"image"
)

// TitleScene displays the title screen with "New Game" and "Load" buttons.
type TitleScene struct {
	screenWidth  int
	screenHeight int

	// Button rectangles for hit testing.
	newGameRect image.Rectangle
	loadRect    image.Rectangle

	// hasSaves indicates whether save files exist. When false, the Load
	// button is greyed out and clicks are ignored.
	hasSaves bool

	// onNewGame is called when the player clicks "New Game".
	onNewGame func()
	// onLoad is called when the player clicks "Load".
	onLoad func()

	// drawFunc is injected by the caller to handle ebiten-dependent rendering.
	drawFunc func(screen image.Image, ts *TitleScene)
}

// NewTitleScene creates a TitleScene. onNewGame is called when "New Game"
// is clicked. onLoad is called when "Load" is clicked. hasSaves controls
// whether the Load button is enabled. drawFunc handles rendering (nil is
// safe — Draw becomes a no-op).
func NewTitleScene(screenWidth, screenHeight int, onNewGame, onLoad func(), hasSaves bool, drawFunc func(image.Image, *TitleScene)) *TitleScene {
	btnW := 180
	btnH := 36
	cx := (screenWidth - btnW) / 2
	baseY := screenHeight/2 + 20

	return &TitleScene{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		newGameRect:  image.Rect(cx, baseY, cx+btnW, baseY+btnH),
		loadRect:     image.Rect(cx, baseY+btnH+12, cx+btnW, baseY+btnH+12+btnH),
		hasSaves:     hasSaves,
		onNewGame:    onNewGame,
		onLoad:       onLoad,
		drawFunc:     drawFunc,
	}
}

// Update is a no-op. Input is handled via HandleClick called from the host.
func (ts *TitleScene) Update() error {
	return nil
}

// Draw renders the title screen via the injected drawFunc.
func (ts *TitleScene) Draw(screen image.Image) {
	if ts.drawFunc != nil {
		ts.drawFunc(screen, ts)
	}
}

// HandleClick checks if the click at (px, py) hits a button and fires the callback.
// Returns true if a button was hit.
func (ts *TitleScene) HandleClick(px, py int) bool {
	pt := image.Pt(px, py)
	if pt.In(ts.newGameRect) && ts.onNewGame != nil {
		ts.onNewGame()
		return true
	}
	if pt.In(ts.loadRect) && ts.onLoad != nil && ts.hasSaves {
		ts.onLoad()
		return true
	}
	return false
}

// OnEnter is called when transitioning to this scene.
func (ts *TitleScene) OnEnter() {}

// OnExit is called when transitioning away from this scene.
func (ts *TitleScene) OnExit() {}

// NewGameRect returns the "New Game" button rectangle for testing/rendering.
func (ts *TitleScene) NewGameRect() image.Rectangle {
	return ts.newGameRect
}

// LoadRect returns the "Load" button rectangle for testing/rendering.
func (ts *TitleScene) LoadRect() image.Rectangle {
	return ts.loadRect
}

// ScreenWidth returns the screen width.
func (ts *TitleScene) ScreenWidth() int {
	return ts.screenWidth
}

// ScreenHeight returns the screen height.
func (ts *TitleScene) ScreenHeight() int {
	return ts.screenHeight
}

// HasSaves reports whether save files are available (controls Load button state).
func (ts *TitleScene) HasSaves() bool {
	return ts.hasSaves
}
