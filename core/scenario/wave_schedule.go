package scenario

import (
	"github.com/nyasuto/seed/core/economy"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/types"
)

// WaveScheduleEntry defines when and what invaders appear in a scenario.
type WaveScheduleEntry struct {
	// TriggerTick is the tick at which this wave activates.
	TriggerTick types.Tick `json:"trigger_tick"`
	// Difficulty is the difficulty multiplier affecting invader levels.
	Difficulty float64 `json:"difficulty"`
	// MinInvaders is the minimum number of invaders in the wave.
	MinInvaders int `json:"min_invaders"`
	// MaxInvaders is the maximum number of invaders in the wave.
	MaxInvaders int `json:"max_invaders"`
	// PreferredClasses is the list of invader class IDs preferred for this wave.
	// If empty, any registered class may be used.
	PreferredClasses []string `json:"preferred_classes,omitempty"`
	// PreferredGoals is the list of goal type names preferred for this wave.
	// If empty, goal assignment follows the default logic in WaveGenerator.
	PreferredGoals []string `json:"preferred_goals,omitempty"`
}

// WaveScheduleBuilder converts scenario-level wave schedule entries
// into invasion.WaveConfig values that Phase 4's WaveGenerator can consume.
type WaveScheduleBuilder struct{}

// BuildSchedule converts a slice of WaveScheduleEntry into a slice of
// invasion.WaveConfig. The PreferredClasses and PreferredGoals fields on
// each entry are advisory hints for the simulation layer; they do not
// affect the WaveConfig output because WaveConfig delegates class/goal
// selection to the WaveGenerator at runtime.
func (b *WaveScheduleBuilder) BuildSchedule(entries []WaveScheduleEntry, rng types.RNG) []invasion.WaveConfig {
	configs := make([]invasion.WaveConfig, len(entries))
	for i, e := range entries {
		configs[i] = invasion.WaveConfig{
			TriggerTick: e.TriggerTick,
			Difficulty:  e.Difficulty,
			MinInvaders: e.MinInvaders,
			MaxInvaders: e.MaxInvaders,
		}
	}
	return configs
}

// CalcFirstWaveTiming estimates a tick at which the first wave should arrive,
// based on the player's starting chi and the cheapest room construction cost.
// The idea is to give the player enough time to build at least some defenses
// but impose time pressure (D002).
//
// The calculation estimates how many ticks it would take to accumulate enough
// chi to build the cheapest room, then returns a tick partway through that
// window so the wave arrives before the player is fully prepared.
func CalcFirstWaveTiming(initialState InitialState, costs economy.ConstructionCost) types.Tick {
	// Find the cheapest room cost.
	cheapest := 0.0
	for _, cost := range costs.RoomCost {
		if cheapest == 0 || cost < cheapest {
			cheapest = cost
		}
	}

	// If no room costs are defined or cheapest is free, return a sensible default.
	if cheapest <= 0 {
		return 100
	}

	// Estimate ticks needed to afford the cheapest room from starting chi.
	// If the player already has enough chi, they still need some build time.
	remaining := cheapest - initialState.StartingChi
	if remaining <= 0 {
		// Player can already afford a room; wave arrives at a short fixed offset.
		return 50
	}

	// Estimate chi income per tick from dragon veins.
	chiPerTick := 0.0
	for _, dv := range initialState.DragonVeins {
		chiPerTick += dv.FlowRate
	}

	// If no income, use a default timeframe.
	if chiPerTick <= 0 {
		return 200
	}

	// Ticks to accumulate enough chi for the cheapest room.
	ticksNeeded := remaining / chiPerTick

	// Return a tick partway through (midpoint) to impose time pressure.
	midpoint := max(types.Tick(ticksNeeded/2), 10)

	return midpoint
}
