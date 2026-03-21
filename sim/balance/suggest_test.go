package balance

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/sim/metrics"
)

func TestSuggestSweep_B06(t *testing.T) {
	alert := metrics.BreakageAlert{
		MetricID:   "B06",
		BrokenSign: "stomp victory with high HP",
		Value:      0.35,
		Threshold:  0.0,
		Direction:  metrics.Above,
	}

	suggestions := SuggestSweep(alert)
	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion for B06")
	}

	s := suggestions[0]
	if s.MetricID != "B06" {
		t.Errorf("MetricID: got %q, want %q", s.MetricID, "B06")
	}
	if s.ParamKey != "wave_schedule.0.difficulty" {
		t.Errorf("ParamKey: got %q, want %q", s.ParamKey, "wave_schedule.0.difficulty")
	}
	if len(s.Values) == 0 {
		t.Error("expected sweep values")
	}
	if s.AdjustmentDirection == "" {
		t.Error("expected non-empty adjustment direction")
	}
}

func TestSuggestSweep_AllMetrics(t *testing.T) {
	// Every metric ID in D002 should have at least one suggestion.
	for _, id := range allMetricIDs {
		alert := metrics.BreakageAlert{
			MetricID:   id,
			BrokenSign: "test",
			Value:      1.0,
			Threshold:  0.0,
			Direction:  metrics.Above,
		}
		suggestions := SuggestSweep(alert)
		if len(suggestions) == 0 {
			t.Errorf("no suggestions for metric %s", id)
		}
	}
}

func TestSuggestSweep_UnknownMetric(t *testing.T) {
	alert := metrics.BreakageAlert{
		MetricID: "B99",
	}
	suggestions := SuggestSweep(alert)
	if suggestions != nil {
		t.Errorf("expected nil for unknown metric, got %v", suggestions)
	}
}

func TestFormatSuggestions_B06(t *testing.T) {
	suggestions := []SweepSuggestion{
		{
			MetricID:            "B06",
			BrokenSign:          "stomp victory with high HP",
			AdjustmentDirection: "侵入者の強さを上げる",
			ParamKey:            "wave_schedule.0.difficulty",
			Values:              []string{"1.0", "1.5", "2.0"},
		},
	}

	output := FormatSuggestions(suggestions)

	if !strings.Contains(output, "B06: stomp victory with high HP") {
		t.Error("missing metric ID and broken sign")
	}
	if !strings.Contains(output, "調整の方向: 侵入者の強さを上げる") {
		t.Error("missing adjustment direction")
	}
	if !strings.Contains(output, "Suggested sweep: wave_schedule.0.difficulty = [1.0, 1.5, 2.0]") {
		t.Errorf("missing sweep suggestion, got:\n%s", output)
	}
}

func TestFormatSuggestions_Empty(t *testing.T) {
	output := FormatSuggestions(nil)
	if output != "" {
		t.Errorf("expected empty string for nil suggestions, got %q", output)
	}
}
