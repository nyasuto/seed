package simulation

import (
	"errors"
	"fmt"

	"github.com/nyasuto/seed/core/economy"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/types"
)

// ErrUnknownCommand is returned when the command type is not recognized
// by the CommandExecutor.
var ErrUnknownCommand = errors.New("unknown command type")

// CommandExecutor applies EventCommand sequences to a GameState.
// Each command type maps to a concrete state mutation:
//   - SpawnWaveCommand  → generates and appends a new invasion wave
//   - ModifyChiCommand  → deposits or withdraws chi from the pool
//   - ModifyConstraintCommand → updates a scenario constraint value
//   - MessageCommand    → appended to the message log (no state change)
type CommandExecutor struct {
	// Messages collects notification texts produced by MessageCommands
	// during the most recent Apply call.
	Messages []string
}

// NewCommandExecutor creates a CommandExecutor ready for use.
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// Apply executes every command in cmds against the given state, in order.
// Messages is reset at the start of each call.
// If any command fails, execution stops and the error is returned.
func (ce *CommandExecutor) Apply(state *GameState, cmds []scenario.EventCommand) error {
	if ce.Messages != nil {
		ce.Messages = ce.Messages[:0]
	}
	for i, cmd := range cmds {
		if err := ce.applyOne(state, cmd); err != nil {
			return fmt.Errorf("command %d: %w", i, err)
		}
	}
	return nil
}

func (ce *CommandExecutor) applyOne(state *GameState, cmd scenario.EventCommand) error {
	switch c := cmd.(type) {
	case *scenario.SpawnWaveCommand:
		return ce.applySpawnWave(state, c)
	case *scenario.ModifyChiCommand:
		return ce.applyModifyChi(state, c)
	case *scenario.ModifyConstraintCommand:
		return ce.applyModifyConstraint(state, c)
	case *scenario.MessageCommand:
		ce.Messages = append(ce.Messages, c.Text)
		return nil
	default:
		return fmt.Errorf("%w: %T", ErrUnknownCommand, cmd)
	}
}

func (ce *CommandExecutor) applySpawnWave(state *GameState, cmd *scenario.SpawnWaveCommand) error {
	config := invasion.WaveConfig{
		TriggerTick: types.Tick(state.Progress.CurrentTick),
		Difficulty:  cmd.Difficulty,
		MinInvaders: cmd.MinInvaders,
		MaxInvaders: cmd.MaxInvaders,
	}

	wg := invasion.NewWaveGenerator(state.InvaderClassRegistry, state.RNG)
	wg.SetNextWaveID(state.NextWaveID)

	wave, err := wg.GenerateWave(config, state.Cave, types.Tick(state.Progress.CurrentTick))
	if err != nil {
		return fmt.Errorf("spawn wave: %w", err)
	}

	state.Waves = append(state.Waves, wave)
	state.NextWaveID = wg.NextWaveID()
	return nil
}

func (ce *CommandExecutor) applyModifyChi(state *GameState, cmd *scenario.ModifyChiCommand) error {
	pool := state.EconomyEngine.ChiPool
	tick := types.Tick(state.Progress.CurrentTick)

	if cmd.Amount >= 0 {
		return pool.Deposit(cmd.Amount, economy.Reward, "event: chi bonus", tick)
	}
	// Withdraw the absolute value. Partial withdrawal is acceptable,
	// so we ignore ErrInsufficientChi.
	err := pool.Withdraw(-cmd.Amount, economy.Deficit, "event: chi penalty", tick)
	if err != nil && !errors.Is(err, economy.ErrInsufficientChi) {
		return err
	}
	return nil
}

func (ce *CommandExecutor) applyModifyConstraint(state *GameState, cmd *scenario.ModifyConstraintCommand) error {
	c := &state.Scenario.Constraints
	switch cmd.Constraint {
	case "max_rooms":
		c.MaxRooms = int(cmd.Value)
	case "max_beasts":
		c.MaxBeasts = int(cmd.Value)
	case "max_ticks":
		c.MaxTicks = types.Tick(cmd.Value)
	default:
		return fmt.Errorf("unknown constraint: %s", cmd.Constraint)
	}
	return nil
}
