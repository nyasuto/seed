package world

import (
	"encoding/json"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// RoomType defines a template for a kind of room that can be placed in the cave.
type RoomType struct {
	// ID is a unique string identifier for this room type (e.g., "dragon_hole").
	ID string `json:"id"`
	// Name is the display name of the room type.
	Name string `json:"name"`
	// Element is the five-element attribute of this room type.
	Element types.Element `json:"element"`
	// BaseChiCapacity is the base amount of chi this room type can store.
	BaseChiCapacity int `json:"base_chi_capacity"`
	// Description is a short description of the room type.
	Description string `json:"description"`
	// MaxBeasts is the maximum number of beasts that can be placed in this room type.
	MaxBeasts int `json:"max_beasts"`
}

// RoomTypeRegistry holds a collection of room types indexed by ID.
type RoomTypeRegistry struct {
	types map[string]RoomType
}

// NewRoomTypeRegistry creates an empty RoomTypeRegistry.
func NewRoomTypeRegistry() *RoomTypeRegistry {
	return &RoomTypeRegistry{
		types: make(map[string]RoomType),
	}
}

// Register adds a room type to the registry.
// Returns an error if a room type with the same ID already exists.
func (r *RoomTypeRegistry) Register(rt RoomType) error {
	if rt.ID == "" {
		return fmt.Errorf("room type ID must not be empty")
	}
	if _, exists := r.types[rt.ID]; exists {
		return fmt.Errorf("room type %q already registered", rt.ID)
	}
	r.types[rt.ID] = rt
	return nil
}

// Get returns the room type with the given ID.
// Returns an error if the ID is not found.
func (r *RoomTypeRegistry) Get(id string) (RoomType, error) {
	rt, ok := r.types[id]
	if !ok {
		return RoomType{}, fmt.Errorf("room type %q not found", id)
	}
	return rt, nil
}

// All returns a slice of all registered room types.
func (r *RoomTypeRegistry) All() []RoomType {
	result := make([]RoomType, 0, len(r.types))
	for _, rt := range r.types {
		result = append(result, rt)
	}
	return result
}

// Len returns the number of registered room types.
func (r *RoomTypeRegistry) Len() int {
	return len(r.types)
}

// roomTypeJSON is the intermediate representation for JSON deserialization
// that uses a string for the element field.
type roomTypeJSON struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Element         string `json:"element"`
	BaseChiCapacity int    `json:"base_chi_capacity"`
	Description     string `json:"description"`
	MaxBeasts       int    `json:"max_beasts"`
}

// elementFromString converts a string name to a types.Element value.
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

// LoadRoomTypesJSON parses JSON data containing an array of room type definitions
// and returns a populated RoomTypeRegistry.
func LoadRoomTypesJSON(data []byte) (*RoomTypeRegistry, error) {
	var raws []roomTypeJSON
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("parsing room types JSON: %w", err)
	}

	reg := NewRoomTypeRegistry()
	for _, raw := range raws {
		elem, err := elementFromString(raw.Element)
		if err != nil {
			return nil, fmt.Errorf("room type %q: %w", raw.ID, err)
		}
		rt := RoomType{
			ID:              raw.ID,
			Name:            raw.Name,
			Element:         elem,
			BaseChiCapacity: raw.BaseChiCapacity,
			Description:     raw.Description,
			MaxBeasts:       raw.MaxBeasts,
		}
		if err := reg.Register(rt); err != nil {
			return nil, err
		}
	}
	return reg, nil
}
