package balance

import (
	"fmt"
	"strings"

	"github.com/nyasuto/seed/sim/adapter/batch"
	"github.com/nyasuto/seed/sim/metrics"
)

// ComparisonRow represents a single row in the sweep comparison table.
type ComparisonRow struct {
	// ParamValue is the parameter value used for this sweep run.
	ParamValue string
	// WinRate is the fraction of games won.
	WinRate float64
	// AlertMetricValue is the value of the target alert metric.
	AlertMetricValue float64
	// AlertResolved is true if the original alert metric is now within threshold.
	AlertResolved bool
	// NewAlerts lists any new breakage alerts that appeared in this sweep run.
	NewAlerts []metrics.BreakageAlert
	// IsBaseline marks this row as the baseline result.
	IsBaseline bool
}

// ComparisonResult holds the full comparison between baseline and sweep results.
type ComparisonResult struct {
	// ParamKey is the parameter that was swept.
	ParamKey string
	// AlertMetricID is the metric ID of the original alert being addressed.
	AlertMetricID string
	// Rows contains one row per sweep value.
	Rows []ComparisonRow
	// BestIndex is the index of the best row (-1 if no good candidate found).
	// "Best" = alert resolved + no new alerts.
	BestIndex int
}

// CompareResults builds a comparison between baseline and sweep results.
// baselineReport is the breakage report from the baseline run.
// alertMetricID is the metric being targeted for resolution.
// sweepResults contains the batch results for each swept parameter value.
func CompareResults(baselineReport metrics.BreakageReport, alertMetricID string, sweepResults []batch.SweepResult) ComparisonResult {
	// Collect baseline alert metric IDs for detecting new alerts.
	baselineAlertIDs := make(map[string]bool)
	for _, a := range baselineReport.Alerts {
		baselineAlertIDs[a.MetricID] = true
	}

	result := ComparisonResult{
		AlertMetricID: alertMetricID,
		BestIndex:     -1,
	}
	if len(sweepResults) > 0 {
		result.ParamKey = sweepResults[0].ParamKey
	}

	for _, sr := range sweepResults {
		winRate, _ := calcSummaryStats(sr.Result.Summaries)

		// Get the value of the target metric from the sweep's breakage data.
		values := metricValues(sr.Result.BreakageData)
		alertValue := values[alertMetricID]

		// Check if the target alert is resolved in this sweep run.
		alertResolved := true
		for _, a := range sr.Result.BreakageReport.Alerts {
			if a.MetricID == alertMetricID {
				alertResolved = false
				break
			}
		}

		// Find new alerts that weren't in the baseline.
		var newAlerts []metrics.BreakageAlert
		for _, a := range sr.Result.BreakageReport.Alerts {
			if !baselineAlertIDs[a.MetricID] {
				newAlerts = append(newAlerts, a)
			}
		}

		result.Rows = append(result.Rows, ComparisonRow{
			ParamValue:       sr.ParamValue,
			WinRate:          winRate,
			AlertMetricValue: alertValue,
			AlertResolved:    alertResolved,
			NewAlerts:        newAlerts,
		})
	}

	// Find the best row: alert resolved + no new alerts.
	for i, row := range result.Rows {
		if row.AlertResolved && len(row.NewAlerts) == 0 {
			result.BestIndex = i
			// Keep the first one found (most conservative parameter change).
			break
		}
	}

	return result
}

// FormatComparison formats the comparison result as a human-readable table
// matching the PRD dashboard output format.
func FormatComparison(cr ComparisonResult) string {
	if len(cr.Rows) == 0 {
		return "No sweep results to compare.\n"
	}

	var sb strings.Builder
	sb.WriteString("Sweep Results:\n")

	// Extract the last segment of the param key for display.
	paramShort := cr.ParamKey
	if idx := strings.LastIndex(paramShort, "."); idx >= 0 {
		paramShort = paramShort[idx+1:]
	}

	for i, row := range cr.Rows {
		// Format: paramShort=value → MetricID=value [status] WinRate=X
		status := "🔴"
		if row.AlertResolved {
			status = "✅"
		}

		line := fmt.Sprintf("  %s=%s → %s=%.2f %s  WinRate=%.2f",
			paramShort, row.ParamValue,
			cr.AlertMetricID, row.AlertMetricValue,
			status, row.WinRate)

		if row.IsBaseline {
			line += "  ← baseline"
		}

		if len(row.NewAlerts) > 0 {
			ids := make([]string, len(row.NewAlerts))
			for j, a := range row.NewAlerts {
				ids[j] = a.MetricID
			}
			line += fmt.Sprintf("  ← %s 🔴 新たな壊れ", strings.Join(ids, ","))
		}

		if i == cr.BestIndex {
			line += "  ← Best"
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	if cr.BestIndex >= 0 {
		best := cr.Rows[cr.BestIndex]
		fmt.Fprintf(&sb, "Best: %s=%s (%s resolved, no new alerts)\n",
			paramShort, best.ParamValue, cr.AlertMetricID)
	} else {
		sb.WriteString("No parameter value resolved the alert without introducing new alerts.\n")
	}

	return sb.String()
}
