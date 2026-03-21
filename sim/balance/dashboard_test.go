package balance

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/sim/metrics"
)

func TestFormatReport_NoAlerts(t *testing.T) {
	report := metrics.BreakageReport{
		Alerts: nil,
		Clean:  []string{"B01", "B02", "B03", "B04", "B05", "B06", "B07", "B08", "B09", "B10", "B11"},
	}
	data := []metrics.BreakageData{
		{
			FirstWaveRecorded: true,
			B01:               32,
			B02:               8,
			B03:               0.18,
			B05:               0.58,
			B09RoomLevelRatio: 0.42,
			B10LayoutEntropy:  0.81,
			B11SurplusRate:    0.15,
		},
	}

	output := FormatReport("tutorial.json", 500, 0.62, 187, report, data)

	// Check header.
	if !strings.Contains(output, "=== Balance Dashboard: tutorial.json ===") {
		t.Error("missing dashboard header")
	}
	// Check baseline stats.
	if !strings.Contains(output, "Baseline (500 games):") {
		t.Error("missing baseline header")
	}
	if !strings.Contains(output, "Win Rate: 62.0%") {
		t.Error("missing win rate")
	}
	if !strings.Contains(output, "Avg Ticks: 187") {
		t.Error("missing avg ticks")
	}
	// Check "No breakage detected".
	if !strings.Contains(output, "No breakage detected.") {
		t.Errorf("expected 'No breakage detected.' but got:\n%s", output)
	}
	// All metrics should be ✅.
	for _, id := range allMetricIDs {
		if !strings.Contains(output, "✅ "+id) {
			t.Errorf("expected ✅ for %s", id)
		}
	}
}

func TestFormatReport_WithAlerts(t *testing.T) {
	report := metrics.BreakageReport{
		Alerts: []metrics.BreakageAlert{
			{
				MetricID:   "B06",
				BrokenSign: "stomp victory with high HP",
				Value:      0.35,
				Threshold:  0.0,
				Direction:  metrics.Above,
			},
			{
				MetricID:   "B07",
				BrokenSign: "early wipe before midgame",
				Value:      0.25,
				Threshold:  0.0,
				Direction:  metrics.Above,
			},
		},
		Clean: []string{"B01", "B02", "B03", "B04", "B05", "B08", "B09", "B10", "B11"},
	}
	data := []metrics.BreakageData{
		{
			FirstWaveRecorded: true,
			B01:               32,
			B02:               8,
			B03:               0.18,
			B05:               0.58,
			B06Stomp:          true,
			B07EarlyWipe:      true,
			B09RoomLevelRatio: 0.42,
			B10LayoutEntropy:  0.81,
			B11SurplusRate:    0.15,
		},
	}

	output := FormatReport("tutorial.json", 500, 0.62, 187, report, data)

	// Check alert count.
	if !strings.Contains(output, "2 alert(s) detected.") {
		t.Errorf("expected '2 alert(s) detected.' but got:\n%s", output)
	}

	// Check alert formatting with 🔴 and broken sign.
	if !strings.Contains(output, "🔴 B06") {
		t.Error("expected 🔴 for B06")
	}
	if !strings.Contains(output, "stomp victory with high HP") {
		t.Error("expected broken sign for B06")
	}
	if !strings.Contains(output, "🔴 B07") {
		t.Error("expected 🔴 for B07")
	}
	if !strings.Contains(output, "early wipe before midgame") {
		t.Error("expected broken sign for B07")
	}

	// Check clean metrics have ✅.
	for _, id := range []string{"B01", "B02", "B03", "B04", "B05", "B08", "B09", "B10", "B11"} {
		if !strings.Contains(output, "✅ "+id) {
			t.Errorf("expected ✅ for %s", id)
		}
	}

	// Ensure threshold is displayed.
	if !strings.Contains(output, "(threshold:") {
		t.Error("expected threshold display")
	}
}

