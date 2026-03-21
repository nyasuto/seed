package metrics

import (
	"math"
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

func TestB06_StompRate(t *testing.T) {
	c := NewCollector()
	c.RecordGameConfig(100, 200, 5) // maxCoreHP=100

	// Simulate a game where CoreHP stays high.
	c.OnTick(scenario.GameSnapshot{
		Tick:   1,
		CoreHP: 90,
	}, nil)

	c.RecordGameResult(simulation.Won)
	c.RecordFinalRoomLevels([]int{3})

	bd := c.BreakageMetrics()
	if !bd.B06Stomp {
		t.Error("B06Stomp should be true when CoreHP 90/100 (90% >= 80%)")
	}

	// Non-stomp: CoreHP at 70%.
	c2 := NewCollector()
	c2.RecordGameConfig(100, 200, 5)
	c2.OnTick(scenario.GameSnapshot{Tick: 1, CoreHP: 70}, nil)
	c2.RecordGameResult(simulation.Won)
	c2.RecordFinalRoomLevels([]int{3})

	bd2 := c2.BreakageMetrics()
	if bd2.B06Stomp {
		t.Error("B06Stomp should be false when CoreHP 70/100 (70% < 80%)")
	}

	// Loss should not be a stomp even with high HP.
	c3 := NewCollector()
	c3.RecordGameConfig(100, 200, 5)
	c3.OnTick(scenario.GameSnapshot{Tick: 1, CoreHP: 95}, nil)
	c3.RecordGameResult(simulation.Lost)

	bd3 := c3.BreakageMetrics()
	if bd3.B06Stomp {
		t.Error("B06Stomp should be false on a loss")
	}
}

func TestB07_EarlyWipeRate(t *testing.T) {
	c := NewCollector()
	c.RecordGameConfig(100, 200, 5) // maxTicks=200

	// Simulate 80 ticks then lose (80 <= 200*0.5=100 → early wipe).
	for tick := 1; tick <= 80; tick++ {
		c.OnTick(scenario.GameSnapshot{
			Tick: types.Tick(tick),
		}, nil)
	}

	c.RecordGameResult(simulation.Lost)

	bd := c.BreakageMetrics()
	if !bd.B07EarlyWipe {
		t.Error("B07EarlyWipe should be true when lost at tick 80 with maxTicks=200")
	}

	// Late loss should not be early wipe.
	c2 := NewCollector()
	c2.RecordGameConfig(100, 200, 5)
	for tick := 1; tick <= 150; tick++ {
		c2.OnTick(scenario.GameSnapshot{Tick: types.Tick(tick)}, nil)
	}
	c2.RecordGameResult(simulation.Lost)

	bd2 := c2.BreakageMetrics()
	if bd2.B07EarlyWipe {
		t.Error("B07EarlyWipe should be false when lost at tick 150 with maxTicks=200")
	}

	// Win should not be early wipe.
	c3 := NewCollector()
	c3.RecordGameConfig(100, 200, 5)
	for tick := 1; tick <= 50; tick++ {
		c3.OnTick(scenario.GameSnapshot{Tick: types.Tick(tick)}, nil)
	}
	c3.RecordGameResult(simulation.Won)
	c3.RecordFinalRoomLevels([]int{1})

	bd3 := c3.BreakageMetrics()
	if bd3.B07EarlyWipe {
		t.Error("B07EarlyWipe should be false on a win")
	}
}

func TestB08_PerfectionRate(t *testing.T) {
	c := NewCollector()
	c.RecordGameConfig(100, 200, 5) // maxRoomLevel=5

	c.OnTick(scenario.GameSnapshot{Tick: 1, CoreHP: 100}, nil)
	c.RecordGameResult(simulation.Won)
	c.RecordFinalRoomLevels([]int{5, 5, 5}) // all rooms at max

	bd := c.BreakageMetrics()
	if !bd.B08Perfection {
		t.Error("B08Perfection should be true when all rooms at MaxLv 5/5")
	}

	// Not all rooms at max.
	c2 := NewCollector()
	c2.RecordGameConfig(100, 200, 5)
	c2.OnTick(scenario.GameSnapshot{Tick: 1, CoreHP: 100}, nil)
	c2.RecordGameResult(simulation.Won)
	c2.RecordFinalRoomLevels([]int{5, 3, 5})

	bd2 := c2.BreakageMetrics()
	if bd2.B08Perfection {
		t.Error("B08Perfection should be false when not all rooms at MaxLv")
	}
}

func TestB09_AvgRoomLevelRatio(t *testing.T) {
	c := NewCollector()
	c.RecordGameConfig(100, 200, 5) // maxRoomLevel=5

	c.OnTick(scenario.GameSnapshot{Tick: 1, CoreHP: 100}, nil)
	c.RecordGameResult(simulation.Won)
	c.RecordFinalRoomLevels([]int{2, 3, 4}) // avg=3.0, ratio=3.0/5.0=0.6

	bd := c.BreakageMetrics()

	const epsilon = 0.001
	want := 0.6
	if bd.B09RoomLevelRatio < want-epsilon || bd.B09RoomLevelRatio > want+epsilon {
		t.Errorf("B09RoomLevelRatio = %f, want %f", bd.B09RoomLevelRatio, want)
	}

	// Loss should not compute ratio.
	c2 := NewCollector()
	c2.RecordGameConfig(100, 200, 5)
	c2.OnTick(scenario.GameSnapshot{Tick: 1}, nil)
	c2.RecordGameResult(simulation.Lost)
	c2.RecordFinalRoomLevels([]int{2, 3, 4})

	bd2 := c2.BreakageMetrics()
	if bd2.B09RoomLevelRatio != 0.0 {
		t.Errorf("B09RoomLevelRatio should be 0.0 on loss, got %f", bd2.B09RoomLevelRatio)
	}
}

func TestB10_LayoutEntropy_DiversePositions(t *testing.T) {
	// 10 games, each placing rooms in completely different positions.
	games := make([][]types.Pos, 10)
	for i := 0; i < 10; i++ {
		games[i] = []types.Pos{
			{X: i * 3, Y: i * 3},
			{X: i*3 + 1, Y: i*3 + 1},
		}
	}

	entropy := CalcLayoutEntropy(games)
	if entropy < 0.9 {
		t.Errorf("B10 LayoutEntropy = %f, want >= 0.9 for diverse positions", entropy)
	}
}

func TestB10_LayoutEntropy_IdenticalPositions(t *testing.T) {
	// 10 games, all placing rooms in exactly the same positions.
	games := make([][]types.Pos, 10)
	for i := 0; i < 10; i++ {
		games[i] = []types.Pos{
			{X: 5, Y: 5},
			{X: 10, Y: 10},
		}
	}

	entropy := CalcLayoutEntropy(games)
	// With only 2 distinct positions used equally, normalized entropy should still be 1.0
	// because the distribution is uniform over the 2 positions.
	// The key is that the positions are always the same across games.
	if entropy < 0.99 {
		t.Errorf("B10 LayoutEntropy = %f, want ~1.0 for uniform distribution over 2 positions", entropy)
	}
}

func TestB10_LayoutEntropy_SinglePosition(t *testing.T) {
	// All games place rooms in exactly the same single position.
	games := make([][]types.Pos, 10)
	for i := 0; i < 10; i++ {
		games[i] = []types.Pos{{X: 5, Y: 5}}
	}

	entropy := CalcLayoutEntropy(games)
	if entropy != 0.0 {
		t.Errorf("B10 LayoutEntropy = %f, want 0.0 for single position", entropy)
	}
}

func TestB10_LayoutEntropy_TooFewGames(t *testing.T) {
	// Fewer than 2 games should return 0.
	entropy := CalcLayoutEntropy([][]types.Pos{{{X: 1, Y: 1}}})
	if entropy != 0.0 {
		t.Errorf("B10 LayoutEntropy = %f, want 0.0 for single game", entropy)
	}

	entropy = CalcLayoutEntropy(nil)
	if entropy != 0.0 {
		t.Errorf("B10 LayoutEntropy = %f, want 0.0 for nil input", entropy)
	}
}

func TestB10_LayoutEntropy_SkewedDistribution(t *testing.T) {
	// 10 games: 9 use position (0,0), 1 uses (1,1).
	// Entropy should be low (skewed).
	games := make([][]types.Pos, 10)
	for i := 0; i < 9; i++ {
		games[i] = []types.Pos{{X: 0, Y: 0}}
	}
	games[9] = []types.Pos{{X: 1, Y: 1}}

	entropy := CalcLayoutEntropy(games)
	// p(0,0)=0.9, p(1,1)=0.1 → H = -(0.9*log2(0.9) + 0.1*log2(0.1)) ≈ 0.469
	// maxH = log2(2) = 1.0 → normalized ≈ 0.469
	if entropy > 0.6 {
		t.Errorf("B10 LayoutEntropy = %f, want < 0.6 for skewed distribution", entropy)
	}
	if entropy < 0.3 {
		t.Errorf("B10 LayoutEntropy = %f, want > 0.3 for skewed distribution", entropy)
	}
}

func TestB11_SurplusRate_HighSurplus(t *testing.T) {
	c := NewCollector()

	// Simulate a game where ChiPool is always high.
	// Peak is established on tick 1, then surplus for all remaining ticks.
	for tick := 1; tick <= 100; tick++ {
		c.OnTick(scenario.GameSnapshot{
			Tick:           types.Tick(tick),
			ChiPoolBalance: 100.0, // constant high value
		}, nil)
	}

	bd := c.BreakageMetrics()
	if bd.B11SurplusRate != 1.0 {
		t.Errorf("B11SurplusRate = %f, want 1.0 for constant high ChiPool", bd.B11SurplusRate)
	}
}

func TestB11_SurplusRate_LowSurplus(t *testing.T) {
	c := NewCollector()

	// Tick 1: peak at 100
	c.OnTick(scenario.GameSnapshot{
		Tick:           1,
		ChiPoolBalance: 100.0,
	}, nil)

	// Ticks 2-100: ChiPool drops to near zero (below 50% of peak).
	for tick := 2; tick <= 100; tick++ {
		c.OnTick(scenario.GameSnapshot{
			Tick:           types.Tick(tick),
			ChiPoolBalance: 10.0, // 10% of peak
		}, nil)
	}

	bd := c.BreakageMetrics()
	// Only tick 1 was at peak (100 >= 50), ticks 2-100 are below threshold.
	// surplus = 1/100 = 0.01
	const epsilon = 0.001
	want := 1.0 / 100.0
	if math.Abs(bd.B11SurplusRate-want) > epsilon {
		t.Errorf("B11SurplusRate = %f, want %f", bd.B11SurplusRate, want)
	}
}

func TestB11_SurplusRate_NoTicks(t *testing.T) {
	c := NewCollector()
	bd := c.BreakageMetrics()
	if bd.B11SurplusRate != 0.0 {
		t.Errorf("B11SurplusRate = %f, want 0.0 for no ticks", bd.B11SurplusRate)
	}
}

func TestB11_SurplusRate_ZeroChiPool(t *testing.T) {
	c := NewCollector()

	// All ticks have zero ChiPool.
	for tick := 1; tick <= 50; tick++ {
		c.OnTick(scenario.GameSnapshot{
			Tick:           types.Tick(tick),
			ChiPoolBalance: 0.0,
		}, nil)
	}

	bd := c.BreakageMetrics()
	if bd.B11SurplusRate != 0.0 {
		t.Errorf("B11SurplusRate = %f, want 0.0 for zero ChiPool", bd.B11SurplusRate)
	}
}

// --- DetectBreakage tests ---

func TestDetectBreakage_AllClean(t *testing.T) {
	// A BreakageData with no violations.
	bd := BreakageData{
		B01:               50,
		B02:               10,
		FirstWaveRecorded: true,
		B03:               0.15, // above 0.05
		B04ZeroBuildable:  false,
		B05:               0.50, // above 0.30
		B06Stomp:          false,
		B07EarlyWipe:      false,
		B08Perfection:     false,
		B09RoomLevelRatio: 0.60, // below 0.80
		B10LayoutEntropy:  0.80, // above threshold
		B11SurplusRate:    0.30, // below 0.50
	}
	th := BreakageThresholds{
		B01MinGraceTicks: 20,
		B10MinEntropy:    0.50,
	}

	report := DetectBreakage(bd, th)
	if len(report.Alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d: %+v", len(report.Alerts), report.Alerts)
	}
	if len(report.Clean) != 11 {
		t.Errorf("expected 11 clean IDs, got %d: %v", len(report.Clean), report.Clean)
	}
}

func TestDetectBreakage_B01_TooEarly(t *testing.T) {
	bd := BreakageData{
		B01:               10,
		B02:               5,
		FirstWaveRecorded: true,
		B03:               0.10,
		B05:               0.50,
		B09RoomLevelRatio: 0.50,
		B11SurplusRate:    0.20,
	}
	th := BreakageThresholds{B01MinGraceTicks: 20}

	report := DetectBreakage(bd, th)
	found := findAlert(report, "B01")
	if found == nil {
		t.Fatal("expected B01 alert")
	}
	if found.Value != 10 {
		t.Errorf("B01 alert Value = %f, want 10", found.Value)
	}
	if found.Direction != Below {
		t.Errorf("B01 alert Direction = %s, want below", found.Direction)
	}
}

func TestDetectBreakage_B01_SkippedWhenThresholdZero(t *testing.T) {
	bd := BreakageData{
		B01:               5,
		B02:               5,
		FirstWaveRecorded: true,
		B03:               0.10,
		B05:               0.50,
		B11SurplusRate:    0.20,
	}
	th := BreakageThresholds{B01MinGraceTicks: 0} // skip

	report := DetectBreakage(bd, th)
	if found := findAlert(report, "B01"); found != nil {
		t.Error("B01 should not alert when threshold is 0")
	}
}

func TestDetectBreakage_B02_TooFewActions(t *testing.T) {
	bd := BreakageData{
		B02:               2,
		FirstWaveRecorded: true,
		B03:               0.10,
		B05:               0.50,
		B11SurplusRate:    0.20,
	}
	th := BreakageThresholds{}

	report := DetectBreakage(bd, th)
	found := findAlert(report, "B02")
	if found == nil {
		t.Fatal("expected B02 alert")
	}
	if found.Value != 2 {
		t.Errorf("B02 alert Value = %f, want 2", found.Value)
	}
}

func TestDetectBreakage_B02_NoWaveNoAlert(t *testing.T) {
	bd := BreakageData{
		B02:               0,
		FirstWaveRecorded: false,
		B03:               0.10,
		B05:               0.50,
		B11SurplusRate:    0.20,
	}
	th := BreakageThresholds{}

	report := DetectBreakage(bd, th)
	if found := findAlert(report, "B02"); found != nil {
		t.Error("B02 should not alert when no wave was recorded")
	}
}

func TestDetectBreakage_BooleanMetrics(t *testing.T) {
	tests := []struct {
		name     string
		bd       BreakageData
		metricID string
	}{
		{
			name:     "B04 zero buildable",
			bd:       BreakageData{B04ZeroBuildable: true, B03: 0.10, B05: 0.50, B11SurplusRate: 0.20},
			metricID: "B04",
		},
		{
			name:     "B06 stomp",
			bd:       BreakageData{B06Stomp: true, B03: 0.10, B05: 0.50, B11SurplusRate: 0.20},
			metricID: "B06",
		},
		{
			name:     "B07 early wipe",
			bd:       BreakageData{B07EarlyWipe: true, B03: 0.10, B05: 0.50, B11SurplusRate: 0.20},
			metricID: "B07",
		},
		{
			name:     "B08 perfection",
			bd:       BreakageData{B08Perfection: true, B03: 0.10, B05: 0.50, B11SurplusRate: 0.20},
			metricID: "B08",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := DetectBreakage(tt.bd, BreakageThresholds{})
			if found := findAlert(report, tt.metricID); found == nil {
				t.Errorf("expected %s alert", tt.metricID)
			}
		})
	}
}

