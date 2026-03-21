package economy

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/types"
)

func TestProcessInvasionEvents_RewardOnDefeat(t *testing.T) {
	pool := NewChiPool(1000)
	pool.Deposit(100, Supply, "initial", 0)
	proc := NewInvasionEconomyProcessor()

	events := []invasion.InvasionEvent{
		{Type: invasion.InvaderDefeated, Tick: 10, RewardChi: 25.0},
		{Type: invasion.InvaderDefeated, Tick: 10, RewardChi: 30.0},
	}

	summary := proc.ProcessInvasionEvents(events, pool, 10)

	if summary.RewardChi != 55.0 {
		t.Errorf("RewardChi = %v, want 55.0", summary.RewardChi)
	}
	if summary.StolenChi != 0 {
		t.Errorf("StolenChi = %v, want 0", summary.StolenChi)
	}
	if summary.NetChi != 55.0 {
		t.Errorf("NetChi = %v, want 55.0", summary.NetChi)
	}
	if pool.Balance() != 155.0 {
		t.Errorf("Balance = %v, want 155.0", pool.Balance())
	}
}

func TestProcessInvasionEvents_TheftOnEscape(t *testing.T) {
	pool := NewChiPool(1000)
	pool.Deposit(100, Supply, "initial", 0)
	proc := NewInvasionEconomyProcessor()

	events := []invasion.InvasionEvent{
		{Type: invasion.InvaderEscaped, Tick: 10, StolenChi: 40.0},
	}

	summary := proc.ProcessInvasionEvents(events, pool, 10)

	if summary.StolenChi != 40.0 {
		t.Errorf("StolenChi = %v, want 40.0", summary.StolenChi)
	}
	if summary.RewardChi != 0 {
		t.Errorf("RewardChi = %v, want 0", summary.RewardChi)
	}
	if summary.NetChi != -40.0 {
		t.Errorf("NetChi = %v, want -40.0", summary.NetChi)
	}
	if pool.Balance() != 60.0 {
		t.Errorf("Balance = %v, want 60.0", pool.Balance())
	}
}

func TestProcessInvasionEvents_RewardAndTheftNetBalance(t *testing.T) {
	pool := NewChiPool(1000)
	pool.Deposit(200, Supply, "initial", 0)
	proc := NewInvasionEconomyProcessor()

	events := []invasion.InvasionEvent{
		{Type: invasion.InvaderDefeated, Tick: 5, RewardChi: 20.0},
		{Type: invasion.InvaderEscaped, Tick: 5, StolenChi: 50.0},
		{Type: invasion.InvaderDefeated, Tick: 5, RewardChi: 15.0},
	}

	summary := proc.ProcessInvasionEvents(events, pool, 5)

	if summary.RewardChi != 35.0 {
		t.Errorf("RewardChi = %v, want 35.0", summary.RewardChi)
	}
	if summary.StolenChi != 50.0 {
		t.Errorf("StolenChi = %v, want 50.0", summary.StolenChi)
	}
	if summary.NetChi != -15.0 {
		t.Errorf("NetChi = %v, want -15.0", summary.NetChi)
	}
	// 200 + 35 (reward) - 50 (theft) = 185
	if pool.Balance() != 185.0 {
		t.Errorf("Balance = %v, want 185.0", pool.Balance())
	}
}

