package human

import (
	"fmt"
	"io"
	"sort"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// defaultRoomWidth is the default width for newly dug rooms.
const defaultRoomWidth = 3

// defaultRoomHeight is the default height for newly dug rooms.
const defaultRoomHeight = 3

// RoomTypeOption represents an available room type for building.
type RoomTypeOption struct {
	// TypeID is the room type identifier.
	TypeID string
	// Name is the display name.
	Name string
	// Element is the five-element attribute.
	Element types.Element
	// Cost is the chi cost to build this room type.
	Cost float64
}

// RoomInfo represents an existing room for corridor/upgrade selection.
type RoomInfo struct {
	// ID is the room ID.
	ID int
	// TypeID is the room type identifier.
	TypeID string
	// Name is the display name.
	Name string
	// Pos is the room's top-left position.
	Pos types.Pos
}

// BuildContext holds the context needed by build submenus.
type BuildContext struct {
	// RoomTypes lists all available room types for building.
	RoomTypes []RoomTypeOption
	// Rooms lists all existing rooms in the cave.
	Rooms []RoomInfo
	// ChiBalance is the current chi pool balance.
	ChiBalance float64
	// CaveWidth is the cave grid width.
	CaveWidth int
	// CaveHeight is the cave grid height.
	CaveHeight int
}

// ShowDigRoomMenu displays the room digging submenu and returns a DigRoomAction.
// Returns nil action when the player chooses to go back.
func ShowDigRoomMenu(ir *InputReader, ctx BuildContext) (simulation.PlayerAction, error) {
	if len(ctx.RoomTypes) == 0 {
		fmt.Fprintln(ir.out, "建設可能な部屋タイプがありません。")
		return nil, nil
	}

	// Sort room types by TypeID for deterministic display order.
	sorted := make([]RoomTypeOption, len(ctx.RoomTypes))
	copy(sorted, ctx.RoomTypes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TypeID < sorted[j].TypeID
	})

	// Display room type choices.
	fmt.Fprintln(ir.out, "")
	fmt.Fprintln(ir.out, "=== 部屋を掘る ===")
	fmt.Fprintln(ir.out, "  0. 戻る")
	for i, rt := range sorted {
		costWarning := ""
		if rt.Cost > ctx.ChiBalance {
			costWarning = " [コスト不足!]"
		}
		fmt.Fprintf(ir.out, "  %d. %s (%s) - コスト: %.1f%s\n",
			i+1, rt.Name, rt.Element, rt.Cost, costWarning)
	}

	// Read room type choice.
	choice, err := ir.ReadIntInRange("部屋タイプ> ", 0, len(sorted))
	if err != nil {
		return nil, err
	}
	if choice == 0 {
		return nil, nil
	}

	selected := sorted[choice-1]

	// Warn if cost is insufficient but still allow proceeding.
	if selected.Cost > ctx.ChiBalance {
		fmt.Fprintf(ir.out, "警告: コスト不足です（必要: %.1f, 残高: %.1f）\n",
			selected.Cost, ctx.ChiBalance)
	}

	// Read coordinates.
	x, y, err := readRoomCoordinates(ir, ctx)
	if err != nil {
		return nil, err
	}

	return simulation.DigRoomAction{
		RoomTypeID: selected.TypeID,
		Pos:        types.Pos{X: x, Y: y},
		Width:      defaultRoomWidth,
		Height:     defaultRoomHeight,
	}, nil
}

// readRoomCoordinates reads X,Y coordinates for room placement.
func readRoomCoordinates(ir *InputReader, ctx BuildContext) (x, y int, err error) {
	maxX := ctx.CaveWidth - defaultRoomWidth
	maxY := ctx.CaveHeight - defaultRoomHeight
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}

	x, err = ir.ReadIntInRange(fmt.Sprintf("X座標 (0-%d)> ", maxX), 0, maxX)
	if err != nil {
		return 0, 0, err
	}
	y, err = ir.ReadIntInRange(fmt.Sprintf("Y座標 (0-%d)> ", maxY), 0, maxY)
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

// ShowDigCorridorMenu displays the corridor digging submenu and returns a DigCorridorAction.
// Returns nil action when the player chooses to go back.
func ShowDigCorridorMenu(ir *InputReader, ctx BuildContext) (simulation.PlayerAction, error) {
	if len(ctx.Rooms) < 2 {
		fmt.Fprintln(ir.out, "通路を掘るには2つ以上の部屋が必要です。")
		return nil, nil
	}

	// Display existing rooms.
	printRoomList(ir.out, ctx.Rooms)

	// Read from room ID.
	fmt.Fprintln(ir.out, "")
	fmt.Fprintln(ir.out, "=== 通路を掘る ===")
	fmt.Fprintln(ir.out, "  0 を入力すると戻ります。")

	fromID, err := ir.ReadInt("始点の部屋ID> ")
	if err != nil {
		return nil, err
	}
	if fromID == 0 {
		return nil, nil
	}

	if !hasRoom(ctx.Rooms, fromID) {
		fmt.Fprintf(ir.out, "部屋ID %d は存在しません。\n", fromID)
		return nil, nil
	}

	// Read to room ID.
	toID, err := ir.ReadInt("終点の部屋ID> ")
	if err != nil {
		return nil, err
	}
	if toID == 0 {
		return nil, nil
	}

	if !hasRoom(ctx.Rooms, toID) {
		fmt.Fprintf(ir.out, "部屋ID %d は存在しません。\n", toID)
		return nil, nil
	}

	if fromID == toID {
		fmt.Fprintln(ir.out, "始点と終点は異なる部屋を指定してください。")
		return nil, nil
	}

	return simulation.DigCorridorAction{
		FromRoomID: fromID,
		ToRoomID:   toID,
	}, nil
}

// printRoomList prints a formatted list of existing rooms.
func printRoomList(w io.Writer, rooms []RoomInfo) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "--- 既存の部屋 ---")
	for _, r := range rooms {
		fmt.Fprintf(w, "  ID:%d %s (%s) at (%d,%d)\n",
			r.ID, r.Name, r.TypeID, r.Pos.X, r.Pos.Y)
	}
}

// hasRoom checks whether the given ID exists in the room list.
func hasRoom(rooms []RoomInfo, id int) bool {
	for _, r := range rooms {
		if r.ID == id {
			return true
		}
	}
	return false
}
