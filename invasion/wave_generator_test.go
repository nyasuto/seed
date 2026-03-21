package invasion

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/testutil"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// newTestCave creates a minimal cave for testing.
func newTestCave(t *testing.T, roomTypes ...string) *world.Cave {
	t.Helper()
	cave, err := world.NewCave(30, 30)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	for i, typeID := range roomTypes {
		_, err := cave.AddRoom(typeID, types.Pos{X: i * 5, Y: 0}, 3, 3, nil)
		if err != nil {
			t.Fatalf("AddRoom(%s): %v", typeID, err)
		}
	}
	return cave
}

func newTestClassRegistry() *InvaderClassRegistry {
	reg := NewInvaderClassRegistry()
	_ = reg.Register(InvaderClass{
		ID: "warrior", Name: "Warrior", Element: types.Wood,
		BaseHP: 100, BaseATK: 25, BaseDEF: 20, BaseSPD: 20,
		RewardChi: 15, PreferredGoal: DestroyCore, RetreatThreshold: 0.3,
	})
	_ = reg.Register(InvaderClass{
		ID: "thief", Name: "Thief", Element: types.Metal,
		BaseHP: 70, BaseATK: 20, BaseDEF: 15, BaseSPD: 40,
		RewardChi: 12, PreferredGoal: StealTreasure, RetreatThreshold: 0.5,
	})
	_ = reg.Register(InvaderClass{
		ID: "hunter", Name: "Hunter", Element: types.Fire,
		BaseHP: 80, BaseATK: 40, BaseDEF: 15, BaseSPD: 25,
		RewardChi: 18, PreferredGoal: HuntBeasts, RetreatThreshold: 0.15,
	})
	return reg
}

func TestWaveGenerator_GenerateWave_InvaderCount(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole", "normal")

	// With fixed RNG, count = min + Intn(max-min+1) = 2 + 0 = 2
	rng := &testutil.FixedRNG{IntValue: 0}
	wg := NewWaveGenerator(reg, rng)

	config := WaveConfig{
		TriggerTick: 50,
		Difficulty:  1.0,
		MinInvaders: 2,
		MaxInvaders: 4,
	}

	wave, err := wg.GenerateWave(config, cave, 50)
	if err != nil {
		t.Fatalf("GenerateWave: %v", err)
	}

	if len(wave.Invaders) != 2 {
		t.Errorf("got %d invaders, want 2", len(wave.Invaders))
	}
	if wave.State != Pending {
		t.Errorf("got state %v, want Pending", wave.State)
	}
	if wave.Difficulty != 1.0 {
		t.Errorf("got difficulty %f, want 1.0", wave.Difficulty)
	}
	if wave.TriggerTick != 50 {
		t.Errorf("got trigger tick %d, want 50", wave.TriggerTick)
	}
}

func TestWaveGenerator_GenerateWave_MaxCount(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole")

	// IntValue=100 → Intn(3) = 100%3 = 1, so count = 2+1 = 3
	rng := &testutil.FixedRNG{IntValue: 100}
	wg := NewWaveGenerator(reg, rng)

	config := WaveConfig{
		TriggerTick: 50,
		Difficulty:  1.0,
		MinInvaders: 2,
		MaxInvaders: 4,
	}

	wave, err := wg.GenerateWave(config, cave, 50)
	if err != nil {
		t.Fatalf("GenerateWave: %v", err)
	}

	if len(wave.Invaders) != 3 {
		t.Errorf("got %d invaders, want 3", len(wave.Invaders))
	}
}

func TestWaveGenerator_GenerateWave_GoalAssignment_WithCore(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole", "storage")

	rng := testutil.NewTestRNG(42)
	wg := NewWaveGenerator(reg, rng)

	config := WaveConfig{
		TriggerTick: 100,
		Difficulty:  1.5,
		MinInvaders: 10,
		MaxInvaders: 10,
	}

	wave, err := wg.GenerateWave(config, cave, 100)
	if err != nil {
		t.Fatalf("GenerateWave: %v", err)
	}

	goalCounts := make(map[GoalType]int)
	for _, inv := range wave.Invaders {
		goalCounts[inv.Goal.Type()]++
	}

	// With both dragon_hole and storage, we expect a mix of goals:
	// DestroyCore (warriors), StealTreasure (thieves), HuntBeasts (hunters)
	if goalCounts[DestroyCore] == 0 {
		t.Error("expected at least one DestroyCore invader when dragon_hole exists")
	}
}

func TestWaveGenerator_GenerateWave_GoalAssignment_NoCoreRoom(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "normal", "storage")

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
		if inv.Goal.Type() == DestroyCore {
			t.Error("should not assign DestroyCore when no dragon_hole exists")
			break
		}
	}
}

func TestWaveGenerator_GenerateWave_GoalAssignment_NoStorage(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole")

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
		if inv.Goal.Type() == StealTreasure {
			t.Error("should not assign StealTreasure when no storage room exists")
			break
		}
	}
}

