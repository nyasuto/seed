package world

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestCaveSerialization_WithRoomsAndCorridors(t *testing.T) {
	cave, err := NewCave(16, 16)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Add two rooms
	entrances1 := []RoomEntrance{
		{Pos: types.Pos{X: 4, Y: 2}, Dir: types.East},
	}
	room1, err := cave.AddRoom("dragon_vein", types.Pos{X: 1, Y: 1}, 4, 3, entrances1)
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	entrances2 := []RoomEntrance{
		{Pos: types.Pos{X: 8, Y: 7}, Dir: types.West},
	}
	room2, err := cave.AddRoom("chi_chamber", types.Pos{X: 8, Y: 6}, 4, 3, entrances2)
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	// Connect rooms
	_, err = cave.ConnectRooms(room1.ID, room2.ID)
	if err != nil {
		t.Fatalf("ConnectRooms: %v", err)
	}

	// Serialize
	data, err := cave.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	// Deserialize
	restored, err := UnmarshalCave(data)
	if err != nil {
		t.Fatalf("UnmarshalCave: %v", err)
	}

	// Verify grid dimensions
	if restored.Grid.Width != cave.Grid.Width || restored.Grid.Height != cave.Grid.Height {
		t.Errorf("grid dimensions: got %dx%d, want %dx%d",
			restored.Grid.Width, restored.Grid.Height, cave.Grid.Width, cave.Grid.Height)
	}

	// Verify all cells match
	for y := 0; y < cave.Grid.Height; y++ {
		for x := 0; x < cave.Grid.Width; x++ {
			pos := types.Pos{X: x, Y: y}
			orig, _ := cave.Grid.At(pos)
			rest, _ := restored.Grid.At(pos)
			if orig != rest {
				t.Errorf("cell at (%d,%d): got %+v, want %+v", x, y, rest, orig)
			}
		}
	}

	// Verify rooms
	if len(restored.Rooms) != len(cave.Rooms) {
		t.Fatalf("rooms count: got %d, want %d", len(restored.Rooms), len(cave.Rooms))
	}
	for i, orig := range cave.Rooms {
		rest := restored.Rooms[i]
		if rest.ID != orig.ID || rest.TypeID != orig.TypeID ||
			rest.Pos != orig.Pos || rest.Width != orig.Width ||
			rest.Height != orig.Height || rest.Level != orig.Level {
			t.Errorf("room %d mismatch: got %+v, want %+v", i, rest, orig)
		}
		if len(rest.Entrances) != len(orig.Entrances) {
			t.Errorf("room %d entrances count: got %d, want %d", i, len(rest.Entrances), len(orig.Entrances))
			continue
		}
		for j, oe := range orig.Entrances {
			re := rest.Entrances[j]
			if re.Pos != oe.Pos || re.Dir != oe.Dir {
				t.Errorf("room %d entrance %d: got %+v, want %+v", i, j, re, oe)
			}
		}
	}

	// Verify corridors
	if len(restored.Corridors) != len(cave.Corridors) {
		t.Fatalf("corridors count: got %d, want %d", len(restored.Corridors), len(cave.Corridors))
	}
	for i, orig := range cave.Corridors {
		rest := restored.Corridors[i]
		if rest.ID != orig.ID || rest.FromRoomID != orig.FromRoomID || rest.ToRoomID != orig.ToRoomID {
			t.Errorf("corridor %d mismatch: got ID=%d From=%d To=%d, want ID=%d From=%d To=%d",
				i, rest.ID, rest.FromRoomID, rest.ToRoomID, orig.ID, orig.FromRoomID, orig.ToRoomID)
		}
		if len(rest.Path) != len(orig.Path) {
			t.Errorf("corridor %d path length: got %d, want %d", i, len(rest.Path), len(orig.Path))
			continue
		}
		for j, op := range orig.Path {
			if rest.Path[j] != op {
				t.Errorf("corridor %d path[%d]: got %v, want %v", i, j, rest.Path[j], op)
			}
		}
	}

	// Verify next IDs are preserved (add another room to check auto-increment)
	entrances3 := []RoomEntrance{
		{Pos: types.Pos{X: 1, Y: 12}, Dir: types.North},
	}
	room3, err := restored.AddRoom("trap_room", types.Pos{X: 1, Y: 11}, 3, 3, entrances3)
	if err != nil {
		t.Fatalf("AddRoom on restored cave: %v", err)
	}
	if room3.ID != cave.nextRoomID {
		t.Errorf("next room ID: got %d, want %d", room3.ID, cave.nextRoomID)
	}
}

func TestCaveSerialization_EmptyCave(t *testing.T) {
	cave, err := NewCave(8, 8)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	data, err := cave.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	restored, err := UnmarshalCave(data)
	if err != nil {
		t.Fatalf("UnmarshalCave: %v", err)
	}

	if restored.Grid.Width != 8 || restored.Grid.Height != 8 {
		t.Errorf("grid dimensions: got %dx%d, want 8x8", restored.Grid.Width, restored.Grid.Height)
	}
	if len(restored.Rooms) != 0 {
		t.Errorf("rooms: got %d, want 0", len(restored.Rooms))
	}
	if len(restored.Corridors) != 0 {
		t.Errorf("corridors: got %d, want 0", len(restored.Corridors))
	}

	// All cells should be Rock
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			cell, _ := restored.Grid.At(types.Pos{X: x, Y: y})
			if cell.Type != Rock || cell.RoomID != 0 {
				t.Errorf("cell at (%d,%d): got %+v, want Rock with RoomID 0", x, y, cell)
			}
		}
	}
}

func TestUnmarshalCave_InvalidJSON(t *testing.T) {
	_, err := UnmarshalCave([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
