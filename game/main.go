// Package main is the entry point for the ChaosForge GUI client.
package main

import (
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/game/asset"
	"github.com/nyasuto/seed/game/controller"
	"github.com/nyasuto/seed/game/input"
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

// Game implements ebiten.Game interface.
type Game struct {
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

	errorMessage string
	errorTimer   int
}

// NewGame creates a Game with a tutorial scenario loaded.
func NewGame() (*Game, error) {
	ctrl, err := controller.NewGameController(tutorialJSON, 42)
	if err != nil {
		return nil, err
	}

	mv := view.NewMapView(mapOffsetX, mapOffsetY)
	cave := ctrl.Engine().State.Cave

	return &Game{
		ctrl:         ctrl,
		provider:     asset.NewPlaceholderProvider(),
		mapView:      mv,
		entity:       view.NewEntityRenderer(mv),
		topBar:       &view.TopBar{},
		mouse:        input.NewMouseTracker(mv, cave.Grid.Width, cave.Grid.Height),
		tooltip:      &view.Tooltip{},
		stateMachine: input.NewInputStateMachine(),
		actionBar:    view.NewActionBar(screenHeight),
	}, nil
}

// Update proceeds the game state.
func (g *Game) Update() error {
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
			g.finishDigRoomElement(elem)
			return nil
		}
		// Click outside panel while panel is open — cancel.
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || !g.elemPanelContains(px, py) {
			g.cancelDigRoom()
			return nil
		}
	}

	g.stateMachine.Update()
	currentMode := g.stateMachine.Mode()

	// Reset flows if mode changed externally (e.g. Escape, or switching to another mode).
	if prevMode != currentMode {
		g.resetDigRoomFlow()
		g.resetDigCorridorFlow()
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

	// Decrement error timer.
	if g.errorTimer > 0 {
		g.errorTimer--
		if g.errorTimer == 0 {
			g.errorMessage = ""
		}
	}

	return nil
}

func (g *Game) handleClick(px, py int) {
	// Check ActionBar first.
	if actionHit, mode, tickHit, tmode := g.actionBar.HandleClick(px, py); actionHit {
		g.stateMachine.SetMode(mode)
		g.resetDigRoomFlow()
		g.resetDigCorridorFlow()
		if mode == input.ModeDigRoom {
			g.digRoomFlow = input.NewDigRoomFlow()
		}
		if mode == input.ModeDigCorridor {
			g.digCorridorFlow = input.NewDigCorridorFlow()
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
}

func (g *Game) finishDigRoomElement(elem types.Element) {
	if g.digRoomFlow == nil {
		return
	}
	registry := g.ctrl.Engine().State.RoomTypeRegistry
	action, err := g.digRoomFlow.SelectElement(elem, registry)
	if err != nil {
		g.showError(err.Error())
		return
	}
	g.ctrl.AddAction(action)
	g.stateMachine.SetMode(input.ModeNormal)
	g.resetDigRoomFlow()
}

func (g *Game) cancelDigRoom() {
	g.stateMachine.SetMode(input.ModeNormal)
	g.resetDigRoomFlow()
}

func (g *Game) resetDigRoomFlow() {
	g.digRoomFlow = nil
	g.elemPanel = nil
}

func (g *Game) resetDigCorridorFlow() {
	g.digCorridorFlow = nil
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
	g.errorMessage = msg
	g.errorTimer = 180 // ~3 seconds at 60 FPS
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	state := g.ctrl.Engine().State
	cave := state.Cave
	registry := state.RoomTypeRegistry

	// Map tiles.
	roomInfo := view.BuildRoomRenderMap(cave, registry)
	g.mapView.Draw(screen, cave, roomInfo, g.provider)

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

	// Error message.
	if g.errorMessage != "" {
		view.DrawText(screen, g.errorMessage, 10, screenHeight-actionBarHeight-20)
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
