package economy

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func defaultTestDeficitParams() *DeficitParams {
	return &DeficitParams{
		MildThreshold:           0.3,
		ModerateThreshold:       0.7,
		MildGrowthPenalty:       0.5,
		ModerateCapacityPenalty: 0.8,
		SevereHPDrain:           5,
		SevereTrapDisable:       true,
	}
}

func TestProcessDeficit_NoDeficit(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)
	_ = pool.Deposit(100.0, Supply, "initial", types.Tick(0))

	maintenance := MaintenanceBreakdown{
		RoomCost:  5.0,
		BeastCost: 3.0,
		TrapCost:  2.0,
		Total:     10.0,
	}

	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != None {
		t.Errorf("expected severity None, got %d", result.Severity)
	}
	if result.Shortage != 0 {
		t.Errorf("expected shortage 0, got %f", result.Shortage)
	}
	if result.GrowthPenalty != 1.0 {
		t.Errorf("expected growth penalty 1.0, got %f", result.GrowthPenalty)
	}
	if result.CapacityPenalty != 1.0 {
		t.Errorf("expected capacity penalty 1.0, got %f", result.CapacityPenalty)
	}
	if result.TrapDisabled {
		t.Error("expected traps not disabled")
	}
	if result.BeastHPDrain != 0 {
		t.Errorf("expected beast hp drain 0, got %d", result.BeastHPDrain)
	}
	// Balance should be 100 - 10 = 90
	if pool.Balance() != 90.0 {
		t.Errorf("expected balance 90.0, got %f", pool.Balance())
	}
}

func TestProcessDeficit_MildDeficit(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)
	// Deposit 8.0, maintenance 10.0 → shortage 2.0, ratio 0.2 < 0.3 → Mild
	_ = pool.Deposit(8.0, Supply, "initial", types.Tick(0))

	maintenance := MaintenanceBreakdown{Total: 10.0}

	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != Mild {
		t.Errorf("expected severity Mild, got %d", result.Severity)
	}
	if result.Shortage != 2.0 {
		t.Errorf("expected shortage 2.0, got %f", result.Shortage)
	}
	if result.GrowthPenalty != 0.5 {
		t.Errorf("expected growth penalty 0.5, got %f", result.GrowthPenalty)
	}
	if result.CapacityPenalty != 1.0 {
		t.Errorf("expected capacity penalty 1.0, got %f", result.CapacityPenalty)
	}
	if result.TrapDisabled {
		t.Error("expected traps not disabled")
	}
	if pool.Balance() != 0 {
		t.Errorf("expected balance 0, got %f", pool.Balance())
	}
}

func TestProcessDeficit_ModerateDeficit(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)
	// Deposit 5.0, maintenance 10.0 → shortage 5.0, ratio 0.5 → Moderate (0.3 <= 0.5 < 0.7)
	_ = pool.Deposit(5.0, Supply, "initial", types.Tick(0))

	maintenance := MaintenanceBreakdown{Total: 10.0}

	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != Moderate {
		t.Errorf("expected severity Moderate, got %d", result.Severity)
	}
	if result.Shortage != 5.0 {
		t.Errorf("expected shortage 5.0, got %f", result.Shortage)
	}
	if result.GrowthPenalty != 0.0 {
		t.Errorf("expected growth penalty 0.0, got %f", result.GrowthPenalty)
	}
	if result.CapacityPenalty != 0.8 {
		t.Errorf("expected capacity penalty 0.8, got %f", result.CapacityPenalty)
	}
	if result.TrapDisabled {
		t.Error("expected traps not disabled for moderate deficit")
	}
	if result.BeastHPDrain != 0 {
		t.Errorf("expected beast hp drain 0 for moderate, got %d", result.BeastHPDrain)
	}
}

func TestProcessDeficit_SevereDeficit(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)
	// Deposit 2.0, maintenance 10.0 → shortage 8.0, ratio 0.8 → Severe (>= 0.7)
	_ = pool.Deposit(2.0, Supply, "initial", types.Tick(0))

	maintenance := MaintenanceBreakdown{Total: 10.0}

	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != Severe {
		t.Errorf("expected severity Severe, got %d", result.Severity)
	}
	if result.Shortage != 8.0 {
		t.Errorf("expected shortage 8.0, got %f", result.Shortage)
	}
	if result.GrowthPenalty != 0.0 {
		t.Errorf("expected growth penalty 0.0, got %f", result.GrowthPenalty)
	}
	if result.CapacityPenalty != 0.8 {
		t.Errorf("expected capacity penalty 0.8, got %f", result.CapacityPenalty)
	}
	if !result.TrapDisabled {
		t.Error("expected traps disabled for severe deficit")
	}
	if result.BeastHPDrain != 5 {
		t.Errorf("expected beast hp drain 5, got %d", result.BeastHPDrain)
	}
}

