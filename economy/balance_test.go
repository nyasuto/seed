package economy

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// standardBalanceConfig creates a standard dungeon configuration for balance testing:
// 6 rooms, 3 beasts, 1 trap, 1 dragon vein, average feng shui.
func standardBalanceConfig() (
	rooms []world.Room,
	beastCount int,
	trapCount int,
	veins []fengshui.DragonVein,
	roomChis map[int]*fengshui.RoomChi,
	caveScore float64,
) {
	rooms = []world.Room{
		{ID: 1, TypeID: "beast_room", Level: 1},
		{ID: 2, TypeID: "beast_room", Level: 1},
		{ID: 3, TypeID: "trap_room", Level: 1},
		{ID: 4, TypeID: "chi_storage", Level: 1},
		{ID: 5, TypeID: "recovery_room", Level: 1},
		{ID: 6, TypeID: "warehouse", Level: 1},
	}
	beastCount = 3
	trapCount = 1
	veins = []fengshui.DragonVein{{ID: 1}}
	roomChis = map[int]*fengshui.RoomChi{
		1: {RoomID: 1, Current: 50, Capacity: 100, Element: types.Wood},
		2: {RoomID: 2, Current: 50, Capacity: 100, Element: types.Fire},
		3: {RoomID: 3, Current: 50, Capacity: 100, Element: types.Earth},
		4: {RoomID: 4, Current: 50, Capacity: 100, Element: types.Metal},
		5: {RoomID: 5, Current: 50, Capacity: 100, Element: types.Water},
		6: {RoomID: 6, Current: 50, Capacity: 100, Element: types.Wood},
	}
	caveScore = 0.5
	return
}

// assumedInvasionInterval is the assumed number of ticks between invasion waves,
// used for balance recovery verification.
const assumedInvasionInterval = 15

func TestBalance_MaintenanceToSupplyRatio(t *testing.T) {
	e := newTestEngine(0)
	rooms, beastCount, trapCount, veins, roomChis, caveScore := standardBalanceConfig()

	const totalTicks = 200
	var totalSupply, totalMaintenance float64

	for tick := types.Tick(1); tick <= totalTicks; tick++ {
		result := e.Tick(tick, veins, roomChis, caveScore, rooms, beastCount, trapCount)
		totalSupply += result.Supply
		totalMaintenance += result.Maintenance.Total
	}

	// Maintenance should consume a significant fraction of supply.
	// This ensures resources are under pressure (D002 principle 3).
	maintenanceRatio := totalMaintenance / totalSupply
	if maintenanceRatio < 0.3 {
		t.Errorf("maintenance/supply ratio too low (%.3f): resources not scarce enough", maintenanceRatio)
	}
	t.Logf("maintenance/supply ratio: %.3f (supply=%.1f, maintenance=%.1f, net=%.1f)",
		maintenanceRatio, totalSupply, totalMaintenance, totalSupply-totalMaintenance)
}

func TestBalance_ConstructionCausesSignificantDrop(t *testing.T) {
	e := newTestEngine(0)
	rooms, beastCount, trapCount, veins, roomChis, caveScore := standardBalanceConfig()

	// Let chi accumulate for 20 ticks.
	for tick := types.Tick(1); tick <= 20; tick++ {
		e.Tick(tick, veins, roomChis, caveScore, rooms, beastCount, trapCount)
	}

	balanceBefore := e.ChiPool.Balance()
	cost, err := e.TryBuildRoom("dragon_den", 21)
	if err != nil {
		t.Fatalf("expected to afford dragon_den after 20 ticks, balance=%.1f: %v", balanceBefore, err)
	}
	balanceAfter := e.ChiPool.Balance()

	// The construction cost should consume a large portion of accumulated balance.
	dropRatio := cost / balanceBefore
	if dropRatio < 0.3 {
		t.Errorf("construction cost (%.1f) should be significant relative to balance (%.1f), ratio=%.2f",
			cost, balanceBefore, dropRatio)
	}
	t.Logf("construction drop: %.1f -> %.1f (cost=%.1f, ratio=%.2f)",
		balanceBefore, balanceAfter, cost, dropRatio)
}

