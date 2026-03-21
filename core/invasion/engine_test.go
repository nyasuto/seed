package invasion

import (
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/testutil"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// setupEngineTest creates a cave with connected rooms and returns the engine and fixtures.
// Layout: entry(1) -> normal(2) -> dragon_hole(3), with a branch: normal(2) -> trap_room(4) -> storage(5)
func setupEngineTest(t *testing.T) (*InvasionEngine, *world.Cave, []*world.Room, world.AdjacencyGraph) {
	t.Helper()
	cave, err := world.NewCave(40, 40)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1 (entry) at (0,0) 3x3 — entrance on east side
	r1, err := cave.AddRoom("normal", types.Pos{X: 0, Y: 0}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 2, Y: 1}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom entry: %v", err)
	}
	// Room 2 (normal) at (6,0) 3x3 — entrances on west, east, south
	r2, err := cave.AddRoom("normal", types.Pos{X: 6, Y: 0}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 6, Y: 1}, Dir: types.West},
		{Pos: types.Pos{X: 8, Y: 1}, Dir: types.East},
		{Pos: types.Pos{X: 7, Y: 2}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom normal: %v", err)
	}
	// Room 3 (dragon_hole) at (12,0) 3x3 — entrance on west side
	r3, err := cave.AddRoom("dragon_hole", types.Pos{X: 12, Y: 0}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 12, Y: 1}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom dragon_hole: %v", err)
	}
	// Room 4 (trap_room) at (6,6) 3x3 — entrances on north and east
	r4, err := cave.AddRoom("trap_room", types.Pos{X: 6, Y: 6}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 7, Y: 6}, Dir: types.North},
		{Pos: types.Pos{X: 8, Y: 7}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom trap_room: %v", err)
	}
	// Room 5 (storage) at (12,6) 3x3 — entrance on west side
	r5, err := cave.AddRoom("storage", types.Pos{X: 12, Y: 6}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 12, Y: 7}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom storage: %v", err)
	}

	// Connect: 1-2, 2-3, 2-4, 4-5
	if _, err := cave.ConnectRooms(r1.ID, r2.ID); err != nil {
		t.Fatalf("ConnectRooms 1-2: %v", err)
	}
	if _, err := cave.ConnectRooms(r2.ID, r3.ID); err != nil {
		t.Fatalf("ConnectRooms 2-3: %v", err)
	}
	if _, err := cave.ConnectRooms(r2.ID, r4.ID); err != nil {
		t.Fatalf("ConnectRooms 2-4: %v", err)
	}
	if _, err := cave.ConnectRooms(r4.ID, r5.ID); err != nil {
		t.Fatalf("ConnectRooms 4-5: %v", err)
	}

	graph := cave.BuildAdjacencyGraph()
	rooms := []*world.Room{r1, r2, r3, r4, r5}

	trapEffects := []TrapEffect{
		{RoomID: r4.ID, Element: types.Earth, DamagePerTrigger: 20, SlowTicks: 2},
	}

	rng := testutil.NewTestRNG(42)
	reg := newTestClassRegistry()
	engine := NewInvasionEngine(cave, graph, DefaultCombatParams(), rng, reg, trapEffects)

	return engine, cave, rooms, graph
}

func makeTestWave(id int, triggerTick types.Tick, invaders []*Invader) *InvasionWave {
	return &InvasionWave{
		ID:          id,
		TriggerTick: triggerTick,
		Invaders:    invaders,
		State:       Pending,
		Difficulty:  1.0,
	}
}

func makeTestInvader(id int, classID string, element types.Element, roomID int, goal Goal) *Invader {
	return &Invader{
		ID:            id,
		ClassID:       classID,
		Name:          classID,
		Element:       element,
		Level:         1,
		HP:            100,
		MaxHP:         100,
		ATK:           25,
		DEF:           20,
		SPD:           20,
		CurrentRoomID: roomID,
		Goal:          goal,
		Memory:        NewExplorationMemory(),
		State:         Advancing,
		EntryTick:     0,
	}
}

