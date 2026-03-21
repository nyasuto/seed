package invasion

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

func newTestCaveWithStorageRoom(t *testing.T) (*world.Cave, int) {
	t.Helper()
	cave, err := world.NewCave(16, 16)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}
	room, err := cave.AddRoom("storage", types.Pos{X: 2, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding storage room: %v", err)
	}
	return cave, room.ID
}

func TestStealTreasureGoal_Type(t *testing.T) {
	g := NewStealTreasureGoal()
	if g.Type() != StealTreasure {
		t.Errorf("expected StealTreasure, got %v", g.Type())
	}
}

func TestStealTreasureGoal_TargetRoomID_KnownTreasure(t *testing.T) {
	cave, storageID := newTestCaveWithStorageRoom(t)
	g := NewStealTreasureGoal()

	mem := &ExplorationMemory{KnownTreasureRooms: []int{storageID}}
	inv := &Invader{CurrentRoomID: 999}

	target := g.TargetRoomID(cave, inv, mem)
	if target != storageID {
		t.Errorf("expected target %d, got %d", storageID, target)
	}
}

func TestStealTreasureGoal_TargetRoomID_UnknownTreasure(t *testing.T) {
	cave, storageID := newTestCaveWithStorageRoom(t)
	g := NewStealTreasureGoal()

	mem := &ExplorationMemory{}
	inv := &Invader{CurrentRoomID: 999}

	// Even without known treasure rooms, scanning finds it.
	target := g.TargetRoomID(cave, inv, mem)
	if target != storageID {
		t.Errorf("expected target %d from scan, got %d", storageID, target)
	}
}

func TestStealTreasureGoal_TargetRoomID_NoStorageRoom(t *testing.T) {
	cave, err := world.NewCave(16, 16)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}
	g := NewStealTreasureGoal()
	mem := &ExplorationMemory{}
	inv := &Invader{CurrentRoomID: 999}

	target := g.TargetRoomID(cave, inv, mem)
	if target != 0 {
		t.Errorf("expected 0 for missing storage, got %d", target)
	}
}

func TestStealTreasureGoal_TargetRoomID_NilMemory(t *testing.T) {
	cave, storageID := newTestCaveWithStorageRoom(t)
	g := NewStealTreasureGoal()

	target := g.TargetRoomID(cave, &Invader{CurrentRoomID: 999}, nil)
	if target != storageID {
		t.Errorf("expected target %d with nil memory, got %d", storageID, target)
	}
}

func TestStealTreasureGoal_TargetRoomID_NearestOfMultiple(t *testing.T) {
	cave, err := world.NewCave(32, 32)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}
	// Add an anchor room where the invader is located.
	anchorRoom, err := cave.AddRoom("dragon_hole", types.Pos{X: 1, Y: 1}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding anchor room: %v", err)
	}
	// Add a far storage room.
	farRoom, err := cave.AddRoom("storage", types.Pos{X: 20, Y: 20}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding far storage room: %v", err)
	}
	// Add a near storage room.
	nearRoom, err := cave.AddRoom("storage", types.Pos{X: 5, Y: 5}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding near storage room: %v", err)
	}

	g := NewStealTreasureGoal()
	inv := &Invader{CurrentRoomID: anchorRoom.ID}
	mem := &ExplorationMemory{KnownTreasureRooms: []int{farRoom.ID, nearRoom.ID}}

	target := g.TargetRoomID(cave, inv, mem)
	if target != nearRoom.ID {
		t.Errorf("expected nearest room %d, got %d", nearRoom.ID, target)
	}
}

func TestStealTreasureGoal_IsAchieved(t *testing.T) {
	cave, storageID := newTestCaveWithStorageRoom(t)
	g := NewStealTreasureGoal()

	tests := []struct {
		name     string
		roomID   int
		achieved bool
	}{
		{"not in storage room", 999, false},
		{"in storage room", storageID, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invader{CurrentRoomID: tt.roomID}
			if got := g.IsAchieved(cave, inv); got != tt.achieved {
				t.Errorf("IsAchieved() = %v, want %v", got, tt.achieved)
			}
		})
	}
}
