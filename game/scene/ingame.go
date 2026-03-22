package scene

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/game/asset"
	"github.com/nyasuto/seed/game/controller"
	"github.com/nyasuto/seed/game/input"
	"github.com/nyasuto/seed/game/view"
)

const (
	inGameActionBarHeight = 36
)

// InGameScene integrates all in-game components: GameController, MapView,
// EntityRenderer, TopBar, ActionBar, InputStateMachine, and interaction flows.
type InGameScene struct {
	ctrl     *controller.GameController
	provider asset.TilesetProvider
	mapView  *view.MapView
	entity   *view.EntityRenderer
	topBar   *view.TopBar
	mouse    *input.MouseTracker
	tooltip  *view.Tooltip

	stateMachine *input.InputStateMachine
	actionBar    *view.ActionBar

	digRoomFlow     *input.DigRoomFlow
	elemPanel       *view.ElementPanel
	digCorridorFlow *input.DigCorridorFlow
	summonFlow      *input.SummonBeastFlow
	upgradeFlow     *input.UpgradeRoomFlow

	feedback  *view.FeedbackOverlay
	infoPanel *view.InfoPanel

	screenWidth  int
	screenHeight int

	// onGameOver is called when the game ends, passing the result.
	onGameOver func(simulation.GameResult, scenario.GameSnapshot)

	// gameOverNotified tracks whether onGameOver has been called already.
	gameOverNotified bool
}

// InGameConfig holds configuration for creating an InGameScene.
type InGameConfig struct {
	Controller   *controller.GameController
	ScreenWidth  int
	ScreenHeight int
	MapOffsetX   int
	MapOffsetY   int
	OnGameOver   func(simulation.GameResult, scenario.GameSnapshot)
}

// NewInGameScene creates an InGameScene with all Phase 1-2 components.
func NewInGameScene(cfg InGameConfig) *InGameScene {
	ctrl := cfg.Controller
	mv := view.NewMapView(cfg.MapOffsetX, cfg.MapOffsetY)
	cave := ctrl.Engine().State.Cave

	return &InGameScene{
		ctrl:         ctrl,
		provider:     asset.NewPlaceholderProvider(),
		mapView:      mv,
		entity:       view.NewEntityRenderer(mv),
		topBar:       &view.TopBar{},
		mouse:        input.NewMouseTracker(mv, cave.Grid.Width, cave.Grid.Height),
		tooltip:      &view.Tooltip{},
		stateMachine: input.NewInputStateMachine(),
		actionBar:    view.NewActionBar(cfg.ScreenHeight),
		feedback:     view.NewFeedbackOverlay(),
		infoPanel:    view.NewInfoPanel(),
		screenWidth:  cfg.ScreenWidth,
		screenHeight: cfg.ScreenHeight,
		onGameOver:   cfg.OnGameOver,
	}
}

// OnEnter is called when the scene becomes active.
func (s *InGameScene) OnEnter() {}

// OnExit is called when the scene is deactivated.
func (s *InGameScene) OnExit() {}

