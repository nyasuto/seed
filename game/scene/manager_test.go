package scene

import (
	"image"
	"testing"
)

// spyScene records lifecycle calls for testing.
type spyScene struct {
	name        string
	enterCount  int
	exitCount   int
	updateCount int
	drawCount   int
}

func (s *spyScene) Update() error {
	s.updateCount++
	return nil
}

func (s *spyScene) Draw(_ image.Image) {
	s.drawCount++
}

func (s *spyScene) OnEnter() {
	s.enterCount++
}

func (s *spyScene) OnExit() {
	s.exitCount++
}

func TestNewSceneManager_CallsOnEnter(t *testing.T) {
	s := &spyScene{name: "initial"}
	sm := NewSceneManager(s)

	if s.enterCount != 1 {
		t.Errorf("expected OnEnter called once, got %d", s.enterCount)
	}
	if sm.Current() != s {
		t.Error("expected current scene to be initial scene")
	}
}

func TestNewSceneManager_NilInitial(t *testing.T) {
	sm := NewSceneManager(nil)
	if sm.Current() != nil {
		t.Error("expected nil current scene")
	}
	if err := sm.Update(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSwitch_CallsOnExitThenOnEnter(t *testing.T) {
	a := &spyScene{name: "A"}
	b := &spyScene{name: "B"}
	sm := NewSceneManager(a)

	sm.Switch(b)

	if a.exitCount != 1 {
		t.Errorf("expected OnExit called on A once, got %d", a.exitCount)
	}
	if b.enterCount != 1 {
		t.Errorf("expected OnEnter called on B once, got %d", b.enterCount)
	}
	if sm.Current() != b {
		t.Error("expected current scene to be B after switch")
	}
}

func TestSwitch_DelegatesUpdateAndDraw(t *testing.T) {
	a := &spyScene{name: "A"}
	b := &spyScene{name: "B"}
	sm := NewSceneManager(a)

	// Update/Draw go to A.
	sm.Update()
	sm.Draw(nil)
	if a.updateCount != 1 {
		t.Errorf("expected A.Update called once, got %d", a.updateCount)
	}
	if a.drawCount != 1 {
		t.Errorf("expected A.Draw called once, got %d", a.drawCount)
	}

	// Switch to B.
	sm.Switch(b)

	// Update/Draw should go to B now.
	sm.Update()
	sm.Draw(nil)
	if b.updateCount != 1 {
		t.Errorf("expected B.Update called once, got %d", b.updateCount)
	}
	if b.drawCount != 1 {
		t.Errorf("expected B.Draw called once, got %d", b.drawCount)
	}
	if a.updateCount != 1 {
		t.Errorf("expected A.Update still 1, got %d", a.updateCount)
	}
	if a.drawCount != 1 {
		t.Errorf("expected A.Draw still 1, got %d", a.drawCount)
	}
}

func TestSwitch_ToNil(t *testing.T) {
	a := &spyScene{name: "A"}
	sm := NewSceneManager(a)

	sm.Switch(nil)

	if a.exitCount != 1 {
		t.Errorf("expected OnExit called on A, got %d", a.exitCount)
	}
	if sm.Current() != nil {
		t.Error("expected nil current scene")
	}
	if err := sm.Update(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMultipleSwitches(t *testing.T) {
	a := &spyScene{name: "A"}
	b := &spyScene{name: "B"}
	c := &spyScene{name: "C"}
	sm := NewSceneManager(a)

	sm.Switch(b)
	sm.Switch(c)

	if a.exitCount != 1 {
		t.Errorf("expected A.OnExit once, got %d", a.exitCount)
	}
	if b.enterCount != 1 {
		t.Errorf("expected B.OnEnter once, got %d", b.enterCount)
	}
	if b.exitCount != 1 {
		t.Errorf("expected B.OnExit once, got %d", b.exitCount)
	}
	if c.enterCount != 1 {
		t.Errorf("expected C.OnEnter once, got %d", c.enterCount)
	}
	if sm.Current() != c {
		t.Error("expected current scene to be C")
	}
}
