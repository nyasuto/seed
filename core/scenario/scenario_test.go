package scenario

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestScenario_ZeroValue(t *testing.T) {
	var s Scenario
	if s.ID != "" {
		t.Errorf("zero value ID = %q, want empty", s.ID)
	}
	if s.Name != "" {
		t.Errorf("zero value Name = %q, want empty", s.Name)
	}
	if s.WinConditions != nil {
		t.Errorf("zero value WinConditions = %v, want nil", s.WinConditions)
	}
	if s.LoseConditions != nil {
		t.Errorf("zero value LoseConditions = %v, want nil", s.LoseConditions)
	}
	if s.WaveSchedule != nil {
		t.Errorf("zero value WaveSchedule = %v, want nil", s.WaveSchedule)
	}
	if s.Events != nil {
		t.Errorf("zero value Events = %v, want nil", s.Events)
	}
}

func TestScenario_FieldAssignment(t *testing.T) {
	s := Scenario{
		ID:          "scenario_01",
		Name:        "Tutorial",
		Description: "A tutorial scenario",
		Difficulty:  "easy",
		InitialState: InitialState{
			CaveWidth:      32,
			CaveHeight:     24,
			TerrainSeed:    42,
			TerrainDensity: 0.3,
			PrebuiltRooms: []RoomPlacement{
				{TypeID: "dragon_hole", Pos: types.Pos{X: 5, Y: 5}, Level: 1},
			},
			DragonVeins: []DragonVeinPlacement{
				{SourcePos: types.Pos{X: 10, Y: 10}, Element: types.Wood, FlowRate: 1.5},
			},
			StartingChi: 100.0,
			StartingBeasts: []BeastPlacement{
				{SpeciesID: "fire_rat", RoomIndex: 0},
				{SpeciesID: "water_snake", RoomIndex: -1},
			},
		},
		WinConditions:  []ConditionDef{{}},
		LoseConditions: []ConditionDef{{}, {}},
		WaveSchedule:   []WaveScheduleEntry{{}},
		Events:         []EventDef{{}},
		Constraints: GameConstraints{
			MaxRooms:           10,
			MaxBeasts:          5,
			MaxTicks:           3000,
			ForbiddenRoomTypes: []string{"metal_room"},
		},
	}

	if s.ID != "scenario_01" {
		t.Errorf("ID = %q, want %q", s.ID, "scenario_01")
	}
	if s.Name != "Tutorial" {
		t.Errorf("Name = %q, want %q", s.Name, "Tutorial")
	}
	if s.Description != "A tutorial scenario" {
		t.Errorf("Description = %q, want %q", s.Description, "A tutorial scenario")
	}
	if s.Difficulty != "easy" {
		t.Errorf("Difficulty = %q, want %q", s.Difficulty, "easy")
	}
	if len(s.WinConditions) != 1 {
		t.Errorf("len(WinConditions) = %d, want 1", len(s.WinConditions))
	}
	if len(s.LoseConditions) != 2 {
		t.Errorf("len(LoseConditions) = %d, want 2", len(s.LoseConditions))
	}
	if len(s.WaveSchedule) != 1 {
		t.Errorf("len(WaveSchedule) = %d, want 1", len(s.WaveSchedule))
	}
	if len(s.Events) != 1 {
		t.Errorf("len(Events) = %d, want 1", len(s.Events))
	}
}

func TestInitialState_Fields(t *testing.T) {
	is := InitialState{
		CaveWidth:      64,
		CaveHeight:     48,
		TerrainSeed:    12345,
		TerrainDensity: 0.5,
		StartingChi:    200.0,
	}

	if is.CaveWidth != 64 {
		t.Errorf("CaveWidth = %d, want 64", is.CaveWidth)
	}
	if is.CaveHeight != 48 {
		t.Errorf("CaveHeight = %d, want 48", is.CaveHeight)
	}
	if is.TerrainSeed != 12345 {
		t.Errorf("TerrainSeed = %d, want 12345", is.TerrainSeed)
	}
	if is.TerrainDensity != 0.5 {
		t.Errorf("TerrainDensity = %f, want 0.5", is.TerrainDensity)
	}
	if is.StartingChi != 200.0 {
		t.Errorf("StartingChi = %f, want 200.0", is.StartingChi)
	}
	if is.PrebuiltRooms != nil {
		t.Errorf("PrebuiltRooms = %v, want nil", is.PrebuiltRooms)
	}
}

