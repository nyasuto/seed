package metrics

import (
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// Collector gathers per-tick statistics from GameSnapshots and produces
// a GameSummary when the game ends.
type Collector struct {
	tickCount  int
	peakBeasts int
	peakRooms  int
	lastSnapshot scenario.GameSnapshot
}

// NewCollector creates a new Collector ready to receive tick data.
func NewCollector() *Collector {
	return &Collector{}
}

// OnTick records statistics from a post-tick snapshot and the actions
// that were executed during that tick.
func (c *Collector) OnTick(snapshot scenario.GameSnapshot, actions []simulation.PlayerAction) {
	c.tickCount++
	c.lastSnapshot = snapshot

	if snapshot.BeastCount > c.peakBeasts {
		c.peakBeasts = snapshot.BeastCount
	}

	// Count rooms built from DigRoom actions in this tick.
	for _, a := range actions {
		if a.ActionType() == "dig_room" {
			c.peakRooms++
		}
	}
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
