package simulation

import (
	"encoding/json"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
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

func TestSimulationRunner_RunWithAI_SimpleAI(t *testing.T) {
	runner := &SimulationRunner{}
	scenJSON := tutorialScenarioJSON()

	result, err := runner.RunWithAI(scenJSON, 42, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("RunWithAI: %v", err)
	}

	if result.Result.Status != Won {
		t.Errorf("expected Won, got %v (reason: %s)", result.Result.Status, result.Result.Reason)
	}
	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}
}

func TestSimulationRunner_RunWithAI_Deterministic(t *testing.T) {
	runner := &SimulationRunner{}
	scenJSON := tutorialScenarioJSON()

	result1, err := runner.RunWithAI(scenJSON, 123, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("RunWithAI run1: %v", err)
	}

	result2, err := runner.RunWithAI(scenJSON, 123, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("RunWithAI run2: %v", err)
	}

	if result1.Result.Status != result2.Result.Status {
		t.Errorf("status mismatch: %v vs %v", result1.Result.Status, result2.Result.Status)
	}
	if result1.TickCount != result2.TickCount {
		t.Errorf("tick count mismatch: %d vs %d", result1.TickCount, result2.TickCount)
	}
	if result1.Result.FinalTick != result2.Result.FinalTick {
		t.Errorf("final tick mismatch: %d vs %d", result1.Result.FinalTick, result2.Result.FinalTick)
	}
}

func TestSimulationRunner_BatchRun_MultipleSeeds(t *testing.T) {
	runner := &SimulationRunner{}
	scenJSON := tutorialScenarioJSON()
	seeds := []int64{1, 2, 3, 42, 100}

	results, err := runner.BatchRun(scenJSON, seeds, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("BatchRun: %v", err)
	}

	if len(results) != len(seeds) {
		t.Fatalf("expected %d results, got %d", len(seeds), len(results))
	}

	for i, r := range results {
		if r.Result.Status != Won {
			t.Errorf("seed %d: expected Won, got %v (reason: %s)", seeds[i], r.Result.Status, r.Result.Reason)
		}
		if r.TickCount == 0 {
			t.Errorf("seed %d: expected TickCount > 0", seeds[i])
		}
	}
}

func TestSimulationRunner_BatchRun_Deterministic(t *testing.T) {
	runner := &SimulationRunner{}
	scenJSON := tutorialScenarioJSON()
	seeds := []int64{10, 20, 30}

	results1, err := runner.BatchRun(scenJSON, seeds, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("BatchRun run1: %v", err)
	}

	results2, err := runner.BatchRun(scenJSON, seeds, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("BatchRun run2: %v", err)
	}

	for i := range seeds {
		if results1[i].Result.Status != results2[i].Result.Status {
			t.Errorf("seed %d: status mismatch: %v vs %v", seeds[i], results1[i].Result.Status, results2[i].Result.Status)
		}
		if results1[i].TickCount != results2[i].TickCount {
			t.Errorf("seed %d: tick count mismatch: %d vs %d", seeds[i], results1[i].TickCount, results2[i].TickCount)
		}
		if results1[i].Statistics.PeakChi != results2[i].Statistics.PeakChi {
			t.Errorf("seed %d: PeakChi mismatch: %f vs %f", seeds[i], results1[i].Statistics.PeakChi, results2[i].Statistics.PeakChi)
		}
		if results1[i].Statistics.DeficitTicks != results2[i].Statistics.DeficitTicks {
			t.Errorf("seed %d: DeficitTicks mismatch: %d vs %d", seeds[i], results1[i].Statistics.DeficitTicks, results2[i].Statistics.DeficitTicks)
		}
	}
}

