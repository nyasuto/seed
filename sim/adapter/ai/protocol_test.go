package ai

import (
	"encoding/json"
	"testing"
)

func TestStateMessage_RoundTrip(t *testing.T) {
	snapshot := json.RawMessage(`{"tick":5,"core_hp":100}`)
	validActions := []ValidAction{
		{Kind: "dig_room", Params: map[string]any{"x": 10.0, "y": 20.0}},
		{Kind: "summon_beast", Params: map[string]any{"element": "wood"}},
	}
	msg := NewStateMessage(5, snapshot, validActions)

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal StateMessage: %v", err)
	}

	var decoded StateMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal StateMessage: %v", err)
	}

	if decoded.Type != "state" {
		t.Errorf("Type = %q, want %q", decoded.Type, "state")
	}
	if decoded.Tick != 5 {
		t.Errorf("Tick = %d, want 5", decoded.Tick)
	}
	if string(decoded.Snapshot) != string(snapshot) {
		t.Errorf("Snapshot = %s, want %s", decoded.Snapshot, snapshot)
	}
	if len(decoded.ValidActions) != 2 {
		t.Fatalf("ValidActions len = %d, want 2", len(decoded.ValidActions))
	}
	if decoded.ValidActions[0].Kind != "dig_room" {
		t.Errorf("ValidActions[0].Kind = %q, want %q", decoded.ValidActions[0].Kind, "dig_room")
	}
	if decoded.ValidActions[1].Kind != "summon_beast" {
		t.Errorf("ValidActions[1].Kind = %q, want %q", decoded.ValidActions[1].Kind, "summon_beast")
	}
}

func TestGameEndMessage_RoundTrip(t *testing.T) {
	summary := json.RawMessage(`{"tick_count":100}`)
	metrics := json.RawMessage(`{"peak_chi":500.0}`)
	msg := NewGameEndMessage("victory", summary, metrics)

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal GameEndMessage: %v", err)
	}

	var decoded GameEndMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal GameEndMessage: %v", err)
	}

	if decoded.Type != "game_end" {
		t.Errorf("Type = %q, want %q", decoded.Type, "game_end")
	}
	if decoded.Result != "victory" {
		t.Errorf("Result = %q, want %q", decoded.Result, "victory")
	}
	if string(decoded.Summary) != string(summary) {
		t.Errorf("Summary = %s, want %s", decoded.Summary, summary)
	}
	if string(decoded.Metrics) != string(metrics) {
		t.Errorf("Metrics = %s, want %s", decoded.Metrics, metrics)
	}
}

func TestGameEndMessage_RoundTrip_NilMetrics(t *testing.T) {
	summary := json.RawMessage(`{"tick_count":50}`)
	msg := NewGameEndMessage("defeat", summary, nil)

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal GameEndMessage: %v", err)
	}

	// metrics should be omitted from JSON when nil
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if _, ok := raw["metrics"]; ok {
		t.Error("metrics should be omitted when nil")
	}

	var decoded GameEndMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal GameEndMessage: %v", err)
	}
	if decoded.Type != "game_end" {
		t.Errorf("Type = %q, want %q", decoded.Type, "game_end")
	}
	if decoded.Result != "defeat" {
		t.Errorf("Result = %q, want %q", decoded.Result, "defeat")
	}
}

func TestErrorMessage_RoundTrip(t *testing.T) {
	msg := NewErrorMessage("invalid action format")

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal ErrorMessage: %v", err)
	}

	var decoded ErrorMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal ErrorMessage: %v", err)
	}

	if decoded.Type != "error" {
		t.Errorf("Type = %q, want %q", decoded.Type, "error")
	}
	if decoded.Message != "invalid action format" {
		t.Errorf("Message = %q, want %q", decoded.Message, "invalid action format")
	}
}

