package invasion

import (
	"encoding/json"
	"fmt"

	"github.com/nyasuto/seed/core/types"
)

// jsonExplorationMemory is the JSON representation of ExplorationMemory.
type jsonExplorationMemory struct {
	VisitedRooms       map[string]types.Tick `json:"visited_rooms"`
	KnownBeastRooms    map[string]bool       `json:"known_beast_rooms"`
	KnownCoreRoom      int                   `json:"known_core_room"`
	KnownTreasureRooms []int                 `json:"known_treasure_rooms"`
}

// jsonGoal is the JSON representation of a Goal, using a type discriminator.
type jsonGoal struct {
	Type              string `json:"type"`
	RequiredStayTicks int    `json:"required_stay_ticks,omitempty"`
	RequiredKills     int    `json:"required_kills,omitempty"`
	Kills             int    `json:"kills,omitempty"`
}

// jsonInvader is the JSON representation of an Invader.
type jsonInvader struct {
	ID            int                    `json:"id"`
	ClassID       string                 `json:"class_id"`
	Name          string                 `json:"name"`
	Element       string                 `json:"element"`
	Level         int                    `json:"level"`
	HP            int                    `json:"hp"`
	MaxHP         int                    `json:"max_hp"`
	ATK           int                    `json:"atk"`
	DEF           int                    `json:"def"`
	SPD           int                    `json:"spd"`
	CurrentRoomID int                    `json:"current_room_id"`
	Goal          jsonGoal               `json:"goal"`
	Memory        *jsonExplorationMemory `json:"memory"`
	State         string                 `json:"state"`
	SlowTicks     int                    `json:"slow_ticks"`
	EntryTick     types.Tick             `json:"entry_tick"`
	StayTicks     int                    `json:"stay_ticks"`
}

// jsonInvasionWave is the JSON representation of an InvasionWave.
type jsonInvasionWave struct {
	ID          int           `json:"id"`
	TriggerTick types.Tick    `json:"trigger_tick"`
	Invaders    []jsonInvader `json:"invaders"`
	State       string        `json:"state"`
	Difficulty  float64       `json:"difficulty"`
}

// invaderStateFromString converts a string to an InvaderState value.
func invaderStateFromString(s string) (InvaderState, error) {
	switch s {
	case "Advancing":
		return Advancing, nil
	case "Fighting":
		return Fighting, nil
	case "Retreating":
		return Retreating, nil
	case "Defeated":
		return Defeated, nil
	case "GoalAchieved":
		return GoalAchieved, nil
	default:
		return 0, fmt.Errorf("unknown invader state %q", s)
	}
}

// waveStateFromString converts a string to a WaveState value.
func waveStateFromString(s string) (WaveState, error) {
	switch s {
	case "Pending":
		return Pending, nil
	case "Active":
		return Active, nil
	case "Completed":
		return Completed, nil
	case "Failed":
		return Failed, nil
	default:
		return 0, fmt.Errorf("unknown wave state %q", s)
	}
}

// marshalGoal converts a Goal interface to its JSON representation.
func marshalGoal(g Goal) jsonGoal {
	jg := jsonGoal{Type: g.Type().String()}
	switch v := g.(type) {
	case *DestroyCoreGoal:
		jg.RequiredStayTicks = v.RequiredStayTicks
	case *HuntBeastsGoal:
		jg.RequiredKills = v.RequiredKills
		jg.Kills = v.Kills
	}
	return jg
}

// unmarshalGoal reconstructs a Goal from its JSON representation.
func unmarshalGoal(jg jsonGoal) (Goal, error) {
	switch jg.Type {
	case "DestroyCore":
		return &DestroyCoreGoal{RequiredStayTicks: jg.RequiredStayTicks}, nil
	case "HuntBeasts":
		return &HuntBeastsGoal{RequiredKills: jg.RequiredKills, Kills: jg.Kills}, nil
	case "StealTreasure":
		return &StealTreasureGoal{}, nil
	default:
		return nil, fmt.Errorf("unknown goal type %q", jg.Type)
	}
}

// marshalMemory converts an ExplorationMemory to its JSON representation.
// Map keys are converted from int to string for JSON compatibility.
func marshalMemory(m *ExplorationMemory) *jsonExplorationMemory {
	if m == nil {
		return nil
	}
	visited := make(map[string]types.Tick, len(m.VisitedRooms))
	for k, v := range m.VisitedRooms {
		visited[fmt.Sprintf("%d", k)] = v
	}
	beasts := make(map[string]bool, len(m.KnownBeastRooms))
	for k, v := range m.KnownBeastRooms {
		beasts[fmt.Sprintf("%d", k)] = v
	}
	treasures := make([]int, len(m.KnownTreasureRooms))
	copy(treasures, m.KnownTreasureRooms)
	return &jsonExplorationMemory{
		VisitedRooms:       visited,
		KnownBeastRooms:    beasts,
		KnownCoreRoom:      m.KnownCoreRoom,
		KnownTreasureRooms: treasures,
	}
}

