// Package main is the entry point for the ChaosForge GUI client.
package main

import (
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/nyasuto/seed/core/simulation"
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
		ctrl:     ctrl,
		provider: asset.NewPlaceholderProvider(),
		mapView:  mv,
		entity:   view.NewEntityRenderer(mv),
		topBar:   &view.TopBar{},
		mouse:    input.NewMouseTracker(mv, cave.Grid.Width, cave.Grid.Height),
		tooltip:  &view.Tooltip{},
	}, nil
}

// Update proceeds the game state.
func (g *Game) Update() error {
	g.mouse.Update()

	// Keyboard-driven tick control.
	if g.ctrl.State() != controller.GameOver {
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

	return nil
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

	// Tooltip on hover.
	cx, cy, ok := g.mouse.CursorCell()
	if ok {
		info := view.BuildTooltipInfo(cave, registry, cx, cy)
		px, py := ebiten.CursorPosition()
		g.tooltip.Draw(screen, info, px, py)
	}

	// Game over overlay.
	if g.ctrl.State() == controller.GameOver {
		result := g.ctrl.Result()
		view.DrawText(screen, "GAME OVER: "+result.Status.String(), screenWidth/2-60, screenHeight/2)
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
