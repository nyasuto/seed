package fengshui

import (
	"math"
	"os"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestEvaluateRoom_SingleRoom(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	flowParams := DefaultFlowParams()
	scoreParams := DefaultScoreParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, flowParams)

	// Fill room 1 to 50% capacity.
	engine.RoomChi[1].Current = 50.0

	ev := NewEvaluator(cave, reg, scoreParams)
	score := ev.EvaluateRoom(1, engine)

	// ChiScore = 0.5 * 100 = 50
	expectedChi := 0.5 * scoreParams.ChiRatioWeight
	if math.Abs(score.ChiScore-expectedChi) > 0.001 {
		t.Errorf("ChiScore = %v, want %v", score.ChiScore, expectedChi)
	}

	// Room 1 is on the vein path, so DragonVeinScore should be the bonus.
	if score.DragonVeinScore != scoreParams.DragonVeinBonus {
		t.Errorf("DragonVeinScore = %v, want %v", score.DragonVeinScore, scoreParams.DragonVeinBonus)
	}

	// Adjacent room 2 has the same element (Wood-Wood), so SameElementBonus.
	if math.Abs(score.AdjacencyScore-scoreParams.SameElementBonus) > 0.001 {
		t.Errorf("AdjacencyScore = %v, want %v", score.AdjacencyScore, scoreParams.SameElementBonus)
	}

	// Total = ChiScore + AdjacencyScore + DragonVeinScore
	expectedTotal := score.ChiScore + score.AdjacencyScore + score.DragonVeinScore
	if math.Abs(score.Total-expectedTotal) > 0.001 {
		t.Errorf("Total = %v, want %v", score.Total, expectedTotal)
	}
}

func TestEvaluateRoom_GeneratesAdjacencyBonus(t *testing.T) {
	// Wood generates Fire.
	cave, source := buildTwoRoomCave(t, "wood_room", "fire_room")
	reg := testRegistry()
	flowParams := DefaultFlowParams()
	scoreParams := DefaultScoreParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, flowParams)
	engine.RoomChi[1].Current = 50.0

	ev := NewEvaluator(cave, reg, scoreParams)
	score := ev.EvaluateRoom(1, engine)

	// Wood generates Fire → GeneratesBonus
	if math.Abs(score.AdjacencyScore-scoreParams.GeneratesBonus) > 0.001 {
		t.Errorf("AdjacencyScore = %v, want %v (generates bonus)", score.AdjacencyScore, scoreParams.GeneratesBonus)
	}
}

func TestEvaluateRoom_OvercomesAdjacencyPenalty(t *testing.T) {
	// Wood overcomes Earth.
	cave, source := buildTwoRoomCave(t, "wood_room", "earth_room")
	reg := testRegistry()
	flowParams := DefaultFlowParams()
	scoreParams := DefaultScoreParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, flowParams)
	engine.RoomChi[1].Current = 50.0

	ev := NewEvaluator(cave, reg, scoreParams)
	score := ev.EvaluateRoom(1, engine)

	// Wood overcomes Earth → OvercomesPenalty (-15)
	if math.Abs(score.AdjacencyScore-scoreParams.OvercomesPenalty) > 0.001 {
		t.Errorf("AdjacencyScore = %v, want %v (overcomes penalty)", score.AdjacencyScore, scoreParams.OvercomesPenalty)
	}
}

func TestEvaluateRoom_DragonVeinBonus(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	flowParams := DefaultFlowParams()
	scoreParams := DefaultScoreParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, flowParams)

	ev := NewEvaluator(cave, reg, scoreParams)

	// Room on the vein path should get the bonus.
	score := ev.EvaluateRoom(1, engine)
	if score.DragonVeinScore != scoreParams.DragonVeinBonus {
		t.Errorf("room on vein: DragonVeinScore = %v, want %v", score.DragonVeinScore, scoreParams.DragonVeinBonus)
	}
}

func TestEvaluateRoom_NoDragonVeinBonus(t *testing.T) {
	cave, _ := buildTwoRoomCave(t, "wood_room", "wood_room")
	reg := testRegistry()
	flowParams := DefaultFlowParams()
	scoreParams := DefaultScoreParams()

	// No veins at all.
	engine := NewChiFlowEngine(cave, []*DragonVein{}, reg, flowParams)

	ev := NewEvaluator(cave, reg, scoreParams)
	score := ev.EvaluateRoom(1, engine)

	if score.DragonVeinScore != 0 {
		t.Errorf("room not on vein: DragonVeinScore = %v, want 0", score.DragonVeinScore)
	}
}

