package senju

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestEvolutionRegistry_GetPaths(t *testing.T) {
	reg := NewEvolutionRegistry()
	path1 := EvolutionPath{
		FromSpeciesID: "suiryu",
		ToSpeciesID:   "souryu",
		Condition:     EvolutionCondition{MinLevel: 15},
		ChiCost:       100,
	}
	path2 := EvolutionPath{
		FromSpeciesID: "suiryu",
		ToSpeciesID:   "shinryu",
		Condition:     EvolutionCondition{MinLevel: 30},
		ChiCost:       200,
	}
	reg.Register(path1)
	reg.Register(path2)

	paths := reg.GetPaths("suiryu")
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}

	// Verify returned slice is a copy.
	paths[0].ChiCost = 999
	original := reg.GetPaths("suiryu")
	if original[0].ChiCost == 999 {
		t.Error("GetPaths should return a copy, not a reference")
	}

	// Non-existent species returns nil.
	if got := reg.GetPaths("nonexistent"); got != nil {
		t.Errorf("expected nil for unknown species, got %v", got)
	}
}

func TestEvolutionRegistry_CheckEvolution(t *testing.T) {
	reg := NewEvolutionRegistry()
	reg.Register(EvolutionPath{
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
		Element:   types.Wood,
		Level:     15,
	}

	tests := []struct {
		name           string
		level          int
		roomElement    types.Element
		roomChiRatio   float64
		chiPoolBalance float64
		wantMatch      bool
	}{
		{
			name:           "all conditions met",
			level:          15,
			roomElement:    types.Wood,
			roomChiRatio:   0.8,
			chiPoolBalance: 200,
			wantMatch:      true,
		},
		{
			name:           "level too low",
			level:          14,
			roomElement:    types.Wood,
			roomChiRatio:   0.8,
			chiPoolBalance: 200,
			wantMatch:      false,
		},
		{
			name:           "wrong room element",
			level:          15,
			roomElement:    types.Fire,
			roomChiRatio:   0.8,
			chiPoolBalance: 200,
			wantMatch:      false,
		},
		{
			name:           "chi ratio too low",
			level:          15,
			roomElement:    types.Wood,
			roomChiRatio:   0.3,
			chiPoolBalance: 200,
			wantMatch:      false,
		},
		{
			name:           "insufficient chi pool",
			level:          15,
			roomElement:    types.Wood,
			roomChiRatio:   0.8,
			chiPoolBalance: 50,
			wantMatch:      false,
		},
		{
			name:           "exact minimum values",
			level:          15,
			roomElement:    types.Wood,
			roomChiRatio:   0.5,
			chiPoolBalance: 100,
			wantMatch:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			beast.Level = tc.level
			result := reg.CheckEvolution(beast, tc.roomElement, tc.roomChiRatio, tc.chiPoolBalance)
			if tc.wantMatch && result == nil {
				t.Error("expected a matching path, got nil")
			}
			if !tc.wantMatch && result != nil {
				t.Errorf("expected nil, got %+v", result)
			}
			if tc.wantMatch && result != nil {
				if result.ToSpeciesID != "souryu" {
					t.Errorf("expected ToSpeciesID souryu, got %s", result.ToSpeciesID)
				}
			}
		})
	}

	// Beast with no evolution paths.
	otherBeast := &Beast{SpeciesID: "unknown", Level: 99}
	if got := reg.CheckEvolution(otherBeast, types.Wood, 1.0, 9999); got != nil {
		t.Errorf("expected nil for unknown species, got %+v", got)
	}
}

func TestEvolutionRegistry_CheckEvolution_NoElementRequirement(t *testing.T) {
	reg := NewEvolutionRegistry()
	reg.Register(EvolutionPath{
		FromSpeciesID: "enhou",
		ToSpeciesID:   "suzaku",
		Condition: EvolutionCondition{
			MinLevel: 15,
			// RequireElement is false — any room element is OK.
		},
		ChiCost: 0,
	})

	beast := &Beast{SpeciesID: "enhou", Level: 15}

	// Should match regardless of room element.
	for _, elem := range []types.Element{types.Wood, types.Fire, types.Earth, types.Metal, types.Water} {
		result := reg.CheckEvolution(beast, elem, 0.0, 0)
		if result == nil {
			t.Errorf("expected match for element %v, got nil", elem)
		}
	}
}

func TestLoadEvolutionJSON(t *testing.T) {
	data := []byte(`[
		{
			"from_species_id": "suiryu",
			"to_species_id": "souryu",
			"min_level": 15,
			"required_room_element": "Wood",
			"min_chi_ratio": 0.5,
			"chi_cost": 100
		},
		{
			"from_species_id": "enhou",
			"to_species_id": "suzaku",
			"min_level": 15,
			"chi_cost": 80
		}
	]`)

	reg, err := LoadEvolutionJSON(data)
	if err != nil {
		t.Fatalf("LoadEvolutionJSON failed: %v", err)
	}

	paths := reg.GetPaths("suiryu")
	if len(paths) != 1 {
		t.Fatalf("expected 1 path for suiryu, got %d", len(paths))
	}
	p := paths[0]
	if p.Condition.MinLevel != 15 {
		t.Errorf("expected MinLevel 15, got %d", p.Condition.MinLevel)
	}
	if !p.Condition.RequireElement {
		t.Error("expected RequireElement to be true")
	}
	if p.Condition.RequiredRoomElement != types.Wood {
		t.Errorf("expected Wood element, got %v", p.Condition.RequiredRoomElement)
	}

	paths2 := reg.GetPaths("enhou")
	if len(paths2) != 1 {
		t.Fatalf("expected 1 path for enhou, got %d", len(paths2))
	}
	if paths2[0].Condition.RequireElement {
		t.Error("expected RequireElement to be false when element not specified")
	}
}

func TestLoadEvolutionJSON_InvalidElement(t *testing.T) {
	data := []byte(`[{
		"from_species_id": "x",
		"to_species_id": "y",
		"min_level": 1,
		"required_room_element": "InvalidElement",
		"chi_cost": 0
	}]`)

	_, err := LoadEvolutionJSON(data)
	if err == nil {
		t.Fatal("expected error for invalid element")
	}
}

func TestLoadEvolutionJSON_InvalidJSON(t *testing.T) {
	_, err := LoadEvolutionJSON([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
