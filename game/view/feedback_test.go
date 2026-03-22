package view

import (
	"testing"

	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/input"
)

func TestFeedbackOverlay_ErrorTimer_ThreeSeconds(t *testing.T) {
	fo := NewFeedbackOverlay()
	fo.ShowError("test error")

	if fo.ErrorTimer() != ErrorDuration {
		t.Fatalf("ErrorTimer() = %d, want %d", fo.ErrorTimer(), ErrorDuration)
	}
	if fo.ErrorMessage() != "test error" {
		t.Fatalf("ErrorMessage() = %q, want %q", fo.ErrorMessage(), "test error")
	}

	// Simulate 180 frames (3 seconds at 60 FPS).
	for i := 0; i < ErrorDuration; i++ {
		fo.Update()
	}

	if fo.ErrorMessage() != "" {
		t.Errorf("after %d frames, ErrorMessage() = %q, want empty", ErrorDuration, fo.ErrorMessage())
	}
	if fo.ErrorTimer() != 0 {
		t.Errorf("after %d frames, ErrorTimer() = %d, want 0", ErrorDuration, fo.ErrorTimer())
	}
}

func TestFeedbackOverlay_ErrorAlpha_FullOpacityBeforeFade(t *testing.T) {
	fo := NewFeedbackOverlay()
	fo.ShowError("visible")

	// At the start (timer=180), alpha should be 1.0.
	if alpha := fo.ErrorAlpha(); alpha != 1.0 {
		t.Errorf("initial ErrorAlpha() = %f, want 1.0", alpha)
	}

	// After 120 frames (timer=60 = errorFadeStart), alpha should still be 1.0.
	for i := 0; i < ErrorDuration-errorFadeStart; i++ {
		fo.Update()
	}
	if alpha := fo.ErrorAlpha(); alpha != 1.0 {
		t.Errorf("at fade start boundary, ErrorAlpha() = %f, want 1.0", alpha)
	}
}

func TestFeedbackOverlay_ErrorAlpha_FadesOut(t *testing.T) {
	fo := NewFeedbackOverlay()
	fo.ShowError("fading")

	// Advance to 30 frames remaining (halfway through fade).
	for i := 0; i < ErrorDuration-30; i++ {
		fo.Update()
	}

	alpha := fo.ErrorAlpha()
	expected := 30.0 / float64(errorFadeStart) // 0.5
	if alpha < expected-0.01 || alpha > expected+0.01 {
		t.Errorf("ErrorAlpha() at 30 frames remaining = %f, want ~%f", alpha, expected)
	}
}

func TestFeedbackOverlay_ErrorAlpha_ZeroWhenExpired(t *testing.T) {
	fo := NewFeedbackOverlay()
	// No error set.
	if alpha := fo.ErrorAlpha(); alpha != 0 {
		t.Errorf("ErrorAlpha() with no error = %f, want 0", alpha)
	}
}

func TestFeedbackOverlay_ShowError_ResetsTimer(t *testing.T) {
	fo := NewFeedbackOverlay()
	fo.ShowError("first")

	// Advance halfway.
	for i := 0; i < 90; i++ {
		fo.Update()
	}

	fo.ShowError("second")
	if fo.ErrorTimer() != ErrorDuration {
		t.Errorf("after reset, ErrorTimer() = %d, want %d", fo.ErrorTimer(), ErrorDuration)
	}
	if fo.ErrorMessage() != "second" {
		t.Errorf("after reset, ErrorMessage() = %q, want %q", fo.ErrorMessage(), "second")
	}
}

func TestModeLabel(t *testing.T) {
	tests := []struct {
		mode input.ActionMode
		want string
	}{
		{input.ModeNormal, ""},
		{input.ModeDigRoom, "掘削モード"},
		{input.ModeDigCorridor, "通路モード"},
		{input.ModeSummon, "召喚モード"},
		{input.ModeUpgrade, "強化モード"},
	}
	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			got := ModeLabel(tt.mode)
			if got != tt.want {
				t.Errorf("ModeLabel(%v) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

func TestCellHighlightFor_DigRoom(t *testing.T) {
	tests := []struct {
		name     string
		cellType world.CellType
		want     CellHighlight
	}{
		{"rock is valid", world.Rock, CellValid},
		{"hard rock is invalid", world.HardRock, CellInvalid},
		{"water is invalid", world.Water, CellInvalid},
		{"room floor is none", world.RoomFloor, CellNone},
		{"corridor is none", world.CorridorFloor, CellNone},
		{"entrance is none", world.Entrance, CellNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CellHighlightFor(input.ModeDigRoom, tt.cellType)
			if got != tt.want {
				t.Errorf("CellHighlightFor(ModeDigRoom, %v) = %v, want %v", tt.cellType, got, tt.want)
			}
		})
	}
}

func TestCellHighlightFor_RoomTargetModes(t *testing.T) {
	modes := []input.ActionMode{input.ModeDigCorridor, input.ModeSummon, input.ModeUpgrade}
	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			if got := CellHighlightFor(mode, world.RoomFloor); got != CellValid {
				t.Errorf("CellHighlightFor(%v, RoomFloor) = %v, want CellValid", mode, got)
			}
			if got := CellHighlightFor(mode, world.Entrance); got != CellValid {
				t.Errorf("CellHighlightFor(%v, Entrance) = %v, want CellValid", mode, got)
			}
			if got := CellHighlightFor(mode, world.Rock); got != CellNone {
				t.Errorf("CellHighlightFor(%v, Rock) = %v, want CellNone", mode, got)
			}
		})
	}
}

func TestCellHighlightFor_NormalMode(t *testing.T) {
	cellTypes := []world.CellType{world.Rock, world.RoomFloor, world.HardRock, world.Water, world.Entrance}
	for _, ct := range cellTypes {
		if got := CellHighlightFor(input.ModeNormal, ct); got != CellNone {
			t.Errorf("CellHighlightFor(ModeNormal, %v) = %v, want CellNone", ct, got)
		}
	}
}
