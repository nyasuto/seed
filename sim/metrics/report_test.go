package metrics

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
)

func TestGenerateJSON_PRDFormat(t *testing.T) {
	config := ReportConfig{
		Scenario: "tutorial.json",
		Games:    10,
		AI:       "simple",
	}

	summaries := []GameSummary{
		{Result: simulation.Won, TotalTicks: 100, RoomsBuilt: 5, FinalCoreHP: 80},
		{Result: simulation.Won, TotalTicks: 120, RoomsBuilt: 6, FinalCoreHP: 60},
		{Result: simulation.Lost, TotalTicks: 80, RoomsBuilt: 3, FinalCoreHP: 0},
		{Result: simulation.Won, TotalTicks: 150, RoomsBuilt: 7, FinalCoreHP: 90},
		{Result: simulation.Won, TotalTicks: 110, RoomsBuilt: 5, FinalCoreHP: 70},
		{Result: simulation.Won, TotalTicks: 130, RoomsBuilt: 6, FinalCoreHP: 50},
		{Result: simulation.Lost, TotalTicks: 60, RoomsBuilt: 2, FinalCoreHP: 0},
		{Result: simulation.Won, TotalTicks: 140, RoomsBuilt: 7, FinalCoreHP: 85},
		{Result: simulation.Won, TotalTicks: 105, RoomsBuilt: 5, FinalCoreHP: 75},
		{Result: simulation.Won, TotalTicks: 115, RoomsBuilt: 6, FinalCoreHP: 65},
	}

	breakageData := make([]BreakageData, 10)
	for i := range breakageData {
		breakageData[i] = BreakageData{
			FirstWaveRecorded: true,
			B01:               30 + i,
			B02:               5 + i,
			B03:               0.1,
			B05:               0.4,
			B11SurplusRate:     0.2,
		}
	}

	report := BreakageReport{
		Alerts: []BreakageAlert{
			{
				MetricID:   "B06",
				BrokenSign: "stomp victory with high HP",
				Value:      0.35,
				Threshold:  0.30,
				Direction:  Above,
			},
		},
		Clean: []string{"B01", "B02", "B03", "B04", "B05", "B07", "B08", "B09", "B10", "B11"},
	}

	result, err := GenerateJSON(config, summaries, breakageData, report)
	if err != nil {
		t.Fatalf("GenerateJSON returned error: %v", err)
	}

	// Verify it's valid JSON.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify top-level keys exist.
	for _, key := range []string{"config", "summary", "breakage_report", "raw_metrics"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	// Verify config section.
	cfg := parsed["config"].(map[string]any)
	if cfg["scenario"] != "tutorial.json" {
		t.Errorf("config.scenario = %v, want tutorial.json", cfg["scenario"])
	}
	if cfg["games"].(float64) != 10 {
		t.Errorf("config.games = %v, want 10", cfg["games"])
	}

	// Verify summary section.
	summary := parsed["summary"].(map[string]any)
	winRate := summary["win_rate"].(float64)
	if winRate != 0.8 {
		t.Errorf("summary.win_rate = %v, want 0.8", winRate)
	}

	// Verify breakage_report section.
	br := parsed["breakage_report"].(map[string]any)
	alerts := br["alerts"].([]any)
	if len(alerts) != 1 {
		t.Errorf("breakage_report.alerts length = %d, want 1", len(alerts))
	}
	clean := br["clean"].([]any)
	if len(clean) != 10 {
		t.Errorf("breakage_report.clean length = %d, want 10", len(clean))
	}

	// Verify raw_metrics section.
	raw := parsed["raw_metrics"].(map[string]any)
	if _, ok := raw["B01_ticks_before_first_wave"]; !ok {
		t.Error("raw_metrics missing B01_ticks_before_first_wave")
	}
	if _, ok := raw["B06_stomp_rate"]; !ok {
		t.Error("raw_metrics missing B06_stomp_rate")
	}
}

func TestGenerateJSON_EmptySummaries(t *testing.T) {
	config := ReportConfig{Scenario: "test.json", Games: 0, AI: "noop"}
	result, err := GenerateJSON(config, nil, nil, BreakageReport{})
	if err != nil {
		t.Fatalf("GenerateJSON returned error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestGenerateCSV_Format(t *testing.T) {
	summaries := []GameSummary{
		{
			Result:              simulation.Won,
			Reason:              "win condition met",
			TotalTicks:          100,
			RoomsBuilt:          5,
			FinalCoreHP:         80,
			PeakChi:             120.5,
			FinalFengShui:       0.85,
			WavesDefeated:       3,
			TotalWaves:          3,
			PeakBeasts:          4,
			TotalDamageDealt:    500,
			TotalDamageReceived: 200,
			DeficitTicks:        10,
			Evolutions:          2,
		},
		{
			Result:      simulation.Lost,
			Reason:      "core HP reached 0",
			TotalTicks:  60,
			RoomsBuilt:  3,
			FinalCoreHP: 0,
		},
	}

	result := GenerateCSV(summaries)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 3 {
		t.Fatalf("CSV has %d lines, want 3 (header + 2 rows)", len(lines))
	}

	// Verify header.
	header := lines[0]
	if !strings.HasPrefix(header, "game,result,reason,") {
		t.Errorf("unexpected header: %s", header)
	}

	// Verify first data row.
	if !strings.HasPrefix(lines[1], "0,Won,") {
		t.Errorf("first row should start with '0,Won,', got: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "1,Lost,") {
		t.Errorf("second row should start with '1,Lost,', got: %s", lines[2])
	}
}

func TestGenerateCSV_EscapeCommaInReason(t *testing.T) {
	summaries := []GameSummary{
		{
			Result: simulation.Won,
			Reason: "condition A, condition B",
		},
	}

	result := GenerateCSV(summaries)
	if !strings.Contains(result, `"condition A, condition B"`) {
		t.Errorf("CSV should escape comma in reason, got: %s", result)
	}
}

func TestPercentile(t *testing.T) {
	vals := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	p10 := percentile(vals, 0.10)
	if p10 < 10 || p10 > 20 {
		t.Errorf("p10 = %v, want between 10 and 20", p10)
	}

	p90 := percentile(vals, 0.90)
	if p90 < 80 || p90 > 100 {
		t.Errorf("p90 = %v, want between 80 and 100", p90)
	}
}

func TestBuildSummary_WinRate(t *testing.T) {
	summaries := []GameSummary{
		{Result: simulation.Won, TotalTicks: 100, RoomsBuilt: 5, FinalCoreHP: 50},
		{Result: simulation.Won, TotalTicks: 200, RoomsBuilt: 10, FinalCoreHP: 100},
		{Result: simulation.Lost, TotalTicks: 50, RoomsBuilt: 2, FinalCoreHP: 0},
		{Result: simulation.Lost, TotalTicks: 80, RoomsBuilt: 3, FinalCoreHP: 0},
	}

	s := buildSummary(summaries)
	if s.WinRate != 0.5 {
		t.Errorf("WinRate = %v, want 0.5", s.WinRate)
	}
	if s.AvgTicks != 107.5 {
		t.Errorf("AvgTicks = %v, want 107.5", s.AvgTicks)
	}
}
