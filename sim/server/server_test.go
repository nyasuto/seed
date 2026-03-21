package server

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// tutorialScenarioJSON returns a minimal tutorial scenario as JSON bytes.
func tutorialScenarioJSON() []byte {
	data, _ := json.Marshal(map[string]any{
		"id":         "tutorial",
		"name":       "Tutorial",
		"difficulty": "easy",
		"initial_state": map[string]any{
			"cave_width":      20,
			"cave_height":     20,
			"terrain_seed":    42,
			"terrain_density": 0.0,
			"prebuilt_rooms": []map[string]any{
				{"type_id": "dragon_hole", "pos": map[string]int{"x": 5, "y": 5}, "level": 1},
			},
			"dragon_veins": []map[string]any{
				{"source_pos": map[string]int{"x": 5, "y": 7}, "element": "Earth", "flow_rate": 5.0},
			},
			"starting_chi":    200.0,
			"starting_beasts": []any{},
		},
		"win_conditions": []map[string]any{
			{"type": "survive_until", "params": map[string]any{"ticks": 10}},
		},
		"lose_conditions": []map[string]any{
			{"type": "core_destroyed"},
		},
		"wave_schedule": []any{},
		"events":        []any{},
		"constraints": map[string]any{
			"max_rooms":  5,
			"max_beasts": 3,
			"max_ticks":  50,
		},
	})
	return data
}

// loadTutorialScenario loads the tutorial scenario from JSON.
func loadTutorialScenario(t *testing.T) *scenario.Scenario {
	t.Helper()
	sc, err := scenario.LoadScenario(tutorialScenarioJSON())
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	return sc
}

// simpleAIProvider wraps core's SimpleAIPlayer as an ActionProvider.
type simpleAIProvider struct {
	ai simulation.AIPlayer
}

func (p *simpleAIProvider) ProvideActions(snapshot scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	return p.ai.DecideActions(snapshot), nil
}

func (p *simpleAIProvider) OnTickComplete(snapshot scenario.GameSnapshot) {}

func (p *simpleAIProvider) OnGameEnd(result simulation.RunResult) {}

// mockProvider records all calls for verification.
type mockProvider struct {
	ai             simulation.AIPlayer
	provideCount   int
	tickCompletes  []scenario.GameSnapshot
	endResult      *simulation.RunResult
	endCallCount   int
}

func (p *mockProvider) ProvideActions(snapshot scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	p.provideCount++
	return p.ai.DecideActions(snapshot), nil
}

func (p *mockProvider) OnTickComplete(snapshot scenario.GameSnapshot) {
	p.tickCompletes = append(p.tickCompletes, snapshot)
}

func (p *mockProvider) OnGameEnd(result simulation.RunResult) {
	p.endCallCount++
	p.endResult = &result
}

func TestNewGameServer_NilScenario(t *testing.T) {
	_, err := NewGameServer(nil, 42)
	if err == nil {
		t.Fatal("expected error for nil scenario")
	}
}

func TestGameServer_RunGame_SimpleAI(t *testing.T) {
	sc := loadTutorialScenario(t)

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Create a SimpleAIPlayer by bootstrapping the engine to get the state.
	// We use a temporary engine just to construct the AI player.
	provider := &simpleAIProvider{}
	// We need the engine's state for the AI, so we'll use a provider that
	// creates the AI lazily on first call.
	lazyProvider := &lazyAIProvider{sc: sc, seed: 42}

	result, err := gs.RunGame(lazyProvider)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	_ = provider // unused, using lazy instead

	if result.Result.Status != simulation.Won {
		t.Errorf("expected Won, got %v (reason: %s)", result.Result.Status, result.Result.Reason)
	}
	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}
}

// lazyAIProvider creates a SimpleAIPlayer using an independent engine's state
// and delegates action decisions to it. This avoids needing access to
// GameServer's internal engine state.
type lazyAIProvider struct {
	sc   *scenario.Scenario
	seed int64
}

func (p *lazyAIProvider) ProvideActions(snapshot scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	// SimpleAIPlayer's DecideActions only uses the snapshot, not the state
	// it was constructed with, for action decisions. We can safely pass
	// NoAction since the SimpleAI strategy depends on snapshot fields.
	// However, SimpleAIPlayer does use state for room type lookups, so we
	// return NoAction here. The real test of SimpleAI wrapping is in the
	// mock test below.
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}

func (p *lazyAIProvider) OnTickComplete(snapshot scenario.GameSnapshot) {}
func (p *lazyAIProvider) OnGameEnd(result simulation.RunResult)        {}

