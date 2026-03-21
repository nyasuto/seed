package scenario

import (
	"encoding/json"
	"fmt"

	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// jsonPos is the JSON representation of types.Pos.
type jsonPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// jsonRoomPlacement is the JSON representation of RoomPlacement.
type jsonRoomPlacement struct {
	TypeID string  `json:"type_id"`
	Pos    jsonPos `json:"pos"`
	Level  int     `json:"level"`
}

// jsonDragonVeinPlacement is the JSON representation of DragonVeinPlacement.
type jsonDragonVeinPlacement struct {
	SourcePos jsonPos `json:"source_pos"`
	Element   string  `json:"element"`
	FlowRate  float64 `json:"flow_rate"`
}

// jsonBeastPlacement is the JSON representation of BeastPlacement.
type jsonBeastPlacement struct {
	SpeciesID string `json:"species_id"`
	RoomIndex int    `json:"room_index"`
}

// jsonInitialState is the JSON representation of InitialState.
type jsonInitialState struct {
	CaveWidth      int                       `json:"cave_width"`
	CaveHeight     int                       `json:"cave_height"`
	TerrainSeed    int64                     `json:"terrain_seed"`
	TerrainDensity float64                   `json:"terrain_density"`
	PrebuiltRooms  []jsonRoomPlacement       `json:"prebuilt_rooms"`
	DragonVeins    []jsonDragonVeinPlacement `json:"dragon_veins"`
	StartingChi    float64                   `json:"starting_chi"`
	StartingBeasts []jsonBeastPlacement      `json:"starting_beasts"`
}

// jsonConditionDef is the JSON representation of ConditionDef.
type jsonConditionDef struct {
	Type   string         `json:"type"`
	Params map[string]any `json:"params,omitempty"`
}

// jsonEventDef is the JSON representation of EventDef.
type jsonEventDef struct {
	ID        string           `json:"id"`
	Condition jsonConditionDef `json:"condition"`
	Commands  []jsonCommandDef `json:"commands"`
	OneShot   bool             `json:"one_shot"`
}

// jsonCommandDef is the JSON representation of CommandDef.
type jsonCommandDef struct {
	Type   string         `json:"type"`
	Params map[string]any `json:"params,omitempty"`
}

// jsonGameConstraints is the JSON representation of GameConstraints.
type jsonGameConstraints struct {
	MaxRooms           int      `json:"max_rooms"`
	MaxBeasts          int      `json:"max_beasts"`
	MaxTicks           uint64   `json:"max_ticks"`
	ForbiddenRoomTypes []string `json:"forbidden_room_types"`
}

// jsonScenario is the top-level JSON representation of Scenario.
type jsonScenario struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Difficulty     string              `json:"difficulty"`
	InitialState   jsonInitialState    `json:"initial_state"`
	WinConditions  []jsonConditionDef  `json:"win_conditions"`
	LoseConditions []jsonConditionDef  `json:"lose_conditions"`
	WaveSchedule   []WaveScheduleEntry `json:"wave_schedule"`
	Events         []jsonEventDef      `json:"events"`
	Constraints    jsonGameConstraints `json:"constraints"`
}

// LoadScenario parses JSON data into a Scenario.
func LoadScenario(data []byte) (*Scenario, error) {
	var raw jsonScenario
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing scenario JSON: %w", err)
	}

	initialState, err := convertInitialState(raw.InitialState)
	if err != nil {
		return nil, fmt.Errorf("initial_state: %w", err)
	}

	s := &Scenario{
		ID:           raw.ID,
		Name:         raw.Name,
		Description:  raw.Description,
		Difficulty:   raw.Difficulty,
		InitialState: initialState,
		WaveSchedule: raw.WaveSchedule,
		Constraints: GameConstraints{
			MaxRooms:           raw.Constraints.MaxRooms,
			MaxBeasts:          raw.Constraints.MaxBeasts,
			MaxTicks:           types.Tick(raw.Constraints.MaxTicks),
			ForbiddenRoomTypes: raw.Constraints.ForbiddenRoomTypes,
		},
	}

	for _, c := range raw.WinConditions {
		s.WinConditions = append(s.WinConditions, ConditionDef{
			Type:   c.Type,
			Params: c.Params,
		})
	}
	for _, c := range raw.LoseConditions {
		s.LoseConditions = append(s.LoseConditions, ConditionDef{
			Type:   c.Type,
			Params: c.Params,
		})
	}

	for _, e := range raw.Events {
		ed := EventDef{
			ID: e.ID,
			Condition: ConditionDef{
				Type:   e.Condition.Type,
				Params: e.Condition.Params,
			},
			OneShot: e.OneShot,
		}
		for _, cmd := range e.Commands {
			ed.Commands = append(ed.Commands, CommandDef{
				Type:   cmd.Type,
				Params: cmd.Params,
			})
		}
		s.Events = append(s.Events, ed)
	}

	return s, nil
}

