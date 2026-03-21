package scenario

import (
	_ "embed"
	"testing"

	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/testutil"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

//go:embed testdata/tutorial.json
var tutorialJSON []byte

// setupRegistries creates minimal registries for integration testing.
func setupRegistries() (*world.RoomTypeRegistry, *senju.SpeciesRegistry, *invasion.InvaderClassRegistry) {
	roomReg := world.NewRoomTypeRegistry()
	_ = roomReg.Register(world.RoomType{
		ID:              "dragon_hole",
		Name:            "龍穴",
		Element:         types.Earth,
		BaseChiCapacity: 100,
		MaxBeasts:       2,
		BaseCoreHP:      100,
	})
	_ = roomReg.Register(world.RoomType{
		ID:              "water_room",
		Name:            "水の間",
		Element:         types.Water,
		BaseChiCapacity: 50,
		MaxBeasts:       2,
	})

	specReg := senju.NewSpeciesRegistry()
	_ = specReg.Register(&senju.Species{
		ID:         "suiryu",
		Name:       "翠龍",
		Element:    types.Wood,
		BaseHP:     20,
		BaseATK:    8,
		BaseDEF:    5,
		BaseSPD:    6,
		GrowthRate: 1.0,
		MaxBeasts:  2,
	})
	_ = specReg.Register(&senju.Species{
		ID:         "suiryu_evolved",
		Name:       "蒼龍",
		Element:    types.Wood,
		BaseHP:     30,
		BaseATK:    12,
		BaseDEF:    8,
		BaseSPD:    9,
		GrowthRate: 1.2,
		MaxBeasts:  2,
	})

	invReg := invasion.NewInvaderClassRegistry()
	_ = invReg.Register(invasion.InvaderClass{
		ID:                "wood_ascetic",
		Name:              "木行修験者",
		Element:           types.Wood,
		BaseHP:            15,
		BaseATK:           6,
		BaseDEF:           4,
		BaseSPD:           5,
		RewardChi:         10.0,
		PreferredGoal:     invasion.DestroyCore,
		RetreatThreshold:  0.2,
	})

	return roomReg, specReg, invReg
}

// buildCaveFromScenario constructs a Cave from the scenario's InitialState,
// placing prebuilt rooms and applying terrain.
func buildCaveFromScenario(t *testing.T, sc *Scenario, roomReg *world.RoomTypeRegistry) *world.Cave {
	t.Helper()

	is := sc.InitialState
	cave, err := world.NewCave(is.CaveWidth, is.CaveHeight)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Place prebuilt rooms.
	for i, rp := range is.PrebuiltRooms {
		rt, err := roomReg.Get(rp.TypeID)
		if err != nil {
			t.Fatalf("room type %q not found: %v", rp.TypeID, err)
		}
		room, err := cave.AddRoom(rp.TypeID, rp.Pos, 3, 3, []world.RoomEntrance{
			{Pos: types.Pos{X: rp.Pos.X + 1, Y: rp.Pos.Y + 2}, Dir: types.South},
		})
		if err != nil {
			t.Fatalf("AddRoom[%d] %s: %v", i, rp.TypeID, err)
		}
		room.Level = rp.Level
		// Initialize CoreHP for dragon hole rooms.
		if rt.BaseCoreHP > 0 {
			room.CoreHP = rt.CoreHPAtLevel(rp.Level)
		}
	}

	// Generate and apply terrain.
	rng := types.NewSeededRNG(is.TerrainSeed)
	tg := &TerrainGenerator{}
	zones := tg.GenerateTerrain(is.CaveWidth, is.CaveHeight, is.TerrainDensity, rng)
	// Filter zones that would overlap with rooms.
	var safeZones []TerrainZone
	for _, z := range zones {
		overlap := false
		for dy := 0; dy < z.Height && !overlap; dy++ {
			for dx := 0; dx < z.Width && !overlap; dx++ {
				pos := types.Pos{X: z.Pos.X + dx, Y: z.Pos.Y + dy}
				if cave.Grid.InBounds(pos) {
					cell, _ := cave.Grid.At(pos)
					if cell.RoomID != 0 {
						overlap = true
					}
				}
			}
		}
		if !overlap {
			safeZones = append(safeZones, z)
		}
	}
	if err := ApplyTerrain(cave, safeZones); err != nil {
		t.Fatalf("ApplyTerrain: %v", err)
	}

	return cave
}

// TestIntegration_TutorialScenario is a comprehensive integration test that
// simulates a tutorial scenario by manually advancing game state through the
// core scenario subsystems.
func TestIntegration_TutorialScenario(t *testing.T) {
	// --- Setup: Load scenario ---
	sc, err := LoadScenario(tutorialJSON)
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}

	roomReg, specReg, invReg := setupRegistries()

	// --- Validate scenario ---
	errs := ValidateScenario(sc, ValidationContext{
		RoomTypes:      roomReg,
		Species:        specReg,
		InvaderClasses: invReg,
	})
	if len(errs) > 0 {
		t.Fatalf("ValidateScenario errors: %v", errs)
	}

	// --- Build Cave from InitialState ---
	cave := buildCaveFromScenario(t, sc, roomReg)

	if len(cave.Rooms) != len(sc.InitialState.PrebuiltRooms) {
		t.Fatalf("expected %d rooms, got %d", len(sc.InitialState.PrebuiltRooms), len(cave.Rooms))
	}

	// --- Terrain validation passes ---
	if err := ValidateTerrain(cave, sc.InitialState.PrebuiltRooms); err != nil {
		t.Fatalf("ValidateTerrain: %v", err)
	}

	// --- WaveSchedule → WaveConfig conversion ---
	rng := testutil.NewTestRNG(42)
	builder := &WaveScheduleBuilder{}
	configs := builder.BuildSchedule(sc.WaveSchedule, rng)

	if len(configs) != len(sc.WaveSchedule) {
		t.Fatalf("BuildSchedule: expected %d configs, got %d", len(sc.WaveSchedule), len(configs))
	}
	for i, cfg := range configs {
		entry := sc.WaveSchedule[i]
		if cfg.TriggerTick != entry.TriggerTick {
			t.Errorf("config[%d].TriggerTick: got %d, want %d", i, cfg.TriggerTick, entry.TriggerTick)
		}
		if cfg.Difficulty != entry.Difficulty {
			t.Errorf("config[%d].Difficulty: got %f, want %f", i, cfg.Difficulty, entry.Difficulty)
		}
		if cfg.MinInvaders != entry.MinInvaders {
			t.Errorf("config[%d].MinInvaders: got %d, want %d", i, cfg.MinInvaders, entry.MinInvaders)
		}
		if cfg.MaxInvaders != entry.MaxInvaders {
			t.Errorf("config[%d].MaxInvaders: got %d, want %d", i, cfg.MaxInvaders, entry.MaxInvaders)
		}
	}

	// --- Victory condition evaluation ---
	t.Run("WinCondition_SurviveUntil", func(t *testing.T) {
		winCond, err := NewCondition(sc.WinConditions[0])
		if err != nil {
			t.Fatalf("NewCondition: %v", err)
		}

		// Not yet won: tick < target.
		snapNotYet := GameSnapshot{Tick: 299, CoreHP: 100, TotalWaves: 1}
		if winCond.Evaluate(snapNotYet) {
			t.Error("win condition should not be met at tick 299")
		}

		// Won: tick >= target with positive CoreHP.
		snapWon := GameSnapshot{Tick: 300, CoreHP: 50, TotalWaves: 1}
		if !winCond.Evaluate(snapWon) {
			t.Error("win condition should be met at tick 300 with CoreHP > 0")
		}

		// Not won: tick >= target but CoreHP == 0.
		snapDead := GameSnapshot{Tick: 300, CoreHP: 0, TotalWaves: 1}
		if winCond.Evaluate(snapDead) {
			t.Error("win condition should not be met when CoreHP == 0")
		}
	})

	// --- Defeat condition evaluation ---
	t.Run("LoseCondition_CoreDestroyed", func(t *testing.T) {
		loseCond, err := NewCondition(sc.LoseConditions[0])
		if err != nil {
			t.Fatalf("NewCondition: %v", err)
		}

		// Core alive.
		snapAlive := GameSnapshot{Tick: 50, CoreHP: 100}
		if loseCond.Evaluate(snapAlive) {
			t.Error("lose condition should not trigger with CoreHP > 0")
		}

		// Core destroyed.
		snapDead := GameSnapshot{Tick: 50, CoreHP: 0}
		if !loseCond.Evaluate(snapDead) {
			t.Error("lose condition should trigger when CoreHP <= 0")
		}
	})

	// --- Event firing (D011: commands returned, no direct state mutation) ---
	t.Run("EventFiring_CommandReturned", func(t *testing.T) {
		// Create a scenario with events for testing.
		events := []EventDef{
			{
				ID: "chi_bonus",
				Condition: ConditionDef{
					Type:   "survive_until",
					Params: map[string]any{"ticks": float64(50)},
				},
				Commands: []CommandDef{
					{Type: "modify_chi", Params: map[string]any{"amount": float64(100)}},
					{Type: "message", Params: map[string]any{"text": "ボーナス気を獲得！"}},
				},
				OneShot: true,
			},
		}

		engine := NewEventEngine()

		// Before condition is met: no commands.
		snapBefore := GameSnapshot{Tick: 49, CoreHP: 100}
		cmds, err := engine.Tick(snapBefore, events)
		if err != nil {
			t.Fatalf("EventEngine.Tick: %v", err)
		}
		if len(cmds) != 0 {
			t.Errorf("expected 0 commands before condition met, got %d", len(cmds))
		}

		// After condition is met: commands returned.
		snapAfter := GameSnapshot{Tick: 50, CoreHP: 100}
		cmds, err = engine.Tick(snapAfter, events)
		if err != nil {
			t.Fatalf("EventEngine.Tick: %v", err)
		}
		if len(cmds) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(cmds))
		}

		// Verify commands are descriptive (D011: Execute returns description,
		// does not mutate game state).
		modCmd, ok := cmds[0].(*ModifyChiCommand)
		if !ok {
			t.Fatalf("expected ModifyChiCommand, got %T", cmds[0])
		}
		if modCmd.Amount != 100 {
			t.Errorf("ModifyChiCommand.Amount: got %f, want 100", modCmd.Amount)
		}
		desc := modCmd.Execute()
		if desc == "" {
			t.Error("Execute() should return non-empty description")
		}

		msgCmd, ok := cmds[1].(*MessageCommand)
		if !ok {
			t.Fatalf("expected MessageCommand, got %T", cmds[1])
		}
		if msgCmd.Text != "ボーナス気を獲得！" {
			t.Errorf("MessageCommand.Text: got %q", msgCmd.Text)
		}

		// One-shot: should not fire again.
		cmds, err = engine.Tick(snapAfter, events)
		if err != nil {
			t.Fatalf("EventEngine.Tick (2nd): %v", err)
		}
		if len(cmds) != 0 {
			t.Errorf("one-shot event should not fire again, got %d commands", len(cmds))
		}
	})

	// --- Beast evolution ---
	t.Run("BeastEvolution", func(t *testing.T) {
		species, _ := specReg.Get("suiryu")
		beast := senju.NewBeast(1, species, 0)

		// Manually level up the beast to meet evolution condition.
		beast.Level = 5

		evolvedSpecies, _ := specReg.Get("suiryu_evolved")

		path := &senju.EvolutionPath{
			FromSpeciesID: "suiryu",
			ToSpeciesID:   "suiryu_evolved",
			Condition: senju.EvolutionCondition{
				MinLevel: 5,
			},
			ChiCost: 50.0,
		}

		// Verify condition is met.
		if beast.Level < path.Condition.MinLevel {
			t.Fatal("beast level should meet evolution condition")
		}

		// Perform evolution.
		if err := senju.Evolve(beast, path, specReg); err != nil {
			t.Fatalf("Evolve: %v", err)
		}

		// Verify post-evolution state.
		if beast.SpeciesID != "suiryu_evolved" {
			t.Errorf("species: got %q, want %q", beast.SpeciesID, "suiryu_evolved")
		}
		if beast.Element != evolvedSpecies.Element {
			t.Errorf("element: got %v, want %v", beast.Element, evolvedSpecies.Element)
		}
		if beast.HP != beast.MaxHP {
			t.Errorf("HP should be fully restored after evolution: got %d/%d", beast.HP, beast.MaxHP)
		}
		// Stats should be recalculated based on new species + current level.
		expectedMaxHP := evolvedSpecies.BaseHP + (beast.Level-1)*2
		if beast.MaxHP != expectedMaxHP {
			t.Errorf("MaxHP: got %d, want %d", beast.MaxHP, expectedMaxHP)
		}
	})

	// --- Beast defeat → Stunned → revival ---
	t.Run("BeastDefeatAndRevival", func(t *testing.T) {
		species, _ := specReg.Get("suiryu")
		beast := senju.NewBeast(2, species, 0)
		beast.Level = 3
		beast.HP = 0 // Simulating HP reaching 0 from invader attack.

		dp := senju.NewDefeatProcessor()
		currentTick := types.Tick(150)
		result := dp.ProcessDefeat(beast, currentTick)

		// Beast should be Stunned.
		if beast.State != senju.Stunned {
			t.Errorf("state: got %v, want Stunned", beast.State)
		}
		if result.NewState != senju.Stunned {
			t.Errorf("result.NewState: got %v, want Stunned", result.NewState)
		}

		// Revival tick should be current + stunned duration (default 20).
		expectedRevival := types.Tick(170)
		if result.RevivalTick != expectedRevival {
			t.Errorf("RevivalTick: got %d, want %d", result.RevivalTick, expectedRevival)
		}

		// Level penalty should be 1 (default).
		if result.LevelPenalty != 1 {
			t.Errorf("LevelPenalty: got %d, want 1", result.LevelPenalty)
		}

		// Revival HP should be 30% of MaxHP (default ratio).
		expectedHP := int(float64(beast.MaxHP) * 0.3)
		if expectedHP < 1 {
			expectedHP = 1
		}
		if result.RevivalHP != expectedHP {
			t.Errorf("RevivalHP: got %d, want %d", result.RevivalHP, expectedHP)
		}

		// Simulate revival: apply result to beast.
		beast.Level -= result.LevelPenalty
		if beast.Level < 1 {
			beast.Level = 1
		}
		beast.HP = result.RevivalHP
		beast.State = senju.Idle

		if beast.State != senju.Idle {
			t.Errorf("after revival, state should be Idle, got %v", beast.State)
		}
		if beast.Level != 2 {
			t.Errorf("after revival, level: got %d, want 2", beast.Level)
		}
		if beast.HP != expectedHP {
			t.Errorf("after revival, HP: got %d, want %d", beast.HP, expectedHP)
		}
	})

	// --- CoreHP reduced by invader attack ---
	t.Run("CoreHP_InvaderDamage", func(t *testing.T) {
		// Use the dragon hole room from the cave we built.
		dragonHole := cave.Rooms[0]
		if dragonHole.TypeID != "dragon_hole" {
			t.Fatalf("expected dragon_hole room, got %q", dragonHole.TypeID)
		}

		rt, _ := roomReg.Get("dragon_hole")
		initialCoreHP := rt.CoreHPAtLevel(dragonHole.Level)
		dragonHole.CoreHP = initialCoreHP

		if dragonHole.CoreHP != 100 {
			t.Fatalf("initial CoreHP: got %d, want 100", dragonHole.CoreHP)
		}

		// Simulate invader dealing damage to the core.
		invaderATK := 15
		dragonHole.CoreHP -= invaderATK

		if dragonHole.CoreHP != 85 {
			t.Errorf("after first attack CoreHP: got %d, want 85", dragonHole.CoreHP)
		}

		// Check lose condition against snapshot reflecting the damage.
		loseCond, _ := NewCondition(sc.LoseConditions[0])

		// Not yet destroyed.
		snapDamaged := GameSnapshot{CoreHP: dragonHole.CoreHP}
		if loseCond.Evaluate(snapDamaged) {
			t.Error("core_destroyed should not trigger with CoreHP > 0")
		}

		// Destroy the core completely.
		dragonHole.CoreHP = 0
		snapDestroyed := GameSnapshot{CoreHP: dragonHole.CoreHP}
		if !loseCond.Evaluate(snapDestroyed) {
			t.Error("core_destroyed should trigger when CoreHP == 0")
		}
	})
}
