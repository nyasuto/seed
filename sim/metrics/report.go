package metrics

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/nyasuto/seed/core/simulation"
)

// ReportConfig holds metadata for the batch run config section of a report.
type ReportConfig struct {
	// Scenario is the scenario name or file path.
	Scenario string
	// Games is the number of games executed.
	Games int
	// AI is the AI strategy used.
	AI string
}

// jsonReport is the top-level JSON structure matching PRD section 3.3.
type jsonReport struct {
	Config         jsonReportConfig   `json:"config"`
	Summary        jsonReportSummary  `json:"summary"`
	BreakageReport jsonBreakageReport `json:"breakage_report"`
	RawMetrics     jsonRawMetrics     `json:"raw_metrics"`
}

type jsonReportConfig struct {
	Scenario string `json:"scenario"`
	Games    int    `json:"games"`
	AI       string `json:"ai"`
}

type jsonReportSummary struct {
	WinRate            float64 `json:"win_rate"`
	AvgTicks           float64 `json:"avg_ticks"`
	AvgRoomsBuilt      float64 `json:"avg_rooms_built"`
	AvgCoreHPRemaining float64 `json:"avg_core_hp_remaining"`
}

type jsonBreakageReport struct {
	Alerts []jsonBreakageAlert `json:"alerts"`
	Clean  []string            `json:"clean"`
}

type jsonBreakageAlert struct {
	MetricID   string  `json:"metric_id"`
	BrokenSign string  `json:"broken_sign"`
	Value      float64 `json:"value"`
	Threshold  float64 `json:"threshold"`
	Direction  string  `json:"direction"`
}

type jsonRawMetrics struct {
	B01 *jsonDistMetric `json:"B01_ticks_before_first_wave,omitempty"`
	B02 *jsonDistMetric `json:"B02_actions_before_first_wave,omitempty"`
	B03 *jsonMeanMetric `json:"B03_terrain_block_rate"`
	B04 float64         `json:"B04_zero_buildable_rate"`
	B05 float64         `json:"B05_wave_overlap_rate"`
	B06 float64         `json:"B06_stomp_rate"`
	B07 float64         `json:"B07_early_wipe_rate"`
	B08 float64         `json:"B08_perfection_rate"`
	B09 float64         `json:"B09_avg_room_level_ratio"`
	B10 float64         `json:"B10_layout_entropy"`
	B11 float64         `json:"B11_resource_surplus_rate"`
}

type jsonDistMetric struct {
	Mean float64 `json:"mean"`
	P10  float64 `json:"p10"`
	P90  float64 `json:"p90"`
}

type jsonMeanMetric struct {
	Mean float64 `json:"mean"`
}

// GenerateJSON produces a JSON report matching PRD section 3.3 format.
func GenerateJSON(config ReportConfig, summaries []GameSummary, breakageData []BreakageData, report BreakageReport) (string, error) {
	jr := jsonReport{
		Config: jsonReportConfig(config),
		Summary:        buildSummary(summaries),
		BreakageReport: buildBreakageReport(report),
		RawMetrics:     buildRawMetrics(breakageData),
	}

	data, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling report JSON: %w", err)
	}
	return string(data), nil
}

// GenerateCSV produces a CSV report of per-game summaries.
// The first row is the header, followed by one row per game.
func GenerateCSV(summaries []GameSummary) string {
	var sb strings.Builder
	sb.WriteString("game,result,reason,total_ticks,rooms_built,final_core_hp,peak_chi,final_feng_shui,waves_defeated,total_waves,peak_beasts,damage_dealt,damage_received,deficit_ticks,evolutions\n")

	for i, s := range summaries {
		fmt.Fprintf(&sb, "%d,%s,%s,%d,%d,%d,%.2f,%.2f,%d,%d,%d,%d,%d,%d,%d\n",
			i,
			s.Result.String(),
			csvEscape(s.Reason),
			s.TotalTicks,
			s.RoomsBuilt,
			s.FinalCoreHP,
			s.PeakChi,
			s.FinalFengShui,
			s.WavesDefeated,
			s.TotalWaves,
			s.PeakBeasts,
			s.TotalDamageDealt,
			s.TotalDamageReceived,
			s.DeficitTicks,
			s.Evolutions,
		)
	}
	return sb.String()
}

