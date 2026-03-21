package invasion

import (
	"encoding/json"
	"fmt"

	"github.com/nyasuto/seed/core/types"
)

// GoalType represents the kind of objective an invader pursues.
type GoalType int

const (
	// DestroyCore means the invader aims to destroy the dragon vein core room.
	DestroyCore GoalType = iota
	// HuntBeasts means the invader aims to hunt beasts in the cave.
	HuntBeasts
	// StealTreasure means the invader aims to steal treasure from storage rooms.
	StealTreasure
)

// String returns the name of the goal type.
func (g GoalType) String() string {
	switch g {
	case DestroyCore:
		return "DestroyCore"
	case HuntBeasts:
		return "HuntBeasts"
	case StealTreasure:
		return "StealTreasure"
	default:
		return "Unknown"
	}
}

// goalTypeFromString converts a string to a GoalType value.
func goalTypeFromString(s string) (GoalType, error) {
	switch s {
	case "DestroyCore":
		return DestroyCore, nil
	case "HuntBeasts":
		return HuntBeasts, nil
	case "StealTreasure":
		return StealTreasure, nil
	default:
		return 0, fmt.Errorf("unknown goal type %q", s)
	}
}

// InvaderClass defines a template for a kind of invader with base stats and behavior.
type InvaderClass struct {
	// ID is a unique string identifier for this invader class.
	ID string `json:"id"`
	// Name is the display name of the invader class.
	Name string `json:"name"`
	// Element is the five-element attribute of this invader class.
	Element types.Element `json:"element"`
	// BaseHP is the base hit points at level 1.
	BaseHP int `json:"base_hp"`
	// BaseATK is the base attack power at level 1.
	BaseATK int `json:"base_atk"`
	// BaseDEF is the base defense power at level 1.
	BaseDEF int `json:"base_def"`
	// BaseSPD is the base speed at level 1.
	BaseSPD int `json:"base_spd"`
	// RewardChi is the amount of chi gained when this invader is defeated.
	RewardChi float64 `json:"reward_chi"`
	// PreferredGoal is the default goal type this class pursues.
	PreferredGoal GoalType `json:"preferred_goal"`
	// RetreatThreshold is the HP ratio (0.0-1.0) at which this invader retreats.
	RetreatThreshold float64 `json:"retreat_threshold"`
	// Description is a short description of the invader class.
	Description string `json:"description"`
}

// InvaderClassRegistry holds a collection of invader classes indexed by ID.
type InvaderClassRegistry struct {
	classes map[string]InvaderClass
}

// NewInvaderClassRegistry creates an empty InvaderClassRegistry.
func NewInvaderClassRegistry() *InvaderClassRegistry {
	return &InvaderClassRegistry{
		classes: make(map[string]InvaderClass),
	}
}

// Register adds an invader class to the registry.
// Returns an error if a class with the same ID already exists.
func (r *InvaderClassRegistry) Register(ic InvaderClass) error {
	if ic.ID == "" {
		return fmt.Errorf("invader class ID must not be empty")
	}
	if _, exists := r.classes[ic.ID]; exists {
		return fmt.Errorf("invader class %q already registered", ic.ID)
	}
	r.classes[ic.ID] = ic
	return nil
}

// Get returns the invader class with the given ID.
// Returns an error if the ID is not found.
func (r *InvaderClassRegistry) Get(id string) (InvaderClass, error) {
	ic, ok := r.classes[id]
	if !ok {
		return InvaderClass{}, fmt.Errorf("invader class %q not found", id)
	}
	return ic, nil
}

// All returns a slice of all registered invader classes.
func (r *InvaderClassRegistry) All() []InvaderClass {
	result := make([]InvaderClass, 0, len(r.classes))
	for _, ic := range r.classes {
		result = append(result, ic)
	}
	return result
}

// Len returns the number of registered invader classes.
func (r *InvaderClassRegistry) Len() int {
	return len(r.classes)
}

// invaderClassJSON is the intermediate representation for JSON deserialization
// that uses strings for element and goal type fields.
type invaderClassJSON struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Element          string  `json:"element"`
	BaseHP           int     `json:"base_hp"`
	BaseATK          int     `json:"base_atk"`
	BaseDEF          int     `json:"base_def"`
	BaseSPD          int     `json:"base_spd"`
	RewardChi        float64 `json:"reward_chi"`
	PreferredGoal    string  `json:"preferred_goal"`
	RetreatThreshold float64 `json:"retreat_threshold"`
	Description      string  `json:"description"`
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

// LoadInvaderClassesJSON parses JSON data containing an array of invader class
// definitions and returns a populated InvaderClassRegistry.
func LoadInvaderClassesJSON(data []byte) (*InvaderClassRegistry, error) {
	var raws []invaderClassJSON
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("parsing invader classes JSON: %w", err)
	}

	reg := NewInvaderClassRegistry()
	for _, raw := range raws {
		elem, err := elementFromString(raw.Element)
		if err != nil {
			return nil, fmt.Errorf("invader class %q: %w", raw.ID, err)
		}
		goal, err := goalTypeFromString(raw.PreferredGoal)
		if err != nil {
			return nil, fmt.Errorf("invader class %q: %w", raw.ID, err)
		}
		ic := InvaderClass{
			ID:               raw.ID,
			Name:             raw.Name,
			Element:          elem,
			BaseHP:           raw.BaseHP,
			BaseATK:          raw.BaseATK,
			BaseDEF:          raw.BaseDEF,
			BaseSPD:          raw.BaseSPD,
			RewardChi:        raw.RewardChi,
			PreferredGoal:    goal,
			RetreatThreshold: raw.RetreatThreshold,
			Description:      raw.Description,
		}
		if err := reg.Register(ic); err != nil {
			return nil, err
		}
	}
	return reg, nil
}
