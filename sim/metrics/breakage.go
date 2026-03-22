package metrics

import (
	"math"

	"github.com/nyasuto/seed/core/types"
)

// BreakageData holds the B01-B09 breakage sign metrics for a single game.
type BreakageData struct {
	// B01 is the tick number when the first invasion wave arrived.
	// Only valid when FirstWaveRecorded is true.
	B01 int
	// B02 is the number of player actions taken before the first wave.
	// Only valid when FirstWaveRecorded is true.
	B02 int
	// FirstWaveRecorded indicates whether a wave arrived during the game.
	FirstWaveRecorded bool
	// B03 is the terrain block rate: fraction of DigRoom attempts that
	// were rejected due to terrain constraints (0.0 to 1.0).
	B03 float64
	// B04ZeroBuildable is true if the game started with zero buildable cells.
	B04ZeroBuildable bool
	// B05 is the wave overlap rate: fraction of wave arrivals that occurred
	// within waveOverlapWindow ticks of a DigRoom attempt.
	B05 float64
	// B06Stomp is true if the player won with CoreHP >= 80% of MaxCoreHP.
	B06Stomp bool
	// B07EarlyWipe is true if the player lost within the first 50% of MaxTicks.
	B07EarlyWipe bool
	// B08Perfection is true if the player won with all rooms at MaxRoomLevel.
	B08Perfection bool
	// B09RoomLevelRatio is the average room level / MaxRoomLevel at game end (win only).
	// Only valid when the game was won and MaxRoomLevel > 0.
	B09RoomLevelRatio float64
	// B10LayoutEntropy is the Shannon entropy of room placement positions
	// across multiple games. Computed post-batch via CalcLayoutEntropy,
	// not per-tick. Higher entropy means more diverse placement.
	B10LayoutEntropy float64
	// B11SurplusRate is the fraction of ticks where ChiPool exceeded
	// surplusThreshold of the observed peak ChiPool. Higher values
	// indicate the player had too many resources throughout the game.
	B11SurplusRate float64
}

// AlertDirection indicates whether a metric broke its threshold by being
// above or below the expected range.
type AlertDirection string

const (
	// Above means the metric value exceeded the upper threshold.
	Above AlertDirection = "above"
	// Below means the metric value fell below the lower threshold.
	Below AlertDirection = "below"
)

// BreakageAlert represents a single metric that violated its threshold.
type BreakageAlert struct {
	// MetricID is the breakage sign identifier (e.g. "B01", "B02").
	MetricID string
	// BrokenSign describes what the violation means for game balance.
	BrokenSign string
	// Value is the observed metric value.
	Value float64
	// Threshold is the threshold that was violated.
	Threshold float64
	// Direction indicates whether Value was above or below Threshold.
	Direction AlertDirection
}

// BreakageReport contains the results of threshold checking across all metrics.
type BreakageReport struct {
	// Alerts lists metrics that violated their thresholds.
	Alerts []BreakageAlert
	// Clean lists metric IDs that passed their threshold checks.
	Clean []string
}

// BreakageThresholds holds configurable thresholds for breakage detection.
// Most thresholds are fixed from D002, but B01 and B10 are scenario-dependent.
type BreakageThresholds struct {
	// B01MinGraceTicks is the minimum number of ticks before the first wave.
	// Scenario-dependent. If 0, B01 check is skipped.
	B01MinGraceTicks int
	// B10MinEntropy is the minimum layout entropy threshold.
	// Scenario-dependent. If 0, B10 check is skipped.
	B10MinEntropy float64
}

// Fixed thresholds from D002.
const (
	thresholdB02MinActions    = 3
	thresholdB03MinBlockRate  = 0.05
	thresholdB05MinOverlap    = 0.30
	thresholdB09MaxLevelRatio = 0.80
	thresholdB11MaxSurplus    = 0.50
)

// allMetricIDs lists all breakage sign metric identifiers in order.
var allMetricIDs = []string{"B01", "B02", "B03", "B04", "B05", "B06", "B07", "B08", "B09", "B10", "B11"}

