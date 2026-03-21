package fengshui

import (
	"math"
	"os"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// testRegistry returns a RoomTypeRegistry with standard room types for testing.
func testRegistry() *world.RoomTypeRegistry {
	reg := world.NewRoomTypeRegistry()
	// Use various elements to test interactions.
	_ = reg.Register(world.RoomType{ID: "wood_room", Name: "Wood Room", Element: types.Wood, BaseChiCapacity: 100})
	_ = reg.Register(world.RoomType{ID: "fire_room", Name: "Fire Room", Element: types.Fire, BaseChiCapacity: 100})
	_ = reg.Register(world.RoomType{ID: "earth_room", Name: "Earth Room", Element: types.Earth, BaseChiCapacity: 100})
	_ = reg.Register(world.RoomType{ID: "metal_room", Name: "Metal Room", Element: types.Metal, BaseChiCapacity: 100})
	_ = reg.Register(world.RoomType{ID: "water_room", Name: "Water Room", Element: types.Water, BaseChiCapacity: 100})
	return reg
}

// buildTwoRoomCave creates a 12x6 cave with two connected rooms and returns
// the cave plus a source position on the corridor.
//
// Layout:
//
//	Room 1 at (1,1) 3x3, entrance (3,2) East
//	Room 2 at (7,1) 3x3, entrance (7,2) West
//	Corridor between them
func buildTwoRoomCave(t *testing.T, typeID1, typeID2 string) (*world.Cave, types.Pos) {
	t.Helper()
	cave, err := world.NewCave(12, 6)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	_, err = cave.AddRoom(typeID1, types.Pos{X: 1, Y: 1}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}
	_, err = cave.AddRoom(typeID2, types.Pos{X: 7, Y: 1}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 7, Y: 2}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}
	_, err = cave.ConnectRooms(1, 2)
	if err != nil {
		t.Fatalf("ConnectRooms: %v", err)
	}
	// Source on the corridor midpoint.
	source := types.Pos{X: 5, Y: 2}
	return cave, source
}

func TestChiFlowEngine_DragonVeinSupply(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, params)

	// Before tick, all rooms have 0 chi.
	for _, rc := range engine.RoomChi {
		if rc.Current != 0 {
			t.Fatalf("initial chi = %v, want 0", rc.Current)
		}
	}

	engine.Tick()

	// Both rooms should have received chi from the dragon vein.
	for _, rc := range engine.RoomChi {
		if rc.Current <= 0 {
			t.Errorf("room %d chi = %v, want > 0 after tick", rc.RoomID, rc.Current)
		}
	}
}

