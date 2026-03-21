package world

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestBuildCorridor_StraightLine(t *testing.T) {
	// Two rooms side by side with a gap of 2 cells between them
	// Room1 at (1,1) 3x3, Room2 at (6,1) 3x3
	// Corridor from (4,2) to (5,2) — straight horizontal line through rock
	grid, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	room1 := &Room{ID: 1, TypeID: "test", Pos: types.Pos{X: 1, Y: 1}, Width: 3, Height: 3,
		Entrances: []RoomEntrance{{Pos: types.Pos{X: 4, Y: 2}, Dir: types.East}},
	}
	room2 := &Room{ID: 2, TypeID: "test", Pos: types.Pos{X: 6, Y: 1}, Width: 3, Height: 3,
		Entrances: []RoomEntrance{{Pos: types.Pos{X: 5, Y: 2}, Dir: types.West}},
	}

	if err := PlaceRoom(grid, room1); err != nil {
		t.Fatalf("PlaceRoom room1: %v", err)
	}
	if err := PlaceRoom(grid, room2); err != nil {
		t.Fatalf("PlaceRoom room2: %v", err)
	}

	fromPos := types.Pos{X: 4, Y: 2} // entrance of room1
	toPos := types.Pos{X: 5, Y: 2}   // entrance of room2

	corridor, err := BuildCorridor(grid, fromPos, toPos, 1, room1.ID, room2.ID)
	if err != nil {
		t.Fatalf("BuildCorridor: %v", err)
	}

	if corridor.ID != 1 {
		t.Errorf("corridor ID = %d, want 1", corridor.ID)
	}
	if corridor.FromRoomID != 1 || corridor.ToRoomID != 2 {
		t.Errorf("corridor rooms = (%d, %d), want (1, 2)", corridor.FromRoomID, corridor.ToRoomID)
	}
	if len(corridor.Path) != 2 {
		t.Errorf("corridor path length = %d, want 2", len(corridor.Path))
	}

	// Verify the path goes from entrance to entrance
	if corridor.Path[0] != fromPos {
		t.Errorf("path start = %v, want %v", corridor.Path[0], fromPos)
	}
	if corridor.Path[len(corridor.Path)-1] != toPos {
		t.Errorf("path end = %v, want %v", corridor.Path[len(corridor.Path)-1], toPos)
	}
}

func TestBuildCorridor_AroundObstacle(t *testing.T) {
	// Room1 at (0,0) 2x2, Room3 (obstacle) at (3,0) 2x2, Room2 at (6,0) 2x2
	// Corridor must go around room3
	grid, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	room1 := &Room{ID: 1, TypeID: "test", Pos: types.Pos{X: 0, Y: 0}, Width: 2, Height: 2}
	room3 := &Room{ID: 3, TypeID: "test", Pos: types.Pos{X: 3, Y: 0}, Width: 2, Height: 2}
	room2 := &Room{ID: 2, TypeID: "test", Pos: types.Pos{X: 6, Y: 0}, Width: 2, Height: 2}

	for _, r := range []*Room{room1, room3, room2} {
		if err := PlaceRoom(grid, r); err != nil {
			t.Fatalf("PlaceRoom room%d: %v", r.ID, err)
		}
	}

	fromPos := types.Pos{X: 2, Y: 0} // just east of room1
	toPos := types.Pos{X: 5, Y: 0}   // just west of room2

	corridor, err := BuildCorridor(grid, fromPos, toPos, 1, room1.ID, room2.ID)
	if err != nil {
		t.Fatalf("BuildCorridor: %v", err)
	}

	// The path must not pass through room3's cells
	for _, pos := range corridor.Path {
		cell, _ := grid.At(pos)
		if cell.RoomID == 3 {
			t.Errorf("corridor passes through obstacle room3 at %v", pos)
		}
	}

	// Path should be longer than the direct 4-cell distance due to detour
	if len(corridor.Path) <= 4 {
		t.Errorf("expected detour path length > 4, got %d", len(corridor.Path))
	}
}

func TestBuildCorridor_Unreachable(t *testing.T) {
	// Create a small grid where a wall of rooms blocks any path
	grid, err := NewGrid(5, 3)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	// Fill the middle column with a room (blocking wall)
	blocker := &Room{ID: 3, TypeID: "test", Pos: types.Pos{X: 2, Y: 0}, Width: 1, Height: 3}
	if err := PlaceRoom(grid, blocker); err != nil {
		t.Fatalf("PlaceRoom blocker: %v", err)
	}

	fromPos := types.Pos{X: 0, Y: 1}
	toPos := types.Pos{X: 4, Y: 1}

	_, err = BuildCorridor(grid, fromPos, toPos, 1, 0, 0)
	if err == nil {
		t.Fatal("expected error for unreachable path, got nil")
	}
	if !errors.Is(err, ErrNoPath) {
		t.Errorf("expected ErrNoPath, got: %v", err)
	}
}

func TestBuildCorridor_OutOfBounds(t *testing.T) {
	grid, err := NewGrid(5, 5)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	_, err = BuildCorridor(grid, types.Pos{X: -1, Y: 0}, types.Pos{X: 4, Y: 4}, 1, 0, 0)
	if !errors.Is(err, ErrOutOfBounds) {
		t.Errorf("expected ErrOutOfBounds for fromPos, got: %v", err)
	}

	_, err = BuildCorridor(grid, types.Pos{X: 0, Y: 0}, types.Pos{X: 10, Y: 10}, 1, 0, 0)
	if !errors.Is(err, ErrOutOfBounds) {
		t.Errorf("expected ErrOutOfBounds for toPos, got: %v", err)
	}
}

func TestBuildCorridor_TraversesExistingCorridor(t *testing.T) {
	// Build two corridors that share some cells
	grid, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	// First corridor: horizontal
	_, err = BuildCorridor(grid, types.Pos{X: 0, Y: 5}, types.Pos{X: 9, Y: 5}, 1, 0, 0)
	if err != nil {
		t.Fatalf("first BuildCorridor: %v", err)
	}

	// Second corridor: vertical, crossing the first
	c2, err := BuildCorridor(grid, types.Pos{X: 5, Y: 0}, types.Pos{X: 5, Y: 9}, 2, 0, 0)
	if err != nil {
		t.Fatalf("second BuildCorridor: %v", err)
	}

	// The second corridor should pass through (5,5) which is already a corridor cell
	foundIntersection := false
	for _, pos := range c2.Path {
		if pos.X == 5 && pos.Y == 5 {
			foundIntersection = true
			break
		}
	}
	if !foundIntersection {
		t.Error("expected second corridor to pass through intersection at (5,5)")
	}
}
