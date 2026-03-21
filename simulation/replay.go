package simulation

import (
	"errors"

	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/types"
)

// Errors returned by replay functions.
var (
	// ErrRNGNotCheckpointable is returned when the engine's RNG does not
	// support state extraction for replay recording.
	ErrRNGNotCheckpointable = errors.New("rng does not support checkpointing")
)

// Replay captures the minimal information needed to deterministically
// reproduce a simulation run: the RNG seed, scenario ID, and every player
// action submitted on each tick.
type Replay struct {
	// Seed is the initial RNG seed used to create the simulation.
	Seed int64
	// ScenarioID identifies the scenario that was played.
	ScenarioID string
	// Actions maps each tick to the player actions submitted on that tick.
	Actions map[types.Tick][]PlayerAction
}

// RecordReplay extracts a Replay from a SimulationEngine that has been run
// with action recording enabled. The engine's RNG must implement
// CheckpointableRNG so the original seed can be recovered.
// Call EnableRecording on the engine before running the simulation.
func RecordReplay(engine *SimulationEngine) (*Replay, error) {
	crng, ok := engine.State.RNG.(types.CheckpointableRNG)
	if !ok {
		return nil, ErrRNGNotCheckpointable
	}

	actions := engine.RecordedActions
	if actions == nil {
		actions = make(map[types.Tick][]PlayerAction)
	}

	return &Replay{
		Seed:       crng.RNGState().Seed,
		ScenarioID: engine.State.Scenario.ID,
		Actions:    actions,
	}, nil
}

// EnableRecording initializes action recording on the engine. This must be
// called before Run or Step to capture player actions for replay.
func EnableRecording(engine *SimulationEngine) {
	if engine.RecordedActions == nil {
		engine.RecordedActions = make(map[types.Tick][]PlayerAction)
	}
}

// PlayReplay recreates a simulation from a Replay and the corresponding
// scenario. It constructs a new engine with the same RNG seed and replays
// the recorded actions tick by tick. The result should be identical to the
// original run due to deterministic design.
func PlayReplay(replay *Replay, sc *scenario.Scenario) (GameResult, error) {
	rng := types.NewCheckpointableRNG(replay.Seed)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		return GameResult{}, err
	}

	// Determine the maximum tick to replay.
	var maxTick types.Tick
	for t := range replay.Actions {
		if t > maxTick {
			maxTick = t
		}
	}

	// Run the simulation providing recorded actions on each tick.
	result, err := engine.Run(int(maxTick)+1, func(snapshot scenario.GameSnapshot) []PlayerAction {
		if actions, ok := replay.Actions[snapshot.Tick]; ok {
			return actions
		}
		return []PlayerAction{NoAction{}}
	})
	if err != nil {
		return GameResult{}, err
	}

	return result, nil
}
