package balance

import (
	"fmt"
	"io"
	"strings"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/adapter/batch"
	"github.com/nyasuto/seed/sim/metrics"
)

// metricName maps metric IDs to human-readable names matching PRD section 4.
var metricName = map[string]string{
	"B01": "TicksBeforeFirstWave",
	"B02": "ActionsBeforeFirstWave",
	"B03": "TerrainBlockRate",
	"B04": "ZeroBuildableRate",
	"B05": "WaveOverlapRate",
	"B06": "StompRate",
	"B07": "EarlyWipeRate",
	"B08": "PerfectionRate",
	"B09": "AvgRoomLevelRatio",
	"B10": "LayoutEntropy",
	"B11": "ResourceSurplusRate",
}

// metricThresholdDesc describes the threshold condition for display.
var metricThresholdDesc = map[string]string{
	"B01": "> min_grace",
	"B02": ">= 3",
	"B03": ">= 0.05",
	"B04": "= false",
	"B05": ">= 0.30",
	"B06": "= false",
	"B07": "= false",
	"B08": "= false",
	"B09": "<= 0.80",
	"B10": ">= min",
	"B11": "<= 0.50",
}

// allMetricIDs lists all breakage sign metric identifiers in order.
var allMetricIDs = []string{"B01", "B02", "B03", "B04", "B05", "B06", "B07", "B08", "B09", "B10", "B11"}

// DashboardConfig holds configuration for a dashboard run.
type DashboardConfig struct {
	// Scenario is the scenario to test.
	Scenario *scenario.Scenario
	// ScenarioName is the display name for the scenario.
	ScenarioName string
	// Games is the number of baseline games to run.
	Games int
	// AI is the AI strategy for batch execution.
	AI batch.AIType
	// BaseSeed is the base RNG seed for the batch.
	BaseSeed int64
	// Output is the writer for dashboard display (typically os.Stdout).
	Output io.Writer
	// Input is the reader for interactive prompts (typically os.Stdin).
	Input io.Reader
}

// Dashboard orchestrates baseline batch execution and breakage sign display.
type Dashboard struct {
	config DashboardConfig
}

// NewDashboard creates a Dashboard with the given configuration.
func NewDashboard(config DashboardConfig) (*Dashboard, error) {
	if config.Scenario == nil {
		return nil, fmt.Errorf("scenario must not be nil")
	}
	if config.Games <= 0 {
		return nil, fmt.Errorf("games must be positive, got %d", config.Games)
	}
	if config.Output == nil {
		return nil, fmt.Errorf("output writer must not be nil")
	}
	return &Dashboard{config: config}, nil
}

// BaselineResult holds the results of the baseline batch execution.
type BaselineResult struct {
	// BatchResult is the raw batch execution result.
	BatchResult batch.BatchResult
	// WinRate is the fraction of games won.
	WinRate float64
	// AvgTicks is the average number of ticks per game.
	AvgTicks float64
}

// Run executes the baseline batch and displays the breakage report.
// It returns the baseline result for use by subsequent dashboard steps.
func (d *Dashboard) Run() (*BaselineResult, error) {
	runner, err := batch.NewBatchRunner(batch.BatchConfig{
		Scenario: d.config.Scenario,
		Games:    d.config.Games,
		BaseSeed: d.config.BaseSeed,
		AI:       d.config.AI,
	})
	if err != nil {
		return nil, fmt.Errorf("creating batch runner: %w", err)
	}

	result, err := runner.Run()
	if err != nil {
		return nil, fmt.Errorf("running baseline batch: %w", err)
	}

	winRate, avgTicks := calcSummaryStats(result.Summaries)

	baseline := &BaselineResult{
		BatchResult: *result,
		WinRate:     winRate,
		AvgTicks:    avgTicks,
	}

	d.printReport(baseline)
	return baseline, nil
}

