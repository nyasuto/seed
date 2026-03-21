package scenario

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestMarshalProgress_NilReturnsError(t *testing.T) {
	_, err := MarshalProgress(nil)
	if err == nil {
		t.Fatal("expected error for nil progress, got nil")
	}
}

func TestMarshalUnmarshalProgress_RoundTrip(t *testing.T) {
	original := &ScenarioProgress{
		ScenarioID:  "standard",
		CurrentTick: 150,
		FiredEventIDs: []string{"evt_tutorial_start", "evt_first_wave"},
		WaveResults: []WaveResult{
			{WaveID: 0, Result: "victory", CompletedTick: 100},
			{WaveID: 1, Result: "defeat", CompletedTick: 140},
		},
		CoreHP: 85,
	}

	data, err := MarshalProgress(original)
	if err != nil {
		t.Fatalf("MarshalProgress: %v", err)
	}

	restored, err := UnmarshalProgress(data)
	if err != nil {
		t.Fatalf("UnmarshalProgress: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("round-trip mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
	}
}

func TestMarshalUnmarshalProgress_EmptySlices(t *testing.T) {
	original := &ScenarioProgress{
		ScenarioID:    "empty_test",
		CurrentTick:   0,
		FiredEventIDs: nil,
		WaveResults:   nil,
		CoreHP:        100,
	}

	data, err := MarshalProgress(original)
	if err != nil {
		t.Fatalf("MarshalProgress: %v", err)
	}

	restored, err := UnmarshalProgress(data)
	if err != nil {
		t.Fatalf("UnmarshalProgress: %v", err)
	}

	// nil slices become empty slices after JSON round-trip; normalize for comparison
	if restored.ScenarioID != original.ScenarioID {
		t.Errorf("ScenarioID: got %q, want %q", restored.ScenarioID, original.ScenarioID)
	}
	if restored.CurrentTick != original.CurrentTick {
		t.Errorf("CurrentTick: got %d, want %d", restored.CurrentTick, original.CurrentTick)
	}
	if restored.CoreHP != original.CoreHP {
		t.Errorf("CoreHP: got %d, want %d", restored.CoreHP, original.CoreHP)
	}
	if len(restored.FiredEventIDs) != 0 {
		t.Errorf("FiredEventIDs: got %v, want empty", restored.FiredEventIDs)
	}
	if len(restored.WaveResults) != 0 {
		t.Errorf("WaveResults: got %v, want empty", restored.WaveResults)
	}
}

func TestUnmarshalProgress_InvalidJSON(t *testing.T) {
	_, err := UnmarshalProgress([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestMarshalProgress_JSONStructure(t *testing.T) {
	p := &ScenarioProgress{
		ScenarioID:    "test_scenario",
		CurrentTick:   types.Tick(42),
		FiredEventIDs: []string{"e1"},
		WaveResults: []WaveResult{
			{WaveID: 0, Result: "victory", CompletedTick: 30},
		},
		CoreHP: 90,
	}

	data, err := MarshalProgress(p)
	if err != nil {
		t.Fatalf("MarshalProgress: %v", err)
	}

	// Verify JSON keys match expected names
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	expectedKeys := []string{"scenario_id", "current_tick", "fired_event_ids", "wave_results", "core_hp"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("missing expected JSON key %q", key)
		}
	}
}
