package balance

import (
	"fmt"
	"strings"

	"github.com/nyasuto/seed/sim/metrics"
)

// SweepSuggestion represents a suggested parameter sweep for resolving a breakage alert.
type SweepSuggestion struct {
	// MetricID is the breakage sign that triggered this suggestion.
	MetricID string
	// BrokenSign is the human-readable description of the breakage.
	BrokenSign string
	// AdjustmentDirection describes the D002 adjustment direction.
	AdjustmentDirection string
	// ParamKey is the dotted path into the scenario JSON to sweep.
	ParamKey string
	// Values is the list of suggested sweep values.
	Values []string
}

// adjustmentRule maps a metric ID to its D002 adjustment direction and suggested parameter sweeps.
type adjustmentRule struct {
	direction   string
	paramKey    string
	valueSets   []string
}

// d002Rules maps each breakage metric to its D002-defined adjustment direction
// and the corresponding scenario parameter + sweep values.
var d002Rules = map[string][]adjustmentRule{
	"B01": {
		{
			direction: "初波までの猶予を伸ばす",
			paramKey:  "wave_schedule.0.trigger_tick",
			valueSets: []string{"15", "20", "25", "30", "40"},
		},
	},
	"B02": {
		{
			direction: "初波までの猶予を伸ばす",
			paramKey:  "wave_schedule.0.trigger_tick",
			valueSets: []string{"15", "20", "25", "30", "40"},
		},
	},
	"B03": {
		{
			direction: "地形制約密度を上げる",
			paramKey:  "initial_state.terrain_density",
			valueSets: []string{"0.10", "0.15", "0.20", "0.25", "0.30"},
		},
	},
	"B04": {
		{
			direction: "地形制約密度を下げる",
			paramKey:  "initial_state.terrain_density",
			valueSets: []string{"0.02", "0.05", "0.08", "0.10", "0.12"},
		},
	},
	"B05": {
		{
			direction: "侵入波を構築中に重ねる",
			paramKey:  "wave_schedule.0.trigger_tick",
			valueSets: []string{"8", "10", "12", "15", "18"},
		},
	},
	"B06": {
		{
			direction: "侵入者の強さを上げる",
			paramKey:  "wave_schedule.0.difficulty",
			valueSets: []string{"1.0", "1.5", "2.0", "2.5", "3.0"},
		},
	},
	"B07": {
		{
			direction: "侵入者の強さを下げる",
			paramKey:  "wave_schedule.0.difficulty",
			valueSets: []string{"0.3", "0.5", "0.7", "1.0", "1.2"},
		},
	},
	"B08": {
		{
			direction: "リソース総量を絞る",
			paramKey:  "initial_state.starting_chi",
			valueSets: []string{"50", "100", "150", "200", "300"},
		},
	},
	"B09": {
		{
			direction: "リソース総量を絞る",
			paramKey:  "initial_state.starting_chi",
			valueSets: []string{"50", "100", "150", "200", "300"},
		},
	},
	"B10": {
		{
			direction: "配置パターンの多様化",
			paramKey:  "initial_state.terrain_density",
			valueSets: []string{"0.05", "0.10", "0.15", "0.20", "0.25"},
		},
	},
	"B11": {
		{
			direction: "リソース供給を下げる",
			paramKey:  "initial_state.starting_chi",
			valueSets: []string{"50", "100", "150", "200", "300"},
		},
	},
}

// SuggestSweep generates parameter sweep suggestions for a breakage alert
// based on the D002 adjustment direction table.
func SuggestSweep(alert metrics.BreakageAlert) []SweepSuggestion {
	rules, ok := d002Rules[alert.MetricID]
	if !ok {
		return nil
	}

	suggestions := make([]SweepSuggestion, 0, len(rules))
	for _, rule := range rules {
		suggestions = append(suggestions, SweepSuggestion{
			MetricID:            alert.MetricID,
			BrokenSign:          alert.BrokenSign,
			AdjustmentDirection: rule.direction,
			ParamKey:            rule.paramKey,
			Values:              rule.valueSets,
		})
	}
	return suggestions
}

// FormatSuggestions formats sweep suggestions as a human-readable string
// matching the PRD dashboard output format.
func FormatSuggestions(suggestions []SweepSuggestion) string {
	if len(suggestions) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, s := range suggestions {
		fmt.Fprintf(&sb, "%s: %s\n", s.MetricID, s.BrokenSign)
		fmt.Fprintf(&sb, "  調整の方向: %s\n", s.AdjustmentDirection)
		fmt.Fprintf(&sb, "  Suggested sweep: %s = [%s]\n",
			s.ParamKey, strings.Join(s.Values, ", "))
		sb.WriteString("\n")
	}
	return sb.String()
}
