package invasion

import (
	"fmt"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// WaveConfig defines the parameters for generating a single invasion wave.
type WaveConfig struct {
	// TriggerTick is the tick at which the wave becomes active.
	TriggerTick types.Tick `json:"trigger_tick"`
	// Difficulty is the difficulty multiplier affecting invader levels.
	Difficulty float64 `json:"difficulty"`
	// MinInvaders is the minimum number of invaders in the wave.
	MinInvaders int `json:"min_invaders"`
	// MaxInvaders is the maximum number of invaders in the wave.
	MaxInvaders int `json:"max_invaders"`
}

// WaveGenerator generates invasion waves based on configuration and cave layout.
type WaveGenerator struct {
	classRegistry *InvaderClassRegistry
	rng           types.RNG
	nextWaveID    int
	nextInvaderID int
}

// NewWaveGenerator creates a new WaveGenerator with the given class registry and RNG.
func NewWaveGenerator(classRegistry *InvaderClassRegistry, rng types.RNG) *WaveGenerator {
	return &WaveGenerator{
		classRegistry: classRegistry,
		rng:           rng,
		nextWaveID:    1,
		nextInvaderID: 1,
	}
}

// SetNextWaveID sets the next wave ID to be assigned.
func (wg *WaveGenerator) SetNextWaveID(id int) {
	wg.nextWaveID = id
}

// NextWaveID returns the next wave ID that will be assigned.
func (wg *WaveGenerator) NextWaveID() int {
	return wg.nextWaveID
}

// GenerateWave creates an InvasionWave based on the given config and cave layout.
// The number of invaders is randomly chosen between MinInvaders and MaxInvaders.
// Goals are assigned based on the cave's room composition:
//   - If a dragon_hole room exists, DestroyCore is the primary goal
//   - If storage rooms exist and a thief class is available, some invaders get StealTreasure
//   - Otherwise, invaders use their class's preferred goal
func (wg *WaveGenerator) GenerateWave(config WaveConfig, cave *world.Cave, tick types.Tick) (*InvasionWave, error) {
	if config.MinInvaders <= 0 || config.MaxInvaders <= 0 {
		return nil, fmt.Errorf("min and max invaders must be positive")
	}
	if config.MinInvaders > config.MaxInvaders {
		return nil, fmt.Errorf("min invaders (%d) must not exceed max invaders (%d)", config.MinInvaders, config.MaxInvaders)
	}

	allClasses := wg.classRegistry.All()
	if len(allClasses) == 0 {
		return nil, fmt.Errorf("no invader classes registered")
	}

	// Determine invader count.
	count := config.MinInvaders
	if config.MaxInvaders > config.MinInvaders {
		count = config.MinInvaders + wg.rng.Intn(config.MaxInvaders-config.MinInvaders+1)
	}

	// Analyze cave layout for goal assignment.
	hasCoreRoom := false
	hasStorageRoom := false
	for _, room := range cave.Rooms {
		if room.TypeID == "dragon_hole" {
			hasCoreRoom = true
		}
		if room.TypeID == "storage" {
			hasStorageRoom = true
		}
	}

	// Determine entry room (first room in the cave).
	entryRoomID := 0
	if len(cave.Rooms) > 0 {
		entryRoomID = cave.Rooms[0].ID
	}

	// Calculate invader level from difficulty.
	level := max(int(config.Difficulty+0.5), 1)

	invaders := make([]*Invader, 0, count)
	for i := 0; i < count; i++ {
		// Pick a random class.
		class := allClasses[wg.rng.Intn(len(allClasses))]

		// Assign goal based on cave layout and class preference.
		goal := wg.assignGoal(class, hasCoreRoom, hasStorageRoom)

		inv := NewInvader(wg.nextInvaderID, class, level, goal, entryRoomID, tick)
		wg.nextInvaderID++
		invaders = append(invaders, inv)
	}

	wave := &InvasionWave{
		ID:          wg.nextWaveID,
		TriggerTick: config.TriggerTick,
		Invaders:    invaders,
		State:       Pending,
		Difficulty:  config.Difficulty,
	}
	wg.nextWaveID++

	return wave, nil
}

// assignGoal determines the appropriate goal for an invader based on cave layout.
// If a core room exists, DestroyCore is the primary goal regardless of class preference.
// If the class prefers StealTreasure but no storage rooms exist, falls back to DestroyCore
// (or HuntBeasts if no core room either).
func (wg *WaveGenerator) assignGoal(class InvaderClass, hasCoreRoom, hasStorageRoom bool) Goal {
	preferred := class.PreferredGoal

	switch preferred {
	case StealTreasure:
		if hasStorageRoom {
			return NewStealTreasureGoal()
		}
		// No storage room — fall back.
		if hasCoreRoom {
			return NewDestroyCoreGoal()
		}
		return NewHuntBeastsGoal()

	case HuntBeasts:
		return NewHuntBeastsGoal()

	case DestroyCore:
		if hasCoreRoom {
			return NewDestroyCoreGoal()
		}
		// No core room — fall back to hunting beasts.
		return NewHuntBeastsGoal()

	default:
		// Unknown preference — use DestroyCore if possible, else HuntBeasts.
		if hasCoreRoom {
			return NewDestroyCoreGoal()
		}
		return NewHuntBeastsGoal()
	}
}
