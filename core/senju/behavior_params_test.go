package senju

import (
	"encoding/json"
	"testing"
)

func TestDefaultBehaviorParams(t *testing.T) {
	p := DefaultBehaviorParams()
	if p == nil {
		t.Fatal("DefaultBehaviorParams returned nil")
	}
	if p.FleeHPThreshold != 0.25 {
		t.Errorf("FleeHPThreshold = %v, want 0.25", p.FleeHPThreshold)
	}
	if p.ChaseTimeoutTicks != 10 {
		t.Errorf("ChaseTimeoutTicks = %d, want 10", p.ChaseTimeoutTicks)
	}
	if p.PatrolRestTicks != 3 {
		t.Errorf("PatrolRestTicks = %d, want 3", p.PatrolRestTicks)
	}
}

func TestLoadBehaviorParams(t *testing.T) {
	data := []byte(`{
		"flee_hp_threshold": 0.5,
		"chase_timeout_ticks": 20,
		"patrol_rest_ticks": 5
	}`)

	p, err := LoadBehaviorParams(data)
	if err != nil {
		t.Fatalf("LoadBehaviorParams error: %v", err)
	}
	if p.FleeHPThreshold != 0.5 {
		t.Errorf("FleeHPThreshold = %v, want 0.5", p.FleeHPThreshold)
	}
	if p.ChaseTimeoutTicks != 20 {
		t.Errorf("ChaseTimeoutTicks = %d, want 20", p.ChaseTimeoutTicks)
	}
	if p.PatrolRestTicks != 5 {
		t.Errorf("PatrolRestTicks = %d, want 5", p.PatrolRestTicks)
	}
}

func TestLoadBehaviorParams_InvalidJSON(t *testing.T) {
	_, err := LoadBehaviorParams([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadBehaviorParams_EmbeddedMatchesDefault(t *testing.T) {
	// Verify the embedded JSON produces valid params.
	p, err := LoadBehaviorParams(defaultBehaviorParamsJSON)
	if err != nil {
		t.Fatalf("embedded JSON parse error: %v", err)
	}
	d := DefaultBehaviorParams()
	if p.FleeHPThreshold != d.FleeHPThreshold {
		t.Errorf("FleeHPThreshold mismatch: embedded=%v default=%v", p.FleeHPThreshold, d.FleeHPThreshold)
	}
	if p.ChaseTimeoutTicks != d.ChaseTimeoutTicks {
		t.Errorf("ChaseTimeoutTicks mismatch: embedded=%d default=%d", p.ChaseTimeoutTicks, d.ChaseTimeoutTicks)
	}
	if p.PatrolRestTicks != d.PatrolRestTicks {
		t.Errorf("PatrolRestTicks mismatch: embedded=%d default=%d", p.PatrolRestTicks, d.PatrolRestTicks)
	}
}

func TestBehaviorParams_JSONRoundTrip(t *testing.T) {
	original := &BehaviorParams{
		FleeHPThreshold:   0.3,
		ChaseTimeoutTicks: 15,
		PatrolRestTicks:   4,
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	restored, err := LoadBehaviorParams(data)
	if err != nil {
		t.Fatalf("LoadBehaviorParams error: %v", err)
	}
	if *original != *restored {
		t.Errorf("round-trip mismatch: got %+v, want %+v", restored, original)
	}
}
