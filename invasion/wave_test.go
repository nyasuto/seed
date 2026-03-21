package invasion

import (
	_ "embed"
	"testing"

	"github.com/ponpoko/chaosseed-core/testutil"
	"github.com/ponpoko/chaosseed-core/types"
)

//go:embed wave_schedule_data.json
var waveScheduleDataJSON []byte

// --- WaveSchedule tests ---

func TestLoadWaveSchedule_EmbeddedJSON(t *testing.T) {
	schedule, err := LoadWaveSchedule(waveScheduleDataJSON)
	if err != nil {
		t.Fatalf("LoadWaveSchedule: %v", err)
	}

	if len(schedule.Waves) != 3 {
		t.Fatalf("got %d waves, want 3", len(schedule.Waves))
	}

	tests := []struct {
		index       int
		triggerTick types.Tick
		difficulty  float64
		minInvaders int
		maxInvaders int
	}{
		{0, 50, 1.0, 2, 3},
		{1, 150, 1.5, 3, 5},
		{2, 300, 2.0, 5, 7},
	}

	for _, tt := range tests {
		w := schedule.Waves[tt.index]
		if w.TriggerTick != tt.triggerTick {
			t.Errorf("wave %d: trigger_tick = %d, want %d", tt.index, w.TriggerTick, tt.triggerTick)
		}
		if w.Difficulty != tt.difficulty {
			t.Errorf("wave %d: difficulty = %f, want %f", tt.index, w.Difficulty, tt.difficulty)
		}
		if w.MinInvaders != tt.minInvaders {
			t.Errorf("wave %d: min_invaders = %d, want %d", tt.index, w.MinInvaders, tt.minInvaders)
		}
		if w.MaxInvaders != tt.maxInvaders {
			t.Errorf("wave %d: max_invaders = %d, want %d", tt.index, w.MaxInvaders, tt.maxInvaders)
		}
	}
}

func TestLoadWaveSchedule_EmptyWaves(t *testing.T) {
	data := []byte(`{"waves": []}`)
	_, err := LoadWaveSchedule(data)
	if err == nil {
		t.Error("expected error for empty waves")
	}
}