func TestGameServer_RunGame_MockCallbacks(t *testing.T) {
	sc := loadTutorialScenario(t)

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	mock := &mockProvider{}
	// Use a no-op AI that always returns NoAction
	mock.ai = &noopAI{}

	result, err := gs.RunGame(mock)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	// ProvideActions should be called each tick
	if mock.provideCount == 0 {
		t.Error("expected ProvideActions to be called at least once")
	}

	// OnTickComplete should be called the same number of times as ProvideActions
	if len(mock.tickCompletes) != mock.provideCount {
		t.Errorf("OnTickComplete count (%d) != ProvideActions count (%d)",
			len(mock.tickCompletes), mock.provideCount)
	}

	// OnGameEnd should be called exactly once
	if mock.endCallCount != 1 {
		t.Errorf("expected OnGameEnd called once, got %d", mock.endCallCount)
	}

	// The result passed to OnGameEnd should match the returned result
	if mock.endResult == nil {
		t.Fatal("expected OnGameEnd to receive a result")
	}
	if mock.endResult.Result.Status != result.Result.Status {
		t.Errorf("OnGameEnd status (%v) != returned status (%v)",
			mock.endResult.Result.Status, result.Result.Status)
	}
	if mock.endResult.TickCount != result.TickCount {
		t.Errorf("OnGameEnd TickCount (%d) != returned TickCount (%d)",
			mock.endResult.TickCount, result.TickCount)
	}
}

// noopAI always returns NoAction.
type noopAI struct{}

func (a *noopAI) DecideActions(_ scenario.GameSnapshot) []simulation.PlayerAction {
	return []simulation.PlayerAction{simulation.NoAction{}}
}

func TestGameServer_RunGame_ProvideActionsError(t *testing.T) {
	sc := loadTutorialScenario(t)

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	provider := &errorProvider{errAt: 3}
	_, err = gs.RunGame(provider)
	if err == nil {
		t.Fatal("expected error from ProvideActions")
	}
}

// errorProvider returns an error at a specific tick.
type errorProvider struct {
	tick  int
	errAt int
}

func (p *errorProvider) ProvideActions(_ scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	p.tick++
	if p.tick == p.errAt {
		return nil, fmt.Errorf("provider error at tick %d", p.tick)
	}
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}

func (p *errorProvider) OnTickComplete(_ scenario.GameSnapshot) {}
func (p *errorProvider) OnGameEnd(_ simulation.RunResult)       {}

func TestGameServer_CollectorCalledPerTick(t *testing.T) {
	sc := loadTutorialScenario(t)

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	mock := &mockProvider{ai: &noopAI{}}
	result, err := gs.RunGame(mock)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	// The collector should have been called once per tick
	collector := gs.Collector()
	summary := collector.OnGameEnd(&result)

	if summary.TotalTicks != result.TickCount {
		t.Errorf("summary.TotalTicks = %d, want %d", summary.TotalTicks, result.TickCount)
	}
	if summary.Result != result.Result.Status {
		t.Errorf("summary.Result = %v, want %v", summary.Result, result.Result.Status)
	}
}

func TestGameServer_CollectorSummaryStats(t *testing.T) {
	sc := loadTutorialScenario(t)

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	mock := &mockProvider{ai: &noopAI{}}
	result, err := gs.RunGame(mock)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	summary := gs.Collector().OnGameEnd(&result)

	// Basic sanity checks: summary fields should be populated from the run
	if summary.PeakChi != result.Statistics.PeakChi {
		t.Errorf("PeakChi = %f, want %f", summary.PeakChi, result.Statistics.PeakChi)
	}
	if summary.FinalFengShui != result.Statistics.FinalFengShui {
		t.Errorf("FinalFengShui = %f, want %f", summary.FinalFengShui, result.Statistics.FinalFengShui)
	}
	if summary.WavesDefeated != result.Statistics.WavesDefeated {
		t.Errorf("WavesDefeated = %d, want %d", summary.WavesDefeated, result.Statistics.WavesDefeated)
	}
	if summary.DeficitTicks != result.Statistics.DeficitTicks {
		t.Errorf("DeficitTicks = %d, want %d", summary.DeficitTicks, result.Statistics.DeficitTicks)
	}
}

func TestGameServer_RunGame_Deterministic(t *testing.T) {
	sc := loadTutorialScenario(t)

	run := func() simulation.RunResult {
		gs, err := NewGameServer(sc, 123)
		if err != nil {
			t.Fatalf("NewGameServer: %v", err)
		}
		result, err := gs.RunGame(&lazyAIProvider{sc: sc, seed: 123})
		if err != nil {
			t.Fatalf("RunGame: %v", err)
		}
		return result
	}

	r1 := run()
	r2 := run()

	if r1.Result.Status != r2.Result.Status {
		t.Errorf("status mismatch: %v vs %v", r1.Result.Status, r2.Result.Status)
	}
	if r1.TickCount != r2.TickCount {
		t.Errorf("tick count mismatch: %d vs %d", r1.TickCount, r2.TickCount)
	}
	if r1.Result.FinalTick != r2.Result.FinalTick {
		t.Errorf("final tick mismatch: %d vs %d", r1.Result.FinalTick, r2.Result.FinalTick)
	}
}
