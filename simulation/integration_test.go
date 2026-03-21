package simulation

import (
	"os"
	"testing"
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
