package server

import (
	"fmt"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// defaultMaxTicks is the fallback when the scenario has no MaxTicks constraint.
const defaultMaxTicks = 10000

// GameServer wraps core's SimulationEngine and drives it via an
// ActionProvider. It handles engine creation, the tick loop, and
// statistics collection.
type GameServer struct {
	scenario *scenario.Scenario
	seed     int64
}

// NewGameServer creates a GameServer for the given scenario and RNG seed.
func NewGameServer(sc *scenario.Scenario, seed int64) (*GameServer, error) {
	if sc == nil {
		return nil, fmt.Errorf("scenario must not be nil")
	}
	return &GameServer{scenario: sc, seed: seed}, nil
}

// RunGame executes a full game using the provided ActionProvider.
// It creates a SimulationEngine, runs the tick loop, and returns the result.
func (gs *GameServer) RunGame(provider ActionProvider) (simulation.RunResult, error) {
	rng := types.NewCheckpointableRNG(gs.seed)
	engine, err := simulation.NewSimulationEngine(gs.scenario, rng)
	if err != nil {
		return simulation.RunResult{}, fmt.Errorf("create engine: %w", err)
	}

	maxTicks := gs.maxTicks()

	for i := range maxTicks {
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
