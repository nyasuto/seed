package ai

import (
	"encoding/json"
	"testing"

	"github.com/nyasuto/seed/core/economy"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// newTestState creates a minimal GameState for testing valid action generation.
// The cave is 20x20 with one room at (2,2) of size 3x3, and 500 chi deposited.
func newTestState(t *testing.T) *simulation.GameState {
	t.Helper()

	cave, err := world.NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	entrances := []world.RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 4}, Dir: types.South},
	}
	_, err = cave.AddRoom("trap_room", types.Pos{X: 2, Y: 2}, 3, 3, entrances)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	roomReg, err := world.LoadDefaultRoomTypes()
	if err != nil {
		t.Fatalf("LoadDefaultRoomTypes: %v", err)
	}

	speciesReg, err := senju.LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("LoadDefaultSpecies: %v", err)
	}

	evolutionReg, err := senju.LoadDefaultEvolution()
	if err != nil {
		t.Fatalf("LoadDefaultEvolution: %v", err)
	}

	chiPool := economy.NewChiPool(1000)
	_ = chiPool.Deposit(500, economy.Supply, "test init", 0)

	sp, err := economy.DefaultSupplyParams()
	if err != nil {
		t.Fatalf("DefaultSupplyParams: %v", err)
	}
	cp, err := economy.DefaultCostParams()
	if err != nil {
		t.Fatalf("DefaultCostParams: %v", err)
	}
	dp, err := economy.DefaultDeficitParams()
	if err != nil {
		t.Fatalf("DefaultDeficitParams: %v", err)
	}
	cc, err := economy.DefaultConstructionCost()
	if err != nil {
		t.Fatalf("DefaultConstructionCost: %v", err)
	}
	bc, err := economy.DefaultBeastCost()
	if err != nil {
		t.Fatalf("DefaultBeastCost: %v", err)
	}
	engine := economy.NewEconomyEngine(chiPool, sp, cp, dp, cc, bc)

	return &simulation.GameState{
		Cave:              cave,
		RoomTypeRegistry:  roomReg,
		SpeciesRegistry:   speciesReg,
		EvolutionRegistry: evolutionReg,
		EconomyEngine:     engine,
		Beasts:            make([]*senju.Beast, 0),
		NextBeastID:       1,
		Progress:          &scenario.ScenarioProgress{CurrentTick: 0},
	}
}

