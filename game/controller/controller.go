package controller

import (
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// GameState represents the current state of the game controller.
type GameState int

const (
	// Playing indicates the game is actively running.
	Playing GameState = iota
	// Paused indicates the game is paused (no tick advancement).
	Paused
	// FastForward indicates the game is auto-advancing multiple ticks per frame.
	FastForward
	// GameOver indicates the game has ended (win or loss).
	GameOver
)

// GameController wraps the core SimulationEngine, managing snapshot updates,
// pending action queuing, and game state transitions for the GUI client.
type GameController struct {
	engine   *simulation.SimulationEngine
	snapshot scenario.GameSnapshot
	pending  []simulation.PlayerAction
	state    GameState
	result   simulation.GameResult
}

// NewGameController creates a GameController by loading the given scenario JSON
// and initializing a SimulationEngine with the provided RNG seed.
func NewGameController(scenarioJSON []byte, seed int64) (*GameController, error) {
	sc, err := scenario.LoadScenario(scenarioJSON)
	if err != nil {
		return nil, err
	}

	rng := types.NewSeededRNG(seed)
	engine, err := simulation.NewSimulationEngine(sc, rng)
	if err != nil {
		return nil, err
	}

	gc := &GameController{
		engine: engine,
		state:  Playing,
	}
	gc.snapshot = simulation.BuildSnapshot(engine.State)

	return gc, nil
}

// Snapshot returns the current read-only game snapshot.
func (gc *GameController) Snapshot() scenario.GameSnapshot {
	return gc.snapshot
}

// State returns the current game state.
func (gc *GameController) State() GameState {
	return gc.state
}

// Result returns the game result. Only meaningful when State is GameOver.
func (gc *GameController) Result() simulation.GameResult {
	return gc.result
}

// Engine returns the underlying SimulationEngine for direct state access
// (e.g., Cave, Beasts, RoomTypeRegistry for rendering).
func (gc *GameController) Engine() *simulation.SimulationEngine {
	return gc.engine
}

// AddAction appends a player action to the pending queue. Actions are
// consumed and cleared on the next call to AdvanceTick.
func (gc *GameController) AddAction(action simulation.PlayerAction) {
	gc.pending = append(gc.pending, action)
}

// PendingActions returns a copy of the current pending action queue.
func (gc *GameController) PendingActions() []simulation.PlayerAction {
	out := make([]simulation.PlayerAction, len(gc.pending))
	copy(out, gc.pending)
	return out
}

// AdvanceTick processes one simulation tick with the pending actions,
// updates the snapshot, and clears the pending queue. If the game is
// over or paused, it returns the last result without advancing.
func (gc *GameController) AdvanceTick() (simulation.GameResult, error) {
	if gc.state == GameOver {
		return gc.result, nil
	}

	actions := gc.pending
	if len(actions) == 0 {
		actions = []simulation.PlayerAction{simulation.NoAction{}}
	}
	gc.pending = nil

	result, err := gc.engine.Step(actions)
	if err != nil {
		return simulation.GameResult{}, err
	}

	gc.snapshot = simulation.BuildSnapshot(gc.engine.State)

	if result.Status != simulation.Running {
		gc.state = GameOver
		gc.result = result
	}

	return result, nil
}
