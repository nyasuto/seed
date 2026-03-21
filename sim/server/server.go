package server

import (
	"fmt"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/sim/metrics"
)

// defaultMaxTicks is the fallback when the scenario has no MaxTicks constraint.
const defaultMaxTicks = 10000

// GameServer wraps core's SimulationEngine and drives it via an
// ActionProvider. It handles engine creation, the tick loop, and
// statistics collection.
type GameServer struct {
	scenario  *scenario.Scenario
	seed      int64
	engine    *simulation.SimulationEngine
	collector *metrics.Collector
}

// NewGameServer creates a GameServer for the given scenario and RNG seed.
func NewGameServer(sc *scenario.Scenario, seed int64) (*GameServer, error) {
	if sc == nil {
		return nil, fmt.Errorf("scenario must not be nil")
	}
	return &GameServer{scenario: sc, seed: seed, collector: metrics.NewCollector()}, nil
}

// RunGame executes a full game using the provided ActionProvider.
// It creates a SimulationEngine, runs the tick loop, and returns the result.
func (gs *GameServer) RunGame(provider ActionProvider) (simulation.RunResult, error) {
	rng := types.NewCheckpointableRNG(gs.seed)
	engine, err := simulation.NewSimulationEngine(gs.scenario, rng)
	if err != nil {
		return simulation.RunResult{}, fmt.Errorf("create engine: %w", err)
	}

	gs.engine = engine
	simulation.EnableRecording(engine)
	defer func() { gs.engine = nil }()

	return gs.runLoop(provider)
}

// ResumeGame continues a game from a previously loaded checkpoint.
// Call LoadCheckpoint before calling this method.
func (gs *GameServer) ResumeGame(provider ActionProvider) (simulation.RunResult, error) {
	if gs.engine == nil {
		return simulation.RunResult{}, fmt.Errorf("no active engine; call LoadCheckpoint first")
	}
	defer func() { gs.engine = nil }()

	return gs.runLoop(provider)
}

// runLoop drives the tick loop using the current engine.
func (gs *GameServer) runLoop(provider ActionProvider) (simulation.RunResult, error) {
	engine := gs.engine
	maxTicks := gs.maxTicks()

	for i := int(engine.State.Progress.CurrentTick); i < maxTicks; i++ {
		snapshot := simulation.BuildSnapshot(engine.State)

		actions, err := provider.ProvideActions(snapshot)
		if err != nil {
			return simulation.RunResult{}, fmt.Errorf("provide actions at tick %d: %w", i, err)
		}
		if actions == nil {
			actions = []simulation.PlayerAction{simulation.NoAction{}}
		}

		result, err := engine.Step(actions)
		if err != nil {
			return simulation.RunResult{}, fmt.Errorf("step failed at tick %d: %w", i, err)
		}

		postSnapshot := simulation.BuildSnapshot(engine.State)
		gs.collector.OnTick(postSnapshot, actions)
		provider.OnTickComplete(postSnapshot)

		if result.Status != simulation.Running {
			runResult := simulation.RunResult{
				Result:     result,
				TickCount:  int(result.FinalTick),
				Statistics: buildStatistics(engine),
			}
			provider.OnGameEnd(runResult)
			return runResult, nil
		}
	}

	runResult := simulation.RunResult{
		Result: simulation.GameResult{
			Status:    simulation.Lost,
			FinalTick: engine.State.Progress.CurrentTick,
			Reason:    "max ticks reached",
		},
		TickCount:  maxTicks,
		Statistics: buildStatistics(engine),
	}
	provider.OnGameEnd(runResult)
	return runResult, nil
}

// maxTicks returns the MaxTicks from the scenario constraints, falling back
// to defaultMaxTicks if not specified.
func (gs *GameServer) maxTicks() int {
	if gs.scenario.Constraints.MaxTicks > 0 {
		return int(gs.scenario.Constraints.MaxTicks)
	}
	return defaultMaxTicks
}

// Collector returns the metrics Collector associated with this GameServer.
func (gs *GameServer) Collector() *metrics.Collector {
	return gs.collector
}

// buildStatistics collects RunStatistics from the engine's accumulated state.
func buildStatistics(engine *simulation.SimulationEngine) simulation.RunStatistics {
	s := engine.State
	snapshot := simulation.BuildSnapshot(s)

	return simulation.RunStatistics{
		PeakChi:        s.PeakChi,
		WavesDefeated:  snapshot.DefeatedWaves,
		FinalFengShui:  snapshot.CaveFengShuiScore,
		Evolutions:     s.EvolutionCount,
		DamageDealt:    s.TotalDamageDealt,
		DamageReceived: s.TotalDamageReceived,
		DeficitTicks:   s.TotalDeficitTicks,
	}
}
