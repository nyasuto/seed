package human

import (
	"fmt"
	"io"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// FormatTickSummary writes a concise tick summary to w by comparing the
// previous and current snapshots. If prev is nil (first tick), it shows
// the full current state.
func FormatTickSummary(w io.Writer, prev *scenario.GameSnapshot, current scenario.GameSnapshot) {
	fmt.Fprintf(w, "\n--- Tick %d ---\n", current.Tick)
	fmt.Fprintf(w, "  CoreHP: %d", current.CoreHP)
	if prev != nil && current.CoreHP < prev.CoreHP {
		diff := prev.CoreHP - current.CoreHP
		fmt.Fprintf(w, " (-%d!)", diff)
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "  気プール: %.1f", current.ChiPoolBalance)
	if prev != nil {
		delta := current.ChiPoolBalance - prev.ChiPoolBalance
		if delta >= 0.05 {
			fmt.Fprintf(w, " (+%.1f)", delta)
		} else if delta <= -0.05 {
			fmt.Fprintf(w, " (%.1f)", delta)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "  仙獣: %d体 (戦闘可能: %d)\n", current.BeastCount, current.AliveBeasts)
	fmt.Fprintf(w, "  撃退波: %d/%d\n", current.DefeatedWaves, current.TotalWaves)

	// Warnings.
	if current.CoreHP <= 20 && current.CoreHP > 0 {
		fmt.Fprintln(w, "  ⚠ CoreHP が危険水域です！")
	}
	if prev != nil && current.AliveBeasts < prev.AliveBeasts {
		lost := prev.AliveBeasts - current.AliveBeasts
		fmt.Fprintf(w, "  ⚠ 仙獣が %d体 戦闘不能になりました\n", lost)
	}
	if prev != nil && current.DefeatedWaves > prev.DefeatedWaves {
		defeated := current.DefeatedWaves - prev.DefeatedWaves
		fmt.Fprintf(w, "  ✓ %d波 撃退しました！\n", defeated)
	}
}

// FormatFastForwardSummary writes a summary of changes that occurred
// during a fast-forward period, comparing the start and end snapshots.
func FormatFastForwardSummary(w io.Writer, start, end scenario.GameSnapshot) {
	ticks := int(end.Tick - start.Tick)
	fmt.Fprintf(w, "\n=== 早送り完了: %dティック (Tick %d → %d) ===\n",
		ticks, start.Tick, end.Tick)

	// CoreHP change.
	if end.CoreHP != start.CoreHP {
		diff := end.CoreHP - start.CoreHP
		fmt.Fprintf(w, "  CoreHP: %d → %d (%+d)\n", start.CoreHP, end.CoreHP, diff)
	} else {
		fmt.Fprintf(w, "  CoreHP: %d (変化なし)\n", end.CoreHP)
	}

	// Chi change.
	chiDelta := end.ChiPoolBalance - start.ChiPoolBalance
	fmt.Fprintf(w, "  気プール: %.1f → %.1f (%+.1f)\n",
		start.ChiPoolBalance, end.ChiPoolBalance, chiDelta)

	// Beast changes.
	if end.BeastCount != start.BeastCount || end.AliveBeasts != start.AliveBeasts {
		fmt.Fprintf(w, "  仙獣: %d体→%d体 (戦闘可能: %d→%d)\n",
			start.BeastCount, end.BeastCount,
			start.AliveBeasts, end.AliveBeasts)
	}

	// Waves.
	if end.DefeatedWaves > start.DefeatedWaves {
		defeated := end.DefeatedWaves - start.DefeatedWaves
		fmt.Fprintf(w, "  撃退波: +%d (計 %d/%d)\n",
			defeated, end.DefeatedWaves, end.TotalWaves)
	}

	// Warnings.
	if end.CoreHP <= 20 && end.CoreHP > 0 {
		fmt.Fprintln(w, "  ⚠ CoreHP が危険水域です！")
	}
}

// FormatGameEnd writes the game result and statistics.
func FormatGameEnd(w io.Writer, result simulation.RunResult) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "========================================")

	switch result.Result.Status {
	case simulation.Won:
		fmt.Fprintln(w, "  ★ 勝利！ ★")
	case simulation.Lost:
		fmt.Fprintln(w, "  ✗ 敗北...")
	default:
		fmt.Fprintln(w, "  ゲーム終了")
	}

	fmt.Fprintf(w, "  理由: %s\n", result.Result.Reason)
	fmt.Fprintf(w, "  最終ティック: %d\n", result.Result.FinalTick)
	fmt.Fprintln(w, "----------------------------------------")
	fmt.Fprintln(w, "  --- 統計 ---")
	fmt.Fprintf(w, "  最大気プール: %.1f\n", result.Statistics.PeakChi)
	fmt.Fprintf(w, "  撃退波数: %d\n", result.Statistics.WavesDefeated)
	fmt.Fprintf(w, "  最終風水スコア: %.2f\n", result.Statistics.FinalFengShui)
	fmt.Fprintf(w, "  進化回数: %d\n", result.Statistics.Evolutions)
	fmt.Fprintf(w, "  与ダメージ合計: %d\n", result.Statistics.DamageDealt)
	fmt.Fprintf(w, "  被ダメージ合計: %d\n", result.Statistics.DamageReceived)
	fmt.Fprintf(w, "  赤字ティック数: %d\n", result.Statistics.DeficitTicks)
	fmt.Fprintln(w, "========================================")
}
