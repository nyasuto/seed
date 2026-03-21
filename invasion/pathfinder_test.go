package invasion

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/testutil"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// makeConnectedCave creates a cave with connected rooms for pathfinding tests.
//
// Layout (20x20 grid):
//
//	Room 1 (dragon_hole) at (1,1) 3x3
//	Room 2 (chi_chamber)  at (6,1) 3x3
//	Room 3 (storage)      at (11,1) 3x3
//	Room 4 (beast_room)   at (6,6) 3x3
//
// Connections: 1-2, 2-3, 2-4 (star topology centered on room 2)
func makeConnectedCave(t *testing.T) (*world.Cave, []*world.Room) {
	t.Helper()
	cave, err := world.NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1: dragon_hole (core) with entrance on east side
	r1, err := cave.AddRoom("dragon_hole", types.Pos{X: 1, Y: 1}, 3, 3,
		[]world.RoomEntrance{{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East}})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	// Room 2: chi_chamber with entrances on west, east, and south
	r2, err := cave.AddRoom("chi_chamber", types.Pos{X: 6, Y: 1}, 3, 3,
		[]world.RoomEntrance{
			{Pos: types.Pos{X: 6, Y: 2}, Dir: types.West},
			{Pos: types.Pos{X: 8, Y: 2}, Dir: types.East},
			{Pos: types.Pos{X: 7, Y: 3}, Dir: types.South},
		})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	// Room 3: storage with entrance on west
	r3, err := cave.AddRoom("storage", types.Pos{X: 11, Y: 1}, 3, 3,
		[]world.RoomEntrance{{Pos: types.Pos{X: 11, Y: 2}, Dir: types.West}})
	if err != nil {
		t.Fatalf("AddRoom 3: %v", err)
	}

	// Room 4: beast_room with entrance on north
	r4, err := cave.AddRoom("beast_room", types.Pos{X: 6, Y: 6}, 3, 3,
		[]world.RoomEntrance{{Pos: types.Pos{X: 7, Y: 6}, Dir: types.North}})
	if err != nil {
		t.Fatalf("AddRoom 4: %v", err)
	}

	// Add beasts to beast_room
	cave.RoomByID(r4.ID).BeastIDs = []int{100, 101}

	// Connect: 1-2, 2-3, 2-4
	if _, err := cave.ConnectRooms(r1.ID, r2.ID); err != nil {
		t.Fatalf("ConnectRooms 1-2: %v", err)
	}
	if _, err := cave.ConnectRooms(r2.ID, r3.ID); err != nil {
		t.Fatalf("ConnectRooms 2-3: %v", err)
	}
	if _, err := cave.ConnectRooms(r2.ID, r4.ID); err != nil {
		t.Fatalf("ConnectRooms 2-4: %v", err)
	}

	rooms := []*world.Room{
		cave.RoomByID(r1.ID),
		cave.RoomByID(r2.ID),
		cave.RoomByID(r3.ID),
		cave.RoomByID(r4.ID),
	}
	return cave, rooms
}

func TestFindPath_DirectConnection(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)

	path := pf.FindPath(rooms[0].ID, rooms[1].ID)
	if len(path) != 2 {
		t.Fatalf("FindPath len = %d, want 2; path = %v", len(path), path)
	}
	if path[0] != rooms[0].ID || path[1] != rooms[1].ID {
		t.Errorf("FindPath = %v, want [%d, %d]", path, rooms[0].ID, rooms[1].ID)
	}
}

func TestFindPath_TwoHops(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)

	// Room 1 -> Room 3 goes through Room 2
	path := pf.FindPath(rooms[0].ID, rooms[2].ID)
	if len(path) != 3 {
		t.Fatalf("FindPath len = %d, want 3; path = %v", len(path), path)
	}
	if path[0] != rooms[0].ID || path[1] != rooms[1].ID || path[2] != rooms[2].ID {
		t.Errorf("FindPath = %v, want [%d, %d, %d]",
			path, rooms[0].ID, rooms[1].ID, rooms[2].ID)
	}
}

func TestFindPath_SameRoom(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)

	path := pf.FindPath(rooms[0].ID, rooms[0].ID)
	if len(path) != 1 || path[0] != rooms[0].ID {
		t.Errorf("FindPath same room = %v, want [%d]", path, rooms[0].ID)
	}
}

func TestFindPath_NoPath(t *testing.T) {
	cave, _ := makeConnectedCave(t)
	// Add an isolated room
	r5, err := cave.AddRoom("chi_chamber", types.Pos{X: 15, Y: 15}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom isolated: %v", err)
	}
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)

	path := pf.FindPath(1, r5.ID)
	if path != nil {
		t.Errorf("FindPath to isolated room = %v, want nil", path)
	}
}

func TestFindNextRoom_KnownCoreRoom(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)
	rng := &testutil.FixedRNG{IntValue: 0}

	// Invader at room 3 (storage), goal is to destroy core (room 1).
	invader := &Invader{
		CurrentRoomID: rooms[2].ID,
		Goal:          NewDestroyCoreGoal(),
		Memory:        NewExplorationMemory(),
	}
	// Record knowledge of core room.
	invader.Memory.RecordCoreRoom(rooms[0].ID)

	next := pf.FindNextRoom(invader, rng)
	// Should move toward core: room 3 -> room 2 -> room 1, so next = room 2
	if next != rooms[1].ID {
		t.Errorf("FindNextRoom = %d, want %d (toward core)", next, rooms[1].ID)
	}
}

