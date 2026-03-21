package senju

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// helper to create a test species.
func testSpecies(element types.Element) *Species {
	return &Species{
		ID:         "test_beast",
		Name:       "TestBeast",
		Element:    element,
		BaseHP:     100,
		BaseATK:    20,
		BaseDEF:    15,
		BaseSPD:    10,
		GrowthRate: 1.0,
		MaxBeasts:  3,
	}
}

// helper to create a room with the given ID.
func testRoom(id int) *world.Room {
	return &world.Room{
		ID:     id,
		TypeID: "beast_room",
		Pos:    types.Pos{X: 0, Y: 0},
		Width:  3,
		Height: 3,
		Level:  1,
	}
}

func testRoomType(maxBeasts int) world.RoomType {
	return world.RoomType{
		ID:              "beast_room",
		Name:            "仙獣部屋",
		Element:         types.Wood,
		BaseChiCapacity: 100,
		MaxBeasts:       maxBeasts,
	}
}

func TestNewBeast(t *testing.T) {
	sp := testSpecies(types.Fire)
	b := NewBeast(1, sp, 100)

	if b.ID != 1 {
		t.Errorf("ID = %d, want 1", b.ID)
	}
	if b.SpeciesID != "test_beast" {
		t.Errorf("SpeciesID = %q, want %q", b.SpeciesID, "test_beast")
	}
	if b.Element != types.Fire {
		t.Errorf("Element = %v, want Fire", b.Element)
	}
	if b.Level != 1 {
		t.Errorf("Level = %d, want 1", b.Level)
	}
	if b.HP != 100 || b.MaxHP != 100 {
		t.Errorf("HP=%d MaxHP=%d, want 100/100", b.HP, b.MaxHP)
	}
	if b.ATK != 20 {
		t.Errorf("ATK = %d, want 20", b.ATK)
	}
	if b.DEF != 15 {
		t.Errorf("DEF = %d, want 15", b.DEF)
	}
	if b.SPD != 10 {
		t.Errorf("SPD = %d, want 10", b.SPD)
	}
	if b.BornTick != 100 {
		t.Errorf("BornTick = %d, want 100", b.BornTick)
	}
	if b.State != Idle {
		t.Errorf("State = %v, want Idle", b.State)
	}
	if b.RoomID != 0 {
		t.Errorf("RoomID = %d, want 0", b.RoomID)
	}
}

func TestPlaceBeast_Success(t *testing.T) {
	sp := testSpecies(types.Wood)
	b := NewBeast(1, sp, 0)
	room := testRoom(10)
	rt := testRoomType(3)

	if err := PlaceBeast(b, room, rt); err != nil {
		t.Fatalf("PlaceBeast: unexpected error: %v", err)
	}
	if b.RoomID != 10 {
		t.Errorf("beast.RoomID = %d, want 10", b.RoomID)
	}
	if room.BeastCount() != 1 {
		t.Errorf("room.BeastCount() = %d, want 1", room.BeastCount())
	}
	if room.BeastIDs[0] != 1 {
		t.Errorf("room.BeastIDs[0] = %d, want 1", room.BeastIDs[0])
	}
}

func TestPlaceBeast_CapacityExceeded(t *testing.T) {
	sp := testSpecies(types.Wood)
	room := testRoom(10)
	rt := testRoomType(2)

	// Fill to capacity.
	b1 := NewBeast(1, sp, 0)
	b2 := NewBeast(2, sp, 0)
	if err := PlaceBeast(b1, room, rt); err != nil {
		t.Fatalf("PlaceBeast b1: %v", err)
	}
	if err := PlaceBeast(b2, room, rt); err != nil {
		t.Fatalf("PlaceBeast b2: %v", err)
	}

	// Third placement should fail.
	b3 := NewBeast(3, sp, 0)
	err := PlaceBeast(b3, room, rt)
	if err == nil {
		t.Fatal("expected error for capacity exceeded, got nil")
	}
	if b3.RoomID != 0 {
		t.Errorf("beast.RoomID = %d, want 0 after failed placement", b3.RoomID)
	}
}

func TestPlaceBeast_RoomTypeNotAllowed(t *testing.T) {
	sp := testSpecies(types.Wood)
	b := NewBeast(1, sp, 0)
	room := testRoom(10)
	rt := testRoomType(0) // MaxBeasts=0 means beasts not allowed

	err := PlaceBeast(b, room, rt)
	if err == nil {
		t.Fatal("expected error for room type not allowing beasts, got nil")
	}
	if b.RoomID != 0 {
		t.Errorf("beast.RoomID = %d, want 0", b.RoomID)
	}
}