func TestDetectBreakage_B03_TooLow(t *testing.T) {
	bd := BreakageData{
		B03:            0.02, // below 0.05
		B05:            0.50,
		B11SurplusRate: 0.20,
	}

	report := DetectBreakage(bd, BreakageThresholds{})
	found := findAlert(report, "B03")
	if found == nil {
		t.Fatal("expected B03 alert")
	}
	if found.Direction != Below {
		t.Errorf("B03 Direction = %s, want below", found.Direction)
	}
}

func TestDetectBreakage_B05_TooLow(t *testing.T) {
	bd := BreakageData{
		B03:            0.10,
		B05:            0.10, // below 0.30
		B11SurplusRate: 0.20,
	}

	report := DetectBreakage(bd, BreakageThresholds{})
	found := findAlert(report, "B05")
	if found == nil {
		t.Fatal("expected B05 alert")
	}
}

func TestDetectBreakage_B09_TooHigh(t *testing.T) {
	bd := BreakageData{
		B03:               0.10,
		B05:               0.50,
		B09RoomLevelRatio: 0.90, // above 0.80
		B11SurplusRate:    0.20,
	}

	report := DetectBreakage(bd, BreakageThresholds{})
	found := findAlert(report, "B09")
	if found == nil {
		t.Fatal("expected B09 alert")
	}
	if found.Direction != Above {
		t.Errorf("B09 Direction = %s, want above", found.Direction)
	}
}

