package server

import (
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// TestIntegration_Tutorial runs a full game with the built-in tutorial
// scenario using a NoAction provider and verifies the GameSummary.
func TestIntegration_Tutorial(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario(tutorial): %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	provider := &noopAI_provider{}
	result, err := gs.RunGame(provider)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	// Tutorial scenario: survive 300 ticks, max_ticks=300
	// With NoAction the game should complete (win by survival or lose by max ticks)
	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}

	summary := gs.Collector().OnGameEnd(&result)

	// Verify summary fields are populated
	if summary.TotalTicks != result.TickCount {
		t.Errorf("TotalTicks = %d, want %d", summary.TotalTicks, result.TickCount)
	}
	if summary.Result != result.Result.Status {
		t.Errorf("Result = %v, want %v", summary.Result, result.Result.Status)
	}

	// CoreHP should be non-negative
	if summary.FinalCoreHP < 0 {
		t.Errorf("FinalCoreHP = %d, want >= 0", summary.FinalCoreHP)
	}

	t.Logf("Tutorial: status=%v reason=%q ticks=%d coreHP=%d peakChi=%.1f fengshui=%.1f waves=%d/%d",
		summary.Result, summary.Reason, summary.TotalTicks,
		summary.FinalCoreHP, summary.PeakChi, summary.FinalFengShui,
		summary.WavesDefeated, summary.TotalWaves)
}

// TestIntegration_Standard runs a full game with the built-in standard
// scenario using a NoAction provider and verifies the GameSummary.
func TestIntegration_Standard(t *testing.T) {
	sc, err := LoadBuiltinScenario("standard")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario(standard): %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	provider := &noopAI_provider{}
	result, err := gs.RunGame(provider)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	// Standard scenario has max_ticks=1000 and multiple waves
	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}

	summary := gs.Collector().OnGameEnd(&result)

	if summary.TotalTicks != result.TickCount {
		t.Errorf("TotalTicks = %d, want %d", summary.TotalTicks, result.TickCount)
	}
	if summary.Result != result.Result.Status {
		t.Errorf("Result = %v, want %v", summary.Result, result.Result.Status)
	}

	// Standard scenario has waves; TotalWaves should be > 0
	if summary.TotalWaves == 0 {
		t.Log("warning: TotalWaves = 0; standard scenario may lack spawn_wave events")
	}

	// PeakChi should be positive (dragon veins supply chi)
	if summary.PeakChi <= 0 {
		t.Errorf("PeakChi = %f, want > 0", summary.PeakChi)
	}

	t.Logf("Standard: status=%v reason=%q ticks=%d coreHP=%d peakChi=%.1f fengshui=%.1f waves=%d/%d dmgDealt=%d dmgRecv=%d",
		summary.Result, summary.Reason, summary.TotalTicks,
		summary.FinalCoreHP, summary.PeakChi, summary.FinalFengShui,
		summary.WavesDefeated, summary.TotalWaves,
		summary.TotalDamageDealt, summary.TotalDamageReceived)
}

// noopAI_provider is a simple ActionProvider that always returns NoAction.
type noopAI_provider struct{}

func (p *noopAI_provider) ProvideActions(_ scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}

func (p *noopAI_provider) OnTickComplete(_ scenario.GameSnapshot) {}
func (p *noopAI_provider) OnGameEnd(_ simulation.RunResult)       {}
