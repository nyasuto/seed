package simulation

import (
	"os"
	"testing"

	"github.com/ponpoko/chaosseed-core/scenario"
)

func loadScenarioFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read scenario file %s: %v", path, err)
	}
	return data
}

func TestIntegration_TutorialSimpleAIWins(t *testing.T) {
	scenJSON := loadScenarioFile(t, "../scenario/testdata/tutorial.json")

	runner := &SimulationRunner{}
	result, err := runner.RunWithAI(scenJSON, 42, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("RunWithAI: %v", err)
	}

	if result.Result.Status != Won {
		t.Errorf("expected SimpleAI to win tutorial scenario, got status=%v reason=%q at tick %d",
			result.Result.Status, result.Result.Reason, result.Result.FinalTick)
	}
	t.Logf("Tutorial completed: status=%v tick=%d waves_defeated=%d peak_chi=%.0f",
		result.Result.Status, result.TickCount, result.Statistics.WavesDefeated, result.Statistics.PeakChi)
}

func TestIntegration_StandardSimpleAIDefendsAtLeast3Waves(t *testing.T) {
	scenJSON := loadScenarioFile(t, "../scenario/testdata/standard.json")

	runner := &SimulationRunner{}
	result, err := runner.RunWithAI(scenJSON, 42, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("RunWithAI: %v", err)
	}

	const minWaves = 3
	if result.Statistics.WavesDefeated < minWaves {
		t.Errorf("expected SimpleAI to defeat at least %d waves in standard scenario, got %d (status=%v reason=%q tick=%d)",
			minWaves, result.Statistics.WavesDefeated, result.Result.Status, result.Result.Reason, result.Result.FinalTick)
	}
	t.Logf("Standard scenario: status=%v tick=%d waves_defeated=%d peak_chi=%.0f fengshui=%.1f",
		result.Result.Status, result.TickCount, result.Statistics.WavesDefeated,
		result.Statistics.PeakChi, result.Statistics.FinalFengShui)
}

// runWithSnapshots runs a simulation and collects per-tick snapshots for determinism comparison.
func runWithSnapshots(t *testing.T, scenJSON []byte, seed int64) (RunResult, []scenario.GameSnapshot) {
	t.Helper()
	runner := &SimulationRunner{}

	engine, sc, err := runner.createEngine(scenJSON, seed)
	if err != nil {
		t.Fatalf("createEngine: %v", err)
	}

	ai := NewSimpleAIPlayer(engine.State)

	maxTicks := runner.maxTicks(sc)
	var snapshots []scenario.GameSnapshot

	result, err := engine.Run(maxTicks, func(snap scenario.GameSnapshot) []PlayerAction {
		snapshots = append(snapshots, snap)
		return ai.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	return RunResult{
		Result:     result,
		TickCount:  int(result.FinalTick),
		Statistics: collectStatistics(engine),
	}, snapshots
}

func TestIntegration_Determinism_SameSeedIdenticalResults(t *testing.T) {
	scenJSON := loadScenarioFile(t, "../scenario/testdata/tutorial.json")
	const seed int64 = 12345

	result1, snapshots1 := runWithSnapshots(t, scenJSON, seed)
	result2, snapshots2 := runWithSnapshots(t, scenJSON, seed)

	// Compare final results
	if result1.Result.Status != result2.Result.Status {
		t.Errorf("status mismatch: run1=%v run2=%v", result1.Result.Status, result2.Result.Status)
	}
	if result1.Result.FinalTick != result2.Result.FinalTick {
		t.Errorf("final tick mismatch: run1=%d run2=%d", result1.Result.FinalTick, result2.Result.FinalTick)
	}
	if result1.Result.Reason != result2.Result.Reason {
		t.Errorf("reason mismatch: run1=%q run2=%q", result1.Result.Reason, result2.Result.Reason)
	}
	if result1.TickCount != result2.TickCount {
		t.Errorf("tick count mismatch: run1=%d run2=%d", result1.TickCount, result2.TickCount)
	}
	if result1.Statistics != result2.Statistics {
		t.Errorf("statistics mismatch:\n  run1=%+v\n  run2=%+v", result1.Statistics, result2.Statistics)
	}

	// Compare per-tick snapshots
	if len(snapshots1) != len(snapshots2) {
		t.Fatalf("snapshot count mismatch: run1=%d run2=%d", len(snapshots1), len(snapshots2))
	}
	for i := range snapshots1 {
		s1, s2 := snapshots1[i], snapshots2[i]
		if s1 != s2 {
			t.Errorf("snapshot mismatch at tick %d:\n  run1=%+v\n  run2=%+v", i, s1, s2)
			break
		}
	}

	t.Logf("Determinism verified: %d ticks, status=%v, identical across 2 runs with seed=%d",
		result1.TickCount, result1.Result.Status, seed)
}
