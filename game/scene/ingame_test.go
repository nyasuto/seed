package scene

import (
	_ "embed"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/game/controller"
	"github.com/nyasuto/seed/game/view"
)

//go:embed testdata/tutorial.json
var tutorialJSON []byte

func newTestController(t *testing.T) *controller.GameController {
	t.Helper()
	gc, err := controller.NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}
	return gc
}

func TestNewInGameScene_CreatesComponents(t *testing.T) {
	ctrl := newTestController(t)

	s := NewInGameScene(InGameConfig{
		Controller:   ctrl,
		ScreenWidth:  1088,
		ScreenHeight: 728,
		MapOffsetX:   32,
		MapOffsetY:   32,
	})

	if s.ctrl != ctrl {
		t.Error("controller not set")
	}
	if s.mapView == nil {
		t.Error("mapView not initialized")
	}
	if s.entity == nil {
		t.Error("entity renderer not initialized")
	}
	if s.stateMachine == nil {
		t.Error("state machine not initialized")
	}
	if s.actionBar == nil {
		t.Error("action bar not initialized")
	}
	if s.feedback == nil {
		t.Error("feedback overlay not initialized")
	}
	if s.provider == nil {
		t.Error("tileset provider not initialized")
	}
}

func TestInGameScene_GameOverCallback(t *testing.T) {
	ctrl := newTestController(t)

	var gotResult simulation.GameResult
	var callCount int

	s := NewInGameScene(InGameConfig{
		Controller:   ctrl,
		ScreenWidth:  1088,
		ScreenHeight: 728,
		MapOffsetX:   32,
		MapOffsetY:   32,
		OnGameOver: func(result simulation.GameResult, _ scenario.GameSnapshot) {
			gotResult = result
			callCount++
		},
	})

	// Advance game to completion.
	for i := 0; i < 350; i++ {
		_, err := ctrl.AdvanceTick()
		if err != nil {
			t.Fatalf("AdvanceTick failed at tick %d: %v", i, err)
		}
		if ctrl.State() == controller.GameOver {
			break
		}
	}

	if ctrl.State() != controller.GameOver {
		t.Fatal("expected game to reach GameOver within 350 ticks")
	}

	// Simulate an Update call — this should trigger the onGameOver callback.
	// We cannot call the full Update (which uses ebiten input), but we can
	// verify the detection logic directly: when state is GameOver, the callback fires.
	// Replicate the detection logic from Update:
	if ctrl.State() == controller.GameOver && !s.gameOverNotified && s.onGameOver != nil {
		s.gameOverNotified = true
		s.onGameOver(ctrl.Result(), ctrl.Snapshot())
	}

	if callCount != 1 {
		t.Errorf("onGameOver called %d times, want 1", callCount)
	}
	if gotResult.Status == simulation.Running {
		t.Error("expected non-Running game result status")
	}

	// Verify callback is not called again.
	if ctrl.State() == controller.GameOver && !s.gameOverNotified && s.onGameOver != nil {
		s.onGameOver(ctrl.Result(), ctrl.Snapshot())
	}
	if callCount != 1 {
		t.Errorf("onGameOver called %d times after second check, want 1", callCount)
	}
}

func TestInGameScene_ControllerAccessor(t *testing.T) {
	ctrl := newTestController(t)
	s := NewInGameScene(InGameConfig{
		Controller:   ctrl,
		ScreenWidth:  1088,
		ScreenHeight: 728,
		MapOffsetX:   32,
		MapOffsetY:   32,
	})

	if s.Controller() != ctrl {
		t.Error("Controller() should return the same controller passed in config")
	}
}

func TestInGameScene_ImplementsScene(t *testing.T) {
	ctrl := newTestController(t)
	s := NewInGameScene(InGameConfig{
		Controller:   ctrl,
		ScreenWidth:  1088,
		ScreenHeight: 728,
		MapOffsetX:   32,
		MapOffsetY:   32,
	})

	// Verify InGameScene implements the Scene interface at compile time.
	var _ Scene = s
}

func TestInGameScene_CurrentTickMode(t *testing.T) {
	ctrl := newTestController(t)
	s := NewInGameScene(InGameConfig{
		Controller:   ctrl,
		ScreenWidth:  1088,
		ScreenHeight: 728,
		MapOffsetX:   32,
		MapOffsetY:   32,
	})

	// Default state is Playing → TickManual.
	if got := s.currentTickMode(); got != view.TickManual {
		t.Errorf("currentTickMode() = %d, want TickManual", got)
	}

	// FastForward.
	ctrl.StartFastForward(5)
	if got := s.currentTickMode(); got != view.TickFastForward {
		t.Errorf("currentTickMode() in FF = %d, want TickFastForward", got)
	}

	// Paused.
	ctrl.StopFastForward()
	ctrl.TogglePause()
	if got := s.currentTickMode(); got != view.TickPaused {
		t.Errorf("currentTickMode() paused = %d, want TickPaused", got)
	}
}
