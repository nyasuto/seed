package human

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

func testBuildContext() BuildContext {
	return BuildContext{
		RoomTypes: []RoomTypeOption{
			{TypeID: "chi_chamber", Name: "蓄気室", Element: types.Water, Cost: 50.0},
			{TypeID: "senju_room", Name: "仙獣部屋", Element: types.Wood, Cost: 30.0},
			{TypeID: "trap_room", Name: "罠部屋", Element: types.Metal, Cost: 40.0},
		},
		Rooms: []RoomInfo{
			{ID: 1, TypeID: "dragon_hole", Name: "龍穴", Pos: types.Pos{X: 5, Y: 5}},
			{ID: 2, TypeID: "senju_room", Name: "仙獣部屋", Pos: types.Pos{X: 10, Y: 5}},
		},
		ChiBalance: 100.0,
		CaveWidth:  30,
		CaveHeight: 30,
	}
}

func TestShowDigRoomMenu_SelectRoomAndCoordinates(t *testing.T) {
	// Select room type 1 (sorted by TypeID: chi_chamber), then coordinates (4, 6).
	input := "1\n4\n6\n"
	ir, out := newTestIR(input)
	ctx := testBuildContext()

	action, err := ShowDigRoomMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action == nil {
		t.Fatal("expected action, got nil")
	}

	dig, ok := action.(simulation.DigRoomAction)
	if !ok {
		t.Fatalf("expected DigRoomAction, got %T", action)
	}

	// Sorted order: chi_chamber, senju_room, trap_room
	if dig.RoomTypeID != "chi_chamber" {
		t.Errorf("expected room type chi_chamber, got %s", dig.RoomTypeID)
	}
	if dig.Pos.X != 4 || dig.Pos.Y != 6 {
		t.Errorf("expected pos (4,6), got (%d,%d)", dig.Pos.X, dig.Pos.Y)
	}
	if dig.Width != defaultRoomWidth || dig.Height != defaultRoomHeight {
		t.Errorf("expected size %dx%d, got %dx%d", defaultRoomWidth, defaultRoomHeight, dig.Width, dig.Height)
	}

	_ = out
}

func TestShowDigRoomMenu_Back(t *testing.T) {
	input := "0\n"
	ir, _ := newTestIR(input)
	ctx := testBuildContext()

	action, err := ShowDigRoomMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatalf("expected nil action for back, got %v", action)
	}
}

func TestShowDigRoomMenu_CostWarning(t *testing.T) {
	// Select trap_room (cost 40.0) with chi balance 20.0.
	input := "3\n2\n2\n"
	ir, out := newTestIR(input)
	ctx := testBuildContext()
	ctx.ChiBalance = 20.0

	action, err := ShowDigRoomMenu(ir, ctx)
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

func TestShowDigRoomMenu_CostWarningInList(t *testing.T) {
	// Chi balance is 35.0: chi_chamber(50) and trap_room(40) should show warning.
	input := "0\n"
	ir, out := newTestIR(input)
	ctx := testBuildContext()
	ctx.ChiBalance = 35.0

	_, _ = ShowDigRoomMenu(ir, ctx)

	output := out.String()
	// Count "[コスト不足!]" occurrences — chi_chamber(50) and trap_room(40) exceed 35.
	count := strings.Count(output, "[コスト不足!]")
	if count != 2 {
		t.Errorf("expected 2 cost warnings in list, got %d\noutput:\n%s", count, output)
	}
}

func TestShowDigRoomMenu_NoRoomTypes(t *testing.T) {
	input := ""
	ir, out := newTestIR(input)
	ctx := testBuildContext()
	ctx.RoomTypes = nil

	action, err := ShowDigRoomMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatal("expected nil action when no room types")
	}
	if !strings.Contains(out.String(), "建設可能な部屋タイプがありません") {
		t.Error("expected no-room-types message")
	}
}

func TestShowDigCorridorMenu_SelectRooms(t *testing.T) {
	input := "1\n2\n"
	ir, _ := newTestIR(input)
	ctx := testBuildContext()

	action, err := ShowDigCorridorMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action == nil {
		t.Fatal("expected action, got nil")
	}

	dig, ok := action.(simulation.DigCorridorAction)
	if !ok {
		t.Fatalf("expected DigCorridorAction, got %T", action)
	}
	if dig.FromRoomID != 1 || dig.ToRoomID != 2 {
		t.Errorf("expected rooms 1->2, got %d->%d", dig.FromRoomID, dig.ToRoomID)
	}
}

func TestShowDigCorridorMenu_Back(t *testing.T) {
	input := "0\n"
	ir, _ := newTestIR(input)
	ctx := testBuildContext()

	action, err := ShowDigCorridorMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatalf("expected nil action for back, got %v", action)
	}
}

func TestShowDigCorridorMenu_BackFromToID(t *testing.T) {
	input := "1\n0\n"
	ir, _ := newTestIR(input)
	ctx := testBuildContext()

	action, err := ShowDigCorridorMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatalf("expected nil action for back at toID, got %v", action)
	}
}

func TestShowDigCorridorMenu_InvalidFromRoom(t *testing.T) {
	input := "99\n"
	ir, out := newTestIR(input)
	ctx := testBuildContext()

	action, err := ShowDigCorridorMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatal("expected nil action for invalid room")
	}
	if !strings.Contains(out.String(), "存在しません") {
		t.Error("expected invalid room message")
	}
}

func TestShowDigCorridorMenu_InvalidToRoom(t *testing.T) {
	input := "1\n99\n"
	ir, out := newTestIR(input)
	ctx := testBuildContext()

	action, err := ShowDigCorridorMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatal("expected nil action for invalid to room")
	}
	if !strings.Contains(out.String(), "存在しません") {
		t.Error("expected invalid room message")
	}
}

func TestShowDigCorridorMenu_SameRoom(t *testing.T) {
	input := "1\n1\n"
	ir, out := newTestIR(input)
	ctx := testBuildContext()

	action, err := ShowDigCorridorMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatal("expected nil action for same room")
	}
	if !strings.Contains(out.String(), "異なる部屋") {
		t.Error("expected same-room error message")
	}
}

func TestShowDigCorridorMenu_TooFewRooms(t *testing.T) {
	input := ""
	ir, out := newTestIR(input)
	ctx := testBuildContext()
	ctx.Rooms = ctx.Rooms[:1] // Only one room.

	action, err := ShowDigCorridorMenu(ir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != nil {
		t.Fatal("expected nil action when too few rooms")
	}
	if !strings.Contains(out.String(), "2つ以上の部屋が必要") {
		t.Error("expected too-few-rooms message")
	}
}

func newTestIR(input string) (*InputReader, *bytes.Buffer) {
	out := &bytes.Buffer{}
	ir := NewInputReader(strings.NewReader(input), out)
	return ir, out
}
