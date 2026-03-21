package world

import (
	"errors"
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestPlaceRoom_RejectsHardRock(t *testing.T) {
	grid, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}
	// Place HardRock in the middle of where the room would go
	_ = grid.Set(types.Pos{X: 2, Y: 2}, Cell{Type: HardRock})

	room := &Room{
		ID:     1,
		TypeID: "wood_room",
		Pos:    types.Pos{X: 1, Y: 1},
		Width:  3,
		Height: 3,
		Level:  1,
	}

	if CanPlaceRoom(grid, room) {
		t.Error("CanPlaceRoom should return false when HardRock is under room")
	}

	err = PlaceRoom(grid, room)
	if err == nil {
		t.Fatal("PlaceRoom should return error for HardRock")
	}
	if !errors.Is(err, ErrRoomOnImpassable) {
		t.Errorf("expected ErrRoomOnImpassable, got: %v", err)
	}
}

func TestPlaceRoom_RejectsWater(t *testing.T) {
	grid, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}
	_ = grid.Set(types.Pos{X: 3, Y: 3}, Cell{Type: Water})

	room := &Room{
		ID:     1,
		TypeID: "wood_room",
		Pos:    types.Pos{X: 2, Y: 2},
		Width:  3,
		Height: 3,
		Level:  1,
	}

	if CanPlaceRoom(grid, room) {
		t.Error("CanPlaceRoom should return false when Water is under room")
	}

	err = PlaceRoom(grid, room)
	if err == nil {
		t.Fatal("PlaceRoom should return error for Water")
	}
	if !errors.Is(err, ErrRoomOnImpassable) {
		t.Errorf("expected ErrRoomOnImpassable, got: %v", err)
	}
}

func TestBuildCorridor_AvoidsHardRock(t *testing.T) {
	// Create a grid with a wall of HardRock that forces a detour
	grid, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	// Place two rooms
	roomA := &Room{ID: 1, TypeID: "wood_room", Pos: types.Pos{X: 0, Y: 4}, Width: 2, Height: 2, Level: 1}
	roomB := &Room{ID: 2, TypeID: "wood_room", Pos: types.Pos{X: 8, Y: 4}, Width: 2, Height: 2, Level: 1}
	if err := PlaceRoom(grid, roomA); err != nil {
		t.Fatalf("PlaceRoom A: %v", err)
	}
	if err := PlaceRoom(grid, roomB); err != nil {
		t.Fatalf("PlaceRoom B: %v", err)
	}

	// Place a vertical wall of HardRock at x=4, y=3..6 (blocks direct path)
	for y := 3; y <= 6; y++ {
		_ = grid.Set(types.Pos{X: 4, Y: y}, Cell{Type: HardRock})
	}

	from := types.Pos{X: 2, Y: 4} // entrance area of room A
	to := types.Pos{X: 7, Y: 4}   // entrance area of room B

	corridor, err := BuildCorridor(grid, from, to, 1, roomA.ID, roomB.ID)
	if err != nil {
		t.Fatalf("BuildCorridor should find a path around HardRock: %v", err)
	}

	// Verify the path does not cross any HardRock cell
	for _, pos := range corridor.Path {
		cell, _ := grid.At(pos)
		if cell.Type == HardRock {
			t.Errorf("corridor path passes through HardRock at (%d,%d)", pos.X, pos.Y)
		}
	}
}

func TestBuildCorridor_AvoidsWater(t *testing.T) {
	grid, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	roomA := &Room{ID: 1, TypeID: "wood_room", Pos: types.Pos{X: 0, Y: 4}, Width: 2, Height: 2, Level: 1}
	roomB := &Room{ID: 2, TypeID: "wood_room", Pos: types.Pos{X: 8, Y: 4}, Width: 2, Height: 2, Level: 1}
	if err := PlaceRoom(grid, roomA); err != nil {
		t.Fatalf("PlaceRoom A: %v", err)
	}
	if err := PlaceRoom(grid, roomB); err != nil {
		t.Fatalf("PlaceRoom B: %v", err)
	}

	// Place a vertical wall of Water at x=4, y=3..6
	for y := 3; y <= 6; y++ {
		_ = grid.Set(types.Pos{X: 4, Y: y}, Cell{Type: Water})
	}

	from := types.Pos{X: 2, Y: 4}
	to := types.Pos{X: 7, Y: 4}

	corridor, err := BuildCorridor(grid, from, to, 1, roomA.ID, roomB.ID)
	if err != nil {
		t.Fatalf("BuildCorridor should find a path around Water: %v", err)
	}

	for _, pos := range corridor.Path {
		cell, _ := grid.At(pos)
		if cell.Type == Water {
			t.Errorf("corridor path passes through Water at (%d,%d)", pos.X, pos.Y)
		}
	}
}

