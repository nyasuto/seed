package scenario

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestLoadScenario_Basic(t *testing.T) {
	data := []byte(`{
		"id": "test_scenario",
		"name": "Test Scenario",
		"description": "A test scenario",
		"difficulty": "easy",
		"initial_state": {
			"cave_width": 16,
			"cave_height": 16,
			"terrain_seed": 42,
			"terrain_density": 0.1,
			"prebuilt_rooms": [
				{"type_id": "dragon_hole", "pos": {"x": 8, "y": 8}, "level": 1}
			],
			"dragon_veins": [
				{"source_pos": {"x": 8, "y": 8}, "element": "Earth", "flow_rate": 1.0}
			],
			"starting_chi": 100.0,
			"starting_beasts": [
				{"species_id": "kirin", "room_index": 0}
			]
		},
		"win_conditions": [
			{"type": "survive_until", "params": {"ticks": 500}}
		],
		"lose_conditions": [
			{"type": "core_destroyed"}
		],
		"wave_schedule": [
			{
				"trigger_tick": 100,
				"difficulty": 1.0,
				"min_invaders": 1,
				"max_invaders": 3
			}
		],
		"events": [
			{
				"id": "bonus",
				"condition": {"type": "survive_until", "params": {"ticks": 200}},
				"commands": [{"type": "modify_chi", "params": {"amount": 50}}],
				"one_shot": true
			}
		],
		"constraints": {
			"max_rooms": 5,
			"max_beasts": 3,
			"max_ticks": 500,
			"forbidden_room_types": ["fire_room"]
		}
	}`)

	s, err := LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario failed: %v", err)
	}

	if s.ID != "test_scenario" {
		t.Errorf("ID = %q, want %q", s.ID, "test_scenario")
	}
	if s.Name != "Test Scenario" {
		t.Errorf("Name = %q, want %q", s.Name, "Test Scenario")
	}
	if s.Difficulty != "easy" {
		t.Errorf("Difficulty = %q, want %q", s.Difficulty, "easy")
	}

	// InitialState
	is := s.InitialState
	if is.CaveWidth != 16 || is.CaveHeight != 16 {
		t.Errorf("cave size = %dx%d, want 16x16", is.CaveWidth, is.CaveHeight)
	}
	if is.TerrainSeed != 42 {
		t.Errorf("TerrainSeed = %d, want 42", is.TerrainSeed)
	}
	if is.TerrainDensity != 0.1 {
		t.Errorf("TerrainDensity = %f, want 0.1", is.TerrainDensity)
	}
	if is.StartingChi != 100.0 {
		t.Errorf("StartingChi = %f, want 100.0", is.StartingChi)
	}

	// PrebuiltRooms
	if len(is.PrebuiltRooms) != 1 {
		t.Fatalf("PrebuiltRooms len = %d, want 1", len(is.PrebuiltRooms))
	}
	room := is.PrebuiltRooms[0]
	if room.TypeID != "dragon_hole" {
		t.Errorf("room TypeID = %q, want %q", room.TypeID, "dragon_hole")
	}
	if room.Pos != (types.Pos{X: 8, Y: 8}) {
		t.Errorf("room Pos = %v, want {8,8}", room.Pos)
	}
	if room.Level != 1 {
		t.Errorf("room Level = %d, want 1", room.Level)
	}

	// DragonVeins
	if len(is.DragonVeins) != 1 {
		t.Fatalf("DragonVeins len = %d, want 1", len(is.DragonVeins))
	}
	dv := is.DragonVeins[0]
	if dv.Element != types.Earth {
		t.Errorf("DragonVein Element = %v, want Earth", dv.Element)
	}
	if dv.FlowRate != 1.0 {
		t.Errorf("DragonVein FlowRate = %f, want 1.0", dv.FlowRate)
	}

	// StartingBeasts
	if len(is.StartingBeasts) != 1 {
		t.Fatalf("StartingBeasts len = %d, want 1", len(is.StartingBeasts))
	}
	if is.StartingBeasts[0].SpeciesID != "kirin" {
		t.Errorf("beast SpeciesID = %q, want %q", is.StartingBeasts[0].SpeciesID, "kirin")
	}

	// Conditions
	if len(s.WinConditions) != 1 {
		t.Fatalf("WinConditions len = %d, want 1", len(s.WinConditions))
	}
	if s.WinConditions[0].Type != "survive_until" {
		t.Errorf("win condition type = %q, want %q", s.WinConditions[0].Type, "survive_until")
	}
	if len(s.LoseConditions) != 1 {
		t.Fatalf("LoseConditions len = %d, want 1", len(s.LoseConditions))
	}
	if s.LoseConditions[0].Type != "core_destroyed" {
		t.Errorf("lose condition type = %q, want %q", s.LoseConditions[0].Type, "core_destroyed")
	}

	// WaveSchedule
	if len(s.WaveSchedule) != 1 {
		t.Fatalf("WaveSchedule len = %d, want 1", len(s.WaveSchedule))
	}
	if s.WaveSchedule[0].TriggerTick != 100 {
		t.Errorf("wave TriggerTick = %d, want 100", s.WaveSchedule[0].TriggerTick)
	}

	// Events
	if len(s.Events) != 1 {
		t.Fatalf("Events len = %d, want 1", len(s.Events))
	}
	if s.Events[0].ID != "bonus" {
		t.Errorf("event ID = %q, want %q", s.Events[0].ID, "bonus")
	}
	if !s.Events[0].OneShot {
		t.Error("event OneShot = false, want true")
	}
	if len(s.Events[0].Commands) != 1 {
		t.Fatalf("event commands len = %d, want 1", len(s.Events[0].Commands))
	}
	if s.Events[0].Commands[0].Type != "modify_chi" {
		t.Errorf("command type = %q, want %q", s.Events[0].Commands[0].Type, "modify_chi")
	}

	// Constraints
	if s.Constraints.MaxRooms != 5 {
		t.Errorf("MaxRooms = %d, want 5", s.Constraints.MaxRooms)
	}
	if s.Constraints.MaxTicks != 500 {
		t.Errorf("MaxTicks = %d, want 500", s.Constraints.MaxTicks)
	}
	if len(s.Constraints.ForbiddenRoomTypes) != 1 || s.Constraints.ForbiddenRoomTypes[0] != "fire_room" {
		t.Errorf("ForbiddenRoomTypes = %v, want [fire_room]", s.Constraints.ForbiddenRoomTypes)
	}
}

