package senju

import (
	"encoding/json"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// jsonBeast is the JSON representation of a Beast.
type jsonBeast struct {
	ID        int         `json:"id"`
	SpeciesID string      `json:"species_id"`
	Name      string      `json:"name"`
	Element   string      `json:"element"`
	RoomID    int         `json:"room_id"`
	Level     int         `json:"level"`
	EXP       int         `json:"exp"`
	HP        int         `json:"hp"`
	MaxHP     int         `json:"max_hp"`
	ATK       int         `json:"atk"`
	DEF       int         `json:"def"`
	SPD       int         `json:"spd"`
	BornTick  types.Tick  `json:"born_tick"`
	State     BeastState  `json:"state"`
}

// MarshalBeasts serializes a slice of beasts to JSON.
func MarshalBeasts(beasts []*Beast) ([]byte, error) {
	jbeasts := make([]jsonBeast, len(beasts))
	for i, b := range beasts {
		jbeasts[i] = jsonBeast{
			ID:        b.ID,
			SpeciesID: b.SpeciesID,
			Name:      b.Name,
			Element:   b.Element.String(),
			RoomID:    b.RoomID,
			Level:     b.Level,
			EXP:       b.EXP,
			HP:        b.HP,
			MaxHP:     b.MaxHP,
			ATK:       b.ATK,
			DEF:       b.DEF,
			SPD:       b.SPD,
			BornTick:  b.BornTick,
			State:     b.State,
		}
	}
	return json.Marshal(jbeasts)
}

// UnmarshalBeasts restores beasts from JSON data. The speciesRegistry is used
// to validate that each beast's SpeciesID exists. Element and Name are restored
// from the serialized data rather than the registry, preserving any runtime
// modifications.
func UnmarshalBeasts(data []byte, speciesRegistry *SpeciesRegistry) ([]*Beast, error) {
	var jbeasts []jsonBeast
	if err := json.Unmarshal(data, &jbeasts); err != nil {
		return nil, fmt.Errorf("unmarshalling beasts: %w", err)
	}

	beasts := make([]*Beast, len(jbeasts))
	for i, jb := range jbeasts {
		// Validate that species exists in registry.
		if _, err := speciesRegistry.Get(jb.SpeciesID); err != nil {
			return nil, fmt.Errorf("beast %d (id=%d): %w", i, jb.ID, err)
		}

		elem, err := elementFromString(jb.Element)
		if err != nil {
			return nil, fmt.Errorf("beast %d (id=%d): %w", i, jb.ID, err)
		}

		beasts[i] = &Beast{
			ID:        jb.ID,
			SpeciesID: jb.SpeciesID,
			Name:      jb.Name,
			Element:   elem,
			RoomID:    jb.RoomID,
			Level:     jb.Level,
			EXP:       jb.EXP,
			HP:        jb.HP,
			MaxHP:     jb.MaxHP,
			ATK:       jb.ATK,
			DEF:       jb.DEF,
			SPD:       jb.SPD,
			BornTick:  jb.BornTick,
			State:     jb.State,
		}
	}
	return beasts, nil
}