func TestFormatReport_PRDFormat(t *testing.T) {
	// Verify the output matches the PRD section 4 format structure.
	report := metrics.BreakageReport{
		Alerts: []metrics.BreakageAlert{
			{
				MetricID:   "B06",
				BrokenSign: "stomp victory with high HP",
				Value:      0.35,
				Threshold:  0.0,
				Direction:  metrics.Above,
			},
		},
		Clean: []string{"B01", "B02", "B03", "B04", "B05", "B07", "B08", "B09", "B10", "B11"},
	}
	data := []metrics.BreakageData{
		{
			FirstWaveRecorded: true,
			B01:               32,
			B02:               8,
			B03:               0.18,
			B05:               0.58,
			B06Stomp:          true,
			B09RoomLevelRatio: 0.42,
			B10LayoutEntropy:  0.81,
			B11SurplusRate:    0.15,
		},
	}

	output := FormatReport("tutorial.json", 500, 0.62, 187, report, data)

	// Verify line-by-line structure.
	lines := strings.Split(output, "\n")
	if len(lines) < 15 {
		t.Fatalf("expected at least 15 lines, got %d:\n%s", len(lines), output)
	}

	// First line: header.
	if !strings.HasPrefix(lines[0], "=== Balance Dashboard:") {
		t.Errorf("line 0: expected header, got %q", lines[0])
	}

	// Check metric names match PRD.
	expectedNames := []string{
		"TicksBeforeFirstWave",
		"ActionsBeforeFirstWave",
		"TerrainBlockRate",
		"ZeroBuildableRate",
		"WaveOverlapRate",
		"StompRate",
		"EarlyWipeRate",
		"PerfectionRate",
		"AvgRoomLevelRatio",
		"LayoutEntropy",
		"ResourceSurplusRate",
	}
	for _, name := range expectedNames {
		if !strings.Contains(output, name) {
			t.Errorf("expected metric name %q in output", name)
		}
	}
}

func TestNewDashboard_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  DashboardConfig
		wantErr string
	}{
		{
			name:    "nil scenario",
			config:  DashboardConfig{Games: 10, Output: &strings.Builder{}},
			wantErr: "scenario must not be nil",
		},
		{
			name:    "zero games",
			config:  DashboardConfig{Scenario: &dummyScenario, Games: 0, Output: &strings.Builder{}},
			wantErr: "games must be positive",
		},
		{
			name:    "nil output",
			config:  DashboardConfig{Scenario: &dummyScenario, Games: 10},
			wantErr: "output writer must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDashboard(tt.config)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestMetricValues_AveragesCorrectly(t *testing.T) {
	data := []metrics.BreakageData{
		{
			FirstWaveRecorded: true,
			B01:               30,
			B02:               6,
			B03:               0.10,
			B06Stomp:          true,
			B09RoomLevelRatio: 0.40,
			B10LayoutEntropy:  0.80,
			B11SurplusRate:    0.10,
		},
		{
			FirstWaveRecorded: true,
			B01:               40,
			B02:               10,
			B03:               0.20,
			B06Stomp:          false,
			B09RoomLevelRatio: 0.60,
			B10LayoutEntropy:  0.80,
			B11SurplusRate:    0.30,
		},
	}

	values := metricValues(data)

	assertClose := func(id string, expected float64) {
		t.Helper()
		got := values[id]
		if diff := got - expected; diff > 0.001 || diff < -0.001 {
			t.Errorf("%s: expected %.3f, got %.3f", id, expected, got)
		}
	}

	assertClose("B01", 35.0)  // (30+40)/2
	assertClose("B02", 8.0)   // (6+10)/2
	assertClose("B03", 0.15)  // (0.10+0.20)/2
	assertClose("B06", 0.50)  // 1 stomp out of 2
	assertClose("B09", 0.50)  // (0.40+0.60)/2
	assertClose("B11", 0.20)  // (0.10+0.30)/2
}

func TestMetricValues_NoFirstWave(t *testing.T) {
	data := []metrics.BreakageData{
		{FirstWaveRecorded: false, B03: 0.10},
		{FirstWaveRecorded: false, B03: 0.20},
	}

	values := metricValues(data)

	if values["B01"] != 0 {
		t.Errorf("B01: expected 0 for no first wave, got %f", values["B01"])
	}
	if values["B02"] != 0 {
		t.Errorf("B02: expected 0 for no first wave, got %f", values["B02"])
	}
}

// dummyScenario is a minimal scenario for validation tests.
// We only need a non-nil pointer; the value is not used.
var dummyScenario = makeDummyScenario()

func makeDummyScenario() scenario.Scenario {
	return scenario.Scenario{}
}
