package scenario

import (
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

// SpawnWaveCommand requests an additional invasion wave.
type SpawnWaveCommand struct {
	// Difficulty is the relative difficulty multiplier for the wave.
	Difficulty float64
	// MinInvaders is the minimum number of invaders in the wave.
	MinInvaders int
	// MaxInvaders is the maximum number of invaders in the wave.
	MaxInvaders int
}

func newSpawnWaveCommand(params map[string]any) (*SpawnWaveCommand, error) {
	difficulty, err := paramFloat64(params, "difficulty")
	if err != nil {
		return nil, fmt.Errorf("spawn_wave: %w", err)
	}
	minInv, err := paramFloat64(params, "min_invaders")
	if err != nil {
		return nil, fmt.Errorf("spawn_wave: %w", err)
	}
	maxInv, err := paramFloat64(params, "max_invaders")
	if err != nil {
		return nil, fmt.Errorf("spawn_wave: %w", err)
	}
	if int(minInv) > int(maxInv) {
		return nil, fmt.Errorf("spawn_wave: min_invaders (%d) exceeds max_invaders (%d)", int(minInv), int(maxInv))
	}
	return &SpawnWaveCommand{
		Difficulty:  difficulty,
		MinInvaders: int(minInv),
		MaxInvaders: int(maxInv),
	}, nil
}

// Execute returns a description of the wave to spawn.
func (c *SpawnWaveCommand) Execute() string {
	return fmt.Sprintf("spawn wave: difficulty=%.1f invaders=%d-%d", c.Difficulty, c.MinInvaders, c.MaxInvaders)
}

// ModifyChiCommand requests a change to the chi pool balance.
type ModifyChiCommand struct {
	// Amount is the chi to add (positive) or subtract (negative).
	Amount float64
}

func newModifyChiCommand(params map[string]any) (*ModifyChiCommand, error) {
	amount, err := paramFloat64(params, "amount")
	if err != nil {
		return nil, fmt.Errorf("modify_chi: %w", err)
	}
	return &ModifyChiCommand{Amount: amount}, nil
}

// Execute returns a description of the chi modification.
func (c *ModifyChiCommand) Execute() string {
	return fmt.Sprintf("modify chi: %+.1f", c.Amount)
}

// ModifyConstraintCommand requests a change to a scenario constraint.
type ModifyConstraintCommand struct {
	// Constraint is the name of the constraint to modify.
	Constraint string
	// Value is the new value for the constraint.
	Value float64
}

func newModifyConstraintCommand(params map[string]any) (*ModifyConstraintCommand, error) {
	constraint, err := paramString(params, "constraint")
	if err != nil {
		return nil, fmt.Errorf("modify_constraint: %w", err)
	}
	value, err := paramFloat64(params, "value")
	if err != nil {
		return nil, fmt.Errorf("modify_constraint: %w", err)
	}
	return &ModifyConstraintCommand{Constraint: constraint, Value: value}, nil
}

// Execute returns a description of the constraint modification.
func (c *ModifyConstraintCommand) Execute() string {
	return fmt.Sprintf("modify constraint: %s=%.1f", c.Constraint, c.Value)
}

// MessageCommand delivers a notification message to the player.
type MessageCommand struct {
	// Text is the message to display.
	Text string
}

func newMessageCommand(params map[string]any) (*MessageCommand, error) {
	text, err := paramString(params, "text")
	if err != nil {
		return nil, fmt.Errorf("message: %w", err)
	}
	return &MessageCommand{Text: text}, nil
}

// Execute returns the message text.
func (c *MessageCommand) Execute() string {
	return c.Text
}

// paramString extracts a string parameter by key from a params map.
func paramString(params map[string]any, key string) (string, error) {
	v, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter %q", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("parameter %q must be a string, got %T", key, v)
	}
	return s, nil
}
