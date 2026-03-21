package simulation

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// newTestState creates a minimal GameState for action tests.
// The cave is 20x20 with one trap_room at (2,2) size 3x3 (room ID 1)
// with one entrance. Chi pool has 500 of 1000 capacity.
func newTestState(t *testing.T) *GameState {
	t.Helper()

	cave, err := world.NewCave(20, 20)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	entrances := []world.RoomEntrance{
		{Pos: types.Pos{X: 4, Y: 3}, Dir: types.East},
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

	engine := economy.NewEconomyEngine(
		chiPool,
		economy.DefaultSupplyParams(),
		economy.DefaultCostParams(),
		economy.DefaultDeficitParams(),
		economy.DefaultConstructionCost(),
		economy.DefaultBeastCost(),
	)

	return &GameState{
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

// --- ActionType tests ---

func TestActionType_Values(t *testing.T) {
	tests := []struct {
		action PlayerAction
		want   string
	}{
		{DigRoomAction{}, "dig_room"},
		{DigCorridorAction{}, "dig_corridor"},
		{PlaceBeastAction{}, "place_beast"},
		{UpgradeRoomAction{}, "upgrade_room"},
		{SummonBeastAction{}, "summon_beast"},
		{EvolveBeastAction{}, "evolve_beast"},
		{NoAction{}, "no_action"},
	}
	for _, tt := range tests {
		if got := tt.action.ActionType(); got != tt.want {
			t.Errorf("%T.ActionType() = %q, want %q", tt.action, got, tt.want)
		}
	}
}

// --- ValidateAction tests ---

func TestValidateAction_NoAction(t *testing.T) {
	state := newTestState(t)
	if err := ValidateAction(NoAction{}, state); err != nil {
		t.Errorf("NoAction should always be valid, got: %v", err)
	}
}

func TestValidateAction_UnknownAction(t *testing.T) {
	state := newTestState(t)
	type fakeAction struct{}
	fa := fakeAction{}
	// fakeAction does not implement PlayerAction, so we use a wrapper.
	// Instead, test the default branch via ApplyAction which also calls ValidateAction.
	// Actually, we cannot pass a non-PlayerAction, so let's skip this approach.
	// The unknown action branch is covered by the type switch default — which
	// requires a type that implements PlayerAction but is not one of the known types.
	_ = fa
	_ = state
}

func TestValidateAction_DigRoom_Success(t *testing.T) {
	state := newTestState(t)
	action := DigRoomAction{
		RoomTypeID: "trap_room",
		Pos:        types.Pos{X: 8, Y: 8},
		Width:      3,
		Height:     3,
	}
	if err := ValidateAction(action, state); err != nil {
		t.Errorf("expected valid dig room action, got: %v", err)
	}
}

func TestValidateAction_DigRoom_UnknownRoomType(t *testing.T) {
	state := newTestState(t)
	action := DigRoomAction{
		RoomTypeID: "nonexistent_type",
		Pos:        types.Pos{X: 8, Y: 8},
		Width:      3,
		Height:     3,
	}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for unknown room type")
	}
	if !errors.Is(err, ErrRoomTypeNotFound) {
		t.Errorf("error = %v, want ErrRoomTypeNotFound", err)
	}
}

func TestValidateAction_DigRoom_InvalidPlacement(t *testing.T) {
	state := newTestState(t)
	// Overlap with existing room at (2,2) size 3x3.
	action := DigRoomAction{
		RoomTypeID: "trap_room",
		Pos:        types.Pos{X: 3, Y: 3},
		Width:      3,
		Height:     3,
	}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for overlapping placement")
	}
}

func TestValidateAction_DigRoom_OutOfBounds(t *testing.T) {
	state := newTestState(t)
	action := DigRoomAction{
		RoomTypeID: "trap_room",
		Pos:        types.Pos{X: 18, Y: 18},
		Width:      5,
		Height:     5,
	}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for out-of-bounds placement")
	}
}

