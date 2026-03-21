package invasion

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// newTestCaveWithBeasts creates a cave with a beast room and a regular room.
// Returns the cave, beast room ID, and non-beast room ID.
func newTestCaveWithBeasts(t *testing.T) (*world.Cave, int, int) {
	t.Helper()
	cave, err := world.NewCave(16, 16)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}
	beastRoom, err := cave.AddRoom("beast_room", types.Pos{X: 2, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding beast room: %v", err)
	}
	// Simulate a beast placed in this room.
	beastRoom.BeastIDs = []int{100}

	otherRoom, err := cave.AddRoom("storage", types.Pos{X: 8, Y: 8}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding other room: %v", err)
	}
	return cave, beastRoom.ID, otherRoom.ID
}

func TestHuntBeastsGoal_Type(t *testing.T) {
	g := NewHuntBeastsGoal()
	if g.Type() != HuntBeasts {
		t.Errorf("expected HuntBeasts, got %v", g.Type())
	}
}

func TestHuntBeastsGoal_TargetRoomID_KnownBeastRoom(t *testing.T) {
	cave, beastRoomID, otherRoomID := newTestCaveWithBeasts(t)
	g := NewHuntBeastsGoal()

	mem := &ExplorationMemory{
		KnownBeastRooms: map[int]bool{beastRoomID: true},
	}
	inv := &Invader{CurrentRoomID: otherRoomID}

	target := g.TargetRoomID(cave, inv, mem)
	if target != beastRoomID {
		t.Errorf("expected target %d, got %d", beastRoomID, target)
	}
}

func TestHuntBeastsGoal_TargetRoomID_UnknownBeastRoom(t *testing.T) {
	cave, beastRoomID, otherRoomID := newTestCaveWithBeasts(t)
	g := NewHuntBeastsGoal()

	// No memory of beast rooms — should scan cave and find the beast room.
	mem := &ExplorationMemory{}
	inv := &Invader{CurrentRoomID: otherRoomID}

	target := g.TargetRoomID(cave, inv, mem)
	if target != beastRoomID {
		t.Errorf("expected target %d from scan, got %d", beastRoomID, target)
	}
}

func TestHuntBeastsGoal_TargetRoomID_NilMemory(t *testing.T) {
	cave, beastRoomID, otherRoomID := newTestCaveWithBeasts(t)
	g := NewHuntBeastsGoal()

	inv := &Invader{CurrentRoomID: otherRoomID}

	target := g.TargetRoomID(cave, inv, nil)
	if target != beastRoomID {
		t.Errorf("expected target %d with nil memory, got %d", beastRoomID, target)
	}
}

func TestHuntBeastsGoal_TargetRoomID_NoBeastRooms(t *testing.T) {
	cave, err := world.NewCave(16, 16)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}
	// Room with no beasts.
	_, err = cave.AddRoom("storage", types.Pos{X: 2, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding room: %v", err)
	}

	g := NewHuntBeastsGoal()
	mem := &ExplorationMemory{}
	inv := &Invader{CurrentRoomID: 0}

	target := g.TargetRoomID(cave, inv, mem)
	if target != 0 {
		t.Errorf("expected 0 for no beasts, got %d", target)
	}
}

func TestHuntBeastsGoal_TargetRoomID_NearestBeastRoom(t *testing.T) {
	cave, err := world.NewCave(32, 32)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}
	// Invader's room at (2,2)
	invRoom, err := cave.AddRoom("storage", types.Pos{X: 2, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding invader room: %v", err)
	}
	// Near beast room at (8,2)
	nearRoom, err := cave.AddRoom("beast_room", types.Pos{X: 8, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding near beast room: %v", err)
	}
	nearRoom.BeastIDs = []int{100}

	// Far beast room at (20,20)
	farRoom, err := cave.AddRoom("beast_room", types.Pos{X: 20, Y: 20}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding far beast room: %v", err)
	}
	farRoom.BeastIDs = []int{200}

	g := NewHuntBeastsGoal()
	inv := &Invader{CurrentRoomID: invRoom.ID}

	target := g.TargetRoomID(cave, inv, nil)
	if target != nearRoom.ID {
		t.Errorf("expected nearest room %d, got %d (far room was %d)", nearRoom.ID, target, farRoom.ID)
	}
}

func TestHuntBeastsGoal_IsAchieved(t *testing.T) {
	cave, err := world.NewCave(16, 16)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}

	tests := []struct {
		name     string
		kills    int
		required int
		achieved bool
	}{
		{"no kills", 0, 2, false},
		{"one kill, need two", 1, 2, false},
		{"exact kills", 2, 2, true},
		{"more than required", 3, 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &HuntBeastsGoal{
				RequiredKills: tt.required,
				Kills:         tt.kills,
			}
			inv := &Invader{}
			if got := g.IsAchieved(cave, inv); got != tt.achieved {
				t.Errorf("IsAchieved() = %v, want %v", got, tt.achieved)
			}
		})
	}
}

func TestHuntBeastsGoal_MemoryBeastRoomFalse(t *testing.T) {
	cave, _, otherRoomID := newTestCaveWithBeasts(t)
	g := NewHuntBeastsGoal()

	// Memory says a room had beasts but they are gone (false).
	mem := &ExplorationMemory{
		KnownBeastRooms: map[int]bool{999: false},
	}
	inv := &Invader{CurrentRoomID: otherRoomID}

	// Should fall back to scanning cave rooms.
	target := g.TargetRoomID(cave, inv, mem)
	// The beast room should still be found via scan since the actual room has beasts.
	if target == 999 {
		t.Error("should not target room with KnownBeastRooms=false")
	}
}
