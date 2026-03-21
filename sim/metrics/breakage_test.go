package metrics

import (
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

func TestB01_TicksBeforeFirstWave(t *testing.T) {
	c := NewCollector()

	// Simulate 30 ticks with no waves. Post-tick snapshot Tick = 1..30.
	for tick := 1; tick <= 30; tick++ {
		c.OnTick(scenario.GameSnapshot{
			Tick:         types.Tick(tick),
			SpawnedWaves: 0,
		}, nil)
	}

	// Wave spawns during tick 30: post-tick snapshot has Tick=31, SpawnedWaves=1.
	c.OnTick(scenario.GameSnapshot{
		Tick:         31,
		SpawnedWaves: 1,
	}, nil)

	bd := c.BreakageMetrics()
	if !bd.FirstWaveRecorded {
		t.Fatal("FirstWaveRecorded should be true")
	}
	if bd.B01 != 30 {
		t.Errorf("B01 = %d, want 30", bd.B01)
	}
}

func TestB02_ActionsBeforeFirstWave(t *testing.T) {
	c := NewCollector()

	// 3 DigRoom actions before wave arrival.
	c.OnTick(scenario.GameSnapshot{Tick: 2, SpawnedWaves: 0, RoomCount: 1}, []simulation.PlayerAction{
		mockAction{actionType: "dig_room"},
	})
	c.OnTick(scenario.GameSnapshot{Tick: 5, SpawnedWaves: 0, RoomCount: 2}, []simulation.PlayerAction{
		mockAction{actionType: "dig_room"},
	})
	c.OnTick(scenario.GameSnapshot{Tick: 10, SpawnedWaves: 0, RoomCount: 3}, []simulation.PlayerAction{
		mockAction{actionType: "dig_room"},
	})

	// No-action ticks.
	c.OnTick(scenario.GameSnapshot{Tick: 20, SpawnedWaves: 0, RoomCount: 3}, []simulation.PlayerAction{
		mockAction{actionType: "no_action"},
	})

	// Wave arrives.
	c.OnTick(scenario.GameSnapshot{Tick: 31, SpawnedWaves: 1, RoomCount: 3}, nil)

	bd := c.BreakageMetrics()
	if !bd.FirstWaveRecorded {
		t.Fatal("FirstWaveRecorded should be true")
	}
	if bd.B02 != 3 {
		t.Errorf("B02 = %d, want 3", bd.B02)
	}
}

func TestB03_TerrainBlockRate(t *testing.T) {
	c := NewCollector()

	// 7 successful DigRoom attempts (room count increases each time).
	for i := 1; i <= 7; i++ {
		c.OnTick(scenario.GameSnapshot{
			Tick:      types.Tick(i),
			RoomCount: i,
		}, []simulation.PlayerAction{
			mockAction{actionType: "dig_room"},
		})
	}

	// 3 blocked DigRoom attempts (room count stays at 7).
	for i := 8; i <= 10; i++ {
		c.OnTick(scenario.GameSnapshot{
			Tick:      types.Tick(i),
			RoomCount: 7,
		}, []simulation.PlayerAction{
			mockAction{actionType: "dig_room"},
		})
	}

	bd := c.BreakageMetrics()
	if bd.B03 != 0.3 {
		t.Errorf("B03 = %f, want 0.3", bd.B03)
	}
}

func TestB04_ZeroBuildable(t *testing.T) {
	c := NewCollector()
	c.RecordBuildableCells(0)

	bd := c.BreakageMetrics()
	if !bd.B04ZeroBuildable {
		t.Error("B04ZeroBuildable should be true when buildable cells = 0")
	}

	c2 := NewCollector()
	c2.RecordBuildableCells(10)

	bd2 := c2.BreakageMetrics()
	if bd2.B04ZeroBuildable {
		t.Error("B04ZeroBuildable should be false when buildable cells > 0")
	}
}

func TestB05_WaveOverlapRate(t *testing.T) {
	c := NewCollector()

	// DigRoom at tick 25 (post-tick snapshot Tick=26).
	c.OnTick(scenario.GameSnapshot{
		Tick:         26,
		SpawnedWaves: 0,
		RoomCount:    1,
	}, []simulation.PlayerAction{
		mockAction{actionType: "dig_room"},
	})

	// No activity for a few ticks.
	c.OnTick(scenario.GameSnapshot{Tick: 28, SpawnedWaves: 0, RoomCount: 1}, nil)

	// Wave arrives at tick 30 (post-tick snapshot Tick=31).
	// Tick 30 - tick 25 = 5, within the overlap window.
	c.OnTick(scenario.GameSnapshot{
		Tick:         31,
		SpawnedWaves: 1,
		RoomCount:    1,
	}, nil)

	bd := c.BreakageMetrics()
	if bd.B05 != 1.0 {
		t.Errorf("B05 = %f, want 1.0 (dig at tick 25, wave at tick 30)", bd.B05)
	}
}

func TestB05_WaveOverlapRate_NoOverlap(t *testing.T) {
	c := NewCollector()

	// DigRoom at tick 10 (post-tick snapshot Tick=11).
	c.OnTick(scenario.GameSnapshot{
		Tick:         11,
		SpawnedWaves: 0,
		RoomCount:    1,
	}, []simulation.PlayerAction{
		mockAction{actionType: "dig_room"},
	})

	// Many ticks pass with no activity.
	for tick := 12; tick <= 30; tick++ {
		c.OnTick(scenario.GameSnapshot{
			Tick:         types.Tick(tick),
			SpawnedWaves: 0,
			RoomCount:    1,
		}, nil)
	}

	// Wave arrives at tick 30 (post-tick Tick=31). Far from dig at tick 10.
	c.OnTick(scenario.GameSnapshot{
		Tick:         31,
		SpawnedWaves: 1,
		RoomCount:    1,
	}, nil)

	bd := c.BreakageMetrics()
	if bd.B05 != 0.0 {
		t.Errorf("B05 = %f, want 0.0 (no overlap)", bd.B05)
	}
}

func TestB01_NoWave(t *testing.T) {
	c := NewCollector()

	for tick := 1; tick <= 10; tick++ {
		c.OnTick(scenario.GameSnapshot{
			Tick:         types.Tick(tick),
			SpawnedWaves: 0,
		}, nil)
	}

	bd := c.BreakageMetrics()
	if bd.FirstWaveRecorded {
		t.Error("FirstWaveRecorded should be false when no wave appeared")
	}
}
