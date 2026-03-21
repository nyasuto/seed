package human

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

func testUnitContext() UnitContext {
	return UnitContext{
		SummonOptions: []SummonOption{
			{Element: types.Wood, Cost: 30.0},
			{Element: types.Fire, Cost: 40.0},
			{Element: types.Water, Cost: 50.0},
		},
		UpgradeOptions: []UpgradeOption{
			{ID: 1, Name: "龍穴", TypeID: "dragon_hole", Level: 1, UpgradeCost: 60.0},
			{ID: 2, Name: "仙獣部屋", TypeID: "senju_room", Level: 1, UpgradeCost: 40.0},
		},
		Rooms: []RoomInfo{
			{ID: 1, TypeID: "dragon_hole", Name: "龍穴", Pos: types.Pos{X: 5, Y: 5}},
			{ID: 2, TypeID: "senju_room", Name: "仙獣部屋", Pos: types.Pos{X: 10, Y: 5}},
		},
		ChiBalance: 100.0,
	}
}

func TestShowSummonBeastMenu_SelectElement(t *testing.T) {
	// Select element 2 (Fire).
	input := "2\n"
	ir, _ := newTestIR(input)
	ctx := testUnitContext()

	action, err := ShowSummonBeastMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action == nil {
		t.Fatal("expected action, got nil")
	}

	summon, ok := action.(simulation.SummonBeastAction)
	if !ok {
		t.Fatalf("expected SummonBeastAction, got %T", action)
	}
	if summon.Element != types.Fire {
		t.Errorf("expected element Fire, got %v", summon.Element)
	}
}

func TestShowSummonBeastMenu_Back(t *testing.T) {
	input := "0\n"
	ir, _ := newTestIR(input)
	ctx := testUnitContext()

	action, err := ShowSummonBeastMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatalf("expected nil action for back, got %v", action)
	}
}

func TestShowSummonBeastMenu_CostWarning(t *testing.T) {
	// Select Water (cost 50.0) with chi balance 35.0.
	input := "3\n"
	ir, out := newTestIR(input)
	ctx := testUnitContext()
	ctx.ChiBalance = 35.0

	action, err := ShowSummonBeastMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action == nil {
		t.Fatal("expected action even with cost warning")
	}

	output := out.String()
	if !strings.Contains(output, "コスト不足") {
		t.Error("expected cost warning in output")
	}
}

func TestShowSummonBeastMenu_CostWarningInList(t *testing.T) {
	// Chi balance 35.0: Fire(40) and Water(50) should show warning in list.
	input := "0\n"
	ir, out := newTestIR(input)
	ctx := testUnitContext()
	ctx.ChiBalance = 35.0

	_, _ = ShowSummonBeastMenu(ir, ctx)

	output := out.String()
	count := strings.Count(output, "[コスト不足!]")
	if count != 2 {
		t.Errorf("expected 2 cost warnings in list, got %d\noutput:\n%s", count, output)
	}
}

func TestShowSummonBeastMenu_NoOptions(t *testing.T) {
	input := ""
	ir, out := newTestIR(input)
	ctx := testUnitContext()
	ctx.SummonOptions = nil

	action, err := ShowSummonBeastMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatal("expected nil action when no options")
	}
	if !strings.Contains(out.String(), "召喚可能な属性がありません") {
		t.Error("expected no-options message")
	}
}

func TestShowUpgradeRoomMenu_SelectRoom(t *testing.T) {
	// Select room 2 (仙獣部屋).
	input := "2\n"
	ir, _ := newTestIR(input)
	ctx := testUnitContext()

	action, err := ShowUpgradeRoomMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action == nil {
		t.Fatal("expected action, got nil")
	}

	upgrade, ok := action.(simulation.UpgradeRoomAction)
	if !ok {
		t.Fatalf("expected UpgradeRoomAction, got %T", action)
	}
	if upgrade.RoomID != 2 {
		t.Errorf("expected room ID 2, got %d", upgrade.RoomID)
	}
}

func TestShowUpgradeRoomMenu_Back(t *testing.T) {
	input := "0\n"
	ir, _ := newTestIR(input)
	ctx := testUnitContext()

	action, err := ShowUpgradeRoomMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatalf("expected nil action for back, got %v", action)
	}
}

func TestShowUpgradeRoomMenu_CostWarning(t *testing.T) {
	// Select 龍穴 (cost 60.0) with chi balance 50.0.
	input := "1\n"
	ir, out := newTestIR(input)
	ctx := testUnitContext()
	ctx.ChiBalance = 50.0

	action, err := ShowUpgradeRoomMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action == nil {
		t.Fatal("expected action even with cost warning")
	}

	output := out.String()
	if !strings.Contains(output, "コスト不足") {
		t.Error("expected cost warning in output")
	}
}

func TestShowUpgradeRoomMenu_CostWarningInList(t *testing.T) {
	// Chi balance 30.0: both rooms (60, 40) should show warning.
	input := "0\n"
	ir, out := newTestIR(input)
	ctx := testUnitContext()
	ctx.ChiBalance = 30.0

	_, _ = ShowUpgradeRoomMenu(ir, ctx)

	output := out.String()
	count := strings.Count(output, "[コスト不足!]")
	if count != 2 {
		t.Errorf("expected 2 cost warnings in list, got %d\noutput:\n%s", count, output)
	}
}

func TestShowUpgradeRoomMenu_NoOptions(t *testing.T) {
	input := ""
	ir, out := newTestIR(input)
	ctx := testUnitContext()
	ctx.UpgradeOptions = nil

	action, err := ShowUpgradeRoomMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatal("expected nil action when no options")
	}
	if !strings.Contains(out.String(), "アップグレード可能な部屋がありません") {
		t.Error("expected no-options message")
	}
}
