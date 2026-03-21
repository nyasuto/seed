package scenario

import "strings"

import "fmt"

// RenderScenarioStatus returns a one-line status summary of the running scenario.
// Format: [<difficulty> | Tick <cur>/<max> | Waves <defeated>/<total> | Core HP <hp> | Win: <condition> <cur>/<target>]
func RenderScenarioStatus(sc *Scenario, prog *ScenarioProgress, snap GameSnapshot) string {
	var result strings.Builder
	result.WriteString("[" + sc.Difficulty)

	// Tick progress
	if sc.Constraints.MaxTicks > 0 {
		fmt.Fprintf(&result, " | Tick %d/%d", prog.CurrentTick, sc.Constraints.MaxTicks)
	} else {
		fmt.Fprintf(&result, " | Tick %d", prog.CurrentTick)
	}

	// Wave progress
	fmt.Fprintf(&result, " | Waves %d/%d", snap.DefeatedWaves, snap.TotalWaves)

	// Core HP
	fmt.Fprintf(&result, " | Core HP %d", prog.CoreHP)

	// Win condition progress (show conditions with measurable targets)
	for _, cond := range sc.WinConditions {
		label := winConditionLabel(cond, snap)
		if label != "" {
			result.WriteString(" | Win: " + label)
		}
	}

	result.WriteString("]")
	return result.String()
}

// winConditionLabel returns a progress label for a win condition, or "" if
// the condition type does not have a meaningful progress display.
func winConditionLabel(cond ConditionDef, snap GameSnapshot) string {
	switch cond.Type {
	case "fengshui_score":
		threshold, err := paramFloat64(cond.Params, "threshold")
		if err != nil {
			return ""
		}
		return fmt.Sprintf("FengShui %.0f/%.0f", snap.CaveFengShuiScore, threshold)
	case "chi_pool":
		threshold, err := paramFloat64(cond.Params, "threshold")
		if err != nil {
			return ""
		}
		return fmt.Sprintf("Chi %.0f/%.0f", snap.ChiPoolBalance, threshold)
	default:
		return ""
	}
}
