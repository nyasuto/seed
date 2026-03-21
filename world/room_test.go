package world

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestCanPlaceRoom_Valid(t *testing.T) {
	g, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	room := &Room{
		ID:     1,
		TypeID: "test",
		Pos:    types.Pos{X: 2, Y: 3},
		Width:  3,
		Height: 2,
	}

	if !CanPlaceRoom(g, room) {
		t.Error("CanPlaceRoom should return true for valid placement on empty grid")
	}
}

func TestCanPlaceRoom_OutOfBounds(t *testing.T) {
	g, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	tests := []struct {
		name string
		room *Room
	}{
		{
			"extends beyond right edge",
			&Room{ID: 1, Pos: types.Pos{X: 8, Y: 0}, Width: 3, Height: 2},
		},
		{
			"extends beyond bottom edge",
			&Room{ID: 1, Pos: types.Pos{X: 0, Y: 9}, Width: 2, Height: 2},
		},
		{
			"negative position",
			&Room{ID: 1, Pos: types.Pos{X: -1, Y: 0}, Width: 2, Height: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if CanPlaceRoom(g, tt.room) {
				t.Error("CanPlaceRoom should return false for out-of-bounds placement")
			}
		})
	}
}

func TestCanPlaceRoom_Overlap(t *testing.T) {
	g, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	// Place a room first
	room1 := &Room{ID: 1, Pos: types.Pos{X: 2, Y: 2}, Width: 3, Height: 3}
	if err := PlaceRoom(g, room1); err != nil {
		t.Fatalf("PlaceRoom: %v", err)
	}

	// Try to place an overlapping room
	room2 := &Room{ID: 2, Pos: types.Pos{X: 4, Y: 4}, Width: 3, Height: 3}
	if CanPlaceRoom(g, room2) {
		t.Error("CanPlaceRoom should return false when overlapping with existing room")
	}
}

func TestPlaceRoom_Success(t *testing.T) {
	g, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	room := &Room{
		ID:     1,
		TypeID: "test",
		Pos:    types.Pos{X: 1, Y: 1},
		Width:  3,
		Height: 2,
		Entrances: []RoomEntrance{
			{Pos: types.Pos{X: 2, Y: 1}, Dir: types.North},
		},
	}

	if err := PlaceRoom(g, room); err != nil {
		t.Fatalf("PlaceRoom: %v", err)
	}

	// Check all room cells are RoomFloor with correct RoomID
	for y := 1; y <= 2; y++ {
		for x := 1; x <= 3; x++ {
			pos := types.Pos{X: x, Y: y}
			cell, err := g.At(pos)
			if err != nil {
				t.Fatalf("At(%v): %v", pos, err)
			}
			if pos.X == 2 && pos.Y == 1 {
				// Entrance cell
				if cell.Type != Entrance {
					t.Errorf("At(%v).Type = %v, want Entrance", pos, cell.Type)
				}
			} else {
				if cell.Type != RoomFloor {
					t.Errorf("At(%v).Type = %v, want RoomFloor", pos, cell.Type)
				}
			}
			if cell.RoomID != 1 {
				t.Errorf("At(%v).RoomID = %d, want 1", pos, cell.RoomID)
			}
		}
	}

	// Check surrounding cells are still Rock
	cell, _ := g.At(types.Pos{X: 0, Y: 0})
	if cell.Type != Rock {
		t.Errorf("surrounding cell should remain Rock, got %v", cell.Type)
	}
}

func TestPlaceRoom_OutOfBounds(t *testing.T) {
	g, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	room := &Room{ID: 1, Pos: types.Pos{X: 9, Y: 9}, Width: 3, Height: 3}
	err = PlaceRoom(g, room)
	if err == nil {
		t.Fatal("PlaceRoom should return error for out-of-bounds room")
	}
	if !errors.Is(err, ErrRoomOutOfBounds) {
		t.Errorf("error = %v, want ErrRoomOutOfBounds", err)
	}
}

