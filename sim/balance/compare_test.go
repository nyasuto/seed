package balance

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/adapter/batch"
	"github.com/nyasuto/seed/sim/metrics"
)

func TestCompareResults_AlertResolved(t *testing.T) {
	baselineReport := metrics.BreakageReport{
		Alerts: []metrics.BreakageAlert{
			{MetricID: "B06", BrokenSign: "stomp victory with high HP", Value: 0.35, Direction: metrics.Above},
		},
		Clean: []string{"B01", "B02", "B03", "B04", "B05", "B07", "B08", "B09", "B10", "B11"},
	}

	sweepResults := []batch.SweepResult{
		{
			ParamKey:   "wave_schedule.0.difficulty",
			ParamValue: "1.5",
			Result: &batch.BatchResult{
				Summaries: []metrics.GameSummary{
					{Result: simulation.Won, TotalTicks: 180},
					{Result: simulation.Won, TotalTicks: 200},
				},
				BreakageReport: metrics.BreakageReport{
					Alerts: []metrics.BreakageAlert{
						{MetricID: "B06", Value: 0.25},
					},
				},
				BreakageData: []metrics.BreakageData{
					{B06Stomp: true},
					{B06Stomp: false},
				},
			},
		},
		{
			ParamKey:   "wave_schedule.0.difficulty",
			ParamValue: "2.0",
			Result: &batch.BatchResult{
				Summaries: []metrics.GameSummary{
					{Result: simulation.Won, TotalTicks: 180},
					{Result: simulation.Lost, TotalTicks: 150},
				},
				BreakageReport: metrics.BreakageReport{
					Alerts: nil,
					Clean:  []string{"B06"},
				},
				BreakageData: []metrics.BreakageData{
					{B06Stomp: false},
					{B06Stomp: false},
				},
			},
		},
	}

	cr := CompareResults(baselineReport, "B06", sweepResults)

	if cr.AlertMetricID != "B06" {
		t.Errorf("AlertMetricID: got %q, want %q", cr.AlertMetricID, "B06")
	}
	if len(cr.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(cr.Rows))
	}

	// First row: B06 still alerted.
	if cr.Rows[0].AlertResolved {
		t.Error("row 0: B06 should NOT be resolved")
	}

	// Second row: B06 resolved.
	if !cr.Rows[1].AlertResolved {
		t.Error("row 1: B06 should be resolved")
	}

	// Best index should be row 1 (resolved, no new alerts).
	if cr.BestIndex != 1 {
		t.Errorf("BestIndex: got %d, want 1", cr.BestIndex)
	}
}

func TestCompareResults_NewAlertWarning(t *testing.T) {
	baselineReport := metrics.BreakageReport{
		Alerts: []metrics.BreakageAlert{
			{MetricID: "B06", BrokenSign: "stomp victory with high HP"},
		},
	}

	sweepResults := []batch.SweepResult{
		{
			ParamKey:   "wave_schedule.0.difficulty",
			ParamValue: "3.0",
			Result: &batch.BatchResult{
				Summaries: []metrics.GameSummary{
					{Result: simulation.Lost, TotalTicks: 80},
				},
				BreakageReport: metrics.BreakageReport{
					Alerts: []metrics.BreakageAlert{
						// B06 resolved but B07 appeared.
						{MetricID: "B07", BrokenSign: "early wipe before midgame", Value: 0.25},
					},
				},
				BreakageData: []metrics.BreakageData{
					{B07EarlyWipe: true},
				},
			},
		},
	}

	cr := CompareResults(baselineReport, "B06", sweepResults)

	if len(cr.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(cr.Rows))
	}

	row := cr.Rows[0]
	if !row.AlertResolved {
		t.Error("B06 should be resolved")
	}
	if len(row.NewAlerts) != 1 {
		t.Fatalf("expected 1 new alert, got %d", len(row.NewAlerts))
	}
	if row.NewAlerts[0].MetricID != "B07" {
		t.Errorf("new alert: got %q, want %q", row.NewAlerts[0].MetricID, "B07")
	}

	// Best index should be -1 (new alert appeared).
	if cr.BestIndex != -1 {
		t.Errorf("BestIndex: got %d, want -1 (new alert appeared)", cr.BestIndex)
	}
}

func TestFormatComparison_Output(t *testing.T) {
	cr := ComparisonResult{
		ParamKey:      "wave_schedule.0.difficulty",
		AlertMetricID: "B06",
		Rows: []ComparisonRow{
			{ParamValue: "1.5", WinRate: 0.60, AlertMetricValue: 0.30, AlertResolved: false},
			{ParamValue: "2.0", WinRate: 0.55, AlertMetricValue: 0.22, AlertResolved: true},
			{
				ParamValue: "3.0", WinRate: 0.28, AlertMetricValue: 0.04, AlertResolved: true,
				NewAlerts: []metrics.BreakageAlert{
					{MetricID: "B07", BrokenSign: "early wipe before midgame"},
				},
			},
		},
		BestIndex: 1,
	}

	output := FormatComparison(cr)

	// Check header.
	if !strings.Contains(output, "Sweep Results:") {
		t.Error("missing header")
	}

	// Check resolved row has ✅.
	if !strings.Contains(output, "difficulty=2.0") {
		t.Errorf("missing param value in output:\n%s", output)
	}

	// Check new alert warning.
	if !strings.Contains(output, "新たな壊れ") {
		t.Error("missing new alert warning")
	}
	if !strings.Contains(output, "B07") {
		t.Error("missing B07 in new alert")
	}

	// Check Best line.
	if !strings.Contains(output, "Best: difficulty=2.0 (B06 resolved, no new alerts)") {
		t.Errorf("missing Best line in output:\n%s", output)
	}
}

func TestFormatComparison_NoBest(t *testing.T) {
	cr := ComparisonResult{
		ParamKey:      "wave_schedule.0.difficulty",
		AlertMetricID: "B06",
		Rows: []ComparisonRow{
			{ParamValue: "1.5", WinRate: 0.60, AlertMetricValue: 0.30, AlertResolved: false},
		},
		BestIndex: -1,
	}

	output := FormatComparison(cr)

	if !strings.Contains(output, "No parameter value resolved the alert") {
		t.Errorf("expected no-best message, got:\n%s", output)
	}
}

func TestFormatComparison_Empty(t *testing.T) {
	cr := ComparisonResult{}
	output := FormatComparison(cr)
	if !strings.Contains(output, "No sweep results to compare.") {
		t.Errorf("expected empty message, got: %q", output)
	}
}