// Update handles input processing and controller updates each frame.
func (s *InGameScene) Update() error {
	s.mouse.Update()

	if s.ctrl.State() == controller.GameOver {
		_, _ = s.ctrl.UpdateTick()
		if !s.gameOverNotified && s.onGameOver != nil {
			s.gameOverNotified = true
			s.onGameOver(s.ctrl.Result(), s.ctrl.Snapshot())
		}
		return nil
	}

	prevMode := s.stateMachine.Mode()

	// Handle element panel clicks before state machine update (panel blocks other input).
	if s.elemPanel != nil && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		px, py := ebiten.CursorPosition()
		if elem, ok := s.elemPanel.HandleClick(px, py); ok {
			s.finishElementSelection(elem)
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || !s.elemPanelContains(px, py) {
			s.cancelElementPanel()
			return nil
		}
	}

	s.stateMachine.Update()
	currentMode := s.stateMachine.Mode()

	// Reset flows if mode changed externally.
	if prevMode != currentMode {
		s.resetDigRoomFlow()
		s.resetDigCorridorFlow()
		s.resetSummonFlow()
		s.resetUpgradeFlow()
	}

	// Handle mouse clicks.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		px, py := ebiten.CursorPosition()
		s.handleClick(px, py)
	}

	// Tick control keys (only when not in element panel).
	if s.elemPanel == nil {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			_, _ = s.ctrl.AdvanceTick()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyW) && currentMode == input.ModeNormal {
			_, _ = s.ctrl.AdvanceTick()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF) {
			s.ctrl.StartFastForward(10)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			if s.ctrl.State() == controller.FastForward {
				s.ctrl.StopFastForward()
			}
		}
	}

	_, _ = s.ctrl.UpdateTick()
	s.feedback.Update()

	// Check game over after tick advancement.
	if s.ctrl.State() == controller.GameOver && !s.gameOverNotified && s.onGameOver != nil {
		s.gameOverNotified = true
		s.onGameOver(s.ctrl.Result(), s.ctrl.Snapshot())
	}

	return nil
}

// Draw renders the in-game scene.
func (s *InGameScene) Draw(screen image.Image) {
	dst := screen.(*ebiten.Image)

	state := s.ctrl.Engine().State
	cave := state.Cave
	registry := state.RoomTypeRegistry

	// Map tiles.
	roomInfo := view.BuildRoomRenderMap(cave, registry)
	s.mapView.Draw(dst, cave, roomInfo, s.provider)

	// Cell highlights.
	currentMode := s.stateMachine.Mode()
	if currentMode != input.ModeNormal {
		for y := 0; y < cave.Grid.Height; y++ {
			for x := 0; x < cave.Grid.Width; x++ {
				cell, err := cave.Grid.At(types.Pos{X: x, Y: y})
				if err != nil {
					continue
				}
				hl := view.CellHighlightFor(currentMode, cell.Type)
				if hl != view.CellNone {
					px, py := s.mapView.CellToScreen(x, y)
					view.DrawCellOverlay(dst, px, py, hl)
				}
			}
		}
	}

	// Entities.
	s.entity.DrawBeasts(dst, cave, state.Beasts, s.provider)
	s.entity.DrawInvaders(dst, cave, state.Waves, s.provider)

	// Top bar.
	snap := s.ctrl.Snapshot()
	maxCoreHP := findMaxCoreHP(state)
	s.topBar.Draw(dst, view.TopBarData{
		ChiPool:    int(snap.ChiPoolBalance),
		MaxChiPool: int(state.EconomyEngine.ChiPool.Cap),
		CoreHP:     snap.CoreHP,
		MaxCoreHP:  maxCoreHP,
		Tick:       int(snap.Tick),
	})

	// Action bar.
	tickMode := s.currentTickMode()
	s.actionBar.Draw(dst, s.stateMachine.Mode(), tickMode)

	// Pending actions display.
	if pending := s.ctrl.PendingActions(); len(pending) > 0 {
		queueText := fmt.Sprintf("Queued: %d action(s)", len(pending))
		view.DrawText(dst, queueText, 10, s.actionBar.BarY()-16)
	}

	// Info panel.
	var infoPanelData view.InfoPanelData
	if roomID := s.infoPanel.SelectedRoomID(); roomID > 0 {
		infoPanelData = view.BuildInfoForRoom(roomID, state)
	} else {
		infoPanelData = view.BuildGameInfo(state, snap)
	}
	s.infoPanel.Draw(dst, infoPanelData)

	// Element selection panel.
	if s.elemPanel != nil {
		s.elemPanel.Draw(dst)
	}

	// Tooltip.
	if s.elemPanel == nil {
		cx, cy, ok := s.mouse.CursorCell()
		if ok {
			info := view.BuildTooltipInfo(cave, registry, cx, cy)
			px, py := ebiten.CursorPosition()
			s.tooltip.Draw(dst, info, px, py)
		}
	}

	// Error message.
	s.feedback.DrawError(dst, 10, s.screenHeight-inGameActionBarHeight-20)

	// Mode label near cursor.
	if s.elemPanel == nil {
		curX, curY := ebiten.CursorPosition()
		view.DrawModeLabel(dst, s.stateMachine.Mode(), curX, curY)
	}

	// Game over overlay.
	if s.ctrl.State() == controller.GameOver {
		result := s.ctrl.Result()
		view.DrawText(dst, "GAME OVER: "+result.Status.String(), s.screenWidth/2-60, s.screenHeight/2)
	}
}

