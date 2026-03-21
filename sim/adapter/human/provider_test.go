package human

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// stubContextBuilder is a test double that returns fixed contexts.
type stubContextBuilder struct {
	build BuildContext
	unit  UnitContext
}

func (s *stubContextBuilder) BuildCtx(_ scenario.GameSnapshot) BuildContext { return s.build }
func (s *stubContextBuilder) UnitCtx(_ scenario.GameSnapshot) UnitContext   { return s.unit }

func defaultStubCtxBuilder() *stubContextBuilder {
	return &stubContextBuilder{
		build: BuildContext{
			RoomTypes: []RoomTypeOption{
				{TypeID: "senju_room", Name: "仙獣部屋", Element: types.Wood, Cost: 30.0},
			},
			Rooms: []RoomInfo{
				{ID: 1, TypeID: "dragon_hole", Name: "龍穴", Pos: types.Pos{X: 5, Y: 5}},
				{ID: 2, TypeID: "senju_room", Name: "仙獣部屋", Pos: types.Pos{X: 10, Y: 5}},
			},
			ChiBalance: 100.0,
			CaveWidth:  30,
			CaveHeight: 30,
		},
		unit: UnitContext{
			SummonOptions: []SummonOption{
				{Element: types.Wood, Cost: 20.0},
			},
			Rooms: []RoomInfo{
				{ID: 1, TypeID: "dragon_hole", Name: "龍穴", Pos: types.Pos{X: 5, Y: 5}},
			},
			ChiBalance: 100.0,
		},
	}
}

func testSnapshot(tick int) scenario.GameSnapshot {
	return scenario.GameSnapshot{
		Tick:              types.Tick(tick),
		CoreHP:            100,
		ChiPoolBalance:    50.0,
		BeastCount:        2,
		AliveBeasts:       2,
		DefeatedWaves:     0,
		TotalWaves:        5,
		CaveFengShuiScore: 0.5,
	}
}

func newHumanProvider(input string) (*HumanProvider, *bytes.Buffer) {
	out := &bytes.Buffer{}
	ir := NewInputReader(strings.NewReader(input), out)
	cb := defaultStubCtxBuilder()
	hp := NewHumanProvider(ir, out, cb)
	return hp, out
}

