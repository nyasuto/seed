package invasion

import (
	"fmt"
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func testClassRegistry(t *testing.T) *InvaderClassRegistry {
	t.Helper()
	reg := NewInvaderClassRegistry()
	classes := []InvaderClass{
		{ID: "bandit", Name: "山賊", Element: types.Fire, BaseHP: 50, BaseATK: 20, BaseDEF: 10, BaseSPD: 15, RewardChi: 10.0, PreferredGoal: StealTreasure, RetreatThreshold: 0.3, Description: "test bandit"},
		{ID: "warrior", Name: "武闘家", Element: types.Metal, BaseHP: 80, BaseATK: 30, BaseDEF: 20, BaseSPD: 10, RewardChi: 20.0, PreferredGoal: DestroyCore, RetreatThreshold: 0.2, Description: "test warrior"},
	}
	for _, c := range classes {
		if err := reg.Register(c); err != nil {
			t.Fatalf("Register class %s: %v", c.ID, err)
		}
	}
	return reg
}

func TestMarshalUnmarshalInvasionState_SaveRestore(t *testing.T) {
	reg := testClassRegistry(t)

	inv1 := &Invader{
		ID: 1, ClassID: "bandit", Name: "山賊", Element: types.Fire,
		Level: 3, HP: 40, MaxHP: 60, ATK: 24, DEF: 12, SPD: 18,
		CurrentRoomID: 5,
		Goal:          &StealTreasureGoal{},
		Memory: &ExplorationMemory{
			VisitedRooms:       map[int]types.Tick{1: 10, 5: 15},
			KnownBeastRooms:    map[int]bool{3: true, 4: false},
			KnownCoreRoom:      2,
			KnownTreasureRooms: []int{7, 9},
		},
		State: Advancing, SlowTicks: 0, EntryTick: 10, StayTicks: 2,
	}

	inv2 := &Invader{
		ID: 2, ClassID: "warrior", Name: "武闘家", Element: types.Metal,
		Level: 5, HP: 30, MaxHP: 112, ATK: 42, DEF: 28, SPD: 14,
		CurrentRoomID: 3,
		Goal:          &HuntBeastsGoal{RequiredKills: 3, Kills: 1},
		Memory: &ExplorationMemory{
			VisitedRooms:       map[int]types.Tick{1: 5, 3: 8},
			KnownBeastRooms:    map[int]bool{3: true},
			KnownCoreRoom:      0,
			KnownTreasureRooms: nil,
		},
		State: Fighting, SlowTicks: 2, EntryTick: 5, StayTicks: 0,
	}

	inv3 := &Invader{
		ID: 3, ClassID: "warrior", Name: "武闘家", Element: types.Metal,
		Level: 2, HP: 88, MaxHP: 88, ATK: 33, DEF: 22, SPD: 11,
		CurrentRoomID: 2,
		Goal:          &DestroyCoreGoal{RequiredStayTicks: 5},
		Memory:        NewExplorationMemory(),
		State:         Retreating, SlowTicks: 0, EntryTick: 12, StayTicks: 3,
	}

	waves := []*InvasionWave{
		{ID: 1, TriggerTick: 10, Invaders: []*Invader{inv1}, State: Active, Difficulty: 1.0},
		{ID: 2, TriggerTick: 50, Invaders: []*Invader{inv2, inv3}, State: Active, Difficulty: 1.5},
	}

	data, err := MarshalInvasionState(waves)
	if err != nil {
		t.Fatalf("MarshalInvasionState: %v", err)
	}

	restored, err := UnmarshalInvasionState(data, reg)
	if err != nil {
		t.Fatalf("UnmarshalInvasionState: %v", err)
	}

	if len(restored) != len(waves) {
		t.Fatalf("wave count: got %d, want %d", len(restored), len(waves))
	}

	for i, orig := range waves {
		rest := restored[i]
		if rest.ID != orig.ID {
			t.Errorf("wave %d ID: got %d, want %d", i, rest.ID, orig.ID)
		}
		if rest.TriggerTick != orig.TriggerTick {
			t.Errorf("wave %d TriggerTick: got %d, want %d", i, rest.TriggerTick, orig.TriggerTick)
		}
		if rest.State != orig.State {
			t.Errorf("wave %d State: got %v, want %v", i, rest.State, orig.State)
		}
		if rest.Difficulty != orig.Difficulty {
			t.Errorf("wave %d Difficulty: got %f, want %f", i, rest.Difficulty, orig.Difficulty)
		}
		if len(rest.Invaders) != len(orig.Invaders) {
			t.Fatalf("wave %d invader count: got %d, want %d", i, len(rest.Invaders), len(orig.Invaders))
		}

		for j, oi := range orig.Invaders {
			ri := rest.Invaders[j]
			assertInvaderEqual(t, i, j, ri, oi)
		}
	}
}

func assertInvaderEqual(t *testing.T, waveIdx, invIdx int, got, want *Invader) {
	t.Helper()
	prefix := func(field string) string {
		return fmt.Sprintf("wave %d invader %d %s", waveIdx, invIdx, field)
	}
	if got.ID != want.ID {
		t.Errorf("%s: got %d, want %d", prefix("ID"), got.ID, want.ID)
	}
	if got.ClassID != want.ClassID {
		t.Errorf("%s: got %s, want %s", prefix("ClassID"), got.ClassID, want.ClassID)
	}
	if got.Name != want.Name {
		t.Errorf("%s: got %s, want %s", prefix("Name"), got.Name, want.Name)
	}
	if got.Element != want.Element {
		t.Errorf("%s: got %v, want %v", prefix("Element"), got.Element, want.Element)
	}
	if got.Level != want.Level {
		t.Errorf("%s: got %d, want %d", prefix("Level"), got.Level, want.Level)
	}
	if got.HP != want.HP {
		t.Errorf("%s: got %d, want %d", prefix("HP"), got.HP, want.HP)
	}
	if got.MaxHP != want.MaxHP {
		t.Errorf("%s: got %d, want %d", prefix("MaxHP"), got.MaxHP, want.MaxHP)
	}
	if got.ATK != want.ATK {
		t.Errorf("%s: got %d, want %d", prefix("ATK"), got.ATK, want.ATK)
	}
	if got.DEF != want.DEF {
		t.Errorf("%s: got %d, want %d", prefix("DEF"), got.DEF, want.DEF)
	}
	if got.SPD != want.SPD {
		t.Errorf("%s: got %d, want %d", prefix("SPD"), got.SPD, want.SPD)
	}
	if got.CurrentRoomID != want.CurrentRoomID {
		t.Errorf("%s: got %d, want %d", prefix("CurrentRoomID"), got.CurrentRoomID, want.CurrentRoomID)
	}
	if got.State != want.State {
		t.Errorf("%s: got %v, want %v", prefix("State"), got.State, want.State)
	}
	if got.SlowTicks != want.SlowTicks {
		t.Errorf("%s: got %d, want %d", prefix("SlowTicks"), got.SlowTicks, want.SlowTicks)
	}
	if got.EntryTick != want.EntryTick {
		t.Errorf("%s: got %d, want %d", prefix("EntryTick"), got.EntryTick, want.EntryTick)
	}
	if got.StayTicks != want.StayTicks {
		t.Errorf("%s: got %d, want %d", prefix("StayTicks"), got.StayTicks, want.StayTicks)
	}

	// Goal
	if got.Goal.Type() != want.Goal.Type() {
		t.Errorf("%s: got %v, want %v", prefix("Goal.Type"), got.Goal.Type(), want.Goal.Type())
	}
	switch wg := want.Goal.(type) {
	case *DestroyCoreGoal:
		gg, ok := got.Goal.(*DestroyCoreGoal)
		if !ok {
			t.Errorf("%s: type mismatch", prefix("Goal"))
		} else if gg.RequiredStayTicks != wg.RequiredStayTicks {
			t.Errorf("%s: got %d, want %d", prefix("Goal.RequiredStayTicks"), gg.RequiredStayTicks, wg.RequiredStayTicks)
		}
	case *HuntBeastsGoal:
		gg, ok := got.Goal.(*HuntBeastsGoal)
		if !ok {
			t.Errorf("%s: type mismatch", prefix("Goal"))
		} else {
			if gg.RequiredKills != wg.RequiredKills {
				t.Errorf("%s: got %d, want %d", prefix("Goal.RequiredKills"), gg.RequiredKills, wg.RequiredKills)
			}
			if gg.Kills != wg.Kills {
				t.Errorf("%s: got %d, want %d", prefix("Goal.Kills"), gg.Kills, wg.Kills)
			}
		}
	}

	// Memory
	assertMemoryEqual(t, prefix("Memory"), got.Memory, want.Memory)
}

func assertMemoryEqual(t *testing.T, prefix string, got, want *ExplorationMemory) {
	t.Helper()
	if len(got.VisitedRooms) != len(want.VisitedRooms) {
		t.Errorf("%s.VisitedRooms len: got %d, want %d", prefix, len(got.VisitedRooms), len(want.VisitedRooms))
	} else {
		for k, wv := range want.VisitedRooms {
			gv, ok := got.VisitedRooms[k]
			if !ok {
				t.Errorf("%s.VisitedRooms[%d]: missing", prefix, k)
			} else if gv != wv {
				t.Errorf("%s.VisitedRooms[%d]: got %d, want %d", prefix, k, gv, wv)
			}
		}
	}
	if len(got.KnownBeastRooms) != len(want.KnownBeastRooms) {
		t.Errorf("%s.KnownBeastRooms len: got %d, want %d", prefix, len(got.KnownBeastRooms), len(want.KnownBeastRooms))
	} else {
		for k, wv := range want.KnownBeastRooms {
			gv, ok := got.KnownBeastRooms[k]
			if !ok {
				t.Errorf("%s.KnownBeastRooms[%d]: missing", prefix, k)
			} else if gv != wv {
				t.Errorf("%s.KnownBeastRooms[%d]: got %v, want %v", prefix, k, gv, wv)
			}
		}
	}
	if got.KnownCoreRoom != want.KnownCoreRoom {
		t.Errorf("%s.KnownCoreRoom: got %d, want %d", prefix, got.KnownCoreRoom, want.KnownCoreRoom)
	}
	if len(got.KnownTreasureRooms) != len(want.KnownTreasureRooms) {
		t.Errorf("%s.KnownTreasureRooms len: got %d, want %d", prefix, len(got.KnownTreasureRooms), len(want.KnownTreasureRooms))
	} else {
		for k, wv := range want.KnownTreasureRooms {
			if got.KnownTreasureRooms[k] != wv {
				t.Errorf("%s.KnownTreasureRooms[%d]: got %d, want %d", prefix, k, got.KnownTreasureRooms[k], wv)
			}
		}
	}
}

func TestMarshalUnmarshalInvasionState_MixedAdvancingRetreating(t *testing.T) {
	reg := testClassRegistry(t)

	inv1 := &Invader{
		ID: 1, ClassID: "bandit", Name: "山賊", Element: types.Fire,
		Level: 1, HP: 50, MaxHP: 50, ATK: 20, DEF: 10, SPD: 15,
		CurrentRoomID: 3, Goal: &StealTreasureGoal{},
		Memory: NewExplorationMemory(), State: Advancing,
		EntryTick: 10,
	}

	inv2 := &Invader{
		ID: 2, ClassID: "warrior", Name: "武闘家", Element: types.Metal,
		Level: 2, HP: 20, MaxHP: 88, ATK: 33, DEF: 22, SPD: 11,
		CurrentRoomID: 1, Goal: &DestroyCoreGoal{RequiredStayTicks: 5},
		Memory: &ExplorationMemory{
			VisitedRooms:       map[int]types.Tick{1: 5, 2: 8, 3: 12},
			KnownBeastRooms:    map[int]bool{2: true},
			KnownCoreRoom:      3,
			KnownTreasureRooms: nil,
		},
		State: Retreating, SlowTicks: 1, EntryTick: 5, StayTicks: 0,
	}

	inv3 := &Invader{
		ID: 3, ClassID: "bandit", Name: "山賊", Element: types.Fire,
		Level: 1, HP: 0, MaxHP: 50, ATK: 20, DEF: 10, SPD: 15,
		CurrentRoomID: 2, Goal: &StealTreasureGoal{},
		Memory: NewExplorationMemory(), State: Defeated,
		EntryTick: 10,
	}

	waves := []*InvasionWave{
		{ID: 1, TriggerTick: 10, Invaders: []*Invader{inv1, inv2, inv3}, State: Active, Difficulty: 1.0},
	}

	data, err := MarshalInvasionState(waves)
	if err != nil {
		t.Fatalf("MarshalInvasionState: %v", err)
	}

	restored, err := UnmarshalInvasionState(data, reg)
	if err != nil {
		t.Fatalf("UnmarshalInvasionState: %v", err)
	}

	if len(restored) != 1 {
		t.Fatalf("wave count: got %d, want 1", len(restored))
	}
	if len(restored[0].Invaders) != 3 {
		t.Fatalf("invader count: got %d, want 3", len(restored[0].Invaders))
	}

	// Verify mixed states preserved.
	states := []InvaderState{Advancing, Retreating, Defeated}
	for i, want := range states {
		if restored[0].Invaders[i].State != want {
			t.Errorf("invader %d state: got %v, want %v", i, restored[0].Invaders[i].State, want)
		}
	}
}

func TestMarshalUnmarshalInvasionState_EmptyList(t *testing.T) {
	reg := testClassRegistry(t)

	data, err := MarshalInvasionState([]*InvasionWave{})
	if err != nil {
		t.Fatalf("MarshalInvasionState: %v", err)
	}

	restored, err := UnmarshalInvasionState(data, reg)
	if err != nil {
		t.Fatalf("UnmarshalInvasionState: %v", err)
	}

	if len(restored) != 0 {
		t.Errorf("count: got %d, want 0", len(restored))
	}
}

func TestMarshalUnmarshalInvasionState_NilSlice(t *testing.T) {
	reg := testClassRegistry(t)

	data, err := MarshalInvasionState(nil)
	if err != nil {
		t.Fatalf("MarshalInvasionState nil: %v", err)
	}

	restored, err := UnmarshalInvasionState(data, reg)
	if err != nil {
		t.Fatalf("UnmarshalInvasionState: %v", err)
	}

	if restored == nil {
		t.Error("restored should not be nil (expected empty slice)")
	}
}

func TestUnmarshalInvasionState_InvalidJSON(t *testing.T) {
	reg := testClassRegistry(t)

	_, err := UnmarshalInvasionState([]byte("not json"), reg)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestUnmarshalInvasionState_UnknownClassID(t *testing.T) {
	reg := testClassRegistry(t)

	data := []byte(`[{"id":1,"trigger_tick":0,"invaders":[{"id":1,"class_id":"unknown","name":"?","element":"Fire","level":1,"hp":10,"max_hp":10,"atk":5,"def":5,"spd":5,"current_room_id":0,"goal":{"type":"StealTreasure"},"memory":null,"state":"Advancing","slow_ticks":0,"entry_tick":0,"stay_ticks":0}],"state":"Active","difficulty":1.0}]`)
	_, err := UnmarshalInvasionState(data, reg)
	if err == nil {
		t.Error("expected error for unknown class ID, got nil")
	}
}