func TestActionMessage_RoundTrip(t *testing.T) {
	msg := ActionMessage{
		Type: "action",
		Actions: []ActionDef{
			{
				Kind: "dig_room",
				Params: map[string]any{
					"room_type_id": "fire_room",
					"x":           5.0,
					"y":           10.0,
					"width":       3.0,
					"height":      3.0,
				},
			},
			{
				Kind: "dig_corridor",
				Params: map[string]any{
					"from_room_id": 1.0,
					"to_room_id":   2.0,
				},
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal ActionMessage: %v", err)
	}

	var decoded ActionMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal ActionMessage: %v", err)
	}

	if decoded.Type != "action" {
		t.Errorf("Type = %q, want %q", decoded.Type, "action")
	}
	if len(decoded.Actions) != 2 {
		t.Fatalf("Actions len = %d, want 2", len(decoded.Actions))
	}
	if decoded.Actions[0].Kind != "dig_room" {
		t.Errorf("Actions[0].Kind = %q, want %q", decoded.Actions[0].Kind, "dig_room")
	}
	if decoded.Actions[0].Params["room_type_id"] != "fire_room" {
		t.Errorf("Actions[0].Params[room_type_id] = %v, want %q", decoded.Actions[0].Params["room_type_id"], "fire_room")
	}
	if decoded.Actions[1].Kind != "dig_corridor" {
		t.Errorf("Actions[1].Kind = %q, want %q", decoded.Actions[1].Kind, "dig_corridor")
	}
}

func TestValidAction_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		va   ValidAction
	}{
		{
			name: "dig_room",
			va: ValidAction{
				Kind: "dig_room",
				Params: map[string]any{
					"x": 5.0, "y": 10.0,
					"room_type_id": "water_room",
				},
			},
		},
		{
			name: "dig_corridor",
			va: ValidAction{
				Kind: "dig_corridor",
				Params: map[string]any{
					"from_room_id": 1.0,
					"to_room_id":   3.0,
				},
			},
		},
		{
			name: "summon_beast",
			va: ValidAction{
				Kind: "summon_beast",
				Params: map[string]any{
					"element": "fire",
					"cost":    50.0,
				},
			},
		},
		{
			name: "upgrade_room",
			va: ValidAction{
				Kind: "upgrade_room",
				Params: map[string]any{
					"room_id": 2.0,
					"cost":    30.0,
				},
			},
		},
		{
			name: "no_action",
			va: ValidAction{
				Kind:   "no_action",
				Params: map[string]any{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.va)
			if err != nil {
				t.Fatalf("marshal ValidAction: %v", err)
			}

			var decoded ValidAction
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal ValidAction: %v", err)
			}

			if decoded.Kind != tt.va.Kind {
				t.Errorf("Kind = %q, want %q", decoded.Kind, tt.va.Kind)
			}
			if len(decoded.Params) != len(tt.va.Params) {
				t.Errorf("Params len = %d, want %d", len(decoded.Params), len(tt.va.Params))
			}
		})
	}
}

func TestJSONLines_MultipleMessages(t *testing.T) {
	// Simulate writing multiple JSON Lines messages and reading them back.
	messages := []any{
		NewStateMessage(1, json.RawMessage(`{"core_hp":100}`), []ValidAction{
			{Kind: "dig_room", Params: map[string]any{"x": 5.0}},
		}),
		NewStateMessage(2, json.RawMessage(`{"core_hp":95}`), []ValidAction{
			{Kind: "no_action", Params: map[string]any{}},
		}),
		NewGameEndMessage("victory", json.RawMessage(`{"tick_count":2}`), nil),
	}

	var lines [][]byte
	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		lines = append(lines, data)
	}

	// Verify each line is valid JSON and has expected type.
	expectedTypes := []string{"state", "state", "game_end"}
	for i, line := range lines {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(line, &raw); err != nil {
			t.Fatalf("line %d: unmarshal: %v", i, err)
		}
		var typ string
		if err := json.Unmarshal(raw["type"], &typ); err != nil {
			t.Fatalf("line %d: unmarshal type: %v", i, err)
		}
		if typ != expectedTypes[i] {
			t.Errorf("line %d: type = %q, want %q", i, typ, expectedTypes[i])
		}
	}
}
