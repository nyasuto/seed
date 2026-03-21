package simulation

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestGameStatus_String(t *testing.T) {
	tests := []struct {
		status GameStatus
		want   string
	}{
		{Running, "Running"},
		{Won, "Won"},
		{Lost, "Lost"},
		{GameStatus(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("GameStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestGameResult_Fields(t *testing.T) {
	r := GameResult{
		Status:    Won,
		FinalTick: types.Tick(150),
		Reason:    "all invaders defeated",
	}
	if r.Status != Won {
		t.Errorf("Status = %v, want Won", r.Status)
	}
	if r.FinalTick != 150 {
		t.Errorf("FinalTick = %d, want 150", r.FinalTick)
	}
	if r.Reason != "all invaders defeated" {
		t.Errorf("Reason = %q, want %q", r.Reason, "all invaders defeated")
	}
}

func TestGameResult_RunningDefault(t *testing.T) {
	var r GameResult
	if r.Status != Running {
		t.Errorf("zero-value Status = %v, want Running", r.Status)
	}
	if r.FinalTick != 0 {
		t.Errorf("zero-value FinalTick = %d, want 0", r.FinalTick)
	}
	if r.Reason != "" {
		t.Errorf("zero-value Reason = %q, want empty", r.Reason)
	}
}

func TestGameStatus_IotaValues(t *testing.T) {
	// Ensure iota ordering is Running=0, Won=1, Lost=2
	if Running != 0 {
		t.Errorf("Running = %d, want 0", Running)
	}
	if Won != 1 {
		t.Errorf("Won = %d, want 1", Won)
	}
	if Lost != 2 {
		t.Errorf("Lost = %d, want 2", Lost)
	}
}