// TestEngine_WaveActivation tests that pending waves are activated at the correct tick.
func TestEngine_WaveActivation(t *testing.T) {
	engine, _, rooms, _ := setupEngineTest(t)
	inv := makeTestInvader(1, "warrior", types.Wood, 1, NewDestroyCoreGoal())
	wave := makeTestWave(1, 10, []*Invader{inv})
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	// Tick 5: wave should NOT activate.
	events := engine.Tick(5, waves, nil, rooms, nil, roomChi)
	if wave.State != Pending {
		t.Errorf("wave state at tick 5: got %v, want Pending", wave.State)
	}
	for _, e := range events {
		if e.Type == WaveStarted {
			t.Error("unexpected WaveStarted event at tick 5")
		}
	}

	// Tick 10: wave should activate.
	events = engine.Tick(10, waves, nil, rooms, nil, roomChi)
	if wave.State != Active {
		t.Errorf("wave state at tick 10: got %v, want Active", wave.State)
	}
	found := false
	for _, e := range events {
		if e.Type == WaveStarted && e.WaveID == 1 {
			found = true
		}
	}
	if !found {
		t.Error("expected WaveStarted event at tick 10")
	}
}

// TestEngine_GoalOrientedMovement tests that invaders move toward dragon_hole.
func TestEngine_GoalOrientedMovement(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	inv := makeTestInvader(1, "warrior", types.Wood, 1, NewDestroyCoreGoal())
	// Record entry room visit.
	inv.Memory.Visit(1, 0, cave, rooms)
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	// Tick: invader should move from room 1 toward room 2 (path to dragon_hole).
	events := engine.Tick(1, waves, nil, rooms, nil, roomChi)

	moved := false
	for _, e := range events {
		if e.Type == InvaderMoved && e.InvaderID == 1 {
			moved = true
			if e.RoomID != 2 {
				t.Errorf("invader moved to room %d, want 2", e.RoomID)
			}
		}
	}
	if !moved {
		t.Error("expected InvaderMoved event")
	}
}

// TestEngine_ExplorationMovement tests that invaders explore unvisited rooms.
func TestEngine_ExplorationMovement(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	// Invader with HuntBeasts goal (no beasts in cave, so it explores).
	inv := makeTestInvader(1, "hunter", types.Fire, 1, NewHuntBeastsGoal())
	inv.Memory.Visit(1, 0, cave, rooms)
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	events := engine.Tick(1, waves, nil, rooms, nil, roomChi)
	moved := false
	for _, e := range events {
		if e.Type == InvaderMoved && e.InvaderID == 1 {
			moved = true
		}
	}
	if !moved {
		t.Error("expected InvaderMoved event during exploration")
	}
}

// TestEngine_CombatMatching tests that beasts and invaders in the same room fight.
func TestEngine_CombatMatching(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	inv := makeTestInvader(1, "warrior", types.Wood, 2, NewDestroyCoreGoal())
	inv.Memory.Visit(1, 0, cave, rooms)
	inv.Memory.Visit(2, 1, cave, rooms)
	inv.SlowTicks = 1 // Prevent movement so combat occurs in room 2.
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}

	beast := &senju.Beast{
		ID:        1,
		SpeciesID: "tiger",
		Name:      "Tiger",
		Element:   types.Fire,
		RoomID:    2,
		Level:     1,
		HP:        80,
		MaxHP:     80,
		ATK:       30,
		DEF:       15,
		SPD:       25,
		State:     senju.Idle,
	}
	beasts := []*senju.Beast{beast}

	roomChi := map[int]*fengshui.RoomChi{
		2: {RoomID: 2, Current: 50, Capacity: 100, Element: types.Fire},
	}

	events := engine.Tick(2, waves, beasts, rooms, nil, roomChi)
	combatFound := false
	for _, e := range events {
		if e.Type == CombatOccurred && e.RoomID == 2 {
			combatFound = true
		}
	}
	if !combatFound {
		t.Error("expected CombatOccurred event in room 2")
	}
}