func TestChiFlowEngine_GeneratesMultiplier(t *testing.T) {
	// Wood generates Fire: vein=Wood, room=Fire should get 1.3x.
	cave, source := buildTwoRoomCave(t, "fire_room", "fire_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	veinGen, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}
	engineGen := NewChiFlowEngine(cave, []*DragonVein{veinGen}, reg, params)
	engineGen.Tick()
	chiGen := engineGen.RoomChi[1].Current

	// Compare with neutral case: vein=Water, room=Fire (neutral).
	cave2, source2 := buildTwoRoomCave(t, "fire_room", "fire_room")
	veinNeutral, err := BuildDragonVein(cave2, source2, types.Earth, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein neutral: %v", err)
	}
	engineNeutral := NewChiFlowEngine(cave2, []*DragonVein{veinNeutral}, reg, params)
	engineNeutral.Tick()
	chiNeutral := engineNeutral.RoomChi[1].Current

	if chiGen <= chiNeutral {
		t.Errorf("generates chi (%v) should be > neutral chi (%v)", chiGen, chiNeutral)
	}
}

func TestChiFlowEngine_OvercomesMultiplier(t *testing.T) {
	// Wood overcomes Earth: vein=Wood, room=Earth should get 0.6x.
	cave, source := buildTwoRoomCave(t, "earth_room", "earth_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	veinOc, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}
	engineOc := NewChiFlowEngine(cave, []*DragonVein{veinOc}, reg, params)
	engineOc.Tick()
	chiOc := engineOc.RoomChi[1].Current

	// Compare with neutral: vein=Fire, room=Earth (neutral — Fire generates Earth, actually).
	// Use Metal for neutral with Earth (Metal does not generate/overcome Earth).
	// Actually: Fire→Earth is generates. Let's use Water vein for Earth room.
	// Water→Earth: Overcomes(Water,Earth)? Water overcomes Fire, not Earth. So neutral.
	// Wait: Overcomes: Wood→Earth, Fire→Metal, Earth→Water, Metal→Wood, Water→Fire.
	// So Water→Earth is neutral.
	cave2, source2 := buildTwoRoomCave(t, "earth_room", "earth_room")
	veinNeutral, err := BuildDragonVein(cave2, source2, types.Water, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein neutral: %v", err)
	}
	engineNeutral := NewChiFlowEngine(cave2, []*DragonVein{veinNeutral}, reg, params)
	engineNeutral.Tick()
	chiNeutral := engineNeutral.RoomChi[1].Current

	if chiOc >= chiNeutral {
		t.Errorf("overcomes chi (%v) should be < neutral chi (%v)", chiOc, chiNeutral)
	}
}

func TestChiFlowEngine_AdjacencyPropagation(t *testing.T) {
	cave, _ := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	// No dragon veins — chi flows only via adjacency propagation.
	engine := NewChiFlowEngine(cave, nil, reg, params)

	// Manually set room 1 chi high, room 2 at zero.
	engine.RoomChi[1].Current = 50.0
	engine.RoomChi[2].Current = 0.0

	engine.Tick()

	// Room 2 should have gained chi from room 1.
	rc2 := engine.RoomChi[2]
	if rc2.Current <= 0 {
		t.Errorf("room 2 chi = %v after tick, want > 0 (propagation from room 1)", rc2.Current)
	}

	// Room 1 should have lost some chi.
	rc1 := engine.RoomChi[1]
	if rc1.Current >= 50.0 {
		t.Errorf("room 1 chi (%v) should be < 50.0 after propagation", rc1.Current)
	}

	// Chi flows from high to low: room 1 should still have more.
	if rc1.Current <= rc2.Current {
		t.Errorf("room 1 chi (%v) should be > room 2 chi (%v)", rc1.Current, rc2.Current)
	}
}

func TestChiFlowEngine_AdjacencyElementMultiplier(t *testing.T) {
	// Test that element affinity affects propagation.
	// Wood generates Fire: propagation from wood room to fire room should be boosted.
	reg := testRegistry()
	params := DefaultFlowParams()

	cave1, _ := buildTwoRoomCave(t, "wood_room", "fire_room")
	vein1, _ := BuildDragonVein(cave1, types.Pos{X: 1, Y: 1}, types.Wood, 10.0)
	engine1 := NewChiFlowEngine(cave1, []*DragonVein{vein1}, reg, params)
	for range 10 {
		engine1.Tick()
	}
	chiFireFromWood := engine1.RoomChi[2].Current

	// Neutral pair: wood room -> metal room (no generates/overcomes between them
	// as source of propagation). Actually Metal→Wood is overcomes, Wood→Metal is neutral.
	// Generates: Wood→Fire, Fire→Earth, Earth→Metal, Metal→Water, Water→Wood
	// So Wood→Metal is neutral.
	cave2, _ := buildTwoRoomCave(t, "wood_room", "metal_room")
	vein2, _ := BuildDragonVein(cave2, types.Pos{X: 1, Y: 1}, types.Wood, 10.0)
	engine2 := NewChiFlowEngine(cave2, []*DragonVein{vein2}, reg, params)
	for range 10 {
		engine2.Tick()
	}
	chiMetalFromWood := engine2.RoomChi[2].Current

	if chiFireFromWood <= chiMetalFromWood {
		t.Errorf("fire room chi (%v, generates) should be > metal room chi (%v, neutral)", chiFireFromWood, chiMetalFromWood)
	}
}

func TestChiFlowEngine_BaseDecay(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, params)

	// Supply chi for 5 ticks.
	for range 5 {
		engine.Tick()
	}
	chiBeforeDecay := engine.RoomChi[1].Current

	// Remove veins so no more supply, only decay and propagation.
	engine.Veins = nil
	engine.Tick()

	chiAfterDecay := engine.RoomChi[1].Current
	if chiAfterDecay >= chiBeforeDecay {
		t.Errorf("chi after decay (%v) should be < before decay (%v)", chiAfterDecay, chiBeforeDecay)
	}
}

func TestChiFlowEngine_CapacityClamp(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	// High flow rate to exceed capacity quickly.
	vein, err := BuildDragonVein(cave, source, types.Wood, 200.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, params)

	// Run many ticks.
	for range 50 {
		engine.Tick()
	}

	for _, rc := range engine.RoomChi {
		if rc.Current > rc.Capacity {
			t.Errorf("room %d chi (%v) exceeds capacity (%v)", rc.RoomID, rc.Current, rc.Capacity)
		}
		if rc.Current < 0 {
			t.Errorf("room %d chi (%v) is negative", rc.RoomID, rc.Current)
		}
	}
}

func TestChiFlowEngine_NoVeinNoDirectSupply(t *testing.T) {
	cave, _ := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	// No dragon veins at all.
	engine := NewChiFlowEngine(cave, nil, reg, params)

	engine.Tick()

	for _, rc := range engine.RoomChi {
		if rc.Current != 0 {
			t.Errorf("room %d chi = %v, want 0 (no dragon vein supply)", rc.RoomID, rc.Current)
		}
	}
}

func TestChiFlowEngine_RoomNotOnVeinNoDirectSupply(t *testing.T) {
	// Create a cave where room 2 is disconnected from the vein path.
	cave, err := world.NewCave(16, 6)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	reg := testRegistry()

	// Room 1 connected to corridor.
	_, err = cave.AddRoom("wood_room", types.Pos{X: 1, Y: 1}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}
	// Room 2 isolated (no corridor).
	_, err = cave.AddRoom("wood_room", types.Pos{X: 12, Y: 1}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 12, Y: 2}, Dir: types.West},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	// Vein from room 1's entrance.
	vein, err := BuildDragonVein(cave, types.Pos{X: 1, Y: 1}, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	params := DefaultFlowParams()
	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, params)

	// Tick: room 2 should get no supply and remain at 0.
	engine.Tick()

	if engine.RoomChi[2].Current != 0 {
		t.Errorf("isolated room 2 chi = %v, want 0", engine.RoomChi[2].Current)
	}
	if engine.RoomChi[1].Current <= 0 {
		t.Errorf("room 1 chi = %v, want > 0", engine.RoomChi[1].Current)
	}
}