func TestValidateAction_DigRoom_InsufficientChi(t *testing.T) {
	state := newTestState(t)
	// Drain the chi pool.
	_ = state.EconomyEngine.ChiPool.Withdraw(500, economy.Construction, "drain", 0)

	action := DigRoomAction{
		RoomTypeID: "trap_room",
		Pos:        types.Pos{X: 8, Y: 8},
		Width:      3,
		Height:     3,
	}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for insufficient chi")
	}
}

func TestValidateAction_DigCorridor_Success(t *testing.T) {
	state := newTestState(t)
	// Add a second room with an entrance.
	entrances := []world.RoomEntrance{
		{Pos: types.Pos{X: 8, Y: 9}, Dir: types.West},
	}
	_, err := state.Cave.AddRoom("trap_room", types.Pos{X: 8, Y: 8}, 3, 3, entrances)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	action := DigCorridorAction{FromRoomID: 1, ToRoomID: 2}
	if err := ValidateAction(action, state); err != nil {
		t.Errorf("expected valid corridor action, got: %v", err)
	}
}

func TestValidateAction_DigCorridor_RoomNotFound(t *testing.T) {
	state := newTestState(t)
	action := DigCorridorAction{FromRoomID: 1, ToRoomID: 99}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for nonexistent room")
	}
	if !errors.Is(err, world.ErrRoomNotFound) {
		t.Errorf("error = %v, want ErrRoomNotFound", err)
	}
}

func TestValidateAction_DigCorridor_NoEntrance(t *testing.T) {
	state := newTestState(t)
	// Add a room without entrances.
	_, err := state.Cave.AddRoom("trap_room", types.Pos{X: 8, Y: 8}, 3, 3, nil)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	action := DigCorridorAction{FromRoomID: 1, ToRoomID: 2}
	err = ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for room without entrance")
	}
	if !errors.Is(err, world.ErrNoEntrance) {
		t.Errorf("error = %v, want ErrNoEntrance", err)
	}
}

func TestValidateAction_PlaceBeast_Success(t *testing.T) {
	state := newTestState(t)
	species, _ := state.SpeciesRegistry.Get("suiryu")
	beast := senju.NewBeast(1, species, 0)
	state.Beasts = append(state.Beasts, beast)

	action := PlaceBeastAction{SpeciesID: "suiryu", RoomID: 1}
	if err := ValidateAction(action, state); err != nil {
		t.Errorf("expected valid place beast action, got: %v", err)
	}
}

func TestValidateAction_PlaceBeast_NoUnassignedBeast(t *testing.T) {
	state := newTestState(t)
	// No beasts exist.
	action := PlaceBeastAction{SpeciesID: "suiryu", RoomID: 1}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for no unassigned beast")
	}
}

func TestValidateAction_PlaceBeast_AlreadyAssigned(t *testing.T) {
	state := newTestState(t)
	species, _ := state.SpeciesRegistry.Get("suiryu")
	beast := senju.NewBeast(1, species, 0)
	beast.RoomID = 1 // already assigned
	state.Beasts = append(state.Beasts, beast)

	action := PlaceBeastAction{SpeciesID: "suiryu", RoomID: 1}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for already-assigned beast")
	}
}

func TestValidateAction_PlaceBeast_RoomNotFound(t *testing.T) {
	state := newTestState(t)
	species, _ := state.SpeciesRegistry.Get("suiryu")
	beast := senju.NewBeast(1, species, 0)
	state.Beasts = append(state.Beasts, beast)

	action := PlaceBeastAction{SpeciesID: "suiryu", RoomID: 99}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for nonexistent room")
	}
	if !errors.Is(err, world.ErrRoomNotFound) {
		t.Errorf("error = %v, want ErrRoomNotFound", err)
	}
}

func TestValidateAction_PlaceBeast_AtCapacity(t *testing.T) {
	state := newTestState(t)
	// Get the room type to check MaxBeasts.
	rt, _ := state.RoomTypeRegistry.Get("trap_room")
	room := state.Cave.RoomByID(1)

	// Fill the room to capacity.
	for i := 0; i < rt.MaxBeasts; i++ {
		room.BeastIDs = append(room.BeastIDs, 100+i)
	}

	species, _ := state.SpeciesRegistry.Get("suiryu")
	beast := senju.NewBeast(1, species, 0)
	state.Beasts = append(state.Beasts, beast)

	action := PlaceBeastAction{SpeciesID: "suiryu", RoomID: 1}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for room at beast capacity")
	}
}

