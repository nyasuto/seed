package scenario

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/types"
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

// --- Validation tests ---

// validScenario returns a minimal valid scenario for use in validation tests.
func validScenario() *Scenario {
	return &Scenario{
		ID:   "valid",
		Name: "Valid",
		InitialState: InitialState{
			CaveWidth:      16,
			CaveHeight:     16,
			TerrainDensity: 0.1,
			PrebuiltRooms: []RoomPlacement{
				{TypeID: "dragon_hole", Pos: types.Pos{X: 8, Y: 8}, Level: 1},
			},
		},
		WinConditions:  []ConditionDef{{Type: "survive_until", Params: json.RawMessage(`{"ticks": 300}`)}},
		LoseConditions: []ConditionDef{{Type: "core_destroyed"}},
		Constraints:    GameConstraints{MaxTicks: 500},
	}
}

func TestValidateScenario_Valid(t *testing.T) {
	s := validScenario()
	errs := ValidateScenario(s, ValidationContext{})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateScenario_NoWinConditions(t *testing.T) {
	s := validScenario()
	s.WinConditions = nil
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "at least one win condition") {
		t.Errorf("expected win condition error, got %v", errs)
	}
}

func TestValidateScenario_NoLoseConditions(t *testing.T) {
	s := validScenario()
	s.LoseConditions = nil
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "at least one lose condition") {
		t.Errorf("expected lose condition error, got %v", errs)
	}
}

func TestValidateScenario_InvalidWinConditionType(t *testing.T) {
	s := validScenario()
	s.WinConditions = []ConditionDef{{Type: "nonexistent"}}
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "win_conditions[0]") {
		t.Errorf("expected win condition type error, got %v", errs)
	}
}

func TestValidateScenario_InvalidLoseConditionType(t *testing.T) {
	s := validScenario()
	s.LoseConditions = []ConditionDef{{Type: "nonexistent"}}
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "lose_conditions[0]") {
		t.Errorf("expected lose condition type error, got %v", errs)
	}
}

func TestValidateScenario_RoomOutOfBounds(t *testing.T) {
	s := validScenario()
	s.InitialState.PrebuiltRooms = append(s.InitialState.PrebuiltRooms,
		RoomPlacement{TypeID: "wood_room", Pos: types.Pos{X: 100, Y: 100}, Level: 1},
	)
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "out of cave bounds") {
		t.Errorf("expected out of bounds error, got %v", errs)
	}
}

func TestValidateScenario_RoomNegativePos(t *testing.T) {
	s := validScenario()
	s.InitialState.PrebuiltRooms = append(s.InitialState.PrebuiltRooms,
		RoomPlacement{TypeID: "wood_room", Pos: types.Pos{X: -1, Y: 5}, Level: 1},
	)
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "out of cave bounds") {
		t.Errorf("expected out of bounds error, got %v", errs)
	}
}

func TestValidateScenario_WaveTriggerExceedsMaxTicks(t *testing.T) {
	s := validScenario()
	s.Constraints.MaxTicks = 200
	s.WaveSchedule = []WaveScheduleEntry{
		{TriggerTick: 300, Difficulty: 1.0, MinInvaders: 1, MaxInvaders: 2},
	}
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "exceeds max_ticks") {
		t.Errorf("expected trigger_tick exceeds max_ticks error, got %v", errs)
	}
}

func TestValidateScenario_WaveTriggerWithZeroMaxTicks(t *testing.T) {
	s := validScenario()
	s.Constraints.MaxTicks = 0
	s.WaveSchedule = []WaveScheduleEntry{
		{TriggerTick: 9999, Difficulty: 1.0, MinInvaders: 1, MaxInvaders: 2},
	}
	errs := ValidateScenario(s, ValidationContext{})
	// MaxTicks == 0 means no limit, so no error expected.
	if containsError(errs, "exceeds max_ticks") {
		t.Errorf("should not error when MaxTicks is 0, got %v", errs)
	}
}

func TestValidateScenario_TerrainDensityOutOfRange(t *testing.T) {
	tests := []struct {
		name    string
		density float64
	}{
		{"negative", -0.1},
		{"too_high", 0.6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := validScenario()
			s.InitialState.TerrainDensity = tc.density
			errs := ValidateScenario(s, ValidationContext{})
			if !containsError(errs, "terrain_density") {
				t.Errorf("expected terrain_density error for %f, got %v", tc.density, errs)
			}
		})
	}
}

