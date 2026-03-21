package senju

import (
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

func TestProcessDefeat_TransitionsToStunned(t *testing.T) {
	beast := &Beast{
		ID:    1,
		Level: 5,
		HP:    -3,
		MaxHP: 100,
		State: Fighting,
	}

	dp := NewDefeatProcessor()
	result := dp.ProcessDefeat(beast, 10)

	if beast.State != Stunned {
		t.Errorf("expected beast state Stunned, got %v", beast.State)
	}
	if beast.HP != 0 {
		t.Errorf("expected beast HP 0, got %d", beast.HP)
	}
	if result.NewState != Stunned {
		t.Errorf("expected result NewState Stunned, got %v", result.NewState)
	}
	if result.BeastID != 1 {
		t.Errorf("expected BeastID 1, got %d", result.BeastID)
	}
}

func TestProcessDefeat_RevivalTick(t *testing.T) {
	beast := &Beast{
		ID:    1,
		Level: 5,
		HP:    0,
		MaxHP: 100,
		State: Fighting,
	}

	dp := NewDefeatProcessor()
	result := dp.ProcessDefeat(beast, 50)

	// Default StunnedDuration is 20, so revival at tick 70.
	if result.RevivalTick != 70 {
		t.Errorf("expected RevivalTick 70, got %d", result.RevivalTick)
	}
}

func TestProcessDefeat_RevivalHP(t *testing.T) {
	beast := &Beast{
		ID:    1,
		Level: 5,
		HP:    0,
		MaxHP: 100,
		State: Fighting,
	}

	dp := NewDefeatProcessor()
	result := dp.ProcessDefeat(beast, 10)

	// Default RevivalHPRatio is 0.3, so 100 * 0.3 = 30.
	if result.RevivalHP != 30 {
		t.Errorf("expected RevivalHP 30, got %d", result.RevivalHP)
	}
}

func TestProcessDefeat_LevelPenalty(t *testing.T) {
	beast := &Beast{
		ID:    1,
		Level: 5,
		HP:    0,
		MaxHP: 100,
		State: Fighting,
	}

	dp := NewDefeatProcessor()
	result := dp.ProcessDefeat(beast, 10)

	// Default LevelPenalty is 1.
	if result.LevelPenalty != 1 {
		t.Errorf("expected LevelPenalty 1, got %d", result.LevelPenalty)
	}
}

func TestProcessDefeat_Level1NeverGoesToZero(t *testing.T) {
	beast := &Beast{
		ID:    1,
		Level: 1,
		HP:    0,
		MaxHP: 50,
		State: Fighting,
	}

	dp := NewDefeatProcessor()
	result := dp.ProcessDefeat(beast, 10)

	// Level 1 beast cannot lose levels; penalty should be clamped to 0.
	if result.LevelPenalty != 0 {
		t.Errorf("expected LevelPenalty 0 for level 1 beast, got %d", result.LevelPenalty)
	}
}

func TestProcessDefeat_RevivalHPMinimumOne(t *testing.T) {
	beast := &Beast{
		ID:    1,
		Level: 3,
		HP:    0,
		MaxHP: 1, // 1 * 0.3 = 0.3 -> int(0.3) = 0 -> clamped to 1
		State: Fighting,
	}

	dp := NewDefeatProcessor()
	result := dp.ProcessDefeat(beast, 10)

	if result.RevivalHP != 1 {
		t.Errorf("expected RevivalHP minimum 1, got %d", result.RevivalHP)
	}
}

func TestProcessDefeat_CustomParams(t *testing.T) {
	beast := &Beast{
		ID:    1,
		Level: 3,
		HP:    -5,
		MaxHP: 200,
		State: Fighting,
	}

	params := &DefeatParams{
		StunnedDuration: 10,
		RevivalHPRatio:  0.5,
		LevelPenalty:    2,
	}
	dp := NewDefeatProcessorWithParams(params)
	result := dp.ProcessDefeat(beast, 100)

	if result.RevivalTick != 110 {
		t.Errorf("expected RevivalTick 110, got %d", result.RevivalTick)
	}
	if result.RevivalHP != 100 {
		t.Errorf("expected RevivalHP 100, got %d", result.RevivalHP)
	}
	if result.LevelPenalty != 2 {
		t.Errorf("expected LevelPenalty 2, got %d", result.LevelPenalty)
	}
}

func TestProcessDefeat_CustomParams_LevelPenaltyClamp(t *testing.T) {
	beast := &Beast{
		ID:    1,
		Level: 2,
		HP:    0,
		MaxHP: 100,
		State: Fighting,
	}

	params := &DefeatParams{
		StunnedDuration: 10,
		RevivalHPRatio:  0.3,
		LevelPenalty:    5, // wants to take 5 levels but beast is level 2
	}
	dp := NewDefeatProcessorWithParams(params)
	result := dp.ProcessDefeat(beast, 10)

	// Level 2, penalty 5 -> new level would be -3, so clamp penalty to 1 (2-1=1).
	if result.LevelPenalty != 1 {
		t.Errorf("expected LevelPenalty 1 (clamped), got %d", result.LevelPenalty)
	}
}

func TestBehaviorEngine_Tick_SkipsStunnedBeast(t *testing.T) {
	cave := &world.Cave{
		Rooms: []*world.Room{
			{ID: 1, TypeID: "D001"},
		},
	}
	ag := cave.BuildAdjacencyGraph()

	reg := world.NewRoomTypeRegistry()
	be := NewBehaviorEngine(cave, ag, reg, nil)

	stunnedBeast := &Beast{
		ID:     1,
		RoomID: 1,
		Level:  3,
		HP:     0,
		MaxHP:  100,
		State:  Stunned,
	}
	activeBeast := &Beast{
		ID:     2,
		RoomID: 1,
		Level:  3,
		HP:     50,
		MaxHP:  100,
		State:  Idle,
	}

	be.AssignBehavior(activeBeast, Guard)

	beasts := []*Beast{stunnedBeast, activeBeast}
	actions := be.Tick(beasts, nil, map[int]*fengshui.RoomChi{})

	// Only the active beast should produce an action.
	for _, a := range actions {
		if a.BeastID == stunnedBeast.ID {
			t.Errorf("stunned beast should not produce any action, got action type %v", a.Action.Type)
		}
	}

	// The active beast should have an action.
	found := false
	for _, a := range actions {
		if a.BeastID == activeBeast.ID {
			found = true
		}
	}
	if !found {
		t.Error("expected active beast to produce an action")
	}
}

func TestLoadDefeatParams(t *testing.T) {
	data := []byte(`{"stunned_duration": 15, "revival_hp_ratio": 0.4, "level_penalty": 2}`)
	params, err := LoadDefeatParams(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.StunnedDuration != 15 {
		t.Errorf("expected StunnedDuration 15, got %d", params.StunnedDuration)
	}
	if params.RevivalHPRatio != 0.4 {
		t.Errorf("expected RevivalHPRatio 0.4, got %f", params.RevivalHPRatio)
	}
	if params.LevelPenalty != 2 {
		t.Errorf("expected LevelPenalty 2, got %d", params.LevelPenalty)
	}
}

func TestLoadDefeatParams_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid}`)
	_, err := LoadDefeatParams(data)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// Verify unused import workaround is not needed.
var _ types.Tick