func TestRoomPlacement_Fields(t *testing.T) {
	rp := RoomPlacement{
		TypeID: "wood_room",
		Pos:    types.Pos{X: 3, Y: 7},
		Level:  2,
	}

	if rp.TypeID != "wood_room" {
		t.Errorf("TypeID = %q, want %q", rp.TypeID, "wood_room")
	}
	if rp.Pos.X != 3 || rp.Pos.Y != 7 {
		t.Errorf("Pos = %v, want {3, 7}", rp.Pos)
	}
	if rp.Level != 2 {
		t.Errorf("Level = %d, want 2", rp.Level)
	}
}

func TestDragonVeinPlacement_Fields(t *testing.T) {
	dvp := DragonVeinPlacement{
		SourcePos: types.Pos{X: 15, Y: 20},
		Element:   types.Fire,
		FlowRate:  2.5,
	}

	if dvp.SourcePos.X != 15 || dvp.SourcePos.Y != 20 {
		t.Errorf("SourcePos = %v, want {15, 20}", dvp.SourcePos)
	}
	if dvp.Element != types.Fire {
		t.Errorf("Element = %v, want Fire", dvp.Element)
	}
	if dvp.FlowRate != 2.5 {
		t.Errorf("FlowRate = %f, want 2.5", dvp.FlowRate)
	}
}

func TestBeastPlacement_Fields(t *testing.T) {
	tests := []struct {
		name      string
		bp        BeastPlacement
		wantID    string
		wantIndex int
	}{
		{
			name:      "AssignedToRoom",
			bp:        BeastPlacement{SpeciesID: "fire_rat", RoomIndex: 0},
			wantID:    "fire_rat",
			wantIndex: 0,
		},
		{
			name:      "Unassigned",
			bp:        BeastPlacement{SpeciesID: "water_snake", RoomIndex: -1},
			wantID:    "water_snake",
			wantIndex: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bp.SpeciesID != tt.wantID {
				t.Errorf("SpeciesID = %q, want %q", tt.bp.SpeciesID, tt.wantID)
			}
			if tt.bp.RoomIndex != tt.wantIndex {
				t.Errorf("RoomIndex = %d, want %d", tt.bp.RoomIndex, tt.wantIndex)
			}
		})
	}
}

func TestGameConstraints_ZeroMeansNoLimit(t *testing.T) {
	var gc GameConstraints
	if gc.MaxRooms != 0 {
		t.Errorf("MaxRooms = %d, want 0 (no limit)", gc.MaxRooms)
	}
	if gc.MaxBeasts != 0 {
		t.Errorf("MaxBeasts = %d, want 0 (no limit)", gc.MaxBeasts)
	}
	if gc.MaxTicks != 0 {
		t.Errorf("MaxTicks = %d, want 0 (no limit)", gc.MaxTicks)
	}
	if gc.ForbiddenRoomTypes != nil {
		t.Errorf("ForbiddenRoomTypes = %v, want nil", gc.ForbiddenRoomTypes)
	}
}

func TestGameConstraints_Fields(t *testing.T) {
	gc := GameConstraints{
		MaxRooms:           15,
		MaxBeasts:          8,
		MaxTicks:           5000,
		ForbiddenRoomTypes: []string{"fire_room", "metal_room"},
	}

	if gc.MaxRooms != 15 {
		t.Errorf("MaxRooms = %d, want 15", gc.MaxRooms)
	}
	if gc.MaxBeasts != 8 {
		t.Errorf("MaxBeasts = %d, want 8", gc.MaxBeasts)
	}
	if gc.MaxTicks != 5000 {
		t.Errorf("MaxTicks = %d, want 5000", gc.MaxTicks)
	}
	if len(gc.ForbiddenRoomTypes) != 2 {
		t.Fatalf("len(ForbiddenRoomTypes) = %d, want 2", len(gc.ForbiddenRoomTypes))
	}
	if gc.ForbiddenRoomTypes[0] != "fire_room" {
		t.Errorf("ForbiddenRoomTypes[0] = %q, want %q", gc.ForbiddenRoomTypes[0], "fire_room")
	}
}
