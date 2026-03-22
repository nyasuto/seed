package input

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

func testRoomTypeRegistry() *world.RoomTypeRegistry {
	reg := world.NewRoomTypeRegistry()
	_ = reg.Register(world.RoomType{ID: "fire_room", Element: types.Fire, MaxBeasts: 1})
	_ = reg.Register(world.RoomType{ID: "water_room", Element: types.Water, MaxBeasts: 1})
	_ = reg.Register(world.RoomType{ID: "wood_room", Element: types.Wood, MaxBeasts: 1})
	_ = reg.Register(world.RoomType{ID: "metal_room", Element: types.Metal, MaxBeasts: 1})
	_ = reg.Register(world.RoomType{ID: "earth_room", Element: types.Earth, MaxBeasts: 1})
	return reg
}

func TestDigRoomFlow_RockCellAndFireElement(t *testing.T) {
	flow := NewDigRoomFlow()

	if flow.Step() != StepSelectCell {
		t.Fatalf("expected StepSelectCell, got %d", flow.Step())
	}

	// Select a Rock cell at (5, 3).
	err := flow.TrySelectCell(5, 3, world.Rock)
	if err != nil {
		t.Fatalf("unexpected error selecting rock cell: %v", err)
	}
	if flow.Step() != StepSelectElement {
		t.Fatalf("expected StepSelectElement, got %d", flow.Step())
	}

	cx, cy := flow.SelectedCell()
	if cx != 5 || cy != 3 {
		t.Errorf("expected selected cell (5,3), got (%d,%d)", cx, cy)
	}

	// Select Fire element.
	action, err := flow.SelectElement(types.Fire, testRoomTypeRegistry())
	if err != nil {
		t.Fatalf("unexpected error selecting element: %v", err)
	}

	if action.RoomTypeID != "fire_room" {
		t.Errorf("expected room type fire_room, got %s", action.RoomTypeID)
	}
	if action.Pos.X != 5 || action.Pos.Y != 3 {
		t.Errorf("expected pos (5,3), got (%d,%d)", action.Pos.X, action.Pos.Y)
	}
	if action.Width != DefaultRoomWidth {
		t.Errorf("expected width %d, got %d", DefaultRoomWidth, action.Width)
	}
	if action.Height != DefaultRoomHeight {
		t.Errorf("expected height %d, got %d", DefaultRoomHeight, action.Height)
	}
	if flow.Step() != StepComplete {
		t.Errorf("expected StepComplete, got %d", flow.Step())
	}
}

func TestDigRoomFlow_HardRockRejected(t *testing.T) {
	flow := NewDigRoomFlow()

	err := flow.TrySelectCell(5, 3, world.HardRock)
	if err == nil {
		t.Fatal("expected error for HardRock cell")
	}
	if flow.Step() != StepSelectCell {
		t.Errorf("expected StepSelectCell after rejection, got %d", flow.Step())
	}
}

func TestDigRoomFlow_WaterRejected(t *testing.T) {
	flow := NewDigRoomFlow()

	err := flow.TrySelectCell(5, 3, world.Water)
	if err == nil {
		t.Fatal("expected error for Water cell")
	}
	if flow.Step() != StepSelectCell {
		t.Errorf("expected StepSelectCell after rejection, got %d", flow.Step())
	}
}

func TestDigRoomFlow_NonRockCellRejected(t *testing.T) {
	tests := []struct {
		name     string
		cellType world.CellType
	}{
		{"RoomFloor", world.RoomFloor},
		{"CorridorFloor", world.CorridorFloor},
		{"Entrance", world.Entrance},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flow := NewDigRoomFlow()
			err := flow.TrySelectCell(5, 3, tt.cellType)
			if err == nil {
				t.Fatalf("expected error for %s cell", tt.name)
			}
			if flow.Step() != StepSelectCell {
				t.Errorf("expected StepSelectCell after rejection, got %d", flow.Step())
			}
		})
	}
}

func TestDigRoomFlow_ElementSelectionCompletesFlow(t *testing.T) {
	flow := NewDigRoomFlow()
	_ = flow.TrySelectCell(10, 7, world.Rock)

	_, err := flow.SelectElement(types.Earth, testRoomTypeRegistry())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After element selection, flow should be complete (caller sets ModeNormal).
	if flow.Step() != StepComplete {
		t.Errorf("expected StepComplete after element selection, got %d", flow.Step())
	}
}

func TestDigRoomFlow_Cancel(t *testing.T) {
	flow := NewDigRoomFlow()
	_ = flow.TrySelectCell(5, 3, world.Rock)

	if flow.Step() != StepSelectElement {
		t.Fatalf("expected StepSelectElement, got %d", flow.Step())
	}

	flow.Cancel()

	if flow.Step() != StepSelectCell {
		t.Errorf("expected StepSelectCell after cancel, got %d", flow.Step())
	}
}

func TestDigRoomFlow_AllElements(t *testing.T) {
	elements := []struct {
		elem       types.Element
		wantTypeID string
	}{
		{types.Wood, "wood_room"},
		{types.Fire, "fire_room"},
		{types.Earth, "earth_room"},
		{types.Metal, "metal_room"},
		{types.Water, "water_room"},
	}
	reg := testRoomTypeRegistry()

	for _, tt := range elements {
		t.Run(tt.elem.String(), func(t *testing.T) {
			flow := NewDigRoomFlow()
			_ = flow.TrySelectCell(1, 1, world.Rock)
			action, err := flow.SelectElement(tt.elem, reg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if action.RoomTypeID != tt.wantTypeID {
				t.Errorf("expected %s, got %s", tt.wantTypeID, action.RoomTypeID)
			}
		})
	}
}
