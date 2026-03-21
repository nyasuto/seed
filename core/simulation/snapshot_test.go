package simulation

import (
	"testing"

	"github.com/nyasuto/seed/core/economy"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/types"
)

func TestBuildSnapshot_EmptyState(t *testing.T) {
	state := &GameState{}
	snap := BuildSnapshot(state)

	if snap.Tick != 0 {
		t.Errorf("Tick = %d, want 0", snap.Tick)
	}
	if snap.CoreHP != 0 {
		t.Errorf("CoreHP = %d, want 0", snap.CoreHP)
	}
	if snap.ChiPoolBalance != 0 {
		t.Errorf("ChiPoolBalance = %f, want 0", snap.ChiPoolBalance)
	}
	if snap.BeastCount != 0 {
		t.Errorf("BeastCount = %d, want 0", snap.BeastCount)
	}
	if snap.AliveBeasts != 0 {
		t.Errorf("AliveBeasts = %d, want 0", snap.AliveBeasts)
	}
	if snap.DefeatedWaves != 0 {
		t.Errorf("DefeatedWaves = %d, want 0", snap.DefeatedWaves)
	}
	if snap.TotalWaves != 0 {
		t.Errorf("TotalWaves = %d, want 0", snap.TotalWaves)
	}
}

func TestBuildSnapshot_ProgressFields(t *testing.T) {
	state := &GameState{
		Progress: &scenario.ScenarioProgress{
			CurrentTick: 42,
			CoreHP:      80,
		},
		ConsecutiveDeficitTicks: 5,
	}

	snap := BuildSnapshot(state)

	if snap.Tick != 42 {
		t.Errorf("Tick = %d, want 42", snap.Tick)
	}
	if snap.CoreHP != 80 {
		t.Errorf("CoreHP = %d, want 80", snap.CoreHP)
	}
	if snap.ConsecutiveDeficitTicks != 5 {
		t.Errorf("ConsecutiveDeficitTicks = %d, want 5", snap.ConsecutiveDeficitTicks)
	}
}

func TestBuildSnapshot_ChiPoolBalance(t *testing.T) {
	pool := economy.NewChiPool(1000)
	_ = pool.Deposit(350, economy.Supply, "test", 0)

	sp, err := economy.DefaultSupplyParams()
	if err != nil {
		t.Fatalf("DefaultSupplyParams: %v", err)
	}
	cp, err := economy.DefaultCostParams()
	if err != nil {
		t.Fatalf("DefaultCostParams: %v", err)
	}
	dp, err := economy.DefaultDeficitParams()
	if err != nil {
		t.Fatalf("DefaultDeficitParams: %v", err)
	}
	cc, err := economy.DefaultConstructionCost()
	if err != nil {
		t.Fatalf("DefaultConstructionCost: %v", err)
	}
	bc, err := economy.DefaultBeastCost()
	if err != nil {
		t.Fatalf("DefaultBeastCost: %v", err)
	}
	engine := economy.NewEconomyEngine(pool, sp, cp, dp, cc, bc)

	state := &GameState{
		EconomyEngine: engine,
	}

	snap := BuildSnapshot(state)

	if snap.ChiPoolBalance != 350 {
		t.Errorf("ChiPoolBalance = %f, want 350", snap.ChiPoolBalance)
	}
}

func TestBuildSnapshot_BeastCounts(t *testing.T) {
	species := &senju.Species{
		ID: "fire_lizard", Name: "Fire Lizard", Element: types.Fire,
		BaseHP: 100, BaseATK: 20, BaseDEF: 10, BaseSPD: 15,
	}

	alive1 := senju.NewBeast(1, species, 0)
	alive2 := senju.NewBeast(2, species, 0)
	stunned := senju.NewBeast(3, species, 0)
	stunned.State = senju.Stunned
	stunned.HP = 0

	state := &GameState{
		Beasts: []*senju.Beast{alive1, alive2, stunned},
	}

	snap := BuildSnapshot(state)

	if snap.BeastCount != 3 {
		t.Errorf("BeastCount = %d, want 3", snap.BeastCount)
	}
	if snap.AliveBeasts != 2 {
		t.Errorf("AliveBeasts = %d, want 2", snap.AliveBeasts)
	}
}

func TestBuildSnapshot_WaveCounts(t *testing.T) {
	waves := []*invasion.InvasionWave{
		{ID: 1, State: invasion.Completed},
		{ID: 2, State: invasion.Active},
		{ID: 3, State: invasion.Failed},
		{ID: 4, State: invasion.Pending},
	}

	state := &GameState{
		Waves: waves,
	}

	snap := BuildSnapshot(state)

	if snap.TotalWaves != 4 {
		t.Errorf("TotalWaves = %d, want 4", snap.TotalWaves)
	}
	if snap.DefeatedWaves != 1 {
		t.Errorf("DefeatedWaves = %d, want 1 (only Completed counts)", snap.DefeatedWaves)
	}
}
