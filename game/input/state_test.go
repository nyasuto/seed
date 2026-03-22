package input

import "testing"

func TestNewInputStateMachine_StartsNormal(t *testing.T) {
	sm := NewInputStateMachine()
	if sm.Mode() != ModeNormal {
		t.Errorf("expected ModeNormal, got %v", sm.Mode())
	}
}

func TestSetMode_ToDigRoom(t *testing.T) {
	sm := NewInputStateMachine()
	sm.SetMode(ModeDigRoom)
	if sm.Mode() != ModeDigRoom {
		t.Errorf("expected ModeDigRoom, got %v", sm.Mode())
	}
}

func TestSetMode_EscapeToNormal(t *testing.T) {
	// Simulates: user is in ModeDigRoom, presses Escape → ModeNormal
	sm := NewInputStateMachine()
	sm.SetMode(ModeDigRoom)
	sm.SetMode(ModeNormal)
	if sm.Mode() != ModeNormal {
		t.Errorf("expected ModeNormal after cancel, got %v", sm.Mode())
	}
}

func TestSetMode_SwitchFromDigRoomToSummon(t *testing.T) {
	// Simulates: user in ModeDigRoom, presses 'S' → ModeSummon
	sm := NewInputStateMachine()
	sm.SetMode(ModeDigRoom)
	if sm.Mode() != ModeDigRoom {
		t.Fatalf("expected ModeDigRoom, got %v", sm.Mode())
	}
	sm.SetMode(ModeSummon)
	if sm.Mode() != ModeSummon {
		t.Errorf("expected ModeSummon, got %v", sm.Mode())
	}
}

func TestSetMode_AllModes(t *testing.T) {
	tests := []struct {
		name string
		mode ActionMode
	}{
		{"Normal", ModeNormal},
		{"DigRoom", ModeDigRoom},
		{"DigCorridor", ModeDigCorridor},
		{"Summon", ModeSummon},
		{"Upgrade", ModeUpgrade},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewInputStateMachine()
			sm.SetMode(tt.mode)
			if sm.Mode() != tt.mode {
				t.Errorf("expected %v, got %v", tt.mode, sm.Mode())
			}
		})
	}
}

func TestSetMode_PreservesLastSet(t *testing.T) {
	sm := NewInputStateMachine()
	sm.SetMode(ModeDigRoom)
	sm.SetMode(ModeUpgrade)
	sm.SetMode(ModeDigCorridor)
	if sm.Mode() != ModeDigCorridor {
		t.Errorf("expected ModeDigCorridor, got %v", sm.Mode())
	}
}
