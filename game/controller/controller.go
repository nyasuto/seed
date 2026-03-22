package controller

import (
	"fmt"

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
	engine       *simulation.SimulationEngine
	snapshot     scenario.GameSnapshot
	pending      []simulation.PlayerAction
	state        GameState
	result       simulation.GameResult
	ffSpeed      int // ticks per UpdateTick call in FastForward mode
	scenarioJSON []byte
}

// NewGameController creates a GameController by loading the given scenario JSON
// and initializing a SimulationEngine with the provided RNG seed.
func NewGameController(scenarioJSON []byte, seed int64) (*GameController, error) {
	sc, err := scenario.LoadScenario(scenarioJSON)
	if err != nil {
		return nil, err
	}

	rng := types.NewCheckpointableRNG(seed)
	engine, err := simulation.NewSimulationEngine(sc, rng)
	if err != nil {
		return nil, err
	}

	gc := &GameController{
		engine:       engine,
		state:        Playing,
		scenarioJSON: scenarioJSON,
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

// ScenarioJSON returns the raw scenario JSON used to create this controller.
// This is needed to restore a checkpoint (scenario is immutable and not serialized).
func (gc *GameController) ScenarioJSON() []byte {
	return gc.scenarioJSON
}

// CreateCheckpoint creates a checkpoint of the current engine state.
func (gc *GameController) CreateCheckpoint() (*simulation.Checkpoint, error) {
	return simulation.CreateCheckpoint(gc.engine)
}

// RestoreFromCheckpoint restores the controller state from a checkpoint.
// The scenario is reloaded from the stored scenarioJSON.
func (gc *GameController) RestoreFromCheckpoint(cp *simulation.Checkpoint) error {
	sc, err := scenario.LoadScenario(gc.scenarioJSON)
	if err != nil {
		return fmt.Errorf("reload scenario: %w", err)
	}
	engine, err := simulation.RestoreCheckpoint(cp, sc)
	if err != nil {
		return fmt.Errorf("restore checkpoint: %w", err)
	}
	gc.engine = engine
	gc.snapshot = simulation.BuildSnapshot(engine.State)
	gc.pending = nil
	gc.state = Playing
	gc.result = simulation.GameResult{}
	return nil
}

// NewGameControllerFromCheckpoint creates a GameController by restoring from
// a checkpoint. The scenarioJSON is needed because the checkpoint does not
// include immutable scenario data.
func NewGameControllerFromCheckpoint(cp *simulation.Checkpoint, scenarioJSON []byte) (*GameController, error) {
	sc, err := scenario.LoadScenario(scenarioJSON)
	if err != nil {
		return nil, fmt.Errorf("load scenario: %w", err)
	}
	engine, err := simulation.RestoreCheckpoint(cp, sc)
	if err != nil {
		return nil, fmt.Errorf("restore checkpoint: %w", err)
	}
	gc := &GameController{
		engine:       engine,
		state:        Playing,
		scenarioJSON: scenarioJSON,
	}
	gc.snapshot = simulation.BuildSnapshot(engine.State)
	return gc, nil
}
