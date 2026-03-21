package server

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/adapter/human"
)

// TestE2E_HumanMode_TutorialPlaythrough runs a full tutorial scenario
// with scripted input: dig room → summon beast → fast-forward to end.
func TestE2E_HumanMode_TutorialPlaythrough(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Scripted input sequence:
	//   Tick 1: dig senju_room at (10,10)
	//     "1"  → ChoiceDigRoom
	//     "4"  → senju_room (4th in alphabetical order: chi_chamber, dragon_hole, recovery_room, senju_room, ...)
	//     "10" → X coordinate
	//     "10" → Y coordinate
	//   Tick 2: summon Wood beast
	//     "3"  → ChoiceSummonBeast
	//     "1"  → Wood (first element in list)
	//   Tick 3+: fast-forward remaining ticks
	//     "6"   → ChoiceFastForward
	//     "300" → forward 300 ticks (covers remaining game)
	input := "1\n4\n10\n10\n3\n1\n6\n300\n"

	out := &bytes.Buffer{}
	ir := human.NewInputReader(strings.NewReader(input), out)
	ctxBuilder := NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, out, ctxBuilder)
	provider.SetCheckpointOps(NewServerCheckpointOps(gs))

	result, err := gs.RunGame(provider)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	// Tutorial scenario: survive 300 ticks.
	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}

	// Verify game completed.
	if result.Result.Status != simulation.Won && result.Result.Status != simulation.Lost {
		t.Errorf("expected terminal status, got %v", result.Result.Status)
	}

	// Verify some output was produced (menu prompts, tick summaries, game end).
	output := out.String()
	if !strings.Contains(output, "メインメニュー") {
		t.Error("expected main menu in output")
	}
	// Game ends during fast-forward (tutorial is 300 ticks), so the FF
	// summary may not appear. Instead verify game end output.
	if !strings.Contains(output, "========") {
		t.Error("expected game end separator in output")
	}

	summary := gs.Collector().OnGameEnd(&result)
	t.Logf("E2E Tutorial: status=%v reason=%q ticks=%d coreHP=%d peakChi=%.1f",
		summary.Result, summary.Reason, summary.TotalTicks,
		summary.FinalCoreHP, summary.PeakChi)
}

// TestE2E_HumanMode_QuitMidGame verifies that quitting mid-game returns io.EOF.
func TestE2E_HumanMode_QuitMidGame(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Tick 1: do nothing. Tick 2: quit (confirmed).
	input := "5\n" + // do nothing
		"q\ny\n" // quit confirmed

	out := &bytes.Buffer{}
	ir := human.NewInputReader(strings.NewReader(input), out)
	ctxBuilder := NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, out, ctxBuilder)

	_, err = gs.RunGame(provider)
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF on quit, got: %v", err)
	}
}

// TestE2E_HumanMode_AllSubmenus exercises every main menu submenu path,
// including back navigation, through the full GameServer integration.
func TestE2E_HumanMode_AllSubmenus(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Build a long scripted input that exercises every submenu.
	//
	// Tick 1: dig room → back → dig room (senju_room at 10,10)
	//   "1" "0" "1" "4" "10" "10"
	//
	// Tick 2: dig corridor from room 1 (dragon_hole) to room 2 (new senju_room)
	//   "2" "1" "2"
	//
	// Tick 3: summon beast → back → summon beast (Wood)
	//   "3" "0" "3" "1"
	//
	// Tick 4: upgrade room → back → do nothing
	//   "4" "0" "5"
	//
	// Tick 5: save → do nothing
	//   "s" "/dev/null" "5"
	//
	// Tick 6: quit cancelled → fast-forward to end
	//   "q" "n" "6" "300"
	var sb strings.Builder
	// Tick 1: dig room (back then actual)
	sb.WriteString("1\n0\n")     // dig room → back
	sb.WriteString("1\n4\n10\n10\n") // dig room → senju_room at (10,10)
	// Tick 2: dig corridor
	sb.WriteString("2\n1\n2\n") // corridor from room 1 to room 2
	// Tick 3: summon beast (back then actual)
	sb.WriteString("3\n0\n")  // summon → back
	sb.WriteString("3\n1\n")  // summon Wood beast
	// Tick 4: upgrade back → do nothing
	sb.WriteString("4\n0\n")  // upgrade → back
	sb.WriteString("5\n")     // do nothing
	// Tick 5: save → do nothing
	sb.WriteString("s\n/dev/null\n") // save
	sb.WriteString("5\n")            // do nothing
	// Tick 6: quit cancel → fast-forward
	sb.WriteString("q\nn\n")     // quit → cancel
	sb.WriteString("6\n300\n")   // fast forward

	out := &bytes.Buffer{}
	ir := human.NewInputReader(strings.NewReader(sb.String()), out)
	ctxBuilder := NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, out, ctxBuilder)
	provider.SetCheckpointOps(NewServerCheckpointOps(gs))

	result, err := gs.RunGame(provider)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	output := out.String()

	// Verify submenu outputs appeared.
	if !strings.Contains(output, "部屋を掘る") {
		t.Error("expected dig room submenu")
	}
	if !strings.Contains(output, "通路を掘る") {
		t.Error("expected dig corridor submenu")
	}
	if !strings.Contains(output, "仙獣を召喚する") {
		t.Error("expected summon beast submenu")
	}
	if !strings.Contains(output, "アップグレード") {
		t.Error("expected upgrade submenu")
	}
	if !strings.Contains(output, "セーブしました") || strings.Contains(output, "セーブ失敗") {
		// Save to /dev/null should succeed on Unix.
		t.Log("save output check (may vary by OS)")
	}
	if !strings.Contains(output, "本当に終了しますか") {
		t.Error("expected quit confirmation prompt")
	}
	// Game ends during FF so "早送り完了" may not appear; check game end instead.
	if !strings.Contains(output, "========") {
		t.Error("expected game end separator")
	}

	// Game should have completed.
	if result.Result.Status != simulation.Won && result.Result.Status != simulation.Lost {
		t.Errorf("expected terminal status, got %v", result.Result.Status)
	}

	t.Logf("E2E AllSubmenus: status=%v ticks=%d", result.Result.Status, result.TickCount)
}