func TestPlaceBeast_AlreadyPlaced(t *testing.T) {
	sp := testSpecies(types.Wood)
	b := NewBeast(1, sp, 0)
	room := testRoom(10)
	rt := testRoomType(3)

	if err := PlaceBeast(b, room, rt); err != nil {
		t.Fatalf("PlaceBeast: %v", err)
	}

	room2 := testRoom(20)
	err := PlaceBeast(b, room2, rt)
	if err == nil {
		t.Fatal("expected error for beast already placed, got nil")
	}
}

func TestRemoveBeast(t *testing.T) {
	sp := testSpecies(types.Wood)
	b := NewBeast(1, sp, 0)
	room := testRoom(10)
	rt := testRoomType(3)

	if err := PlaceBeast(b, room, rt); err != nil {
		t.Fatalf("PlaceBeast: %v", err)
	}
	if err := RemoveBeast(b, room); err != nil {
		t.Fatalf("RemoveBeast: %v", err)
	}
	if b.RoomID != 0 {
		t.Errorf("beast.RoomID = %d, want 0", b.RoomID)
	}
	if room.BeastCount() != 0 {
		t.Errorf("room.BeastCount() = %d, want 0", room.BeastCount())
	}
}

func TestRemoveBeast_WrongRoom(t *testing.T) {
	sp := testSpecies(types.Wood)
	b := NewBeast(1, sp, 0)
	b.RoomID = 10
	room := testRoom(20) // different room

	err := RemoveBeast(b, room)
	if err == nil {
		t.Fatal("expected error for beast not in room, got nil")
	}
}

func TestMoveBeast(t *testing.T) {
	sp := testSpecies(types.Wood)
	b := NewBeast(1, sp, 0)
	from := testRoom(10)
	to := testRoom(20)
	rt := testRoomType(3)

	if err := PlaceBeast(b, from, rt); err != nil {
		t.Fatalf("PlaceBeast: %v", err)
	}
	if err := MoveBeast(b, from, to, rt); err != nil {
		t.Fatalf("MoveBeast: %v", err)
	}
	if b.RoomID != 20 {
		t.Errorf("beast.RoomID = %d, want 20", b.RoomID)
	}
	if from.BeastCount() != 0 {
		t.Errorf("from.BeastCount() = %d, want 0", from.BeastCount())
	}
	if to.BeastCount() != 1 {
		t.Errorf("to.BeastCount() = %d, want 1", to.BeastCount())
	}
}

func TestMoveBeast_TargetFull(t *testing.T) {
	sp := testSpecies(types.Wood)
	from := testRoom(10)
	to := testRoom(20)
	rt := testRoomType(1)

	// Place a beast in 'to' to fill it.
	occupant := NewBeast(99, sp, 0)
	if err := PlaceBeast(occupant, to, rt); err != nil {
		t.Fatalf("PlaceBeast occupant: %v", err)
	}

	// Place our beast in 'from'.
	b := NewBeast(1, sp, 0)
	if err := PlaceBeast(b, from, rt); err != nil {
		t.Fatalf("PlaceBeast: %v", err)
	}

	// Move should fail and rollback.
	err := MoveBeast(b, from, to, rt)
	if err == nil {
		t.Fatal("expected error for full target room, got nil")
	}
	// Beast should be back in from room after rollback.
	if b.RoomID != 10 {
		t.Errorf("beast.RoomID = %d, want 10 (rollback)", b.RoomID)
	}
	if from.BeastCount() != 1 {
		t.Errorf("from.BeastCount() = %d, want 1 (rollback)", from.BeastCount())
	}
}

func TestCalcCombatStats_NilRoomChi(t *testing.T) {
	sp := testSpecies(types.Fire)
	b := NewBeast(1, sp, 0)

	stats := b.CalcCombatStats(nil)
	if stats.ATK != b.ATK || stats.DEF != b.DEF || stats.SPD != b.SPD {
		t.Errorf("nil roomChi: stats = %+v, want base stats ATK=%d DEF=%d SPD=%d",
			stats, b.ATK, b.DEF, b.SPD)
	}
}

func TestCalcCombatStats_Generates(t *testing.T) {
	// Wood generates Fire. Room=Wood, Beast=Fire → room generates beast → 1.3x
	sp := testSpecies(types.Fire)
	b := NewBeast(1, sp, 0)
	roomChi := &fengshui.RoomChi{RoomID: 1, Current: 50, Capacity: 100, Element: types.Wood}

	stats := b.CalcCombatStats(roomChi)
	expectedATK := 26 // 20 * 1.3 = 26
	expectedDEF := 20 // 15 * 1.3 = 19.5 → 20 (rounded)
	expectedSPD := 13 // 10 * 1.3 = 13
	if stats.ATK != expectedATK {
		t.Errorf("generates: ATK = %d, want %d", stats.ATK, expectedATK)
	}
	if stats.DEF != expectedDEF {
		t.Errorf("generates: DEF = %d, want %d", stats.DEF, expectedDEF)
	}
	if stats.SPD != expectedSPD {
		t.Errorf("generates: SPD = %d, want %d", stats.SPD, expectedSPD)
	}
}

