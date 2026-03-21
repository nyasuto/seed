package simulation

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/testutil"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// newExecutorTestState creates a minimal GameState suitable for executor tests.
// Includes a 20x20 cave with a dragon_hole room, an invader class registry,
// chi pool with 500/1000, and scenario with default constraints.
func newExecutorTestState(t *testing.T) *GameState {
	t.Helper()

	cave, err := world.NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	entrances := []world.RoomEntrance{
		{Pos: types.Pos{X: 4, Y: 3}, Dir: types.East},
	}
	_, err = cave.AddRoom("dragon_hole", types.Pos{X: 2, Y: 2}, 3, 3, entrances)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	chiPool := economy.NewChiPool(1000)
	_ = chiPool.Deposit(500, economy.Supply, "test init", 0)

	engine := economy.NewEconomyEngine(
		chiPool,
		economy.DefaultSupplyParams(),
		economy.DefaultCostParams(),
		economy.DefaultDeficitParams(),
		economy.DefaultConstructionCost(),
		economy.DefaultBeastCost(),
	)

	invReg := invasion.NewInvaderClassRegistry()
	_ = invReg.Register(invasion.InvaderClass{
		ID: "warrior", Name: "Warrior", Element: types.Wood,
		BaseHP: 100, BaseATK: 25, BaseDEF: 20, BaseSPD: 20,
		RewardChi: 15, PreferredGoal: invasion.DestroyCore, RetreatThreshold: 0.3,
	})

	rng := &testutil.FixedRNG{IntValue: 0}

	return &GameState{
		Cave:                 cave,
		EconomyEngine:        engine,
		InvaderClassRegistry: invReg,
		Waves:                make([]*invasion.InvasionWave, 0),
		NextWaveID:           1,
		Progress:             &scenario.ScenarioProgress{CurrentTick: 10},
		Scenario:             &scenario.Scenario{},
		RNG:                  rng,
	}
}

func TestCommandExecutor_SpawnWave(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	cmd := &scenario.SpawnWaveCommand{
		Difficulty:  1.5,
		MinInvaders: 2,
		MaxInvaders: 3,
	}

	err := exec.Apply(state, []scenario.EventCommand{cmd})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(state.Waves) != 1 {
		t.Fatalf("expected 1 wave, got %d", len(state.Waves))
	}

	wave := state.Waves[0]
	if wave.ID != 1 {
		t.Errorf("wave ID = %d, want 1", wave.ID)
	}
	if wave.Difficulty != 1.5 {
		t.Errorf("wave Difficulty = %f, want 1.5", wave.Difficulty)
	}
	if len(wave.Invaders) < 2 || len(wave.Invaders) > 3 {
		t.Errorf("invader count = %d, want 2-3", len(wave.Invaders))
	}
	if state.NextWaveID != 2 {
		t.Errorf("NextWaveID = %d, want 2", state.NextWaveID)
	}
}

func TestCommandExecutor_SpawnWave_MultipleWaves(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	cmd1 := &scenario.SpawnWaveCommand{Difficulty: 1.0, MinInvaders: 1, MaxInvaders: 1}
	cmd2 := &scenario.SpawnWaveCommand{Difficulty: 2.0, MinInvaders: 1, MaxInvaders: 1}

	err := exec.Apply(state, []scenario.EventCommand{cmd1, cmd2})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(state.Waves) != 2 {
		t.Fatalf("expected 2 waves, got %d", len(state.Waves))
	}
	if state.Waves[0].ID != 1 || state.Waves[1].ID != 2 {
		t.Errorf("wave IDs = %d, %d; want 1, 2", state.Waves[0].ID, state.Waves[1].ID)
	}
	if state.NextWaveID != 3 {
		t.Errorf("NextWaveID = %d, want 3", state.NextWaveID)
	}
}

func TestCommandExecutor_ModifyChi_Deposit(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	cmd := &scenario.ModifyChiCommand{Amount: 100}

	err := exec.Apply(state, []scenario.EventCommand{cmd})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	got := state.EconomyEngine.ChiPool.Balance()
	want := 600.0
	if got != want {
		t.Errorf("chi balance = %f, want %f", got, want)
	}
}

func TestCommandExecutor_ModifyChi_Withdraw(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	cmd := &scenario.ModifyChiCommand{Amount: -200}

	err := exec.Apply(state, []scenario.EventCommand{cmd})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	got := state.EconomyEngine.ChiPool.Balance()
	want := 300.0
	if got != want {
		t.Errorf("chi balance = %f, want %f", got, want)
	}
}

func TestCommandExecutor_ModifyChi_WithdrawExceedsBalance(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	// Withdraw more than available (500). Should succeed with partial withdrawal.
	cmd := &scenario.ModifyChiCommand{Amount: -999}

	err := exec.Apply(state, []scenario.EventCommand{cmd})
	if err != nil {
		t.Fatalf("Apply should not return error for partial withdrawal: %v", err)
	}

	got := state.EconomyEngine.ChiPool.Balance()
	if got != 0 {
		t.Errorf("chi balance = %f, want 0", got)
	}
}

