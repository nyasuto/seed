package view

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

func newTestCaveWithRoom(t *testing.T) (*world.Cave, *world.RoomTypeRegistry) {
	t.Helper()

	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	registry := world.NewRoomTypeRegistry()
	if err := registry.Register(world.RoomType{
		ID:      "fire_room",
		Name:    "Fire Room",
		Element: types.Fire,
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, err = cave.AddRoom("fire_room", types.Pos{X: 2, Y: 3}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	return cave, registry
}

func TestBuildTooltipInfo_RockCell(t *testing.T) {
	cave, registry := newTestCaveWithRoom(t)

	info := BuildTooltipInfo(cave, registry, 0, 0)
	if len(info.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %v", len(info.Lines), info.Lines)
	}
	want := "(0, 0) Rock"
	if info.Lines[0] != want {
		t.Errorf("line[0] = %q, want %q", info.Lines[0], want)
	}
}

func TestBuildTooltipInfo_RoomCell(t *testing.T) {
	cave, registry := newTestCaveWithRoom(t)

	// Cell (3, 4) is inside the room placed at (2,3) with size 3x3.
	info := BuildTooltipInfo(cave, registry, 3, 4)
	if len(info.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(info.Lines), info.Lines)
	}

	wantLine0 := "(3, 4) RoomFloor"
	if info.Lines[0] != wantLine0 {
		t.Errorf("line[0] = %q, want %q", info.Lines[0], wantLine0)
	}

	wantLine1 := "Room #1  Fire Lv1"
	if info.Lines[1] != wantLine1 {
		t.Errorf("line[1] = %q, want %q", info.Lines[1], wantLine1)
	}
}

func TestBuildTooltipInfo_OutOfBounds(t *testing.T) {
	cave, registry := newTestCaveWithRoom(t)

	info := BuildTooltipInfo(cave, registry, 99, 99)
	if len(info.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %v", len(info.Lines), info.Lines)
	}
	if info.Lines[0] != "(invalid)" {
		t.Errorf("line[0] = %q, want %q", info.Lines[0], "(invalid)")
	}
}

func TestBuildTooltipInfo_HardRock(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Manually set a cell to HardRock.
	_ = cave.Grid.Set(types.Pos{X: 5, Y: 5}, world.Cell{Type: world.HardRock})

	registry := world.NewRoomTypeRegistry()
	info := BuildTooltipInfo(cave, registry, 5, 5)

	want := "(5, 5) HardRock"
	if info.Lines[0] != want {
		t.Errorf("line[0] = %q, want %q", info.Lines[0], want)
	}
}
