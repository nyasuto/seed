package input

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

func TestSummonBeastFlow_RoomAndFireElement(t *testing.T) {
	flow := NewSummonBeastFlow()

	if flow.Step() != SummonStepSelectRoom {
		t.Fatalf("expected SummonStepSelectRoom, got %d", flow.Step())
	}

	// Select a room cell with capacity.
	err := flow.TrySelectRoom(world.RoomFloor, 2, true)
	if err != nil {
		t.Fatalf("unexpected error selecting room: %v", err)
	}
	if flow.Step() != SummonStepSelectElement {
		t.Fatalf("expected SummonStepSelectElement, got %d", flow.Step())
	}
	if flow.RoomID() != 2 {
		t.Errorf("expected roomID 2, got %d", flow.RoomID())
	}

	// Select Fire element.
	action := flow.SelectElement(types.Fire)
	if action.Element != types.Fire {
		t.Errorf("expected Fire element, got %v", action.Element)
	}
	if action.ActionType() != "summon_beast" {
		t.Errorf("expected summon_beast action type, got %s", action.ActionType())
	}
	if flow.Step() != SummonStepComplete {
		t.Errorf("expected SummonStepComplete, got %d", flow.Step())
	}
}

func TestSummonBeastFlow_NonRoomCellRejected(t *testing.T) {
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
			flow := NewSummonBeastFlow()
			err := flow.TrySelectRoom(tt.cellType, 0, true)
			if err == nil {
				t.Fatalf("expected error for %s cell", tt.name)
			}
			if flow.Step() != SummonStepSelectRoom {
				t.Errorf("expected SummonStepSelectRoom after rejection, got %d", flow.Step())
			}
		})
	}
}

func TestSummonBeastFlow_RoomAtCapacityRejected(t *testing.T) {
	flow := NewSummonBeastFlow()

	err := flow.TrySelectRoom(world.RoomFloor, 2, false)
	if err == nil {
		t.Fatal("expected error for room at capacity")
	}
	if flow.Step() != SummonStepSelectRoom {
		t.Errorf("expected SummonStepSelectRoom after rejection, got %d", flow.Step())
	}
}

func TestSummonBeastFlow_Cancel(t *testing.T) {
	flow := NewSummonBeastFlow()
	_ = flow.TrySelectRoom(world.RoomFloor, 3, true)

	if flow.Step() != SummonStepSelectElement {
		t.Fatalf("expected SummonStepSelectElement, got %d", flow.Step())
	}

	flow.Cancel()

	if flow.Step() != SummonStepSelectRoom {
		t.Errorf("expected SummonStepSelectRoom after cancel, got %d", flow.Step())
	}
	if flow.RoomID() != 0 {
		t.Errorf("expected roomID 0 after cancel, got %d", flow.RoomID())
	}
}

func TestSummonBeastFlow_AllElements(t *testing.T) {
	elements := []types.Element{types.Wood, types.Fire, types.Earth, types.Metal, types.Water}

	for _, elem := range elements {
		t.Run(elem.String(), func(t *testing.T) {
			flow := NewSummonBeastFlow()
			_ = flow.TrySelectRoom(world.RoomFloor, 1, true)
			action := flow.SelectElement(elem)
			if action.Element != elem {
				t.Errorf("expected %v, got %v", elem, action.Element)
			}
		})
	}
}

func TestSummonBeastFlow_EntranceCellAccepted(t *testing.T) {
	flow := NewSummonBeastFlow()

	err := flow.TrySelectRoom(world.Entrance, 5, true)
	if err != nil {
		t.Fatalf("unexpected error for Entrance cell: %v", err)
	}
	if flow.Step() != SummonStepSelectElement {
		t.Errorf("expected SummonStepSelectElement, got %d", flow.Step())
	}
}
