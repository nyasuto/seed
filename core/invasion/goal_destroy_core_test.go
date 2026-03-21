package invasion

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

func newTestCaveWithCoreRoom(t *testing.T) (*world.Cave, int) {
	t.Helper()
	cave, err := world.NewCave(16, 16)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}
	room, err := cave.AddRoom("dragon_hole", types.Pos{X: 2, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding core room: %v", err)
	}
	return cave, room.ID
}

func TestDestroyCoreGoal_Type(t *testing.T) {
	g := NewDestroyCoreGoal()
	if g.Type() != DestroyCore {
		t.Errorf("expected DestroyCore, got %v", g.Type())
	}
}

func TestDestroyCoreGoal_TargetRoomID_KnownCore(t *testing.T) {
	cave, coreID := newTestCaveWithCoreRoom(t)
	g := NewDestroyCoreGoal()

	mem := &ExplorationMemory{KnownCoreRoom: coreID}
	inv := &Invader{}

	target := g.TargetRoomID(cave, inv, mem)
	if target != coreID {
		t.Errorf("expected target %d, got %d", coreID, target)
	}
}

func TestDestroyCoreGoal_TargetRoomID_UnknownCore(t *testing.T) {
	cave, coreID := newTestCaveWithCoreRoom(t)
	g := NewDestroyCoreGoal()

	mem := &ExplorationMemory{}
	inv := &Invader{}

	// Even without known core room, scanning finds it.
	target := g.TargetRoomID(cave, inv, mem)
	if target != coreID {
		t.Errorf("expected target %d from scan, got %d", coreID, target)
	}
}

func TestDestroyCoreGoal_TargetRoomID_NoCoreRoom(t *testing.T) {
	cave, err := world.NewCave(16, 16)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}
	g := NewDestroyCoreGoal()
	mem := &ExplorationMemory{}
	inv := &Invader{}

	target := g.TargetRoomID(cave, inv, mem)
	if target != 0 {
		t.Errorf("expected 0 for missing core, got %d", target)
	}
}

func TestDestroyCoreGoal_TargetRoomID_NilMemory(t *testing.T) {
	cave, coreID := newTestCaveWithCoreRoom(t)
	g := NewDestroyCoreGoal()

	target := g.TargetRoomID(cave, &Invader{}, nil)
	if target != coreID {
		t.Errorf("expected target %d with nil memory, got %d", coreID, target)
	}
}

func TestDestroyCoreGoal_IsAchieved(t *testing.T) {
	cave, coreID := newTestCaveWithCoreRoom(t)
	g := NewDestroyCoreGoal()

	tests := []struct {
		name     string
		roomID   int
		stay     int
		achieved bool
	}{
		{"not in core room", 999, 10, false},
		{"in core, not enough ticks", coreID, DefaultDestroyCoreTicks - 1, false},
		{"in core, exact ticks", coreID, DefaultDestroyCoreTicks, true},
		{"in core, extra ticks", coreID, DefaultDestroyCoreTicks + 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invader{
				CurrentRoomID: tt.roomID,
				StayTicks:     tt.stay,
			}
			if got := g.IsAchieved(cave, inv); got != tt.achieved {
				t.Errorf("IsAchieved() = %v, want %v", got, tt.achieved)
			}
		})
	}
}
