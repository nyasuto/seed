package senju

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func testSpeciesRegistry(t *testing.T) *SpeciesRegistry {
	t.Helper()
	reg := NewSpeciesRegistry()
	species := []*Species{
		{ID: "suiryu", Name: "翠龍", Element: types.Wood, BaseHP: 100, BaseATK: 30, BaseDEF: 30, BaseSPD: 25, GrowthRate: 1.0, MaxBeasts: 3},
		{ID: "enhou", Name: "炎鳳", Element: types.Fire, BaseHP: 80, BaseATK: 45, BaseDEF: 20, BaseSPD: 30, GrowthRate: 1.1, MaxBeasts: 3},
	}
	for _, s := range species {
		if err := reg.Register(s); err != nil {
			t.Fatalf("Register species %s: %v", s.ID, err)
		}
	}
	return reg
}

func TestMarshalUnmarshalBeasts_SaveRestore(t *testing.T) {
	reg := testSpeciesRegistry(t)
	suiryu, _ := reg.Get("suiryu")
	enhou, _ := reg.Get("enhou")

	b1 := NewBeast(1, suiryu, types.Tick(10))
	b1.RoomID = 5
	b1.Level = 3
	b1.EXP = 42
	b1.HP = 80
	b1.MaxHP = 120
	b1.ATK = 35
	b1.DEF = 33
	b1.SPD = 28
	b1.State = Patrolling

	b2 := NewBeast(2, enhou, types.Tick(20))
	b2.RoomID = 7
	b2.Level = 5
	b2.EXP = 10
	b2.State = Fighting

	beasts := []*Beast{b1, b2}

	data, err := MarshalBeasts(beasts)
	if err != nil {
		t.Fatalf("MarshalBeasts: %v", err)
	}

	restored, err := UnmarshalBeasts(data, reg)
	if err != nil {
		t.Fatalf("UnmarshalBeasts: %v", err)
	}

	if len(restored) != len(beasts) {
		t.Fatalf("count: got %d, want %d", len(restored), len(beasts))
	}

	for i, orig := range beasts {
		rest := restored[i]
		if rest.ID != orig.ID {
			t.Errorf("beast %d ID: got %d, want %d", i, rest.ID, orig.ID)
		}
		if rest.SpeciesID != orig.SpeciesID {
			t.Errorf("beast %d SpeciesID: got %s, want %s", i, rest.SpeciesID, orig.SpeciesID)
		}
		if rest.Name != orig.Name {
			t.Errorf("beast %d Name: got %s, want %s", i, rest.Name, orig.Name)
		}
		if rest.Element != orig.Element {
			t.Errorf("beast %d Element: got %v, want %v", i, rest.Element, orig.Element)
		}
		if rest.RoomID != orig.RoomID {
			t.Errorf("beast %d RoomID: got %d, want %d", i, rest.RoomID, orig.RoomID)
		}
		if rest.Level != orig.Level {
			t.Errorf("beast %d Level: got %d, want %d", i, rest.Level, orig.Level)
		}
		if rest.EXP != orig.EXP {
			t.Errorf("beast %d EXP: got %d, want %d", i, rest.EXP, orig.EXP)
		}
		if rest.HP != orig.HP {
			t.Errorf("beast %d HP: got %d, want %d", i, rest.HP, orig.HP)
		}
		if rest.MaxHP != orig.MaxHP {
			t.Errorf("beast %d MaxHP: got %d, want %d", i, rest.MaxHP, orig.MaxHP)
		}
		if rest.ATK != orig.ATK {
			t.Errorf("beast %d ATK: got %d, want %d", i, rest.ATK, orig.ATK)
		}
		if rest.DEF != orig.DEF {
			t.Errorf("beast %d DEF: got %d, want %d", i, rest.DEF, orig.DEF)
		}
		if rest.SPD != orig.SPD {
			t.Errorf("beast %d SPD: got %d, want %d", i, rest.SPD, orig.SPD)
		}
		if rest.BornTick != orig.BornTick {
			t.Errorf("beast %d BornTick: got %d, want %d", i, rest.BornTick, orig.BornTick)
		}
		if rest.State != orig.State {
			t.Errorf("beast %d State: got %v, want %v", i, rest.State, orig.State)
		}
	}
}

func TestMarshalUnmarshalBeasts_EmptyList(t *testing.T) {
	reg := testSpeciesRegistry(t)

	data, err := MarshalBeasts([]*Beast{})
	if err != nil {
		t.Fatalf("MarshalBeasts: %v", err)
	}

	restored, err := UnmarshalBeasts(data, reg)
	if err != nil {
		t.Fatalf("UnmarshalBeasts: %v", err)
	}

	if len(restored) != 0 {
		t.Errorf("count: got %d, want 0", len(restored))
	}
}

func TestUnmarshalBeasts_InvalidJSON(t *testing.T) {
	reg := testSpeciesRegistry(t)

	_, err := UnmarshalBeasts([]byte("not json"), reg)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestUnmarshalBeasts_UnknownSpeciesID(t *testing.T) {
	reg := testSpeciesRegistry(t)

	// JSON with a species ID that does not exist in the registry.
	data := []byte(`[{"id":1,"species_id":"unknown_species","name":"???","element":"Wood","room_id":0,"level":1,"exp":0,"hp":50,"max_hp":50,"atk":10,"def":10,"spd":10,"born_tick":0,"state":0}]`)
	_, err := UnmarshalBeasts(data, reg)
	if err == nil {
		t.Error("expected error for unknown species ID, got nil")
	}
}

func TestMarshalUnmarshalBeasts_NilSlice(t *testing.T) {
	reg := testSpeciesRegistry(t)

	data, err := MarshalBeasts(nil)
	if err != nil {
		t.Fatalf("MarshalBeasts nil: %v", err)
	}

	restored, err := UnmarshalBeasts(data, reg)
	if err != nil {
		t.Fatalf("UnmarshalBeasts: %v", err)
	}

	if restored == nil {
		t.Error("restored should not be nil (expected empty slice)")
	}
}