func TestBalance_RecoverySlowerThanInvasionInterval(t *testing.T) {
	e := newTestEngine(0)
	rooms, beastCount, trapCount, veins, roomChis, caveScore := standardBalanceConfig()

	const totalTicks = 200
	var totalSupply, totalMaintenance float64

	for tick := types.Tick(1); tick <= totalTicks; tick++ {
		result := e.Tick(tick, veins, roomChis, caveScore, rooms, beastCount, trapCount)
		totalSupply += result.Supply
		totalMaintenance += result.Maintenance.Total
	}

	avgNetPerTick := (totalSupply - totalMaintenance) / totalTicks
	if avgNetPerTick <= 0 {
		t.Fatalf("expected positive net income per tick, got %.3f", avgNetPerTick)
	}

	// After building the most expensive room (dragon_den = 50 chi),
	// recovery time should exceed invasion wave interval.
	cc := DefaultConstructionCost()
	dragonDenCost := cc.CalcRoomCost("dragon_den")
	recoveryTicks := int(dragonDenCost / avgNetPerTick)

	if recoveryTicks <= assumedInvasionInterval {
		t.Errorf("recovery ticks (%d) should exceed invasion interval (%d); net/tick=%.3f, cost=%.1f",
			recoveryTicks, assumedInvasionInterval, avgNetPerTick, dragonDenCost)
	}
	t.Logf("recovery ticks: %d (invasion interval: %d, net/tick=%.3f)",
		recoveryTicks, assumedInvasionInterval, avgNetPerTick)
}

func TestBalance_CannotMaxAllByTick200(t *testing.T) {
	rooms, _, _, _, _, _ := standardBalanceConfig()

	const totalTicks = 200
	const maxLevel = 5

	// Calculate total cost to upgrade all 6 rooms to max level and summon 3 beasts.
	cc := DefaultConstructionCost()
	bc := DefaultBeastCost()

	var totalUpgradeCost float64
	for _, r := range rooms {
		for lv := 1; lv < maxLevel; lv++ {
			totalUpgradeCost += cc.CalcUpgradeCost(r.TypeID, lv)
		}
	}

	var totalSummonCost float64
	for _, el := range []types.Element{types.Wood, types.Fire, types.Earth} {
		totalSummonCost += bc.CalcSummonCost(el)
	}

	totalCostToMax := totalUpgradeCost + totalSummonCost

	// Calculate total net income over 200 ticks.
	e := newTestEngine(0)
	rooms2, beastCount, trapCount, veins, roomChis, caveScore := standardBalanceConfig()
	var totalSupply, totalMaintenance float64

	for tick := types.Tick(1); tick <= totalTicks; tick++ {
		result := e.Tick(tick, veins, roomChis, caveScore, rooms2, beastCount, trapCount)
		totalSupply += result.Supply
		totalMaintenance += result.Maintenance.Total
	}

	totalNetIncome := totalSupply - totalMaintenance

	// Even if every chi earned were spent on upgrades (ignoring cap constraints),
	// the total income should not cover maxing everything.
	if totalNetIncome >= totalCostToMax {
		t.Errorf("total net income (%.1f) should be less than total cost to max all (%.1f)",
			totalNetIncome, totalCostToMax)
	}
	t.Logf("net income: %.1f, cost to max: %.1f (deficit=%.1f)",
		totalNetIncome, totalCostToMax, totalCostToMax-totalNetIncome)
}

func TestBalance_DeficitRecoveryCycle(t *testing.T) {
	// Simulate dragon vein loss (e.g., during invasion damage) → deficit → recovery.
	// Phase 1: No veins, maintenance drains balance to deficit.
	// Phase 2: Vein restored, supply recovers balance.
	e := newTestEngine(20) // Small initial balance.
	rooms, beastCount, trapCount, _, roomChis, caveScore := standardBalanceConfig()

	noVeins := []fengshui.DragonVein{}
	deficitHit := false
	var deficitTick types.Tick

	// Phase 1: Run without veins until deficit occurs.
	for tick := types.Tick(1); tick <= 50; tick++ {
		result := e.Tick(tick, noVeins, roomChis, caveScore, rooms, beastCount, trapCount)
		if result.DeficitResult.Severity != None {
			deficitHit = true
			deficitTick = tick
			// Verify penalty is applied.
			if result.DeficitResult.GrowthPenalty >= 1.0 {
				t.Error("expected growth penalty during deficit")
			}
			break
		}
	}
	if !deficitHit {
		t.Fatal("expected deficit with no veins and ongoing maintenance")
	}
	t.Logf("deficit occurred at tick %d", deficitTick)

	// Phase 2: Restore vein and verify recovery.
	veins := []fengshui.DragonVein{{ID: 1}}
	recovered := false

	for tick := deficitTick + 1; tick <= deficitTick+40; tick++ {
		result := e.Tick(tick, veins, roomChis, caveScore, rooms, beastCount, trapCount)
		if result.DeficitResult.Severity == None && result.Balance > 0 {
			recovered = true
			t.Logf("recovered at tick %d (balance=%.1f)", tick, result.Balance)
			break
		}
	}
	if !recovered {
		t.Error("expected recovery after restoring dragon vein supply")
	}
}
