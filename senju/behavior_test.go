package senju

import "testing"

// --- Patrol Tests ---

func TestPatrolBehavior_RouteGeneration(t *testing.T) {
	patrol := NewPatrolBehavior(10, []int{11, 12, 13}, 0)
	if len(patrol.PatrolRoute) != 4 {
		t.Fatalf("expected route length 4, got %d", len(patrol.PatrolRoute))
	}
	if patrol.PatrolRoute[0] != 10 {
		t.Errorf("expected home room 10 at index 0, got %d", patrol.PatrolRoute[0])
	}
}

func TestPatrolBehavior_CyclesThroughRooms(t *testing.T) {
	// RestTicks=0 means move every tick.
	patrol := NewPatrolBehavior(10, []int{11, 12}, 0)

	beast := &Beast{ID: 1, RoomID: 10}
	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{},
	}

	// First action: route index advances from 0 (room 10) to 1 (room 11).
	action := patrol.DecideAction(ctx)
	if action.Type != MoveToRoom {
		t.Fatalf("tick 1: expected MoveToRoom, got %s", action.Type)
	}
	if action.TargetRoomID != 11 {
		t.Errorf("tick 1: expected target room 11, got %d", action.TargetRoomID)
	}

	// Simulate beast moved to room 11.
	ctx.RoomID = 11
	action = patrol.DecideAction(ctx)
	if action.Type != MoveToRoom {
		t.Fatalf("tick 2: expected MoveToRoom, got %s", action.Type)
	}
	if action.TargetRoomID != 12 {
		t.Errorf("tick 2: expected target room 12, got %d", action.TargetRoomID)
	}

	// Simulate beast moved to room 12. Next should cycle back to 10.
	ctx.RoomID = 12
	action = patrol.DecideAction(ctx)
	if action.Type != MoveToRoom {
		t.Fatalf("tick 3: expected MoveToRoom, got %s", action.Type)
	}
	if action.TargetRoomID != 10 {
		t.Errorf("tick 3: expected target room 10 (cycle), got %d", action.TargetRoomID)
	}
}

func TestPatrolBehavior_RestTicks(t *testing.T) {
	// RestTicks=2 means stay 2 ticks before moving.
	patrol := NewPatrolBehavior(10, []int{11}, 2)

	beast := &Beast{ID: 1, RoomID: 10}
	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{},
	}

	// Tick 1: rest (0 < 2)
	action := patrol.DecideAction(ctx)
	if action.Type != Stay {
		t.Errorf("tick 1: expected Stay (resting), got %s", action.Type)
	}

	// Tick 2: rest (1 < 2)
	action = patrol.DecideAction(ctx)
	if action.Type != Stay {
		t.Errorf("tick 2: expected Stay (resting), got %s", action.Type)
	}

	// Tick 3: rest done (2 >= 2), move to next room.
	action = patrol.DecideAction(ctx)
	if action.Type != MoveToRoom {
		t.Errorf("tick 3: expected MoveToRoom, got %s", action.Type)
	}
	if action.TargetRoomID != 11 {
		t.Errorf("tick 3: expected target room 11, got %d", action.TargetRoomID)
	}
}

func TestPatrolBehavior_InvaderInCurrentRoom_Attack(t *testing.T) {
	patrol := NewPatrolBehavior(10, []int{11, 12}, 0)
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{10: {99}},
	}

	action := patrol.DecideAction(ctx)
	if action.Type != Attack {
		t.Errorf("expected Attack, got %s", action.Type)
	}
	if action.TargetBeastID != 99 {
		t.Errorf("expected target invader 99, got %d", action.TargetBeastID)
	}
}

func TestPatrolBehavior_InvaderInAdjacentRoom_MoveToward(t *testing.T) {
	patrol := NewPatrolBehavior(10, []int{11, 12}, 0)
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{12: {99}},
	}

	action := patrol.DecideAction(ctx)
	if action.Type != MoveToRoom {
		t.Errorf("expected MoveToRoom toward invader, got %s", action.Type)
	}
	if action.TargetRoomID != 12 {
		t.Errorf("expected target room 12 (where invader is), got %d", action.TargetRoomID)
	}
}

func TestPatrolBehavior_Type(t *testing.T) {
	patrol := NewPatrolBehavior(10, nil, 0)
	if patrol.Type() != Patrol {
		t.Errorf("expected Patrol, got %s", patrol.Type())
	}
}

// --- Chase Tests ---

func TestChaseBehavior_InvaderInSameRoom_Attack(t *testing.T) {
	chase := NewChaseBehavior(99, 10)
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{10: {99}},
	}

	action := chase.DecideAction(ctx)
	if action.Type != Attack {
		t.Errorf("expected Attack, got %s", action.Type)
	}
	if action.TargetBeastID != 99 {
		t.Errorf("expected target 99, got %d", action.TargetBeastID)
	}
}

func TestChaseBehavior_InvaderInAdjacentRoom_MoveToward(t *testing.T) {
	chase := NewChaseBehavior(99, 10)
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{12: {99}},
	}

	action := chase.DecideAction(ctx)
	if action.Type != MoveToRoom {
		t.Errorf("expected MoveToRoom, got %s", action.Type)
	}
	if action.TargetRoomID != 12 {
		t.Errorf("expected target room 12, got %d", action.TargetRoomID)
	}
}