// unmarshalMemory reconstructs an ExplorationMemory from its JSON representation.
func unmarshalMemory(jm *jsonExplorationMemory) (*ExplorationMemory, error) {
	if jm == nil {
		return NewExplorationMemory(), nil
	}
	visited := make(map[int]types.Tick, len(jm.VisitedRooms))
	for k, v := range jm.VisitedRooms {
		var id int
		if _, err := fmt.Sscanf(k, "%d", &id); err != nil {
			return nil, fmt.Errorf("invalid visited room key %q: %w", k, err)
		}
		visited[id] = v
	}
	beasts := make(map[int]bool, len(jm.KnownBeastRooms))
	for k, v := range jm.KnownBeastRooms {
		var id int
		if _, err := fmt.Sscanf(k, "%d", &id); err != nil {
			return nil, fmt.Errorf("invalid beast room key %q: %w", k, err)
		}
		beasts[id] = v
	}
	treasures := make([]int, len(jm.KnownTreasureRooms))
	copy(treasures, jm.KnownTreasureRooms)
	return &ExplorationMemory{
		VisitedRooms:       visited,
		KnownBeastRooms:    beasts,
		KnownCoreRoom:      jm.KnownCoreRoom,
		KnownTreasureRooms: treasures,
	}, nil
}

// MarshalInvasionState serializes a slice of invasion waves to JSON.
func MarshalInvasionState(waves []*InvasionWave) ([]byte, error) {
	jwaves := make([]jsonInvasionWave, len(waves))
	for i, w := range waves {
		jinvaders := make([]jsonInvader, len(w.Invaders))
		for j, inv := range w.Invaders {
			jinvaders[j] = jsonInvader{
				ID:            inv.ID,
				ClassID:       inv.ClassID,
				Name:          inv.Name,
				Element:       inv.Element.String(),
				Level:         inv.Level,
				HP:            inv.HP,
				MaxHP:         inv.MaxHP,
				ATK:           inv.ATK,
				DEF:           inv.DEF,
				SPD:           inv.SPD,
				CurrentRoomID: inv.CurrentRoomID,
				Goal:          marshalGoal(inv.Goal),
				Memory:        marshalMemory(inv.Memory),
				State:         inv.State.String(),
				SlowTicks:     inv.SlowTicks,
				EntryTick:     inv.EntryTick,
				StayTicks:     inv.StayTicks,
			}
		}
		jwaves[i] = jsonInvasionWave{
			ID:          w.ID,
			TriggerTick: w.TriggerTick,
			Invaders:    jinvaders,
			State:       w.State.String(),
			Difficulty:  w.Difficulty,
		}
	}
	return json.Marshal(jwaves)
}

// UnmarshalInvasionState restores invasion waves from JSON data.
// The classRegistry is used to validate that each invader's ClassID exists.
func UnmarshalInvasionState(data []byte, classRegistry *InvaderClassRegistry) ([]*InvasionWave, error) {
	var jwaves []jsonInvasionWave
	if err := json.Unmarshal(data, &jwaves); err != nil {
		return nil, fmt.Errorf("unmarshalling invasion state: %w", err)
	}

	waves := make([]*InvasionWave, len(jwaves))
	for i, jw := range jwaves {
		wState, err := waveStateFromString(jw.State)
		if err != nil {
			return nil, fmt.Errorf("wave %d (id=%d): %w", i, jw.ID, err)
		}

		invaders := make([]*Invader, len(jw.Invaders))
		for j, ji := range jw.Invaders {
			if _, err := classRegistry.Get(ji.ClassID); err != nil {
				return nil, fmt.Errorf("wave %d invader %d (id=%d): %w", i, j, ji.ID, err)
			}

			elem, err := elementFromString(ji.Element)
			if err != nil {
				return nil, fmt.Errorf("wave %d invader %d (id=%d): %w", i, j, ji.ID, err)
			}

			iState, err := invaderStateFromString(ji.State)
			if err != nil {
				return nil, fmt.Errorf("wave %d invader %d (id=%d): %w", i, j, ji.ID, err)
			}

			goal, err := unmarshalGoal(ji.Goal)
			if err != nil {
				return nil, fmt.Errorf("wave %d invader %d (id=%d): %w", i, j, ji.ID, err)
			}

			memory, err := unmarshalMemory(ji.Memory)
			if err != nil {
				return nil, fmt.Errorf("wave %d invader %d (id=%d): %w", i, j, ji.ID, err)
			}

			invaders[j] = &Invader{
				ID:            ji.ID,
				ClassID:       ji.ClassID,
				Name:          ji.Name,
				Element:       elem,
				Level:         ji.Level,
				HP:            ji.HP,
				MaxHP:         ji.MaxHP,
				ATK:           ji.ATK,
				DEF:           ji.DEF,
				SPD:           ji.SPD,
				CurrentRoomID: ji.CurrentRoomID,
				Goal:          goal,
				Memory:        memory,
				State:         iState,
				SlowTicks:     ji.SlowTicks,
				EntryTick:     ji.EntryTick,
				StayTicks:     ji.StayTicks,
			}
		}

		waves[i] = &InvasionWave{
			ID:          jw.ID,
			TriggerTick: jw.TriggerTick,
			Invaders:    invaders,
			State:       wState,
			Difficulty:  jw.Difficulty,
		}
	}
	return waves, nil
}
