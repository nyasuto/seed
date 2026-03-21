package senju

import "github.com/nyasuto/seed/core/types"

// Species defines the base characteristics of a beast type.
// Each species has an elemental affinity and base combat stats that
// individual beasts inherit upon creation.
type Species struct {
	// ID is the unique identifier for this species (e.g., "suiryu", "enhou").
	ID string

	// Name is the display name (e.g., "翠龍", "炎鳳").
	Name string

	// Element is the elemental affinity of this species.
	Element types.Element

	// BaseHP is the starting hit points for beasts of this species.
	BaseHP int

	// BaseATK is the starting attack power.
	BaseATK int

	// BaseDEF is the starting defense power.
	BaseDEF int

	// BaseSPD is the starting speed.
	BaseSPD int

	// GrowthRate is a multiplier applied to EXP gain per tick.
	// Higher values mean faster leveling.
	GrowthRate float64

	// MaxBeasts is a hint for the maximum number of this species
	// that can be placed in a single room.
	MaxBeasts int

	// Description is a short flavor text for this species.
	Description string
}
