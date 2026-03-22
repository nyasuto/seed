// Package main is the entry point for the ChaosForge GUI client.
package main

import (
	_ "embed"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"image/color"

	"github.com/nyasuto/seed/game/asset"
	"github.com/nyasuto/seed/game/controller"
	"github.com/nyasuto/seed/game/scene"
	"github.com/nyasuto/seed/game/view"
)

const (
	screenWidth  = 1088
	screenHeight = 728
	mapOffsetX   = 32
	mapOffsetY   = 32
)

//go:embed testdata/tutorial.json
var tutorialJSON []byte

//go:embed testdata/standard.json
var standardJSON []byte

// builtinScenarios returns the list of selectable scenario entries.
func builtinScenarios() []scene.ScenarioEntry {
	return []scene.ScenarioEntry{
		{
			ID:          "tutorial",
			Name:        "チュートリアル",
			Description: "基本操作を学ぶための簡単なシナリオ",
			Difficulty:  "easy",
			Data:        tutorialJSON,
		},
		{
			ID:          "standard",
			Name:        "標準シナリオ",
			Description: "中規模マップでの本格的な洞窟経営シナリオ",
			Difficulty:  "normal",
			Data:        standardJSON,
		},
	}
}

// Game implements ebiten.Game interface.
// It delegates Update/Draw to the active Scene via SceneManager.
type Game struct {
	scenes *scene.SceneManager
}

// NewGame creates a Game starting at the title screen.
func NewGame() (*Game, error) {
	g := &Game{}
	g.scenes = scene.NewSceneManager(g.makeTitleScene())
	return g, nil
}

func (g *Game) makeTitleScene() *scene.TitleScene {
	return scene.NewTitleScene(screenWidth, screenHeight,
		func() { g.showScenarioSelect() },
		func() { /* Load stub — Phase 4 */ },
		drawTitleScene,
	)
}

// showScenarioSelect transitions to the scenario selection screen.
func (g *Game) showScenarioSelect() {
	entries := builtinScenarios()
	selectScene := scene.NewScenarioSelectScene(screenWidth, screenHeight, entries,
		func(entry scene.ScenarioEntry) { g.startInGame(entry.Data) },
		func() { g.showTitle() },
		drawSelectScene,
	)
	g.scenes.Switch(selectScene)
}

// showTitle transitions back to the title screen.
func (g *Game) showTitle() {
	g.scenes.Switch(g.makeTitleScene())
}

// showResult transitions to the result screen.
func (g *Game) showResult(result simulation.GameResult, snap scenario.GameSnapshot) {
	data := scene.BuildResultData(result, snap)
	resultScene := scene.NewResultScene(screenWidth, screenHeight, data,
		func() { g.showScenarioSelect() },
		func() { g.showTitle() },
		drawResultScene,
	)
	g.scenes.Switch(resultScene)
}

// startInGame initializes the GameController and switches to the InGame scene.
func (g *Game) startInGame(scenarioJSON []byte) {
	ctrl, err := controller.NewGameController(scenarioJSON, 42)
	if err != nil {
		log.Printf("failed to start game: %v", err)
		return
	}

	inGame := scene.NewInGameScene(scene.InGameConfig{
		Controller:   ctrl,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		MapOffsetX:   mapOffsetX,
		MapOffsetY:   mapOffsetY,
		OnGameOver: func(result simulation.GameResult, snap scenario.GameSnapshot) {
			g.showResult(result, snap)
		},
	})
	g.scenes.Switch(inGame)
}

