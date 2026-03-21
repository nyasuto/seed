package scenario

import (
	"testing"

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
	for dy := 0; dy < 3; dy++ {
		for dx := 0; dx < 2; dx++ {
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
