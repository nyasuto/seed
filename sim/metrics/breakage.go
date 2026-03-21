package metrics

// BreakageData holds the B01-B05 breakage sign metrics for a single game.
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
}