func TestValidateScenario_NoDragonHole(t *testing.T) {
	s := validScenario()
	s.InitialState.PrebuiltRooms = []RoomPlacement{
		{TypeID: "wood_room", Pos: types.Pos{X: 5, Y: 5}, Level: 1},
	}
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "dragon_hole") {
		t.Errorf("expected dragon_hole error, got %v", errs)
	}
}

func TestValidateScenario_InvalidEventCondition(t *testing.T) {
	s := validScenario()
	s.Events = []EventDef{
		{
			ID:        "bad_event",
			Condition: ConditionDef{Type: "nonexistent"},
			Commands:  []CommandDef{{Type: "message", Params: json.RawMessage(`{"text": "hi"}`)}},
		},
	}
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "events[0]") {
		t.Errorf("expected event condition error, got %v", errs)
	}
}

func TestValidateScenario_InvalidEventCommand(t *testing.T) {
	s := validScenario()
	s.Events = []EventDef{
		{
			ID:        "bad_cmd_event",
			Condition: ConditionDef{Type: "core_destroyed"},
			Commands:  []CommandDef{{Type: "nonexistent_cmd"}},
		},
	}
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "commands[0]") {
		t.Errorf("expected event command error, got %v", errs)
	}
}

func TestValidateScenario_MultipleErrors(t *testing.T) {
	s := &Scenario{
		ID: "broken",
		InitialState: InitialState{
			CaveWidth:      8,
			CaveHeight:     8,
			TerrainDensity: 0.8, // out of range
			PrebuiltRooms: []RoomPlacement{
				{TypeID: "wood_room", Pos: types.Pos{X: 100, Y: 100}, Level: 1}, // out of bounds, no dragon_hole
			},
		},
		// no win or lose conditions
	}
	errs := ValidateScenario(s, ValidationContext{})
	if len(errs) < 4 {
		t.Errorf("expected at least 4 errors (no win, no lose, out of bounds, terrain_density, no dragon_hole), got %d: %v", len(errs), errs)
	}
}

// --- Testdata JSON loading tests ---

func TestLoadScenario_TutorialJSON(t *testing.T) {
	data := readTestdata(t, "tutorial.json")
	s, err := LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario tutorial.json: %v", err)
	}

	if s.ID != "tutorial" {
		t.Errorf("ID = %q, want %q", s.ID, "tutorial")
	}
	if s.Difficulty != "easy" {
		t.Errorf("Difficulty = %q, want %q", s.Difficulty, "easy")
	}
	if s.InitialState.CaveWidth != 16 || s.InitialState.CaveHeight != 16 {
		t.Errorf("cave size = %dx%d, want 16x16", s.InitialState.CaveWidth, s.InitialState.CaveHeight)
	}
	if len(s.WinConditions) != 1 || s.WinConditions[0].Type != "survive_until" {
		t.Errorf("unexpected win conditions: %v", s.WinConditions)
	}
	if len(s.LoseConditions) != 1 || s.LoseConditions[0].Type != "core_destroyed" {
		t.Errorf("unexpected lose conditions: %v", s.LoseConditions)
	}
	if len(s.WaveSchedule) != 1 {
		t.Errorf("WaveSchedule len = %d, want 1", len(s.WaveSchedule))
	}
	if len(s.InitialState.StartingBeasts) != 1 {
		t.Errorf("StartingBeasts len = %d, want 1", len(s.InitialState.StartingBeasts))
	}

	// Should pass validation (no registry context).
	errs := ValidateScenario(s, ValidationContext{})
	if len(errs) != 0 {
		t.Errorf("validation errors: %v", errs)
	}
}