func (s *InGameScene) handleClick(px, py int) {
	// Check ActionBar first.
	if actionHit, mode, tickHit, tmode := s.actionBar.HandleClick(px, py); actionHit {
		s.stateMachine.SetMode(mode)
		s.resetDigRoomFlow()
		s.resetDigCorridorFlow()
		s.resetSummonFlow()
		s.resetUpgradeFlow()
		switch mode {
		case input.ModeDigRoom:
			s.digRoomFlow = input.NewDigRoomFlow()
		case input.ModeDigCorridor:
			s.digCorridorFlow = input.NewDigCorridorFlow()
		case input.ModeSummon:
			s.summonFlow = input.NewSummonBeastFlow()
		case input.ModeUpgrade:
			s.upgradeFlow = input.NewUpgradeRoomFlow()
		case input.ModeNormal:
			_, _ = s.ctrl.AdvanceTick()
		}
		return
	} else if tickHit {
		switch tmode {
		case view.TickManual:
			_, _ = s.ctrl.AdvanceTick()
		case view.TickFastForward:
			s.ctrl.StartFastForward(10)
		case view.TickPaused:
			if s.ctrl.State() == controller.FastForward {
				s.ctrl.StopFastForward()
			}
		}
		return
	}

	// Handle DigRoom cell selection.
	if s.stateMachine.Mode() == input.ModeDigRoom && s.digRoomFlow == nil {
		s.digRoomFlow = input.NewDigRoomFlow()
	}

	if s.digRoomFlow != nil && s.digRoomFlow.Step() == input.StepSelectCell {
		cx, cy, ok := s.mouse.CursorCell()
		if !ok {
			return
		}
		cave := s.ctrl.Engine().State.Cave
		cell, err := cave.Grid.At(types.Pos{X: cx, Y: cy})
		if err != nil {
			return
		}
		if selErr := s.digRoomFlow.TrySelectCell(cx, cy, cell.Type); selErr != nil {
			s.showError(selErr.Error())
			return
		}
		s.elemPanel = view.NewElementPanel(s.screenWidth/2, s.screenHeight/2)
		return
	}

	// Handle DigCorridor room selection.
	if s.stateMachine.Mode() == input.ModeDigCorridor && s.digCorridorFlow == nil {
		s.digCorridorFlow = input.NewDigCorridorFlow()
	}

	if s.digCorridorFlow != nil && s.digCorridorFlow.Step() != input.CorridorStepComplete {
		cx, cy, ok := s.mouse.CursorCell()
		if !ok {
			return
		}
		cave := s.ctrl.Engine().State.Cave
		cell, err := cave.Grid.At(types.Pos{X: cx, Y: cy})
		if err != nil {
			return
		}
		action, selErr := s.digCorridorFlow.TrySelectRoom(cell.Type, cell.RoomID)
		if selErr != nil {
			s.showError(selErr.Error())
			return
		}
		if action != nil {
			s.ctrl.AddAction(action)
			s.stateMachine.SetMode(input.ModeNormal)
			s.resetDigCorridorFlow()
		}
	}

	// Handle SummonBeast room selection.
	if s.stateMachine.Mode() == input.ModeSummon && s.summonFlow == nil {
		s.summonFlow = input.NewSummonBeastFlow()
	}

	if s.summonFlow != nil && s.summonFlow.Step() == input.SummonStepSelectRoom {
		cx, cy, ok := s.mouse.CursorCell()
		if !ok {
			return
		}
		state := s.ctrl.Engine().State
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
		if selErr := s.summonFlow.TrySelectRoom(cell.Type, cell.RoomID, hasCapacity); selErr != nil {
			s.showError(selErr.Error())
			return
		}
		s.elemPanel = view.NewElementPanel(s.screenWidth/2, s.screenHeight/2)
		return
	}

	// Handle UpgradeRoom room selection.
	if s.stateMachine.Mode() == input.ModeUpgrade && s.upgradeFlow == nil {
		s.upgradeFlow = input.NewUpgradeRoomFlow()
	}

	if s.upgradeFlow != nil && !s.upgradeFlow.Complete() {
		cx, cy, ok := s.mouse.CursorCell()
		if !ok {
			return
		}
		cave := s.ctrl.Engine().State.Cave
		cell, err := cave.Grid.At(types.Pos{X: cx, Y: cy})
		if err != nil {
			return
		}
		action, selErr := s.upgradeFlow.TrySelectRoom(cell.Type, cell.RoomID)
		if selErr != nil {
			s.showError(selErr.Error())
			return
		}
		s.ctrl.AddAction(action)
		s.stateMachine.SetMode(input.ModeNormal)
		s.resetUpgradeFlow()
		return
	}

	// ModeNormal: click on room cell to show info panel.
	if s.stateMachine.Mode() == input.ModeNormal {
		cx, cy, ok := s.mouse.CursorCell()
		if !ok {
			return
		}
		cave := s.ctrl.Engine().State.Cave
		cell, err := cave.Grid.At(types.Pos{X: cx, Y: cy})
		if err != nil {
			return
		}
		if cell.RoomID > 0 {
			s.infoPanel.SelectRoom(cell.RoomID)
		} else {
			s.infoPanel.ClearSelection()
		}
	}
}

