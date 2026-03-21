package invasion

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

func makeTestCaveWithRooms(t *testing.T) (*world.Cave, []*world.Room) {
	t.Helper()
	cave, err := world.NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1: dragon_hole (core)
	r1, err := cave.AddRoom("dragon_hole", types.Pos{X: 1, Y: 1}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom dragon_hole: %v", err)
	}

	// Room 2: storage (treasure)
	r2, err := cave.AddRoom("storage", types.Pos{X: 6, Y: 1}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom storage: %v", err)
	}

	// Room 3: beast_room (beasts)
	r3, err := cave.AddRoom("beast_room", types.Pos{X: 11, Y: 1}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom beast_room: %v", err)
	}

	// Room 4: chi_chamber (generic)
	r4, err := cave.AddRoom("chi_chamber", types.Pos{X: 1, Y: 6}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom chi_chamber: %v", err)
	}

	room1 := cave.RoomByID(r1.ID)
	room2 := cave.RoomByID(r2.ID)
	room3 := cave.RoomByID(r3.ID)
	room4 := cave.RoomByID(r4.ID)

	// Add beasts to beast_room.
	room3.BeastIDs = []int{100, 101}

	rooms := []*world.Room{room1, room2, room3, room4}
	return cave, rooms
}

func TestExplorationMemory_Visit_RecordsFirstVisitTick(t *testing.T) {
	cave, rooms := makeTestCaveWithRooms(t)
	mem := NewExplorationMemory()

	roomID := rooms[3].ID // chi_chamber
	mem.Visit(roomID, 10, cave, rooms)

	if !mem.HasVisited(roomID) {
		t.Error("room should be marked as visited")
	}
	if tick, ok := mem.VisitedRooms[roomID]; !ok || tick != 10 {
		t.Errorf("VisitedRooms[%d] = %d, want 10", roomID, tick)
	}

	// Second visit should NOT update the tick.
	mem.Visit(roomID, 20, cave, rooms)
	if tick := mem.VisitedRooms[roomID]; tick != 10 {
		t.Errorf("VisitedRooms[%d] = %d after second visit, want 10 (first visit tick)", roomID, tick)
	}
}

func TestExplorationMemory_Visit_DiscoversCoreRoom(t *testing.T) {
	cave, rooms := makeTestCaveWithRooms(t)
	mem := NewExplorationMemory()

	coreRoom := rooms[0] // dragon_hole
	mem.Visit(coreRoom.ID, 5, cave, rooms)

	if mem.KnownCoreRoom != coreRoom.ID {
		t.Errorf("KnownCoreRoom = %d, want %d", mem.KnownCoreRoom, coreRoom.ID)
	}
}

func TestExplorationMemory_Visit_DiscoversTreasureRoom(t *testing.T) {
	cave, rooms := makeTestCaveWithRooms(t)
	mem := NewExplorationMemory()

	storageRoom := rooms[1] // storage
	mem.Visit(storageRoom.ID, 5, cave, rooms)

	if len(mem.KnownTreasureRooms) != 1 || mem.KnownTreasureRooms[0] != storageRoom.ID {
		t.Errorf("KnownTreasureRooms = %v, want [%d]", mem.KnownTreasureRooms, storageRoom.ID)
	}

	// Visiting again should NOT add duplicate.
	mem.Visit(storageRoom.ID, 10, cave, rooms)
	if len(mem.KnownTreasureRooms) != 1 {
		t.Errorf("KnownTreasureRooms has %d entries after duplicate visit, want 1", len(mem.KnownTreasureRooms))
	}
}

func TestExplorationMemory_Visit_DiscoversBeastRoom(t *testing.T) {
	cave, rooms := makeTestCaveWithRooms(t)
	mem := NewExplorationMemory()

	beastRoom := rooms[2] // beast_room with 2 beasts
	mem.Visit(beastRoom.ID, 5, cave, rooms)

	if !mem.KnownBeastRooms[beastRoom.ID] {
		t.Errorf("KnownBeastRooms[%d] = false, want true", beastRoom.ID)
	}
}