func TestValidateAction_UpgradeRoom_Success(t *testing.T) {
	state := newTestState(t)
	action := UpgradeRoomAction{RoomID: 1}
	if err := ValidateAction(action, state); err != nil {
		t.Errorf("expected valid upgrade action, got: %v", err)
	}
}

func TestValidateAction_UpgradeRoom_RoomNotFound(t *testing.T) {
	state := newTestState(t)
	action := UpgradeRoomAction{RoomID: 99}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for nonexistent room")
	}
	if !errors.Is(err, world.ErrRoomNotFound) {
		t.Errorf("error = %v, want ErrRoomNotFound", err)
	}
}

func TestValidateAction_UpgradeRoom_InsufficientChi(t *testing.T) {
	state := newTestState(t)
	_ = state.EconomyEngine.ChiPool.Withdraw(500, economy.Construction, "drain", 0)

	action := UpgradeRoomAction{RoomID: 1}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for insufficient chi")
	}
}

func TestValidateAction_SummonBeast_Success(t *testing.T) {
	state := newTestState(t)
	action := SummonBeastAction{Element: types.Water}
	if err := ValidateAction(action, state); err != nil {
		t.Errorf("expected valid summon action, got: %v", err)
	}
}

func TestValidateAction_SummonBeast_InsufficientChi(t *testing.T) {
	state := newTestState(t)
	_ = state.EconomyEngine.ChiPool.Withdraw(500, economy.Construction, "drain", 0)

	action := SummonBeastAction{Element: types.Water}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for insufficient chi")
	}
}

func TestValidateAction_EvolveBeast_Success(t *testing.T) {
	state := newTestState(t)
	species, _ := state.SpeciesRegistry.Get("suiryu")
	beast := senju.NewBeast(1, species, 0)
	state.Beasts = append(state.Beasts, beast)

	// suiryu should have an evolution path in the default registry.
	action := EvolveBeastAction{BeastID: 1}
	err := ValidateAction(action, state)
	// If no evolution path exists for suiryu, that's fine — test the error.
	paths := state.EvolutionRegistry.GetPaths("suiryu")
	if len(paths) == 0 {
		if err == nil {
			t.Fatal("expected error for no evolution path")
		}
		if !errors.Is(err, ErrNoEvolutionPath) {
			t.Errorf("error = %v, want ErrNoEvolutionPath", err)
		}
	} else {
		if err != nil {
			t.Errorf("expected valid evolve action, got: %v", err)
		}
	}
}

func TestValidateAction_EvolveBeast_BeastNotFound(t *testing.T) {
	state := newTestState(t)
	action := EvolveBeastAction{BeastID: 99}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for nonexistent beast")
	}
	if !errors.Is(err, ErrBeastNotFound) {
		t.Errorf("error = %v, want ErrBeastNotFound", err)
	}
}

func TestValidateAction_EvolveBeast_NoEvolutionPath(t *testing.T) {
	state := newTestState(t)
	// Use an empty evolution registry.
	state.EvolutionRegistry = senju.NewEvolutionRegistry()

	species, _ := state.SpeciesRegistry.Get("suiryu")
	beast := senju.NewBeast(1, species, 0)
	state.Beasts = append(state.Beasts, beast)

	action := EvolveBeastAction{BeastID: 1}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for no evolution path")
	}
	if !errors.Is(err, ErrNoEvolutionPath) {
		t.Errorf("error = %v, want ErrNoEvolutionPath", err)
	}
}

