package senju

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/ponpoko/chaosseed-core/types"
)

// ErrSpeciesNotFound is returned when a species ID is not in the registry.
var ErrSpeciesNotFound = errors.New("species not found")

// SpeciesRegistry manages a collection of species definitions.
type SpeciesRegistry struct {
	species map[string]*Species
}

// NewSpeciesRegistry creates an empty SpeciesRegistry.
func NewSpeciesRegistry() *SpeciesRegistry {
	return &SpeciesRegistry{
		species: make(map[string]*Species),
	}
}

// Register adds a species to the registry. Returns an error if the ID
// is already registered.
func (r *SpeciesRegistry) Register(s *Species) error {
	if _, exists := r.species[s.ID]; exists {
		return fmt.Errorf("species %q already registered", s.ID)
	}
	r.species[s.ID] = s
	return nil
}

// Get returns the species with the given ID, or ErrSpeciesNotFound.
func (r *SpeciesRegistry) Get(id string) (*Species, error) {
	s, ok := r.species[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSpeciesNotFound, id)
	}
	return s, nil
}

// All returns all registered species sorted by ID for deterministic ordering.
func (r *SpeciesRegistry) All() []*Species {
	result := make([]*Species, 0, len(r.species))
	for _, s := range r.species {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// Len returns the number of registered species.
func (r *SpeciesRegistry) Len() int {
	return len(r.species)
}

// speciesJSON is the intermediate representation for JSON deserialization.
type speciesJSON struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Element     string  `json:"element"`
	BaseHP      int     `json:"base_hp"`
	BaseATK     int     `json:"base_atk"`
	BaseDEF     int     `json:"base_def"`
	BaseSPD     int     `json:"base_spd"`
	GrowthRate  float64 `json:"growth_rate"`
	MaxBeasts   int     `json:"max_beasts"`
	Description string  `json:"description"`
}

// LoadSpeciesJSON loads species definitions from JSON data into a new registry.
func LoadSpeciesJSON(data []byte) (*SpeciesRegistry, error) {
	var items []speciesJSON
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("unmarshal species json: %w", err)
	}

	reg := NewSpeciesRegistry()
	for _, item := range items {
		elem, err := elementFromString(item.Element)
		if err != nil {
			return nil, fmt.Errorf("species %q: %w", item.ID, err)
		}
		s := &Species{
			ID:          item.ID,
			Name:        item.Name,
			Element:     elem,
			BaseHP:      item.BaseHP,
			BaseATK:     item.BaseATK,
			BaseDEF:     item.BaseDEF,
			BaseSPD:     item.BaseSPD,
			GrowthRate:  item.GrowthRate,
			MaxBeasts:   item.MaxBeasts,
			Description: item.Description,
		}
		if err := reg.Register(s); err != nil {
			return nil, err
		}
	}
	return reg, nil
}

// elementFromString converts a string element name to a types.Element value.
func elementFromString(s string) (types.Element, error) {
	switch s {
	case "Wood":
		return types.Wood, nil
	case "Fire":
		return types.Fire, nil
	case "Earth":
		return types.Earth, nil
	case "Metal":
		return types.Metal, nil
	case "Water":
		return types.Water, nil
	default:
		return 0, fmt.Errorf("unknown element %q", s)
	}
}
