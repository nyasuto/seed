package simulation

import (
	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
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

	return scenario.GameSnapshot{
		Tick:                    tick,
		CoreHP:                 coreHP,
		ChiPoolBalance:         chiPoolBalance,
		BeastCount:             len(state.Beasts),
		AliveBeasts:            aliveBeasts,
		DefeatedWaves:          defeatedWaves,
		TotalWaves:             len(state.Waves),
		CaveFengShuiScore:      caveFengShuiScore,
		ConsecutiveDeficitTicks: state.ConsecutiveDeficitTicks,
	}
}
