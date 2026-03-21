package scenario

import (
	"testing"

	"github.com/nyasuto/seed/core/economy"
	"github.com/nyasuto/seed/core/testutil"
	"github.com/nyasuto/seed/core/types"
)

func TestWaveScheduleBuilder_BuildSchedule(t *testing.T) {
	entries := []WaveScheduleEntry{
		{
			TriggerTick:      100,
			Difficulty:       1.0,
			MinInvaders:      2,
			MaxInvaders:      4,
			PreferredClasses: []string{"warrior"},
			PreferredGoals:   []string{"DestroyCore"},
		},
		{
			TriggerTick: 300,
			Difficulty:  2.5,
			MinInvaders: 5,
			MaxInvaders: 8,
		},
	}

	builder := &WaveScheduleBuilder{}
	rng := testutil.NewTestRNG(42)
	configs := builder.BuildSchedule(entries, rng)

	if len(configs) != 2 {
		t.Fatalf("got %d configs, want 2", len(configs))
	}

	tests := []struct {
		name        string
		idx         int
		triggerTick types.Tick
		difficulty  float64
		minInvaders int
		maxInvaders int
	}{
		{"first wave", 0, 100, 1.0, 2, 4},
		{"second wave", 1, 300, 2.5, 5, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := configs[tt.idx]
			if c.TriggerTick != tt.triggerTick {
				t.Errorf("TriggerTick = %d, want %d", c.TriggerTick, tt.triggerTick)
			}
			if c.Difficulty != tt.difficulty {
				t.Errorf("Difficulty = %f, want %f", c.Difficulty, tt.difficulty)
			}
			if c.MinInvaders != tt.minInvaders {
				t.Errorf("MinInvaders = %d, want %d", c.MinInvaders, tt.minInvaders)
			}
			if c.MaxInvaders != tt.maxInvaders {
				t.Errorf("MaxInvaders = %d, want %d", c.MaxInvaders, tt.maxInvaders)
			}
		})
	}
}

func TestWaveScheduleBuilder_BuildSchedule_Empty(t *testing.T) {
	builder := &WaveScheduleBuilder{}
	rng := testutil.NewTestRNG(1)
	configs := builder.BuildSchedule(nil, rng)
	if len(configs) != 0 {
		t.Errorf("got %d configs for nil entries, want 0", len(configs))
	}
}

func TestCalcFirstWaveTiming_BeforeConstructionComplete(t *testing.T) {
	costs := economy.ConstructionCost{
		RoomCost: map[string]float64{
			"dragon_hole": 100.0,
			"trap_room":   50.0,
			"storage":     80.0,
		},
	}

	initialState := InitialState{
		StartingChi: 10.0,
		DragonVeins: []DragonVeinPlacement{
			{
				SourcePos: types.Pos{X: 0, Y: 0},
				Element:   types.Wood,
				FlowRate:  1.0,
			},
		},
	}

	timing := CalcFirstWaveTiming(initialState, costs)

	// Cheapest room = 50.0. Remaining = 50.0 - 10.0 = 40.0.
	// Chi per tick = 1.0. Ticks needed = 40.0.
	// Midpoint = 20. Wave should arrive before full construction is possible.
	ticksToAfford := types.Tick((50.0 - 10.0) / 1.0)
	if timing >= ticksToAfford {
		t.Errorf("timing %d should be before ticks to afford cheapest room %d", timing, ticksToAfford)
	}
	if timing < 10 {
		t.Errorf("timing %d should be at least 10", timing)
	}
}

func TestCalcFirstWaveTiming_AlreadyAffordable(t *testing.T) {
	costs := economy.ConstructionCost{
		RoomCost: map[string]float64{
			"trap_room": 20.0,
		},
	}

	initialState := InitialState{
		StartingChi: 100.0,
	}

	timing := CalcFirstWaveTiming(initialState, costs)
	// Player can already afford the room, so timing should be the short default.
	if timing != 50 {
		t.Errorf("timing = %d, want 50 for already-affordable scenario", timing)
	}
}

func TestCalcFirstWaveTiming_NoRoomCosts(t *testing.T) {
	costs := economy.ConstructionCost{
		RoomCost: map[string]float64{},
	}
	initialState := InitialState{StartingChi: 10.0}

	timing := CalcFirstWaveTiming(initialState, costs)
	if timing != 100 {
		t.Errorf("timing = %d, want 100 for no room costs", timing)
	}
}

func TestCalcFirstWaveTiming_NoDragonVeins(t *testing.T) {
	costs := economy.ConstructionCost{
		RoomCost: map[string]float64{
			"trap_room": 50.0,
		},
	}
	initialState := InitialState{
		StartingChi: 10.0,
		DragonVeins: nil,
	}

	timing := CalcFirstWaveTiming(initialState, costs)
	// No chi income → default 200.
	if timing != 200 {
		t.Errorf("timing = %d, want 200 for no dragon veins", timing)
	}
}

func TestCalcFirstWaveTiming_MultipleDragonVeins(t *testing.T) {
	costs := economy.ConstructionCost{
		RoomCost: map[string]float64{
			"trap_room": 100.0,
		},
	}
	initialState := InitialState{
		StartingChi: 0.0,
		DragonVeins: []DragonVeinPlacement{
			{FlowRate: 2.0},
			{FlowRate: 3.0},
		},
	}

	timing := CalcFirstWaveTiming(initialState, costs)
	// Cheapest = 100, remaining = 100, chiPerTick = 5.0.
	// ticksNeeded = 20, midpoint = 10.
	if timing != 10 {
		t.Errorf("timing = %d, want 10", timing)
	}
}
