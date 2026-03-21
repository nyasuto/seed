package senju

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

// helper to create a species registry with base and evolved species.
func setupEvolutionSpecies(t *testing.T) *SpeciesRegistry {
	t.Helper()
	reg := NewSpeciesRegistry()
	species := []*Species{
		{ID: "suiryu", Name: "翠龍", Element: types.Wood, BaseHP: 30, BaseATK: 10, BaseDEF: 8, BaseSPD: 12},
		{ID: "souryu", Name: "蒼龍", Element: types.Wood, BaseHP: 50, BaseATK: 18, BaseDEF: 15, BaseSPD: 20},
		{ID: "enhou", Name: "炎鳳", Element: types.Fire, BaseHP: 25, BaseATK: 14, BaseDEF: 6, BaseSPD: 14},
		{ID: "suzaku", Name: "朱雀", Element: types.Fire, BaseHP: 45, BaseATK: 24, BaseDEF: 12, BaseSPD: 22},
	}
	for _, s := range species {
		if err := reg.Register(s); err != nil {
			t.Fatalf("register species %s: %v", s.ID, err)
		}
	}
	return reg
}

func TestEvolve_BasicExecution(t *testing.T) {
	specReg := setupEvolutionSpecies(t)

	beast := &Beast{
		ID:        1,
		SpeciesID: "suiryu",
		Name:      "翠龍",
		Element:   types.Wood,
		Level:     15,
		HP:        20, // damaged
		MaxHP:     58, // 30 + (15-1)*2
		ATK:       24, // 10 + (15-1)*1
		DEF:       22, // 8 + (15-1)*1
		SPD:       26, // 12 + (15-1)*1
	}

	path := &EvolutionPath{
		FromSpeciesID: "suiryu",
		ToSpeciesID:   "souryu",
		Condition:     EvolutionCondition{MinLevel: 15},
		ChiCost:       100,
	}

	if err := Evolve(beast, path, specReg); err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	// Species should be updated.
	if beast.SpeciesID != "souryu" {
		t.Errorf("SpeciesID: want souryu, got %s", beast.SpeciesID)
	}
	if beast.Name != "蒼龍" {
		t.Errorf("Name: want 蒼龍, got %s", beast.Name)
	}
	if beast.Element != types.Wood {
		t.Errorf("Element: want Wood, got %v", beast.Element)
	}

	// Level should be preserved.
	if beast.Level != 15 {
		t.Errorf("Level: want 15, got %d", beast.Level)
	}

	// Stats should be recalculated from new species base + level.
	// MaxHP = 50 + (15-1)*2 = 78
	wantMaxHP := 50 + (15-1)*2
	if beast.MaxHP != wantMaxHP {
		t.Errorf("MaxHP: want %d, got %d", wantMaxHP, beast.MaxHP)
	}
	// HP fully restored.
	if beast.HP != beast.MaxHP {
		t.Errorf("HP should be fully restored: want %d, got %d", beast.MaxHP, beast.HP)
	}
	// ATK = 18 + (15-1)*1 = 32
	wantATK := 18 + (15-1)*1
	if beast.ATK != wantATK {
		t.Errorf("ATK: want %d, got %d", wantATK, beast.ATK)
	}
	// DEF = 15 + (15-1)*1 = 29
	wantDEF := 15 + (15-1)*1
	if beast.DEF != wantDEF {
		t.Errorf("DEF: want %d, got %d", wantDEF, beast.DEF)
	}
	// SPD = 20 + (15-1)*1 = 34
	wantSPD := 20 + (15-1)*1
	if beast.SPD != wantSPD {
		t.Errorf("SPD: want %d, got %d", wantSPD, beast.SPD)
	}
}

func TestEvolve_StatsRecalculation(t *testing.T) {
	specReg := setupEvolutionSpecies(t)

	tests := []struct {
		name      string
		level     int
		wantMaxHP int
		wantATK   int
		wantDEF   int
		wantSPD   int
	}{
		{
			name:      "level 1 uses base stats only",
			level:     1,
			wantMaxHP: 50,  // 50 + 0
			wantATK:   18,  // 18 + 0
			wantDEF:   15,  // 15 + 0
			wantSPD:   20,  // 20 + 0
		},
		{
			name:      "level 20 higher stats",
			level:     20,
			wantMaxHP: 88,  // 50 + 19*2
			wantATK:   37,  // 18 + 19
			wantDEF:   34,  // 15 + 19
			wantSPD:   39,  // 20 + 19
		},
	}

	path := &EvolutionPath{
		FromSpeciesID: "suiryu",
		ToSpeciesID:   "souryu",
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			beast := &Beast{
				SpeciesID: "suiryu",
				Level:     tc.level,
				HP:        1,
			}
			if err := Evolve(beast, path, specReg); err != nil {
				t.Fatalf("Evolve failed: %v", err)
			}
			if beast.MaxHP != tc.wantMaxHP {
				t.Errorf("MaxHP: want %d, got %d", tc.wantMaxHP, beast.MaxHP)
			}
			if beast.HP != tc.wantMaxHP {
				t.Errorf("HP should equal MaxHP: want %d, got %d", tc.wantMaxHP, beast.HP)
			}
			if beast.ATK != tc.wantATK {
				t.Errorf("ATK: want %d, got %d", tc.wantATK, beast.ATK)
			}
			if beast.DEF != tc.wantDEF {
				t.Errorf("DEF: want %d, got %d", tc.wantDEF, beast.DEF)
			}
			if beast.SPD != tc.wantSPD {
				t.Errorf("SPD: want %d, got %d", tc.wantSPD, beast.SPD)
			}
		})
	}
}

