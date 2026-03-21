package metrics

import (
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// stompThreshold is the fraction of MaxCoreHP above which a win is considered a stomp.
const stompThreshold = 0.8

// earlyWipeFraction is the fraction of MaxTicks within which a loss is considered early.
const earlyWipeFraction = 0.5

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

	// B06-B09: game-end metrics config
	maxCoreHP       int // maximum core HP (set via RecordGameConfig)
	maxTicks        int // maximum ticks in the scenario (set via RecordGameConfig)
	maxRoomLevel    int // maximum room level (set via RecordGameConfig)
	gameConfigSet   bool
	gameResult      simulation.GameStatus // set via RecordGameResult
	gameResultSet   bool
	finalRoomLevels []int // room levels at game end (set via RecordFinalRoomLevels)
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

// RecordGameConfig records scenario configuration needed for B06-B09.
// maxCoreHP is the starting core HP, maxTicks is the scenario tick limit,
// and maxRoomLevel is the maximum level a room can reach.
func (c *Collector) RecordGameConfig(maxCoreHP, maxTicks, maxRoomLevel int) {
	c.maxCoreHP = maxCoreHP
	c.maxTicks = maxTicks
	c.maxRoomLevel = maxRoomLevel
	c.gameConfigSet = true
}

// RecordGameResult records the final game outcome for B06-B09 calculation.
func (c *Collector) RecordGameResult(status simulation.GameStatus) {
	c.gameResult = status
	c.gameResultSet = true
}

// RecordFinalRoomLevels records the level of each room at game end
// for B08 (PerfectionRate) and B09 (AvgRoomLevelRatio) calculation.
func (c *Collector) RecordFinalRoomLevels(levels []int) {
	c.finalRoomLevels = levels
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

// BreakageMetrics returns the B01-B09 breakage sign metrics collected
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

	// B06: StompRate — won with CoreHP >= 80% of MaxCoreHP.
	if c.gameConfigSet && c.gameResultSet && c.gameResult == simulation.Won && c.maxCoreHP > 0 {
		ratio := float64(c.lastSnapshot.CoreHP) / float64(c.maxCoreHP)
		bd.B06Stomp = ratio >= stompThreshold
	}

	// B07: EarlyWipeRate — lost within the first 50% of MaxTicks.
	if c.gameConfigSet && c.gameResultSet && c.gameResult == simulation.Lost && c.maxTicks > 0 {
		bd.B07EarlyWipe = c.tickCount <= int(float64(c.maxTicks)*earlyWipeFraction)
	}

	// B08/B09: PerfectionRate and AvgRoomLevelRatio (win only).
	if c.gameConfigSet && c.gameResultSet && c.gameResult == simulation.Won && c.maxRoomLevel > 0 && len(c.finalRoomLevels) > 0 {
		allMax := true
		levelSum := 0
		for _, lv := range c.finalRoomLevels {
			levelSum += lv
			if lv < c.maxRoomLevel {
				allMax = false
			}
		}
		bd.B08Perfection = allMax
		bd.B09RoomLevelRatio = float64(levelSum) / float64(len(c.finalRoomLevels)*c.maxRoomLevel)
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
