package sim

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/adapter/ai"
	"github.com/nyasuto/seed/sim/adapter/batch"
	"github.com/nyasuto/seed/sim/adapter/human"
	"github.com/nyasuto/seed/sim/balance"
	"github.com/nyasuto/seed/sim/server"
)

// Phase 6: 全モード統合テスト
//
// Task 6-A の 8 項目を検証する:
//   1. Human Mode: スクリプト入力でチュートリアルシナリオ完走
//   2. AI Mode: パイプ経由でチュートリアルシナリオ完走
//   3. Batch Mode: 1,000ゲーム × SimpleAI → BreakageReport 検証
//   4. Balance: ダッシュボード起動 → BreakageReport 表示
//   5. リプレイ: 記録 → 再生 → 決定論性確認
//   6. チェックポイント: Human Mode で保存 → 復元 → 続行
//   7. go test -race (テスト自体が -race 対応)
//   8. go vet (CI / make check で検証)

// --- 1. Human Mode ---

func TestPhase6_HumanMode_TutorialComplete(t *testing.T) {
	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := server.NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Scripted: dig room → summon beast → fast-forward to end.
	input := "1\n4\n10\n10\n3\n1\n6\n300\n"

	out := &bytes.Buffer{}
	ir := human.NewInputReader(strings.NewReader(input), out)
	ctxBuilder := server.NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, out, ctxBuilder)
	provider.SetCheckpointOps(server.NewServerCheckpointOps(gs))

	result, err := gs.RunGame(provider)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	assertTerminalResult(t, "HumanMode", result)

	output := out.String()
	if !strings.Contains(output, "メインメニュー") {
		t.Error("expected main menu in output")
	}
	if !strings.Contains(output, "========") {
		t.Error("expected game end separator")
	}

	t.Logf("Human: status=%v ticks=%d", result.Result.Status, result.TickCount)
}

// --- 2. AI Mode ---

func TestPhase6_AIMode_TutorialComplete(t *testing.T) {
	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := server.NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	builder := ai.NewStateBuilder(gs.Engine)
	provider := ai.NewAIProvider(inR, outW, builder)

	clientErr := make(chan error, 1)
	go func() {
		defer inW.Close()
		client := newTestAIClient(outR, inW)
		for {
			msg, err := client.readMessage()
			if err != nil {
				clientErr <- err
				return
			}
			switch msg["type"] {
			case "state":
				if err := client.waitAction(); err != nil {
					clientErr <- fmt.Errorf("send wait: %w", err)
					return
				}
			case "game_end":
				clientErr <- nil
				return
			}
		}
	}()

	result, err := gs.RunGame(provider)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}
	if cErr := <-clientErr; cErr != nil {
		t.Fatalf("client error: %v", cErr)
	}

	assertTerminalResult(t, "AIMode", result)
	t.Logf("AI: status=%v ticks=%d", result.Result.Status, result.TickCount)
}

// --- 3. Batch Mode: 1,000 games × SimpleAI ---

func TestPhase6_BatchMode_1000Games(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 1000-game batch in short mode")
	}

	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	runner, err := batch.NewBatchRunner(batch.BatchConfig{
		Scenario: sc,
		Games:    1000,
		BaseSeed: 42,
		AI:       batch.AISimple,
		Parallel: 0,
	})
	if err != nil {
		t.Fatalf("NewBatchRunner: %v", err)
	}

	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Summaries) != 1000 {
		t.Fatalf("expected 1000 summaries, got %d", len(result.Summaries))
	}

	for i, s := range result.Summaries {
		if s.Result != simulation.Won && s.Result != simulation.Lost {
			t.Errorf("game %d: unexpected result %v", i, s.Result)
		}
	}

	// Known tutorial-specific alerts (acceptable due to easy scenario design).
	knownTutorialAlerts := map[string]bool{
		"B03": true, // low terrain density
		"B05": true, // single wave, no construction overlap
		"B11": true, // generous chi surplus
	}

	for _, a := range result.BreakageReport.Alerts {
		if !knownTutorialAlerts[a.MetricID] {
			t.Errorf("unexpected alert %s: %s (value=%.4f, threshold=%.4f)",
				a.MetricID, a.BrokenSign, a.Value, a.Threshold)
		}
	}

	t.Logf("Batch 1000: alerts=%d (known=%d), clean=%d",
		len(result.BreakageReport.Alerts),
		len(knownTutorialAlerts),
		len(result.BreakageReport.Clean))
}

// --- 4. Balance Dashboard ---

func TestPhase6_Balance_DashboardBreakageReport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping balance dashboard in short mode")
	}

	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	var output strings.Builder
	config := balance.DashboardConfig{
		Scenario:     sc,
		ScenarioName: "tutorial",
		Games:        50,
		AI:           batch.AISimple,
		BaseSeed:     42,
		Output:       &output,
	}

	dash, err := balance.NewDashboard(config)
	if err != nil {
		t.Fatalf("NewDashboard: %v", err)
	}

	baseline, err := dash.Run()
	if err != nil {
		t.Fatalf("Dashboard.Run: %v", err)
	}

	if len(baseline.BatchResult.Summaries) != 50 {
		t.Errorf("expected 50 summaries, got %d", len(baseline.BatchResult.Summaries))
	}

	out := output.String()
	if !strings.Contains(out, "Breakage Report:") {
		t.Error("missing breakage report in dashboard output")
	}

	t.Logf("Balance: WinRate=%.1f%% AvgTicks=%.0f Alerts=%d",
		baseline.WinRate*100, baseline.AvgTicks,
		len(baseline.BatchResult.BreakageReport.Alerts))
}

