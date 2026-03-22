package scene

import (
	"testing"
	"time"
)

func TestLoadScene_SelectEntry(t *testing.T) {
	entries := []LoadEntry{
		{Path: "/tmp/a.json", Filename: "a.json", SavedAt: time.Now(), ScenarioID: "tutorial"},
		{Path: "/tmp/b.json", Filename: "b.json", SavedAt: time.Now(), ScenarioID: "standard"},
	}

	var selected LoadEntry
	ls := NewLoadScene(1088, 728, entries, func(e LoadEntry) { selected = e }, nil, nil)

	// Click center of the first entry button.
	r := ls.ButtonRects()[0]
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2

	if !ls.HandleClick(cx, cy) {
		t.Error("expected HandleClick to return true for save entry button")
	}
	if selected.Filename != "a.json" {
		t.Errorf("selected entry = %q, want a.json", selected.Filename)
	}
}

func TestLoadScene_BackButton(t *testing.T) {
	entries := []LoadEntry{
		{Path: "/tmp/a.json", Filename: "a.json", SavedAt: time.Now(), ScenarioID: "tutorial"},
	}

	backCalled := false
	ls := NewLoadScene(1088, 728, entries, nil, func() { backCalled = true }, nil)

	r := ls.BackRect()
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2

	if !ls.HandleClick(cx, cy) {
		t.Error("expected HandleClick to return true for back button")
	}
	if !backCalled {
		t.Error("expected onBack callback to be called")
	}
}

func TestLoadScene_ClickOutside(t *testing.T) {
	entries := []LoadEntry{
		{Path: "/tmp/a.json", Filename: "a.json", SavedAt: time.Now(), ScenarioID: "tutorial"},
	}

	ls := NewLoadScene(1088, 728, entries,
		func(e LoadEntry) { t.Error("should not fire") },
		func() { t.Error("should not fire") },
		nil,
	)

	if ls.HandleClick(0, 0) {
		t.Error("expected HandleClick to return false for click outside buttons")
	}
}

func TestLoadScene_EmptyEntries(t *testing.T) {
	ls := NewLoadScene(1088, 728, nil, nil, nil, nil)

	if len(ls.Entries()) != 0 {
		t.Errorf("expected empty entries, got %d", len(ls.Entries()))
	}
	if len(ls.ButtonRects()) != 0 {
		t.Errorf("expected no button rects, got %d", len(ls.ButtonRects()))
	}
}

func TestFormatEntryLabel_WithScenarioID(t *testing.T) {
	e := LoadEntry{
		Filename:   "save_20260101_120000.json",
		SavedAt:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		ScenarioID: "tutorial",
	}
	label := FormatEntryLabel(e)
	want := "2026-01-01 12:00:00  [tutorial]"
	if label != want {
		t.Errorf("label = %q, want %q", label, want)
	}
}

func TestFormatEntryLabel_WithoutScenarioID(t *testing.T) {
	e := LoadEntry{
		Filename: "save_20260101_120000.json",
		SavedAt:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	label := FormatEntryLabel(e)
	want := "2026-01-01 12:00:00  save_20260101_120000.json"
	if label != want {
		t.Errorf("label = %q, want %q", label, want)
	}
}