func TestProcessDeficit_RecoveryFromDeficit(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)

	// First tick: severe deficit
	_ = pool.Deposit(2.0, Supply, "initial", types.Tick(0))
	maintenance := MaintenanceBreakdown{Total: 10.0}
	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != Severe {
		t.Fatalf("expected severe deficit on tick 1, got %d", result.Severity)
	}

	// Second tick: supply exceeds maintenance → no deficit
	_ = pool.Deposit(20.0, Supply, "resupply", types.Tick(2))
	result = dp.ProcessDeficit(pool, maintenance, types.Tick(3))

	if result.Severity != None {
		t.Errorf("expected no deficit after recovery, got %d", result.Severity)
	}
	if result.GrowthPenalty != 1.0 {
		t.Errorf("expected growth penalty 1.0 after recovery, got %f", result.GrowthPenalty)
	}
	if result.CapacityPenalty != 1.0 {
		t.Errorf("expected capacity penalty 1.0 after recovery, got %f", result.CapacityPenalty)
	}
	if result.TrapDisabled {
		t.Error("expected traps not disabled after recovery")
	}
	// Balance: 20.0 - 10.0 = 10.0
	if pool.Balance() != 10.0 {
		t.Errorf("expected balance 10.0 after recovery, got %f", pool.Balance())
	}
}

func TestProcessDeficit_ZeroMaintenance(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)

	maintenance := MaintenanceBreakdown{Total: 0.0}
	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != None {
		t.Errorf("expected no deficit for zero maintenance, got %d", result.Severity)
	}
}

func TestProcessDeficit_ZeroBalanceTotalDeficit(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)
	// Balance is 0, maintenance 10 → shortage 10, ratio 1.0 → Severe

	maintenance := MaintenanceBreakdown{Total: 10.0}
	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != Severe {
		t.Errorf("expected severe deficit, got %d", result.Severity)
	}
	if result.Shortage != 10.0 {
		t.Errorf("expected shortage 10.0, got %f", result.Shortage)
	}
}

func TestDeficitParams_LoadFromJSON(t *testing.T) {
	p, err := DefaultDeficitParams()
	if err != nil {
		t.Fatalf("DefaultDeficitParams: %v", err)
	}
	if p.MildThreshold != 0.3 {
		t.Errorf("expected mild threshold 0.3, got %f", p.MildThreshold)
	}
	if p.ModerateThreshold != 0.7 {
		t.Errorf("expected moderate threshold 0.7, got %f", p.ModerateThreshold)
	}
	if p.SevereHPDrain != 5 {
		t.Errorf("expected severe hp drain 5, got %d", p.SevereHPDrain)
	}
	if !p.SevereTrapDisable {
		t.Error("expected severe trap disable true")
	}
}

func TestDeficitParams_InvalidJSON(t *testing.T) {
	_, err := LoadDeficitParams([]byte("invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestProcessDeficit_BoundaryMildThreshold(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)
	// Deposit 7.0, maintenance 10.0 → shortage 3.0, ratio 0.3 → exactly at mild threshold → Moderate
	_ = pool.Deposit(7.0, Supply, "initial", types.Tick(0))

	maintenance := MaintenanceBreakdown{Total: 10.0}
	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != Moderate {
		t.Errorf("expected Moderate at boundary 0.3, got %d", result.Severity)
	}
}

func TestProcessDeficit_BoundaryModerateThreshold(t *testing.T) {
	dp := NewDeficitProcessor(defaultTestDeficitParams())
	pool := NewChiPool(200.0)
	// Deposit 3.0, maintenance 10.0 → shortage 7.0, ratio 0.7 → exactly at moderate threshold → Severe
	_ = pool.Deposit(3.0, Supply, "initial", types.Tick(0))

	maintenance := MaintenanceBreakdown{Total: 10.0}
	result := dp.ProcessDeficit(pool, maintenance, types.Tick(1))

	if result.Severity != Severe {
		t.Errorf("expected Severe at boundary 0.7, got %d", result.Severity)
	}
}