func TestFindNextRoom_UnknownTarget_ExploresUnvisited(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)
	rng := &testutil.FixedRNG{IntValue: 0}

	// Invader at room 2, goal is destroy core but core is unknown.
	invader := &Invader{
		CurrentRoomID: rooms[1].ID,
		Goal:          NewDestroyCoreGoal(),
		Memory:        NewExplorationMemory(),
	}
	// Mark room 2 as visited, but neighbors are unvisited.
	invader.Memory.VisitedRooms[rooms[1].ID] = 1

	next := pf.FindNextRoom(invader, rng)
	// Should pick an unvisited neighbor. All 3 neighbors are unvisited.
	neighbors := graph.Neighbors(rooms[1].ID)
	found := false
	for _, n := range neighbors {
		if n == next {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("FindNextRoom = %d, not a neighbor of room %d", next, rooms[1].ID)
	}
	if invader.Memory.HasVisited(next) {
		t.Errorf("FindNextRoom chose visited room %d, should prefer unvisited", next)
	}
}

func TestFindNextRoom_Backtrack(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)
	rng := &testutil.FixedRNG{IntValue: 0}

	// Invader at room 3 (leaf node), all neighbors visited, but room 4 unvisited.
	invader := &Invader{
		CurrentRoomID: rooms[2].ID, // storage (room 3)
		Goal:          NewDestroyCoreGoal(),
		Memory:        NewExplorationMemory(),
	}
	// Mark rooms 1, 2, 3 as visited; room 4 is unvisited.
	invader.Memory.VisitedRooms[rooms[0].ID] = 1
	invader.Memory.VisitedRooms[rooms[1].ID] = 2
	invader.Memory.VisitedRooms[rooms[2].ID] = 3

	next := pf.FindNextRoom(invader, rng)
	// Should backtrack toward room 4 via room 2.
	if next != rooms[1].ID {
		t.Errorf("FindNextRoom backtrack = %d, want %d (toward unvisited room 4)", next, rooms[1].ID)
	}
}

func TestFindNextRoom_FullyExplored_RandomMove(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)
	rng := &testutil.FixedRNG{IntValue: 0}

	// Invader at room 2, all rooms visited, no goal target known.
	invader := &Invader{
		CurrentRoomID: rooms[1].ID,
		Goal:          NewDestroyCoreGoal(),
		Memory:        NewExplorationMemory(),
	}
	// Mark all rooms visited.
	for _, r := range rooms {
		invader.Memory.VisitedRooms[r.ID] = 1
	}
	// Core room is not in the cave with matching TypeID "dragon_hole"
	// Actually room 1 IS dragon_hole, so the goal will find it.
	// Let's use HuntBeastsGoal with kills already met to get no target.
	huntGoal := NewHuntBeastsGoal()
	huntGoal.Kills = huntGoal.RequiredKills // already achieved
	invader.Goal = huntGoal

	next := pf.FindNextRoom(invader, rng)
	// Should pick a random neighbor.
	neighbors := graph.Neighbors(rooms[1].ID)
	found := false
	for _, n := range neighbors {
		if n == next {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("FindNextRoom random = %d, not a neighbor of room %d", next, rooms[1].ID)
	}
}

func TestFindNextRoom_ThiefTargetsStorage(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)
	rng := &testutil.FixedRNG{IntValue: 0}

	// Invader at room 1 (core), goal is steal treasure.
	invader := &Invader{
		CurrentRoomID: rooms[0].ID,
		Goal:          NewStealTreasureGoal(),
		Memory:        NewExplorationMemory(),
	}
	// Record knowledge of storage room (room 3).
	invader.Memory.RecordTreasureRoom(rooms[2].ID)

	next := pf.FindNextRoom(invader, rng)
	// Path: room 1 -> room 2 -> room 3, so next should be room 2.
	if next != rooms[1].ID {
		t.Errorf("FindNextRoom thief = %d, want %d (toward storage)", next, rooms[1].ID)
	}
}

func TestFindNextRoom_HunterTargetsBeastRoom(t *testing.T) {
	cave, rooms := makeConnectedCave(t)
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)
	rng := &testutil.FixedRNG{IntValue: 0}

	// Invader at room 1, goal is hunt beasts.
	invader := &Invader{
		CurrentRoomID: rooms[0].ID,
		Goal:          NewHuntBeastsGoal(),
		Memory:        NewExplorationMemory(),
	}
	// Record knowledge of beast room (room 4).
	invader.Memory.RecordBeastRoom(rooms[3].ID, true)

	next := pf.FindNextRoom(invader, rng)
	// Path: room 1 -> room 2 -> room 4, so next should be room 2.
	if next != rooms[1].ID {
		t.Errorf("FindNextRoom hunter = %d, want %d (toward beast room)", next, rooms[1].ID)
	}
}

func TestFindNextRoom_IsolatedRoom_StaysInPlace(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	r1, err := cave.AddRoom("chi_chamber", types.Pos{X: 1, Y: 1}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}
	graph := cave.BuildAdjacencyGraph()
	pf := NewPathfinder(cave, graph)
	rng := &testutil.FixedRNG{IntValue: 0}

	invader := &Invader{
		CurrentRoomID: r1.ID,
		Goal:          NewDestroyCoreGoal(),
		Memory:        NewExplorationMemory(),
	}

	next := pf.FindNextRoom(invader, rng)
	if next != r1.ID {
		t.Errorf("FindNextRoom isolated = %d, want %d (stay in place)", next, r1.ID)
	}
}