func TestEvaluateAll(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "fire_room")
	reg := testRegistry()
	flowParams := DefaultFlowParams()
	scoreParams := DefaultScoreParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, flowParams)
	engine.RoomChi[1].Current = 50.0
	engine.RoomChi[2].Current = 30.0

	ev := NewEvaluator(cave, reg, scoreParams)
	scores := ev.EvaluateAll(engine)

	if len(scores) != 2 {
		t.Fatalf("EvaluateAll returned %d scores, want 2", len(scores))
	}

	// Verify each score has a valid RoomID.
	roomIDs := make(map[int]bool)
	for _, s := range scores {
		roomIDs[s.RoomID] = true
		if s.Total == 0 && s.ChiScore == 0 && s.AdjacencyScore == 0 && s.DragonVeinScore == 0 {
			t.Errorf("room %d: all scores are zero, expected some values", s.RoomID)
		}
	}
	if !roomIDs[1] || !roomIDs[2] {
		t.Errorf("expected room IDs 1 and 2, got %v", roomIDs)
	}
}

func TestCaveTotal(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "fire_room")
	reg := testRegistry()
	flowParams := DefaultFlowParams()
	scoreParams := DefaultScoreParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, flowParams)
	engine.RoomChi[1].Current = 50.0
	engine.RoomChi[2].Current = 30.0

	ev := NewEvaluator(cave, reg, scoreParams)

	total := ev.CaveTotal(engine)
	scores := ev.EvaluateAll(engine)
	expectedTotal := 0.0
	for _, s := range scores {
		expectedTotal += s.Total
	}

	if math.Abs(total-expectedTotal) > 0.001 {
		t.Errorf("CaveTotal = %v, want %v", total, expectedTotal)
	}
}

func TestEvaluateRoom_CustomParams(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "fire_room")
	reg := testRegistry()
	flowParams := DefaultFlowParams()

	// Use custom score params with different values.
	customParams := &ScoreParams{
		GeneratesBonus:   40.0,
		OvercomesPenalty: -30.0,
		SameElementBonus: 10.0,
		DragonVeinBonus:  60.0,
		ChiRatioWeight:   200.0,
	}

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, flowParams)
	engine.RoomChi[1].Current = 50.0

	ev := NewEvaluator(cave, reg, customParams)
	score := ev.EvaluateRoom(1, engine)

	// ChiScore = 0.5 * 200 = 100
	expectedChi := 0.5 * customParams.ChiRatioWeight
	if math.Abs(score.ChiScore-expectedChi) > 0.001 {
		t.Errorf("ChiScore = %v, want %v", score.ChiScore, expectedChi)
	}

	// Wood generates Fire → custom GeneratesBonus = 40
	if math.Abs(score.AdjacencyScore-customParams.GeneratesBonus) > 0.001 {
		t.Errorf("AdjacencyScore = %v, want %v", score.AdjacencyScore, customParams.GeneratesBonus)
	}

	// DragonVein bonus = 60
	if score.DragonVeinScore != customParams.DragonVeinBonus {
		t.Errorf("DragonVeinScore = %v, want %v", score.DragonVeinScore, customParams.DragonVeinBonus)
	}
}

func TestDefaultScoreParams(t *testing.T) {
	p := DefaultScoreParams()
	if p.GeneratesBonus != 20.0 {
		t.Errorf("GeneratesBonus = %v, want 20.0", p.GeneratesBonus)
	}
	if p.OvercomesPenalty != -15.0 {
		t.Errorf("OvercomesPenalty = %v, want -15.0", p.OvercomesPenalty)
	}
	if p.SameElementBonus != 5.0 {
		t.Errorf("SameElementBonus = %v, want 5.0", p.SameElementBonus)
	}
	if p.DragonVeinBonus != 30.0 {
		t.Errorf("DragonVeinBonus = %v, want 30.0", p.DragonVeinBonus)
	}
	if p.ChiRatioWeight != 100.0 {
		t.Errorf("ChiRatioWeight = %v, want 100.0", p.ChiRatioWeight)
	}
}

func TestLoadScoreParams(t *testing.T) {
	// Write a temporary JSON file.
	dir := t.TempDir()
	path := dir + "/params.json"
	data := []byte(`{
		"generates_bonus": 25.0,
		"overcomes_penalty": -10.0,
		"same_element_bonus": 8.0,
		"dragon_vein_bonus": 35.0,
		"chi_ratio_weight": 150.0
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	p, err := LoadScoreParams(path)
	if err != nil {
		t.Fatalf("LoadScoreParams: %v", err)
	}
	if p.GeneratesBonus != 25.0 {
		t.Errorf("GeneratesBonus = %v, want 25.0", p.GeneratesBonus)
	}
	if p.OvercomesPenalty != -10.0 {
		t.Errorf("OvercomesPenalty = %v, want -10.0", p.OvercomesPenalty)
	}
	if p.ChiRatioWeight != 150.0 {
		t.Errorf("ChiRatioWeight = %v, want 150.0", p.ChiRatioWeight)
	}
}

func TestLoadScoreParams_FileNotFound(t *testing.T) {
	_, err := LoadScoreParams("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