// DetectBreakage checks all B01-B11 metrics against their thresholds and
// returns a BreakageReport with alerts for violations and clean IDs for passes.
func DetectBreakage(bd BreakageData, th BreakageThresholds) BreakageReport {
	var alerts []BreakageAlert
	alerted := make(map[string]bool)

	// B01: first wave arrives too early (below minimum grace ticks).
	if th.B01MinGraceTicks > 0 && bd.FirstWaveRecorded {
		if float64(bd.B01) < float64(th.B01MinGraceTicks) {
			alerts = append(alerts, BreakageAlert{
				MetricID:   "B01",
				BrokenSign: "enemy arrives before planning",
				Value:      float64(bd.B01),
				Threshold:  float64(th.B01MinGraceTicks),
				Direction:  Below,
			})
			alerted["B01"] = true
		}
	}

	// B02: too few actions before first wave (< 3).
	if bd.FirstWaveRecorded {
		if float64(bd.B02) < thresholdB02MinActions {
			alerts = append(alerts, BreakageAlert{
				MetricID:   "B02",
				BrokenSign: "insufficient actions before first wave",
				Value:      float64(bd.B02),
				Threshold:  thresholdB02MinActions,
				Direction:  Below,
			})
			alerted["B02"] = true
		}
	}

	// B03: terrain block rate too low (< 0.05) — no constraint fun.
	if bd.B03 < thresholdB03MinBlockRate {
		alerts = append(alerts, BreakageAlert{
			MetricID:   "B03",
			BrokenSign: "terrain constraints too weak",
			Value:      bd.B03,
			Threshold:  thresholdB03MinBlockRate,
			Direction:  Below,
		})
		alerted["B03"] = true
	}

	// B04: zero buildable cells — game starts stuck.
	if bd.B04ZeroBuildable {
		alerts = append(alerts, BreakageAlert{
			MetricID:   "B04",
			BrokenSign: "zero buildable cells at start",
			Value:      1.0,
			Threshold:  0.0,
			Direction:  Above,
		})
		alerted["B04"] = true
	}

	// B05: wave overlap rate too low (< 0.30) — no interrupted fun.
	if bd.B05 < thresholdB05MinOverlap {
		alerts = append(alerts, BreakageAlert{
			MetricID:   "B05",
			BrokenSign: "waves do not overlap with construction",
			Value:      bd.B05,
			Threshold:  thresholdB05MinOverlap,
			Direction:  Below,
		})
		alerted["B05"] = true
	}

	// B06: stomp victory — won too easily.
	if bd.B06Stomp {
		alerts = append(alerts, BreakageAlert{
			MetricID:   "B06",
			BrokenSign: "stomp victory with high HP",
			Value:      1.0,
			Threshold:  0.0,
			Direction:  Above,
		})
		alerted["B06"] = true
	}

	// B07: early wipe — lost too quickly.
	if bd.B07EarlyWipe {
		alerts = append(alerts, BreakageAlert{
			MetricID:   "B07",
			BrokenSign: "early wipe before midgame",
			Value:      1.0,
			Threshold:  0.0,
			Direction:  Above,
		})
		alerted["B07"] = true
	}

	// B08: perfection — all rooms at max level.
	if bd.B08Perfection {
		alerts = append(alerts, BreakageAlert{
			MetricID:   "B08",
			BrokenSign: "all rooms reached max level",
			Value:      1.0,
			Threshold:  0.0,
			Direction:  Above,
		})
		alerted["B08"] = true
	}

	// B09: room level ratio too high (> 0.80) — game too easy.
	if bd.B09RoomLevelRatio > thresholdB09MaxLevelRatio {
		alerts = append(alerts, BreakageAlert{
			MetricID:   "B09",
			BrokenSign: "room levels too high at game end",
			Value:      bd.B09RoomLevelRatio,
			Threshold:  thresholdB09MaxLevelRatio,
			Direction:  Above,
		})
		alerted["B09"] = true
	}

	// B10: layout entropy too low — rooms always placed in same spots.
	if th.B10MinEntropy > 0 {
		if bd.B10LayoutEntropy < th.B10MinEntropy {
			alerts = append(alerts, BreakageAlert{
				MetricID:   "B10",
				BrokenSign: "low layout diversity across games",
				Value:      bd.B10LayoutEntropy,
				Threshold:  th.B10MinEntropy,
				Direction:  Below,
			})
			alerted["B10"] = true
		}
	}

	// B11: surplus rate too high (> 0.50) — too many resources.
	if bd.B11SurplusRate > thresholdB11MaxSurplus {
		alerts = append(alerts, BreakageAlert{
			MetricID:   "B11",
			BrokenSign: "chi pool always has surplus",
			Value:      bd.B11SurplusRate,
			Threshold:  thresholdB11MaxSurplus,
			Direction:  Above,
		})
		alerted["B11"] = true
	}

	// Build clean list from non-alerted metrics.
	var clean []string
	for _, id := range allMetricIDs {
		if !alerted[id] {
			clean = append(clean, id)
		}
	}

	return BreakageReport{
		Alerts: alerts,
		Clean:  clean,
	}
}

// CalcLayoutEntropy computes the Shannon entropy of room placement positions
// across multiple games. Each element of gameRoomPositions is the list of
// room positions from one game run. The entropy is normalized by log2(N)
// where N is the number of distinct positions observed, yielding a value
// in [0, 1]. A value near 1.0 means rooms are placed in many different
// positions across games; near 0.0 means rooms always end up in the same spots.
// Returns 0.0 if fewer than 2 games or no rooms are provided.
func CalcLayoutEntropy(gameRoomPositions [][]types.Pos) float64 {
	if len(gameRoomPositions) < 2 {
		return 0.0
	}

	// Count how many times each position is used across all games.
	freq := make(map[types.Pos]int)
	total := 0
	for _, positions := range gameRoomPositions {
		for _, pos := range positions {
			freq[pos]++
			total++
		}
	}

	if total == 0 || len(freq) <= 1 {
		return 0.0
	}

	// Shannon entropy: H = -Σ p(x) * log2(p(x))
	var entropy float64
	for _, count := range freq {
		p := float64(count) / float64(total)
		entropy -= p * math.Log2(p)
	}

	// Normalize by maximum possible entropy (uniform distribution over all positions).
	maxEntropy := math.Log2(float64(len(freq)))
	if maxEntropy == 0 {
		return 0.0
	}

	return entropy / maxEntropy
}
