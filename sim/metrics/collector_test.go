package metrics

import (
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// mockAction implements simulation.PlayerAction for testing.
type mockAction struct {
	actionType string
}

func (a mockAction) ActionType() string { return a.actionType }

func TestCollector_OnTick_TracksPeakBeasts(t *testing.T) {
	c := NewCollector()

	snapshots := []scenario.GameSnapshot{
		{BeastCount: 2},
		{BeastCount: 5},
		{BeastCount: 3},
	}
	for _, s := range snapshots {
		c.OnTick(s, nil)
	}

	result := &simulation.RunResult{
		Result:    simulation.GameResult{Status: simulation.Won, Reason: "test"},
		TickCount: 3,
	}
	summary := c.OnGameEnd(result)

	if summary.PeakBeasts != 5 {
		t.Errorf("PeakBeasts = %d, want 5", summary.PeakBeasts)
	}
}

func TestCollector_OnTick_CountsRoomsBuilt(t *testing.T) {
	c := NewCollector()

	actions := []simulation.PlayerAction{
		mockAction{actionType: "dig_room"},
		mockAction{actionType: "no_action"},
		mockAction{actionType: "dig_room"},
	}
	c.OnTick(scenario.GameSnapshot{}, actions)

	result := &simulation.RunResult{
		Result:    simulation.GameResult{Status: simulation.Won},
		TickCount: 1,
	}
	summary := c.OnGameEnd(result)

	if summary.RoomsBuilt != 2 {
		t.Errorf("RoomsBuilt = %d, want 2", summary.RoomsBuilt)
	}
}

func TestCollector_OnGameEnd_PopulatesSummary(t *testing.T) {
	c := NewCollector()

	c.OnTick(scenario.GameSnapshot{
		CoreHP:     80,
		BeastCount: 3,
		TotalWaves: 5,
	}, nil)
	c.OnTick(scenario.GameSnapshot{
		CoreHP:     60,
		BeastCount: 2,
		TotalWaves: 5,
	}, nil)

	result := &simulation.RunResult{
		Result:    simulation.GameResult{Status: simulation.Lost, Reason: "core destroyed"},
		TickCount: 2,
		Statistics: simulation.RunStatistics{
			PeakChi:        150.0,
			WavesDefeated:  3,
			FinalFengShui:  0.75,
			Evolutions:     1,
			DamageDealt:    200,
			DamageReceived: 100,
			DeficitTicks:   5,
		},
	}
	summary := c.OnGameEnd(result)

	if summary.Result != simulation.Lost {
		t.Errorf("Result = %v, want Lost", summary.Result)
	}
	if summary.Reason != "core destroyed" {
		t.Errorf("Reason = %q, want %q", summary.Reason, "core destroyed")
	}
	if summary.TotalTicks != 2 {
		t.Errorf("TotalTicks = %d, want 2", summary.TotalTicks)
	}
	if summary.FinalCoreHP != 60 {
		t.Errorf("FinalCoreHP = %d, want 60", summary.FinalCoreHP)
	}
	if summary.PeakChi != 150.0 {
		t.Errorf("PeakChi = %f, want 150.0", summary.PeakChi)
	}
	if summary.FinalFengShui != 0.75 {
		t.Errorf("FinalFengShui = %f, want 0.75", summary.FinalFengShui)
	}
	if summary.WavesDefeated != 3 {
		t.Errorf("WavesDefeated = %d, want 3", summary.WavesDefeated)
	}
	if summary.TotalWaves != 5 {
		t.Errorf("TotalWaves = %d, want 5", summary.TotalWaves)
	}
	if summary.PeakBeasts != 3 {
		t.Errorf("PeakBeasts = %d, want 3", summary.PeakBeasts)
	}
	if summary.TotalDamageDealt != 200 {
		t.Errorf("TotalDamageDealt = %d, want 200", summary.TotalDamageDealt)
	}
	if summary.TotalDamageReceived != 100 {
		t.Errorf("TotalDamageReceived = %d, want 100", summary.TotalDamageReceived)
	}
	if summary.DeficitTicks != 5 {
		t.Errorf("DeficitTicks = %d, want 5", summary.DeficitTicks)
	}
	if summary.Evolutions != 1 {
		t.Errorf("Evolutions = %d, want 1", summary.Evolutions)
	}
}

func TestNewCollector_Fresh(t *testing.T) {
	c := NewCollector()
	if c.tickCount != 0 {
		t.Errorf("tickCount = %d, want 0", c.tickCount)
	}
	if c.peakBeasts != 0 {
		t.Errorf("peakBeasts = %d, want 0", c.peakBeasts)
	}
}