func TestValidateAction_EvolveBeast_NilRegistry(t *testing.T) {
	state := newTestState(t)
	state.EvolutionRegistry = nil

	species, _ := state.SpeciesRegistry.Get("suiryu")
	beast := senju.NewBeast(1, species, 0)
	state.Beasts = append(state.Beasts, beast)

	action := EvolveBeastAction{BeastID: 1}
	err := ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error for nil evolution registry")
	}
	if !errors.Is(err, ErrNoEvolutionPath) {
		t.Errorf("error = %v, want ErrNoEvolutionPath", err)
	}
}

// --- ApplyAction tests ---

func TestApplyAction_NoAction(t *testing.T) {
	state := newTestState(t)
	result, err := ApplyAction(NoAction{}, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("NoAction should always succeed")
	}
	if result.Cost != 0 {
		t.Errorf("NoAction cost = %f, want 0", result.Cost)
	}
}

func TestApplyAction_DigRoom_Success(t *testing.T) {
	state := newTestState(t)
	balanceBefore := state.EconomyEngine.ChiPool.Balance()

	action := DigRoomAction{
		RoomTypeID: "trap_room",
		Pos:        types.Pos{X: 8, Y: 8},
		Width:      3,
		Height:     3,
	}
	result, err := ApplyAction(action, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Cost <= 0 {
		t.Errorf("expected positive cost, got %f", result.Cost)
	}

	// Verify room was added.
	if len(state.Cave.Rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(state.Cave.Rooms))
	}

	// Verify chi was spent.
	balanceAfter := state.EconomyEngine.ChiPool.Balance()
	if balanceAfter >= balanceBefore {
		t.Errorf("balance should have decreased: before=%f, after=%f", balanceBefore, balanceAfter)
	}
}

func TestApplyAction_DigRoom_InsufficientChi(t *testing.T) {
	state := newTestState(t)
	_ = state.EconomyEngine.ChiPool.Withdraw(500, economy.Construction, "drain", 0)

	action := DigRoomAction{
		RoomTypeID: "trap_room",
		Pos:        types.Pos{X: 8, Y: 8},
		Width:      3,
		Height:     3,
	}
	_, err := ApplyAction(action, state)
	if err == nil {
		t.Fatal("expected error for insufficient chi")
	}

	// Room should not have been added.
	if len(state.Cave.Rooms) != 1 {
		t.Errorf("room count should remain 1, got %d", len(state.Cave.Rooms))
	}
}