func TestChiFlowEngine_SteadyState(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 5.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, params)

	// Run many ticks to approach steady state.
	var prev1, prev2 float64
	for range 200 {
		engine.Tick()
		prev1 = engine.RoomChi[1].Current
		prev2 = engine.RoomChi[2].Current
	}

	// Run a few more ticks and check that values barely change.
	engine.Tick()
	cur1 := engine.RoomChi[1].Current
	cur2 := engine.RoomChi[2].Current

	const epsilon = 0.01
	if math.Abs(cur1-prev1) > epsilon {
		t.Errorf("room 1 not at steady state: delta = %v", math.Abs(cur1-prev1))
	}
	if math.Abs(cur2-prev2) > epsilon {
		t.Errorf("room 2 not at steady state: delta = %v", math.Abs(cur2-prev2))
	}
}

func TestChiFlowEngine_OnCaveChanged(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, params)

	// Initial state: 2 rooms.
	if len(engine.RoomChi) != 2 {
		t.Fatalf("initial RoomChi count = %d, want 2", len(engine.RoomChi))
	}

	// Add a third room and connect it.
	_, err = cave.AddRoom("fire_room", types.Pos{X: 1, Y: 4}, 2, 1, []world.RoomEntrance{
		{Pos: types.Pos{X: 2, Y: 4}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom 3: %v", err)
	}

	engine.OnCaveChanged(cave)

	// Should now have 3 RoomChi entries.
	if len(engine.RoomChi) != 3 {
		t.Errorf("after OnCaveChanged RoomChi count = %d, want 3", len(engine.RoomChi))
	}

	rc3, ok := engine.RoomChi[3]
	if !ok {
		t.Fatal("room 3 not found in RoomChi after OnCaveChanged")
	}
	if rc3.Element != types.Fire {
		t.Errorf("room 3 element = %v, want Fire", rc3.Element)
	}
}

func TestDefaultFlowParams(t *testing.T) {
	p := DefaultFlowParams()
	if p.GeneratesMultiplier != 1.3 {
		t.Errorf("GeneratesMultiplier = %v, want 1.3", p.GeneratesMultiplier)
	}
	if p.OvercomesMultiplier != 0.6 {
		t.Errorf("OvercomesMultiplier = %v, want 0.6", p.OvercomesMultiplier)
	}
	if p.SameElementMultiplier != 1.1 {
		t.Errorf("SameElementMultiplier = %v, want 1.1", p.SameElementMultiplier)
	}
	if p.NeutralMultiplier != 1.0 {
		t.Errorf("NeutralMultiplier = %v, want 1.0", p.NeutralMultiplier)
	}
	if p.BaseDecayRate != 0.02 {
		t.Errorf("BaseDecayRate = %v, want 0.02", p.BaseDecayRate)
	}
}

func TestLoadFlowParams(t *testing.T) {
	// Write a temp JSON file and load it.
	dir := t.TempDir()
	path := dir + "/params.json"
	data := []byte(`{"generates_multiplier":2.0,"overcomes_multiplier":0.5,"same_element_multiplier":1.5,"neutral_multiplier":1.0,"base_decay_rate":0.05}`)
	if err := writeTestFile(path, data); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	p, err := LoadFlowParams(path)
	if err != nil {
		t.Fatalf("LoadFlowParams: %v", err)
	}
	if p.GeneratesMultiplier != 2.0 {
		t.Errorf("GeneratesMultiplier = %v, want 2.0", p.GeneratesMultiplier)
	}
	if p.BaseDecayRate != 0.05 {
		t.Errorf("BaseDecayRate = %v, want 0.05", p.BaseDecayRate)
	}
}

func TestLoadFlowParams_FileNotFound(t *testing.T) {
	_, err := LoadFlowParams("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func writeTestFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
