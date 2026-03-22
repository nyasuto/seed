package simulation

import (
	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/types"
)

// BuildSnapshot constructs a read-only GameSnapshot from the current GameState.
// The snapshot captures a point-in-time view suitable for condition evaluation
// and external action providers.
func BuildSnapshot(state *GameState) scenario.GameSnapshot {
	aliveBeasts := 0
	for _, b := range state.Beasts {
		if b.State != senju.Stunned && b.HP > 0 {
			aliveBeasts++
		}
	}

	defeatedWaves := 0
	for _, w := range state.Waves {
		if w.State == invasion.Completed {
			defeatedWaves++
		}
	}

	var caveFengShuiScore float64
	if state.Cave != nil && state.RoomTypeRegistry != nil && state.ChiFlowEngine != nil {
		params := state.ScoreParams
		if params == nil {
			params = fengshui.DefaultScoreParams()
		}
		ev := fengshui.NewEvaluator(state.Cave, state.RoomTypeRegistry, params)
		caveFengShuiScore = ev.CaveTotal(state.ChiFlowEngine)
	}

	var chiPoolBalance float64
	if state.EconomyEngine != nil && state.EconomyEngine.ChiPool != nil {
		chiPoolBalance = state.EconomyEngine.ChiPool.Balance()
	}

	var tick types.Tick
	var coreHP int
	if state.Progress != nil {
		tick = state.Progress.CurrentTick
		coreHP = state.Progress.CoreHP
	}

	var roomCount int
	if state.Cave != nil {
		roomCount = len(state.Cave.Rooms)
	}

	return scenario.GameSnapshot{
		Tick:                    tick,
		CoreHP:                 coreHP,
		ChiPoolBalance:         chiPoolBalance,
		BeastCount:             len(state.Beasts),
		AliveBeasts:            aliveBeasts,
		DefeatedWaves:          defeatedWaves,
		TotalWaves:             maxInt(len(state.Waves), state.ScheduledWaves),
		CaveFengShuiScore:      caveFengShuiScore,
		ConsecutiveDeficitTicks: state.ConsecutiveDeficitTicks,
		SpawnedWaves:           len(state.Waves),
		RoomCount:              roomCount,
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