func TestDetectBreakage_B10_TooLow(t *testing.T) {
	bd := BreakageData{
		B03:              0.10,
		B05:              0.50,
		B10LayoutEntropy: 0.20, // below 0.50
		B11SurplusRate:   0.20,
	}

	report := DetectBreakage(bd, BreakageThresholds{B10MinEntropy: 0.50})
	found := findAlert(report, "B10")
	if found == nil {
		t.Fatal("expected B10 alert")
	}
}

func TestDetectBreakage_B10_SkippedWhenThresholdZero(t *testing.T) {
	bd := BreakageData{
		B03:              0.10,
		B05:              0.50,
		B10LayoutEntropy: 0.10,
		B11SurplusRate:   0.20,
	}

	report := DetectBreakage(bd, BreakageThresholds{B10MinEntropy: 0})
	if found := findAlert(report, "B10"); found != nil {
		t.Error("B10 should not alert when threshold is 0")
	}
}

func TestDetectBreakage_B11_TooHigh(t *testing.T) {
	bd := BreakageData{
		B03:            0.10,
		B05:            0.50,
		B11SurplusRate: 0.70, // above 0.50
	}

	report := DetectBreakage(bd, BreakageThresholds{})
	found := findAlert(report, "B11")
	if found == nil {
		t.Fatal("expected B11 alert")
	}
	if found.Direction != Above {
		t.Errorf("B11 Direction = %s, want above", found.Direction)
	}
}