func TestExplorationMemory_Visit_RecordsNoBeastsInEmptyRoom(t *testing.T) {
	cave, rooms := makeTestCaveWithRooms(t)
	mem := NewExplorationMemory()

	emptyRoom := rooms[3] // chi_chamber, no beasts
	mem.Visit(emptyRoom.ID, 5, cave, rooms)

	if hasBeast, ok := mem.KnownBeastRooms[emptyRoom.ID]; !ok || hasBeast {
		t.Errorf("KnownBeastRooms[%d] = %v, want false", emptyRoom.ID, hasBeast)
	}
}

func TestExplorationMemory_HasVisited(t *testing.T) {
	mem := NewExplorationMemory()

	if mem.HasVisited(1) {
		t.Error("unvisited room should return false")
	}

	mem.VisitedRooms[1] = 5
	if !mem.HasVisited(1) {
		t.Error("visited room should return true")
	}
}

func TestExplorationMemory_VisitedCount(t *testing.T) {
	mem := NewExplorationMemory()
	if mem.VisitedCount() != 0 {
		t.Errorf("VisitedCount = %d, want 0", mem.VisitedCount())
	}

	mem.VisitedRooms[1] = 5
	mem.VisitedRooms[2] = 10
	if mem.VisitedCount() != 2 {
		t.Errorf("VisitedCount = %d, want 2", mem.VisitedCount())
	}
}

func TestExplorationMemory_RecordBeastRoom(t *testing.T) {
	mem := NewExplorationMemory()

	mem.RecordBeastRoom(5, true)
	if !mem.KnownBeastRooms[5] {
		t.Error("RecordBeastRoom(5, true) should set KnownBeastRooms[5] = true")
	}

	mem.RecordBeastRoom(5, false)
	if mem.KnownBeastRooms[5] {
		t.Error("RecordBeastRoom(5, false) should set KnownBeastRooms[5] = false")
	}
}

func TestExplorationMemory_RecordCoreRoom(t *testing.T) {
	mem := NewExplorationMemory()

	mem.RecordCoreRoom(42)
	if mem.KnownCoreRoom != 42 {
		t.Errorf("KnownCoreRoom = %d, want 42", mem.KnownCoreRoom)
	}
}

func TestExplorationMemory_RecordTreasureRoom(t *testing.T) {
	mem := NewExplorationMemory()

	mem.RecordTreasureRoom(10)
	mem.RecordTreasureRoom(20)
	mem.RecordTreasureRoom(10) // duplicate

	if len(mem.KnownTreasureRooms) != 2 {
		t.Errorf("KnownTreasureRooms has %d entries, want 2", len(mem.KnownTreasureRooms))
	}
}

func TestExplorationMemory_Visit_UnknownRoom(t *testing.T) {
	cave, rooms := makeTestCaveWithRooms(t)
	mem := NewExplorationMemory()

	// Visit a room ID that doesn't exist in the rooms slice.
	mem.Visit(9999, 5, cave, rooms)

	// Should still be recorded as visited.
	if !mem.HasVisited(9999) {
		t.Error("unknown room should still be recorded as visited")
	}

	// But no special rooms should be discovered.
	if mem.KnownCoreRoom != 0 {
		t.Error("KnownCoreRoom should remain 0 for unknown room")
	}
	if len(mem.KnownTreasureRooms) != 0 {
		t.Error("KnownTreasureRooms should remain empty for unknown room")
	}
}

func TestExplorationMemory_Visit_MultipleRooms(t *testing.T) {
	cave, rooms := makeTestCaveWithRooms(t)
	mem := NewExplorationMemory()

	// Visit all rooms.
	for i, room := range rooms {
		mem.Visit(room.ID, types.Tick(i+1), cave, rooms)
	}

	if mem.VisitedCount() != 4 {
		t.Errorf("VisitedCount = %d, want 4", mem.VisitedCount())
	}
	if mem.KnownCoreRoom != rooms[0].ID {
		t.Errorf("KnownCoreRoom = %d, want %d", mem.KnownCoreRoom, rooms[0].ID)
	}
	if len(mem.KnownTreasureRooms) != 1 {
		t.Errorf("KnownTreasureRooms has %d entries, want 1", len(mem.KnownTreasureRooms))
	}
	if !mem.KnownBeastRooms[rooms[2].ID] {
		t.Errorf("beast room should be known")
	}
}
