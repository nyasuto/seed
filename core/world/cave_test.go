package world

import (
	"errors"
	"sort"
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestNewCave(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	if cave.Grid.Width != 20 || cave.Grid.Height != 20 {
		t.Errorf("grid size = %dx%d, want 20x20", cave.Grid.Width, cave.Grid.Height)
	}
	if len(cave.Rooms) != 0 {
		t.Errorf("new cave should have 0 rooms, got %d", len(cave.Rooms))
	}
	if len(cave.Corridors) != 0 {
		t.Errorf("new cave should have 0 corridors, got %d", len(cave.Corridors))
	}
}

func TestNewCave_InvalidSize(t *testing.T) {
	_, err := NewCave(0, 10)
	if err == nil {
		t.Error("NewCave(0, 10) should return error")
	}
}

func TestCave_AddRoom(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	entrances := []RoomEntrance{
		{Pos: types.Pos{X: 4, Y: 2}, Dir: types.East},
	}
	room, err := cave.AddRoom("test", types.Pos{X: 2, Y: 2}, 3, 3, entrances)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}
	if room.ID != 1 {
		t.Errorf("first room ID = %d, want 1", room.ID)
	}
	if len(cave.Rooms) != 1 {
		t.Errorf("cave should have 1 room, got %d", len(cave.Rooms))
	}
}

func TestCave_AddRoom_Overlap(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	_, err = cave.AddRoom("test", types.Pos{X: 2, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom first: %v", err)
	}

	_, err = cave.AddRoom("test", types.Pos{X: 3, Y: 3}, 3, 3, nil)
	if err == nil {
		t.Error("AddRoom should fail for overlapping room")
	}
	if !errors.Is(err, ErrRoomOverlap) {
		t.Errorf("error = %v, want ErrRoomOverlap", err)
	}
}

func TestCave_AddRoom_AutoID(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	r1, err := cave.AddRoom("test", types.Pos{X: 1, Y: 1}, 2, 2, nil)
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}
	r2, err := cave.AddRoom("test", types.Pos{X: 5, Y: 5}, 2, 2, nil)
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	if r1.ID != 1 || r2.ID != 2 {
		t.Errorf("room IDs = (%d, %d), want (1, 2)", r1.ID, r2.ID)
	}
}

func TestCave_ConnectRooms(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1 at (1,1) 3x3 with entrance on right side
	r1, err := cave.AddRoom("test", types.Pos{X: 1, Y: 1}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	// Room 2 at (8,1) 3x3 with entrance on left side
	r2, err := cave.AddRoom("test", types.Pos{X: 8, Y: 1}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 8, Y: 2}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	corridor, err := cave.ConnectRooms(r1.ID, r2.ID)
	if err != nil {
		t.Fatalf("ConnectRooms: %v", err)
	}
	if corridor.FromRoomID != r1.ID || corridor.ToRoomID != r2.ID {
		t.Errorf("corridor rooms = (%d, %d), want (%d, %d)",
			corridor.FromRoomID, corridor.ToRoomID, r1.ID, r2.ID)
	}
	if len(corridor.Path) == 0 {
		t.Error("corridor path should not be empty")
	}
	if len(cave.Corridors) != 1 {
		t.Errorf("cave should have 1 corridor, got %d", len(cave.Corridors))
	}
}

func TestCave_ConnectRooms_RoomNotFound(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	_, err = cave.ConnectRooms(1, 2)
	if err == nil {
		t.Error("ConnectRooms should fail for non-existent rooms")
	}
	if !errors.Is(err, ErrRoomNotFound) {
		t.Errorf("error = %v, want ErrRoomNotFound", err)
	}
}

func TestCave_ConnectRooms_NoEntrance(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Add rooms without entrances
	_, err = cave.AddRoom("test", types.Pos{X: 1, Y: 1}, 2, 2, nil)
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}
	_, err = cave.AddRoom("test", types.Pos{X: 5, Y: 5}, 2, 2, nil)
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	_, err = cave.ConnectRooms(1, 2)
	if err == nil {
		t.Error("ConnectRooms should fail for rooms without entrances")
	}
	if !errors.Is(err, ErrNoEntrance) {
		t.Errorf("error = %v, want ErrNoEntrance", err)
	}
}

func TestCave_RoomByID(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	r, err := cave.AddRoom("test", types.Pos{X: 1, Y: 1}, 2, 2, nil)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	found := cave.RoomByID(r.ID)
	if found == nil {
		t.Fatal("RoomByID should find the room")
	}
	if found.ID != r.ID {
		t.Errorf("RoomByID returned room with ID %d, want %d", found.ID, r.ID)
	}

	notFound := cave.RoomByID(999)
	if notFound != nil {
		t.Error("RoomByID should return nil for non-existent room")
	}
}