func TestChaseBehavior_InvaderNotVisible_Stay(t *testing.T) {
	chase := NewChaseBehavior(99, 10)
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{},
	}

	action := chase.DecideAction(ctx)
	if action.Type != Stay {
		t.Errorf("expected Stay (invader not visible), got %s", action.Type)
	}
}

func TestChaseBehavior_Timeout(t *testing.T) {
	chase := NewChaseBehavior(99, 3)
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{},
	}

	// Tick through the timeout.
	for i := 0; i < 3; i++ {
		action := chase.DecideAction(ctx)
		if action.Type != Stay {
			t.Errorf("tick %d: expected Stay, got %s", i+1, action.Type)
		}
	}

	// One more tick should exceed timeout.
	action := chase.DecideAction(ctx)
	if action.Type != Stay {
		t.Errorf("after timeout: expected Stay, got %s", action.Type)
	}
	if !chase.TimedOut() {
		t.Error("expected TimedOut() to be true")
	}
}

func TestChaseBehavior_AttacksOtherInvaderInSameRoom(t *testing.T) {
	// Target is 99, but only invader 88 is in the room.
	chase := NewChaseBehavior(99, 10)
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{10: {88}},
	}

	action := chase.DecideAction(ctx)
	if action.Type != Attack {
		t.Errorf("expected Attack, got %s", action.Type)
	}
	if action.TargetBeastID != 88 {
		t.Errorf("expected target 88 (first available invader), got %d", action.TargetBeastID)
	}
}

func TestChaseBehavior_Type(t *testing.T) {
	chase := NewChaseBehavior(99, 10)
	if chase.Type() != Chase {
		t.Errorf("expected Chase, got %s", chase.Type())
	}
}

// --- Flee Tests ---

func TestShouldFlee_BelowThreshold(t *testing.T) {
	beast := &Beast{ID: 1, HP: 20, MaxHP: 100}
	if !ShouldFlee(beast, 0.25) {
		t.Error("expected ShouldFlee to be true at 20% HP")
	}
}

func TestShouldFlee_AtThreshold(t *testing.T) {
	beast := &Beast{ID: 1, HP: 25, MaxHP: 100}
	if !ShouldFlee(beast, 0.25) {
		t.Error("expected ShouldFlee to be true at exactly 25% HP")
	}
}

func TestShouldFlee_AboveThreshold(t *testing.T) {
	beast := &Beast{ID: 1, HP: 50, MaxHP: 100}
	if ShouldFlee(beast, 0.25) {
		t.Error("expected ShouldFlee to be false at 50% HP")
	}
}

func TestShouldFlee_ZeroMaxHP(t *testing.T) {
	beast := &Beast{ID: 1, HP: 0, MaxHP: 0}
	if ShouldFlee(beast, 0.25) {
		t.Error("expected ShouldFlee to be false when MaxHP is 0")
	}
}

func TestFleeBehavior_InRecoveryRoom_Stay(t *testing.T) {
	flee := NewFleeBehavior(0.25, map[int]string{10: "recovery_room"})
	beast := &Beast{ID: 1, RoomID: 10, HP: 10, MaxHP: 100}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{},
	}

	action := flee.DecideAction(ctx)
	if action.Type != Stay {
		t.Errorf("expected Stay in recovery room, got %s", action.Type)
	}
}

func TestFleeBehavior_AdjacentRecoveryRoom_Retreat(t *testing.T) {
	flee := NewFleeBehavior(0.25, map[int]string{
		10: "senju_room",
		11: "recovery_room",
		12: "senju_room",
	})
	beast := &Beast{ID: 1, RoomID: 10, HP: 10, MaxHP: 100}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{10: {99}},
	}

	action := flee.DecideAction(ctx)
	if action.Type != Retreat {
		t.Errorf("expected Retreat toward recovery room, got %s", action.Type)
	}
	if action.TargetRoomID != 11 {
		t.Errorf("expected target room 11 (recovery room), got %d", action.TargetRoomID)
	}
}

func TestFleeBehavior_NoRecoveryRoom_FleeFromInvaders(t *testing.T) {
	flee := NewFleeBehavior(0.25, map[int]string{
		10: "senju_room",
		11: "senju_room",
		12: "senju_room",
	})
	beast := &Beast{ID: 1, RoomID: 10, HP: 10, MaxHP: 100}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{11: {99}},
	}

	action := flee.DecideAction(ctx)
	if action.Type != Retreat {
		t.Errorf("expected Retreat, got %s", action.Type)
	}
	// Room 12 has no invaders, room 11 has one, so prefer room 12.
	if action.TargetRoomID != 12 {
		t.Errorf("expected target room 12 (away from invader), got %d", action.TargetRoomID)
	}
}

func TestFleeBehavior_NoAdjacentRooms_Stay(t *testing.T) {
	flee := NewFleeBehavior(0.25, map[int]string{10: "senju_room"})
	beast := &Beast{ID: 1, RoomID: 10, HP: 10, MaxHP: 100}

	ctx := BehaviorContext{
		Beast:           beast,
		RoomID:          10,
		AdjacentRoomIDs: []int{},
		RoomBeasts:      map[int][]int{10: {1}},
		InvaderRoomIDs:  map[int][]int{},
	}

	action := flee.DecideAction(ctx)
	if action.Type != Stay {
		t.Errorf("expected Stay (no adjacent rooms to flee to), got %s", action.Type)
	}
}

func TestFleeBehavior_Type(t *testing.T) {
	flee := NewFleeBehavior(0.25, nil)
	if flee.Type() != Flee {
		t.Errorf("expected Flee, got %s", flee.Type())
	}
}
