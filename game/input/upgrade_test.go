package input

import (
	"testing"

	"github.com/nyasuto/seed/core/world"
)

func TestUpgradeRoomFlow_TrySelectRoom_Success(t *testing.T) {
	f := NewUpgradeRoomFlow()
	action, err := f.TrySelectRoom(world.RoomFloor, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.RoomID != 1 {
		t.Errorf("RoomID = %d, want 1", action.RoomID)
	}
	if !f.Complete() {
		t.Error("expected flow to be complete")
	}
}

func TestUpgradeRoomFlow_TrySelectRoom_Entrance(t *testing.T) {
	f := NewUpgradeRoomFlow()
	action, err := f.TrySelectRoom(world.Entrance, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.RoomID != 3 {
		t.Errorf("RoomID = %d, want 3", action.RoomID)
	}
}

func TestUpgradeRoomFlow_TrySelectRoom_NotRoom(t *testing.T) {
	tests := []struct {
		name     string
		cellType world.CellType
	}{
		{"Rock", world.Rock},
		{"HardRock", world.HardRock},
		{"Water", world.Water},
		{"CorridorFloor", world.CorridorFloor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewUpgradeRoomFlow()
			action, err := f.TrySelectRoom(tt.cellType, 0)
			if err == nil {
				t.Fatal("expected error for non-room cell")
			}
			if action != nil {
				t.Error("expected nil action on error")
			}
			if f.Complete() {
				t.Error("flow should not be complete on error")
			}
		})
	}
}

func TestUpgradeRoomFlow_TrySelectRoom_ZeroRoomID(t *testing.T) {
	f := NewUpgradeRoomFlow()
	action, err := f.TrySelectRoom(world.RoomFloor, 0)
	if err == nil {
		t.Fatal("expected error for zero roomID")
	}
	if action != nil {
		t.Error("expected nil action on error")
	}
}

func TestUpgradeRoomFlow_Cancel(t *testing.T) {
	f := NewUpgradeRoomFlow()
	_, _ = f.TrySelectRoom(world.RoomFloor, 1)
	if !f.Complete() {
		t.Fatal("expected complete after selection")
	}
	f.Cancel()
	if f.Complete() {
		t.Error("expected not complete after cancel")
	}
}
