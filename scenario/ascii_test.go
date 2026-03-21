package scenario

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestRenderScenarioStatus_Standard(t *testing.T) {
	sc := &Scenario{
		Difficulty: "standard",
		Constraints: GameConstraints{
			MaxTicks: 500,
		},
		WinConditions: []ConditionDef{
			{Type: "fengshui_score", Params: map[string]any{"threshold": 80.0}},
		},
	}
	prog := &ScenarioProgress{
		CurrentTick: 150,
		CoreHP:      85,
	}
	snap := GameSnapshot{
		DefeatedWaves:     2,
		TotalWaves:        5,
		CaveFengShuiScore: 45,
	}

	got := RenderScenarioStatus(sc, prog, snap)
	want := "[standard | Tick 150/500 | Waves 2/5 | Core HP 85 | Win: FengShui 45/80]"

	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestRenderScenarioStatus_NoMaxTicks(t *testing.T) {
	sc := &Scenario{
		Difficulty: "tutorial",
		WinConditions: []ConditionDef{
			{Type: "defeat_all_waves"},
		},
	}
	prog := &ScenarioProgress{
		CurrentTick: 42,
		CoreHP:      100,
	}
	snap := GameSnapshot{
		DefeatedWaves: 1,
		TotalWaves:    3,
	}

	got := RenderScenarioStatus(sc, prog, snap)
	want := "[tutorial | Tick 42 | Waves 1/3 | Core HP 100]"

	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestRenderScenarioStatus_MultipleWinConditions(t *testing.T) {
	sc := &Scenario{
		Difficulty: "hard",
		Constraints: GameConstraints{
			MaxTicks: types.Tick(1000),
		},
		WinConditions: []ConditionDef{
			{Type: "survive_until", Params: map[string]any{"ticks": 1000.0}},
			{Type: "fengshui_score", Params: map[string]any{"threshold": 60.0}},
			{Type: "chi_pool", Params: map[string]any{"threshold": 200.0}},
		},
	}
	prog := &ScenarioProgress{
		CurrentTick: 300,
		CoreHP:      50,
	}
	snap := GameSnapshot{
		DefeatedWaves:     0,
		TotalWaves:        8,
		CaveFengShuiScore: 30,
		ChiPoolBalance:    120,
	}

	got := RenderScenarioStatus(sc, prog, snap)
	want := "[hard | Tick 300/1000 | Waves 0/8 | Core HP 50 | Win: FengShui 30/60 | Win: Chi 120/200]"

	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestRenderScenarioStatus_ZeroWaves(t *testing.T) {
	sc := &Scenario{
		Difficulty: "peaceful",
		Constraints: GameConstraints{
			MaxTicks: 200,
		},
		WinConditions: []ConditionDef{
			{Type: "survive_until", Params: map[string]any{"ticks": 200.0}},
		},
	}
	prog := &ScenarioProgress{
		CurrentTick: 0,
		CoreHP:      100,
	}
	snap := GameSnapshot{
		DefeatedWaves: 0,
		TotalWaves:    0,
	}

	got := RenderScenarioStatus(sc, prog, snap)
	want := "[peaceful | Tick 0/200 | Waves 0/0 | Core HP 100]"

	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
