package input

import (
	"testing"

	"github.com/nyasuto/seed/core/world"
)

func TestDigCorridorFlow_SelectTwoRooms(t *testing.T) {
	flow := NewDigCorridorFlow()

	if flow.Step() != CorridorStepSelectFirst {
		t.Fatalf("expected CorridorStepSelectFirst, got %d", flow.Step())
	}

	// Select first room (room 1).
	action, err := flow.TrySelectRoom(world.RoomFloor, 1)
	if err != nil {
		t.Fatalf("unexpected error selecting first room: %v", err)
	}
	if action != nil {
		t.Fatal("expected nil action after first room selection")
	}
	if flow.Step() != CorridorStepSelectSecond {
		t.Fatalf("expected CorridorStepSelectSecond, got %d", flow.Step())
	}
	if flow.FromRoomID() != 1 {
		t.Errorf("expected fromRoomID=1, got %d", flow.FromRoomID())
	}

	// Select second room (room 3).
	action, err = flow.TrySelectRoom(world.RoomFloor, 3)
	if err != nil {
		t.Fatalf("unexpected error selecting second room: %v", err)
	}
	if action == nil {
		t.Fatal("expected non-nil action after second room selection")
	}
	if action.FromRoomID != 1 {
		t.Errorf("expected FromRoomID=1, got %d", action.FromRoomID)
	}
	if action.ToRoomID != 3 {
		t.Errorf("expected ToRoomID=3, got %d", action.ToRoomID)
	}
	if flow.Step() != CorridorStepComplete {
		t.Errorf("expected CorridorStepComplete, got %d", flow.Step())
	}
}

func TestDigCorridorFlow_NonRoomCellRejected(t *testing.T) {
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
			flow := NewDigCorridorFlow()
			_, err := flow.TrySelectRoom(tt.cellType, 1)
			if err == nil {
				t.Fatalf("expected error for %s cell", tt.name)
			}
			if flow.Step() != CorridorStepSelectFirst {
				t.Errorf("expected CorridorStepSelectFirst after rejection, got %d", flow.Step())
			}
		})
	}
}

func TestDigCorridorFlow_SameRoomRejected(t *testing.T) {
	flow := NewDigCorridorFlow()

	// Select first room.
	_, err := flow.TrySelectRoom(world.RoomFloor, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to select the same room as second.
	_, err = flow.TrySelectRoom(world.RoomFloor, 2)
	if err == nil {
		t.Fatal("expected error for same room selection")
	}
	// Should remain at second step (not reset).
	if flow.Step() != CorridorStepSelectSecond {
		t.Errorf("expected CorridorStepSelectSecond after same-room rejection, got %d", flow.Step())
	}
}

func TestDigCorridorFlow_CancelResetsToFirst(t *testing.T) {
	flow := NewDigCorridorFlow()

	// Select first room, then cancel.
	_, _ = flow.TrySelectRoom(world.RoomFloor, 1)
	if flow.Step() != CorridorStepSelectSecond {
		t.Fatalf("expected CorridorStepSelectSecond, got %d", flow.Step())
	}

	flow.Cancel()

	if flow.Step() != CorridorStepSelectFirst {
		t.Errorf("expected CorridorStepSelectFirst after cancel, got %d", flow.Step())
	}
	if flow.FromRoomID() != 0 {
		t.Errorf("expected fromRoomID=0 after cancel, got %d", flow.FromRoomID())
	}
}

func TestDigCorridorFlow_EntranceCellAccepted(t *testing.T) {
	flow := NewDigCorridorFlow()

	// Entrance cells should also be accepted as room cells.
	_, err := flow.TrySelectRoom(world.Entrance, 1)
	if err != nil {
		t.Fatalf("unexpected error for Entrance cell: %v", err)
	}
	if flow.Step() != CorridorStepSelectSecond {
		t.Errorf("expected CorridorStepSelectSecond, got %d", flow.Step())
	}

	action, err := flow.TrySelectRoom(world.Entrance, 2)
	if err != nil {
		t.Fatalf("unexpected error for Entrance cell: %v", err)
	}
	if action == nil {
		t.Fatal("expected non-nil action")
	}
	if action.FromRoomID != 1 || action.ToRoomID != 2 {
		t.Errorf("expected FromRoomID=1, ToRoomID=2, got %d, %d", action.FromRoomID, action.ToRoomID)
	}
}

func TestDigCorridorFlow_ZeroRoomIDRejected(t *testing.T) {
	flow := NewDigCorridorFlow()

	// RoomID=0 means the cell doesn't belong to any room.
	_, err := flow.TrySelectRoom(world.RoomFloor, 0)
	if err == nil {
		t.Fatal("expected error for roomID=0")
	}
	if flow.Step() != CorridorStepSelectFirst {
		t.Errorf("expected CorridorStepSelectFirst after rejection, got %d", flow.Step())
	}
}