func TestLoadWaveSchedule_InvalidJSON(t *testing.T) {
	_, err := LoadWaveSchedule([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadWaveSchedule_InvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			"zero_min_invaders",
			`{"waves": [{"trigger_tick": 50, "difficulty": 1.0, "min_invaders": 0, "max_invaders": 3}]}`,
		},
		{
			"negative_max_invaders",
			`{"waves": [{"trigger_tick": 50, "difficulty": 1.0, "min_invaders": 1, "max_invaders": -1}]}`,
		},
		{
			"min_exceeds_max",
			`{"waves": [{"trigger_tick": 50, "difficulty": 1.0, "min_invaders": 5, "max_invaders": 3}]}`,
		},
		{
			"zero_difficulty",
			`{"waves": [{"trigger_tick": 50, "difficulty": 0, "min_invaders": 1, "max_invaders": 3}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadWaveSchedule([]byte(tt.data))
			if err == nil {
				t.Error("expected error for invalid config")
			}
		})
	}
}

// --- WaveGenerator with schedule integration ---

func TestWaveGenerator_GenerateFromSchedule(t *testing.T) {
	schedule, err := LoadWaveSchedule(waveScheduleDataJSON)
	if err != nil {
		t.Fatalf("LoadWaveSchedule: %v", err)
	}

	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole", "storage")
	rng := testutil.NewTestRNG(42)
	wg := NewWaveGenerator(reg, rng)

	for i, config := range schedule.Waves {
		wave, err := wg.GenerateWave(config, cave, config.TriggerTick)
		if err != nil {
			t.Fatalf("wave %d: GenerateWave: %v", i, err)
		}

		if len(wave.Invaders) < config.MinInvaders || len(wave.Invaders) > config.MaxInvaders {
			t.Errorf("wave %d: got %d invaders, want between %d and %d",
				i, len(wave.Invaders), config.MinInvaders, config.MaxInvaders)
		}
		if wave.TriggerTick != config.TriggerTick {
			t.Errorf("wave %d: trigger tick = %d, want %d", i, wave.TriggerTick, config.TriggerTick)
		}
		if wave.Difficulty != config.Difficulty {
			t.Errorf("wave %d: difficulty = %f, want %f", i, wave.Difficulty, config.Difficulty)
		}
	}
}

// --- InvasionWave lifecycle tests ---

func TestInvasionWave_AliveCount(t *testing.T) {
	inv1 := &Invader{ID: 1, State: Advancing}
	inv2 := &Invader{ID: 2, State: Defeated}
	inv3 := &Invader{ID: 3, State: Retreating}

	wave := &InvasionWave{
		ID:       1,
		Invaders: []*Invader{inv1, inv2, inv3},
		State:    Active,
	}

	if got := wave.AliveCount(); got != 2 {
		t.Errorf("AliveCount() = %d, want 2", got)
	}
}

func TestInvasionWave_DefeatedCount(t *testing.T) {
	inv1 := &Invader{ID: 1, State: Defeated}
	inv2 := &Invader{ID: 2, State: Defeated}
	inv3 := &Invader{ID: 3, State: Advancing}

	wave := &InvasionWave{
		ID:       1,
		Invaders: []*Invader{inv1, inv2, inv3},
		State:    Active,
	}

	if got := wave.DefeatedCount(); got != 2 {
		t.Errorf("DefeatedCount() = %d, want 2", got)
	}
}

func TestInvasionWave_IsActive(t *testing.T) {
	wave := &InvasionWave{State: Pending}
	if wave.IsActive() {
		t.Error("Pending wave should not be active")
	}

	wave.State = Active
	if !wave.IsActive() {
		t.Error("Active wave should be active")
	}

	wave.State = Completed
	if wave.IsActive() {
		t.Error("Completed wave should not be active")
	}
}

func TestInvasionWave_IsCompleted(t *testing.T) {
	tests := []struct {
		state WaveState
		want  bool
	}{
		{Pending, false},
		{Active, false},
		{Completed, true},
		{Failed, true},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			wave := &InvasionWave{State: tt.state}
			if got := wave.IsCompleted(); got != tt.want {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Goal assignment tests ---

func TestWaveGenerator_GoalAssignment_WithCoreAndStorage(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole", "storage")
	rng := testutil.NewTestRNG(42)
	wg := NewWaveGenerator(reg, rng)

	config := WaveConfig{
		TriggerTick: 100,
		Difficulty:  1.5,
		MinInvaders: 20,
		MaxInvaders: 20,
	}

	wave, err := wg.GenerateWave(config, cave, 100)
	if err != nil {
		t.Fatalf("GenerateWave: %v", err)
	}

	goalCounts := make(map[GoalType]int)
	for _, inv := range wave.Invaders {
		goalCounts[inv.Goal.Type()]++
	}

	if goalCounts[DestroyCore] == 0 {
		t.Error("expected DestroyCore invaders when dragon_hole exists")
	}
	if goalCounts[StealTreasure] == 0 {
		t.Error("expected StealTreasure invaders when storage exists")
	}
}

func TestWaveGenerator_GoalAssignment_NoCoreNoStorage(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "normal")
	rng := testutil.NewTestRNG(42)
	wg := NewWaveGenerator(reg, rng)

	config := WaveConfig{
		TriggerTick: 100,
		Difficulty:  1.0,
		MinInvaders: 10,
		MaxInvaders: 10,
	}

	wave, err := wg.GenerateWave(config, cave, 100)
	if err != nil {
		t.Fatalf("GenerateWave: %v", err)
	}

	for _, inv := range wave.Invaders {
		gt := inv.Goal.Type()
		if gt == DestroyCore {
			t.Error("should not assign DestroyCore without dragon_hole")
		}
		if gt == StealTreasure {
			t.Error("should not assign StealTreasure without storage")
		}
	}
}

// --- Difficulty scaling test ---

func TestWaveGenerator_DifficultyScaling_AcrossWaves(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole")

	schedule, err := LoadWaveSchedule(waveScheduleDataJSON)
	if err != nil {
		t.Fatalf("LoadWaveSchedule: %v", err)
	}

	rng := testutil.NewTestRNG(1)
	wg := NewWaveGenerator(reg, rng)

	var prevMaxHP int
	for i, config := range schedule.Waves {
		wave, err := wg.GenerateWave(config, cave, config.TriggerTick)
		if err != nil {
			t.Fatalf("wave %d: %v", i, err)
		}

		// Find max HP in wave.
		maxHP := 0
		for _, inv := range wave.Invaders {
			if inv.HP > maxHP {
				maxHP = inv.HP
			}
		}

		if i > 0 && maxHP < prevMaxHP {
			t.Errorf("wave %d: max HP (%d) should not decrease from wave %d (%d) with higher difficulty",
				i, maxHP, i-1, prevMaxHP)
		}
		prevMaxHP = maxHP
	}
}