func TestWaveGenerator_GenerateWave_DifficultyScaling(t *testing.T) {
	reg := NewInvaderClassRegistry()
	_ = reg.Register(InvaderClass{
		ID: "test_class", Name: "Test", Element: types.Wood,
		BaseHP: 100, BaseATK: 20, BaseDEF: 10, BaseSPD: 15,
		PreferredGoal: DestroyCore, RetreatThreshold: 0.3,
	})
	cave := newTestCave(t, "dragon_hole")

	tests := []struct {
		name       string
		difficulty float64
		wantLevel  int
	}{
		{"difficulty_1.0", 1.0, 1},
		{"difficulty_1.5", 1.5, 2},
		{"difficulty_2.0", 2.0, 2},
		{"difficulty_2.8", 2.8, 3},
		{"difficulty_0.3", 0.3, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := &testutil.FixedRNG{IntValue: 0}
			wg := NewWaveGenerator(reg, rng)
			config := WaveConfig{
				TriggerTick: 50,
				Difficulty:  tt.difficulty,
				MinInvaders: 1,
				MaxInvaders: 1,
			}
			wave, err := wg.GenerateWave(config, cave, 50)
			if err != nil {
				t.Fatalf("GenerateWave: %v", err)
			}
			if wave.Invaders[0].Level != tt.wantLevel {
				t.Errorf("got level %d, want %d", wave.Invaders[0].Level, tt.wantLevel)
			}
		})
	}
}

func TestWaveGenerator_GenerateWave_IncrementingIDs(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole")
	rng := testutil.NewTestRNG(1)
	wg := NewWaveGenerator(reg, rng)

	config := WaveConfig{
		TriggerTick: 50,
		Difficulty:  1.0,
		MinInvaders: 2,
		MaxInvaders: 2,
	}

	wave1, err := wg.GenerateWave(config, cave, 50)
	if err != nil {
		t.Fatalf("GenerateWave 1: %v", err)
	}
	wave2, err := wg.GenerateWave(config, cave, 100)
	if err != nil {
		t.Fatalf("GenerateWave 2: %v", err)
	}

	if wave1.ID != 1 {
		t.Errorf("first wave ID = %d, want 1", wave1.ID)
	}
	if wave2.ID != 2 {
		t.Errorf("second wave ID = %d, want 2", wave2.ID)
	}

	// Invader IDs should be globally unique.
	ids := make(map[int]bool)
	for _, inv := range wave1.Invaders {
		ids[inv.ID] = true
	}
	for _, inv := range wave2.Invaders {
		if ids[inv.ID] {
			t.Errorf("duplicate invader ID %d across waves", inv.ID)
		}
	}
}

func TestWaveGenerator_GenerateWave_InvalidConfig(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "dragon_hole")
	rng := &testutil.FixedRNG{IntValue: 0}
	wg := NewWaveGenerator(reg, rng)

	tests := []struct {
		name   string
		config WaveConfig
	}{
		{"zero_min", WaveConfig{TriggerTick: 50, Difficulty: 1.0, MinInvaders: 0, MaxInvaders: 3}},
		{"negative_max", WaveConfig{TriggerTick: 50, Difficulty: 1.0, MinInvaders: 1, MaxInvaders: -1}},
		{"min_exceeds_max", WaveConfig{TriggerTick: 50, Difficulty: 1.0, MinInvaders: 5, MaxInvaders: 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := wg.GenerateWave(tt.config, cave, 50)
			if err == nil {
				t.Error("expected error for invalid config")
			}
		})
	}
}

func TestWaveGenerator_GenerateWave_EmptyRegistry(t *testing.T) {
	reg := NewInvaderClassRegistry()
	cave := newTestCave(t, "dragon_hole")
	rng := &testutil.FixedRNG{IntValue: 0}
	wg := NewWaveGenerator(reg, rng)

	config := WaveConfig{
		TriggerTick: 50,
		Difficulty:  1.0,
		MinInvaders: 1,
		MaxInvaders: 3,
	}

	_, err := wg.GenerateWave(config, cave, 50)
	if err == nil {
		t.Error("expected error for empty class registry")
	}
}

func TestWaveGenerator_GenerateWave_EntryRoom(t *testing.T) {
	reg := newTestClassRegistry()
	cave := newTestCave(t, "entrance", "dragon_hole")
	rng := &testutil.FixedRNG{IntValue: 0}
	wg := NewWaveGenerator(reg, rng)

	config := WaveConfig{
		TriggerTick: 50,
		Difficulty:  1.0,
		MinInvaders: 1,
		MaxInvaders: 1,
	}

	wave, err := wg.GenerateWave(config, cave, 50)
	if err != nil {
		t.Fatalf("GenerateWave: %v", err)
	}

	// Entry room should be the first room (ID from cave).
	expectedRoomID := cave.Rooms[0].ID
	if wave.Invaders[0].CurrentRoomID != expectedRoomID {
		t.Errorf("got entry room %d, want %d", wave.Invaders[0].CurrentRoomID, expectedRoomID)
	}
}
