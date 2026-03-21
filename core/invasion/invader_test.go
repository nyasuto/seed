package invasion

import (
	_ "embed"
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

//go:embed invader_class_data.json
var invaderClassDataJSON []byte

func TestInvaderClassRegistry_LoadJSON(t *testing.T) {
	reg, err := LoadInvaderClassesJSON(invaderClassDataJSON)
	if err != nil {
		t.Fatalf("loading invader classes: %v", err)
	}
	if reg.Len() != 5 {
		t.Errorf("expected 5 classes, got %d", reg.Len())
	}

	tests := []struct {
		id      string
		element types.Element
		goal    GoalType
	}{
		{"wood_ascetic", types.Wood, DestroyCore},
		{"fire_fighter", types.Fire, HuntBeasts},
		{"earth_knight", types.Earth, DestroyCore},
		{"metal_thief", types.Metal, StealTreasure},
		{"water_taoist", types.Water, DestroyCore},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			ic, err := reg.Get(tt.id)
			if err != nil {
				t.Fatalf("getting class %q: %v", tt.id, err)
			}
			if ic.Element != tt.element {
				t.Errorf("element = %v, want %v", ic.Element, tt.element)
			}
			if ic.PreferredGoal != tt.goal {
				t.Errorf("preferred goal = %v, want %v", ic.PreferredGoal, tt.goal)
			}
		})
	}
}

func TestInvaderClassRegistry_GetNotFound(t *testing.T) {
	reg := NewInvaderClassRegistry()
	_, err := reg.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent class")
	}
}

func TestNewInvader_LevelOne(t *testing.T) {
	class := InvaderClass{
		ID:               "test_class",
		Name:             "テスト侵入者",
		Element:          types.Wood,
		BaseHP:           100,
		BaseATK:          30,
		BaseDEF:          20,
		BaseSPD:          15,
		RewardChi:        10.0,
		PreferredGoal:    DestroyCore,
		RetreatThreshold: 0.3,
	}
	goal := NewDestroyCoreGoal()
	inv := NewInvader(1, class, 1, goal, 5, types.Tick(10))

	if inv.ID != 1 {
		t.Errorf("ID = %d, want 1", inv.ID)
	}
	if inv.ClassID != "test_class" {
		t.Errorf("ClassID = %q, want %q", inv.ClassID, "test_class")
	}
	if inv.Element != types.Wood {
		t.Errorf("Element = %v, want Wood", inv.Element)
	}
	if inv.Level != 1 {
		t.Errorf("Level = %d, want 1", inv.Level)
	}
	if inv.HP != 100 {
		t.Errorf("HP = %d, want 100", inv.HP)
	}
	if inv.MaxHP != 100 {
		t.Errorf("MaxHP = %d, want 100", inv.MaxHP)
	}
	if inv.ATK != 30 {
		t.Errorf("ATK = %d, want 30", inv.ATK)
	}
	if inv.DEF != 20 {
		t.Errorf("DEF = %d, want 20", inv.DEF)
	}
	if inv.SPD != 15 {
		t.Errorf("SPD = %d, want 15", inv.SPD)
	}
	if inv.CurrentRoomID != 5 {
		t.Errorf("CurrentRoomID = %d, want 5", inv.CurrentRoomID)
	}
	if inv.State != Advancing {
		t.Errorf("State = %v, want Advancing", inv.State)
	}
	if inv.EntryTick != 10 {
		t.Errorf("EntryTick = %d, want 10", inv.EntryTick)
	}
	if inv.Memory == nil {
		t.Fatal("Memory should not be nil")
	}
	if inv.Goal == nil {
		t.Fatal("Goal should not be nil")
	}
}

