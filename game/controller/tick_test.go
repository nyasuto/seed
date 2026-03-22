package controller

import (
	"testing"
)

func TestManualMode_NoAutoAdvance(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	if gc.State() != Playing {
		t.Fatalf("initial state = %d, want Playing", gc.State())
	}

	tickBefore := gc.Snapshot().Tick

	// UpdateTick in Playing mode should not advance.
	n, err := gc.UpdateTick()
	if err != nil {
		t.Fatalf("UpdateTick failed: %v", err)
	}
	if n != 0 {
		t.Errorf("UpdateTick in Playing mode advanced %d ticks, want 0", n)
	}

	tickAfter := gc.Snapshot().Tick
	if tickAfter != tickBefore {
		t.Errorf("tick changed from %d to %d without explicit AdvanceTick", tickBefore, tickAfter)
	}
}

func TestFastForward_AdvancesMultipleTicks(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	tickBefore := gc.Snapshot().Tick

	gc.StartFastForward(10)
	if gc.State() != FastForward {
		t.Fatalf("state = %d, want FastForward", gc.State())
	}
	if gc.FastForwardSpeed() != 10 {
		t.Errorf("FastForwardSpeed = %d, want 10", gc.FastForwardSpeed())
	}

	n, err := gc.UpdateTick()
	if err != nil {
		t.Fatalf("UpdateTick failed: %v", err)
	}
	if n != 10 {
		t.Errorf("UpdateTick advanced %d ticks, want 10", n)
	}

	tickAfter := gc.Snapshot().Tick
	if tickAfter != tickBefore+10 {
		t.Errorf("tick = %d, want %d", tickAfter, tickBefore+10)
	}
}

func TestPause_ResumeRestoresProgress(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	// Advance a few ticks first.
	for i := 0; i < 5; i++ {
		if _, err := gc.AdvanceTick(); err != nil {
			t.Fatalf("AdvanceTick failed: %v", err)
		}
	}
	tickBeforePause := gc.Snapshot().Tick

	// Pause.
	gc.TogglePause()
	if gc.State() != Paused {
		t.Fatalf("state = %d, want Paused", gc.State())
	}

	// UpdateTick while paused should not advance.
	n, err := gc.UpdateTick()
	if err != nil {
		t.Fatalf("UpdateTick failed: %v", err)
	}
	if n != 0 {
		t.Errorf("UpdateTick in Paused mode advanced %d ticks, want 0", n)
	}
	if gc.Snapshot().Tick != tickBeforePause {
		t.Errorf("tick changed during pause")
	}

	// Resume.
	gc.TogglePause()
	if gc.State() != Playing {
		t.Fatalf("state = %d, want Playing after resume", gc.State())
	}

	// Manual advance should work again.
	if _, err := gc.AdvanceTick(); err != nil {
		t.Fatalf("AdvanceTick after resume failed: %v", err)
	}
	if gc.Snapshot().Tick != tickBeforePause+1 {
		t.Errorf("tick = %d, want %d after resume + advance", gc.Snapshot().Tick, tickBeforePause+1)
	}
}

func TestFastForward_StopsOnGameOver(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	// Use a large speed so we hit game over within a few UpdateTick calls.
	gc.StartFastForward(100)

	totalAdvanced := 0
	for i := 0; i < 10; i++ {
		n, err := gc.UpdateTick()
		if err != nil {
			t.Fatalf("UpdateTick failed: %v", err)
		}
		totalAdvanced += n
		if gc.State() == GameOver {
			break
		}
	}

	if gc.State() != GameOver {
		t.Fatal("expected GameOver state")
	}

	// After game over, UpdateTick should not advance further.
	n, err := gc.UpdateTick()
	if err != nil {
		t.Fatalf("UpdateTick after GameOver failed: %v", err)
	}
	if n != 0 {
		t.Errorf("UpdateTick after GameOver advanced %d ticks, want 0", n)
	}
}

func TestTogglePause_NoEffectWhenGameOver(t *testing.T) {
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
		if result.Status != 0 { // not Running
			break
		}
	}

	if gc.State() != GameOver {
		t.Fatal("expected GameOver state")
	}

	gc.TogglePause()
	if gc.State() != GameOver {
		t.Errorf("TogglePause changed GameOver state to %d", gc.State())
	}
}

func TestStartFastForward_NoEffectWhenGameOver(t *testing.T) {
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
		if result.Status != 0 {
			break
		}
	}

	gc.StartFastForward(10)
	if gc.State() != GameOver {
		t.Errorf("StartFastForward changed GameOver state to %d", gc.State())
	}
}

func TestStopFastForward_ReturnsToPlaying(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	gc.StartFastForward(5)
	if gc.State() != FastForward {
		t.Fatalf("state = %d, want FastForward", gc.State())
	}

	gc.StopFastForward()
	if gc.State() != Playing {
		t.Errorf("state = %d, want Playing after StopFastForward", gc.State())
	}
	if gc.FastForwardSpeed() != 0 {
		t.Errorf("FastForwardSpeed = %d, want 0 after stop", gc.FastForwardSpeed())
	}
}

func TestStartFastForward_InvalidSpeed(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController failed: %v", err)
	}

	gc.StartFastForward(0)
	if gc.FastForwardSpeed() != DefaultFastForwardSpeed {
		t.Errorf("FastForwardSpeed = %d, want %d for speed=0", gc.FastForwardSpeed(), DefaultFastForwardSpeed)
	}

	gc.StopFastForward()
	gc.StartFastForward(-1)
	if gc.FastForwardSpeed() != DefaultFastForwardSpeed {
		t.Errorf("FastForwardSpeed = %d, want %d for speed=-1", gc.FastForwardSpeed(), DefaultFastForwardSpeed)
	}
}