// TestCave_AddTwoRooms_Connect_AdjacencyGraph is the integration test:
// create cave → add 2 rooms → connect via corridor → verify adjacency graph.
func TestCave_AddTwoRooms_Connect_AdjacencyGraph(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Add room 1 at (1,1) 3x3 with entrance on the right side
	r1, err := cave.AddRoom("chi_chamber", types.Pos{X: 1, Y: 1}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	// Add room 2 at (8,1) 3x3 with entrance on the left side
	r2, err := cave.AddRoom("trap_room", types.Pos{X: 8, Y: 1}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 8, Y: 2}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	// Connect rooms
	_, err = cave.ConnectRooms(r1.ID, r2.ID)
	if err != nil {
		t.Fatalf("ConnectRooms: %v", err)
	}

	// Build adjacency graph and verify
	graph := cave.BuildAdjacencyGraph()

	// r1 and r2 should be neighbors
	neighbors := graph.Neighbors(r1.ID)
	if len(neighbors) != 1 || neighbors[0] != r2.ID {
		t.Errorf("room %d neighbors = %v, want [%d]", r1.ID, neighbors, r2.ID)
	}

	neighbors = graph.Neighbors(r2.ID)
	if len(neighbors) != 1 || neighbors[0] != r1.ID {
		t.Errorf("room %d neighbors = %v, want [%d]", r2.ID, neighbors, r1.ID)
	}

	// Path should exist between the two rooms
	if !graph.PathExists(r1.ID, r2.ID) {
		t.Error("PathExists should return true for connected rooms")
	}
	if !graph.PathExists(r2.ID, r1.ID) {
		t.Error("PathExists should return true in reverse direction")
	}
}

// TestCave_AdjacencyGraph_ThreeRooms tests graph with 3 rooms in a chain: R1-R2-R3.
func TestCave_AdjacencyGraph_ThreeRooms(t *testing.T) {
	cave, err := NewCave(30, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1 at (1,1) with entrance on right
	r1, err := cave.AddRoom("test", types.Pos{X: 1, Y: 1}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	// Room 2 at (8,1) with entrances on left and right
	r2, err := cave.AddRoom("test", types.Pos{X: 8, Y: 1}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 8, Y: 2}, Dir: types.West},
		{Pos: types.Pos{X: 10, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	// Room 3 at (15,1) with entrance on left
	r3, err := cave.AddRoom("test", types.Pos{X: 15, Y: 1}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 15, Y: 2}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom 3: %v", err)
	}

	// Connect R1-R2 and R2-R3
	_, err = cave.ConnectRooms(r1.ID, r2.ID)
	if err != nil {
		t.Fatalf("ConnectRooms(1,2): %v", err)
	}
	_, err = cave.ConnectRooms(r2.ID, r3.ID)
	if err != nil {
		t.Fatalf("ConnectRooms(2,3): %v", err)
	}

	graph := cave.BuildAdjacencyGraph()

	// R2 should have 2 neighbors
	neighbors := graph.Neighbors(r2.ID)
	sort.Ints(neighbors)
	if len(neighbors) != 2 || neighbors[0] != r1.ID || neighbors[1] != r3.ID {
		t.Errorf("room %d neighbors = %v, want [%d %d]", r2.ID, neighbors, r1.ID, r3.ID)
	}

	// R1 and R3 should be reachable via R2
	if !graph.PathExists(r1.ID, r3.ID) {
		t.Error("PathExists(R1, R3) should return true (connected via R2)")
	}

	// R1 should not be a direct neighbor of R3
	r1Neighbors := graph.Neighbors(r1.ID)
	for _, n := range r1Neighbors {
		if n == r3.ID {
			t.Error("R1 should not be a direct neighbor of R3")
		}
	}
}

// TestCave_AdjacencyGraph_DisconnectedRooms tests that disconnected rooms
// have no path between them.
func TestCave_AdjacencyGraph_DisconnectedRooms(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	r1, err := cave.AddRoom("test", types.Pos{X: 1, Y: 1}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	r2, err := cave.AddRoom("test", types.Pos{X: 10, Y: 10}, 3, 3, []RoomEntrance{
		{Pos: types.Pos{X: 10, Y: 11}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	graph := cave.BuildAdjacencyGraph()

	if graph.PathExists(r1.ID, r2.ID) {
		t.Error("PathExists should return false for disconnected rooms")
	}

	// Both rooms should exist in the graph with no neighbors
	n1 := graph.Neighbors(r1.ID)
	if len(n1) != 0 {
		t.Errorf("disconnected room %d should have 0 neighbors, got %v", r1.ID, n1)
	}
}

// TestAdjacencyGraph_Neighbors_UnknownRoom tests that Neighbors returns nil
// for a room ID not in the graph.
func TestAdjacencyGraph_Neighbors_UnknownRoom(t *testing.T) {
	cave, err := NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	graph := cave.BuildAdjacencyGraph()
	if neighbors := graph.Neighbors(999); neighbors != nil {
		t.Errorf("Neighbors(999) = %v, want nil", neighbors)
	}
}

// TestAdjacencyGraph_PathExists_SameRoom tests that PathExists returns true
// when from and to are the same room.
func TestAdjacencyGraph_PathExists_SameRoom(t *testing.T) {
	cave, err := NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	r, err := cave.AddRoom("test", types.Pos{X: 1, Y: 1}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	graph := cave.BuildAdjacencyGraph()
	if !graph.PathExists(r.ID, r.ID) {
		t.Error("PathExists should return true when from == to")
	}
}
