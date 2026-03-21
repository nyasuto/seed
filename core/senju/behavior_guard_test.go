package senju

import "testing"

func TestGuardBehavior_NoInvaders_Stay(t *testing.T) {
	guard := &GuardBehavior{}
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:          beast,
		RoomID:         10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:     map[int][]int{10: {1}},
		InvaderRoomIDs: map[int][]int{},
		RoomChi:        nil,
	}

	action := guard.DecideAction(ctx)
	if action.Type != Stay {
		t.Errorf("expected Stay, got %s", action.Type)
	}
}

func TestGuardBehavior_InvaderPresent_Attack(t *testing.T) {
	guard := &GuardBehavior{}
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:          beast,
		RoomID:         10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:     map[int][]int{10: {1}},
		InvaderRoomIDs: map[int][]int{10: {99, 100}},
		RoomChi:        nil,
	}

	action := guard.DecideAction(ctx)
	if action.Type != Attack {
		t.Errorf("expected Attack, got %s", action.Type)
	}
	if action.TargetBeastID != 99 {
		t.Errorf("expected target 99 (first invader), got %d", action.TargetBeastID)
	}
}

func TestGuardBehavior_InvaderInAdjacentRoom_Stay(t *testing.T) {
	guard := &GuardBehavior{}
	beast := &Beast{ID: 1, RoomID: 10}

	ctx := BehaviorContext{
		Beast:          beast,
		RoomID:         10,
		AdjacentRoomIDs: []int{11, 12},
		RoomBeasts:     map[int][]int{10: {1}},
		InvaderRoomIDs: map[int][]int{11: {99}},
		RoomChi:        nil,
	}

	action := guard.DecideAction(ctx)
	if action.Type != Stay {
		t.Errorf("expected Stay (invader is in adjacent room, not current room), got %s", action.Type)
	}
}

func TestGuardBehavior_Type(t *testing.T) {
	guard := &GuardBehavior{}
	if guard.Type() != Guard {
		t.Errorf("expected Guard, got %s", guard.Type())
	}
}
