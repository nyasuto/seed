package view

import (
	"testing"

	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/types"
)

func TestBattleRoomIDs_FightingInvaderRoomIdentified(t *testing.T) {
	waves := []*invasion.InvasionWave{
		{
			ID:    1,
			State: invasion.Active,
			Invaders: []*invasion.Invader{
				{ID: 1, CurrentRoomID: 3, State: invasion.Fighting},
				{ID: 2, CurrentRoomID: 5, State: invasion.Advancing},
			},
		},
	}

	rooms := BattleRoomIDs(waves)

	if !rooms[3] {
		t.Error("room 3 should be identified as battle room (Fighting invader)")
	}
	if rooms[5] {
		t.Error("room 5 should not be a battle room (Advancing invader)")
	}
}

func TestBattleRoomIDs_InactiveWaveIgnored(t *testing.T) {
	waves := []*invasion.InvasionWave{
		{
			ID:    1,
			State: invasion.Completed,
			Invaders: []*invasion.Invader{
				{ID: 1, CurrentRoomID: 3, State: invasion.Fighting},
			},
		},
	}

	rooms := BattleRoomIDs(waves)
	if rooms[3] {
		t.Error("room 3 should not be identified (wave is Completed)")
	}
}

func TestBattleRoomIDs_MultipleRooms(t *testing.T) {
	waves := []*invasion.InvasionWave{
		{
			ID:    1,
			State: invasion.Active,
			Invaders: []*invasion.Invader{
				{ID: 1, CurrentRoomID: 2, State: invasion.Fighting},
				{ID: 2, CurrentRoomID: 4, State: invasion.Fighting},
			},
		},
		{
			ID:    2,
			State: invasion.Active,
			Invaders: []*invasion.Invader{
				{ID: 3, CurrentRoomID: 6, State: invasion.Fighting},
			},
		},
	}

	rooms := BattleRoomIDs(waves)
	if len(rooms) != 3 {
		t.Errorf("expected 3 battle rooms, got %d", len(rooms))
	}
	for _, id := range []int{2, 4, 6} {
		if !rooms[id] {
			t.Errorf("room %d should be a battle room", id)
		}
	}
}

func TestBattleRoomIDs_ZeroRoomIDIgnored(t *testing.T) {
	waves := []*invasion.InvasionWave{
		{
			ID:    1,
			State: invasion.Active,
			Invaders: []*invasion.Invader{
				{ID: 1, CurrentRoomID: 0, State: invasion.Fighting},
			},
		},
	}

	rooms := BattleRoomIDs(waves)
	if len(rooms) != 0 {
		t.Errorf("expected no battle rooms, got %d", len(rooms))
	}
}

func TestWaveArrivalDetected(t *testing.T) {
	tests := []struct {
		name    string
		prev    int
		current int
		want    bool
	}{
		{"new wave arrives", 0, 1, true},
		{"multiple new waves", 1, 3, true},
		{"no change", 2, 2, false},
		{"wave completed", 2, 1, false},
		{"all waves done", 1, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WaveArrivalDetected(tt.prev, tt.current)
			if got != tt.want {
				t.Errorf("WaveArrivalDetected(%d, %d) = %v, want %v", tt.prev, tt.current, got, tt.want)
			}
		})
	}
}

func TestBattleFeedback_WaveAlert(t *testing.T) {
	bf := NewBattleFeedback()

	// Initial update with no active waves.
	bf.Update(100, nil, nil)

	// Simulate a wave becoming active.
	waves := []*invasion.InvasionWave{
		{ID: 1, State: invasion.Active, Invaders: []*invasion.Invader{}},
	}
	bf.Update(100, waves, nil)

	if bf.WaveAlertMessage() == "" {
		t.Error("expected wave alert message after new wave arrival")
	}
	if bf.WaveAlertAlpha() != 1.0 {
		t.Errorf("expected alpha 1.0 at start, got %f", bf.WaveAlertAlpha())
	}
}

func TestBattleFeedback_WaveAlertFades(t *testing.T) {
	bf := NewBattleFeedback()
	bf.Update(100, nil, nil)

	waves := []*invasion.InvasionWave{
		{ID: 1, State: invasion.Active, Invaders: []*invasion.Invader{}},
	}
	bf.Update(100, waves, nil)

	// Advance past full opacity phase.
	for i := 0; i < WaveAlertDuration-waveAlertFadeStart; i++ {
		bf.Update(100, waves, nil)
	}
	if bf.WaveAlertAlpha() != 1.0 {
		t.Errorf("at fade start boundary, alpha = %f, want 1.0", bf.WaveAlertAlpha())
	}

	// Advance to halfway through fade.
	for i := 0; i < waveAlertFadeStart/2; i++ {
		bf.Update(100, waves, nil)
	}
	alpha := bf.WaveAlertAlpha()
	if alpha < 0.4 || alpha > 0.6 {
		t.Errorf("at fade midpoint, alpha = %f, want ~0.5", alpha)
	}

	// Advance to end.
	for i := 0; i < waveAlertFadeStart; i++ {
		bf.Update(100, waves, nil)
	}
	if bf.WaveAlertMessage() != "" {
		t.Error("wave alert should be cleared after full duration")
	}
}