// drawTitleScene renders the title screen using ebiten drawing primitives.
func drawTitleScene(screen image.Image, ts *scene.TitleScene) {
	dst := screen.(*ebiten.Image)
	dst.Fill(asset.ColorUIBackground)

	sw := ts.ScreenWidth()
	sh := ts.ScreenHeight()

	// Title text.
	title := "ChaosForge"
	tw := view.TextWidth(title)
	view.DrawText(dst, title, (sw-tw)/2, sh/2-60)

	subtitle := "- Feng Shui Corridor Chronicle -"
	stw := view.TextWidth(subtitle)
	view.DrawText(dst, subtitle, (sw-stw)/2, sh/2-36)

	// Buttons.
	ngBtn := view.ButtonFromRect(ts.NewGameRect(), "New Game")
	ldBtn := view.ButtonFromRect(ts.LoadRect(), "Load")

	px, py := ebiten.CursorPosition()
	ngState := view.ButtonNormal
	if ngBtn.Contains(px, py) {
		ngState = view.ButtonHover
	}
	ngBtn.Draw(dst, ngState)
	ldBtn.Draw(dst, view.ButtonNormal)
}

// drawSelectScene renders the scenario selection screen using ebiten drawing primitives.
func drawSelectScene(screen image.Image, ss *scene.ScenarioSelectScene) {
	dst := screen.(*ebiten.Image)
	dst.Fill(asset.ColorUIBackground)

	sw := ss.ScreenWidth()
	px, py := ebiten.CursorPosition()

	// Header.
	header := "Select Scenario"
	hw := view.TextWidth(header)
	view.DrawText(dst, header, (sw-hw)/2, 40)

	// Scenario buttons.
	entries := ss.Entries()
	rects := ss.ButtonRects()
	for i, r := range rects {
		label := entries[i].Name + "  [" + entries[i].Difficulty + "]"
		btn := view.ButtonFromRect(r, label)
		state := view.ButtonNormal
		if btn.Contains(px, py) {
			state = view.ButtonHover
		}
		btn.Draw(dst, state)

		// Description below button.
		desc := entries[i].Description
		dw := view.TextWidth(desc)
		view.DrawColoredText(dst, desc, (sw-dw)/2, r.Max.Y+2, asset.ColorUIBorder, 1.0)
	}

	// Back button.
	backBtn := view.ButtonFromRect(ss.BackRect(), "Back")
	backBtn.Draw(dst, view.ButtonNormal)
}

// drawResultScene renders the result screen using ebiten drawing primitives.
func drawResultScene(screen image.Image, rs *scene.ResultScene) {
	dst := screen.(*ebiten.Image)
	dst.Fill(asset.ColorUIBackground)

	sw := rs.ScreenWidth()
	data := rs.Data()
	lines := data.ResultLines()

	// Draw result lines centered.
	startY := rs.ScreenHeight()/2 - 100
	for i, line := range lines {
		tw := view.TextWidth(line)
		x := (sw - tw) / 2
		y := startY + i*view.LineHeight
		if i == 0 {
			// Title line in color: green for victory, red for defeat.
			c := color.RGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF}
			if !data.Won {
				c = asset.ColorFire
			}
			view.DrawColoredText(dst, line, x, y, c, 1.0)
		} else {
			view.DrawText(dst, line, x, y)
		}
	}

	// Buttons.
	retryBtn := view.ButtonFromRect(rs.RetryRect(), "Retry")
	titleBtn := view.ButtonFromRect(rs.TitleRect(), "Title")

	px, py := ebiten.CursorPosition()
	retryState := view.ButtonNormal
	if retryBtn.Contains(px, py) {
		retryState = view.ButtonHover
	}
	titleState := view.ButtonNormal
	if titleBtn.Contains(px, py) {
		titleState = view.ButtonHover
	}
	retryBtn.Draw(dst, retryState)
	titleBtn.Draw(dst, titleState)
}

// Update delegates to the active scene via SceneManager.
func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		px, py := ebiten.CursorPosition()
		switch s := g.scenes.Current().(type) {
		case *scene.TitleScene:
			s.HandleClick(px, py)
		case *scene.ScenarioSelectScene:
			s.HandleClick(px, py)
		case *scene.ResultScene:
			s.HandleClick(px, py)
		}
	}
	return g.scenes.Update()
}

// Draw delegates to the active scene via SceneManager.
func (g *Game) Draw(screen *ebiten.Image) {
	g.scenes.Draw(image.Image(screen))
}

// Layout returns the game's logical screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("ChaosForge")

	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
