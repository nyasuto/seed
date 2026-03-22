// Package main is the entry point for the ChaosForge GUI client.
package main

import (
	_ "embed"
	"fmt"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/game/asset"
	"github.com/nyasuto/seed/game/controller"
	"github.com/nyasuto/seed/game/input"
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

	ctrl     *controller.GameController
	provider asset.TilesetProvider
	mapView  *view.MapView
	entity   *view.EntityRenderer
	topBar   *view.TopBar
	mouse    *input.MouseTracker
	tooltip  *view.Tooltip

	stateMachine *input.InputStateMachine
	actionBar    *view.ActionBar

	// DigRoom flow state.
	digRoomFlow *input.DigRoomFlow
	elemPanel   *view.ElementPanel

	// DigCorridor flow state.
	digCorridorFlow *input.DigCorridorFlow

	// SummonBeast flow state.
	summonFlow *input.SummonBeastFlow

	// UpgradeRoom flow state.
	upgradeFlow *input.UpgradeRoomFlow

	feedback *view.FeedbackOverlay
}

// inGameProxy wraps Game's existing Update/Draw logic as a Scene.
// This is a transitional structure until full scene extraction in Task 3-C.
type inGameProxy struct {
	game *Game
}

func (p *inGameProxy) Update() error { return p.game.updateInGame() }
func (p *inGameProxy) Draw(screen image.Image) {
	p.game.drawInGame(screen.(*ebiten.Image))
}
func (p *inGameProxy) OnEnter() {}
func (p *inGameProxy) OnExit()  {}

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

// startInGame initializes the GameController and in-game components,
// then switches to the InGame scene.
func (g *Game) startInGame(scenarioJSON []byte) {
	ctrl, err := controller.NewGameController(scenarioJSON, 42)
	if err != nil {
		log.Printf("failed to start game: %v", err)
		return
	}

	mv := view.NewMapView(mapOffsetX, mapOffsetY)
	cave := ctrl.Engine().State.Cave

	g.ctrl = ctrl
	g.provider = asset.NewPlaceholderProvider()
	g.mapView = mv
	g.entity = view.NewEntityRenderer(mv)
	g.topBar = &view.TopBar{}
	g.mouse = input.NewMouseTracker(mv, cave.Grid.Width, cave.Grid.Height)
	g.tooltip = &view.Tooltip{}
	g.stateMachine = input.NewInputStateMachine()
	g.actionBar = view.NewActionBar(screenHeight)
	g.feedback = view.NewFeedbackOverlay()

	g.scenes.Switch(&inGameProxy{game: g})
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

// Update delegates to the active scene via SceneManager.
// For title and select scenes, it also handles mouse click input
// (since those scenes are ebiten-free and expose HandleClick).
func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		px, py := ebiten.CursorPosition()
		switch s := g.scenes.Current().(type) {
		case *scene.TitleScene:
			s.HandleClick(px, py)
		case *scene.ScenarioSelectScene:
			s.HandleClick(px, py)
		}
	}
	return g.scenes.Update()
}

// updateInGame contains the in-game update logic (delegated from inGameProxy).
func (g *Game) updateInGame() error {
	g.mouse.Update()

	if g.ctrl.State() == controller.GameOver {
		g.ctrl.UpdateTick()
		return nil
	}

	prevMode := g.stateMachine.Mode()

	// Handle element panel clicks before state machine update (panel blocks other input).
	if g.elemPanel != nil && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		px, py := ebiten.CursorPosition()
		if elem, ok := g.elemPanel.HandleClick(px, py); ok {
			g.finishElementSelection(elem)
			return nil
		}
		// Click outside panel while panel is open — cancel.
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || !g.elemPanelContains(px, py) {
			g.cancelElementPanel()
			return nil
		}
	}

	g.stateMachine.Update()
	currentMode := g.stateMachine.Mode()

	// Reset flows if mode changed externally (e.g. Escape, or switching to another mode).
	if prevMode != currentMode {
		g.resetDigRoomFlow()
		g.resetDigCorridorFlow()
		g.resetSummonFlow()
		g.resetUpgradeFlow()
	}

	// Handle mouse clicks.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		px, py := ebiten.CursorPosition()
		g.handleClick(px, py)
	}

	// Tick control keys (only when not in element panel).
	if g.elemPanel == nil {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.ctrl.AdvanceTick()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyW) && currentMode == input.ModeNormal {
			// Wait: advance tick with no player action.
			g.ctrl.AdvanceTick()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF) {
			g.ctrl.StartFastForward(10)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			if g.ctrl.State() == controller.FastForward {
				g.ctrl.StopFastForward()
			}
		}
	}

	// Auto-advance in FastForward mode.
	g.ctrl.UpdateTick()

	// Update feedback overlay (error timer, etc.).
	g.feedback.Update()

	return nil
}