// TestEngine_TrapEffect tests that traps trigger on advancing invaders.
func TestEngine_TrapEffect(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	// Place invader directly in trap room.
	inv := makeTestInvader(1, "warrior", types.Wood, 4, NewDestroyCoreGoal())
	inv.Memory.Visit(1, 0, cave, rooms)
	inv.Memory.Visit(2, 1, cave, rooms)
	inv.Memory.Visit(4, 2, cave, rooms)
	inv.SlowTicks = 1 // Prevent movement to stay in trap room.
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	hpBefore := inv.HP
	events := engine.Tick(3, waves, nil, rooms, nil, roomChi)

	trapFound := false
	for _, e := range events {
		if e.Type == TrapTriggered && e.InvaderID == 1 && e.RoomID == 4 {
			trapFound = true
			if e.Damage <= 0 {
				t.Errorf("trap damage = %d, want > 0", e.Damage)
			}
		}
	}
	if !trapFound {
		t.Error("expected TrapTriggered event in room 4")
	}
	if inv.HP >= hpBefore {
		t.Error("invader HP should be reduced after trap")
	}
}

// TestEngine_RetreatOnLowHP tests that invaders retreat when HP is low.
func TestEngine_RetreatOnLowHP(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	inv := makeTestInvader(1, "warrior", types.Wood, 2, NewDestroyCoreGoal())
	inv.HP = 20 // Below retreat threshold (100 * 0.3 = 30).
	inv.Memory.Visit(1, 0, cave, rooms)
	inv.Memory.Visit(2, 1, cave, rooms)
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	events := engine.Tick(2, waves, nil, rooms, nil, roomChi)

	retreatFound := false
	for _, e := range events {
		if e.Type == InvaderRetreating && e.InvaderID == 1 {
			retreatFound = true
			if e.Details != "LowHP" {
				t.Errorf("retreat reason = %q, want LowHP", e.Details)
			}
		}
	}
	if !retreatFound {
		t.Error("expected InvaderRetreating event for low HP")
	}
	if inv.State != Retreating {
		t.Errorf("invader state = %v, want Retreating", inv.State)
	}
}

// TestEngine_MoraleBreakRetreat tests that invaders retreat when >= 50% companions are defeated.
func TestEngine_MoraleBreakRetreat(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	inv1 := makeTestInvader(1, "warrior", types.Wood, 2, NewDestroyCoreGoal())
	inv1.State = Defeated
	inv2 := makeTestInvader(2, "warrior", types.Wood, 2, NewDestroyCoreGoal())
	inv2.Memory.Visit(1, 0, cave, rooms)
	inv2.Memory.Visit(2, 1, cave, rooms)
	inv2.SlowTicks = 1 // prevent movement

	wave := makeTestWave(1, 0, []*Invader{inv1, inv2})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	events := engine.Tick(2, waves, nil, rooms, nil, roomChi)

	retreatFound := false
	for _, e := range events {
		if e.Type == InvaderRetreating && e.InvaderID == 2 {
			retreatFound = true
			if e.Details != "MoraleBroken" {
				t.Errorf("retreat reason = %q, want MoraleBroken", e.Details)
			}
		}
	}
	if !retreatFound {
		t.Error("expected InvaderRetreating due to morale break")
	}
}

// TestEngine_GoalAchievedRetreat tests that invaders with a steal goal retreat after achieving.
func TestEngine_GoalAchievedRetreat(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	// Invader in storage room (5) with StealTreasure goal — already achieved.
	inv := makeTestInvader(1, "thief", types.Metal, 5, NewStealTreasureGoal())
	inv.Memory.Visit(1, 0, cave, rooms)
	inv.Memory.Visit(2, 1, cave, rooms)
	inv.Memory.Visit(4, 2, cave, rooms)
	inv.Memory.Visit(5, 3, cave, rooms)
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := map[int]*fengshui.RoomChi{
		5: {RoomID: 5, Current: 100, Capacity: 200, Element: types.Metal},
	}

	events := engine.Tick(4, waves, nil, rooms, nil, roomChi)

	goalFound := false
	retreatFound := false
	for _, e := range events {
		if e.Type == GoalAchievedEvent && e.InvaderID == 1 {
			goalFound = true
		}
		if e.Type == InvaderRetreating && e.InvaderID == 1 {
			retreatFound = true
			if e.StolenChi <= 0 {
				t.Errorf("stolen chi = %f, want > 0", e.StolenChi)
			}
		}
	}
	if !goalFound {
		t.Error("expected GoalAchievedEvent")
	}
	if !retreatFound {
		t.Error("expected InvaderRetreating after goal achieved")
	}
}

