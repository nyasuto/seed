package scenario

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrUnknownCommandType indicates that NewCommand was called with a
// CommandDef whose Type is not recognised by the factory.
var ErrUnknownCommandType = errors.New("unknown command type")

// EventCommand represents a command produced by an event.
// Execute returns a human-readable description of what the command does.
// The actual game-state mutation is performed by the simulation layer.
type EventCommand interface {
	Execute() string
}

// NewCommand creates an EventCommand from a data-driven CommandDef.
// Returns ErrUnknownCommandType when the def.Type is not supported.
func NewCommand(def CommandDef) (EventCommand, error) {
	switch def.Type {
	case "spawn_wave":
		return newSpawnWaveCommand(def.Params)
	case "modify_chi":
		return newModifyChiCommand(def.Params)
	case "modify_constraint":
		return newModifyConstraintCommand(def.Params)
	case "message":
		return newMessageCommand(def.Params)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownCommandType, def.Type)
	}
}

// spawnWaveParams holds the typed parameters for SpawnWaveCommand.
type spawnWaveParams struct {
	Difficulty  float64 `json:"difficulty"`
	MinInvaders int     `json:"min_invaders"`
	MaxInvaders int     `json:"max_invaders"`
}

// SpawnWaveCommand requests an additional invasion wave.
type SpawnWaveCommand struct {
	// Difficulty is the relative difficulty multiplier for the wave.
	Difficulty float64
	// MinInvaders is the minimum number of invaders in the wave.
	MinInvaders int
	// MaxInvaders is the maximum number of invaders in the wave.
	MaxInvaders int
}

func newSpawnWaveCommand(params json.RawMessage) (*SpawnWaveCommand, error) {
	var p spawnWaveParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("spawn_wave params: %w", err)
	}
	if p.Difficulty == 0 {
		return nil, fmt.Errorf("spawn_wave: missing required parameter \"difficulty\"")
	}
	if p.MinInvaders > p.MaxInvaders {
		return nil, fmt.Errorf("spawn_wave: min_invaders (%d) exceeds max_invaders (%d)", p.MinInvaders, p.MaxInvaders)
	}
	return &SpawnWaveCommand{
		Difficulty:  p.Difficulty,
		MinInvaders: p.MinInvaders,
		MaxInvaders: p.MaxInvaders,
	}, nil
}

// Execute returns a description of the wave to spawn.
func (c *SpawnWaveCommand) Execute() string {
	return fmt.Sprintf("spawn wave: difficulty=%.1f invaders=%d-%d", c.Difficulty, c.MinInvaders, c.MaxInvaders)
}

// modifyChiParams holds the typed parameters for ModifyChiCommand.
type modifyChiParams struct {
	Amount float64 `json:"amount"`
}

// ModifyChiCommand requests a change to the chi pool balance.
type ModifyChiCommand struct {
	// Amount is the chi to add (positive) or subtract (negative).
	Amount float64
}

func newModifyChiCommand(params json.RawMessage) (*ModifyChiCommand, error) {
	var p modifyChiParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("modify_chi params: %w", err)
	}
	if p.Amount == 0 {
		return nil, fmt.Errorf("modify_chi: missing required parameter \"amount\"")
	}
	return &ModifyChiCommand{Amount: p.Amount}, nil
}

// Execute returns a description of the chi modification.
func (c *ModifyChiCommand) Execute() string {
	return fmt.Sprintf("modify chi: %+.1f", c.Amount)
}

// modifyConstraintParams holds the typed parameters for ModifyConstraintCommand.
type modifyConstraintParams struct {
	Constraint string   `json:"constraint"`
	Value      *float64 `json:"value"`
}

// ModifyConstraintCommand requests a change to a scenario constraint.
type ModifyConstraintCommand struct {
	// Constraint is the name of the constraint to modify.
	Constraint string
	// Value is the new value for the constraint.
	Value float64
}

func newModifyConstraintCommand(params json.RawMessage) (*ModifyConstraintCommand, error) {
	var p modifyConstraintParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("modify_constraint params: %w", err)
	}
	if p.Constraint == "" {
		return nil, fmt.Errorf("modify_constraint: missing required parameter \"constraint\"")
	}
	if p.Value == nil {
		return nil, fmt.Errorf("modify_constraint: missing required parameter \"value\"")
	}
	return &ModifyConstraintCommand{Constraint: p.Constraint, Value: *p.Value}, nil
}

// Execute returns a description of the constraint modification.
func (c *ModifyConstraintCommand) Execute() string {
	return fmt.Sprintf("modify constraint: %s=%.1f", c.Constraint, c.Value)
}

// messageParams holds the typed parameters for MessageCommand.
type messageParams struct {
	Text string `json:"text"`
}

// MessageCommand delivers a notification message to the player.
type MessageCommand struct {
	// Text is the message to display.
	Text string
}

func newMessageCommand(params json.RawMessage) (*MessageCommand, error) {
	var p messageParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("message params: %w", err)
	}
	if p.Text == "" {
		return nil, fmt.Errorf("message: missing required parameter \"text\"")
	}
	return &MessageCommand{Text: p.Text}, nil
}

// Execute returns the message text.
func (c *MessageCommand) Execute() string {
	return c.Text
}
