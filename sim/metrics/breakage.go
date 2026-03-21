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
