package input

import "testing"

func TestActionMode_String(t *testing.T) {
	tests := []struct {
		mode ActionMode
		want string
	}{
		{ModeNormal, "Normal"},
		{ModeDigRoom, "DigRoom"},
		{ModeDigCorridor, "DigCorridor"},
		{ModeSummon, "Summon"},
		{ModeUpgrade, "Upgrade"},
		{ActionMode(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("ActionMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}
