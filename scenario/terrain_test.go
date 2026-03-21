package scenario

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/testutil"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

func TestApplyTerrain_Basic(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	zones := []TerrainZone{
		{Pos: types.Pos{X: 1, Y: 1}, Width: 2, Height: 3, Type: world.HardRock},
		{Pos: types.Pos{X: 5, Y: 5}, Width: 1, Height: 1, Type: world.Water},
	}

	if err := ApplyTerrain(cave, zones); err != nil {
		t.Fatalf("ApplyTerrain: %v", err)
	}

	// Verify HardRock zone
	for dy := range 3 {
		for dx := range 2 {
			cell, _ := cave.Grid.At(types.Pos{X: 1 + dx, Y: 1 + dy})
			if cell.Type != world.HardRock {
				t.Errorf("pos (%d,%d): got %v, want HardRock", 1+dx, 1+dy, cell.Type)
			}
		}
	}

	// Verify Water zone
	cell, _ := cave.Grid.At(types.Pos{X: 5, Y: 5})
	if cell.Type != world.Water {
		t.Errorf("pos (5,5): got %v, want Water", cell.Type)
	}

	// Verify unaffected cell is still Rock
	cell, _ = cave.Grid.At(types.Pos{X: 0, Y: 0})
	if cell.Type != world.Rock {
		t.Errorf("pos (0,0): got %v, want Rock", cell.Type)
	}
}

func TestApplyTerrain_OutOfBounds(t *testing.T) {
	cave, err := world.NewCave(5, 5)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	zones := []TerrainZone{
		{Pos: types.Pos{X: 4, Y: 4}, Width: 2, Height: 2, Type: world.HardRock},
	}

	if err := ApplyTerrain(cave, zones); err == nil {
		t.Error("expected error for out-of-bounds zone, got nil")
	}
}

func TestApplyTerrain_OverlapsRoom(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Simulate a room by setting a cell's RoomID
	pos := types.Pos{X: 3, Y: 3}
	if err := cave.Grid.Set(pos, world.Cell{Type: world.RoomFloor, RoomID: 1}); err != nil {
		t.Fatalf("Grid.Set: %v", err)
	}

	zones := []TerrainZone{
		{Pos: types.Pos{X: 2, Y: 2}, Width: 3, Height: 3, Type: world.HardRock},
	}

	if err := ApplyTerrain(cave, zones); err == nil {
		t.Error("expected error for zone overlapping room, got nil")
	}
}

func TestApplyTerrain_EmptyZones(t *testing.T) {
	cave, err := world.NewCave(5, 5)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	if err := ApplyTerrain(cave, nil); err != nil {
		t.Errorf("ApplyTerrain with nil zones: %v", err)
	}

	if err := ApplyTerrain(cave, []TerrainZone{}); err != nil {
		t.Errorf("ApplyTerrain with empty zones: %v", err)
	}
}

func TestGenerateTerrain_Deterministic(t *testing.T) {
	tg := &TerrainGenerator{}

	rng1 := testutil.NewTestRNG(42)
	zones1 := tg.GenerateTerrain(16, 16, 0.1, rng1)

	rng2 := testutil.NewTestRNG(42)
	zones2 := tg.GenerateTerrain(16, 16, 0.1, rng2)

	if len(zones1) != len(zones2) {
		t.Fatalf("deterministic mismatch: %d zones vs %d zones", len(zones1), len(zones2))
	}

	for i := range zones1 {
		if zones1[i] != zones2[i] {
			t.Errorf("zone %d: %+v != %+v", i, zones1[i], zones2[i])
		}
	}
}

func TestValidateTerrain_NoRooms(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	if err := ValidateTerrain(cave, nil); err != nil {
		t.Errorf("ValidateTerrain with no rooms: %v", err)
	}
}