// convertInitialState converts jsonInitialState to InitialState.
func convertInitialState(raw jsonInitialState) (InitialState, error) {
	is := InitialState{
		CaveWidth:      raw.CaveWidth,
		CaveHeight:     raw.CaveHeight,
		TerrainSeed:    raw.TerrainSeed,
		TerrainDensity: raw.TerrainDensity,
		StartingChi:    raw.StartingChi,
	}

	for _, r := range raw.PrebuiltRooms {
		rp := RoomPlacement{
			TypeID: r.TypeID,
			Pos:    types.Pos{X: r.Pos.X, Y: r.Pos.Y},
			Level:  r.Level,
		}
		if rp.Level == 0 {
			rp.Level = 1
		}
		is.PrebuiltRooms = append(is.PrebuiltRooms, rp)
	}

	for i, dv := range raw.DragonVeins {
		elem, err := elementFromString(dv.Element)
		if err != nil {
			return InitialState{}, fmt.Errorf("dragon_veins[%d]: %w", i, err)
		}
		is.DragonVeins = append(is.DragonVeins, DragonVeinPlacement{
			SourcePos: types.Pos{X: dv.SourcePos.X, Y: dv.SourcePos.Y},
			Element:   elem,
			FlowRate:  dv.FlowRate,
		})
	}

	for _, b := range raw.StartingBeasts {
		is.StartingBeasts = append(is.StartingBeasts, BeastPlacement{
			SpeciesID: b.SpeciesID,
			RoomIndex: b.RoomIndex,
		})
	}

	return is, nil
}

// elementFromString converts a string element name to types.Element.
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

// ValidationContext provides registries needed for scenario validation.
type ValidationContext struct {
	RoomTypes      *world.RoomTypeRegistry
	Species        *senju.SpeciesRegistry
	InvaderClasses *invasion.InvaderClassRegistry
}

// ValidateScenario checks the structural integrity of a scenario.
// It returns all validation errors found (not just the first one).
func ValidateScenario(s *Scenario, ctx ValidationContext) []error {
	var errs []error

	// At least one win condition.
	if len(s.WinConditions) == 0 {
		errs = append(errs, fmt.Errorf("scenario must have at least one win condition"))
	}

	// At least one lose condition.
	if len(s.LoseConditions) == 0 {
		errs = append(errs, fmt.Errorf("scenario must have at least one lose condition"))
	}

	// Validate win/lose condition types.
	for i, c := range s.WinConditions {
		if _, err := NewCondition(c); err != nil {
			errs = append(errs, fmt.Errorf("win_conditions[%d]: %w", i, err))
		}
	}
	for i, c := range s.LoseConditions {
		if _, err := NewCondition(c); err != nil {
			errs = append(errs, fmt.Errorf("lose_conditions[%d]: %w", i, err))
		}
	}

	w := s.InitialState.CaveWidth
	h := s.InitialState.CaveHeight

	// Room placements within cave bounds.
	for i, r := range s.InitialState.PrebuiltRooms {
		if r.Pos.X < 0 || r.Pos.Y < 0 || r.Pos.X >= w || r.Pos.Y >= h {
			errs = append(errs, fmt.Errorf("prebuilt_rooms[%d]: position (%d,%d) out of cave bounds (%d x %d)", i, r.Pos.X, r.Pos.Y, w, h))
		}
	}

	// WaveSchedule TriggerTick within MaxTicks.
	if s.Constraints.MaxTicks > 0 {
		for i, ws := range s.WaveSchedule {
			if ws.TriggerTick > s.Constraints.MaxTicks {
				errs = append(errs, fmt.Errorf("wave_schedule[%d]: trigger_tick %d exceeds max_ticks %d", i, ws.TriggerTick, s.Constraints.MaxTicks))
			}
		}
	}

	// Validate referenced RoomTypeIDs.
	if ctx.RoomTypes != nil {
		for i, r := range s.InitialState.PrebuiltRooms {
			if _, err := ctx.RoomTypes.Get(r.TypeID); err != nil {
				errs = append(errs, fmt.Errorf("prebuilt_rooms[%d]: %w", i, err))
			}
		}
	}

	// Validate referenced SpeciesIDs.
	if ctx.Species != nil {
		for i, b := range s.InitialState.StartingBeasts {
			if _, err := ctx.Species.Get(b.SpeciesID); err != nil {
				errs = append(errs, fmt.Errorf("starting_beasts[%d]: %w", i, err))
			}
		}
	}

	// Validate referenced InvaderClassIDs.
	if ctx.InvaderClasses != nil {
		for i, ws := range s.WaveSchedule {
			for _, classID := range ws.PreferredClasses {
				if _, err := ctx.InvaderClasses.Get(classID); err != nil {
					errs = append(errs, fmt.Errorf("wave_schedule[%d]: preferred_classes: %w", i, err))
				}
			}
		}
	}

	// TerrainDensity in [0.0, 0.5].
	td := s.InitialState.TerrainDensity
	if td < 0.0 || td > 0.5 {
		errs = append(errs, fmt.Errorf("terrain_density %.2f out of range [0.0, 0.5]", td))
	}

	// PrebuiltRooms must contain a dragon hole.
	hasDragonHole := false
	for _, r := range s.InitialState.PrebuiltRooms {
		if r.TypeID == "dragon_hole" {
			hasDragonHole = true
			break
		}
	}
	if !hasDragonHole {
		errs = append(errs, fmt.Errorf("prebuilt_rooms must contain a dragon_hole"))
	}

	// Validate event definitions.
	for i, e := range s.Events {
		if _, err := NewCondition(e.Condition); err != nil {
			errs = append(errs, fmt.Errorf("events[%d] %q condition: %w", i, e.ID, err))
		}
		for j, cmd := range e.Commands {
			if _, err := NewCommand(cmd); err != nil {
				errs = append(errs, fmt.Errorf("events[%d] %q commands[%d]: %w", i, e.ID, j, err))
			}
		}
	}

	return errs
}