func TestHumanProvider_DoNothing(t *testing.T) {
	// Select "5" (do nothing).
	hp, _ := newHumanProvider("5\n")
	snap := testSnapshot(1)

	actions, err := hp.ProvideActions(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if _, ok := actions[0].(simulation.NoAction); !ok {
		t.Errorf("expected NoAction, got %T", actions[0])
	}
}

func TestHumanProvider_DigRoom(t *testing.T) {
	// Select "1" (dig room), then room type 1, coordinates 4,6.
	hp, _ := newHumanProvider("1\n1\n4\n6\n")
	snap := testSnapshot(1)

	actions, err := hp.ProvideActions(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	dig, ok := actions[0].(simulation.DigRoomAction)
	if !ok {
		t.Fatalf("expected DigRoomAction, got %T", actions[0])
	}
	if dig.RoomTypeID != "senju_room" {
		t.Errorf("expected room type senju_room, got %s", dig.RoomTypeID)
	}
}

func TestHumanProvider_SummonBeast(t *testing.T) {
	// Select "3" (summon beast), then element 1.
	hp, _ := newHumanProvider("3\n1\n")
	snap := testSnapshot(1)

	actions, err := hp.ProvideActions(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	summon, ok := actions[0].(simulation.SummonBeastAction)
	if !ok {
		t.Fatalf("expected SummonBeastAction, got %T", actions[0])
	}
	if summon.Element != types.Wood {
		t.Errorf("expected Wood element, got %v", summon.Element)
	}
}

func TestHumanProvider_FastForward(t *testing.T) {
	// Select "6" (fast forward), then 50 ticks.
	hp, out := newHumanProvider("6\n50\n")
	snap := testSnapshot(1)

	// First call: initiates fast-forward.
	actions, err := hp.ProvideActions(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := actions[0].(simulation.NoAction); !ok {
		t.Fatalf("expected NoAction, got %T", actions[0])
	}

	// Remaining 49 ticks should return NoAction without prompting.
	for i := 0; i < 49; i++ {
		snap.Tick = types.Tick(2 + i)
		actions, err = hp.ProvideActions(snap)
		if err != nil {
			t.Fatalf("tick %d: unexpected error: %v", i, err)
		}
		if _, ok := actions[0].(simulation.NoAction); !ok {
			t.Fatalf("tick %d: expected NoAction, got %T", i, actions[0])
		}
	}

	// After 50 ticks total, OnTickComplete should produce FF summary.
	out.Reset()
	snap.Tick = 50
	hp.OnTickComplete(snap)
	output := out.String()
	if !strings.Contains(output, "早送り完了") {
		t.Errorf("expected fast-forward summary, got: %s", output)
	}
}

func TestHumanProvider_FastForwardSkipsDisplay(t *testing.T) {
	// FF 3 ticks.
	hp, out := newHumanProvider("6\n3\n")
	snap := testSnapshot(1)

	// Tick 1: start FF.
	_, err := hp.ProvideActions(snap)
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()

	// OnTickComplete during FF should produce no output.
	snap.Tick = 1
	hp.OnTickComplete(snap)
	if out.Len() > 0 {
		// During FF (fastForward > 0), display should be skipped.
		// After this tick, fastForward goes from 2 to 1 (still > 0 at next ProvideActions).
	}

	// Tick 2: still FF.
	snap.Tick = 2
	_, err = hp.ProvideActions(snap)
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	hp.OnTickComplete(snap)

	// Tick 3: last FF tick (fastForward decremented to 0 in ProvideActions).
	snap.Tick = 3
	_, err = hp.ProvideActions(snap)
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	snap.Tick = 3
	hp.OnTickComplete(snap)
	output := out.String()
	if !strings.Contains(output, "早送り完了") {
		t.Errorf("expected FF summary after last tick, got: %s", output)
	}
}

func TestHumanProvider_OnTickComplete_ShowsCombatLog(t *testing.T) {
	hp, out := newHumanProvider("")

	prev := testSnapshot(1)
	hp.prevSnapshot = &prev

	// Current tick: CoreHP decreased, a beast was lost.
	current := testSnapshot(2)
	current.CoreHP = 80
	current.AliveBeasts = 1

	hp.OnTickComplete(current)
	output := out.String()

	if !strings.Contains(output, "Tick 2") {
		t.Errorf("expected tick number, got: %s", output)
	}
	if !strings.Contains(output, "-20!") {
		t.Errorf("expected CoreHP decrease indicator, got: %s", output)
	}
	if !strings.Contains(output, "1体 戦闘不能") {
		t.Errorf("expected beast loss warning, got: %s", output)
	}
}

func TestHumanProvider_OnGameEnd(t *testing.T) {
	hp, out := newHumanProvider("")

	result := simulation.RunResult{
		Result: simulation.GameResult{
			Status:    simulation.Won,
			FinalTick: 100,
			Reason:    "all waves defeated",
		},
		TickCount: 100,
		Statistics: simulation.RunStatistics{
			PeakChi:        250.0,
			WavesDefeated:  5,
			FinalFengShui:  0.85,
			Evolutions:     2,
			DamageDealt:    500,
			DamageReceived: 200,
			DeficitTicks:   3,
		},
	}

	hp.OnGameEnd(result)
	output := out.String()

	if !strings.Contains(output, "勝利") {
		t.Errorf("expected victory message, got: %s", output)
	}
	if !strings.Contains(output, "all waves defeated") {
		t.Errorf("expected reason, got: %s", output)
	}
	if !strings.Contains(output, "250.0") {
		t.Errorf("expected peak chi, got: %s", output)
	}
}

func TestHumanProvider_SubMenuBack_ReturnsToMainMenu(t *testing.T) {
	// Select "1" (dig room), then "0" (back), then "5" (do nothing).
	hp, _ := newHumanProvider("1\n0\n5\n")
	snap := testSnapshot(1)

	actions, err := hp.ProvideActions(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := actions[0].(simulation.NoAction); !ok {
		t.Errorf("expected NoAction after back+doNothing, got %T", actions[0])
	}
}

func TestHumanProvider_Quit_Confirmed(t *testing.T) {
	// Select "q", then "y" to confirm.
	hp, _ := newHumanProvider("q\ny\n")
	snap := testSnapshot(1)

	_, err := hp.ProvideActions(snap)
	if err != io.EOF {
		t.Errorf("expected io.EOF on confirmed quit, got: %v", err)
	}
}

func TestHumanProvider_Quit_Cancelled(t *testing.T) {
	// Select "q", then "n" to cancel, then "5" (do nothing).
	hp, _ := newHumanProvider("q\nn\n5\n")
	snap := testSnapshot(1)

	actions, err := hp.ProvideActions(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := actions[0].(simulation.NoAction); !ok {
		t.Errorf("expected NoAction after cancelled quit, got %T", actions[0])
	}
}