// FormatReport formats the dashboard output as a string without executing a batch.
// This is useful for testing the display format with pre-computed results.
func FormatReport(scenarioName string, games int, winRate, avgTicks float64, report metrics.BreakageReport, breakageData []metrics.BreakageData) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "=== Balance Dashboard: %s ===\n", scenarioName)
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "Baseline (%d games):\n", games)
	fmt.Fprintf(&sb, "  Win Rate: %.1f%%   Avg Ticks: %.0f\n", winRate*100, avgTicks)
	sb.WriteString("\n")
	sb.WriteString("Breakage Report:\n")

	// Build a map of alerted metrics for quick lookup.
	alertMap := make(map[string]metrics.BreakageAlert)
	for _, a := range report.Alerts {
		alertMap[a.MetricID] = a
	}

	// Build value map from breakage data.
	values := metricValues(breakageData)

	for _, id := range allMetricIDs {
		name := metricName[id]
		threshDesc := metricThresholdDesc[id]

		if alert, ok := alertMap[id]; ok {
			fmt.Fprintf(&sb, "  🔴 %-3s %-26s %s   (threshold: %s) ← %s\n",
				id, name, formatValue(alert.Value), threshDesc, alert.BrokenSign)
		} else {
			val := values[id]
			fmt.Fprintf(&sb, "  ✅ %-3s %-26s %s   (threshold: %s)\n",
				id, name, formatValue(val), threshDesc)
		}
	}

	sb.WriteString("\n")
	alertCount := len(report.Alerts)
	if alertCount == 0 {
		sb.WriteString("No breakage detected.\n")
	} else {
		fmt.Fprintf(&sb, "%d alert(s) detected.\n", alertCount)
	}

	return sb.String()
}

func (d *Dashboard) printReport(baseline *BaselineResult) {
	output := FormatReport(
		d.config.ScenarioName,
		d.config.Games,
		baseline.WinRate,
		baseline.AvgTicks,
		baseline.BatchResult.BreakageReport,
		baseline.BatchResult.BreakageData,
	)
	_, _ = fmt.Fprint(d.config.Output, output)
}

func calcSummaryStats(summaries []metrics.GameSummary) (winRate, avgTicks float64) {
	n := len(summaries)
	if n == 0 {
		return 0, 0
	}
	var wins int
	var tickSum float64
	for _, s := range summaries {
		if s.Result == simulation.Won {
			wins++
		}
		tickSum += float64(s.TotalTicks)
	}
	fn := float64(n)
	return float64(wins) / fn, tickSum / fn
}

// metricValues extracts the average metric values from breakage data.
func metricValues(data []metrics.BreakageData) map[string]float64 {
	values := make(map[string]float64)
	n := len(data)
	if n == 0 {
		return values
	}

	fn := float64(n)
	var b01Sum, b02Sum float64
	var b01Count, b02Count int

	for _, d := range data {
		if d.FirstWaveRecorded {
			b01Sum += float64(d.B01)
			b01Count++
			b02Sum += float64(d.B02)
			b02Count++
		}
		values["B03"] += d.B03
		if d.B04ZeroBuildable {
			values["B04"]++
		}
		values["B05"] += d.B05
		if d.B06Stomp {
			values["B06"]++
		}
		if d.B07EarlyWipe {
			values["B07"]++
		}
		if d.B08Perfection {
			values["B08"]++
		}
		values["B09"] += d.B09RoomLevelRatio
		values["B11"] += d.B11SurplusRate
	}

	if b01Count > 0 {
		values["B01"] = b01Sum / float64(b01Count)
	}
	if b02Count > 0 {
		values["B02"] = b02Sum / float64(b02Count)
	}
	values["B03"] /= fn
	values["B04"] /= fn
	values["B05"] /= fn
	values["B06"] /= fn
	values["B07"] /= fn
	values["B08"] /= fn
	values["B09"] /= fn
	values["B11"] /= fn

	// B10 is computed separately (layout entropy), use first entry if available.
	if n > 0 {
		values["B10"] = data[0].B10LayoutEntropy
	}

	return values
}

func formatValue(v float64) string {
	if v == float64(int(v)) && v < 1000 {
		return fmt.Sprintf("%.1f", v)
	}
	return fmt.Sprintf("%.2f", v)
}
