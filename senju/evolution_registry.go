package senju

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ponpoko/chaosseed-core/types"
)

// EvolutionRegistry manages evolution paths for all species.
type EvolutionRegistry struct {
	// paths maps FromSpeciesID to a list of possible evolution paths.
	paths map[string][]EvolutionPath
}

// NewEvolutionRegistry creates an empty EvolutionRegistry.
func NewEvolutionRegistry() *EvolutionRegistry {
	return &EvolutionRegistry{
		paths: make(map[string][]EvolutionPath),
	}
}

// Register adds an evolution path to the registry.
func (r *EvolutionRegistry) Register(path EvolutionPath) {
	r.paths[path.FromSpeciesID] = append(r.paths[path.FromSpeciesID], path)
}

// GetPaths returns all evolution paths available for the given species ID.
// Returns nil if no paths exist.
func (r *EvolutionRegistry) GetPaths(speciesID string) []EvolutionPath {
	paths := r.paths[speciesID]
	if len(paths) == 0 {
		return nil
	}
	// Return a copy to prevent mutation.
	result := make([]EvolutionPath, len(paths))
	copy(result, paths)
	return result
}

// CheckEvolution checks whether the given beast meets the conditions for any
// evolution path. It evaluates room element, chi ratio, and chi pool balance.
// Returns the first matching path, or nil if no conditions are met.
//
// Parameters:
//   - beast: the beast to check for evolution eligibility
//   - roomElement: the element of the room the beast is in
//   - roomChiRatio: the chi fill ratio (0.0–1.0) of the beast's room
//   - chiPoolBalance: the current chi pool balance available for evolution cost
func (r *EvolutionRegistry) CheckEvolution(beast *Beast, roomElement types.Element, roomChiRatio float64, chiPoolBalance float64) *EvolutionPath {
	paths := r.paths[beast.SpeciesID]
	for i := range paths {
		p := &paths[i]
		c := p.Condition

		// Check minimum level.
		if beast.Level < c.MinLevel {
			continue
		}

		// Check room element requirement.
		if c.RequireElement && roomElement != c.RequiredRoomElement {
			continue
		}

		// Check minimum chi ratio.
		if c.MinChiRatio > 0 && roomChiRatio < c.MinChiRatio {
			continue
		}

		// Check chi cost affordability.
		if p.ChiCost > 0 && chiPoolBalance < p.ChiCost {
			continue
		}

		// All conditions met.
		result := paths[i]
		return &result
	}
	return nil
}

// evolutionPathJSON is the intermediate representation for JSON deserialization.
type evolutionPathJSON struct {
	FromSpeciesID       string  `json:"from_species_id"`
	ToSpeciesID         string  `json:"to_species_id"`
	MinLevel            int     `json:"min_level"`
	RequiredRoomElement string  `json:"required_room_element,omitempty"`
	MinChiRatio         float64 `json:"min_chi_ratio,omitempty"`
	ChiCost             float64 `json:"chi_cost"`
}

// LoadEvolutionJSON loads evolution path definitions from JSON data.
func LoadEvolutionJSON(data []byte) (*EvolutionRegistry, error) {
	var items []evolutionPathJSON
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("unmarshal evolution json: %w", err)
	}

	reg := NewEvolutionRegistry()
	for _, item := range items {
		cond := EvolutionCondition{
			MinLevel:    item.MinLevel,
			MinChiRatio: item.MinChiRatio,
		}

		if item.RequiredRoomElement != "" {
			elem, err := elementFromString(item.RequiredRoomElement)
			if err != nil {
				return nil, fmt.Errorf("evolution %s->%s: %w", item.FromSpeciesID, item.ToSpeciesID, err)
			}
			cond.RequiredRoomElement = elem
			cond.RequireElement = true
		}

		reg.Register(EvolutionPath{
			FromSpeciesID: item.FromSpeciesID,
			ToSpeciesID:   item.ToSpeciesID,
			Condition:     cond,
			ChiCost:       item.ChiCost,
		})
	}

	// Sort paths per species for deterministic ordering.
	for key := range reg.paths {
		sort.Slice(reg.paths[key], func(i, j int) bool {
			return reg.paths[key][i].ToSpeciesID < reg.paths[key][j].ToSpeciesID
		})
	}

	return reg, nil
}
