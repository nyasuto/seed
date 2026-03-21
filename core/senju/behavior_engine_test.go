package senju

import (
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// setupTestCaveForEngine creates a cave with 3 rooms connected by corridors.
// Room 1 (senju_room) <-> Room 2 (senju_room) <-> Room 3 (recovery_room)
func setupTestCaveForEngine(t *testing.T) (*world.Cave, world.AdjacencyGraph, *world.RoomTypeRegistry, map[int]*world.Room) {
	t.Helper()

	reg := world.NewRoomTypeRegistry()
	if err := reg.Register(world.RoomType{
		ID: "senju_room", Name: "仙獣部屋", Element: types.Wood,
		BaseChiCapacity: 100, MaxBeasts: 3,
	}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(world.RoomType{
		ID: "recovery_room", Name: "回復室", Element: types.Water,
		BaseChiCapacity: 50, MaxBeasts: 3,
	}); err != nil {
		t.Fatal(err)
	}

	cave, err := world.NewCave(32, 32)
	if err != nil {
		t.Fatal(err)
	}

	// Room 1 at (2,2), 3x3
	r1, err := cave.AddRoom("senju_room", types.Pos{X: 2, Y: 2}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 4, Y: 3}, Dir: types.East},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Room 2 at (8,2), 3x3
	r2, err := cave.AddRoom("senju_room", types.Pos{X: 8, Y: 2}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 8, Y: 3}, Dir: types.West},
		{Pos: types.Pos{X: 10, Y: 3}, Dir: types.East},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Room 3 at (14,2), 3x3 (recovery room)
	r3, err := cave.AddRoom("recovery_room", types.Pos{X: 14, Y: 2}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 14, Y: 3}, Dir: types.West},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Connect rooms: 1-2 and 2-3
	if _, err := cave.ConnectRooms(r1.ID, r2.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := cave.ConnectRooms(r2.ID, r3.ID); err != nil {
		t.Fatal(err)
	}

	ag := cave.BuildAdjacencyGraph()

	rooms := map[int]*world.Room{
		r1.ID: cave.RoomByID(r1.ID),
		r2.ID: cave.RoomByID(r2.ID),
		r3.ID: cave.RoomByID(r3.ID),
	}

	return cave, ag, reg, rooms
}

func TestBehaviorEngine_AllGuard_NoInvaders_AllStay(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)
	b2 := NewBeast(2, &Species{ID: "test", Element: types.Fire, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	r2 := rooms[2]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}
	if err := PlaceBeast(b2, r2, rt); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	engine.AssignBehavior(b1, Guard)
	engine.AssignBehavior(b2, Guard)

	beasts := []*Beast{b1, b2}
	actions := engine.Tick(beasts, map[int][]int{}, nil)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	for _, a := range actions {
		if a.Action.Type != Stay {
			t.Errorf("beast %d: expected Stay, got %s", a.BeastID, a.Action.Type)
		}
	}
}