// TestEngine_WaveCompletedOnAllDefeated tests that wave is Completed when all invaders defeated.
func TestEngine_WaveCompletedOnAllDefeated(t *testing.T) {
	engine, _, rooms, _ := setupEngineTest(t)

	inv := makeTestInvader(1, "warrior", types.Wood, 1, NewDestroyCoreGoal())
	inv.HP = 0
	inv.State = Defeated
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	events := engine.Tick(1, waves, nil, rooms, nil, roomChi)

	completed := false
	for _, e := range events {
		if e.Type == WaveCompleted && e.WaveID == 1 {
			completed = true
		}
	}
	if !completed {
		t.Error("expected WaveCompleted event when all invaders defeated")
	}
	if wave.State != Completed {
		t.Errorf("wave state = %v, want Completed", wave.State)
	}
}

// TestEngine_InvaderDefeatedEvent tests that defeated invaders emit an event.
func TestEngine_InvaderDefeatedEvent(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	inv := makeTestInvader(1, "warrior", types.Wood, 2, NewDestroyCoreGoal())
	inv.HP = 0 // Will be caught by the defeated check.
	inv.Memory.Visit(1, 0, cave, rooms)
	inv.Memory.Visit(2, 1, cave, rooms)
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	events := engine.Tick(2, waves, nil, rooms, nil, roomChi)

	defeated := false
	for _, e := range events {
		if e.Type == InvaderDefeated && e.InvaderID == 1 {
			defeated = true
			if e.RewardChi <= 0 {
				t.Errorf("reward chi = %f, want > 0", e.RewardChi)
			}
		}
	}
	if !defeated {
		t.Error("expected InvaderDefeated event")
	}
}

// TestEngine_RewardChiCollection tests CollectRewards aggregation.
func TestEngine_RewardChiCollection(t *testing.T) {
	engine, _, _, _ := setupEngineTest(t)

	events := []InvasionEvent{
		{Type: InvaderDefeated, RewardChi: 15.0},
		{Type: InvaderDefeated, RewardChi: 12.0},
		{Type: InvaderMoved},
	}

	total := engine.CollectRewards(events)
	if total != 27.0 {
		t.Errorf("CollectRewards = %f, want 27.0", total)
	}
}

// TestEngine_StolenChiCollection tests CollectStolenChi aggregation.
func TestEngine_StolenChiCollection(t *testing.T) {
	engine, _, _, _ := setupEngineTest(t)

	events := []InvasionEvent{
		{Type: InvaderEscaped, StolenChi: 50.0},
		{Type: InvaderEscaped, StolenChi: 30.0},
		{Type: InvaderMoved},
	}

	total := engine.CollectStolenChi(events)
	if total != 80.0 {
		t.Errorf("CollectStolenChi = %f, want 80.0", total)
	}
}

// TestEngine_SlowTickSkipsMovement tests that slowed invaders don't move.
func TestEngine_SlowTickSkipsMovement(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	inv := makeTestInvader(1, "warrior", types.Wood, 2, NewDestroyCoreGoal())
	inv.Memory.Visit(1, 0, cave, rooms)
	inv.Memory.Visit(2, 1, cave, rooms)
	inv.SlowTicks = 3
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	events := engine.Tick(2, waves, nil, rooms, nil, roomChi)

	for _, e := range events {
		if e.Type == InvaderMoved && e.InvaderID == 1 {
			t.Error("slowed invader should not move")
		}
	}
	if inv.SlowTicks != 2 {
		t.Errorf("SlowTicks = %d, want 2", inv.SlowTicks)
	}
	if inv.CurrentRoomID != 2 {
		t.Errorf("invader room = %d, want 2 (no movement)", inv.CurrentRoomID)
	}
}