func (s *InGameScene) finishElementSelection(elem types.Element) {
	s.elemPanel = nil
	if s.digRoomFlow != nil {
		registry := s.ctrl.Engine().State.RoomTypeRegistry
		action, err := s.digRoomFlow.SelectElement(elem, registry)
		if err != nil {
			s.showError(err.Error())
			return
		}
		s.ctrl.AddAction(action)
		s.stateMachine.SetMode(input.ModeNormal)
		s.resetDigRoomFlow()
		return
	}
	if s.summonFlow != nil {
		action := s.summonFlow.SelectElement(elem)
		s.ctrl.AddAction(action)
		s.stateMachine.SetMode(input.ModeNormal)
		s.resetSummonFlow()
		return
	}
}

func (s *InGameScene) cancelElementPanel() {
	s.elemPanel = nil
	s.stateMachine.SetMode(input.ModeNormal)
	s.resetDigRoomFlow()
	s.resetSummonFlow()
}

func (s *InGameScene) resetDigRoomFlow() {
	s.digRoomFlow = nil
	s.elemPanel = nil
}

func (s *InGameScene) resetDigCorridorFlow() {
	s.digCorridorFlow = nil
}

func (s *InGameScene) resetSummonFlow() {
	s.summonFlow = nil
	if s.digRoomFlow == nil {
		s.elemPanel = nil
	}
}

func (s *InGameScene) resetUpgradeFlow() {
	s.upgradeFlow = nil
}

func (s *InGameScene) elemPanelContains(px, py int) bool {
	cx := s.screenWidth / 2
	cy := s.screenHeight / 2
	hw := 120
	hh := 80
	return px >= cx-hw && px <= cx+hw && py >= cy-hh && py <= cy+hh
}

func (s *InGameScene) showError(msg string) {
	s.feedback.ShowError(msg)
}

func (s *InGameScene) currentTickMode() view.TickMode {
	switch s.ctrl.State() {
	case controller.FastForward:
		return view.TickFastForward
	case controller.Paused:
		return view.TickPaused
	default:
		return view.TickManual
	}
}

// Controller returns the GameController for external access (e.g., save/load).
func (s *InGameScene) Controller() *controller.GameController {
	return s.ctrl
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
