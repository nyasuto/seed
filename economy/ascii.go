package economy

import "fmt"

// RenderEconomyStatus returns a one-line summary of the economy state.
// Format: [Chi: <current>/<cap> | +<supply> -<maintenance> = <net>/tick | <status>]
// where <status> is OK, MILD, MODERATE, or SEVERE depending on the deficit severity.
func RenderEconomyStatus(engine *EconomyEngine, lastTick *EconomyTickResult) string {
	current := engine.ChiPool.Balance()
	cap := engine.ChiPool.Cap

	var supply, maintenance, net float64
	var status string

	if lastTick != nil {
		supply = lastTick.Supply
		maintenance = lastTick.Maintenance.Total
		net = supply - maintenance
		status = severityLabel(lastTick.DeficitResult.Severity)
	} else {
		status = "OK"
	}

	return fmt.Sprintf("[Chi: %.1f/%.1f | +%.1f -%.1f = %+.1f/tick | %s]",
		current, cap, supply, maintenance, net, status)
}

// severityLabel returns a display label for the given deficit severity.
func severityLabel(s DeficitSeverity) string {
	switch s {
	case Mild:
		return "MILD"
	case Moderate:
		return "MODERATE"
	case Severe:
		return "SEVERE"
	default:
		return "OK"
	}
}
