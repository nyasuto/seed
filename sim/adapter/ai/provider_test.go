package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// makeStateBuilder creates a StateBuilder backed by the test state.
func makeStateBuilder(t *testing.T) *StateBuilder {
	t.Helper()
	state := newTestState(t)
	engine := &simulation.SimulationEngine{State: state}
	return NewStateBuilder(func() *simulation.SimulationEngine { return engine })
}

// makeSnapshot creates a minimal GameSnapshot for testing.
func makeSnapshot() scenario.GameSnapshot {
	return scenario.GameSnapshot{
		Tick:           1,
		CoreHP:         100,
		ChiPoolBalance: 500,
	}
}

// readJSONLine reads the next JSON line from a scanner and unmarshals it into v.
func readJSONLine(t *testing.T, scanner *bufio.Scanner, v any) {
	t.Helper()
	if !scanner.Scan() {
		t.Fatalf("expected a JSON line, got EOF; err=%v", scanner.Err())
	}
	if err := json.Unmarshal(scanner.Bytes(), v); err != nil {
		t.Fatalf("unmarshal %q: %v", scanner.Text(), err)
	}
}

func TestAIProvider_ProvideActions_WaitAction(t *testing.T) {
	// Client sends a wait action.
	clientInput := `{"type":"action","actions":[{"kind":"wait","params":{}}]}` + "\n"
	in := strings.NewReader(clientInput)
	var out bytes.Buffer
	builder := makeStateBuilder(t)

	provider := NewAIProvider(in, &out, builder)
	actions, err := provider.ProvideActions(makeSnapshot())
	if err != nil {
		t.Fatalf("ProvideActions: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].ActionType() != "no_action" {
		t.Errorf("action type = %q, want \"no_action\"", actions[0].ActionType())
	}

	// Verify a StateMessage was written.
	scanner := bufio.NewScanner(&out)
	var stateMsg StateMessage
	readJSONLine(t, scanner, &stateMsg)
	if stateMsg.Type != "state" {
		t.Errorf("output type = %q, want \"state\"", stateMsg.Type)
	}
	if stateMsg.Tick != 1 {
		t.Errorf("tick = %d, want 1", stateMsg.Tick)
	}
}

func TestAIProvider_ProvideActions_InvalidJSON_ThenValid(t *testing.T) {
	// First line: invalid JSON. Second line: valid action after re-sent state.
	clientInput := "not json\n" +
		`{"type":"action","actions":[{"kind":"wait","params":{}}]}` + "\n"
	in := strings.NewReader(clientInput)
	var out bytes.Buffer
	builder := makeStateBuilder(t)

	provider := NewAIProvider(in, &out, builder)
	actions, err := provider.ProvideActions(makeSnapshot())
	if err != nil {
		t.Fatalf("ProvideActions: %v", err)
	}

	if len(actions) != 1 || actions[0].ActionType() != "no_action" {
		t.Fatalf("expected 1 no_action, got %v", actions)
	}

	// Output should contain: state, error, state (re-sent).
	scanner := bufio.NewScanner(bytes.NewReader(out.Bytes()))
	var msg1 map[string]any
	readJSONLine(t, scanner, &msg1)
	if msg1["type"] != "state" {
		t.Errorf("first output type = %v, want \"state\"", msg1["type"])
	}

	var errMsg ErrorMessage
	readJSONLine(t, scanner, &errMsg)
	if errMsg.Type != "error" {
		t.Errorf("second output type = %q, want \"error\"", errMsg.Type)
	}
	if !strings.Contains(errMsg.Message, "invalid JSON") {
		t.Errorf("error message = %q, want to contain \"invalid JSON\"", errMsg.Message)
	}

	var msg3 map[string]any
	readJSONLine(t, scanner, &msg3)
	if msg3["type"] != "state" {
		t.Errorf("third output type = %v, want \"state\"", msg3["type"])
	}
}

func TestAIProvider_ProvideActions_InvalidAction_NotInValidActions(t *testing.T) {
	// Send an action that is not in valid_actions (upgrade room 999 which doesn't exist).
	clientInput := `{"type":"action","actions":[{"kind":"upgrade_room","params":{"room_id":999}}]}` + "\n" +
		`{"type":"action","actions":[{"kind":"wait","params":{}}]}` + "\n"
	in := strings.NewReader(clientInput)
	var out bytes.Buffer
	builder := makeStateBuilder(t)

	provider := NewAIProvider(in, &out, builder)
	actions, err := provider.ProvideActions(makeSnapshot())
	if err != nil {
		t.Fatalf("ProvideActions: %v", err)
	}

	// Should have retried and got the wait action.
	if len(actions) != 1 || actions[0].ActionType() != "no_action" {
		t.Fatalf("expected 1 no_action, got %v", actions)
	}

	// Check error message was sent.
	scanner := bufio.NewScanner(bytes.NewReader(out.Bytes()))
	var stateMsg map[string]any
	readJSONLine(t, scanner, &stateMsg) // initial state
	var errMsg ErrorMessage
	readJSONLine(t, scanner, &errMsg) // error
	if !strings.Contains(errMsg.Message, "not in valid_actions") {
		t.Errorf("error message = %q, want to contain \"not in valid_actions\"", errMsg.Message)
	}
}

func TestAIProvider_ProvideActions_EOF(t *testing.T) {
	// Empty input = EOF.
	in := strings.NewReader("")
	var out bytes.Buffer
	builder := makeStateBuilder(t)

	provider := NewAIProvider(in, &out, builder)
	_, err := provider.ProvideActions(makeSnapshot())
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestAIProvider_OnGameEnd_Victory(t *testing.T) {
	var out bytes.Buffer
	builder := makeStateBuilder(t)
	provider := NewAIProvider(strings.NewReader(""), &out, builder)

	result := simulation.RunResult{
		Result: simulation.GameResult{
			Status: simulation.Won,
			Reason: "all waves defeated",
		},
		TickCount: 100,
	}
	provider.OnGameEnd(result)

	scanner := bufio.NewScanner(&out)
	var msg GameEndMessage
	readJSONLine(t, scanner, &msg)
	if msg.Type != "game_end" {
		t.Errorf("type = %q, want \"game_end\"", msg.Type)
	}
	if msg.Result != "victory" {
		t.Errorf("result = %q, want \"victory\"", msg.Result)
	}
}

func TestAIProvider_OnGameEnd_Defeat(t *testing.T) {
	var out bytes.Buffer
	builder := makeStateBuilder(t)
	provider := NewAIProvider(strings.NewReader(""), &out, builder)

	result := simulation.RunResult{
		Result: simulation.GameResult{
			Status: simulation.Lost,
			Reason: "core destroyed",
		},
		TickCount: 50,
	}
	provider.OnGameEnd(result)

	scanner := bufio.NewScanner(&out)
	var msg GameEndMessage
	readJSONLine(t, scanner, &msg)
	if msg.Result != "defeat" {
		t.Errorf("result = %q, want \"defeat\"", msg.Result)
	}
}

func TestAIProvider_WrongMessageType(t *testing.T) {
	// Send a message with wrong type field.
	clientInput := `{"type":"state","tick":1}` + "\n" +
		`{"type":"action","actions":[{"kind":"wait","params":{}}]}` + "\n"
	in := strings.NewReader(clientInput)
	var out bytes.Buffer
	builder := makeStateBuilder(t)

	provider := NewAIProvider(in, &out, builder)
	actions, err := provider.ProvideActions(makeSnapshot())
	if err != nil {
		t.Fatalf("ProvideActions: %v", err)
	}

	if len(actions) != 1 || actions[0].ActionType() != "no_action" {
		t.Fatalf("expected retry to succeed with wait, got %v", actions)
	}
}

func TestAIProvider_EmptyActionsArray(t *testing.T) {
	// Send an action message with empty actions array.
	clientInput := `{"type":"action","actions":[]}` + "\n" +
		`{"type":"action","actions":[{"kind":"wait","params":{}}]}` + "\n"
	in := strings.NewReader(clientInput)
	var out bytes.Buffer
	builder := makeStateBuilder(t)

	provider := NewAIProvider(in, &out, builder)
	actions, err := provider.ProvideActions(makeSnapshot())
	if err != nil {
		t.Fatalf("ProvideActions: %v", err)
	}

	if len(actions) != 1 || actions[0].ActionType() != "no_action" {
		t.Fatalf("expected retry to succeed with wait, got %v", actions)
	}
}

func TestConvertAction_AllTypes(t *testing.T) {
	tests := []struct {
		name       string
		def        ActionDef
		wantType   string
		wantErr    bool
	}{
		{
			name:     "wait",
			def:      ActionDef{Kind: "wait", Params: map[string]any{}},
			wantType: "no_action",
		},
		{
			name: "dig_room",
			def: ActionDef{Kind: "dig_room", Params: map[string]any{
				"room_type_id": "trap_room",
				"x":            float64(5),
				"y":            float64(6),
				"width":        float64(3),
				"height":       float64(3),
			}},
			wantType: "dig_room",
		},
		{
			name: "dig_corridor",
			def: ActionDef{Kind: "dig_corridor", Params: map[string]any{
				"from_room_id": float64(1),
				"to_room_id":   float64(2),
			}},
			wantType: "dig_corridor",
		},
		{
			name: "summon_beast",
			def: ActionDef{Kind: "summon_beast", Params: map[string]any{
				"element": "Fire",
			}},
			wantType: "summon_beast",
		},
		{
			name: "upgrade_room",
			def: ActionDef{Kind: "upgrade_room", Params: map[string]any{
				"room_id": float64(1),
			}},
			wantType: "upgrade_room",
		},
		{
			name: "evolve_beast",
			def: ActionDef{Kind: "evolve_beast", Params: map[string]any{
				"beast_id": float64(1),
			}},
			wantType: "evolve_beast",
		},
		{
			name: "place_beast",
			def: ActionDef{Kind: "place_beast", Params: map[string]any{
				"species_id": "fire_lizard",
				"room_id":    float64(1),
			}},
			wantType: "place_beast",
		},
		{
			name:    "unknown",
			def:     ActionDef{Kind: "unknown", Params: map[string]any{}},
			wantErr: true,
		},
		{
			name: "summon_beast_bad_element",
			def: ActionDef{Kind: "summon_beast", Params: map[string]any{
				"element": "Void",
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := convertAction(tt.def)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if action.ActionType() != tt.wantType {
				t.Errorf("action type = %q, want %q", action.ActionType(), tt.wantType)
			}
		})
	}
}

func TestParseElement(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Wood", "Wood"},
		{"wood", "Wood"},
		{"fire", "Fire"},
		{"EARTH", "Earth"},
		{"Metal", "Metal"},
		{"water", "Water"},
	}
	for _, tt := range tests {
		elem, err := parseElement(tt.input)
		if err != nil {
			t.Errorf("parseElement(%q): %v", tt.input, err)
			continue
		}
		if elem.String() != tt.want {
			t.Errorf("parseElement(%q) = %v, want %v", tt.input, elem, tt.want)
		}
	}

	_, err := parseElement("Void")
	if err == nil {
		t.Error("expected error for unknown element \"Void\"")
	}
}

func TestAIProvider_MaxRetriesExceeded_FallsBackToWait(t *testing.T) {
	// Send 4 invalid lines + no valid one = exceed 3 retries.
	clientInput := "bad1\nbad2\nbad3\nbad4\n"
	in := strings.NewReader(clientInput)
	var out bytes.Buffer
	builder := makeStateBuilder(t)

	provider := NewAIProvider(in, &out, builder)
	actions, err := provider.ProvideActions(makeSnapshot())
	if err != nil {
		t.Fatalf("expected fallback to wait, got error: %v", err)
	}

	if len(actions) != 1 || actions[0].ActionType() != "no_action" {
		t.Fatalf("expected 1 no_action (fallback), got %v", actions)
	}
}
