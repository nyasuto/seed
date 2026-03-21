package metrics

import (
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// waveOverlapWindow is the number of ticks to look back when checking
// whether construction overlapped with a wave arrival.
const waveOverlapWindow = 5

// Collector gathers per-tick statistics from GameSnapshots and produces
// a GameSummary when the game ends.
type Collector struct {
	tickCount    int
	peakBeasts   int
	peakRooms    int
	lastSnapshot scenario.GameSnapshot

	// B01/B02: first wave tracking
	firstWaveRecorded bool
	b01               types.Tick // tick when first wave arrived
	b02               int        // player actions before first wave
	prevSpawnedWaves  int        // SpawnedWaves from previous tick

	// B03: terrain block rate
	digRoomAttempts int
	digRoomBlocked  int
	prevRoomCount   int

	// B04: buildable cells at game start
	initialBuildableCells int
	buildableCellsSet     bool

	// B05: wave overlap
	recentDigTicks  []types.Tick // ticks where DigRoom was attempted
	waveOverlapHits int          // number of wave arrivals overlapping with construction
	waveArrivals    int          // total wave arrivals observed
}

// NewCollector creates a new Collector ready to receive tick data.
func NewCollector() *Collector {
	return &Collector{}
}

// RecordBuildableCells records the number of buildable cells at game start
// for B04 (ZeroBuildableRate) calculation.
func (c *Collector) RecordBuildableCells(count int) {
	c.initialBuildableCells = count
	c.buildableCellsSet = true
}

// OnTick records statistics from a post-tick snapshot and the actions
// that were executed during that tick.
func (c *Collector) OnTick(snapshot scenario.GameSnapshot, actions []simulation.PlayerAction) {
	c.tickCount++

	if snapshot.BeastCount > c.peakBeasts {
		c.peakBeasts = snapshot.BeastCount
	}

	// Count DigRoom actions attempted this tick.
	digRoomThisTick := 0
	playerActionsThisTick := 0
	for _, a := range actions {
		at := a.ActionType()
		if at == "dig_room" {
			c.peakRooms++
			digRoomThisTick++
		}
		if at != "no_action" {
			playerActionsThisTick++
		}
	}

	// B01/B02: detect first wave arrival.
	if !c.firstWaveRecorded {
		c.b02 += playerActionsThisTick
		if snapshot.SpawnedWaves > c.prevSpawnedWaves && c.prevSpawnedWaves == 0 {
			c.firstWaveRecorded = true
			c.b01 = snapshot.Tick - 1 // snapshot.Tick is post-tick (incremented), so actual tick = Tick-1
		}
	}
	c.prevSpawnedWaves = snapshot.SpawnedWaves

	// B03: terrain block rate.
	if digRoomThisTick > 0 {
		c.digRoomAttempts += digRoomThisTick
		// Actual rooms added = post-tick room count - previous room count.
		roomsAdded := snapshot.RoomCount - c.prevRoomCount
		if roomsAdded < 0 {
			roomsAdded = 0
		}
		blocked := digRoomThisTick - roomsAdded
		if blocked < 0 {
			blocked = 0
		}
		c.digRoomBlocked += blocked
	}
	c.prevRoomCount = snapshot.RoomCount

	// B05: track recent DigRoom ticks and wave overlap.
	currentTick := snapshot.Tick - 1 // actual tick that just executed
	if digRoomThisTick > 0 {
		c.recentDigTicks = append(c.recentDigTicks, currentTick)
	}

	// Check if new waves arrived this tick.
	newWaves := snapshot.SpawnedWaves - c.lastSnapshot.SpawnedWaves
	if c.tickCount > 1 && newWaves > 0 {
		c.waveArrivals += newWaves
		// Check if any DigRoom happened within the overlap window.
		if c.hasRecentDig(currentTick) {
			c.waveOverlapHits += newWaves
		}
	}

	// Prune old dig ticks beyond the overlap window.
	c.pruneDigTicks(currentTick)

	c.lastSnapshot = snapshot
}

// hasRecentDig checks if DigRoom was attempted within waveOverlapWindow
// ticks before the given tick.
func (c *Collector) hasRecentDig(currentTick types.Tick) bool {
	for _, dt := range c.recentDigTicks {
		if currentTick-dt <= types.Tick(waveOverlapWindow) {
			return true
		}
	}
	return false
}

// pruneDigTicks removes dig tick entries older than the overlap window.
func (c *Collector) pruneDigTicks(currentTick types.Tick) {
	cutoff := types.Tick(0)
	if currentTick > types.Tick(waveOverlapWindow) {
		cutoff = currentTick - types.Tick(waveOverlapWindow)
	}
	n := 0
	for _, dt := range c.recentDigTicks {
		if dt >= cutoff {
			c.recentDigTicks[n] = dt
			n++
		}
	}
	c.recentDigTicks = c.recentDigTicks[:n]
}

// BreakageMetrics returns the B01-B05 breakage sign metrics collected
// during the game.
func (c *Collector) BreakageMetrics() BreakageData {
	bd := BreakageData{}

	if c.firstWaveRecorded {
		bd.B01 = int(c.b01)
		bd.B02 = c.b02
		bd.FirstWaveRecorded = true
	}

	if c.digRoomAttempts > 0 {
		bd.B03 = float64(c.digRoomBlocked) / float64(c.digRoomAttempts)
	}

	bd.B04ZeroBuildable = c.buildableCellsSet && c.initialBuildableCells == 0

	if c.waveArrivals > 0 {
		bd.B05 = float64(c.waveOverlapHits) / float64(c.waveArrivals)
	}

	return bd
}

// OnGameEnd finalizes collection and returns a GameSummary populated
// from the RunResult and accumulated per-tick data.
func (c *Collector) OnGameEnd(result *simulation.RunResult) *GameSummary {
	return &GameSummary{
		Result:              result.Result.Status,
		Reason:              result.Result.Reason,
		TotalTicks:          result.TickCount,
		RoomsBuilt:          c.peakRooms,
		FinalCoreHP:         c.lastSnapshot.CoreHP,
		PeakChi:             result.Statistics.PeakChi,
		FinalFengShui:       result.Statistics.FinalFengShui,
		WavesDefeated:       result.Statistics.WavesDefeated,
		TotalWaves:          c.lastSnapshot.TotalWaves,
		PeakBeasts:          c.peakBeasts,
		TotalDamageDealt:    result.Statistics.DamageDealt,
		TotalDamageReceived: result.Statistics.DamageReceived,
		DeficitTicks:        result.Statistics.DeficitTicks,
		Evolutions:          result.Statistics.Evolutions,
	}
}
