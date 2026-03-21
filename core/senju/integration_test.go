package senju

import (
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// testBeastRegistry creates a room type registry with beast-compatible rooms
// for each element (all with MaxBeasts >= 3 and high BaseChiCapacity).
func testBeastRegistry() *world.RoomTypeRegistry {
	reg := world.NewRoomTypeRegistry()
	_ = reg.Register(world.RoomType{ID: "wood_room", Name: "Wood Room", Element: types.Wood, BaseChiCapacity: 500, MaxBeasts: 3})
	_ = reg.Register(world.RoomType{ID: "fire_room", Name: "Fire Room", Element: types.Fire, BaseChiCapacity: 500, MaxBeasts: 3})
	_ = reg.Register(world.RoomType{ID: "water_room", Name: "Water Room", Element: types.Water, BaseChiCapacity: 500, MaxBeasts: 3})
	_ = reg.Register(world.RoomType{ID: "earth_room", Name: "Earth Room", Element: types.Earth, BaseChiCapacity: 500, MaxBeasts: 3})
	return reg
}

// TestIntegration_BeastGrowthSimulation places 3 beasts in rooms with different
// element affinities and runs a 30-tick growth simulation to verify:
//   - Beasts in same-element rooms grow faster than beasts in overcomes rooms
//   - Chi is consumed from RoomChi as beasts grow
//   - CalcCombatStats varies by room element affinity
//
// Room layout (32x32 cave):
//
//	Room 1 (Wood)  at (2,2)  3x3 — entrance south  ← 翠龍 (Wood, same element → 1.1x)
//	Room 2 (Fire)  at (8,2)  3x3 — entrance south  ← 炎鳳 (Fire, same element → 1.1x)
//	Room 3 (Earth) at (14,2) 3x3 — entrance south  ← 水蛇 (Water in Earth room; Earth overcomes Water → 0.7x)
//
// Dragon vein: Wood element from (3,5), FlowRate=10.0
func TestIntegration_BeastGrowthSimulation(t *testing.T) {
	// --- Setup registries ---
	roomReg := testBeastRegistry()
	speciesReg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("LoadDefaultSpecies: %v", err)
	}

	// --- Build cave ---
	cave, err := world.NewCave(32, 32)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	type roomDef struct {
		typeID    string
		pos       types.Pos
		entrances []world.RoomEntrance
	}

	defs := []roomDef{
		{
			typeID: "wood_room",
			pos:    types.Pos{X: 2, Y: 2},
			entrances: []world.RoomEntrance{
				{Pos: types.Pos{X: 3, Y: 4}, Dir: types.South},
			},
		},
		{
			typeID: "fire_room",
			pos:    types.Pos{X: 8, Y: 2},
			entrances: []world.RoomEntrance{
				{Pos: types.Pos{X: 9, Y: 4}, Dir: types.South},
			},
		},
		{
			typeID: "earth_room",
			pos:    types.Pos{X: 14, Y: 2},
			entrances: []world.RoomEntrance{
				{Pos: types.Pos{X: 15, Y: 4}, Dir: types.South},
			},
		},
	}

	rooms := make([]*world.Room, 0, len(defs))
	for i, d := range defs {
		r, err := cave.AddRoom(d.typeID, d.pos, 3, 3, d.entrances)
		if err != nil {
			t.Fatalf("AddRoom[%d] %s: %v", i, d.typeID, err)
		}
		rooms = append(rooms, r)
	}

	// Connect rooms: 1-2, 2-3 so dragon vein can reach all.
	connections := [][2]int{
		{rooms[0].ID, rooms[1].ID},
		{rooms[1].ID, rooms[2].ID},
	}
	for i, c := range connections {
		_, err := cave.ConnectRooms(c[0], c[1])
		if err != nil {
			t.Fatalf("ConnectRooms[%d] (%d,%d): %v", i, c[0], c[1], err)
		}
	}

	// --- Build dragon vein ---
	vein, err := fengshui.BuildDragonVein(cave, types.Pos{X: 3, Y: 5}, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	reachedRooms := vein.RoomsOnPath(cave)
	if len(reachedRooms) == 0 {
		t.Fatal("dragon vein reaches no rooms")
	}

	// --- Initialize chi flow engine and pre-fill chi ---
	flowParams := fengshui.DefaultFlowParams()
	chiEngine := fengshui.NewChiFlowEngine(cave, []*fengshui.DragonVein{vein}, roomReg, flowParams)

	// Pre-fill chi so rooms have sufficient supply before beasts start consuming.
	for range 50 {
		chiEngine.Tick()
	}

	// All rooms should have positive chi after pre-fill.
	for id, rc := range chiEngine.RoomChi {
		if rc.Current <= 0 {
			t.Fatalf("room %d has no chi after pre-fill (current=%v)", id, rc.Current)
		}
	}

	// --- Create and place beasts ---
	suiryu, err := speciesReg.Get("suiryu")
	if err != nil {
		t.Fatalf("Get suiryu: %v", err)
	}
	enhou, err := speciesReg.Get("enhou")
	if err != nil {
		t.Fatalf("Get enhou: %v", err)
	}
	suija, err := speciesReg.Get("suija")
	if err != nil {
		t.Fatalf("Get suija: %v", err)
	}

	beast1 := NewBeast(1, suiryu, 0) // Wood beast
	beast2 := NewBeast(2, enhou, 0)  // Fire beast
	beast3 := NewBeast(3, suija, 0)  // Water beast

	rt1, _ := roomReg.Get("wood_room")
	if err := PlaceBeast(beast1, rooms[0], rt1); err != nil {
		t.Fatalf("PlaceBeast 1: %v", err)
	}

	rt2, _ := roomReg.Get("fire_room")
	if err := PlaceBeast(beast2, rooms[1], rt2); err != nil {
		t.Fatalf("PlaceBeast 2: %v", err)
	}

	rt3, _ := roomReg.Get("earth_room")
	if err := PlaceBeast(beast3, rooms[2], rt3); err != nil {
		t.Fatalf("PlaceBeast 3: %v", err)
	}

	// --- Run 30 ticks of growth simulation ---
	growthParams := DefaultGrowthParams()
	growthEngine := NewGrowthEngine(growthParams, speciesReg)

	beasts := []*Beast{beast1, beast2, beast3}
	roomMap := make(map[int]*world.Room)
	for _, r := range rooms {
		roomMap[r.ID] = r
	}

	// Track total EXP gained per beast from events (beast.EXP resets on level-up).
	totalEXP := map[int]int{beast1.ID: 0, beast2.ID: 0, beast3.ID: 0}
	starvedCount := map[int]int{beast1.ID: 0, beast2.ID: 0, beast3.ID: 0}

	for range 30 {
		chiEngine.Tick()
		events := growthEngine.Tick(beasts, chiEngine.RoomChi, roomMap)
		for _, e := range events {
			switch e.Type {
			case EXPGained:
				totalEXP[e.BeastID] += e.EXPGained
			case ChiStarved:
				starvedCount[e.BeastID]++
			}
		}
	}

	// --- Verify: beasts in favorable rooms gained more total EXP ---
	// Wood beast (same element 1.1x, GrowthRate 1.0): ~11 EXP/tick
	// Fire beast (same element 1.1x, GrowthRate 1.1): ~12 EXP/tick
	// Water beast (overcomes 0.7x, GrowthRate 0.95): ~6 EXP/tick

	t.Logf("Total EXP: Wood=%d, Fire=%d, Water=%d", totalEXP[1], totalEXP[2], totalEXP[3])
	t.Logf("Levels: Wood=%d, Fire=%d, Water=%d", beast1.Level, beast2.Level, beast3.Level)
	t.Logf("Starved ticks: Wood=%d, Fire=%d, Water=%d", starvedCount[1], starvedCount[2], starvedCount[3])

	// Water beast (overcomes) should have gained less total EXP than favorable beasts.
	if totalEXP[beast3.ID] >= totalEXP[beast1.ID] {
		t.Errorf("overcomes beast (Water in Earth) totalEXP=%d should be < same-element beast (Wood in Wood) totalEXP=%d",
			totalEXP[beast3.ID], totalEXP[beast1.ID])
	}
	if totalEXP[beast3.ID] >= totalEXP[beast2.ID] {
		t.Errorf("overcomes beast (Water in Earth) totalEXP=%d should be < same-element beast (Fire in Fire) totalEXP=%d",
			totalEXP[beast3.ID], totalEXP[beast2.ID])
	}

	// All beasts should have gained some EXP (chi should be sufficient).
	for _, b := range beasts {
		if totalEXP[b.ID] <= 0 {
			t.Errorf("beast %d (%s) should have gained EXP, totalEXP=%d", b.ID, b.Name, totalEXP[b.ID])
		}
	}

	// --- Verify: chi consumption reflected in RoomChi ---
	// All rooms should have non-negative chi.
	for _, b := range beasts {
		rc := chiEngine.RoomChi[b.RoomID]
		if rc.Current < 0 {
			t.Errorf("room %d chi = %v, should not be negative", b.RoomID, rc.Current)
		}
	}

	// Chi should have been consumed: at least some rooms should not be at full capacity.
	anyConsumed := false
	for _, b := range beasts {
		rc := chiEngine.RoomChi[b.RoomID]
		if rc.Current < rc.Capacity {
			anyConsumed = true
			break
		}
	}
	if !anyConsumed {
		t.Error("expected chi consumption to reduce at least one room below capacity")
	}

	// --- Verify: CalcCombatStats varies by room placement ---
	// Wood beast in Wood room (same element → 1.1x multiplier on ATK/DEF/SPD)
	stats1 := beast1.CalcCombatStats(chiEngine.RoomChi[beast1.RoomID])
	// Fire beast in Fire room (same element → 1.1x)
	stats2 := beast2.CalcCombatStats(chiEngine.RoomChi[beast2.RoomID])
	// Water beast in Earth room (overcomes → 0.7x multiplier)
	stats3 := beast3.CalcCombatStats(chiEngine.RoomChi[beast3.RoomID])

	t.Logf("CombatStats ATK: Wood=%d, Fire=%d, Water=%d", stats1.ATK, stats2.ATK, stats3.ATK)

	// Wood beast (base ATK 30 + level-ups, ×1.1) should outperform
	// Water beast (base ATK 25 + fewer level-ups, ×0.7).
	if stats1.ATK <= stats3.ATK {
		t.Errorf("Wood beast in Wood room ATK=%d should be > Water beast in Earth room ATK=%d",
			stats1.ATK, stats3.ATK)
	}

	// Fire beast (base ATK 45, ×1.1) should have highest effective ATK.
	if stats2.ATK <= stats3.ATK {
		t.Errorf("Fire beast in Fire room ATK=%d should be > Water beast in Earth room ATK=%d",
			stats2.ATK, stats3.ATK)
	}

	// Verify CalcCombatStats(nil) returns base stats (no affinity applied).
	baseStats := beast1.CalcCombatStats(nil)
	if baseStats.ATK != beast1.ATK {
		t.Errorf("CalcCombatStats(nil) ATK=%d should equal beast base ATK=%d",
			baseStats.ATK, beast1.ATK)
	}
}
