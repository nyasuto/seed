package view

import (
	"testing"

	"github.com/nyasuto/seed/game/input"
)

func TestActionBar_HandleClick_ActionButton(t *testing.T) {
	ab := NewActionBar(728)

	// Click on first action button (Dig D).
	btn := ab.actionButtons[0]
	cx := (btn.Rect.Min.X + btn.Rect.Max.X) / 2
	cy := (btn.Rect.Min.Y + btn.Rect.Max.Y) / 2

	actionHit, mode, tickHit, _ := ab.HandleClick(cx, cy)
	if !actionHit {
		t.Fatal("expected action hit for Dig button click")
	}
	if mode != input.ModeDigRoom {
		t.Errorf("mode = %v, want ModeDigRoom", mode)
	}
	if tickHit {
		t.Error("unexpected tick hit")
	}
}

func TestActionBar_HandleClick_AllActionButtons(t *testing.T) {
	ab := NewActionBar(728)

	expected := []input.ActionMode{
		input.ModeDigRoom,
		input.ModeDigCorridor,
		input.ModeSummon,
		input.ModeUpgrade,
		input.ModeNormal, // Wait
	}

	for i, btn := range ab.actionButtons {
		cx := (btn.Rect.Min.X + btn.Rect.Max.X) / 2
		cy := (btn.Rect.Min.Y + btn.Rect.Max.Y) / 2

		actionHit, mode, _, _ := ab.HandleClick(cx, cy)
		if !actionHit {
			t.Errorf("button %d: expected action hit", i)
		}
		if mode != expected[i] {
			t.Errorf("button %d: mode = %v, want %v", i, mode, expected[i])
		}
	}
}

func TestActionBar_HandleClick_TickButton(t *testing.T) {
	ab := NewActionBar(728)

	// Click on first tick button (Play).
	btn := ab.tickButtons[0]
	cx := (btn.Rect.Min.X + btn.Rect.Max.X) / 2
	cy := (btn.Rect.Min.Y + btn.Rect.Max.Y) / 2

	actionHit, _, tickHit, tmode := ab.HandleClick(cx, cy)
	if actionHit {
		t.Error("unexpected action hit")
	}
	if !tickHit {
		t.Fatal("expected tick hit for Play button click")
	}
	if tmode != TickManual {
		t.Errorf("tmode = %v, want TickManual", tmode)
	}
}

func TestActionBar_HandleClick_Miss(t *testing.T) {
	ab := NewActionBar(728)

	// Click outside all buttons.
	actionHit, _, tickHit, _ := ab.HandleClick(0, 0)
	if actionHit || tickHit {
		t.Error("expected no hit for click at (0,0)")
	}
}

func TestActionBar_HandleClick_OutsideBar(t *testing.T) {
	ab := NewActionBar(728)

	// Click in the middle of the screen (well above the bar).
	actionHit, _, tickHit, _ := ab.HandleClick(400, 400)
	if actionHit || tickHit {
		t.Error("expected no hit for click above bar")
	}
}

func TestActionBar_BarY(t *testing.T) {
	ab := NewActionBar(728)
	expected := 728 - actionBarHeight
	if ab.BarY() != expected {
		t.Errorf("BarY() = %d, want %d", ab.BarY(), expected)
	}
}
