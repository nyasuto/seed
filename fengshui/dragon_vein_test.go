package fengshui

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// helper to create a small cave with rooms and corridors for testing.
// Layout (10x10):
//
//	Room 1 at (1,1) 3x3 with entrance at (3,2) facing East
//	Room 2 at (6,1) 3x3 with entrance at (6,2) facing West
//	Corridor connecting them via entrance cells
func setupTestCave(t *testing.T) *world.Cave {
	t.Helper()
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1: top-left area
	_, err = cave.AddRoom("dragon_den", types.Pos{X: 1, Y: 1}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	// Room 2: top-right area
	_, err = cave.AddRoom("chi_chamber", types.Pos{X: 6, Y: 1}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 6, Y: 2}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	// Connect rooms
	_, err = cave.ConnectRooms(1, 2)
	if err != nil {
		t.Fatalf("ConnectRooms: %v", err)
	}

	return cave
}

func TestBuildDragonVein_ReachesRooms(t *testing.T) {
	cave := setupTestCave(t)

	// Source from the corridor between the rooms (entrance outside cell)
	sourcePos := types.Pos{X: 4, Y: 2}
	dv, err := BuildDragonVein(cave, sourcePos, types.Wood, 1.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	if dv.SourcePos != sourcePos {
		t.Errorf("SourcePos = %v, want %v", dv.SourcePos, sourcePos)
	}
	if dv.Element != types.Wood {
		t.Errorf("Element = %v, want Wood", dv.Element)
	}
	if dv.FlowRate != 1.0 {
		t.Errorf("FlowRate = %v, want 1.0", dv.FlowRate)
	}

	rooms := dv.RoomsOnPath(cave)
	if len(rooms) != 2 {
		t.Fatalf("RoomsOnPath = %v, want 2 rooms", rooms)
	}

	roomSet := make(map[int]bool)
	for _, id := range rooms {
		roomSet[id] = true
	}
	if !roomSet[1] || !roomSet[2] {
		t.Errorf("expected rooms 1 and 2 on path, got %v", rooms)
	}
}

func TestBuildDragonVein_UnreachableError(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Source on rock cell — unreachable
	_, err = BuildDragonVein(cave, types.Pos{X: 5, Y: 5}, types.Fire, 1.0)
	if err == nil {
		t.Fatal("expected error for rock source, got nil")
	}
	if !errors.Is(err, ErrUnreachable) {
		t.Errorf("expected ErrUnreachable, got %v", err)
	}
}

func TestBuildDragonVein_OutOfBounds(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	_, err = BuildDragonVein(cave, types.Pos{X: -1, Y: 0}, types.Water, 1.0)
	if err == nil {
		t.Fatal("expected error for out-of-bounds source, got nil")
	}
}

func TestRebuildDragonVein_PathExtends(t *testing.T) {
	cave := setupTestCave(t)

	// Build initial vein from corridor
	sourcePos := types.Pos{X: 4, Y: 2}
	original, err := BuildDragonVein(cave, sourcePos, types.Metal, 2.5)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}
	original.ID = 42

	originalPathLen := len(original.Path)
	originalRooms := original.RoomsOnPath(cave)

	// Add a third room and connect it
	_, err = cave.AddRoom("trap_room", types.Pos{X: 1, Y: 6}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 7}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 3: %v", err)
	}
	_, err = cave.ConnectRooms(1, 3)
	if err != nil {
		t.Fatalf("ConnectRooms 1-3: %v", err)
	}

	// Rebuild the vein
	rebuilt, err := RebuildDragonVein(cave, original)
	if err != nil {
		t.Fatalf("RebuildDragonVein: %v", err)
	}

	// ID preserved
	if rebuilt.ID != 42 {
		t.Errorf("rebuilt ID = %d, want 42", rebuilt.ID)
	}
	// Element and FlowRate preserved
	if rebuilt.Element != types.Metal {
		t.Errorf("rebuilt Element = %v, want Metal", rebuilt.Element)
	}
	if rebuilt.FlowRate != 2.5 {
		t.Errorf("rebuilt FlowRate = %v, want 2.5", rebuilt.FlowRate)
	}

	// Path should be longer now
	if len(rebuilt.Path) <= originalPathLen {
		t.Errorf("rebuilt path len = %d, should be > original %d", len(rebuilt.Path), originalPathLen)
	}

	// Should now reach 3 rooms
	rebuiltRooms := rebuilt.RoomsOnPath(cave)
	if len(rebuiltRooms) <= len(originalRooms) {
		t.Errorf("rebuilt rooms = %v, should have more rooms than original %v", rebuiltRooms, originalRooms)
	}

	roomSet := make(map[int]bool)
	for _, id := range rebuiltRooms {
		roomSet[id] = true
	}
	if !roomSet[3] {
		t.Errorf("expected room 3 on rebuilt path, got %v", rebuiltRooms)
	}
}

func TestBuildDragonVein_DisconnectedRoomNotReached(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1 with entrance
	_, err = cave.AddRoom("dragon_den", types.Pos{X: 1, Y: 1}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	// Room 2 far away, no corridor connecting them
	_, err = cave.AddRoom("chi_chamber", types.Pos{X: 7, Y: 7}, 2, 2, []world.RoomEntrance{
		{Pos: types.Pos{X: 7, Y: 8}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	// Build vein from room 1's entrance area
	dv, err := BuildDragonVein(cave, types.Pos{X: 1, Y: 1}, types.Earth, 1.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	rooms := dv.RoomsOnPath(cave)
	for _, id := range rooms {
		if id == 2 {
			t.Errorf("disconnected room 2 should not be on path, got rooms %v", rooms)
		}
	}
	// Room 1 should be reachable (source is inside it)
	found := false
	for _, id := range rooms {
		if id == 1 {
			found = true
		}
	}
	if !found {
		t.Errorf("room 1 should be on path, got rooms %v", rooms)
	}
}