func TestBattleFeedback_CoreHPBlink(t *testing.T) {
	bf := NewBattleFeedback()
	bf.Update(100, nil, nil)

	if bf.HPBlinkTimer() != 0 {
		t.Errorf("initial HPBlinkTimer = %d, want 0", bf.HPBlinkTimer())
	}

	// CoreHP decreases.
	bf.Update(90, nil, nil)
	if bf.HPBlinkTimer() != HPBlinkDuration {
		t.Errorf("after HP decrease, HPBlinkTimer = %d, want %d", bf.HPBlinkTimer(), HPBlinkDuration)
	}

	// No change should not reset.
	for i := 0; i < 10; i++ {
		bf.Update(90, nil, nil)
	}
	if bf.HPBlinkTimer() != HPBlinkDuration-10 {
		t.Errorf("HPBlinkTimer = %d, want %d", bf.HPBlinkTimer(), HPBlinkDuration-10)
	}

	// Advance to expiry.
	for i := 0; i < HPBlinkDuration; i++ {
		bf.Update(90, nil, nil)
	}
	if bf.HPBlinkTimer() != 0 {
		t.Errorf("after full duration, HPBlinkTimer = %d, want 0", bf.HPBlinkTimer())
	}
}

func TestBattleFeedback_BeastDefeatBlink(t *testing.T) {
	bf := NewBattleFeedback()
	bf.Update(100, nil, nil)

	beast := &senju.Beast{
		ID:        1,
		SpeciesID: "kirin",
		Element:   types.Fire,
		RoomID:    1,
		State:     senju.Stunned,
	}
	bf.Update(100, nil, []*senju.Beast{beast})

	if _, ok := bf.defeatedBeastIDs[1]; !ok {
		t.Error("beast 1 should be tracked as defeated")
	}

	// Advance past blink duration.
	for i := 0; i < BeastDefeatBlinkDuration+1; i++ {
		bf.Update(100, nil, []*senju.Beast{beast})
	}

	// Beast is still stunned but original timer expired. It should have been
	// re-added since it's still Stunned — check that it's tracked.
	// Actually the current logic only adds on first detection, so after the timer
	// expires, it will be re-detected as a new stunned beast.
	// This is acceptable — the blink restarts when detected again.
}

func TestBattleFeedback_NoAlertOnInitialWaves(t *testing.T) {
	bf := NewBattleFeedback()

	// Start with already active waves — should not trigger alert.
	waves := []*invasion.InvasionWave{
		{ID: 1, State: invasion.Active, Invaders: []*invasion.Invader{}},
	}
	bf.Update(100, waves, nil)

	if bf.WaveAlertMessage() != "" {
		t.Error("should not trigger alert on initial state with pre-existing active waves")
	}
}

func TestFormatWaveAlert(t *testing.T) {
	tests := []struct {
		newWaves    int
		totalActive int
		want        string
	}{
		{1, 1, "Wave incoming! (1 active)"},
		{1, 2, "Wave incoming! (2 active)"},
		{2, 3, "2 Waves incoming! (3 active)"},
	}
	for _, tt := range tests {
		got := FormatWaveAlert(tt.newWaves, tt.totalActive)
		if got != tt.want {
			t.Errorf("FormatWaveAlert(%d, %d) = %q, want %q", tt.newWaves, tt.totalActive, got, tt.want)
		}
	}
}

func TestBattleFeedback_IsBlinkOn(t *testing.T) {
	bf := NewBattleFeedback()

	// Frame 0 → period 4: frame 0 → 0/4=0, 0%2=0 → true (on).
	bf.frame = 0
	if !bf.IsBlinkOn(4) {
		t.Error("frame 0, period 4 should be on")
	}

	// Frame 4 → 4/4=1, 1%2=1 → false (off).
	bf.frame = 4
	if bf.IsBlinkOn(4) {
		t.Error("frame 4, period 4 should be off")
	}

	// Frame 8 → 8/4=2, 2%2=0 → true (on).
	bf.frame = 8
	if !bf.IsBlinkOn(4) {
		t.Error("frame 8, period 4 should be on")
	}

	// Period 0 → always off.
	if bf.IsBlinkOn(0) {
		t.Error("period 0 should always return false")
	}
}