func TestCommandExecutor_ModifyConstraint(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		value      float64
		check      func(*testing.T, *scenario.Scenario)
	}{
		{
			name:       "max_rooms",
			constraint: "max_rooms",
			value:      10,
			check: func(t *testing.T, s *scenario.Scenario) {
				if s.Constraints.MaxRooms != 10 {
					t.Errorf("MaxRooms = %d, want 10", s.Constraints.MaxRooms)
				}
			},
		},
		{
			name:       "max_beasts",
			constraint: "max_beasts",
			value:      5,
			check: func(t *testing.T, s *scenario.Scenario) {
				if s.Constraints.MaxBeasts != 5 {
					t.Errorf("MaxBeasts = %d, want 5", s.Constraints.MaxBeasts)
				}
			},
		},
		{
			name:       "max_ticks",
			constraint: "max_ticks",
			value:      200,
			check: func(t *testing.T, s *scenario.Scenario) {
				if s.Constraints.MaxTicks != 200 {
					t.Errorf("MaxTicks = %d, want 200", s.Constraints.MaxTicks)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newExecutorTestState(t)
			exec := NewCommandExecutor()

			cmd := &scenario.ModifyConstraintCommand{
				Constraint: tt.constraint,
				Value:      tt.value,
			}

			err := exec.Apply(state, []scenario.EventCommand{cmd})
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}
			tt.check(t, state.Scenario)
		})
	}
}

func TestCommandExecutor_ModifyConstraint_Unknown(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	cmd := &scenario.ModifyConstraintCommand{
		Constraint: "nonexistent",
		Value:      42,
	}

	err := exec.Apply(state, []scenario.EventCommand{cmd})
	if err == nil {
		t.Fatal("expected error for unknown constraint")
	}
}

func TestCommandExecutor_Message(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	cmd := &scenario.MessageCommand{Text: "Wave incoming!"}

	err := exec.Apply(state, []scenario.EventCommand{cmd})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(exec.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(exec.Messages))
	}
	if exec.Messages[0] != "Wave incoming!" {
		t.Errorf("message = %q, want %q", exec.Messages[0], "Wave incoming!")
	}
}

func TestCommandExecutor_Message_ResetsBetweenCalls(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	_ = exec.Apply(state, []scenario.EventCommand{
		&scenario.MessageCommand{Text: "first"},
	})
	_ = exec.Apply(state, []scenario.EventCommand{
		&scenario.MessageCommand{Text: "second"},
	})

	if len(exec.Messages) != 1 || exec.Messages[0] != "second" {
		t.Errorf("Messages = %v, want [second]", exec.Messages)
	}
}

func TestCommandExecutor_UnknownCommand(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	err := exec.Apply(state, []scenario.EventCommand{unknownCmd{}})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !errors.Is(err, ErrUnknownCommand) {
		t.Errorf("error = %v, want ErrUnknownCommand", err)
	}
}

// unknownCmd is a test-only EventCommand implementation.
type unknownCmd struct{}

func (unknownCmd) Execute() string { return "unknown" }

func TestCommandExecutor_MixedCommands(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	cmds := []scenario.EventCommand{
		&scenario.MessageCommand{Text: "Prepare!"},
		&scenario.ModifyChiCommand{Amount: 50},
		&scenario.ModifyConstraintCommand{Constraint: "max_rooms", Value: 8},
		&scenario.SpawnWaveCommand{Difficulty: 1.0, MinInvaders: 1, MaxInvaders: 1},
	}

	err := exec.Apply(state, cmds)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(exec.Messages) != 1 || exec.Messages[0] != "Prepare!" {
		t.Errorf("Messages = %v, want [Prepare!]", exec.Messages)
	}
	if state.EconomyEngine.ChiPool.Balance() != 550 {
		t.Errorf("chi balance = %f, want 550", state.EconomyEngine.ChiPool.Balance())
	}
	if state.Scenario.Constraints.MaxRooms != 8 {
		t.Errorf("MaxRooms = %d, want 8", state.Scenario.Constraints.MaxRooms)
	}
	if len(state.Waves) != 1 {
		t.Errorf("waves count = %d, want 1", len(state.Waves))
	}
}

func TestCommandExecutor_EmptyCommands(t *testing.T) {
	state := newExecutorTestState(t)
	exec := NewCommandExecutor()

	err := exec.Apply(state, nil)
	if err != nil {
		t.Fatalf("Apply with nil: %v", err)
	}

	err = exec.Apply(state, []scenario.EventCommand{})
	if err != nil {
		t.Fatalf("Apply with empty slice: %v", err)
	}
}