func TestDetectBreakage_MultipleAlerts(t *testing.T) {
	bd := BreakageData{
		B01:               5,
		B02:               1,
		FirstWaveRecorded: true,
		B03:               0.01, // below 0.05
		B04ZeroBuildable:  true,
		B05:               0.10, // below 0.30
		B06Stomp:          true,
		B07EarlyWipe:      true,
		B08Perfection:     true,
		B09RoomLevelRatio: 0.95, // above 0.80
		B10LayoutEntropy:  0.10, // below 0.50
		B11SurplusRate:    0.80, // above 0.50
	}
	th := BreakageThresholds{
		B01MinGraceTicks: 20,
		B10MinEntropy:    0.50,
	}

	report := DetectBreakage(bd, th)
	if len(report.Alerts) != 11 {
		t.Errorf("expected 11 alerts, got %d", len(report.Alerts))
		for _, a := range report.Alerts {
			t.Logf("  alert: %s", a.MetricID)
		}
	}
	if len(report.Clean) != 0 {
		t.Errorf("expected 0 clean, got %d: %v", len(report.Clean), report.Clean)
	}
}

// findAlert returns the first alert with the given MetricID, or nil.
func findAlert(report BreakageReport, metricID string) *BreakageAlert {
	for i := range report.Alerts {
		if report.Alerts[i].MetricID == metricID {
			return &report.Alerts[i]
		}
	}
	return nil
}