func TestCalcCombatStats_SameElement(t *testing.T) {
	sp := testSpecies(types.Fire)
	b := NewBeast(1, sp, 0)
	roomChi := &fengshui.RoomChi{RoomID: 1, Current: 50, Capacity: 100, Element: types.Fire}

	stats := b.CalcCombatStats(roomChi)
	expectedATK := 22 // 20 * 1.1 = 22
	expectedDEF := 17 // 15 * 1.1 = 16.5 → 17 (rounded)
	expectedSPD := 11 // 10 * 1.1 = 11
	if stats.ATK != expectedATK {
		t.Errorf("same: ATK = %d, want %d", stats.ATK, expectedATK)
	}
	if stats.DEF != expectedDEF {
		t.Errorf("same: DEF = %d, want %d", stats.DEF, expectedDEF)
	}
	if stats.SPD != expectedSPD {
		t.Errorf("same: SPD = %d, want %d", stats.SPD, expectedSPD)
	}
}

func TestCalcCombatStats_Overcomes(t *testing.T) {
	// Water overcomes Fire. Room=Water, Beast=Fire → room overcomes beast → 0.7x
	sp := testSpecies(types.Fire)
	b := NewBeast(1, sp, 0)
	roomChi := &fengshui.RoomChi{RoomID: 1, Current: 50, Capacity: 100, Element: types.Water}

	stats := b.CalcCombatStats(roomChi)
	expectedATK := 14 // 20 * 0.7 = 14
	expectedDEF := 11 // 15 * 0.7 = 10.5 → 11 (rounded)
	expectedSPD := 7  // 10 * 0.7 = 7
	if stats.ATK != expectedATK {
		t.Errorf("overcomes: ATK = %d, want %d", stats.ATK, expectedATK)
	}
	if stats.DEF != expectedDEF {
		t.Errorf("overcomes: DEF = %d, want %d", stats.DEF, expectedDEF)
	}
	if stats.SPD != expectedSPD {
		t.Errorf("overcomes: SPD = %d, want %d", stats.SPD, expectedSPD)
	}
}

func TestCalcCombatStats_Neutral(t *testing.T) {
	// Fire and Earth: Fire generates Earth, not the other way. Earth does not generate Fire.
	// Earth overcomes Water, not Fire. So Fire beast in Metal room = neutral.
	sp := testSpecies(types.Fire)
	b := NewBeast(1, sp, 0)
	roomChi := &fengshui.RoomChi{RoomID: 1, Current: 50, Capacity: 100, Element: types.Metal}

	stats := b.CalcCombatStats(roomChi)
	// Metal overcomes Wood, not Fire. Fire overcomes Metal. But we check room→beast.
	// Metal does not generate Fire. Metal is not Fire. Metal overcomes Wood not Fire. So neutral.
	// Wait: Fire overcomes Metal means Overcomes(Fire, Metal)=true. But we check Overcomes(room, beast) = Overcomes(Metal, Fire).
	// Metal overcomes Wood. So Overcomes(Metal, Fire) = false. Neutral.
	if stats.ATK != b.ATK {
		t.Errorf("neutral: ATK = %d, want %d", stats.ATK, b.ATK)
	}
	if stats.DEF != b.DEF {
		t.Errorf("neutral: DEF = %d, want %d", stats.DEF, b.DEF)
	}
	if stats.SPD != b.SPD {
		t.Errorf("neutral: SPD = %d, want %d", stats.SPD, b.SPD)
	}
}

func TestRoomAffinity(t *testing.T) {
	tests := []struct {
		name    string
		beast   types.Element
		room    types.Element
		want    float64
	}{
		{"generates", types.Fire, types.Wood, 1.3},      // Wood generates Fire
		{"same", types.Fire, types.Fire, 1.1},            // same element
		{"overcomes", types.Fire, types.Water, 0.7},      // Water overcomes Fire
		{"neutral", types.Fire, types.Metal, 1.0},         // no relationship
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoomAffinity(tt.beast, tt.room)
			if got != tt.want {
				t.Errorf("RoomAffinity(%v, %v) = %v, want %v", tt.beast, tt.room, got, tt.want)
			}
		})
	}
}

func TestGrowthAffinity(t *testing.T) {
	// Should return same values as RoomAffinity.
	got := GrowthAffinity(types.Fire, types.Wood)
	want := RoomAffinity(types.Fire, types.Wood)
	if got != want {
		t.Errorf("GrowthAffinity != RoomAffinity: %v != %v", got, want)
	}
}

func TestBeastState_String(t *testing.T) {
	tests := []struct {
		state BeastState
		want  string
	}{
		{Idle, "Idle"},
		{Patrolling, "Patrolling"},
		{Chasing, "Chasing"},
		{Fighting, "Fighting"},
		{Recovering, "Recovering"},
		{BeastState(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("BeastState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}