func TestProcessInvasionEvents_MultipleEvents(t *testing.T) {
	pool := NewChiPool(1000)
	pool.Deposit(500, Supply, "initial", 0)
	proc := NewInvasionEconomyProcessor()

	events := []invasion.InvasionEvent{
		{Type: invasion.InvaderDefeated, Tick: 20, RewardChi: 10.0},
		{Type: invasion.InvaderDefeated, Tick: 20, RewardChi: 15.0},
		{Type: invasion.InvaderDefeated, Tick: 20, RewardChi: 20.0},
		{Type: invasion.InvaderEscaped, Tick: 20, StolenChi: 30.0},
		{Type: invasion.InvaderEscaped, Tick: 20, StolenChi: 25.0},
		{Type: invasion.BeastDefeated, Tick: 20},
		{Type: invasion.BeastDefeated, Tick: 20},
		{Type: invasion.CombatOccurred, Tick: 20, Damage: 5}, // ignored
		{Type: invasion.WaveStarted, Tick: 20},                // ignored
	}

	summary := proc.ProcessInvasionEvents(events, pool, 20)

	if summary.RewardChi != 45.0 {
		t.Errorf("RewardChi = %v, want 45.0", summary.RewardChi)
	}
	if summary.StolenChi != 55.0 {
		t.Errorf("StolenChi = %v, want 55.0", summary.StolenChi)
	}
	if summary.NetChi != -10.0 {
		t.Errorf("NetChi = %v, want -10.0", summary.NetChi)
	}
	if summary.BeastsLost != 2 {
		t.Errorf("BeastsLost = %v, want 2", summary.BeastsLost)
	}
	// 500 + 45 - 55 = 490
	if pool.Balance() != 490.0 {
		t.Errorf("Balance = %v, want 490.0", pool.Balance())
	}
}

func TestProcessInvasionEvents_EmptyEvents(t *testing.T) {
	pool := NewChiPool(1000)
	pool.Deposit(100, Supply, "initial", 0)
	proc := NewInvasionEconomyProcessor()

	summary := proc.ProcessInvasionEvents(nil, pool, 1)

	if summary.RewardChi != 0 || summary.StolenChi != 0 || summary.NetChi != 0 || summary.BeastsLost != 0 {
		t.Errorf("expected zero summary for empty events, got %+v", summary)
	}
	if pool.Balance() != 100.0 {
		t.Errorf("Balance = %v, want 100.0 (unchanged)", pool.Balance())
	}
}

func TestProcessInvasionEvents_TheftExceedsBalance(t *testing.T) {
	pool := NewChiPool(1000)
	pool.Deposit(20, Supply, "initial", 0)
	proc := NewInvasionEconomyProcessor()

	events := []invasion.InvasionEvent{
		{Type: invasion.InvaderEscaped, Tick: 10, StolenChi: 50.0},
	}

	summary := proc.ProcessInvasionEvents(events, pool, 10)

	if summary.StolenChi != 50.0 {
		t.Errorf("StolenChi = %v, want 50.0", summary.StolenChi)
	}
	// ChiPool does partial withdrawal; balance goes to 0.
	if pool.Balance() != 0 {
		t.Errorf("Balance = %v, want 0 (partial withdrawal)", pool.Balance())
	}
}

func TestProcessInvasionEvents_TransactionHistory(t *testing.T) {
	pool := NewChiPool(1000)
	proc := NewInvasionEconomyProcessor()

	events := []invasion.InvasionEvent{
		{Type: invasion.InvaderDefeated, Tick: 5, RewardChi: 30.0},
		{Type: invasion.InvaderEscaped, Tick: 5, StolenChi: 10.0},
	}

	proc.ProcessInvasionEvents(events, pool, 5)

	if len(pool.History) != 2 {
		t.Fatalf("History length = %v, want 2", len(pool.History))
	}
	if pool.History[0].Type != Reward {
		t.Errorf("History[0].Type = %v, want Reward", pool.History[0].Type)
	}
	if pool.History[0].Amount != 30.0 {
		t.Errorf("History[0].Amount = %v, want 30.0", pool.History[0].Amount)
	}
	if pool.History[0].Tick != types.Tick(5) {
		t.Errorf("History[0].Tick = %v, want 5", pool.History[0].Tick)
	}
	if pool.History[1].Type != Theft {
		t.Errorf("History[1].Type = %v, want Theft", pool.History[1].Type)
	}
	if pool.History[1].Amount != 10.0 {
		t.Errorf("History[1].Amount = %v, want 10.0", pool.History[1].Amount)
	}
}
