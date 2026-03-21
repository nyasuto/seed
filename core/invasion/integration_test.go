package invasion

import (
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/testutil"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// TestIntegration_FullInvasionSimulation runs a 50-tick invasion simulation
// with 6 rooms, 3 beasts, and a wave of 3 invaders (2 warriors + 1 thief).
//
// Cave layout (40x40):
//
//	Room 1 (normal)       entry at (0,0)   — invasion entry point (蓄気室 role)
//	Room 2 (senju_room)   at (6,0)         — hub / patrol beast room
//	Room 3 (dragon_hole)  at (12,0)        — core, guard+chase beasts
//	Room 4 (trap_room)    at (6,6)         — traps on path to storage
//	Room 5 (storage)      at (12,6)        — treasure target
//	Room 6 (senju_room)   at (18,0)        — second beast room
//
// Connectivity: 1-2, 2-3, 2-4, 4-5, 2-6 (all rooms reachable)
//
// Invader wave (trigger tick 1):
//   - Warrior ×2 (DestroyCore) → head to room 3 via room 2
//   - Thief ×1 (StealTreasure, HP=300) → head to room 5 via rooms 2,4
//
// Beasts:
//   - Guard  (Fire, ATK=60, HP=200) in room 3 — high ATK kills warriors in ~2 rounds
//   - Chase  (Water, ATK=30, HP=80)  in room 3 — fights remaining warrior after first dies
//   - Patrol (Wood, ATK=20, HP=60)   in room 6 — stationed, not on invader path
//
// Expected flow:
//  1. Warriors reach room 3, fight guard+chase beasts → warrior2 killed, warrior1 low HP retreat
//  2. Thief enters trap room 4, takes repeated trap damage while slowed
//  3. After both warriors Defeated, thief sees ≥50% morale break → retreat
//  4. Wave completes with RewardChi from defeated warrior
//
// Verification items:
//   - 修行者が龍穴方向に移動すること
//   - 盗賊が倉庫方向に移動すること
//   - 罠部屋通過でダメージを受けること
//   - Guard仙獣が龍穴前で迎撃すること
//   - Chase仙獣が侵入者を追跡すること
//   - HP低下した侵入者が撤退すること
//   - 士気崩壊による撤退が発生すること
//   - 波完了時にRewardChiが集計されること
//   - 全過程がInvasionEventログとして記録されること
func TestIntegration_FullInvasionSimulation(t *testing.T) {
	// --- Build cave ---
	cave, err := world.NewCave(40, 40)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	r1, err := cave.AddRoom("normal", types.Pos{X: 0, Y: 0}, 3, 3, []world.RoomEntrance{
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
	r3, err := cave.AddRoom("dragon_hole", types.Pos{X: 12, Y: 0}, 3, 3, []world.RoomEntrance{
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
	r6, err := cave.AddRoom("senju_room", types.Pos{X: 18, Y: 0}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 18, Y: 1}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom r6: %v", err)
	}

	for _, c := range [][2]int{{r1.ID, r2.ID}, {r2.ID, r3.ID}, {r2.ID, r4.ID}, {r4.ID, r5.ID}, {r2.ID, r6.ID}} {
		if _, err := cave.ConnectRooms(c[0], c[1]); err != nil {
			t.Fatalf("ConnectRooms(%d,%d): %v", c[0], c[1], err)
		}
	}

	rooms := []*world.Room{r1, r2, r3, r4, r5, r6}
	graph := cave.BuildAdjacencyGraph()

	// --- Trap effects ---
	trapEffects := []TrapEffect{
		{RoomID: r4.ID, Element: types.Earth, DamagePerTrigger: 20, SlowTicks: 2},
	}

	// --- Invader class registry ---
	reg := newTestClassRegistry()

	// --- Invasion engine ---
	rng := testutil.NewTestRNG(42)
	engine := NewInvasionEngine(cave, graph, DefaultCombatParams(), rng, reg, trapEffects)

	// --- Create invaders ---
	warrior1 := makeTestInvader(1, "warrior", types.Wood, r1.ID, NewDestroyCoreGoal())
	warrior1.Memory.Visit(r1.ID, 0, cave, rooms)

	warrior2 := makeTestInvader(2, "warrior", types.Wood, r1.ID, NewDestroyCoreGoal())
	warrior2.Memory.Visit(r1.ID, 0, cave, rooms)

	// Thief with high HP to survive trap room and trigger morale break (not low HP retreat).
	thiefGoal := NewStealTreasureGoal()
	thief := makeTestInvader(3, "thief", types.Metal, r1.ID, thiefGoal)
	thief.HP = 300
	thief.MaxHP = 300
	thief.SPD = 40
	thief.Memory.Visit(r1.ID, 0, cave, rooms)

	wave := &InvasionWave{
		ID:          1,
		TriggerTick: 1,
		Invaders:    []*Invader{warrior1, warrior2, thief},
		State:       Pending,
		Difficulty:  1.0,
	}
	waves := []*InvasionWave{wave}

	// --- Create beasts ---
	// Guard beast in dragon_hole room: high ATK to kill warriors quickly.
	guardBeast := &senju.Beast{
		ID: 1, SpeciesID: "enhou", Name: "Guard-Enhou",
		Element: types.Fire, RoomID: r3.ID, Level: 3,
		HP: 200, MaxHP: 200, ATK: 60, DEF: 30, SPD: 20,
		State: senju.Idle,
	}
	// Chase beast also in dragon_hole room: moderate ATK, high SPD.
	chaseBeast := &senju.Beast{
		ID: 2, SpeciesID: "suija", Name: "Chase-Suija",
		Element: types.Water, RoomID: r3.ID, Level: 2,
		HP: 80, MaxHP: 80, ATK: 30, DEF: 20, SPD: 30,
		State: senju.Idle,
	}
	// Patrol beast in senju_room (not on invader path).
	patrolBeast := &senju.Beast{
		ID: 3, SpeciesID: "suiryu", Name: "Patrol-Suiryu",
		Element: types.Wood, RoomID: r6.ID, Level: 2,
		HP: 60, MaxHP: 60, ATK: 20, DEF: 15, SPD: 18,
		State: senju.Idle,
	}
	beasts := []*senju.Beast{guardBeast, chaseBeast, patrolBeast}

	// --- Room chi ---
	roomChi := map[int]*fengshui.RoomChi{
		r1.ID: {RoomID: r1.ID, Current: 50, Capacity: 200, Element: types.Wood},
		r2.ID: {RoomID: r2.ID, Current: 80, Capacity: 200, Element: types.Wood},
		r3.ID: {RoomID: r3.ID, Current: 100, Capacity: 300, Element: types.Fire},
		r4.ID: {RoomID: r4.ID, Current: 60, Capacity: 200, Element: types.Earth},
		r5.ID: {RoomID: r5.ID, Current: 200, Capacity: 300, Element: types.Metal},
		r6.ID: {RoomID: r6.ID, Current: 70, Capacity: 200, Element: types.Water},
	}

	// --- Run 50-tick simulation ---
	var allEvents []InvasionEvent
	for tick := types.Tick(1); tick <= 50; tick++ {
		events := engine.Tick(tick, waves, beasts, rooms, nil, roomChi)
		allEvents = append(allEvents, events...)
	}

	// --- Helpers ---
	eventsByType := make(map[InvasionEventType][]InvasionEvent)
	for _, e := range allEvents {
		eventsByType[e.Type] = append(eventsByType[e.Type], e)
	}

	hasEventType := func(t InvasionEventType) bool { return len(eventsByType[t]) > 0 }

	hasEventForInvader := func(et InvasionEventType, invaderID int) bool {
		for _, e := range eventsByType[et] {
			if e.InvaderID == invaderID {
				return true
			}
		}
		return false
	}

	hasEventWithDetail := func(et InvasionEventType, detail string) bool {
		for _, e := range eventsByType[et] {
			if e.Details == detail {
				return true
			}
		}
		return false
	}

	// --- Log event summary ---
	summary := make(map[InvasionEventType]int)
	for _, e := range allEvents {
		summary[e.Type]++
	}
	t.Logf("event summary: %v", summary)
	t.Logf("total events: %d", len(allEvents))
	for _, inv := range wave.Invaders {
		t.Logf("invader %d (%s): state=%v, HP=%d/%d, room=%d",
			inv.ID, inv.ClassID, inv.State, inv.HP, inv.MaxHP, inv.CurrentRoomID)
	}
	for _, b := range beasts {
		t.Logf("beast %d (%s): state=%v, HP=%d/%d, room=%d",
			b.ID, b.Name, b.State, b.HP, b.MaxHP, b.RoomID)
	}

	// ===== Verification =====

	// 1. Wave started
	if !hasEventType(WaveStarted) {
		t.Error("expected WaveStarted event")
	}

	// 2. 修行者が龍穴方向に移動すること
	// Warriors should move toward dragon_hole (room 2 then room 3).
	warriorMovedToHub := false
	warriorMovedToCore := false
	for _, e := range eventsByType[InvaderMoved] {
		if e.InvaderID == warrior1.ID || e.InvaderID == warrior2.ID {
			if e.RoomID == r2.ID {
				warriorMovedToHub = true
			}
			if e.RoomID == r3.ID {
				warriorMovedToCore = true
			}
		}
	}
	if !warriorMovedToHub {
		t.Error("expected warriors to move through hub room (room 2)")
	}
	if !warriorMovedToCore {
		t.Error("expected warriors to reach dragon_hole (room 3)")
	}

	// 3. 盗賊が倉庫方向に移動すること
	// Thief should move toward storage via hub (room 2) then trap room (room 4).
	thiefMovedToHub := false
	thiefMovedToTrap := false
	for _, e := range eventsByType[InvaderMoved] {
		if e.InvaderID == thief.ID {
			if e.RoomID == r2.ID {
				thiefMovedToHub = true
			}
			if e.RoomID == r4.ID {
				thiefMovedToTrap = true
			}
		}
	}
	if !thiefMovedToHub {
		t.Error("expected thief to move through hub room (room 2)")
	}
	if !thiefMovedToTrap {
		t.Error("expected thief to move through trap room (room 4) toward storage")
	}

	// 4. 罠部屋通過でダメージを受けること
	trapEvents := eventsByType[TrapTriggered]
	if len(trapEvents) == 0 {
		t.Error("expected TrapTriggered events in trap room")
	}
	for _, e := range trapEvents {
		if e.RoomID != r4.ID {
			t.Errorf("trap event in room %d, want room %d", e.RoomID, r4.ID)
		}
		if e.Damage <= 0 {
			t.Errorf("trap damage = %d, want > 0", e.Damage)
		}
	}

	// 5. Guard仙獣が龍穴前で迎撃すること
	guardFought := false
	for _, e := range eventsByType[CombatOccurred] {
		if e.BeastID == guardBeast.ID && e.RoomID == r3.ID {
			guardFought = true
			break
		}
	}
	if !guardFought {
		t.Error("expected guard beast to fight invaders in dragon_hole (room 3)")
	}

	// 6. Chase仙獣が侵入者を追跡すること (combat with chase beast)
	chaseFought := false
	for _, e := range eventsByType[CombatOccurred] {
		if e.BeastID == chaseBeast.ID && e.RoomID == r3.ID {
			chaseFought = true
			break
		}
	}
	if !chaseFought {
		t.Error("expected chase beast to fight invaders in room 3")
	}

	// 7. HP低下した侵入者が撤退すること
	if !hasEventWithDetail(InvaderRetreating, "LowHP") {
		t.Error("expected InvaderRetreating event with reason LowHP")
	}

	// 8. 士気崩壊による撤退が発生すること
	if !hasEventWithDetail(InvaderRetreating, "MoraleBroken") {
		t.Error("expected InvaderRetreating event with reason MoraleBroken")
	}

	// 9. 波完了時にRewardChiが集計されること
	// At least one warrior should have been defeated (HP→0) in combat.
	if !hasEventType(InvaderDefeated) {
		t.Error("expected InvaderDefeated event (warrior killed by guard beast)")
	}
	totalReward := engine.CollectRewards(allEvents)
	if totalReward <= 0 {
		t.Errorf("reward chi = %.1f, want > 0 (defeated invaders should yield chi)", totalReward)
	}
	t.Logf("reward chi: %.1f", totalReward)

	// 10. 全過程がInvasionEventログとして記録されること
	if len(allEvents) == 0 {
		t.Fatal("no events generated during simulation")
	}
	requiredEventTypes := []InvasionEventType{
		WaveStarted, InvaderMoved, TrapTriggered, CombatOccurred,
		InvaderRetreating, InvaderDefeated,
	}
	for _, et := range requiredEventTypes {
		if !hasEventType(et) {
			t.Errorf("expected event type %v in event log", et)
		}
	}
	for _, e := range allEvents {
		if e.Tick < 1 || e.Tick > 50 {
			t.Errorf("event %v has tick %d outside range [1,50]", e.Type, e.Tick)
		}
	}

	// 11. Wave should complete (all invaders Defeated after retreats/kills)
	waveResolved := hasEventType(WaveCompleted) || hasEventType(WaveFailed)
	if !waveResolved {
		activeCount := 0
		for _, inv := range wave.Invaders {
			if inv.State != Defeated {
				activeCount++
			}
		}
		t.Logf("wave not yet resolved after 50 ticks: %d invaders still active", activeCount)
	}

	// 12. Verify retreat events reference correct invaders
	// LowHP retreat should be a warrior (fought guard beast, lost HP)
	lowHPRetreatInvader := -1
	for _, e := range eventsByType[InvaderRetreating] {
		if e.Details == "LowHP" {
			lowHPRetreatInvader = e.InvaderID
			break
		}
	}
	if lowHPRetreatInvader != -1 {
		// Should be warrior1 or warrior2
		if lowHPRetreatInvader != warrior1.ID && lowHPRetreatInvader != warrior2.ID {
			t.Errorf("LowHP retreat invader ID=%d, expected warrior (%d or %d)",
				lowHPRetreatInvader, warrior1.ID, warrior2.ID)
		}
	}

	// MoraleBroken retreat should be the thief (after both warriors defeated)
	moraleRetreatInvader := -1
	for _, e := range eventsByType[InvaderRetreating] {
		if e.Details == "MoraleBroken" {
			moraleRetreatInvader = e.InvaderID
			break
		}
	}
	if moraleRetreatInvader != -1 && moraleRetreatInvader != thief.ID {
		t.Errorf("MoraleBroken retreat invader ID=%d, expected thief (%d)",
			moraleRetreatInvader, thief.ID)
	}

	// 13. Escaped invaders should have InvaderEscaped events
	escapedCount := len(eventsByType[InvaderEscaped])
	t.Logf("escaped invaders: %d", escapedCount)

	// 14. InvaderDefeated events should have positive RewardChi
	for _, e := range eventsByType[InvaderDefeated] {
		if e.RewardChi <= 0 {
			t.Errorf("InvaderDefeated (invader %d) has RewardChi=%.1f, want > 0",
				e.InvaderID, e.RewardChi)
		}
	}

	// 15. Check invader movement event includes the invader who is not fighting
	// (thief should have InvaderMoved events independent of warriors)
	if !hasEventForInvader(InvaderMoved, thief.ID) {
		t.Error("expected thief to have InvaderMoved events")
	}
}