func (g *Game) handleClick(px, py int) {
	// Check ActionBar first.
	if actionHit, mode, tickHit, tmode := g.actionBar.HandleClick(px, py); actionHit {
		g.stateMachine.SetMode(mode)
		g.resetDigRoomFlow()
		g.resetDigCorridorFlow()
		g.resetSummonFlow()
		g.resetUpgradeFlow()
		switch mode {
		case input.ModeDigRoom:
			g.digRoomFlow = input.NewDigRoomFlow()
		case input.ModeDigCorridor:
			g.digCorridorFlow = input.NewDigCorridorFlow()
		case input.ModeSummon:
			g.summonFlow = input.NewSummonBeastFlow()
		case input.ModeUpgrade:
			g.upgradeFlow = input.NewUpgradeRoomFlow()
		case input.ModeNormal:
			// "Wait W" button — advance tick with no player action.
			g.ctrl.AdvanceTick()
		}
		return
	} else if tickHit {
		switch tmode {
		case view.TickManual:
			g.ctrl.AdvanceTick()
		case view.TickFastForward:
			g.ctrl.StartFastForward(10)
		case view.TickPaused:
			if g.ctrl.State() == controller.FastForward {
				g.ctrl.StopFastForward()
			}
		}
		return
	}

	// Handle DigRoom cell selection.
	if g.stateMachine.Mode() == input.ModeDigRoom && g.digRoomFlow == nil {
		g.digRoomFlow = input.NewDigRoomFlow()
	}

	if g.digRoomFlow != nil && g.digRoomFlow.Step() == input.StepSelectCell {
		cx, cy, ok := g.mouse.CursorCell()
		if !ok {
			return
		}
		cave := g.ctrl.Engine().State.Cave
		cell, err := cave.Grid.At(types.Pos{X: cx, Y: cy})
		if err != nil {
			return
		}
		if selErr := g.digRoomFlow.TrySelectCell(cx, cy, cell.Type); selErr != nil {
			g.showError(selErr.Error())
			return
		}
		// Cell selected — show element panel.
		g.elemPanel = view.NewElementPanel(screenWidth/2, screenHeight/2)
		return
	}

	// Handle DigCorridor room selection.
	if g.stateMachine.Mode() == input.ModeDigCorridor && g.digCorridorFlow == nil {
		g.digCorridorFlow = input.NewDigCorridorFlow()
	}

	if g.digCorridorFlow != nil && g.digCorridorFlow.Step() != input.CorridorStepComplete {
		cx, cy, ok := g.mouse.CursorCell()
		if !ok {
			return
		}
		cave := g.ctrl.Engine().State.Cave
		cell, err := cave.Grid.At(types.Pos{X: cx, Y: cy})
		if err != nil {
			return
		}
		action, selErr := g.digCorridorFlow.TrySelectRoom(cell.Type, cell.RoomID)
		if selErr != nil {
			g.showError(selErr.Error())
			return
		}
		if action != nil {
			g.ctrl.AddAction(action)
			g.stateMachine.SetMode(input.ModeNormal)
			g.resetDigCorridorFlow()
		}
	}

	// Handle SummonBeast room selection.
	if g.stateMachine.Mode() == input.ModeSummon && g.summonFlow == nil {
		g.summonFlow = input.NewSummonBeastFlow()
	}

	if g.summonFlow != nil && g.summonFlow.Step() == input.SummonStepSelectRoom {
		cx, cy, ok := g.mouse.CursorCell()
		if !ok {
			return
		}
		state := g.ctrl.Engine().State
		cell, err := state.Cave.Grid.At(types.Pos{X: cx, Y: cy})
		if err != nil {
			return
		}
		hasCapacity := false
		if cell.RoomID > 0 {
			room := state.Cave.RoomByID(cell.RoomID)
			if room != nil {
				rt, rtErr := state.RoomTypeRegistry.Get(room.TypeID)
				if rtErr == nil {
					hasCapacity = room.HasBeastCapacity(rt)
				}
			}
		}
		if selErr := g.summonFlow.TrySelectRoom(cell.Type, cell.RoomID, hasCapacity); selErr != nil {
			g.showError(selErr.Error())
			return
		}
		// Room selected — show element panel.
		g.elemPanel = view.NewElementPanel(screenWidth/2, screenHeight/2)
		return
	}

	// Handle UpgradeRoom room selection.
	if g.stateMachine.Mode() == input.ModeUpgrade && g.upgradeFlow == nil {
		g.upgradeFlow = input.NewUpgradeRoomFlow()
	}

	if g.upgradeFlow != nil && !g.upgradeFlow.Complete() {
		cx, cy, ok := g.mouse.CursorCell()
		if !ok {
			return
		}
		cave := g.ctrl.Engine().State.Cave
		cell, err := cave.Grid.At(types.Pos{X: cx, Y: cy})
		if err != nil {
			return
		}
		action, selErr := g.upgradeFlow.TrySelectRoom(cell.Type, cell.RoomID)
		if selErr != nil {
			g.showError(selErr.Error())
			return
		}
		g.ctrl.AddAction(action)
		g.stateMachine.SetMode(input.ModeNormal)
		g.resetUpgradeFlow()
	}
}

