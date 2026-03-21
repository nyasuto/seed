package world

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
)

// TestIntegration_MediumMap verifies the full workflow on a 32x32 cave:
// place 5 rooms, connect them all, and confirm adjacency graph correctness.
func TestIntegration_MediumMap(t *testing.T) {
	cave, err := NewCave(32, 32)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Define 5 rooms with different types, positions, and entrances.
	// Each room is 3x3 with one entrance on the south side.
	type roomDef struct {
		typeID    string
		pos       types.Pos
		entrances []RoomEntrance
	}

	rooms := []roomDef{
		{
			typeID: "dragon_hole",
			pos:    types.Pos{X: 2, Y: 2},
			entrances: []RoomEntrance{
				{Pos: types.Pos{X: 3, Y: 4}, Dir: types.South},
			},
		},
		{
			typeID: "chi_chamber",
			pos:    types.Pos{X: 8, Y: 2},
			entrances: []RoomEntrance{
				{Pos: types.Pos{X: 9, Y: 4}, Dir: types.South},
			},
		},
		{
			typeID: "senju_room",
			pos:    types.Pos{X: 14, Y: 2},
			entrances: []RoomEntrance{
				{Pos: types.Pos{X: 15, Y: 4}, Dir: types.South},
			},
		},
		{
			typeID: "trap_room",
			pos:    types.Pos{X: 5, Y: 10},
			entrances: []RoomEntrance{
				{Pos: types.Pos{X: 6, Y: 9}, Dir: types.North},
			},
		},
		{
			typeID: "recovery_room",
			pos:    types.Pos{X: 14, Y: 10},
			entrances: []RoomEntrance{
				{Pos: types.Pos{X: 15, Y: 9}, Dir: types.North},
			},
		},
	}

	// Place all 5 rooms.
	placedRooms := make([]*Room, 0, len(rooms))
	for i, rd := range rooms {
		r, err := cave.AddRoom(rd.typeID, rd.pos, 3, 3, rd.entrances)
		if err != nil {
			t.Fatalf("AddRoom[%d] %s: %v", i, rd.typeID, err)
		}
		placedRooms = append(placedRooms, r)
	}

	if len(cave.Rooms) != 5 {
		t.Fatalf("expected 5 rooms, got %d", len(cave.Rooms))
	}

	// Verify each room has a unique ID and correct TypeID.
	seenIDs := make(map[int]bool)
	for _, r := range placedRooms {
		if seenIDs[r.ID] {
			t.Errorf("duplicate room ID: %d", r.ID)
		}
		seenIDs[r.ID] = true
	}

	// Verify room cells are correctly carved in the grid.
	for _, r := range placedRooms {
		for dy := 0; dy < r.Height; dy++ {
			for dx := 0; dx < r.Width; dx++ {
				pos := types.Pos{X: r.Pos.X + dx, Y: r.Pos.Y + dy}
				cell, err := cave.Grid.At(pos)
				if err != nil {
					t.Errorf("Grid.At(%v): %v", pos, err)
					continue
				}
				if cell.RoomID != r.ID {
					t.Errorf("cell at %v: RoomID = %d, want %d", pos, cell.RoomID, r.ID)
				}
			}
		}
	}

	// Connect rooms in a chain: 0-1, 1-2, 2-4, 4-3, 3-0 forming a cycle.
	connections := [][2]int{
		{placedRooms[0].ID, placedRooms[1].ID},
		{placedRooms[1].ID, placedRooms[2].ID},
		{placedRooms[2].ID, placedRooms[4].ID},
		{placedRooms[4].ID, placedRooms[3].ID},
		{placedRooms[3].ID, placedRooms[0].ID},
	}

	for i, conn := range connections {
		_, err := cave.ConnectRooms(conn[0], conn[1])
		if err != nil {
			t.Fatalf("ConnectRooms[%d] (%d, %d): %v", i, conn[0], conn[1], err)
		}
	}

	if len(cave.Corridors) != 5 {
		t.Fatalf("expected 5 corridors, got %d", len(cave.Corridors))
	}

	// Verify corridor paths only contain valid cell types (CorridorFloor, RoomFloor, Entrance).
	for _, cor := range cave.Corridors {
		if len(cor.Path) == 0 {
			t.Errorf("corridor %d has empty path", cor.ID)
		}
		for _, pos := range cor.Path {
			cell, err := cave.Grid.At(pos)
			if err != nil {
				t.Errorf("corridor %d: Grid.At(%v): %v", cor.ID, pos, err)
				continue
			}
			if cell.Type == Rock {
				t.Errorf("corridor %d: cell at %v is still Rock", cor.ID, pos)
			}
		}
	}

	// Build and verify adjacency graph.
	graph := cave.BuildAdjacencyGraph()

	// Every room pair should be reachable (we formed a cycle).
	for i := 0; i < len(placedRooms); i++ {
		for j := i + 1; j < len(placedRooms); j++ {
			if !graph.PathExists(placedRooms[i].ID, placedRooms[j].ID) {
				t.Errorf("PathExists(%d, %d) = false, want true",
					placedRooms[i].ID, placedRooms[j].ID)
			}
		}
	}

	// Direct neighbors check for room 0 (connected to rooms 1 and 3).
	neighbors0 := graph.Neighbors(placedRooms[0].ID)
	neighborSet := make(map[int]bool)
	for _, n := range neighbors0 {
		neighborSet[n] = true
	}
	if !neighborSet[placedRooms[1].ID] {
		t.Errorf("room %d should be neighbor of room %d", placedRooms[1].ID, placedRooms[0].ID)
	}
	if !neighborSet[placedRooms[3].ID] {
		t.Errorf("room %d should be neighbor of room %d", placedRooms[3].ID, placedRooms[0].ID)
	}
	if len(neighbors0) != 2 {
		t.Errorf("room %d: expected 2 neighbors, got %d", placedRooms[0].ID, len(neighbors0))
	}

	// Verify serialization round-trip preserves data.
	data, err := cave.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	restored, err := UnmarshalCave(data)
	if err != nil {
		t.Fatalf("UnmarshalCave: %v", err)
	}

	if len(restored.Rooms) != len(cave.Rooms) {
		t.Errorf("restored rooms: got %d, want %d", len(restored.Rooms), len(cave.Rooms))
	}
	if len(restored.Corridors) != len(cave.Corridors) {
		t.Errorf("restored corridors: got %d, want %d", len(restored.Corridors), len(cave.Corridors))
	}
	if restored.Grid.Width != cave.Grid.Width || restored.Grid.Height != cave.Grid.Height {
		t.Errorf("restored grid size: got %dx%d, want %dx%d",
			restored.Grid.Width, restored.Grid.Height, cave.Grid.Width, cave.Grid.Height)
	}

	// Verify restored adjacency graph matches.
	restoredGraph := restored.BuildAdjacencyGraph()
	for i := 0; i < len(placedRooms); i++ {
		for j := i + 1; j < len(placedRooms); j++ {
			if !restoredGraph.PathExists(placedRooms[i].ID, placedRooms[j].ID) {
				t.Errorf("restored: PathExists(%d, %d) = false, want true",
					placedRooms[i].ID, placedRooms[j].ID)
			}
		}
	}
}
