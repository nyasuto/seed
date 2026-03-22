package controller

import (
	_ "embed"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

//go:embed testdata/tutorial.json
var tutorialJSON []byte

func TestNewGameController_LoadsScenario(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	snap := gc.Snapshot()
	if snap.Tick != 0 {
		t.Errorf("initial tick = %d, want 0", snap.Tick)
	}
	if gc.State() != Playing {
		t.Errorf("initial state = %d, want Playing", gc.State())
	}
}

func TestAdvanceTick_UpdatesSnapshot(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	snapBefore := gc.Snapshot()

	result, err := gc.AdvanceTick()
	if err != nil {
		t.Fatalf("AdvanceTick failed: %v", err)
	}
	if result.Status != simulation.Running {
		t.Fatalf("unexpected game status: %v", result.Status)
	}

	snapAfter := gc.Snapshot()
	if snapAfter.Tick != snapBefore.Tick+1 {
		t.Errorf("tick after advance = %d, want %d", snapAfter.Tick, snapBefore.Tick+1)
	}
}

func TestAddAction_PassedToEngine(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	// Add a DigRoom action targeting a wall cell.
	// We need to find a valid wall position. Use a simple approach:
	// dig at a known wall position in the tutorial cave (which is 16x16).
	action := simulation.DigRoomAction{
		RoomTypeID: "guardian_chamber",
		Pos:        types.Pos{X: 2, Y: 2},
		Width:      3,
		Height:     3,
	}
	gc.AddAction(action)

	pending := gc.PendingActions()
	if len(pending) != 1 {
		t.Fatalf("pending actions = %d, want 1", len(pending))
	}
	if pending[0].ActionType() != "dig_room" {
		t.Errorf("action type = %q, want %q", pending[0].ActionType(), "dig_room")
	}

	// AdvanceTick consumes pending actions.
	_, err = gc.AdvanceTick()
	if err != nil {
		t.Fatalf("AdvanceTick failed: %v", err)
	}

	// Pending should be cleared after advance.
	if len(gc.PendingActions()) != 0 {
		t.Errorf("pending actions after advance = %d, want 0", len(gc.PendingActions()))
	}
}

func TestAdvanceTick_GameOver(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	// Advance until game ends (tutorial wins at tick 300).
	for i := 0; i < 350; i++ {
		result, err := gc.AdvanceTick()
		if err != nil {
			t.Fatalf("AdvanceTick failed at tick %d: %v", i, err)
		}
		if result.Status != simulation.Running {
			if gc.State() != GameOver {
				t.Errorf("state = %d, want GameOver", gc.State())
			}
			return
		}
	}
	t.Fatal("game did not end within 350 ticks")
}

func TestAdvanceTick_AfterGameOver_NoOp(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	// Run to game over.
	for i := 0; i < 350; i++ {
		result, err := gc.AdvanceTick()
		if err != nil {
			t.Fatalf("AdvanceTick failed: %v", err)
		}
		if result.Status != simulation.Running {
			break
		}
	}

	if gc.State() != GameOver {
		t.Fatal("expected GameOver state")
	}

	snapBefore := gc.Snapshot()
	result, err := gc.AdvanceTick()
	if err != nil {
		t.Fatalf("AdvanceTick after GameOver failed: %v", err)
	}

	snapAfter := gc.Snapshot()
	if snapAfter.Tick != snapBefore.Tick {
		t.Errorf("tick changed after GameOver: %d → %d", snapBefore.Tick, snapAfter.Tick)
	}
	if result.Status == simulation.Running {
		t.Error("expected non-Running status after GameOver")
	}
}

func TestEngine_DirectAccess(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	engine := gc.Engine()
	if engine == nil {
		t.Fatal("Engine() returned nil")
	}
	if engine.State.Cave == nil {
		t.Fatal("engine.State.Cave is nil")
	}
	if engine.State.RoomTypeRegistry == nil {
		t.Fatal("engine.State.RoomTypeRegistry is nil")
	}
}