func TestEvolve_ElementChange(t *testing.T) {
	specReg := setupEvolutionSpecies(t)

	beast := &Beast{
		SpeciesID: "enhou",
		Element:   types.Fire,
		Level:     15,
		HP:        10,
	}

	path := &EvolutionPath{
		FromSpeciesID: "enhou",
		ToSpeciesID:   "suzaku",
	}

	if err := Evolve(beast, path, specReg); err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if beast.Element != types.Fire {
		t.Errorf("Element: want Fire, got %v", beast.Element)
	}
	if beast.Name != "朱雀" {
		t.Errorf("Name: want 朱雀, got %s", beast.Name)
	}
}

func TestEvolve_TargetNotFound(t *testing.T) {
	specReg := setupEvolutionSpecies(t)

	beast := &Beast{
		SpeciesID: "suiryu",
		Level:     15,
	}

	path := &EvolutionPath{
		FromSpeciesID: "suiryu",
		ToSpeciesID:   "nonexistent",
	}

	err := Evolve(beast, path, specReg)
	if err == nil {
		t.Fatal("expected error for nonexistent target species")
	}
	if !errors.Is(err, ErrEvolutionTargetNotFound) {
		t.Errorf("expected ErrEvolutionTargetNotFound, got %v", err)
	}

	// Beast should remain unchanged on error.
	if beast.SpeciesID != "suiryu" {
		t.Errorf("SpeciesID should not change on error: got %s", beast.SpeciesID)
	}
}

func TestEvolve_HPFullyRestored(t *testing.T) {
	specReg := setupEvolutionSpecies(t)

	beast := &Beast{
		SpeciesID: "suiryu",
		Level:     15,
		HP:        1, // nearly dead
		MaxHP:     58,
	}

	path := &EvolutionPath{
		FromSpeciesID: "suiryu",
		ToSpeciesID:   "souryu",
	}

	if err := Evolve(beast, path, specReg); err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if beast.HP != beast.MaxHP {
		t.Errorf("HP should be fully restored after evolution: HP=%d, MaxHP=%d", beast.HP, beast.MaxHP)
	}
}

func TestEvolve_PreservesNonStatFields(t *testing.T) {
	specReg := setupEvolutionSpecies(t)

	beast := &Beast{
		ID:        42,
		SpeciesID: "suiryu",
		Level:     15,
		EXP:       500,
		RoomID:    7,
		BornTick:  100,
		State:     Patrolling,
		HP:        20,
	}

	path := &EvolutionPath{
		FromSpeciesID: "suiryu",
		ToSpeciesID:   "souryu",
	}

	if err := Evolve(beast, path, specReg); err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	// Fields that should be preserved.
	if beast.ID != 42 {
		t.Errorf("ID should be preserved: want 42, got %d", beast.ID)
	}
	if beast.Level != 15 {
		t.Errorf("Level should be preserved: want 15, got %d", beast.Level)
	}
	if beast.EXP != 500 {
		t.Errorf("EXP should be preserved: want 500, got %d", beast.EXP)
	}
	if beast.RoomID != 7 {
		t.Errorf("RoomID should be preserved: want 7, got %d", beast.RoomID)
	}
	if beast.BornTick != 100 {
		t.Errorf("BornTick should be preserved: want 100, got %d", beast.BornTick)
	}
	if beast.State != Patrolling {
		t.Errorf("State should be preserved: want Patrolling, got %v", beast.State)
	}
}

func TestCheckEvolutionThenEvolve_Integration(t *testing.T) {
	specReg := setupEvolutionSpecies(t)

	evoReg := NewEvolutionRegistry()
	evoReg.Register(EvolutionPath{
		FromSpeciesID: "suiryu",
		ToSpeciesID:   "souryu",
		Condition: EvolutionCondition{
			MinLevel:            15,
			RequiredRoomElement: types.Wood,
			RequireElement:      true,
			MinChiRatio:         0.5,
		},
		ChiCost: 100,
	})

	beast := &Beast{
		SpeciesID: "suiryu",
		Name:      "翠龍",
		Element:   types.Wood,
		Level:     15,
		HP:        20,
		MaxHP:     58,
		ATK:       24,
		DEF:       22,
		SPD:       26,
	}

	// Condition not met: level too low.
	beast.Level = 14
	if path := evoReg.CheckEvolution(beast, types.Wood, 0.8, 200); path != nil {
		t.Error("should not evolve at level 14")
	}

	// Condition not met: wrong element.
	beast.Level = 15
	if path := evoReg.CheckEvolution(beast, types.Fire, 0.8, 200); path != nil {
		t.Error("should not evolve in Fire room")
	}

	// Condition not met: chi ratio too low.
	if path := evoReg.CheckEvolution(beast, types.Wood, 0.3, 200); path != nil {
		t.Error("should not evolve with 0.3 chi ratio")
	}

	// Condition not met: insufficient chi pool.
	if path := evoReg.CheckEvolution(beast, types.Wood, 0.8, 50); path != nil {
		t.Error("should not evolve with 50 chi pool")
	}

	// All conditions met — evolve.
	path := evoReg.CheckEvolution(beast, types.Wood, 0.8, 200)
	if path == nil {
		t.Fatal("expected evolution path, got nil")
	}

	if err := Evolve(beast, path, specReg); err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if beast.SpeciesID != "souryu" {
		t.Errorf("after evolution SpeciesID: want souryu, got %s", beast.SpeciesID)
	}
	if beast.HP != beast.MaxHP {
		t.Errorf("HP should be fully restored: HP=%d, MaxHP=%d", beast.HP, beast.MaxHP)
	}
	// Verify stat increase from evolution.
	// Old MaxHP=58 (base30 + 14*2), New MaxHP=78 (base50 + 14*2)
	if beast.MaxHP <= 58 {
		t.Errorf("MaxHP should increase after evolution: got %d", beast.MaxHP)
	}
}
