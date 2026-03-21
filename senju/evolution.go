package senju

import (
	"errors"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// ErrEvolutionTargetNotFound is returned when the target species of an
// evolution path is not in the species registry.
var ErrEvolutionTargetNotFound = errors.New("evolution target species not found")

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

// Evolve performs the evolution of a beast along the given path.
// It changes the beast's species, updates its element, and recalculates
// stats based on the new species' base values while preserving the current level.
// The beast's HP is fully restored after evolution.
func Evolve(beast *Beast, path *EvolutionPath, speciesRegistry *SpeciesRegistry) error {
	newSpecies, err := speciesRegistry.Get(path.ToSpeciesID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrEvolutionTargetNotFound, path.ToSpeciesID)
	}

	beast.SpeciesID = newSpecies.ID
	beast.Name = newSpecies.Name
	beast.Element = newSpecies.Element

	// Recalculate stats using the same formula as level-up (growth.go).
	beast.MaxHP = newSpecies.BaseHP + (beast.Level-1)*2
	beast.HP = beast.MaxHP
	beast.ATK = newSpecies.BaseATK + (beast.Level-1)*1
	beast.DEF = newSpecies.BaseDEF + (beast.Level-1)*1
	beast.SPD = newSpecies.BaseSPD + (beast.Level-1)*1

	return nil
}
