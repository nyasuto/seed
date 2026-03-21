package economy

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/world"
)

func TestCalcTickMaintenance_NoRooms(t *testing.T) {
	mc := NewMaintenanceCalculator(mustLoadCostParams(t))
	result := mc.CalcTickMaintenance(nil, 0, 0)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %f", result.Total)
	}
	if result.RoomCost != 0 {
		t.Errorf("expected room cost 0, got %f", result.RoomCost)
	}
	if result.BeastCost != 0 {
		t.Errorf("expected beast cost 0, got %f", result.BeastCost)
	}
	if result.TrapCost != 0 {
		t.Errorf("expected trap cost 0, got %f", result.TrapCost)
	}
}

func TestCalcTickMaintenance_RoomTypeCosts(t *testing.T) {
	params := mustLoadCostParams(t)
	mc := NewMaintenanceCalculator(params)

	rooms := []world.Room{
		{TypeID: "dragon_hole"},
		{TypeID: "chi_storage"},
		{TypeID: "beast_room"},
	}
	result := mc.CalcTickMaintenance(rooms, 0, 0)

	expected := params.RoomMaintenancePerTick["dragon_hole"] +
		params.RoomMaintenancePerTick["chi_storage"] +
		params.RoomMaintenancePerTick["beast_room"]

	if result.RoomCost != expected {
		t.Errorf("expected room cost %f, got %f", expected, result.RoomCost)
	}
	if result.Total != expected {
		t.Errorf("expected total %f, got %f", expected, result.Total)
	}
}

func TestCalcTickMaintenance_BeastsIncreaseCost(t *testing.T) {
	params := mustLoadCostParams(t)
	mc := NewMaintenanceCalculator(params)

	rooms := []world.Room{{TypeID: "beast_room"}}
	result := mc.CalcTickMaintenance(rooms, 3, 0)

	expectedBeast := 3 * params.BeastMaintenancePerTick
	if result.BeastCost != expectedBeast {
		t.Errorf("expected beast cost %f, got %f", expectedBeast, result.BeastCost)
	}

	expectedTotal := params.RoomMaintenancePerTick["beast_room"] + expectedBeast
	if result.Total != expectedTotal {
		t.Errorf("expected total %f, got %f", expectedTotal, result.Total)
	}
}

func TestCalcTickMaintenance_Traps(t *testing.T) {
	params := mustLoadCostParams(t)
	mc := NewMaintenanceCalculator(params)

	rooms := []world.Room{{TypeID: "trap_room"}}
	result := mc.CalcTickMaintenance(rooms, 0, 2)

	expectedTrap := 2 * params.TrapMaintenancePerTick
	if result.TrapCost != expectedTrap {
		t.Errorf("expected trap cost %f, got %f", expectedTrap, result.TrapCost)
	}
}

func TestCalcTickMaintenance_AllCategories(t *testing.T) {
	params := mustLoadCostParams(t)
	mc := NewMaintenanceCalculator(params)

	rooms := []world.Room{
		{TypeID: "dragon_hole"},
		{TypeID: "trap_room"},
	}
	result := mc.CalcTickMaintenance(rooms, 2, 1)

	expectedRoom := params.RoomMaintenancePerTick["dragon_hole"] +
		params.RoomMaintenancePerTick["trap_room"]
	expectedBeast := 2 * params.BeastMaintenancePerTick
	expectedTrap := 1 * params.TrapMaintenancePerTick
	expectedTotal := expectedRoom + expectedBeast + expectedTrap

	if result.RoomCost != expectedRoom {
		t.Errorf("expected room cost %f, got %f", expectedRoom, result.RoomCost)
	}
	if result.BeastCost != expectedBeast {
		t.Errorf("expected beast cost %f, got %f", expectedBeast, result.BeastCost)
	}
	if result.TrapCost != expectedTrap {
		t.Errorf("expected trap cost %f, got %f", expectedTrap, result.TrapCost)
	}
	if result.Total != expectedTotal {
		t.Errorf("expected total %f, got %f", expectedTotal, result.Total)
	}
}

func TestCalcChiPoolCap_NoStorageRooms(t *testing.T) {
	params := mustLoadCostParams(t)
	rooms := []world.Room{
		{TypeID: "dragon_hole"},
		{TypeID: "beast_room"},
	}
	cap := CalcChiPoolCap(rooms, params)
	if cap != params.ChiPoolBaseCap {
		t.Errorf("expected base cap %f, got %f", params.ChiPoolBaseCap, cap)
	}
}

func TestCalcChiPoolCap_WithStorageRooms(t *testing.T) {
	params := mustLoadCostParams(t)
	rooms := []world.Room{
		{TypeID: "chi_storage", Level: 1},
		{TypeID: "chi_storage", Level: 1},
		{TypeID: "dragon_hole", Level: 1},
	}
	expected := params.ChiPoolBaseCap + 2*params.ChiPoolCapPerStorageRoom
	cap := CalcChiPoolCap(rooms, params)
	if cap != expected {
		t.Errorf("expected cap %f, got %f", expected, cap)
	}
}

func TestCalcChiPoolCap_StorageLevelUp(t *testing.T) {
	params := mustLoadCostParams(t)
	rooms := []world.Room{
		{TypeID: "chi_storage", Level: 3},
	}
	// base + 1 room bonus + (3-1) level bonuses
	expected := params.ChiPoolBaseCap +
		params.ChiPoolCapPerStorageRoom +
		2*params.ChiPoolCapPerStorageLevel
	cap := CalcChiPoolCap(rooms, params)
	if cap != expected {
		t.Errorf("expected cap %f, got %f", expected, cap)
	}
}

func TestCalcChiPoolCap_EmptyRooms(t *testing.T) {
	params := mustLoadCostParams(t)
	cap := CalcChiPoolCap(nil, params)
	if cap != params.ChiPoolBaseCap {
		t.Errorf("expected base cap %f, got %f", params.ChiPoolBaseCap, cap)
	}
}

func TestLoadCostParams(t *testing.T) {
	p := mustLoadCostParams(t)
	if p == nil {
		t.Fatal("DefaultCostParams returned nil")
	}
	if p.ChiPoolBaseCap != 100.0 {
		t.Errorf("expected base cap 100.0, got %f", p.ChiPoolBaseCap)
	}
	if p.BeastMaintenancePerTick != 0.3 {
		t.Errorf("expected beast maintenance 0.3, got %f", p.BeastMaintenancePerTick)
	}
	if len(p.RoomMaintenancePerTick) != 6 {
		t.Errorf("expected 6 room types, got %d", len(p.RoomMaintenancePerTick))
	}
}

func TestLoadCostParams_InvalidJSON(t *testing.T) {
	_, err := LoadCostParams([]byte("invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCalcTickMaintenance_UnknownRoomType(t *testing.T) {
	params := mustLoadCostParams(t)
	mc := NewMaintenanceCalculator(params)

	rooms := []world.Room{{TypeID: "unknown_type"}}
	result := mc.CalcTickMaintenance(rooms, 0, 0)
	if result.RoomCost != 0 {
		t.Errorf("expected room cost 0 for unknown type, got %f", result.RoomCost)
	}
}
