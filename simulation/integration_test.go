package simulation

import (
	"embed"
	"testing"

	"github.com/ponpoko/chaosseed-core/scenario"
)

//go:embed testdata
var testdataFS embed.FS

func loadScenarioFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := testdataFS.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("failed to read scenario file %s: %v", name, err)
	}
	return data
}

func TestIntegration_TutorialSimpleAIWins(t *testing.T) {
	scenJSON := loadScenarioFile(t, "tutorial.json")

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
	scenJSON := loadScenarioFile(t, "standard.json")

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

func TestIntegration_CheckpointRestore(t *testing.T) {
	scenJSON := loadScenarioFile(t, "tutorial.json")
	const seed int64 = 99

	runner := &SimulationRunner{}

	// --- Full run: collect snapshots for all ticks ---
	fullEngine, fullSC, err := runner.createEngine(scenJSON, seed)
	if err != nil {
		t.Fatalf("createEngine (full): %v", err)
	}
	fullAI := NewSimpleAIPlayer(fullEngine.State)
	maxTicks := runner.maxTicks(fullSC)

	var fullSnapshots []scenario.GameSnapshot
	fullResult, err := fullEngine.Run(maxTicks, func(snap scenario.GameSnapshot) []PlayerAction {
		fullSnapshots = append(fullSnapshots, snap)
		return fullAI.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("full Run: %v", err)
	}

	if len(fullSnapshots) < 100 {
		t.Fatalf("simulation ended too early (%d ticks), need at least 100 for checkpoint test", len(fullSnapshots))
	}

	// --- Partial run to tick 100, then checkpoint ---
	cpEngine, cpSC, err := runner.createEngine(scenJSON, seed)
	if err != nil {
		t.Fatalf("createEngine (checkpoint): %v", err)
	}
	cpAI := NewSimpleAIPlayer(cpEngine.State)

	// Run exactly 100 ticks.
	for i := range 100 {
		snap := BuildSnapshot(cpEngine.State)
		actions := cpAI.DecideActions(snap)
		if actions == nil {
			actions = []PlayerAction{NoAction{}}
		}
		result, err := cpEngine.Step(actions)
		if err != nil {
			t.Fatalf("Step at tick %d: %v", i, err)
		}
		if result.Status != Running {
			t.Fatalf("game ended early at tick %d before checkpoint", i)
		}
	}

	cp, err := CreateCheckpoint(cpEngine)
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}

	// --- Restore from checkpoint and continue ---
	restoredEngine, err := RestoreCheckpoint(cp, cpSC)
	if err != nil {
		t.Fatalf("RestoreCheckpoint: %v", err)
	}
	restoredAI := NewSimpleAIPlayer(restoredEngine.State)

	var restoredSnapshots []scenario.GameSnapshot
	restoredResult, err := restoredEngine.Run(maxTicks-100, func(snap scenario.GameSnapshot) []PlayerAction {
		restoredSnapshots = append(restoredSnapshots, snap)
		return restoredAI.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("restored Run: %v", err)
	}

	// --- Compare results ---
	if fullResult.Status != restoredResult.Status {
		t.Errorf("status mismatch: full=%v restored=%v", fullResult.Status, restoredResult.Status)
	}
	if fullResult.FinalTick != restoredResult.FinalTick {
		t.Errorf("final tick mismatch: full=%d restored=%d", fullResult.FinalTick, restoredResult.FinalTick)
	}
	if fullResult.Reason != restoredResult.Reason {
		t.Errorf("reason mismatch: full=%q restored=%q", fullResult.Reason, restoredResult.Reason)
	}

	// Compare per-tick snapshots from tick 100 onward.
	postCheckpointSnapshots := fullSnapshots[100:]
	if len(postCheckpointSnapshots) != len(restoredSnapshots) {
		t.Fatalf("snapshot count mismatch after checkpoint: full=%d restored=%d",
			len(postCheckpointSnapshots), len(restoredSnapshots))
	}
	for i := range postCheckpointSnapshots {
		if postCheckpointSnapshots[i] != restoredSnapshots[i] {
			t.Errorf("snapshot mismatch at tick %d (post-checkpoint index %d):\n  full=%+v\n  restored=%+v",
				100+i, i, postCheckpointSnapshots[i], restoredSnapshots[i])
			break
		}
	}

	t.Logf("Checkpoint restore verified: checkpoint at tick 100, %d ticks after restore, status=%v",
		len(restoredSnapshots), restoredResult.Status)
}

func TestIntegration_ReplayProducesSameResult(t *testing.T) {
	scenJSON := loadScenarioFile(t, "tutorial.json")
	const seed int64 = 77

	runner := &SimulationRunner{}

	// --- Original run with recording ---
	engine, sc, err := runner.createEngine(scenJSON, seed)
	if err != nil {
		t.Fatalf("createEngine: %v", err)
	}
	EnableRecording(engine)

	ai := NewSimpleAIPlayer(engine.State)
	maxTicks := runner.maxTicks(sc)

	var origSnapshots []scenario.GameSnapshot
	origResult, err := engine.Run(maxTicks, func(snap scenario.GameSnapshot) []PlayerAction {
		origSnapshots = append(origSnapshots, snap)
		return ai.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("original Run: %v", err)
	}

	// --- Record replay and round-trip through JSON ---
	replay, err := RecordReplay(engine)
	if err != nil {
		t.Fatalf("RecordReplay: %v", err)
	}

	data, err := MarshalReplay(replay)
	if err != nil {
		t.Fatalf("MarshalReplay: %v", err)
	}

	restored, err := UnmarshalReplay(data)
	if err != nil {
		t.Fatalf("UnmarshalReplay: %v", err)
	}

	// --- Play back the replay ---
	replayResult, err := PlayReplay(restored, sc)
	if err != nil {
		t.Fatalf("PlayReplay: %v", err)
	}

	// --- Compare results ---
	if origResult.Status != replayResult.Status {
		t.Errorf("status mismatch: original=%v replay=%v", origResult.Status, replayResult.Status)
	}
	if origResult.FinalTick != replayResult.FinalTick {
		t.Errorf("final tick mismatch: original=%d replay=%d", origResult.FinalTick, replayResult.FinalTick)
	}
	if origResult.Reason != replayResult.Reason {
		t.Errorf("reason mismatch: original=%q replay=%q", origResult.Reason, replayResult.Reason)
	}

	// Verify replay data integrity
	if restored.Seed != seed {
		t.Errorf("replay seed mismatch: expected %d got %d", seed, restored.Seed)
	}
	if len(restored.Actions) == 0 {
		t.Error("replay has no recorded actions")
	}

	t.Logf("Replay verified: %d ticks, %d action ticks recorded, status=%v, JSON size=%d bytes",
		origResult.FinalTick, len(restored.Actions), origResult.Status, len(data))
}

func TestIntegration_StressTest_LargeMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	scenJSON := loadScenarioFile(t, "stress.json")

	runner := &SimulationRunner{}
	result, err := runner.RunWithAI(scenJSON, 42, func(state *GameState) AIPlayer {
		return NewSimpleAIPlayer(state)
	})
	if err != nil {
		t.Fatalf("RunWithAI: %v", err)
	}

	// Verify the simulation completed within 1000 ticks
	const maxTicks = 1000
	if result.TickCount > maxTicks {
		t.Errorf("stress test exceeded %d ticks: got %d", maxTicks, result.TickCount)
	}

	// Verify the simulation actually ran a meaningful number of ticks
	const minTicks = 50
	if result.TickCount < minTicks {
		t.Errorf("stress test ended too quickly (%d ticks), expected at least %d", result.TickCount, minTicks)
	}

	// Verify game reached a terminal state (not still running)
	if result.Result.Status == Running {
		t.Errorf("stress test did not reach a terminal state after %d ticks", result.TickCount)
	}

	// Log detailed results for analysis
	t.Logf("Stress test completed: status=%v reason=%q tick=%d/%d waves_defeated=%d peak_chi=%.0f fengshui=%.1f",
		result.Result.Status, result.Result.Reason, result.TickCount, maxTicks,
		result.Statistics.WavesDefeated, result.Statistics.PeakChi, result.Statistics.FinalFengShui)
}

func TestIntegration_Determinism_SameSeedIdenticalResults(t *testing.T) {
	scenJSON := loadScenarioFile(t, "tutorial.json")
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
