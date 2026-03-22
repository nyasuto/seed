// Package main is the entry point for the ChaosForge GUI client.
package main

import (
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
	"github.com/nyasuto/seed/game/input"
	"github.com/nyasuto/seed/game/view"
)

const (
	screenWidth  = 1088
	screenHeight = 728
)

//go:embed testdata/tutorial.json
var tutorialJSON []byte

// Game implements ebiten.Game interface.
type Game struct {
	cave     *world.Cave
	registry *world.RoomTypeRegistry
	provider asset.TilesetProvider
	mapView  *view.MapView
	roomInfo map[int]view.RoomRenderInfo
	mouse    *input.MouseTracker
	tooltip  *view.Tooltip
}

// NewGame creates a Game with a tutorial scenario Cave loaded for rendering.
func NewGame() (*Game, error) {
	sc, err := scenario.LoadScenario(tutorialJSON)
	if err != nil {
		return nil, err
	}
	rng := types.NewSeededRNG(42)
	engine, err := simulation.NewSimulationEngine(sc, rng)
	if err != nil {
		return nil, err
	}

	mv := view.NewMapView(32, 32)
	roomInfo := view.BuildRoomRenderMap(engine.State.Cave, engine.State.RoomTypeRegistry)

	return &Game{
		cave:     engine.State.Cave,
		registry: engine.State.RoomTypeRegistry,
		provider: asset.NewPlaceholderProvider(),
		mapView:  mv,
		roomInfo: roomInfo,
		mouse:    input.NewMouseTracker(mv, engine.State.Cave.Grid.Width, engine.State.Cave.Grid.Height),
		tooltip:  &view.Tooltip{},
	}, nil
}

// Update proceeds the game state.
func (g *Game) Update() error {
	g.mouse.Update()
	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	g.mapView.Draw(screen, g.cave, g.roomInfo, g.provider)

	cx, cy, ok := g.mouse.CursorCell()
	if ok {
		info := view.BuildTooltipInfo(g.cave, g.registry, cx, cy)
		px, py := ebiten.CursorPosition()
		g.tooltip.Draw(screen, info, px, py)
	}
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