func TestPlaceRoom_Overlap(t *testing.T) {
	g, err := NewGrid(10, 10)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	room1 := &Room{ID: 1, Pos: types.Pos{X: 2, Y: 2}, Width: 3, Height: 3}
	if err := PlaceRoom(g, room1); err != nil {
		t.Fatalf("PlaceRoom first room: %v", err)
	}

	room2 := &Room{ID: 2, Pos: types.Pos{X: 3, Y: 3}, Width: 3, Height: 3}
	err = PlaceRoom(g, room2)
	if err == nil {
		t.Fatal("PlaceRoom should return error for overlapping room")
	}
	if !errors.Is(err, ErrRoomOverlap) {
		t.Errorf("error = %v, want ErrRoomOverlap", err)
	}
}

func TestRoomTypeRegistry_LoadJSON(t *testing.T) {
	reg, err := LoadDefaultRoomTypes()
	if err != nil {
		t.Fatalf("LoadRoomTypesJSON: %v", err)
	}

	if reg.Len() != 6 {
		t.Errorf("registry has %d types, want 6", reg.Len())
	}

	// Verify a specific room type
	rt, err := reg.Get("dragon_hole")
	if err != nil {
		t.Fatalf("Get dragon_hole: %v", err)
	}
	if rt.Name != "龍穴" {
		t.Errorf("dragon_hole Name = %q, want %q", rt.Name, "龍穴")
	}
	if rt.Element != types.Earth {
		t.Errorf("dragon_hole Element = %v, want Earth", rt.Element)
	}
	if rt.BaseChiCapacity != 100 {
		t.Errorf("dragon_hole BaseChiCapacity = %d, want 100", rt.BaseChiCapacity)
	}
}

func TestRoomType_MaxBeasts(t *testing.T) {
	reg, err := LoadDefaultRoomTypes()
	if err != nil {
		t.Fatalf("LoadDefaultRoomTypes: %v", err)
	}

	expected := map[string]int{
		"dragon_hole":   1,
		"chi_chamber":   0,
		"senju_room":    3,
		"trap_room":     2,
		"recovery_room": 0,
		"storage":       0,
	}

	for id, want := range expected {
		rt, err := reg.Get(id)
		if err != nil {
			t.Fatalf("Get(%q): %v", id, err)
		}
		if rt.MaxBeasts != want {
			t.Errorf("%s MaxBeasts = %d, want %d", id, rt.MaxBeasts, want)
		}
	}
}

func TestRoom_BeastCount(t *testing.T) {
	room := &Room{ID: 1}
	if room.BeastCount() != 0 {
		t.Errorf("empty room BeastCount = %d, want 0", room.BeastCount())
	}

	room.BeastIDs = []int{10, 20}
	if room.BeastCount() != 2 {
		t.Errorf("BeastCount = %d, want 2", room.BeastCount())
	}
}

func TestRoom_HasBeastCapacity(t *testing.T) {
	senjuType := RoomType{ID: "senju_room", MaxBeasts: 3}
	storageType := RoomType{ID: "storage", MaxBeasts: 0}

	tests := []struct {
		name     string
		room     *Room
		roomType RoomType
		want     bool
	}{
		{
			"empty room with capacity",
			&Room{ID: 1},
			senjuType,
			true,
		},
		{
			"room with space remaining",
			&Room{ID: 1, BeastIDs: []int{10, 20}},
			senjuType,
			true,
		},
		{
			"room at capacity",
			&Room{ID: 1, BeastIDs: []int{10, 20, 30}},
			senjuType,
			false,
		},
		{
			"room type disallows beasts",
			&Room{ID: 1},
			storageType,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.room.HasBeastCapacity(tt.roomType)
			if got != tt.want {
				t.Errorf("HasBeastCapacity = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoomTypeRegistry_LoadJSON_AllTypes(t *testing.T) {
	reg, err := LoadDefaultRoomTypes()
	if err != nil {
		t.Fatalf("LoadRoomTypesJSON: %v", err)
	}

	expectedIDs := []string{
		"dragon_hole", "chi_chamber", "senju_room",
		"trap_room", "recovery_room", "storage",
	}
	for _, id := range expectedIDs {
		if _, err := reg.Get(id); err != nil {
			t.Errorf("expected room type %q to be registered", id)
		}
	}
}
