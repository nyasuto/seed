package economy

import (
	"errors"
	"math"
	"testing"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// newTestEngine creates an EconomyEngine with default parameters and the given initial balance.
// It calls t.Fatal if any default parameter loading fails.
func newTestEngine(t *testing.T, initialBalance float64) *EconomyEngine {
	t.Helper()
	pool := NewChiPool(1000)
	if initialBalance > 0 {
		_ = pool.Deposit(initialBalance, Supply, "initial", 0)
	}
	sp, cp, dp, cc, bc := mustLoadDefaults(t)
	return NewEconomyEngine(pool, sp, cp, dp, cc, bc)
}

// mustLoadDefaults loads all default economy parameters, calling t.Fatal on error.
func mustLoadDefaults(t *testing.T) (*SupplyParams, *CostParams, *DeficitParams, *ConstructionCost, *BeastCost) {
	t.Helper()
	sp, err := DefaultSupplyParams()
	if err != nil {
		t.Fatalf("DefaultSupplyParams: %v", err)
	}
	cp, err := DefaultCostParams()
	if err != nil {
		t.Fatalf("DefaultCostParams: %v", err)
	}
	dp, err := DefaultDeficitParams()
	if err != nil {
		t.Fatalf("DefaultDeficitParams: %v", err)
	}
	cc, err := DefaultConstructionCost()
	if err != nil {
		t.Fatalf("DefaultConstructionCost: %v", err)
	}
	bc, err := DefaultBeastCost()
	if err != nil {
		t.Fatalf("DefaultBeastCost: %v", err)
	}
	return sp, cp, dp, cc, bc
}

// mustLoadCostParams loads default CostParams, calling t.Fatal on error.
func mustLoadCostParams(t *testing.T) *CostParams {
	t.Helper()
	cp, err := DefaultCostParams()
	if err != nil {
		t.Fatalf("DefaultCostParams: %v", err)
	}
	return cp
}

func TestNewEconomyEngine(t *testing.T) {
	e := newTestEngine(t,100)
	if e.ChiPool == nil {
		t.Fatal("ChiPool should not be nil")
	}
	if e.SupplyCalc == nil {
		t.Fatal("SupplyCalc should not be nil")
	}
	if e.MaintenanceCalc == nil {
		t.Fatal("MaintenanceCalc should not be nil")
	}
	if e.DeficitProc == nil {
		t.Fatal("DeficitProc should not be nil")
	}
	if e.Construction == nil {
		t.Fatal("Construction should not be nil")
	}
	if e.Beast == nil {
		t.Fatal("Beast should not be nil")
	}
	if e.ChiPool.Balance() != 100 {
		t.Errorf("expected balance 100, got %f", e.ChiPool.Balance())
	}
}

func TestTick_SupplyAndMaintenance(t *testing.T) {
	e := newTestEngine(t,50)

	veins := []fengshui.DragonVein{{ID: 1}}
	roomChis := map[int]*fengshui.RoomChi{
		1: {RoomID: 1, Current: 50, Capacity: 100, Element: types.Wood},
	}
	rooms := []world.Room{
		{ID: 1, TypeID: "beast_room", Level: 1},
	}

	result := e.Tick(1, veins, roomChis, 0.5, rooms, 1, 0)

	if result.Tick != 1 {
		t.Errorf("expected tick 1, got %d", result.Tick)
	}
	if result.Supply <= 0 {
		t.Errorf("expected positive supply, got %f", result.Supply)
	}
	if result.Maintenance.Total <= 0 {
		t.Errorf("expected positive maintenance, got %f", result.Maintenance.Total)
	}
	if result.DeficitResult.Severity != None {
		t.Errorf("expected no deficit, got severity %d", result.DeficitResult.Severity)
	}
	// Balance should be: initial(50) + supply - maintenance (cap is 100, so no clamping).
	expectedBalance := 50 + result.Supply - result.Maintenance.Total
	if math.Abs(result.Balance-expectedBalance) > 0.001 {
		t.Errorf("expected balance ~%f, got %f", expectedBalance, result.Balance)
	}
	if result.ChiPoolCap <= 0 {
		t.Errorf("expected positive chi pool cap, got %f", result.ChiPoolCap)
	}
}

func TestTick_DeficitPenalty(t *testing.T) {
	// Start with very low balance and no supply (no veins).
	e := newTestEngine(t,0)

	rooms := []world.Room{
		{ID: 1, TypeID: "beast_room", Level: 1},
		{ID: 2, TypeID: "trap_room", Level: 1},
		{ID: 3, TypeID: "dragon_den", Level: 1},
	}

	result := e.Tick(1, nil, nil, 0, rooms, 3, 2)

	if result.Supply != 0 {
		t.Errorf("expected zero supply with no veins, got %f", result.Supply)
	}
	if result.Maintenance.Total <= 0 {
		t.Errorf("expected positive maintenance, got %f", result.Maintenance.Total)
	}
	// With zero balance and positive maintenance, should have severe deficit.
	if result.DeficitResult.Severity == None {
		t.Error("expected deficit severity > None")
	}
	if result.DeficitResult.Shortage <= 0 {
		t.Errorf("expected positive shortage, got %f", result.DeficitResult.Shortage)
	}
}

func TestTick_ChiPoolCapRecalculated(t *testing.T) {
	e := newTestEngine(t,50)
	params := e.CostParams

	rooms := []world.Room{
		{ID: 1, TypeID: "chi_storage", Level: 3},
	}

	result := e.Tick(1, nil, nil, 0, rooms, 0, 0)

	expectedCap := params.ChiPoolBaseCap + params.ChiPoolCapPerStorageRoom + 2*params.ChiPoolCapPerStorageLevel
	if math.Abs(result.ChiPoolCap-expectedCap) > 0.001 {
		t.Errorf("expected cap %f, got %f", expectedCap, result.ChiPoolCap)
	}
	if math.Abs(e.ChiPool.Cap-expectedCap) > 0.001 {
		t.Errorf("expected pool cap updated to %f, got %f", expectedCap, e.ChiPool.Cap)
	}
}

func TestTryBuildRoom_Success(t *testing.T) {
	e := newTestEngine(t,200)
	cc := e.Construction

	cost, err := e.TryBuildRoom("beast_room", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedCost := cc.CalcRoomCost("beast_room")
	if cost != expectedCost {
		t.Errorf("expected cost %f, got %f", expectedCost, cost)
	}
	if math.Abs(e.ChiPool.Balance()-(200-expectedCost)) > 0.001 {
		t.Errorf("expected balance %f, got %f", 200-expectedCost, e.ChiPool.Balance())
	}
}

func TestTryBuildRoom_InsufficientChi(t *testing.T) {
	e := newTestEngine(t,1) // Very low balance.

	_, err := e.TryBuildRoom("dragon_den", 1)
	if !errors.Is(err, ErrInsufficientChi) {
		t.Errorf("expected ErrInsufficientChi, got %v", err)
	}
	// Balance should be unchanged.
	if e.ChiPool.Balance() != 1 {
		t.Errorf("expected balance 1, got %f", e.ChiPool.Balance())
	}
}

func TestTryBuildRoom_UnknownType(t *testing.T) {
	e := newTestEngine(t,200)

	_, err := e.TryBuildRoom("nonexistent_room", 1)
	if err == nil {
		t.Error("expected error for unknown room type")
	}
}

func TestTrySummonBeast_Success(t *testing.T) {
	e := newTestEngine(t,200)
	bc := e.Beast

	cost, err := e.TrySummonBeast(types.Wood, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedCost := bc.CalcSummonCost(types.Wood)
	if cost != expectedCost {
		t.Errorf("expected cost %f, got %f", expectedCost, cost)
	}
	if math.Abs(e.ChiPool.Balance()-(200-expectedCost)) > 0.001 {
		t.Errorf("expected balance %f, got %f", 200-expectedCost, e.ChiPool.Balance())
	}
}

func TestTrySummonBeast_InsufficientChi(t *testing.T) {
	e := newTestEngine(t,1)

	_, err := e.TrySummonBeast(types.Metal, 1)
	if !errors.Is(err, ErrInsufficientChi) {
		t.Errorf("expected ErrInsufficientChi, got %v", err)
	}
}

func TestTryUpgradeRoom_Success(t *testing.T) {
	e := newTestEngine(t,200)

	cost, err := e.TryUpgradeRoom("beast_room", 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedCost := e.Construction.CalcUpgradeCost("beast_room", 2)
	if cost != expectedCost {
		t.Errorf("expected cost %f, got %f", expectedCost, cost)
	}
}

func TestTryUpgradeRoom_InsufficientChi(t *testing.T) {
	e := newTestEngine(t,1)

	_, err := e.TryUpgradeRoom("dragon_den", 3, 1)
	if !errors.Is(err, ErrInsufficientChi) {
		t.Errorf("expected ErrInsufficientChi, got %v", err)
	}
}

func TestTryDigCorridor_Success(t *testing.T) {
	e := newTestEngine(t,200)

	cost, err := e.TryDigCorridor(5, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedCost := e.Construction.CalcCorridorCost(5)
	if cost != expectedCost {
		t.Errorf("expected cost %f, got %f", expectedCost, cost)
	}
}

func TestTryDigCorridor_InsufficientChi(t *testing.T) {
	e := newTestEngine(t,1)

	_, err := e.TryDigCorridor(100, 1)
	if !errors.Is(err, ErrInsufficientChi) {
		t.Errorf("expected ErrInsufficientChi, got %v", err)
	}
}

func TestTryDigCorridor_ZeroLength(t *testing.T) {
	e := newTestEngine(t,200)

	cost, err := e.TryDigCorridor(0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost != 0 {
		t.Errorf("expected zero cost for zero length, got %f", cost)
	}
}

func TestTryBuildRoom_TransactionRecorded(t *testing.T) {
	e := newTestEngine(t,200)
	histLen := len(e.ChiPool.History)

	cost, err := e.TryBuildRoom("beast_room", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(e.ChiPool.History) != histLen+1 {
		t.Fatalf("expected %d transactions, got %d", histLen+1, len(e.ChiPool.History))
	}
	tx := e.ChiPool.History[len(e.ChiPool.History)-1]
	if tx.Type != Construction {
		t.Errorf("expected Construction transaction type, got %d", tx.Type)
	}
	if tx.Amount != cost {
		t.Errorf("expected transaction amount %f, got %f", cost, tx.Amount)
	}
	if tx.Tick != 5 {
		t.Errorf("expected tick 5, got %d", tx.Tick)
	}
}

func TestTrySummonBeast_TransactionRecorded(t *testing.T) {
	e := newTestEngine(t,200)
	histLen := len(e.ChiPool.History)

	cost, err := e.TrySummonBeast(types.Fire, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(e.ChiPool.History) != histLen+1 {
		t.Fatalf("expected %d transactions, got %d", histLen+1, len(e.ChiPool.History))
	}
	tx := e.ChiPool.History[len(e.ChiPool.History)-1]
	if tx.Type != BeastSummon {
		t.Errorf("expected BeastSummon transaction type, got %d", tx.Type)
	}
	if tx.Amount != cost {
		t.Errorf("expected transaction amount %f, got %f", cost, tx.Amount)
	}
}

func TestTryUpgradeRoom_TransactionRecorded(t *testing.T) {
	e := newTestEngine(t,200)
	histLen := len(e.ChiPool.History)

	cost, err := e.TryUpgradeRoom("chi_storage", 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(e.ChiPool.History) != histLen+1 {
		t.Fatalf("expected %d transactions, got %d", histLen+1, len(e.ChiPool.History))
	}
	tx := e.ChiPool.History[len(e.ChiPool.History)-1]
	if tx.Type != RoomUpgrade {
		t.Errorf("expected RoomUpgrade transaction type, got %d", tx.Type)
	}
	if tx.Amount != cost {
		t.Errorf("expected transaction amount %f, got %f", cost, tx.Amount)
	}
}