func TestSnapshotToJSON(t *testing.T) {
	snapshot := scenario.GameSnapshot{
		Tick:           5,
		CoreHP:         100,
		ChiPoolBalance: 250.5,
		BeastCount:     2,
		AliveBeasts:    1,
	}

	data, err := SnapshotToJSON(snapshot)
	if err != nil {
		t.Fatalf("SnapshotToJSON: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["Tick"] != float64(5) {
		t.Errorf("Tick = %v, want 5", decoded["Tick"])
	}
	if decoded["CoreHP"] != float64(100) {
		t.Errorf("CoreHP = %v, want 100", decoded["CoreHP"])
	}
	if decoded["ChiPoolBalance"] != 250.5 {
		t.Errorf("ChiPoolBalance = %v, want 250.5", decoded["ChiPoolBalance"])
	}
}

func TestBuildValidActions_InitialState(t *testing.T) {
	state := newTestState(t)

	actions := BuildValidActions(state)

	// Should always have a "wait" action.
	hasWait := false
	hasDigRoom := false
	hasSummon := false
	hasUpgrade := false
	for _, a := range actions {
		switch a.Kind {
		case "wait":
			hasWait = true
		case "dig_room":
			hasDigRoom = true
		case "summon_beast":
			hasSummon = true
		case "upgrade_room":
			hasUpgrade = true
		}
	}

	if !hasWait {
		t.Error("expected wait action to always be present")
	}
	if !hasDigRoom {
		t.Error("expected dig_room actions (cave has open space and chi)")
	}
	if !hasSummon {
		t.Error("expected summon_beast actions (chi pool has balance)")
	}
	if !hasUpgrade {
		t.Error("expected upgrade_room actions (existing room can be upgraded)")
	}
}

func TestBuildValidActions_ChiInsufficient_NoSummon(t *testing.T) {
	state := newTestState(t)

	// Drain chi to near zero.
	bal := state.EconomyEngine.ChiPool.Balance()
	_ = state.EconomyEngine.ChiPool.Withdraw(bal-0.01, economy.Construction, "drain", 0)

	actions := BuildValidActions(state)

	for _, a := range actions {
		if a.Kind == "summon_beast" {
			t.Errorf("summon_beast should not appear when chi is insufficient, got element=%v", a.Params["element"])
		}
		if a.Kind == "dig_room" {
			t.Errorf("dig_room should not appear when chi is insufficient")
		}
	}
}

func TestBuildValidActions_MaxRoomsReached_NoDigRoom(t *testing.T) {
	state := newTestState(t)

	// Set MaxRooms = 1 (we already have 1 room).
	state.Scenario = &scenario.Scenario{
		Constraints: scenario.GameConstraints{MaxRooms: 1},
	}

	actions := BuildValidActions(state)

	for _, a := range actions {
		if a.Kind == "dig_room" {
			t.Error("dig_room should not appear when MaxRooms is reached")
		}
	}

	// Wait should still be present.
	hasWait := false
	for _, a := range actions {
		if a.Kind == "wait" {
			hasWait = true
		}
	}
	if !hasWait {
		t.Error("wait should always be present")
	}
}

func TestBuildValidActions_DigRoomParams(t *testing.T) {
	state := newTestState(t)

	actions := BuildValidActions(state)

	// Find a dig_room action and verify its params.
	for _, a := range actions {
		if a.Kind != "dig_room" {
			continue
		}
		if _, ok := a.Params["room_type_id"]; !ok {
			t.Error("dig_room should have room_type_id param")
		}
		if _, ok := a.Params["x"]; !ok {
			t.Error("dig_room should have x param")
		}
		if _, ok := a.Params["y"]; !ok {
			t.Error("dig_room should have y param")
		}
		if w, ok := a.Params["width"]; !ok || w != 3 {
			t.Errorf("dig_room width = %v, want 3", w)
		}
		if h, ok := a.Params["height"]; !ok || h != 3 {
			t.Errorf("dig_room height = %v, want 3", h)
		}
		if _, ok := a.Params["cost"]; !ok {
			t.Error("dig_room should have cost param")
		}
		return // Only need to check one.
	}
	t.Fatal("no dig_room action found")
}

func TestBuildValidActions_DigCorridor(t *testing.T) {
	state := newTestState(t)

	// Add a second room so corridors are possible.
	entrances := []world.RoomEntrance{
		{Pos: types.Pos{X: 9, Y: 11}, Dir: types.South},
	}
	_, err := state.Cave.AddRoom("fire_room", types.Pos{X: 8, Y: 9}, 3, 3, entrances)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	actions := BuildValidActions(state)

	hasCorridor := false
	for _, a := range actions {
		if a.Kind == "dig_corridor" {
			hasCorridor = true
			if _, ok := a.Params["from_room_id"]; !ok {
				t.Error("dig_corridor should have from_room_id param")
			}
			if _, ok := a.Params["to_room_id"]; !ok {
				t.Error("dig_corridor should have to_room_id param")
			}
		}
	}
	if !hasCorridor {
		t.Error("expected dig_corridor action when two rooms with entrances exist")
	}
}

func TestBuildValidActions_PlaceBeast(t *testing.T) {
	state := newTestState(t)

	// Summon a beast so there's an unassigned one.
	species := state.SpeciesRegistry.All()
	if len(species) == 0 {
		t.Fatal("no species registered")
	}
	beast := senju.NewBeast(state.NextBeastID, species[0], 0)
	state.NextBeastID++
	state.Beasts = append(state.Beasts, beast)

	actions := BuildValidActions(state)

	hasPlace := false
	for _, a := range actions {
		if a.Kind == "place_beast" {
			hasPlace = true
			if _, ok := a.Params["species_id"]; !ok {
				t.Error("place_beast should have species_id param")
			}
			if _, ok := a.Params["room_id"]; !ok {
				t.Error("place_beast should have room_id param")
			}
		}
	}
	if !hasPlace {
		t.Error("expected place_beast action when unassigned beast exists")
	}
}

func TestBuildValidActions_NoPlaceBeast_WhenAllAssigned(t *testing.T) {
	state := newTestState(t)

	// Add a beast that's already assigned to a room.
	species := state.SpeciesRegistry.All()
	if len(species) == 0 {
		t.Fatal("no species registered")
	}
	beast := senju.NewBeast(state.NextBeastID, species[0], 0)
	beast.RoomID = 1 // assigned
	state.NextBeastID++
	state.Beasts = append(state.Beasts, beast)

	actions := BuildValidActions(state)

	for _, a := range actions {
		if a.Kind == "place_beast" {
			t.Error("place_beast should not appear when all beasts are assigned")
		}
	}
}

func TestBuildValidActions_SummonBeastParams(t *testing.T) {
	state := newTestState(t)

	actions := BuildValidActions(state)

	for _, a := range actions {
		if a.Kind != "summon_beast" {
			continue
		}
		elem, ok := a.Params["element"]
		if !ok {
			t.Error("summon_beast should have element param")
		}
		if _, ok := elem.(string); !ok {
			t.Errorf("summon_beast element should be string, got %T", elem)
		}
		if _, ok := a.Params["cost"]; !ok {
			t.Error("summon_beast should have cost param")
		}
		return
	}
	t.Fatal("no summon_beast action found")
}

func TestBuildStateMessage(t *testing.T) {
	state := newTestState(t)

	snapshot := simulation.BuildSnapshot(state)

	sb := NewStateBuilder(func() *simulation.SimulationEngine {
		return &simulation.SimulationEngine{State: state}
	})

	msg, err := sb.BuildStateMessage(snapshot)
	if err != nil {
		t.Fatalf("BuildStateMessage: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}
	if msg.Type != "state" {
		t.Errorf("Type = %q, want %q", msg.Type, "state")
	}
	if msg.Tick != int(snapshot.Tick) {
		t.Errorf("Tick = %d, want %d", msg.Tick, snapshot.Tick)
	}
	if len(msg.ValidActions) == 0 {
		t.Error("expected non-empty ValidActions")
	}
	if len(msg.Snapshot) == 0 {
		t.Error("expected non-empty Snapshot JSON")
	}
}

func TestBuildStateMessage_NilEngine(t *testing.T) {
	sb := NewStateBuilder(func() *simulation.SimulationEngine {
		return nil
	})

	msg, err := sb.BuildStateMessage(scenario.GameSnapshot{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Error("expected nil message when engine is nil")
	}
}
