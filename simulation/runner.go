package simulation

import (
	"fmt"

	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/types"
)

// defaultMaxTicks is used when the scenario does not specify a MaxTicks constraint.
const defaultMaxTicks = 10000

// AIPlayerFactory creates an AIPlayer given the initialized game state.
// This allows the runner to construct the engine first, then hand off
// the state to the AI player factory.
type AIPlayerFactory func(state *GameState) AIPlayer

// RunStatistics captures aggregate metrics from a completed simulation run.
type RunStatistics struct {
	// PeakChi is the highest chi pool balance observed during the run.
	PeakChi float64
	// WavesDefeated is the number of invasion waves successfully repelled.
	WavesDefeated int
	// FinalFengShui is the cave's feng shui score at the end of the run.
	FinalFengShui float64
	// Evolutions is the total number of beast evolutions that occurred.
	Evolutions int
	// DamageDealt is the total damage dealt to invaders.
	DamageDealt int
	// DamageReceived is the total damage received by beasts and the core.
	DamageReceived int
	// DeficitTicks is the total number of ticks where the economy was in deficit.
	DeficitTicks int
}

// RunResult captures the outcome of a completed simulation run.
type RunResult struct {
	// Result is the game outcome (Won/Lost and reason).
	Result GameResult
	// TickCount is the number of ticks actually executed.
	TickCount int
	// Statistics holds aggregate metrics from the run.
	Statistics RunStatistics
}

// SimulationRunner provides high-level APIs for running simulations
// from raw scenario JSON bytes. It handles scenario loading, engine
// creation, and result collection.
type SimulationRunner struct{}

// RunWithAI loads a scenario from JSON, creates an engine with the given
// seed, constructs an AIPlayer via the factory, and runs the simulation
// to completion. The returned RunResult contains the game outcome and
// the number of ticks executed.
func (r *SimulationRunner) RunWithAI(scenarioJSON []byte, seed int64, factory AIPlayerFactory) (RunResult, error) {
	engine, sc, err := r.createEngine(scenarioJSON, seed)
	if err != nil {
		return RunResult{}, err
	}

	ai := factory(engine.State)

	maxTicks := r.maxTicks(sc)
	result, err := engine.Run(maxTicks, func(snapshot scenario.GameSnapshot) []PlayerAction {
		return ai.DecideActions(snapshot)
	})
	if err != nil {
		return RunResult{}, fmt.Errorf("run failed: %w", err)
	}

	return RunResult{
		Result:     result,
		TickCount:  int(result.FinalTick),
		Statistics: collectStatistics(engine),
	}, nil
}

// RunInteractive loads a scenario from JSON, creates an engine with the
// given seed, and runs an interactive simulation loop using channels.
// Each tick, the current GameSnapshot is sent on snapshotCh, and the
// runner waits for player actions on actionCh. The loop continues until
// a terminal condition is reached or actionCh is closed (which results
// in a loss). When the simulation ends, snapshotCh is closed.
func (r *SimulationRunner) RunInteractive(scenarioJSON []byte, seed int64, actionCh <-chan []PlayerAction, snapshotCh chan<- scenario.GameSnapshot) (RunResult, error) {
	engine, sc, err := r.createEngine(scenarioJSON, seed)
	if err != nil {
		close(snapshotCh)
		return RunResult{}, err
	}

	maxTicks := r.maxTicks(sc)
	defer close(snapshotCh)

	for i := 0; i < maxTicks; i++ {
		snapshot := BuildSnapshot(engine.State)
		snapshotCh <- snapshot

		actions, ok := <-actionCh
		if !ok {
			// Channel closed — player disconnected.
			return RunResult{
				Result: GameResult{
					Status:    Lost,
					FinalTick: engine.State.Progress.CurrentTick,
					Reason:    "player disconnected",
				},
				TickCount:  i,
				Statistics: collectStatistics(engine),
			}, nil
		}
		if actions == nil {
			actions = []PlayerAction{NoAction{}}
		}

		result, err := engine.Step(actions)
		if err != nil {
			return RunResult{}, fmt.Errorf("step failed at tick %d: %w", i, err)
		}
		if result.Status != Running {
			return RunResult{
				Result:     result,
				TickCount:  int(result.FinalTick),
				Statistics: collectStatistics(engine),
			}, nil
		}
	}

	return RunResult{
		Result: GameResult{
			Status:    Lost,
			FinalTick: engine.State.Progress.CurrentTick,
			Reason:    "max ticks reached",
		},
		TickCount:  maxTicks,
		Statistics: collectStatistics(engine),
	}, nil
}

// createEngine loads a scenario from JSON and creates a SimulationEngine.
func (r *SimulationRunner) createEngine(scenarioJSON []byte, seed int64) (*SimulationEngine, *scenario.Scenario, error) {
	sc, err := scenario.LoadScenario(scenarioJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("load scenario: %w", err)
	}

	rng := types.NewCheckpointableRNG(seed)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		return nil, nil, fmt.Errorf("create engine: %w", err)
	}

	return engine, sc, nil
}

// maxTicks returns the MaxTicks from the scenario constraints, falling back
// to defaultMaxTicks if not specified.
func (r *SimulationRunner) maxTicks(sc *scenario.Scenario) int {
	if sc.Constraints.MaxTicks > 0 {
		return int(sc.Constraints.MaxTicks)
	}
	return defaultMaxTicks
}

// collectStatistics builds RunStatistics from the engine's accumulated
// state after a simulation run completes.
func collectStatistics(engine *SimulationEngine) RunStatistics {
	s := engine.State
	snapshot := BuildSnapshot(s)

	return RunStatistics{
		PeakChi:        s.PeakChi,
		WavesDefeated:  snapshot.DefeatedWaves,
		FinalFengShui:  snapshot.CaveFengShuiScore,
		Evolutions:     s.EvolutionCount,
		DamageDealt:    s.TotalDamageDealt,
		DamageReceived: s.TotalDamageReceived,
		DeficitTicks:   s.TotalDeficitTicks,
	}
}