// --- 5. Replay: record → replay → determinism ---

// replaySavingProvider wraps an ActionProvider and saves the replay file
// during OnGameEnd (when the engine is still active).
type replaySavingProvider struct {
	server.ActionProvider
	gs         *server.GameServer
	replayPath string
	saveErr    error
}

func (p *replaySavingProvider) OnGameEnd(result simulation.RunResult) {
	p.ActionProvider.OnGameEnd(result)
	p.saveErr = p.gs.SaveReplayTo(p.replayPath)
}

func TestPhase6_Replay_Determinism(t *testing.T) {
	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	dir := t.TempDir()
	replayPath := filepath.Join(dir, "replay.json")

	// Run 1: play full game with a noop provider, save replay at game end.
	gs1, err := server.NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Use scripted Human input: wait → fast-forward to end.
	input := "5\n6\n300\n"
	out := &bytes.Buffer{}
	ir := human.NewInputReader(strings.NewReader(input), out)
	ctxBuilder := server.NewGameContextBuilder(gs1)
	baseProvider := human.NewHumanProvider(ir, out, ctxBuilder)
	baseProvider.SetCheckpointOps(server.NewServerCheckpointOps(gs1))

	rsp := &replaySavingProvider{
		ActionProvider: baseProvider,
		gs:             gs1,
		replayPath:     replayPath,
	}

	originalResult, err := gs1.RunGame(rsp)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}
	if rsp.saveErr != nil {
		t.Fatalf("SaveReplayTo: %v", rsp.saveErr)
	}
	assertTerminalResult(t, "Replay-Original", originalResult)

	// Run 2: replay playback.
	gs2, err := server.NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer for replay: %v", err)
	}

	replayResult, err := gs2.PlayReplayFrom(replayPath)
	if err != nil {
		t.Fatalf("PlayReplayFrom: %v", err)
	}

	// Verify determinism: same status and final tick.
	if originalResult.Result.Status != replayResult.Status {
		t.Errorf("status mismatch: original=%v replay=%v",
			originalResult.Result.Status, replayResult.Status)
	}
	if originalResult.Result.FinalTick != replayResult.FinalTick {
		t.Errorf("final tick mismatch: original=%d replay=%d",
			originalResult.Result.FinalTick, replayResult.FinalTick)
	}

	t.Logf("Replay: original=%v@tick%d, replay=%v@tick%d",
		originalResult.Result.Status, originalResult.Result.FinalTick,
		replayResult.Status, replayResult.FinalTick)
}

// --- 6. Checkpoint: save → load → resume ---

func TestPhase6_Checkpoint_SaveLoadResume(t *testing.T) {
	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	dir := t.TempDir()
	cpPath := filepath.Join(dir, "checkpoint.json")

	gs, err := server.NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Phase 1: play a few ticks, save checkpoint, then trigger load.
	// Script: wait → save → wait → load → (ErrCheckpointLoaded)
	phase1Input := fmt.Sprintf("5\ns\n%s\n5\nl\n%s\n", cpPath, cpPath)
	// Phase 2 (after resume): fast-forward to end.
	phase2Input := "6\n300\n"

	combinedInput := phase1Input + phase2Input
	out := &bytes.Buffer{}
	ir := human.NewInputReader(strings.NewReader(combinedInput), out)
	ctxBuilder := server.NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, out, ctxBuilder)
	provider.SetCheckpointOps(server.NewServerCheckpointOps(gs))

	result, err := gs.RunGame(provider)

	// RunGame returns ErrCheckpointLoaded when checkpoint is loaded.
	if !errors.Is(err, human.ErrCheckpointLoaded) {
		if err != nil {
			t.Fatalf("RunGame: unexpected error %v (expected ErrCheckpointLoaded)", err)
		}
		t.Fatal("RunGame: expected ErrCheckpointLoaded but got nil")
	}

	// Resume from checkpoint (mimics runHumanMode's loop).
	result, err = gs.ResumeGame(provider)
	if err != nil {
		t.Fatalf("ResumeGame: %v", err)
	}

	assertTerminalResult(t, "Checkpoint-Resume", result)

	output := out.String()
	if !strings.Contains(output, "セーブしました") {
		t.Error("expected save confirmation in output")
	}
	if !strings.Contains(output, "ロードしました") {
		t.Error("expected load confirmation in output")
	}

	t.Logf("Checkpoint: status=%v ticks=%d", result.Result.Status, result.TickCount)
}

// --- Helpers (shared with phase4_integration_test.go) ---
// assertTerminalResult and test AI client helpers are defined in
// phase4_integration_test.go. They are reused here since both files
// are in the same package (sim).