func TestSimulationRunner_BatchRun_EmptySeeds(t *testing.T) {
	runner := &SimulationRunner{}
	scenJSON := tutorialScenarioJSON()

	results, err := runner.BatchRun(scenJSON, []int64{}, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("BatchRun: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSimulationRunner_BatchRun_InvalidJSON(t *testing.T) {
	runner := &SimulationRunner{}
	_, err := runner.BatchRun([]byte("{invalid"), []int64{1}, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSimulationRunner_RunStatistics_Collection(t *testing.T) {
	runner := &SimulationRunner{}
	scenJSON := tutorialScenarioJSON()

	result, err := runner.RunWithAI(scenJSON, 42, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("RunWithAI: %v", err)
	}

	stats := result.Statistics
	// PeakChi should be positive (chi was flowing during the run)
	if stats.PeakChi <= 0 {
		t.Errorf("expected PeakChi > 0, got %f", stats.PeakChi)
	}
	// FinalFengShui should be non-negative
	if stats.FinalFengShui < 0 {
		t.Errorf("expected FinalFengShui >= 0, got %f", stats.FinalFengShui)
	}
	// No waves in tutorial scenario, so WavesDefeated should be 0
	if stats.WavesDefeated != 0 {
		t.Errorf("expected WavesDefeated == 0 (no waves), got %d", stats.WavesDefeated)
	}
}

func TestSimulationRunner_RunWithAI_InvalidJSON(t *testing.T) {
	runner := &SimulationRunner{}
	_, err := runner.RunWithAI([]byte("{invalid"), 1, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSimulationRunner_RunInteractive_ChannelBased(t *testing.T) {
	runner := &SimulationRunner{}
	scenJSON := tutorialScenarioJSON()

	actionCh := make(chan []PlayerAction)
	snapshotCh := make(chan scenario.GameSnapshot)

	var result RunResult
	var runErr error
	done := make(chan struct{})

	go func() {
		result, runErr = runner.RunInteractive(scenJSON, 42, actionCh, snapshotCh)
		close(done)
	}()

	// Play: send NoAction each tick until the game ends.
	tickCount := 0
	for snapshot := range snapshotCh {
		_ = snapshot
		tickCount++
		actionCh <- []PlayerAction{NoAction{}}
	}
	<-done

	if runErr != nil {
		t.Fatalf("RunInteractive: %v", runErr)
	}
	if result.Result.Status != Won {
		t.Errorf("expected Won, got %v (reason: %s)", result.Result.Status, result.Result.Reason)
	}
	if tickCount == 0 {
		t.Error("expected at least one snapshot")
	}
}

func TestSimulationRunner_RunInteractive_PlayerDisconnect(t *testing.T) {
	runner := &SimulationRunner{}
	scenJSON := tutorialScenarioJSON()

	actionCh := make(chan []PlayerAction)
	snapshotCh := make(chan scenario.GameSnapshot)

	var result RunResult
	var runErr error
	done := make(chan struct{})

	go func() {
		result, runErr = runner.RunInteractive(scenJSON, 42, actionCh, snapshotCh)
		close(done)
	}()

	// Receive first snapshot, then close actionCh to simulate disconnect.
	<-snapshotCh
	close(actionCh)

	// Drain remaining snapshots (snapshotCh will be closed by runner).
	for range snapshotCh {
	}
	<-done

	if runErr != nil {
		t.Fatalf("RunInteractive: %v", runErr)
	}
	if result.Result.Status != Lost {
		t.Errorf("expected Lost, got %v", result.Result.Status)
	}
	if result.Result.Reason != "player disconnected" {
		t.Errorf("expected 'player disconnected', got %q", result.Result.Reason)
	}
}

func TestSimulationRunner_RunInteractive_MaxTicks(t *testing.T) {
	// Create a scenario with very long survive requirement but low max_ticks.
	data, _ := json.Marshal(map[string]any{
		"id":         "long",
		"name":       "Long",
		"difficulty": "hard",
		"initial_state": map[string]any{
			"cave_width":      20,
			"cave_height":     20,
			"terrain_seed":    1,
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
			{"type": "survive_until", "params": map[string]any{"ticks": 9999}},
		},
		"lose_conditions": []map[string]any{
			{"type": "core_destroyed"},
		},
		"wave_schedule": []any{},
		"events":        []any{},
		"constraints": map[string]any{
			"max_ticks": 5,
		},
	})

	runner := &SimulationRunner{}
	actionCh := make(chan []PlayerAction)
	snapshotCh := make(chan scenario.GameSnapshot)

	var result RunResult
	var runErr error
	done := make(chan struct{})

	go func() {
		result, runErr = runner.RunInteractive(data, 42, actionCh, snapshotCh)
		close(done)
	}()

	for range snapshotCh {
		actionCh <- []PlayerAction{NoAction{}}
	}
	<-done

	if runErr != nil {
		t.Fatalf("RunInteractive: %v", runErr)
	}
	if result.Result.Status != Lost {
		t.Errorf("expected Lost, got %v", result.Result.Status)
	}
	if result.Result.Reason != "max ticks reached" {
		t.Errorf("expected 'max ticks reached', got %q", result.Result.Reason)
	}
}
