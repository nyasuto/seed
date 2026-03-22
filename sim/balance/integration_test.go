package balance

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/sim/adapter/batch"
	"github.com/nyasuto/seed/sim/server"
)

// TestIntegration_DashboardFullFlow runs the full dashboard flow:
// baseline execution → breakage report → sweep suggestions → comparison.
// This verifies that all dashboard components work together end-to-end.
func TestIntegration_DashboardFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	var output strings.Builder
	config := DashboardConfig{
		Scenario:     sc,
		ScenarioName: "tutorial",
		Games:        50,
		AI:           batch.AISimple,
		BaseSeed:     42,
		Output:       &output,
	}

	dash, err := NewDashboard(config)
	if err != nil {
		t.Fatalf("NewDashboard: %v", err)
	}

	// Step 1: Run baseline.
	baseline, err := dash.Run()
	if err != nil {
		t.Fatalf("Dashboard.Run: %v", err)
	}

	// Verify baseline result.
	if len(baseline.BatchResult.Summaries) != 50 {
		t.Errorf("expected 50 summaries, got %d", len(baseline.BatchResult.Summaries))
	}
	if baseline.WinRate < 0 || baseline.WinRate > 1 {
		t.Errorf("win rate out of range: %f", baseline.WinRate)
	}
	if baseline.AvgTicks <= 0 {
		t.Errorf("avg ticks should be positive: %f", baseline.AvgTicks)
	}

	// Verify output contains expected sections.
	out := output.String()
	if !strings.Contains(out, "=== Balance Dashboard: tutorial ===") {
		t.Error("missing dashboard header in output")
	}
	if !strings.Contains(out, "Breakage Report:") {
		t.Error("missing breakage report section")
	}

	t.Logf("Baseline: WinRate=%.1f%% AvgTicks=%.0f Alerts=%d",
		baseline.WinRate*100, baseline.AvgTicks,
		len(baseline.BatchResult.BreakageReport.Alerts))

	// Step 2: For each alert, generate suggestions.
	alerts := baseline.BatchResult.BreakageReport.Alerts
	for _, alert := range alerts {
		suggestions := SuggestSweep(alert)
		if len(suggestions) == 0 {
			t.Errorf("no suggestions for alert %s", alert.MetricID)
			continue
		}

		formatted := FormatSuggestions(suggestions)
		if formatted == "" {
			t.Errorf("empty formatted suggestions for %s", alert.MetricID)
		}

		t.Logf("Alert %s: %d suggestion(s)", alert.MetricID, len(suggestions))

		// Step 3: Run sweep for first suggestion and compare.
		s := suggestions[0]
		scenarioJSON, jsonErr := server.LoadBuiltinScenarioJSON("tutorial")
		if jsonErr != nil {
			t.Fatalf("LoadBuiltinScenarioJSON: %v", jsonErr)
		}

		sweepParam := batch.SweepParam{
			Key:    s.ParamKey,
			Values: s.Values,
		}
		baseConfig := batch.BatchConfig{
			Games:    20,
			BaseSeed: 42,
			AI:       batch.AISimple,
		}
		sweepResults, sweepErr := batch.RunSweep(scenarioJSON, sweepParam, baseConfig)
		if sweepErr != nil {
			t.Fatalf("RunSweep(%s): %v", s.ParamKey, sweepErr)
		}

		if len(sweepResults) != len(s.Values) {
			t.Errorf("expected %d sweep results, got %d", len(s.Values), len(sweepResults))
		}

		// Step 4: Compare results.
		comparison := CompareResults(baseline.BatchResult.BreakageReport, alert.MetricID, sweepResults)
		if comparison.AlertMetricID != alert.MetricID {
			t.Errorf("comparison metric = %s, want %s", comparison.AlertMetricID, alert.MetricID)
		}
		if len(comparison.Rows) != len(sweepResults) {
			t.Errorf("comparison rows = %d, want %d", len(comparison.Rows), len(sweepResults))
		}

		formatted = FormatComparison(comparison)
		if !strings.Contains(formatted, "Sweep Results:") {
			t.Error("missing sweep results header in comparison output")
		}

		t.Logf("  Sweep %s: %d rows, best=%d", s.ParamKey, len(comparison.Rows), comparison.BestIndex)

		// Only test one alert's sweep to keep test duration reasonable.
		break
	}
}

// TestIntegration_DashboardNoAlerts verifies the dashboard works when
// all metrics are clean (no alerts to process).
func TestIntegration_DashboardNoAlerts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	// Run with noop AI — results may differ but dashboard should still work.
	var output strings.Builder
	config := DashboardConfig{
		Scenario:     sc,
		ScenarioName: "tutorial",
		Games:        10,
		AI:           batch.AINoop,
		BaseSeed:     42,
		Output:       &output,
	}

	dash, err := NewDashboard(config)
	if err != nil {
		t.Fatalf("NewDashboard: %v", err)
	}

	baseline, err := dash.Run()
	if err != nil {
		t.Fatalf("Dashboard.Run: %v", err)
	}

	// Verify the dashboard ran and produced output.
	out := output.String()
	if !strings.Contains(out, "=== Balance Dashboard:") {
		t.Error("missing dashboard header")
	}

	t.Logf("NoopAI: WinRate=%.1f%% AvgTicks=%.0f Alerts=%d",
		baseline.WinRate*100, baseline.AvgTicks,
		len(baseline.BatchResult.BreakageReport.Alerts))
}