func TestLoadScenario_DefaultLevel(t *testing.T) {
	data := []byte(`{
		"id": "test",
		"name": "Test",
		"initial_state": {
			"cave_width": 8,
			"cave_height": 8,
			"prebuilt_rooms": [
				{"type_id": "dragon_hole", "pos": {"x": 4, "y": 4}}
			]
		},
		"win_conditions": [{"type": "survive_until", "params": {"ticks": 100}}],
		"lose_conditions": [{"type": "core_destroyed"}]
	}`)

	s, err := LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario failed: %v", err)
	}

	if s.InitialState.PrebuiltRooms[0].Level != 1 {
		t.Errorf("default Level = %d, want 1", s.InitialState.PrebuiltRooms[0].Level)
	}
}

func TestLoadScenario_InvalidJSON(t *testing.T) {
	_, err := LoadScenario([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadScenario_InvalidElement(t *testing.T) {
	data := []byte(`{
		"id": "test",
		"initial_state": {
			"cave_width": 8,
			"cave_height": 8,
			"dragon_veins": [
				{"source_pos": {"x": 4, "y": 4}, "element": "Invalid", "flow_rate": 1.0}
			]
		}
	}`)

	_, err := LoadScenario(data)
	if err == nil {
		t.Fatal("expected error for invalid element")
	}
}

func TestLoadScenario_AllElements(t *testing.T) {
	elements := []struct {
		name string
		want types.Element
	}{
		{"Wood", types.Wood},
		{"Fire", types.Fire},
		{"Earth", types.Earth},
		{"Metal", types.Metal},
		{"Water", types.Water},
	}

	for _, tc := range elements {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte(`{
				"id": "test",
				"initial_state": {
					"cave_width": 8,
					"cave_height": 8,
					"dragon_veins": [
						{"source_pos": {"x": 4, "y": 4}, "element": "` + tc.name + `", "flow_rate": 1.0}
					]
				}
			}`)

			s, err := LoadScenario(data)
			if err != nil {
				t.Fatalf("LoadScenario failed: %v", err)
			}
			if s.InitialState.DragonVeins[0].Element != tc.want {
				t.Errorf("Element = %v, want %v", s.InitialState.DragonVeins[0].Element, tc.want)
			}
		})
	}
}

func TestLoadScenario_EmptyArrays(t *testing.T) {
	data := []byte(`{
		"id": "minimal",
		"name": "Minimal",
		"initial_state": {
			"cave_width": 8,
			"cave_height": 8
		}
	}`)

	s, err := LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario failed: %v", err)
	}

	if s.WinConditions != nil {
		t.Errorf("WinConditions = %v, want nil", s.WinConditions)
	}
	if s.WaveSchedule != nil {
		t.Errorf("WaveSchedule = %v, want nil", s.WaveSchedule)
	}
	if s.Events != nil {
		t.Errorf("Events = %v, want nil", s.Events)
	}
}
