package economy_test

import (
	"math"
	"testing"

	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/testutil"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// TestIntegration_FullEconomySimulation runs an 80-tick simulation combining
// Cave, ChiFlowEngine, beasts, invasion waves, and the EconomyEngine.
//
// Cave layout (40×40):
//
//	Room 1 (dragon_hole)  at (0,0)   3×3 — core room, entrance to cave
//	Room 2 (senju_room)   at (6,0)   3×3 — beast room with 2 beasts
//	Room 3 (chi_chamber)  at (12,0)  3×3 — chi storage for pool cap
//	Room 4 (trap_room)    at (6,6)   3×3 — traps for invaders
//	Room 5 (storage)      at (12,6)  3×3 — treasure target
//
// Connectivity: 1-2, 2-3, 2-4, 4-5
//
// Verification items:
//  1. Chi supply deposited every tick into ChiPool
//  2. Maintenance costs withdrawn every tick
//  3. Invasion reward added to ChiPool
//  4. Thief escape loss subtracted from ChiPool
//  5. Deficit penalty applied when balance is low
//  6. Construction cost deducted from ChiPool with transaction
//  7. ChiPoolCap linked to chi_storage room level
//  8. Full transaction history preserved through serialization
func TestIntegration_FullEconomySimulation(t *testing.T) {
	// --- Build cave ---
	cave, err := world.NewCave(40, 40)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	r1, err := cave.AddRoom("dragon_hole", types.Pos{X: 0, Y: 0}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 2, Y: 1}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom r1: %v", err)
	}
	r2, err := cave.AddRoom("senju_room", types.Pos{X: 6, Y: 0}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 6, Y: 1}, Dir: types.West},
		{Pos: types.Pos{X: 8, Y: 1}, Dir: types.East},
		{Pos: types.Pos{X: 7, Y: 2}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom r2: %v", err)
	}
	r3, err := cave.AddRoom("chi_chamber", types.Pos{X: 12, Y: 0}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 12, Y: 1}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom r3: %v", err)
	}
	r4, err := cave.AddRoom("trap_room", types.Pos{X: 6, Y: 6}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 7, Y: 6}, Dir: types.North},
		{Pos: types.Pos{X: 8, Y: 7}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom r4: %v", err)
	}
	r5, err := cave.AddRoom("storage", types.Pos{X: 12, Y: 6}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 12, Y: 7}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom r5: %v", err)
	}

	for _, c := range [][2]int{{r1.ID, r2.ID}, {r2.ID, r3.ID}, {r2.ID, r4.ID}, {r4.ID, r5.ID}} {
		if _, err := cave.ConnectRooms(c[0], c[1]); err != nil {
			t.Fatalf("ConnectRooms(%d,%d): %v", c[0], c[1], err)
		}
	}

	// --- Room type registry for chi flow ---
	roomReg, err := world.LoadDefaultRoomTypes()
	if err != nil {
		t.Fatalf("LoadDefaultRoomTypes: %v", err)
	}

	// --- Dragon vein (from entrance at r1) ---
	vein, err := fengshui.BuildDragonVein(cave, types.Pos{X: 2, Y: 1}, types.Wood, 8.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	// --- Chi flow engine ---
	flowParams := fengshui.DefaultFlowParams()
	chiEngine := fengshui.NewChiFlowEngine(cave, []*fengshui.DragonVein{vein}, roomReg, flowParams)

	// --- Beasts (2 in senju_room) ---
	guardBeast := &senju.Beast{
		ID: 1, SpeciesID: "enhou", Name: "Guard-Enhou",
		Element: types.Fire, RoomID: r2.ID, Level: 3,
		HP: 250, MaxHP: 250, ATK: 80, DEF: 30, SPD: 25,
		State: senju.Idle,
	}
	patrolBeast := &senju.Beast{
		ID: 2, SpeciesID: "suiryu", Name: "Patrol-Suiryu",
		Element: types.Wood, RoomID: r2.ID, Level: 2,
		HP: 100, MaxHP: 100, ATK: 40, DEF: 20, SPD: 20,
		State: senju.Idle,
	}
	beasts := []*senju.Beast{guardBeast, patrolBeast}
	beastCount := len(beasts)

	// --- Invasion setup ---
	invReg := invasion.NewInvaderClassRegistry()
	_ = invReg.Register(invasion.InvaderClass{
		ID: "warrior", Name: "Warrior", Element: types.Wood,
		BaseHP: 100, BaseATK: 25, BaseDEF: 20, BaseSPD: 20,
		RewardChi: 15, PreferredGoal: invasion.DestroyCore, RetreatThreshold: 0.3,
	})
	_ = invReg.Register(invasion.InvaderClass{
		ID: "thief", Name: "Thief", Element: types.Metal,
		BaseHP: 70, BaseATK: 20, BaseDEF: 15, BaseSPD: 40,
		RewardChi: 12, PreferredGoal: invasion.StealTreasure, RetreatThreshold: 0.5,
	})

	rng := testutil.NewTestRNG(42)
	graph := cave.BuildAdjacencyGraph()
	trapEffects := []invasion.TrapEffect{
		{RoomID: r4.ID, Element: types.Earth, DamagePerTrigger: 20, SlowTicks: 2},
	}
	invEngine := invasion.NewInvasionEngine(cave, graph, invasion.DefaultCombatParams(), rng, invReg, trapEffects)

	// Warrior invader — enters at r1 heading for dragon_hole
	warrior := &invasion.Invader{
		ID: 1, ClassID: "warrior", Name: "Warrior",
		Element: types.Wood, Level: 1,
		HP: 100, MaxHP: 100, ATK: 25, DEF: 20, SPD: 20,
		CurrentRoomID: r1.ID,
		Goal:          invasion.NewDestroyCoreGoal(),
		Memory:        invasion.NewExplorationMemory(),
		State:         invasion.Advancing,
		EntryTick:     0,
	}
	rooms := []*world.Room{r1, r2, r3, r4, r5}
	warrior.Memory.Visit(r1.ID, 0, cave, rooms)

	// Thief — enters at r1 heading for storage, high HP to survive traps
	thief := &invasion.Invader{
		ID: 2, ClassID: "thief", Name: "Thief",
		Element: types.Metal, Level: 1,
		HP: 200, MaxHP: 200, ATK: 20, DEF: 15, SPD: 40,
		CurrentRoomID: r1.ID,
		Goal:          invasion.NewStealTreasureGoal(),
		Memory:        invasion.NewExplorationMemory(),
		State:         invasion.Advancing,
		EntryTick:     0,
	}
	thief.Memory.Visit(r1.ID, 0, cave, rooms)

	wave := &invasion.InvasionWave{
		ID:          1,
		TriggerTick: 10,
		Invaders:    []*invasion.Invader{warrior, thief},
		State:       invasion.Pending,
		Difficulty:  1.0,
	}
	waves := []*invasion.InvasionWave{wave}

	// --- Economy engine ---
	// Use rooms with economy-specific TypeIDs for maintenance calculation.
	// chi_storage maps to chi_chamber in world, but economy uses its own IDs.
	econRooms := []world.Room{
		{ID: r1.ID, TypeID: "dragon_hole", Level: 1},
		{ID: r2.ID, TypeID: "beast_room", Level: 1},
		{ID: r3.ID, TypeID: "chi_storage", Level: 1}, // chi_chamber → chi_storage for economy
		{ID: r4.ID, TypeID: "trap_room", Level: 1},
		{ID: r5.ID, TypeID: "warehouse", Level: 1},
	}
	trapCount := 1

	supplyParams, err := economy.DefaultSupplyParams()
	if err != nil {
		t.Fatalf("DefaultSupplyParams: %v", err)
	}
	costParams, err := economy.DefaultCostParams()
	if err != nil {
		t.Fatalf("DefaultCostParams: %v", err)
	}
	deficitParams, err := economy.DefaultDeficitParams()
	if err != nil {
		t.Fatalf("DefaultDeficitParams: %v", err)
	}
	constructionCost, err := economy.DefaultConstructionCost()
	if err != nil {
		t.Fatalf("DefaultConstructionCost: %v", err)
	}
	beastCostTable, err := economy.DefaultBeastCost()
	if err != nil {
		t.Fatalf("DefaultBeastCost: %v", err)
	}

	chiPool := economy.NewChiPool(200.0) // initial cap 200
	econEngine := economy.NewEconomyEngine(chiPool, supplyParams, costParams, deficitParams, constructionCost, beastCostTable)

	// Give initial chi to start with
	_ = chiPool.Deposit(50.0, economy.Supply, "initial deposit", 0)

	// --- Run 80-tick simulation ---
	var econResults []economy.EconomyTickResult
	var allInvasionEvents []invasion.InvasionEvent

	supplyReceived := false
	maintenanceCharged := false
	rewardDeposited := false
	theftLossRecorded := false
	deficitOccurred := false

	for tick := types.Tick(1); tick <= 80; tick++ {
		// 1. Chi flow tick
		chiEngine.Tick()

		// 2. Invasion tick
		invEvents := invEngine.Tick(tick, waves, beasts, rooms, nil, chiEngine.RoomChi)
		allInvasionEvents = append(allInvasionEvents, invEvents...)

		// 3. Process invasion rewards/losses into economy
		for _, ev := range invEvents {
			if ev.Type == invasion.InvaderDefeated && ev.RewardChi > 0 {
				_ = chiPool.Deposit(ev.RewardChi, economy.Reward, "invader defeated", tick)
				rewardDeposited = true
			}
			if ev.Type == invasion.InvaderEscaped && ev.StolenChi > 0 {
				_ = chiPool.Withdraw(ev.StolenChi, economy.Theft, "thief escaped", tick)
				theftLossRecorded = true
			}
		}

		// 4. Economy tick — supply, maintenance, deficit
		// Build vein list as value slice for economy engine
		veins := []fengshui.DragonVein{*vein}
		caveScore := 0.5 // moderate feng shui score
		result := econEngine.Tick(tick, veins, chiEngine.RoomChi, caveScore, econRooms, beastCount, trapCount)
		econResults = append(econResults, result)

		if result.Supply > 0 {
			supplyReceived = true
		}
		if result.Maintenance.Total > 0 {
			maintenanceCharged = true
		}
		if result.DeficitResult.Severity != economy.None {
			deficitOccurred = true
		}
	}

	// ===== Verification =====

	// 1. Chi supply deposited every tick into ChiPool
	if !supplyReceived {
		t.Error("expected chi supply to be deposited into ChiPool at least once")
	}
	// Verify supply is positive on every tick (we have 1 vein)
	for i, r := range econResults {
		if r.Supply <= 0 {
			t.Errorf("tick %d: supply=%.2f, want > 0", i+1, r.Supply)
			break
		}
	}

	// 2. Maintenance costs withdrawn every tick
	if !maintenanceCharged {
		t.Error("expected maintenance costs to be charged at least once")
	}
	// Verify maintenance breakdown includes room + beast + trap costs
	firstMaint := econResults[0].Maintenance
	if firstMaint.RoomCost <= 0 {
		t.Errorf("expected room maintenance > 0, got %.2f", firstMaint.RoomCost)
	}
	if firstMaint.BeastCost <= 0 {
		t.Errorf("expected beast maintenance > 0, got %.2f", firstMaint.BeastCost)
	}
	if firstMaint.TrapCost <= 0 {
		t.Errorf("expected trap maintenance > 0, got %.2f", firstMaint.TrapCost)
	}
	if firstMaint.Total <= 0 {
		t.Errorf("expected total maintenance > 0, got %.2f", firstMaint.Total)
	}

	// 3. Invasion reward deposited into ChiPool
	// The warrior should be defeated by our guard beast.
	if !rewardDeposited {
		t.Log("warrior state:", warrior.State, "HP:", warrior.HP)
		t.Error("expected invasion reward to be deposited into ChiPool from defeated invader")
	}

	// Verify reward transactions exist in history
	rewardTxCount := 0
	for _, tx := range chiPool.History {
		if tx.Type == economy.Reward {
			rewardTxCount++
		}
	}
	if rewardDeposited && rewardTxCount == 0 {
		t.Error("reward deposited but no Reward transaction found in history")
	}

	// 4. Thief escape loss subtracted from ChiPool
	// The thief may or may not escape depending on simulation dynamics.
	// We log but don't require it since the thief might be killed too.
	if theftLossRecorded {
		theftTxCount := 0
		for _, tx := range chiPool.History {
			if tx.Type == economy.Theft {
				theftTxCount++
			}
		}
		if theftTxCount == 0 {
			t.Error("theft loss recorded but no Theft transaction found in history")
		}
		t.Logf("theft loss recorded: %d transactions", theftTxCount)
	} else {
		t.Log("no thief escape occurred (thief may have been defeated or is still active)")
	}

	// 5. Deficit penalty when balance runs low
	// Force a deficit scenario: drain balance and run another tick.
	balanceBefore := chiPool.Balance()
	// Withdraw all remaining balance to trigger deficit
	if balanceBefore > 0 {
		_ = chiPool.Withdraw(balanceBefore, economy.Construction, "drain for deficit test", 81)
	}
	veins := []fengshui.DragonVein{*vein}
	// Set supply very low by using 0 caveScore
	deficitResult := econEngine.Tick(81, veins, chiEngine.RoomChi, 0.0, econRooms, beastCount, trapCount)
	if deficitResult.DeficitResult.Severity == economy.None {
		// Supply might cover maintenance even with caveScore 0.
		// Drain again after supply deposit.
		if chiPool.Balance() > 0 {
			_ = chiPool.Withdraw(chiPool.Balance(), economy.Construction, "drain again", 82)
		}
		deficitResult = econEngine.Tick(82, nil, nil, 0.0, econRooms, beastCount, trapCount)
	}
	if deficitResult.DeficitResult.Severity == economy.None {
		t.Error("expected deficit to occur when ChiPool is empty and maintenance is due")
	} else {
		t.Logf("deficit severity: %v, shortage: %.2f", deficitResult.DeficitResult.Severity, deficitResult.DeficitResult.Shortage)
		// Verify deficit penalties are set
		if deficitResult.DeficitResult.Shortage <= 0 {
			t.Errorf("expected positive shortage in deficit, got %.2f", deficitResult.DeficitResult.Shortage)
		}
	}

	// Also check that deficit occurred naturally during the 80-tick run if supply < maintenance
	t.Logf("natural deficit during simulation: %v", deficitOccurred)

	// 6. Construction cost deducted from ChiPool with transaction
	// Refill pool for construction test
	_ = chiPool.Deposit(100.0, economy.Supply, "refill for construction", 83)
	balanceBefore = chiPool.Balance()
	cost, err := econEngine.TryBuildRoom("chi_storage", 83)
	if err != nil {
		t.Fatalf("TryBuildRoom: %v", err)
	}
	if cost <= 0 {
		t.Errorf("expected positive construction cost, got %.2f", cost)
	}
	balanceAfter := chiPool.Balance()
	expectedBalance := balanceBefore - cost
	if math.Abs(balanceAfter-expectedBalance) > 0.001 {
		t.Errorf("balance after construction: got %.2f, want %.2f", balanceAfter, expectedBalance)
	}
	// Verify construction transaction recorded
	lastTx := chiPool.History[len(chiPool.History)-1]
	if lastTx.Type != economy.Construction {
		t.Errorf("last transaction type = %v, want Construction", lastTx.Type)
	}

	// 7. ChiPoolCap linked to chi_storage room level
	// Current rooms include one chi_storage level 1 → base + 1 × storagePerRoom
	expectedCap := costParams.ChiPoolBaseCap + costParams.ChiPoolCapPerStorageRoom
	actualCap := economy.CalcChiPoolCap(econRooms, costParams)
	if math.Abs(actualCap-expectedCap) > 0.001 {
		t.Errorf("ChiPoolCap with 1 storage room: got %.2f, want %.2f", actualCap, expectedCap)
	}

	// Upgrade storage room level and verify cap increases
	econRoomsUpgraded := make([]world.Room, len(econRooms))
	copy(econRoomsUpgraded, econRooms)
	for i := range econRoomsUpgraded {
		if econRoomsUpgraded[i].TypeID == "chi_storage" {
			econRoomsUpgraded[i].Level = 3
		}
	}
	expectedCapUpgraded := costParams.ChiPoolBaseCap + costParams.ChiPoolCapPerStorageRoom +
		2*costParams.ChiPoolCapPerStorageLevel
	actualCapUpgraded := economy.CalcChiPoolCap(econRoomsUpgraded, costParams)
	if math.Abs(actualCapUpgraded-expectedCapUpgraded) > 0.001 {
		t.Errorf("ChiPoolCap with storage lv3: got %.2f, want %.2f", actualCapUpgraded, expectedCapUpgraded)
	}
	if actualCapUpgraded <= actualCap {
		t.Errorf("upgraded cap (%.2f) should be > base cap (%.2f)", actualCapUpgraded, actualCap)
	}

	// 8. Full transaction history preserved through serialization
	historyLen := len(chiPool.History)
	if historyLen == 0 {
		t.Fatal("expected non-empty transaction history")
	}
	t.Logf("transaction history length: %d", historyLen)

	data, err := economy.MarshalEconomyState(econEngine)
	if err != nil {
		t.Fatalf("MarshalEconomyState: %v", err)
	}

	restored, err := economy.UnmarshalEconomyState(data, supplyParams, costParams, deficitParams, constructionCost, beastCostTable)
	if err != nil {
		t.Fatalf("UnmarshalEconomyState: %v", err)
	}

	// Verify restored state matches
	if math.Abs(restored.ChiPool.Balance()-chiPool.Balance()) > 0.001 {
		t.Errorf("restored balance = %.2f, want %.2f", restored.ChiPool.Balance(), chiPool.Balance())
	}
	if math.Abs(restored.ChiPool.Cap-chiPool.Cap) > 0.001 {
		t.Errorf("restored cap = %.2f, want %.2f", restored.ChiPool.Cap, chiPool.Cap)
	}
	if len(restored.ChiPool.History) != historyLen {
		t.Errorf("restored history length = %d, want %d", len(restored.ChiPool.History), historyLen)
	}

	// Verify individual transaction records are preserved
	for i, tx := range chiPool.History {
		rtx := restored.ChiPool.History[i]
		if tx.Tick != rtx.Tick {
			t.Errorf("tx[%d] tick: got %d, want %d", i, rtx.Tick, tx.Tick)
		}
		if math.Abs(tx.Amount-rtx.Amount) > 0.001 {
			t.Errorf("tx[%d] amount: got %.4f, want %.4f", i, rtx.Amount, tx.Amount)
		}
		if tx.Type != rtx.Type {
			t.Errorf("tx[%d] type: got %v, want %v", i, rtx.Type, tx.Type)
		}
		if tx.Reason != rtx.Reason {
			t.Errorf("tx[%d] reason: got %q, want %q", i, rtx.Reason, tx.Reason)
		}
	}

	// --- Summary ---
	t.Logf("final balance: %.2f / cap: %.2f", chiPool.Balance(), chiPool.Cap)
	t.Logf("invasion events: %d", len(allInvasionEvents))

	// Count transaction types for summary
	txTypeCounts := make(map[economy.TransactionType]int)
	for _, tx := range chiPool.History {
		txTypeCounts[tx.Type]++
	}
	t.Logf("transaction type breakdown: %v", txTypeCounts)
}
