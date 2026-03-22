package scene

import (
	"testing"
)

func TestTitleScene_NewGameButtonClick(t *testing.T) {
	called := false
	ts := NewTitleScene(1088, 728, func() { called = true }, nil, nil)

	// Click center of the New Game button.
	r := ts.NewGameRect()
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2

	if !ts.HandleClick(cx, cy) {
		t.Error("expected HandleClick to return true for New Game button")
	}
	if !called {
		t.Error("expected onNewGame callback to be called")
	}
}

func TestTitleScene_LoadButtonClick(t *testing.T) {
	called := false
	ts := NewTitleScene(1088, 728, nil, func() { called = true }, nil)

	r := ts.LoadRect()
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2

	if !ts.HandleClick(cx, cy) {
		t.Error("expected HandleClick to return true for Load button")
	}
	if !called {
		t.Error("expected onLoad callback to be called")
	}
}

func TestTitleScene_ClickOutsideButtons(t *testing.T) {
	ts := NewTitleScene(1088, 728, func() { t.Error("should not fire") }, func() { t.Error("should not fire") }, nil)

	if ts.HandleClick(0, 0) {
		t.Error("expected HandleClick to return false for click outside buttons")
	}
}

func TestTitleScene_ButtonsDoNotOverlap(t *testing.T) {
	ts := NewTitleScene(1088, 728, nil, nil, nil)
	ng := ts.NewGameRect()
	ld := ts.LoadRect()

	if ng.Overlaps(ld) {
		t.Errorf("New Game button %v overlaps Load button %v", ng, ld)
	}
}

func TestTitleScene_TransitionToScenarioSelect(t *testing.T) {
	transitioned := false
	sm := NewSceneManager(nil)

	titleScene := NewTitleScene(1088, 728, func() {
		transitioned = true
		sm.Switch(&spyScene{name: "select"})
	}, nil, nil)

	sm.Switch(titleScene)

	// Click the New Game button.
	r := titleScene.NewGameRect()
	titleScene.HandleClick((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)

	if !transitioned {
		t.Error("expected transition to scenario select scene")
	}
	if _, ok := sm.Current().(*spyScene); !ok {
		t.Error("expected current scene to be the select scene spy")
	}
}