func (g *Game) finishElementSelection(elem types.Element) {
	if g.digRoomFlow != nil {
		registry := g.ctrl.Engine().State.RoomTypeRegistry
		action, err := g.digRoomFlow.SelectElement(elem, registry)
		if err != nil {
			g.showError(err.Error())
			return
		}
		g.ctrl.AddAction(action)
		g.stateMachine.SetMode(input.ModeNormal)
		g.resetDigRoomFlow()
		return
	}
	if g.summonFlow != nil {
		action := g.summonFlow.SelectElement(elem)
		g.ctrl.AddAction(action)
		g.stateMachine.SetMode(input.ModeNormal)
		g.resetSummonFlow()
		return
	}
}

func (g *Game) cancelElementPanel() {
	g.stateMachine.SetMode(input.ModeNormal)
	g.resetDigRoomFlow()
	g.resetSummonFlow()
}

func (g *Game) resetDigRoomFlow() {
	g.digRoomFlow = nil
	g.elemPanel = nil
}

func (g *Game) resetDigCorridorFlow() {
	g.digCorridorFlow = nil
}

func (g *Game) resetSummonFlow() {
	g.summonFlow = nil
	if g.digRoomFlow == nil {
		g.elemPanel = nil
	}
}

func (g *Game) resetUpgradeFlow() {
	g.upgradeFlow = nil
}

func (g *Game) elemPanelContains(px, py int) bool {
	// Simple bounds check — element panel covers center of screen.
	cx := screenWidth / 2
	cy := screenHeight / 2
	hw := 120 // elemPanelWidth/2
	hh := 80  // elemPanelHeight/2
	return px >= cx-hw && px <= cx+hw && py >= cy-hh && py <= cy+hh
}

func (g *Game) showError(msg string) {
	g.feedback.ShowError(msg)
}

// Draw delegates to the active scene via SceneManager.
func (g *Game) Draw(screen *ebiten.Image) {
	g.scenes.Draw(image.Image(screen))
}