// TestEngine_BeastDefeatedEvent tests that beasts with 0 HP emit BeastDefeated.
func TestEngine_BeastDefeatedEvent(t *testing.T) {
	engine, _, rooms, _ := setupEngineTest(t)

	beast := &senju.Beast{
		ID:        1,
		SpeciesID: "tiger",
		Name:      "Tiger",
		Element:   types.Fire,
		RoomID:    2,
		Level:     1,
		HP:        0,
		MaxHP:     80,
		ATK:       30,
		DEF:       15,
		SPD:       25,
		State:     senju.Fighting,
	}
	beasts := []*senju.Beast{beast}

	// No active waves, just check beast defeat detection.
	wave := makeTestWave(1, 100, nil) // far in the future, won't activate
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	events := engine.Tick(1, waves, beasts, rooms, nil, roomChi)

	beastDefeated := false
	for _, e := range events {
		if e.Type == BeastDefeated && e.BeastID == 1 {
			beastDefeated = true
		}
	}
	if !beastDefeated {
		t.Error("expected BeastDefeated event")
	}
	if beast.State != senju.Recovering {
		t.Errorf("beast state = %v, want Recovering", beast.State)
	}
}

// TestEngine_BuildInvaderPositions tests the position map builder.
func TestEngine_BuildInvaderPositions(t *testing.T) {
	engine, _, _, _ := setupEngineTest(t)

	inv1 := makeTestInvader(1, "warrior", types.Wood, 2, NewDestroyCoreGoal())
	inv2 := makeTestInvader(2, "hunter", types.Fire, 2, NewHuntBeastsGoal())
	inv3 := makeTestInvader(3, "thief", types.Metal, 4, NewStealTreasureGoal())
	inv3.State = Retreating // Should not be included.

	wave := &InvasionWave{ID: 1, State: Active, Invaders: []*Invader{inv1, inv2, inv3}}
	waves := []*InvasionWave{wave}

	positions := engine.BuildInvaderPositions(waves)

	if len(positions[2]) != 2 {
		t.Errorf("invaders in room 2 = %d, want 2", len(positions[2]))
	}
	if len(positions[4]) != 0 {
		t.Errorf("invaders in room 4 = %d, want 0 (retreating excluded)", len(positions[4]))
	}
}

// TestEngine_InvasionEventTypeString tests the String method of InvasionEventType.
func TestEngine_InvasionEventTypeString(t *testing.T) {
	tests := []struct {
		eventType InvasionEventType
		want      string
	}{
		{WaveStarted, "WaveStarted"},
		{WaveCompleted, "WaveCompleted"},
		{WaveFailed, "WaveFailed"},
		{InvaderMoved, "InvaderMoved"},
		{InvaderDefeated, "InvaderDefeated"},
		{InvaderRetreating, "InvaderRetreating"},
		{InvaderEscaped, "InvaderEscaped"},
		{CombatOccurred, "CombatOccurred"},
		{BeastDefeated, "BeastDefeated"},
		{TrapTriggered, "TrapTriggered"},
		{GoalAchievedEvent, "GoalAchievedEvent"},
		{InvasionEventType(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.eventType.String(); got != tt.want {
			t.Errorf("InvasionEventType(%d).String() = %q, want %q", tt.eventType, got, tt.want)
		}
	}
}

// TestEngine_RetreatPathAndEscape tests that retreating invaders move toward entry and escape.
func TestEngine_RetreatPathAndEscape(t *testing.T) {
	engine, cave, rooms, _ := setupEngineTest(t)

	inv := makeTestInvader(1, "warrior", types.Wood, 2, NewDestroyCoreGoal())
	inv.State = Retreating
	inv.Memory.Visit(1, 0, cave, rooms)
	inv.Memory.Visit(2, 1, cave, rooms)
	wave := makeTestWave(1, 0, []*Invader{inv})
	wave.State = Active
	waves := []*InvasionWave{wave}
	roomChi := make(map[int]*fengshui.RoomChi)

	// First tick: should move from room 2 to room 1.
	events := engine.Tick(2, waves, nil, rooms, nil, roomChi)

	retreatMove := false
	escaped := false
	for _, e := range events {
		if e.Type == InvaderRetreating && e.InvaderID == 1 && e.RoomID == 1 {
			retreatMove = true
		}
		if e.Type == InvaderEscaped && e.InvaderID == 1 {
			escaped = true
		}
	}
	if !retreatMove {
		t.Error("expected InvaderRetreating movement toward entry room")
	}
	if !escaped {
		t.Error("expected InvaderEscaped when reaching entry room")
	}
}
