package senju

import (
	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/world"
)

// GrowthEventType represents the type of growth event that occurred.
type GrowthEventType int

const (
	// EXPGained indicates the beast gained experience points.
	EXPGained GrowthEventType = iota
	// LevelUp indicates the beast leveled up.
	LevelUp
	// ChiStarved indicates the beast could not grow due to insufficient chi.
	ChiStarved
)

// String returns the name of the growth event type.
func (t GrowthEventType) String() string {
	switch t {
	case EXPGained:
		return "EXPGained"
	case LevelUp:
		return "LevelUp"
	case ChiStarved:
		return "ChiStarved"
	default:
		return "Unknown"
	}
}

// GrowthEvent records a growth-related event for a beast during a tick.
type GrowthEvent struct {
	// BeastID is the ID of the beast this event belongs to.
	BeastID int
	// Type is the kind of growth event.
	Type GrowthEventType
	// OldLevel is the beast's level before this event (relevant for LevelUp).
	OldLevel int
	// NewLevel is the beast's level after this event (relevant for LevelUp).
	NewLevel int
	// EXPGained is the amount of EXP gained in this event.
	EXPGained int
}

// GrowthEngine handles beast growth and leveling mechanics.
type GrowthEngine struct {
	params          *GrowthParams
	speciesRegistry *SpeciesRegistry
}

// NewGrowthEngine creates a new GrowthEngine with the given parameters
// and species registry.
func NewGrowthEngine(params *GrowthParams, speciesRegistry *SpeciesRegistry) *GrowthEngine {
	return &GrowthEngine{
		params:          params,
		speciesRegistry: speciesRegistry,
	}
}

// Tick processes one tick of growth for all beasts. It consumes chi from
// rooms, awards EXP, and handles level-ups. Returns a list of growth events.
func (ge *GrowthEngine) Tick(beasts []*Beast, roomChi map[int]*fengshui.RoomChi, rooms map[int]*world.Room) []GrowthEvent {
	var events []GrowthEvent

	for _, beast := range beasts {
		if beast.RoomID == 0 {
			continue
		}
		if beast.Level >= ge.params.MaxLevel {
			continue
		}

		chi, ok := roomChi[beast.RoomID]
		if !ok {
			events = append(events, GrowthEvent{
				BeastID: beast.ID,
				Type:    ChiStarved,
			})
			continue
		}

		// Try to consume chi
		if chi.Current < ge.params.ChiConsumptionPerTick {
			events = append(events, GrowthEvent{
				BeastID: beast.ID,
				Type:    ChiStarved,
			})
			continue
		}
		chi.Current -= ge.params.ChiConsumptionPerTick

		// Calculate EXP gain
		species, err := ge.speciesRegistry.Get(beast.SpeciesID)
		if err != nil {
			continue
		}

		affinity := GrowthAffinity(beast.Element, chi.Element)
		expGain := max(int(float64(ge.params.BaseEXPPerTick)*affinity*species.GrowthRate), 1)

		beast.EXP += expGain
		events = append(events, GrowthEvent{
			BeastID:   beast.ID,
			Type:      EXPGained,
			EXPGained: expGain,
		})

		// Check for level up
		threshold := ge.params.LevelUpThreshold(beast.Level)
		for beast.EXP >= threshold && beast.Level < ge.params.MaxLevel {
			oldLevel := beast.Level
			beast.Level++
			beast.EXP -= threshold

			// Recalculate stats on level up
			beast.MaxHP = species.BaseHP + (beast.Level-1)*2
			beast.HP = beast.MaxHP
			beast.ATK = species.BaseATK + (beast.Level-1)*1
			beast.DEF = species.BaseDEF + (beast.Level-1)*1
			beast.SPD = species.BaseSPD + (beast.Level-1)*1

			events = append(events, GrowthEvent{
				BeastID:  beast.ID,
				Type:     LevelUp,
				OldLevel: oldLevel,
				NewLevel: beast.Level,
			})

			threshold = ge.params.LevelUpThreshold(beast.Level)
		}
	}

	return events
}
