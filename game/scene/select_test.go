package scene

import (
	"testing"
)

func testEntries() []ScenarioEntry {
	return []ScenarioEntry{
		{
			ID:          "tutorial",
			Name:        "チュートリアル",
			Description: "基本操作を学ぶための簡単なシナリオ",
			Difficulty:  "easy",
			Data:        []byte(`{}`),
		},
		{
			ID:          "standard",
			Name:        "標準シナリオ",
			Description: "中規模マップでの本格的な洞窟経営シナリオ",
			Difficulty:  "normal",
			Data:        []byte(`{}`),
		},
	}
}

func TestScenarioSelectScene_SelectScenario(t *testing.T) {
	entries := testEntries()
	var selected ScenarioEntry
	ss := NewScenarioSelectScene(1088, 728, entries, func(e ScenarioEntry) {
		selected = e
	}, nil, nil)

	// Verify buttons are created for each entry.
	rects := ss.ButtonRects()
	if len(rects) != len(entries) {
		t.Fatalf("expected %d button rects, got %d", len(entries), len(rects))
	}

	// Click the first scenario button.
	r := rects[0]
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2
	if !ss.HandleClick(cx, cy) {
		t.Error("expected HandleClick to return true for first scenario button")
	}

	if selected.ID != "tutorial" {
		t.Errorf("expected selected scenario ID 'tutorial', got %q", selected.ID)
	}
}

func TestScenarioSelectScene_SelectCreatesInGameScene(t *testing.T) {
	entries := testEntries()
	sm := NewSceneManager(nil)

	var selectedEntry ScenarioEntry
	ss := NewScenarioSelectScene(1088, 728, entries, func(e ScenarioEntry) {
		selectedEntry = e
		sm.Switch(&spyScene{name: "ingame-" + e.ID})
	}, nil, nil)

	sm.Switch(ss)

	// Click the second scenario button.
	rects := ss.ButtonRects()
	r := rects[1]
	ss.HandleClick((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)

	if selectedEntry.ID != "standard" {
		t.Errorf("expected selected 'standard', got %q", selectedEntry.ID)
	}
	spy, ok := sm.Current().(*spyScene)
	if !ok {
		t.Fatal("expected current scene to be spyScene")
	}
	if spy.name != "ingame-standard" {
		t.Errorf("expected scene name 'ingame-standard', got %q", spy.name)
	}
}

func TestScenarioSelectScene_BackButton(t *testing.T) {
	entries := testEntries()
	backCalled := false
	ss := NewScenarioSelectScene(1088, 728, entries, nil, func() {
		backCalled = true
	}, nil)

	r := ss.BackRect()
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2

	if !ss.HandleClick(cx, cy) {
		t.Error("expected HandleClick to return true for back button")
	}
	if !backCalled {
		t.Error("expected onBack callback to be called")
	}
}

func TestScenarioSelectScene_ClickOutside(t *testing.T) {
	entries := testEntries()
	ss := NewScenarioSelectScene(1088, 728, entries, func(e ScenarioEntry) {
		t.Error("should not fire")
	}, func() {
		t.Error("should not fire")
	}, nil)

	if ss.HandleClick(0, 0) {
		t.Error("expected HandleClick to return false for click outside")
	}
}

func TestScenarioSelectScene_ButtonsMatchEntries(t *testing.T) {
	entries := testEntries()
	ss := NewScenarioSelectScene(1088, 728, entries, nil, nil, nil)

	if len(ss.ButtonRects()) != len(entries) {
		t.Errorf("expected %d button rects, got %d", len(entries), len(ss.ButtonRects()))
	}

	stored := ss.Entries()
	for i, e := range entries {
		if stored[i].ID != e.ID {
			t.Errorf("entry[%d]: expected ID %q, got %q", i, e.ID, stored[i].ID)
		}
	}
}

func TestScenarioSelectScene_EmptyEntries(t *testing.T) {
	ss := NewScenarioSelectScene(1088, 728, nil, nil, nil, nil)

	if len(ss.ButtonRects()) != 0 {
		t.Errorf("expected 0 button rects for empty entries, got %d", len(ss.ButtonRects()))
	}
}