func TestNewInvader_LevelScaling(t *testing.T) {
	class := InvaderClass{
		ID:      "test_class",
		Name:    "テスト",
		Element: types.Fire,
		BaseHP:  100,
		BaseATK: 30,
		BaseDEF: 20,
		BaseSPD: 10,
	}

	tests := []struct {
		name                        string
		level                       int
		wantHP, wantATK, wantDEF, wantSPD int
	}{
		{"level 1", 1, 100, 30, 20, 10},
		{"level 2", 2, 110, 33, 22, 11},
		{"level 5", 5, 140, 42, 28, 14},
		{"level 10", 10, 190, 57, 38, 19},
		{"level 11", 11, 200, 60, 40, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInvader(1, class, tt.level, NewDestroyCoreGoal(), 1, 0)
			if inv.HP != tt.wantHP {
				t.Errorf("HP = %d, want %d", inv.HP, tt.wantHP)
			}
			if inv.MaxHP != tt.wantHP {
				t.Errorf("MaxHP = %d, want %d", inv.MaxHP, tt.wantHP)
			}
			if inv.ATK != tt.wantATK {
				t.Errorf("ATK = %d, want %d", inv.ATK, tt.wantATK)
			}
			if inv.DEF != tt.wantDEF {
				t.Errorf("DEF = %d, want %d", inv.DEF, tt.wantDEF)
			}
			if inv.SPD != tt.wantSPD {
				t.Errorf("SPD = %d, want %d", inv.SPD, tt.wantSPD)
			}
			if inv.Level != tt.level {
				t.Errorf("Level = %d, want %d", inv.Level, tt.level)
			}
		})
	}
}

func TestNewInvader_LevelZeroClamped(t *testing.T) {
	class := InvaderClass{
		ID:      "test",
		BaseHP:  100,
		BaseATK: 30,
		BaseDEF: 20,
		BaseSPD: 10,
	}
	inv := NewInvader(1, class, 0, NewDestroyCoreGoal(), 1, 0)
	if inv.Level != 1 {
		t.Errorf("Level = %d, want 1 (clamped from 0)", inv.Level)
	}
	if inv.HP != 100 {
		t.Errorf("HP = %d, want 100 (level 1 base)", inv.HP)
	}
}

func TestGoalType_TargetRoomDecision(t *testing.T) {
	// Build a cave with dragon_hole, beast room, and storage room.
	cave, err := world.NewCave(20, 20)
	if err != nil {
		t.Fatalf("creating cave: %v", err)
	}

	coreRoom, err := cave.AddRoom("dragon_hole", types.Pos{X: 2, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding core room: %v", err)
	}

	beastRoom, err := cave.AddRoom("beast_room", types.Pos{X: 8, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding beast room: %v", err)
	}
	// Simulate a beast in the beast room.
	beastRoom.BeastIDs = append(beastRoom.BeastIDs, 99)

	storageRoom, err := cave.AddRoom("storage", types.Pos{X: 14, Y: 2}, 3, 3, nil)
	if err != nil {
		t.Fatalf("adding storage room: %v", err)
	}

	tests := []struct {
		name       string
		goal       Goal
		wantRoomID int
	}{
		{
			"DestroyCore targets dragon_hole",
			NewDestroyCoreGoal(),
			coreRoom.ID,
		},
		{
			"HuntBeasts targets beast room",
			NewHuntBeastsGoal(),
			beastRoom.ID,
		},
		{
			"StealTreasure targets storage",
			NewStealTreasureGoal(),
			storageRoom.ID,
		},
	}

	class := InvaderClass{
		ID:      "test",
		Element: types.Wood,
		BaseHP:  100,
		BaseATK: 20,
		BaseDEF: 10,
		BaseSPD: 10,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInvader(1, class, 1, tt.goal, coreRoom.ID, 0)
			target := tt.goal.TargetRoomID(cave, inv, inv.Memory)
			if target != tt.wantRoomID {
				t.Errorf("TargetRoomID = %d, want %d", target, tt.wantRoomID)
			}
		})
	}
}

func TestInvaderState_String(t *testing.T) {
	tests := []struct {
		state InvaderState
		want  string
	}{
		{Advancing, "Advancing"},
		{Fighting, "Fighting"},
		{Retreating, "Retreating"},
		{Defeated, "Defeated"},
		{GoalAchieved, "GoalAchieved"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestNewExplorationMemory(t *testing.T) {
	mem := NewExplorationMemory()
	if mem.VisitedRooms == nil {
		t.Error("VisitedRooms should be initialized")
	}
	if mem.KnownBeastRooms == nil {
		t.Error("KnownBeastRooms should be initialized")
	}
	if mem.KnownCoreRoom != 0 {
		t.Errorf("KnownCoreRoom = %d, want 0", mem.KnownCoreRoom)
	}
	if mem.KnownTreasureRooms != nil {
		t.Errorf("KnownTreasureRooms should be nil initially")
	}
}
