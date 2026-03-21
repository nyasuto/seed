package senju

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

func setupGrowthTest(t *testing.T) (*GrowthEngine, *SpeciesRegistry) {
	t.Helper()
	reg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("failed to load species: %v", err)
	}
	params := DefaultGrowthParams()
	engine := NewGrowthEngine(params, reg)
	return engine, reg
}

func makeBeastInRoom(t *testing.T, reg *SpeciesRegistry, speciesID string, beastID, roomID int) *Beast {
	t.Helper()
	sp, err := reg.Get(speciesID)
	if err != nil {
		t.Fatalf("species not found: %v", err)
	}
	b := NewBeast(beastID, sp, 0)
	b.RoomID = roomID
	return b
}

func TestGrowthEngine_BasicEXPGain(t *testing.T) {
	engine, reg := setupGrowthTest(t)
	beast := makeBeastInRoom(t, reg, "suiryu", 1, 10)

	roomChi := map[int]*fengshui.RoomChi{
		10: {RoomID: 10, Current: 100, Capacity: 100, Element: types.Wood},
	}
	rooms := map[int]*world.Room{
		10: {ID: 10},
	}

	events := engine.Tick([]*Beast{beast}, roomChi, rooms)

	var expEvent *GrowthEvent
	for i := range events {
		if events[i].Type == EXPGained {
			expEvent = &events[i]
			break
		}
	}
	if expEvent == nil {
		t.Fatal("expected EXPGained event")
	}
	if beast.EXP <= 0 {
		t.Errorf("expected EXP > 0, got %d", beast.EXP)
	}
}

func TestGrowthEngine_ChiStarved(t *testing.T) {
	engine, reg := setupGrowthTest(t)
	beast := makeBeastInRoom(t, reg, "suiryu", 1, 10)

	// Not enough chi
	roomChi := map[int]*fengshui.RoomChi{
		10: {RoomID: 10, Current: 0.5, Capacity: 100, Element: types.Wood},
	}
	rooms := map[int]*world.Room{
		10: {ID: 10},
	}

	events := engine.Tick([]*Beast{beast}, roomChi, rooms)

	if len(events) != 1 || events[0].Type != ChiStarved {
		t.Errorf("expected ChiStarved event, got %v", events)
	}
	if beast.EXP != 0 {
		t.Errorf("expected EXP = 0 when starved, got %d", beast.EXP)
	}
}

func TestGrowthEngine_ChiStarved_NoRoomChi(t *testing.T) {
	engine, reg := setupGrowthTest(t)
	beast := makeBeastInRoom(t, reg, "suiryu", 1, 10)

	// Room not in chi map
	roomChi := map[int]*fengshui.RoomChi{}
	rooms := map[int]*world.Room{
		10: {ID: 10},
	}

	events := engine.Tick([]*Beast{beast}, roomChi, rooms)

	if len(events) != 1 || events[0].Type != ChiStarved {
		t.Errorf("expected ChiStarved event, got %v", events)
	}
}

func TestGrowthEngine_LevelUp(t *testing.T) {
	engine, reg := setupGrowthTest(t)
	beast := makeBeastInRoom(t, reg, "suiryu", 1, 10)

	// Set EXP just below threshold: LevelUpBase + LevelUpPerLevel * 1 = 100 + 50 = 150
	// suiryu GrowthRate=1.0, Wood in Wood room → SameElement affinity=1.1
	// EXP per tick = 10 * 1.1 * 1.0 = 11
	beast.EXP = 140 // Need 150 total to level up, one tick gives 11

	roomChi := map[int]*fengshui.RoomChi{
		10: {RoomID: 10, Current: 100, Capacity: 100, Element: types.Wood},
	}
	rooms := map[int]*world.Room{
		10: {ID: 10},
	}

	events := engine.Tick([]*Beast{beast}, roomChi, rooms)

	hasLevelUp := false
	for _, e := range events {
		if e.Type == LevelUp {
			hasLevelUp = true
			if e.OldLevel != 1 || e.NewLevel != 2 {
				t.Errorf("expected level 1→2, got %d→%d", e.OldLevel, e.NewLevel)
			}
		}
	}
	if !hasLevelUp {
		t.Fatal("expected LevelUp event")
	}
	if beast.Level != 2 {
		t.Errorf("expected level 2, got %d", beast.Level)
	}
}

func TestGrowthEngine_AffinityAffectsGrowth(t *testing.T) {
	engine, reg := setupGrowthTest(t)

	// Same species in different element rooms
	beastSame := makeBeastInRoom(t, reg, "suiryu", 1, 10) // Wood beast
	beastGen := makeBeastInRoom(t, reg, "suiryu", 2, 20)
	beastOver := makeBeastInRoom(t, reg, "suiryu", 3, 30)

	roomChi := map[int]*fengshui.RoomChi{
		10: {RoomID: 10, Current: 100, Capacity: 100, Element: types.Wood},  // Same: 1.1
		20: {RoomID: 20, Current: 100, Capacity: 100, Element: types.Water}, // Water generates Wood: 1.3
		30: {RoomID: 30, Current: 100, Capacity: 100, Element: types.Metal}, // Metal overcomes Wood: 0.7
	}
	rooms := map[int]*world.Room{
		10: {ID: 10}, 20: {ID: 20}, 30: {ID: 30},
	}

	engine.Tick([]*Beast{beastSame, beastGen, beastOver}, roomChi, rooms)

	// Water generates Wood → highest EXP
	// Same element → middle EXP
	// Metal overcomes Wood → lowest EXP
	if beastGen.EXP <= beastSame.EXP {
		t.Errorf("generates room should give more EXP: gen=%d, same=%d", beastGen.EXP, beastSame.EXP)
	}
	if beastSame.EXP <= beastOver.EXP {
		t.Errorf("same element room should give more EXP than overcomes: same=%d, over=%d", beastSame.EXP, beastOver.EXP)
	}
}