func TestValidateTerrain_RoomOnImpassable(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Place HardRock at (3,3)
	zones := []TerrainZone{
		{Pos: types.Pos{X: 3, Y: 3}, Width: 1, Height: 1, Type: world.HardRock},
	}
	if err := ApplyTerrain(cave, zones); err != nil {
		t.Fatalf("ApplyTerrain: %v", err)
	}

	rooms := []RoomPlacement{
		{TypeID: "dragon_hole", Pos: types.Pos{X: 3, Y: 3}},
	}

	err = ValidateTerrain(cave, rooms)
	if err == nil {
		t.Fatal("expected error for room on impassable terrain, got nil")
	}
	if !errors.Is(err, ErrTerrainBlocksRoom) {
		t.Errorf("expected ErrTerrainBlocksRoom, got: %v", err)
	}
}

func TestValidateTerrain_RoomOnWater(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	zones := []TerrainZone{
		{Pos: types.Pos{X: 5, Y: 5}, Width: 1, Height: 1, Type: world.Water},
	}
	if err := ApplyTerrain(cave, zones); err != nil {
		t.Fatalf("ApplyTerrain: %v", err)
	}

	rooms := []RoomPlacement{
		{TypeID: "wood_room", Pos: types.Pos{X: 5, Y: 5}},
	}

	err = ValidateTerrain(cave, rooms)
	if err == nil {
		t.Fatal("expected error for room on Water terrain, got nil")
	}
	if !errors.Is(err, ErrTerrainBlocksRoom) {
		t.Errorf("expected ErrTerrainBlocksRoom, got: %v", err)
	}
}

func TestValidateTerrain_Disconnected(t *testing.T) {
	// Create a 10x10 cave with a wall of HardRock splitting it vertically.
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Wall at x=5, full height
	zones := []TerrainZone{
		{Pos: types.Pos{X: 5, Y: 0}, Width: 1, Height: 10, Type: world.HardRock},
	}
	if err := ApplyTerrain(cave, zones); err != nil {
		t.Fatalf("ApplyTerrain: %v", err)
	}

	rooms := []RoomPlacement{
		{TypeID: "dragon_hole", Pos: types.Pos{X: 1, Y: 1}},
		{TypeID: "wood_room", Pos: types.Pos{X: 8, Y: 8}},
	}

	err = ValidateTerrain(cave, rooms)
	if err == nil {
		t.Fatal("expected error for disconnected rooms, got nil")
	}
	if !errors.Is(err, ErrTerrainDisconnected) {
		t.Errorf("expected ErrTerrainDisconnected, got: %v", err)
	}
}

func TestValidateTerrain_Connected(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Partial wall (leaves a gap at y=9)
	zones := []TerrainZone{
		{Pos: types.Pos{X: 5, Y: 0}, Width: 1, Height: 9, Type: world.HardRock},
	}
	if err := ApplyTerrain(cave, zones); err != nil {
		t.Fatalf("ApplyTerrain: %v", err)
	}

	rooms := []RoomPlacement{
		{TypeID: "dragon_hole", Pos: types.Pos{X: 1, Y: 1}},
		{TypeID: "wood_room", Pos: types.Pos{X: 8, Y: 8}},
	}

	if err := ValidateTerrain(cave, rooms); err != nil {
		t.Errorf("expected no error for connected rooms, got: %v", err)
	}
}

func TestValidateTerrain_OutOfBounds(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	rooms := []RoomPlacement{
		{TypeID: "dragon_hole", Pos: types.Pos{X: 15, Y: 15}},
	}

	if err := ValidateTerrain(cave, rooms); err == nil {
		t.Error("expected error for out-of-bounds room, got nil")
	}
}

func TestValidateTerrain_SingleRoom(t *testing.T) {
	cave, err := world.NewCave(10, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	rooms := []RoomPlacement{
		{TypeID: "dragon_hole", Pos: types.Pos{X: 5, Y: 5}},
	}

	if err := ValidateTerrain(cave, rooms); err != nil {
		t.Errorf("single room on clear terrain should pass: %v", err)
	}
}