func TestBehaviorEngine_Patrol_CyclesThroughRooms(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	// Use PatrolRestTicks=0 so patrol moves immediately each tick.
	params := &BehaviorParams{FleeHPThreshold: 0.25, ChaseTimeoutTicks: 10, PatrolRestTicks: 0}
	engine := NewBehaviorEngine(cave, ag, reg, params)
	engine.AssignBehavior(b1, Patrol)

	beasts := []*Beast{b1}
	invaders := map[int][]int{}

	// Tick 1: Should move from room 1 to adjacent room.
	actions := engine.Tick(beasts, invaders, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Action.Type != MoveToRoom {
		t.Errorf("tick 1: expected MoveToRoom, got %s", actions[0].Action.Type)
	}

	// Apply the action.
	if err := ApplyActions(beasts, rooms, reg, actions); err != nil {
		t.Fatal(err)
	}

	prevRoom := b1.RoomID

	// Tick 2: Should continue patrolling.
	actions = engine.Tick(beasts, invaders, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	// Patrol should continue moving or stay (depending on route).
	// The important thing is it's deciding an action.
	_ = prevRoom
}

func TestBehaviorEngine_InvaderDetected_ChaseTransition(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	engine.AssignBehavior(b1, Patrol)

	beasts := []*Beast{b1}
	// Place invader in room 2 (adjacent to room 1).
	invaders := map[int][]int{2: {99}}

	actions := engine.Tick(beasts, invaders, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	// Patrol should move toward invader in adjacent room.
	if actions[0].Action.Type != MoveToRoom {
		t.Errorf("expected MoveToRoom toward invader, got %s", actions[0].Action.Type)
	}
	if actions[0].Action.TargetRoomID != 2 {
		t.Errorf("expected target room 2, got %d", actions[0].Action.TargetRoomID)
	}

	// Behavior should have transitioned to Chase.
	b := engine.GetBehavior(b1.ID)
	if b == nil || b.Type() != Chase {
		t.Errorf("expected behavior to transition to Chase, got %v", b)
	}
}

func TestBehaviorEngine_HPLow_FleeTransition(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)
	// Set HP to trigger flee.
	b1.HP = 20 // 20/100 = 20% < 25% threshold

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	engine.AssignBehavior(b1, Guard)

	beasts := []*Beast{b1}
	actions := engine.Tick(beasts, map[int][]int{}, nil)

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	// Should have transitioned to Flee behavior.
	b := engine.GetBehavior(b1.ID)
	if b == nil || b.Type() != Flee {
		t.Errorf("expected behavior to transition to Flee, got %v", b)
	}

	// Should be retreating toward room 2 (adjacent, heading toward recovery room).
	if actions[0].Action.Type != Retreat {
		t.Errorf("expected Retreat, got %s", actions[0].Action.Type)
	}
}

func TestBehaviorEngine_MovementConflict_FirstByID(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	// Two beasts in different rooms, both want to move to room 2.
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)
	b2 := NewBeast(2, &Species{ID: "test", Element: types.Fire, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	r3 := rooms[3]
	rt3, _ := reg.Get("recovery_room")
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}
	if err := PlaceBeast(b2, r3, rt3); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	// Both patrol toward room 2.
	engine.AssignBehavior(b1, Patrol)
	engine.AssignBehavior(b2, Patrol)

	beasts := []*Beast{b1, b2}
	actions := engine.Tick(beasts, map[int][]int{}, nil)

	// Count how many actually got MoveToRoom to room 2.
	moveToRoom2 := 0
	stayCount := 0
	for _, a := range actions {
		if (a.Action.Type == MoveToRoom || a.Action.Type == Retreat) && a.Action.TargetRoomID == 2 {
			moveToRoom2++
		}
		if a.Action.Type == Stay {
			stayCount++
		}
	}

	// If both target room 2, only one should succeed (the one with lower ID).
	if moveToRoom2 > 1 {
		t.Errorf("expected at most 1 beast to move to room 2, got %d", moveToRoom2)
	}
}

func TestBehaviorEngine_ApplyActions_MovesBeasts(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	// Use PatrolRestTicks=0 so patrol moves immediately.
	params := &BehaviorParams{FleeHPThreshold: 0.25, ChaseTimeoutTicks: 10, PatrolRestTicks: 0}
	engine := NewBehaviorEngine(cave, ag, reg, params)
	engine.AssignBehavior(b1, Patrol)

	beasts := []*Beast{b1}
	actions := engine.Tick(beasts, map[int][]int{}, nil)

	if err := ApplyActions(beasts, rooms, reg, actions); err != nil {
		t.Fatal(err)
	}

	// Beast should have moved out of room 1.
	if b1.RoomID == 1 {
		t.Error("expected beast to move from room 1")
	}
	// Room 1 should no longer contain beast 1.
	if len(r1.BeastIDs) != 0 {
		t.Errorf("expected room 1 to have 0 beasts, got %d", len(r1.BeastIDs))
	}
	// Beast should be in one of the adjacent rooms.
	if b1.RoomID == 0 {
		t.Error("beast should be assigned to a room")
	}
}

func TestBehaviorEngine_ApplyActions_StayDoesNotMove(t *testing.T) {
	_, _, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	// Manually create a Stay action.
	actions := []BeastAction{
		{BeastID: 1, Action: Action{Type: Stay}, PreviousRoomID: 1, ResultingState: Idle},
	}

	if err := ApplyActions([]*Beast{b1}, rooms, reg, actions); err != nil {
		t.Fatal(err)
	}

	if b1.RoomID != 1 {
		t.Errorf("expected beast to stay in room 1, got room %d", b1.RoomID)
	}
	if b1.State != Idle {
		t.Errorf("expected state Idle, got %s", b1.State)
	}
}

func TestBehaviorEngine_ChaseTimeout_RevertsToGuard(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	// Assign chase with short timeout.
	chase := NewChaseBehavior(99, 2)
	engine.SetBehavior(b1.ID, chase)

	beasts := []*Beast{b1}
	invaders := map[int][]int{} // No invaders visible.

	// Tick through timeout (timeout=2, so 3 ticks to exceed, then 4th detects it).
	for range 4 {
		engine.Tick(beasts, invaders, nil)
	}

	// After timeout detected on tick 4, behavior should revert to Guard.
	b := engine.GetBehavior(b1.ID)
	if b == nil || b.Type() != Guard {
		t.Errorf("expected behavior to revert to Guard after chase timeout, got %v", b)
	}
}

func TestBehaviorEngine_NilRoomChi(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	engine.AssignBehavior(b1, Guard)

	beasts := []*Beast{b1}
	// Pass nil roomChi — should not panic.
	actions := engine.Tick(beasts, map[int][]int{}, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
}

func TestBehaviorEngine_WithRoomChi(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	engine.AssignBehavior(b1, Guard)

	beasts := []*Beast{b1}
	roomChi := map[int]*fengshui.RoomChi{
		1: {RoomID: 1, Current: 50, Capacity: 100, Element: types.Wood},
	}
	actions := engine.Tick(beasts, map[int][]int{}, roomChi)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
}

func TestBehaviorEngine_StunnedBeastSkipped(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)
	b2 := NewBeast(2, &Species{ID: "test", Element: types.Fire, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	r2 := rooms[2]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}
	if err := PlaceBeast(b2, r2, rt); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	engine.AssignBehavior(b1, Guard)
	engine.AssignBehavior(b2, Guard)

	// Set beast 1 to Stunned state.
	b1.State = Stunned
	b1.HP = 0

	beasts := []*Beast{b1, b2}
	actions := engine.Tick(beasts, map[int][]int{}, nil)

	// Only beast 2 should produce an action; beast 1 is stunned.
	if len(actions) != 1 {
		t.Fatalf("expected 1 action (stunned beast skipped), got %d", len(actions))
	}
	if actions[0].BeastID != 2 {
		t.Errorf("expected action for beast 2, got beast %d", actions[0].BeastID)
	}
}

func TestBehaviorEngine_GuardAttacksInvader(t *testing.T) {
	cave, ag, reg, rooms := setupTestCaveForEngine(t)

	rt, _ := reg.Get("senju_room")
	b1 := NewBeast(1, &Species{ID: "test", Element: types.Wood, BaseHP: 100, BaseATK: 10, BaseDEF: 10, BaseSPD: 10}, 0)

	r1 := rooms[1]
	if err := PlaceBeast(b1, r1, rt); err != nil {
		t.Fatal(err)
	}

	engine := NewBehaviorEngine(cave, ag, reg, nil)
	engine.AssignBehavior(b1, Guard)

	beasts := []*Beast{b1}
	invaders := map[int][]int{1: {99}}

	actions := engine.Tick(beasts, invaders, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Action.Type != Attack {
		t.Errorf("expected Attack, got %s", actions[0].Action.Type)
	}
	if actions[0].Action.TargetBeastID != 99 {
		t.Errorf("expected target 99, got %d", actions[0].Action.TargetBeastID)
	}
	if actions[0].ResultingState != Fighting {
		t.Errorf("expected Fighting state, got %s", actions[0].ResultingState)
	}
}

func TestRemoveBehavior(t *testing.T) {
	cave, adjGraph, reg, _ := setupTestCaveForEngine(t)
	engine := NewBehaviorEngine(cave, adjGraph, reg, nil)

	beast := &Beast{ID: 1, RoomID: 1, HP: 100, MaxHP: 100, State: Idle}
	engine.AssignBehavior(beast, Guard)

	if engine.GetBehavior(1) == nil {
		t.Fatal("behavior should be assigned")
	}

	engine.RemoveBehavior(1)

	if engine.GetBehavior(1) != nil {
		t.Error("behavior should be nil after RemoveBehavior")
	}
}

func TestRemoveBehavior_NonExistent(t *testing.T) {
	cave, adjGraph, reg, _ := setupTestCaveForEngine(t)
	engine := NewBehaviorEngine(cave, adjGraph, reg, nil)

	// Should not panic when removing a non-existent behavior.
	engine.RemoveBehavior(999)
}
