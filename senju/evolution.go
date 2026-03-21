package senju

import "github.com/ponpoko/chaosseed-core/types"

// EvolutionCondition defines the requirements for a beast to evolve.
// All non-zero conditions must be met simultaneously.
type EvolutionCondition struct {
	// MinLevel is the minimum beast level required for evolution.
	MinLevel int

	// RequiredRoomElement is the element the beast's room must have.
	// A zero value means no element restriction (any room element is acceptable).
	// Use HasElementRequirement to check if this constraint is active.
	RequiredRoomElement types.Element

	// RequireElement indicates whether RequiredRoomElement is an active constraint.
	// This is needed because the zero value of Element (Wood) is a valid element.
	RequireElement bool

	// MinChiRatio is the minimum chi fill ratio (0.0–1.0) the room must have.
	// A zero value means no chi ratio requirement.
	MinChiRatio float64
}

// EvolutionPath defines a single evolution route from one species to another.
type EvolutionPath struct {
	// FromSpeciesID is the species that can evolve.
	FromSpeciesID string

	// ToSpeciesID is the species it evolves into.
	ToSpeciesID string

	// Condition defines the requirements that must be met for this evolution.
	Condition EvolutionCondition

	// ChiCost is the amount of chi consumed from the economy pool to perform the evolution.
	ChiCost float64
}
