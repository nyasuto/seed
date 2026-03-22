package scene

import (
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

func TestResultData_VictoryText(t *testing.T) {
	data := ResultData{
		Won:           true,
		Reason:        "all waves defeated",
		TotalTicks:    500,
		RoomCount:     8,
		DefeatedWaves: 5,
		TotalWaves:    5,
		FinalCoreHP:   80,
	}
	lines := data.ResultLines()
	if len(lines) == 0 {
		t.Fatal("expected non-empty lines")
	}
	if lines[0] != "VICTORY" {
		t.Errorf("expected first line VICTORY, got %q", lines[0])
	}
}

func TestResultData_DefeatText(t *testing.T) {
	data := ResultData{
		Won:           false,
		Reason:        "core HP reached 0",
		TotalTicks:    120,
		RoomCount:     3,
		DefeatedWaves: 1,
		TotalWaves:    5,
		FinalCoreHP:   0,
	}
	lines := data.ResultLines()
	if len(lines) == 0 {
		t.Fatal("expected non-empty lines")
	}
	if lines[0] != "DEFEAT" {
		t.Errorf("expected first line DEFEAT, got %q", lines[0])
	}
}

func TestResultData_StatisticsFormat(t *testing.T) {
	data := ResultData{
		Won:           true,
		Reason:        "survived",
		TotalTicks:    300,
		RoomCount:     5,
		DefeatedWaves: 3,
		TotalWaves:    4,
		FinalCoreHP:   50,
	}
	lines := data.ResultLines()

	expected := map[string]bool{
		"Total Ticks: 300":       false,
		"Rooms Built: 5":         false,
		"Waves Defeated: 3 / 4":  false,
		"Final Core HP: 50":      false,
	}

	for _, line := range lines {
		if _, ok := expected[line]; ok {
			expected[line] = true
		}
	}
	for text, found := range expected {
		if !found {
			t.Errorf("expected line %q not found in result lines", text)
		}
	}
}

func TestBuildResultData_Won(t *testing.T) {
	result := simulation.GameResult{
		Status:    simulation.Won,
		FinalTick: 400,
		Reason:    "all waves defeated",
	}
	snap := scenario.GameSnapshot{
		CoreHP:        75,
		RoomCount:     6,
		DefeatedWaves: 5,
		TotalWaves:    5,
	}
	data := BuildResultData(result, snap)
	if !data.Won {
		t.Error("expected Won to be true")
	}
	if data.TotalTicks != 400 {
		t.Errorf("expected TotalTicks=400, got %d", data.TotalTicks)
	}
	if data.FinalCoreHP != 75 {
		t.Errorf("expected FinalCoreHP=75, got %d", data.FinalCoreHP)
	}
}

func TestBuildResultData_Lost(t *testing.T) {
	result := simulation.GameResult{
		Status:    simulation.Lost,
		FinalTick: 150,
		Reason:    "core HP reached 0",
	}
	snap := scenario.GameSnapshot{
		CoreHP:        0,
		RoomCount:     2,
		DefeatedWaves: 1,
		TotalWaves:    5,
	}
	data := BuildResultData(result, snap)
	if data.Won {
		t.Error("expected Won to be false")
	}
	if data.Reason != "core HP reached 0" {
		t.Errorf("expected reason %q, got %q", "core HP reached 0", data.Reason)
	}
}

func TestResultScene_TitleButtonClick(t *testing.T) {
	called := false
	rs := NewResultScene(1088, 728, ResultData{Won: true}, nil, func() { called = true }, nil)

	r := rs.TitleRect()
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2

	if !rs.HandleClick(cx, cy) {
		t.Error("expected HandleClick to return true for Title button")
	}
	if !called {
		t.Error("expected onTitle callback to be called")
	}
}

func TestResultScene_RetryButtonClick(t *testing.T) {
	called := false
	rs := NewResultScene(1088, 728, ResultData{Won: false}, func() { called = true }, nil, nil)

	r := rs.RetryRect()
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2

	if !rs.HandleClick(cx, cy) {
		t.Error("expected HandleClick to return true for Retry button")
	}
	if !called {
		t.Error("expected onRetry callback to be called")
	}
}

func TestResultScene_ClickOutsideButtons(t *testing.T) {
	rs := NewResultScene(1088, 728, ResultData{}, func() { t.Error("should not fire") }, func() { t.Error("should not fire") }, nil)

	if rs.HandleClick(0, 0) {
		t.Error("expected HandleClick to return false for click outside buttons")
	}
}

func TestResultScene_TransitionToTitle(t *testing.T) {
	sm := NewSceneManager(nil)
	data := ResultData{Won: true}

	resultScene := NewResultScene(1088, 728, data, nil, func() {
		sm.Switch(&spyScene{name: "title"})
	}, nil)

	sm.Switch(resultScene)

	r := resultScene.TitleRect()
	resultScene.HandleClick((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)

	if _, ok := sm.Current().(*spyScene); !ok {
		t.Error("expected current scene to be the title scene spy")
	}
}

func TestResultScene_ButtonsDoNotOverlap(t *testing.T) {
	rs := NewResultScene(1088, 728, ResultData{}, nil, nil, nil)
	retry := rs.RetryRect()
	title := rs.TitleRect()

	if retry.Overlaps(title) {
		t.Errorf("Retry button %v overlaps Title button %v", retry, title)
	}
}
