package ai

import "encoding/json"

// Server → Client messages

// StateMessage is sent each tick with the current game state and available actions.
type StateMessage struct {
	Type         string          `json:"type"`
	Tick         int             `json:"tick"`
	Snapshot     json.RawMessage `json:"snapshot"`
	ValidActions []ValidAction   `json:"valid_actions"`
}

// GameEndMessage is sent when the game reaches a terminal state.
type GameEndMessage struct {
	Type    string          `json:"type"`
	Result  string          `json:"result"`
	Summary json.RawMessage `json:"summary"`
	Metrics json.RawMessage `json:"metrics,omitempty"`
}

// ErrorMessage is sent when the server encounters an error processing client input.
type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Client → Server messages

// ActionMessage is sent by the client to specify actions for the current tick.
type ActionMessage struct {
	Type    string      `json:"type"`
	Actions []ActionDef `json:"actions"`
}

// ActionDef defines a single action within an ActionMessage.
type ActionDef struct {
	Kind   string         `json:"kind"`
	Params map[string]any `json:"params"`
}

// ValidAction describes an action the player can take this tick.
type ValidAction struct {
	Kind   string         `json:"kind"`
	Params map[string]any `json:"params"`
}

// NewStateMessage creates a StateMessage with type pre-filled.
func NewStateMessage(tick int, snapshot json.RawMessage, validActions []ValidAction) StateMessage {
	return StateMessage{
		Type:         "state",
		Tick:         tick,
		Snapshot:     snapshot,
		ValidActions: validActions,
	}
}

// NewGameEndMessage creates a GameEndMessage with type pre-filled.
func NewGameEndMessage(result string, summary json.RawMessage, metrics json.RawMessage) GameEndMessage {
	return GameEndMessage{
		Type:    "game_end",
		Result:  result,
		Summary: summary,
		Metrics: metrics,
	}
}

// NewErrorMessage creates an ErrorMessage with type pre-filled.
func NewErrorMessage(message string) ErrorMessage {
	return ErrorMessage{
		Type:    "error",
		Message: message,
	}
}