func buildSummary(summaries []GameSummary) jsonReportSummary {
	n := len(summaries)
	if n == 0 {
		return jsonReportSummary{}
	}

	var wins int
	var tickSum, roomSum, hpSum float64
	for _, s := range summaries {
		if s.Result == simulation.Won {
			wins++
		}
		tickSum += float64(s.TotalTicks)
		roomSum += float64(s.RoomsBuilt)
		hpSum += float64(s.FinalCoreHP)
	}

	fn := float64(n)
	return jsonReportSummary{
		WinRate:            roundTo(float64(wins)/fn, 4),
		AvgTicks:           roundTo(tickSum/fn, 1),
		AvgRoomsBuilt:      roundTo(roomSum/fn, 1),
		AvgCoreHPRemaining: roundTo(hpSum/fn, 1),
	}
}

func buildBreakageReport(report BreakageReport) jsonBreakageReport {
	alerts := make([]jsonBreakageAlert, 0, len(report.Alerts))
	for _, a := range report.Alerts {
		alerts = append(alerts, jsonBreakageAlert{
			MetricID:   a.MetricID,
			BrokenSign: a.BrokenSign,
			Value:      a.Value,
			Threshold:  a.Threshold,
			Direction:  string(a.Direction),
		})
	}

	clean := report.Clean
	if clean == nil {
		clean = []string{}
	}

	return jsonBreakageReport{
		Alerts: alerts,
		Clean:  clean,
	}
}

func buildRawMetrics(breakageData []BreakageData) jsonRawMetrics {
	n := len(breakageData)
	if n == 0 {
		return jsonRawMetrics{}
	}

	fn := float64(n)

	// B01/B02: distribution metrics (only for games with first wave).
	var b01Vals, b02Vals []float64
	var b03Sum, b05Sum, b11Sum float64
	var b04Count, b06Count, b07Count, b08Count int
	var b09Sum float64
	var b09Count int

	for _, d := range breakageData {
		if d.FirstWaveRecorded {
			b01Vals = append(b01Vals, float64(d.B01))
			b02Vals = append(b02Vals, float64(d.B02))
		}
		b03Sum += d.B03
		if d.B04ZeroBuildable {
			b04Count++
		}
		b05Sum += d.B05
		if d.B06Stomp {
			b06Count++
		}
		if d.B07EarlyWipe {
			b07Count++
		}
		if d.B08Perfection {
			b08Count++
		}
		if d.B09RoomLevelRatio > 0 {
			b09Sum += d.B09RoomLevelRatio
			b09Count++
		}
		b11Sum += d.B11SurplusRate
	}

	raw := jsonRawMetrics{
		B03: &jsonMeanMetric{Mean: roundTo(b03Sum/fn, 4)},
		B04: roundTo(float64(b04Count)/fn, 4),
		B05: roundTo(b05Sum/fn, 4),
		B06: roundTo(float64(b06Count)/fn, 4),
		B07: roundTo(float64(b07Count)/fn, 4),
		B08: roundTo(float64(b08Count)/fn, 4),
		B10: 0,
		B11: roundTo(b11Sum/fn, 4),
	}

	if b09Count > 0 {
		raw.B09 = roundTo(b09Sum/float64(b09Count), 4)
	}

	if len(b01Vals) > 0 {
		raw.B01 = distMetric(b01Vals)
	}
	if len(b02Vals) > 0 {
		raw.B02 = distMetric(b02Vals)
	}

	return raw
}

// distMetric computes mean, p10, and p90 for a slice of values.
func distMetric(vals []float64) *jsonDistMetric {
	if len(vals) == 0 {
		return nil
	}
	sort.Float64s(vals)

	sum := 0.0
	for _, v := range vals {
		sum += v
	}

	return &jsonDistMetric{
		Mean: roundTo(sum/float64(len(vals)), 1),
		P10:  percentile(vals, 0.10),
		P90:  percentile(vals, 0.90),
	}
}

// percentile returns the p-th percentile of a sorted slice using nearest-rank.
func percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	rank := p * float64(n-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))
	if lower == upper || upper >= n {
		return sorted[lower]
	}
	frac := rank - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

func roundTo(val float64, decimals int) float64 {
	pow := math.Pow10(decimals)
	return math.Round(val*pow) / pow
}

// csvEscape wraps a string in quotes if it contains commas.
func csvEscape(s string) string {
	if strings.Contains(s, ",") || strings.Contains(s, "\"") || strings.Contains(s, "\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}