func TestGrowthEngine_MaxLevelClamp(t *testing.T) {
	params := &GrowthParams{
		BaseEXPPerTick:        1000,
		LevelUpBase:           10,
		LevelUpPerLevel:       0,
		ChiConsumptionPerTick: 0.1,
		MaxLevel:              3,
	}
	reg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("failed to load species: %v", err)
	}
	engine := NewGrowthEngine(params, reg)
	beast := makeBeastInRoom(t, reg, "suiryu", 1, 10)

	roomChi := map[int]*fengshui.RoomChi{
		10: {RoomID: 10, Current: 1000, Capacity: 1000, Element: types.Wood},
	}
	rooms := map[int]*world.Room{
		10: {ID: 10},
	}

	// Run enough ticks to exceed max level
	for range 10 {
		engine.Tick([]*Beast{beast}, roomChi, rooms)
	}

	if beast.Level > params.MaxLevel {
		t.Errorf("level %d exceeded max %d", beast.Level, params.MaxLevel)
	}
	if beast.Level != params.MaxLevel {
		t.Errorf("expected level %d, got %d", params.MaxLevel, beast.Level)
	}
}

func TestGrowthEngine_ChiConsumption(t *testing.T) {
	engine, reg := setupGrowthTest(t)
	beast := makeBeastInRoom(t, reg, "suiryu", 1, 10)

	initialChi := 50.0
	roomChi := map[int]*fengshui.RoomChi{
		10: {RoomID: 10, Current: initialChi, Capacity: 100, Element: types.Wood},
	}
	rooms := map[int]*world.Room{
		10: {ID: 10},
	}

	engine.Tick([]*Beast{beast}, roomChi, rooms)

	expectedChi := initialChi - engine.params.ChiConsumptionPerTick
	if roomChi[10].Current != expectedChi {
		t.Errorf("expected chi %.1f after consumption, got %.1f", expectedChi, roomChi[10].Current)
	}
}

func TestGrowthEngine_GrowthEventGeneration(t *testing.T) {
	engine, reg := setupGrowthTest(t)
	beast := makeBeastInRoom(t, reg, "suiryu", 1, 10)

	roomChi := map[int]*fengshui.RoomChi{
		10: {RoomID: 10, Current: 100, Capacity: 100, Element: types.Wood},
	}
	rooms := map[int]*world.Room{
		10: {ID: 10},
	}

	events := engine.Tick([]*Beast{beast}, roomChi, rooms)

	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}

	// First event should be EXPGained
	if events[0].Type != EXPGained {
		t.Errorf("expected EXPGained event, got %v", events[0].Type)
	}
	if events[0].BeastID != beast.ID {
		t.Errorf("expected beast ID %d, got %d", beast.ID, events[0].BeastID)
	}
	if events[0].EXPGained <= 0 {
		t.Errorf("expected positive EXP in event, got %d", events[0].EXPGained)
	}
}

func TestGrowthEngine_UnassignedBeastSkipped(t *testing.T) {
	engine, reg := setupGrowthTest(t)
	sp, _ := reg.Get("suiryu")
	beast := NewBeast(1, sp, 0) // RoomID = 0 (unassigned)

	roomChi := map[int]*fengshui.RoomChi{}
	rooms := map[int]*world.Room{}

	events := engine.Tick([]*Beast{beast}, roomChi, rooms)

	if len(events) != 0 {
		t.Errorf("expected no events for unassigned beast, got %d", len(events))
	}
}

func TestLoadGrowthParams(t *testing.T) {
	data := []byte(`{
		"base_exp_per_tick": 20,
		"level_up_base": 200,
		"level_up_per_level": 100,
		"chi_consumption_per_tick": 5.0,
		"max_level": 99
	}`)

	params, err := LoadGrowthParams(data)
	if err != nil {
		t.Fatalf("failed to load params: %v", err)
	}
	if params.BaseEXPPerTick != 20 {
		t.Errorf("BaseEXPPerTick = %d, want 20", params.BaseEXPPerTick)
	}
	if params.MaxLevel != 99 {
		t.Errorf("MaxLevel = %d, want 99", params.MaxLevel)
	}
	if params.ChiConsumptionPerTick != 5.0 {
		t.Errorf("ChiConsumptionPerTick = %f, want 5.0", params.ChiConsumptionPerTick)
	}
}

func TestLoadGrowthParams_EmbeddedDefaults(t *testing.T) {
	params, err := LoadDefaultGrowthParams()
	if err != nil {
		t.Fatalf("failed to load embedded params: %v", err)
	}
	defaults := DefaultGrowthParams()
	if params.BaseEXPPerTick != defaults.BaseEXPPerTick {
		t.Errorf("BaseEXPPerTick = %d, want %d", params.BaseEXPPerTick, defaults.BaseEXPPerTick)
	}
	if params.MaxLevel != defaults.MaxLevel {
		t.Errorf("MaxLevel = %d, want %d", params.MaxLevel, defaults.MaxLevel)
	}
}

func TestGrowthParams_LevelUpThreshold(t *testing.T) {
	params := DefaultGrowthParams()

	tests := []struct {
		level    int
		expected int
	}{
		{1, 150},  // 100 + 50*1
		{2, 200},  // 100 + 50*2
		{10, 600}, // 100 + 50*10
	}

	for _, tt := range tests {
		got := params.LevelUpThreshold(tt.level)
		if got != tt.expected {
			t.Errorf("LevelUpThreshold(%d) = %d, want %d", tt.level, got, tt.expected)
		}
	}
}
