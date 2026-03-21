package senju

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestLoadDefaultSpecies_AllFiveLoaded(t *testing.T) {
	reg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("LoadDefaultSpecies() error: %v", err)
	}
	if got := reg.Len(); got != 5 {
		t.Errorf("Len() = %d, want 5", got)
	}
}

func TestSpeciesRegistry_Get_AllSpecies(t *testing.T) {
	reg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("LoadDefaultSpecies() error: %v", err)
	}

	tests := []struct {
		id      string
		name    string
		element types.Element
	}{
		{"suiryu", "翠龍", types.Wood},
		{"enhou", "炎鳳", types.Fire},
		{"ganki", "岩亀", types.Earth},
		{"kinrou", "金狼", types.Metal},
		{"suija", "水蛇", types.Water},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			s, err := reg.Get(tt.id)
			if err != nil {
				t.Fatalf("Get(%q) error: %v", tt.id, err)
			}
			if s.Name != tt.name {
				t.Errorf("Name = %q, want %q", s.Name, tt.name)
			}
			if s.Element != tt.element {
				t.Errorf("Element = %v, want %v", s.Element, tt.element)
			}
		})
	}
}

func TestSpeciesRegistry_Get_NotFound(t *testing.T) {
	reg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("LoadDefaultSpecies() error: %v", err)
	}

	_, err = reg.Get("nonexistent")
	if err == nil {
		t.Fatal("Get(nonexistent) expected error, got nil")
	}
	if !errors.Is(err, ErrSpeciesNotFound) {
		t.Errorf("error = %v, want ErrSpeciesNotFound", err)
	}
}

func TestSpeciesRegistry_All_Sorted(t *testing.T) {
	reg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("LoadDefaultSpecies() error: %v", err)
	}

	all := reg.All()
	if len(all) != 5 {
		t.Fatalf("All() returned %d species, want 5", len(all))
	}

	// Verify sorted by ID
	for i := 1; i < len(all); i++ {
		if all[i-1].ID >= all[i].ID {
			t.Errorf("All() not sorted: %q >= %q at index %d", all[i-1].ID, all[i].ID, i)
		}
	}
}

func TestSpeciesRegistry_Register_Duplicate(t *testing.T) {
	reg := NewSpeciesRegistry()
	s := &Species{ID: "test", Name: "Test", Element: types.Wood}

	if err := reg.Register(s); err != nil {
		t.Fatalf("first Register() error: %v", err)
	}
	if err := reg.Register(s); err == nil {
		t.Error("second Register() expected error for duplicate, got nil")
	}
}

func TestLoadSpeciesJSON_InvalidJSON(t *testing.T) {
	_, err := LoadSpeciesJSON([]byte(`not json`))
	if err == nil {
		t.Error("LoadSpeciesJSON(invalid) expected error, got nil")
	}
}

func TestLoadSpeciesJSON_InvalidElement(t *testing.T) {
	data := []byte(`[{"id":"bad","name":"Bad","element":"Void","base_hp":10,"base_atk":5,"base_def":5,"base_spd":5,"growth_rate":1.0,"max_beasts":1,"description":"test"}]`)
	_, err := LoadSpeciesJSON(data)
	if err == nil {
		t.Error("LoadSpeciesJSON(invalid element) expected error, got nil")
	}
}

func TestSpecies_StatsReasonable(t *testing.T) {
	reg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("LoadDefaultSpecies() error: %v", err)
	}

	for _, s := range reg.All() {
		t.Run(s.ID, func(t *testing.T) {
			if s.BaseHP <= 0 {
				t.Errorf("BaseHP = %d, want > 0", s.BaseHP)
			}
			if s.BaseATK <= 0 {
				t.Errorf("BaseATK = %d, want > 0", s.BaseATK)
			}
			if s.BaseDEF <= 0 {
				t.Errorf("BaseDEF = %d, want > 0", s.BaseDEF)
			}
			if s.BaseSPD <= 0 {
				t.Errorf("BaseSPD = %d, want > 0", s.BaseSPD)
			}
			if s.GrowthRate <= 0 {
				t.Errorf("GrowthRate = %f, want > 0", s.GrowthRate)
			}
			if s.MaxBeasts <= 0 {
				t.Errorf("MaxBeasts = %d, want > 0", s.MaxBeasts)
			}
		})
	}
}