func TestBuildCorridor_UnreachableDueToImpassable(t *testing.T) {
	// Surround destination with HardRock so no path exists
	grid, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	roomA := &Room{ID: 1, TypeID: "wood_room", Pos: types.Pos{X: 0, Y: 0}, Width: 2, Height: 2, Level: 1}
	roomB := &Room{ID: 2, TypeID: "wood_room", Pos: types.Pos{X: 5, Y: 5}, Width: 2, Height: 2, Level: 1}
	if err := PlaceRoom(grid, roomA); err != nil {
		t.Fatalf("PlaceRoom A: %v", err)
	}
	if err := PlaceRoom(grid, roomB); err != nil {
		t.Fatalf("PlaceRoom B: %v", err)
	}

	// Surround room B with HardRock (ring at distance 1 from the room)
	for x := 3; x <= 8; x++ {
		for y := 3; y <= 8; y++ {
			pos := types.Pos{X: x, Y: y}
			cell, _ := grid.At(pos)
			// Only set HardRock on Rock cells (don't overwrite room B)
			if cell.Type == Rock {
				_ = grid.Set(pos, Cell{Type: HardRock})
			}
		}
	}

	from := types.Pos{X: 1, Y: 1}
	to := types.Pos{X: 5, Y: 5}

	_, err = BuildCorridor(grid, from, to, 1, roomA.ID, roomB.ID)
	if err == nil {
		t.Fatal("BuildCorridor should return error when path is blocked by impassable terrain")
	}
	if !errors.Is(err, ErrNoPath) {
		t.Errorf("expected ErrNoPath, got: %v", err)
	}
}

func TestCoreHP_Initialization(t *testing.T) {
	reg, err := LoadDefaultRoomTypes()
	if err != nil {
		t.Fatalf("LoadDefaultRoomTypes: %v", err)
	}

	dragonHole, err := reg.Get("dragon_hole")
	if err != nil {
		t.Fatalf("Get dragon_hole: %v", err)
	}

	if dragonHole.BaseCoreHP == 0 {
		t.Fatal("dragon_hole BaseCoreHP should be non-zero")
	}

	tests := []struct {
		name  string
		level int
		want  int
	}{
		{"level 0", 0, 0},
		{"level 1", 1, 100},
		{"level 2", 2, 200},
		{"level 5", 5, 500},
		{"negative level", -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dragonHole.CoreHPAtLevel(tt.level)
			if got != tt.want {
				t.Errorf("CoreHPAtLevel(%d) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}

	// Non-core rooms should always return 0
	chiChamber, err := reg.Get("chi_chamber")
	if err != nil {
		t.Fatalf("Get chi_chamber: %v", err)
	}
	if chiChamber.CoreHPAtLevel(1) != 0 {
		t.Error("non-core room CoreHPAtLevel should return 0")
	}
}

func TestCoreHP_RoomField(t *testing.T) {
	reg, err := LoadDefaultRoomTypes()
	if err != nil {
		t.Fatalf("LoadDefaultRoomTypes: %v", err)
	}

	dragonHole, err := reg.Get("dragon_hole")
	if err != nil {
		t.Fatalf("Get dragon_hole: %v", err)
	}

	// Simulate initializing a Room with CoreHP from its type
	room := &Room{
		ID:     1,
		TypeID: "dragon_hole",
		Pos:    types.Pos{X: 5, Y: 5},
		Width:  3,
		Height: 3,
		Level:  1,
		CoreHP: dragonHole.CoreHPAtLevel(1),
	}

	if room.CoreHP != 100 {
		t.Errorf("dragon_hole room CoreHP = %d, want 100", room.CoreHP)
	}

	// Non-core room should have CoreHP 0
	normalRoom := &Room{
		ID:     2,
		TypeID: "wood_room",
		Pos:    types.Pos{X: 0, Y: 0},
		Width:  3,
		Height: 3,
		Level:  1,
		CoreHP: 0,
	}

	if normalRoom.CoreHP != 0 {
		t.Errorf("non-core room CoreHP = %d, want 0", normalRoom.CoreHP)
	}
}
