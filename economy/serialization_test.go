package economy

import (
	"encoding/json"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestMarshalUnmarshalEconomyState_RoundTrip(t *testing.T) {
	supplyParams := DefaultSupplyParams()
	costParams := DefaultCostParams()
	deficitParams := DefaultDeficitParams()
	constructionCost := DefaultConstructionCost()
	beastCost := DefaultBeastCost()

	chiPool := NewChiPool(100.0)
	engine := NewEconomyEngine(chiPool, supplyParams, costParams, deficitParams, constructionCost, beastCost)

	// Add some balance and transactions.
	_ = chiPool.Deposit(50.0, Supply, "tick supply", types.Tick(1))
	_ = chiPool.Withdraw(10.0, RoomMaintenance, "room upkeep", types.Tick(1))
	_ = chiPool.Deposit(20.0, Reward, "invasion reward", types.Tick(2))
	_ = chiPool.Withdraw(5.0, BeastMaintenance, "beast upkeep", types.Tick(2))

	// Marshal.
	data, err := MarshalEconomyState(engine)
	if err != nil {
		t.Fatalf("MarshalEconomyState: %v", err)
	}

	// Verify JSON is valid.
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Unmarshal.
	restored, err := UnmarshalEconomyState(data, supplyParams, costParams, deficitParams, constructionCost, beastCost)
	if err != nil {
		t.Fatalf("UnmarshalEconomyState: %v", err)
	}

	// Verify ChiPool state.
	if restored.ChiPool.Current != engine.ChiPool.Current {
		t.Errorf("Current: got %v, want %v", restored.ChiPool.Current, engine.ChiPool.Current)
	}
	if restored.ChiPool.Cap != engine.ChiPool.Cap {
		t.Errorf("Cap: got %v, want %v", restored.ChiPool.Cap, engine.ChiPool.Cap)
	}

	// Verify transaction history.
	if len(restored.ChiPool.History) != len(engine.ChiPool.History) {
		t.Fatalf("History length: got %d, want %d", len(restored.ChiPool.History), len(engine.ChiPool.History))
	}
	for i, got := range restored.ChiPool.History {
		want := engine.ChiPool.History[i]
		if got.Tick != want.Tick {
			t.Errorf("History[%d].Tick: got %v, want %v", i, got.Tick, want.Tick)
		}
		if got.Amount != want.Amount {
			t.Errorf("History[%d].Amount: got %v, want %v", i, got.Amount, want.Amount)
		}
		if got.Type != want.Type {
			t.Errorf("History[%d].Type: got %v, want %v", i, got.Type, want.Type)
		}
		if got.Reason != want.Reason {
			t.Errorf("History[%d].Reason: got %q, want %q", i, got.Reason, want.Reason)
		}
		if got.BalanceAfter != want.BalanceAfter {
			t.Errorf("History[%d].BalanceAfter: got %v, want %v", i, got.BalanceAfter, want.BalanceAfter)
		}
	}

	// Verify calculators are reconstructed (engine can still function).
	if restored.SupplyCalc == nil {
		t.Error("SupplyCalc not reconstructed")
	}
	if restored.MaintenanceCalc == nil {
		t.Error("MaintenanceCalc not reconstructed")
	}
	if restored.DeficitProc == nil {
		t.Error("DeficitProc not reconstructed")
	}
	if restored.Construction == nil {
		t.Error("Construction not reconstructed")
	}
	if restored.Beast == nil {
		t.Error("Beast not reconstructed")
	}
	if restored.CostParams == nil {
		t.Error("CostParams not reconstructed")
	}
}

func TestMarshalUnmarshalEconomyState_EmptyHistory(t *testing.T) {
	supplyParams := DefaultSupplyParams()
	costParams := DefaultCostParams()
	deficitParams := DefaultDeficitParams()
	constructionCost := DefaultConstructionCost()
	beastCost := DefaultBeastCost()

	chiPool := NewChiPool(200.0)
	engine := NewEconomyEngine(chiPool, supplyParams, costParams, deficitParams, constructionCost, beastCost)

	data, err := MarshalEconomyState(engine)
	if err != nil {
		t.Fatalf("MarshalEconomyState: %v", err)
	}

	restored, err := UnmarshalEconomyState(data, supplyParams, costParams, deficitParams, constructionCost, beastCost)
	if err != nil {
		t.Fatalf("UnmarshalEconomyState: %v", err)
	}

	if restored.ChiPool.Current != 0 {
		t.Errorf("Current: got %v, want 0", restored.ChiPool.Current)
	}
	if restored.ChiPool.Cap != 200.0 {
		t.Errorf("Cap: got %v, want 200", restored.ChiPool.Cap)
	}
	if len(restored.ChiPool.History) != 0 {
		t.Errorf("History length: got %d, want 0", len(restored.ChiPool.History))
	}
}

func TestMarshalEconomyState_NilEngine(t *testing.T) {
	_, err := MarshalEconomyState(nil)
	if err == nil {
		t.Error("expected error for nil engine")
	}
}

func TestMarshalEconomyState_NilChiPool(t *testing.T) {
	engine := &EconomyEngine{}
	_, err := MarshalEconomyState(engine)
	if err == nil {
		t.Error("expected error for nil chi pool")
	}
}

func TestUnmarshalEconomyState_InvalidJSON(t *testing.T) {
	supplyParams := DefaultSupplyParams()
	costParams := DefaultCostParams()
	deficitParams := DefaultDeficitParams()
	constructionCost := DefaultConstructionCost()
	beastCost := DefaultBeastCost()

	_, err := UnmarshalEconomyState([]byte("invalid"), supplyParams, costParams, deficitParams, constructionCost, beastCost)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMarshalUnmarshalEconomyState_WithAllTransactionTypes(t *testing.T) {
	supplyParams := DefaultSupplyParams()
	costParams := DefaultCostParams()
	deficitParams := DefaultDeficitParams()
	constructionCost := DefaultConstructionCost()
	beastCost := DefaultBeastCost()

	chiPool := NewChiPool(1000.0)
	engine := NewEconomyEngine(chiPool, supplyParams, costParams, deficitParams, constructionCost, beastCost)

	// Add transactions of all types.
	txTypes := []struct {
		amount  float64
		txType  TransactionType
		reason  string
		tick    types.Tick
		deposit bool
	}{
		{100, Supply, "supply", 1, true},
		{10, RoomMaintenance, "room", 1, false},
		{5, BeastMaintenance, "beast", 2, false},
		{3, TrapMaintenance, "trap", 2, false},
		{50, Reward, "reward", 3, true},
		{8, Theft, "theft", 3, false},
		{20, Construction, "build", 4, false},
		{15, BeastSummon, "summon", 4, false},
		{12, RoomUpgrade, "upgrade", 5, false},
		{2, Deficit, "deficit", 5, false},
	}

	for _, tx := range txTypes {
		if tx.deposit {
			_ = chiPool.Deposit(tx.amount, tx.txType, tx.reason, tx.tick)
		} else {
			_ = chiPool.Withdraw(tx.amount, tx.txType, tx.reason, tx.tick)
		}
	}

	data, err := MarshalEconomyState(engine)
	if err != nil {
		t.Fatalf("MarshalEconomyState: %v", err)
	}

	restored, err := UnmarshalEconomyState(data, supplyParams, costParams, deficitParams, constructionCost, beastCost)
	if err != nil {
		t.Fatalf("UnmarshalEconomyState: %v", err)
	}

	if len(restored.ChiPool.History) != len(engine.ChiPool.History) {
		t.Fatalf("History length: got %d, want %d", len(restored.ChiPool.History), len(engine.ChiPool.History))
	}

	// Verify all transaction types are preserved.
	for i, got := range restored.ChiPool.History {
		want := engine.ChiPool.History[i]
		if got.Type != want.Type {
			t.Errorf("History[%d].Type: got %v, want %v", i, got.Type, want.Type)
		}
	}
}