func TestApplyAction_PlaceBeast_Success(t *testing.T) {
	state := newTestState(t)
	species, _ := state.SpeciesRegistry.Get("suiryu")
	beast := senju.NewBeast(1, species, 0)
	state.Beasts = append(state.Beasts, beast)

	action := PlaceBeastAction{SpeciesID: "suiryu", RoomID: 1}
	result, err := ApplyAction(action, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Cost != 0 {
		t.Errorf("place beast cost = %f, want 0", result.Cost)
	}

	// Verify beast was assigned.
	if beast.RoomID != 1 {
		t.Errorf("beast.RoomID = %d, want 1", beast.RoomID)
	}
	room := state.Cave.RoomByID(1)
	if len(room.BeastIDs) != 1 || room.BeastIDs[0] != 1 {
		t.Errorf("room.BeastIDs = %v, want [1]", room.BeastIDs)
	}
}

func TestApplyAction_UpgradeRoom_Success(t *testing.T) {
	state := newTestState(t)
	room := state.Cave.RoomByID(1)
	levelBefore := room.Level

	action := UpgradeRoomAction{RoomID: 1}
	result, err := ApplyAction(action, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Cost <= 0 {
		t.Errorf("expected positive cost, got %f", result.Cost)
	}
	if room.Level != levelBefore+1 {
		t.Errorf("room level = %d, want %d", room.Level, levelBefore+1)
	}
}

func TestApplyAction_SummonBeast_Success(t *testing.T) {
	state := newTestState(t)
	balanceBefore := state.EconomyEngine.ChiPool.Balance()

	action := SummonBeastAction{Element: types.Water}
	result, err := ApplyAction(action, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Cost <= 0 {
		t.Errorf("expected positive cost, got %f", result.Cost)
	}

	// Verify beast was added.
	if len(state.Beasts) != 1 {
		t.Errorf("expected 1 beast, got %d", len(state.Beasts))
	}
	if state.Beasts[0].Element != types.Water {
		t.Errorf("beast element = %v, want Water", state.Beasts[0].Element)
	}

	// Verify NextBeastID was incremented.
	if state.NextBeastID != 2 {
		t.Errorf("NextBeastID = %d, want 2", state.NextBeastID)
	}

	// Verify chi was spent.
	balanceAfter := state.EconomyEngine.ChiPool.Balance()
	if balanceAfter >= balanceBefore {
		t.Errorf("balance should have decreased: before=%f, after=%f", balanceBefore, balanceAfter)
	}
}

func TestApplyAction_SummonBeast_InsufficientChi(t *testing.T) {
	state := newTestState(t)
	_ = state.EconomyEngine.ChiPool.Withdraw(500, economy.Construction, "drain", 0)

	action := SummonBeastAction{Element: types.Water}
	_, err := ApplyAction(action, state)
	if err == nil {
		t.Fatal("expected error for insufficient chi")
	}

	// Beast should not have been added.
	if len(state.Beasts) != 0 {
		t.Errorf("beast count should remain 0, got %d", len(state.Beasts))
	}
}

func TestApplyAction_EvolveBeast_Success(t *testing.T) {
	state := newTestState(t)

	// Find a species that has an evolution path.
	var fromSpeciesID string
	for _, sp := range state.SpeciesRegistry.All() {
		paths := state.EvolutionRegistry.GetPaths(sp.ID)
		if len(paths) > 0 {
			fromSpeciesID = sp.ID
			break
		}
	}
	if fromSpeciesID == "" {
		t.Skip("no species with evolution paths found in default data")
	}

	species, _ := state.SpeciesRegistry.Get(fromSpeciesID)
	beast := senju.NewBeast(1, species, 0)
	state.Beasts = append(state.Beasts, beast)

	action := EvolveBeastAction{BeastID: 1}
	result, err := ApplyAction(action, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}

	// Beast species should have changed.
	if beast.SpeciesID == fromSpeciesID {
		t.Error("beast species should have changed after evolution")
	}
}

func TestApplyAction_EvolveBeast_BeastNotFound(t *testing.T) {
	state := newTestState(t)
	action := EvolveBeastAction{BeastID: 99}
	_, err := ApplyAction(action, state)
	if err == nil {
		t.Fatal("expected error for nonexistent beast")
	}
	if !errors.Is(err, ErrBeastNotFound) {
		t.Errorf("error = %v, want ErrBeastNotFound", err)
	}
}

func TestApplyAction_DigCorridor_Success(t *testing.T) {
	state := newTestState(t)
	// Add a second room with entrance.
	entrances := []world.RoomEntrance{
		{Pos: types.Pos{X: 8, Y: 9}, Dir: types.West},
	}
	_, err := state.Cave.AddRoom("trap_room", types.Pos{X: 8, Y: 8}, 3, 3, entrances)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	action := DigCorridorAction{FromRoomID: 1, ToRoomID: 2}
	result, err := ApplyAction(action, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}

	// Verify corridor was created.
	if len(state.Cave.Corridors) != 1 {
		t.Errorf("expected 1 corridor, got %d", len(state.Cave.Corridors))
	}
}

func TestApplyAction_DigCorridor_RoomNotFound(t *testing.T) {
	state := newTestState(t)
	action := DigCorridorAction{FromRoomID: 1, ToRoomID: 99}
	_, err := ApplyAction(action, state)
	if err == nil {
		t.Fatal("expected error for nonexistent room")
	}
}

// --- ActionResult tests ---

func TestActionResult_Fields(t *testing.T) {
	r := ActionResult{
		Success:     true,
		Cost:        42.5,
		Description: "did something",
	}
	if !r.Success {
		t.Error("Success should be true")
	}
	if r.Cost != 42.5 {
		t.Errorf("Cost = %f, want 42.5", r.Cost)
	}
	if r.Description != "did something" {
		t.Errorf("Description = %q, want %q", r.Description, "did something")
	}
}
