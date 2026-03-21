package scenario

import "fmt"

// RenderScenarioStatus returns a one-line status summary of the running scenario.
// Format: [<difficulty> | Tick <cur>/<max> | Waves <defeated>/<total> | Core HP <hp> | Win: <condition> <cur>/<target>]
func RenderScenarioStatus(sc *Scenario, prog *ScenarioProgress, snap GameSnapshot) string {
	result := "[" + sc.Difficulty

	// Tick progress
	if sc.Constraints.MaxTicks > 0 {
		result += fmt.Sprintf(" | Tick %d/%d", prog.CurrentTick, sc.Constraints.MaxTicks)
	} else {
		result += fmt.Sprintf(" | Tick %d", prog.CurrentTick)
	}

	// Wave progress
	result += fmt.Sprintf(" | Waves %d/%d", snap.DefeatedWaves, snap.TotalWaves)

	// Core HP
	result += fmt.Sprintf(" | Core HP %d", prog.CoreHP)

	// Win condition progress (show conditions with measurable targets)
	for _, cond := range sc.WinConditions {
		label := winConditionLabel(cond, snap)
		if label != "" {
			result += " | Win: " + label
		}
	}

	result += "]"
	return result
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
