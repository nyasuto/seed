package scene

import (
	"fmt"
	"image"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// ResultData holds the information displayed on the result screen.
type ResultData struct {
	Won           bool
	Reason        string
	TotalTicks    int
	RoomCount     int
	DefeatedWaves int
	TotalWaves    int
	FinalCoreHP   int
}

// ResultLines returns formatted text lines for the result screen.
func (d ResultData) ResultLines() []string {
	var lines []string
	if d.Won {
		lines = append(lines, "VICTORY")
	} else {
		lines = append(lines, "DEFEAT")
	}
	lines = append(lines, "")
	lines = append(lines, d.Reason)
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Total Ticks: %d", d.TotalTicks))
	lines = append(lines, fmt.Sprintf("Rooms Built: %d", d.RoomCount))
	lines = append(lines, fmt.Sprintf("Waves Defeated: %d / %d", d.DefeatedWaves, d.TotalWaves))
	lines = append(lines, fmt.Sprintf("Final Core HP: %d", d.FinalCoreHP))
	return lines
}

// BuildResultData creates ResultData from the game result and snapshot.
func BuildResultData(result simulation.GameResult, snap scenario.GameSnapshot) ResultData {
	return ResultData{
		Won:           result.Status == simulation.Won,
		Reason:        result.Reason,
		TotalTicks:    int(result.FinalTick),
		RoomCount:     snap.RoomCount,
		DefeatedWaves: snap.DefeatedWaves,
		TotalWaves:    snap.TotalWaves,
		FinalCoreHP:   snap.CoreHP,
	}
}

// ResultScene displays the game result with statistics and navigation buttons.
type ResultScene struct {
	data         ResultData
	screenWidth  int
	screenHeight int

	retryRect image.Rectangle
	titleRect image.Rectangle

	onRetry func()
	onTitle func()

	drawFunc func(screen image.Image, rs *ResultScene)
}

// NewResultScene creates a ResultScene.
// onRetry is called when "Retry" is clicked. onTitle is called when "Title" is clicked.
// drawFunc handles rendering (nil is safe — Draw becomes a no-op).
func NewResultScene(screenWidth, screenHeight int, data ResultData, onRetry, onTitle func(), drawFunc func(image.Image, *ResultScene)) *ResultScene {
	btnW := 180
	btnH := 36
	cx := (screenWidth - btnW) / 2
	baseY := screenHeight/2 + 80

	return &ResultScene{
		data:         data,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		retryRect:    image.Rect(cx, baseY, cx+btnW, baseY+btnH),
		titleRect:    image.Rect(cx, baseY+btnH+12, cx+btnW, baseY+btnH+12+btnH),
		onRetry:      onRetry,
		onTitle:      onTitle,
		drawFunc:     drawFunc,
	}
}

// Update is a no-op. Input is handled via HandleClick called from the host.
func (rs *ResultScene) Update() error {
	return nil
}

// Draw renders the result screen via the injected drawFunc.
func (rs *ResultScene) Draw(screen image.Image) {
	if rs.drawFunc != nil {
		rs.drawFunc(screen, rs)
	}
}

// HandleClick checks if the click at (px, py) hits a button and fires the callback.
// Returns true if a button was hit.
func (rs *ResultScene) HandleClick(px, py int) bool {
	pt := image.Pt(px, py)
	if pt.In(rs.retryRect) && rs.onRetry != nil {
		rs.onRetry()
		return true
	}
	if pt.In(rs.titleRect) && rs.onTitle != nil {
		rs.onTitle()
		return true
	}
	return false
}

// OnEnter is called when transitioning to this scene.
func (rs *ResultScene) OnEnter() {}

// OnExit is called when transitioning away from this scene.
func (rs *ResultScene) OnExit() {}

// Data returns the result data for rendering.
func (rs *ResultScene) Data() ResultData {
	return rs.data
}

// RetryRect returns the "Retry" button rectangle for testing/rendering.
func (rs *ResultScene) RetryRect() image.Rectangle {
	return rs.retryRect
}

// TitleRect returns the "Title" button rectangle for testing/rendering.
func (rs *ResultScene) TitleRect() image.Rectangle {
	return rs.titleRect
}

// ScreenWidth returns the screen width.
func (rs *ResultScene) ScreenWidth() int {
	return rs.screenWidth
}

// ScreenHeight returns the screen height.
func (rs *ResultScene) ScreenHeight() int {
	return rs.screenHeight
}