// drawInGame contains the in-game draw logic (delegated from inGameProxy).
func (g *Game) drawInGame(screen *ebiten.Image) {
	state := g.ctrl.Engine().State
	cave := state.Cave
	registry := state.RoomTypeRegistry

	// Map tiles.
	roomInfo := view.BuildRoomRenderMap(cave, registry)
	g.mapView.Draw(screen, cave, roomInfo, g.provider)

	// Cell highlights (valid/invalid overlay based on current mode).
	currentMode := g.stateMachine.Mode()
	if currentMode != input.ModeNormal {
		for y := 0; y < cave.Grid.Height; y++ {
			for x := 0; x < cave.Grid.Width; x++ {
				cell, err := cave.Grid.At(types.Pos{X: x, Y: y})
				if err != nil {
					continue
				}
				hl := view.CellHighlightFor(currentMode, cell.Type)
				if hl != view.CellNone {
					px, py := g.mapView.CellToScreen(x, y)
					view.DrawCellOverlay(screen, px, py, hl)
				}
			}
		}
	}

	// Entities (beasts + invaders).
	g.entity.DrawBeasts(screen, cave, state.Beasts, g.provider)
	g.entity.DrawInvaders(screen, cave, state.Waves, g.provider)

	// Top bar.
	snap := g.ctrl.Snapshot()
	maxCoreHP := findMaxCoreHP(state)
	g.topBar.Draw(screen, view.TopBarData{
		ChiPool:    int(snap.ChiPoolBalance),
		MaxChiPool: int(state.EconomyEngine.ChiPool.Cap),
		CoreHP:     snap.CoreHP,
		MaxCoreHP:  maxCoreHP,
		Tick:       int(snap.Tick),
	})

	// Action bar.
	tickMode := g.currentTickMode()
	g.actionBar.Draw(screen, g.stateMachine.Mode(), tickMode)

	// Pending actions display (above action bar).
	if pending := g.ctrl.PendingActions(); len(pending) > 0 {
		queueText := fmt.Sprintf("Queued: %d action(s)", len(pending))
		view.DrawText(screen, queueText, 10, g.actionBar.BarY()-16)
	}

	// Element selection panel (blocks other UI).
	if g.elemPanel != nil {
		g.elemPanel.Draw(screen)
	}

	// Tooltip on hover (hide when panel is open).
	if g.elemPanel == nil {
		cx, cy, ok := g.mouse.CursorCell()
		if ok {
			info := view.BuildTooltipInfo(cave, registry, cx, cy)
			px, py := ebiten.CursorPosition()
			g.tooltip.Draw(screen, info, px, py)
		}
	}

	// Error message (red, fading).
	g.feedback.DrawError(screen, 10, screenHeight-actionBarHeight-20)

	// Mode label near cursor.
	if g.elemPanel == nil {
		curX, curY := ebiten.CursorPosition()
		view.DrawModeLabel(screen, g.stateMachine.Mode(), curX, curY)
	}

	// Game over overlay.
	if g.ctrl.State() == controller.GameOver {
		result := g.ctrl.Result()
		view.DrawText(screen, "GAME OVER: "+result.Status.String(), screenWidth/2-60, screenHeight/2)
	}
}

const actionBarHeight = 36

func (g *Game) currentTickMode() view.TickMode {
	switch g.ctrl.State() {
	case controller.FastForward:
		return view.TickFastForward
	case controller.Paused:
		return view.TickPaused
	default:
		return view.TickManual
	}
}

// findMaxCoreHP returns the max core HP from the dragon hole room.
func findMaxCoreHP(state *simulation.GameState) int {
	for _, room := range state.Cave.Rooms {
		rt, err := state.RoomTypeRegistry.Get(room.TypeID)
		if err != nil {
			continue
		}
		if rt.BaseCoreHP > 0 {
			return rt.CoreHPAtLevel(room.Level)
		}
	}
	return 0
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