func TestLoadScenario_StandardJSON(t *testing.T) {
	data := readTestdata(t, "standard.json")
	s, err := LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario standard.json: %v", err)
	}

	if s.ID != "standard" {
		t.Errorf("ID = %q, want %q", s.ID, "standard")
	}
	if s.Difficulty != "normal" {
		t.Errorf("Difficulty = %q, want %q", s.Difficulty, "normal")
	}
	if s.InitialState.CaveWidth != 32 || s.InitialState.CaveHeight != 32 {
		t.Errorf("cave size = %dx%d, want 32x32", s.InitialState.CaveWidth, s.InitialState.CaveHeight)
	}
	if len(s.WinConditions) != 2 {
		t.Errorf("WinConditions len = %d, want 2", len(s.WinConditions))
	}
	if len(s.LoseConditions) != 2 {
		t.Errorf("LoseConditions len = %d, want 2", len(s.LoseConditions))
	}
	if len(s.WaveSchedule) != 5 {
		t.Errorf("WaveSchedule len = %d, want 5", len(s.WaveSchedule))
	}
	if len(s.Events) != 7 {
		t.Errorf("Events len = %d, want 7", len(s.Events))
	}
	if len(s.InitialState.DragonVeins) != 2 {
		t.Errorf("DragonVeins len = %d, want 2", len(s.InitialState.DragonVeins))
	}

	errs := ValidateScenario(s, ValidationContext{})
	if len(errs) != 0 {
		t.Errorf("validation errors: %v", errs)
	}
}

func TestLoadScenario_AntipatternRichJSON(t *testing.T) {
	data := readTestdata(t, "antipattern_rich.json")
	s, err := LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario antipattern_rich.json: %v", err)
	}

	if s.ID != "antipattern_rich" {
		t.Errorf("ID = %q, want %q", s.ID, "antipattern_rich")
	}
	if s.InitialState.StartingChi != 9999.0 {
		t.Errorf("StartingChi = %f, want 9999.0", s.InitialState.StartingChi)
	}
	if len(s.InitialState.DragonVeins) != 3 {
		t.Errorf("DragonVeins len = %d, want 3", len(s.InitialState.DragonVeins))
	}
	if len(s.WaveSchedule) != 2 {
		t.Errorf("WaveSchedule len = %d, want 2", len(s.WaveSchedule))
	}

	errs := ValidateScenario(s, ValidationContext{})
	if len(errs) != 0 {
		t.Errorf("validation errors: %v", errs)
	}
}

func TestLoadScenario_AntipatternImpossibleJSON(t *testing.T) {
	data := readTestdata(t, "antipattern_impossible.json")
	s, err := LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario antipattern_impossible.json: %v", err)
	}

	if s.ID != "antipattern_impossible" {
		t.Errorf("ID = %q, want %q", s.ID, "antipattern_impossible")
	}
	if s.InitialState.TerrainDensity != 0.5 {
		t.Errorf("TerrainDensity = %f, want 0.5", s.InitialState.TerrainDensity)
	}
	if s.InitialState.StartingChi != 10.0 {
		t.Errorf("StartingChi = %f, want 10.0", s.InitialState.StartingChi)
	}
	if len(s.WaveSchedule) != 3 {
		t.Errorf("WaveSchedule len = %d, want 3", len(s.WaveSchedule))
	}
	if s.WaveSchedule[0].TriggerTick != 10 {
		t.Errorf("first wave TriggerTick = %d, want 10", s.WaveSchedule[0].TriggerTick)
	}

	// TerrainDensity 0.5 is at the boundary, should still pass.
	errs := ValidateScenario(s, ValidationContext{})
	if len(errs) != 0 {
		t.Errorf("validation errors: %v", errs)
	}
}

// --- Rejection tests for invalid scenarios ---

func TestLoadScenario_RejectMissingConditionParams(t *testing.T) {
	// survive_until without ticks param should fail at validation.
	s := validScenario()
	s.WinConditions = []ConditionDef{{Type: "survive_until"}} // missing "ticks"
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "win_conditions[0]") {
		t.Errorf("expected validation error for missing params, got %v", errs)
	}
}

func TestLoadScenario_RejectInvalidConditionParams(t *testing.T) {
	// fengshui_score without threshold param.
	s := validScenario()
	s.WinConditions = []ConditionDef{{Type: "fengshui_score"}} // missing "threshold"
	errs := ValidateScenario(s, ValidationContext{})
	if !containsError(errs, "win_conditions[0]") {
		t.Errorf("expected validation error for missing threshold, got %v", errs)
	}
}

// --- Helpers ---

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	return data
}

func containsError(errs []error, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e.Error(), substr) {
			return true
		}
	}
	return false
}
