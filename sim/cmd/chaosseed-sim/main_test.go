package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/nyasuto/seed/sim/adapter/human"
	"github.com/nyasuto/seed/sim/server"
)

func TestRunHumanMode_Starts(t *testing.T) {
	// Verify that the game starts with tutorial scenario.
	// We can't run the full interactive mode, but we can test the setup
	// by creating the components and verifying they connect.
	sc, err := server.LoadScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}
	gs, err := server.NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Create a HumanProvider with scripted input: just quit immediately.
	input := strings.NewReader("q\ny\n")
	out := &strings.Builder{}
	ir := human.NewInputReader(input, out)
	cb := server.NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, out, cb)
	provider.SetCheckpointOps(server.NewServerCheckpointOps(gs))

	_, err = gs.RunGame(provider)
	// Player quit → io.EOF wrapped in error.
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF from quit, got: %v", err)
	}
}

func TestRunHumanMode_SaveAndLoad(t *testing.T) {
	sc, err := server.LoadScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}
	gs, err := server.NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	dir := t.TempDir()
	savePath := dir + "/save.json"

	// Play 1 tick (do nothing), save, then quit.
	input := strings.NewReader("5\ns\n" + savePath + "\nq\ny\n")
	out := &strings.Builder{}
	ir := human.NewInputReader(input, out)
	cb := server.NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, out, cb)
	provider.SetCheckpointOps(server.NewServerCheckpointOps(gs))

	_, err = gs.RunGame(provider)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF from quit, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "セーブしました") {
		t.Errorf("expected save confirmation in output")
	}

	// Now load the checkpoint and resume.
	gs2, err := server.NewGameServer(sc, 0)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	ops2 := server.NewServerCheckpointOps(gs2)

	// Load, then immediately quit.
	input2 := strings.NewReader("l\n" + savePath + "\nq\ny\n")
	out2 := &strings.Builder{}
	ir2 := human.NewInputReader(input2, out2)
	cb2 := server.NewGameContextBuilder(gs2)
	provider2 := human.NewHumanProvider(ir2, out2, cb2)
	provider2.SetCheckpointOps(ops2)

	// RunGame starts a new game; when provider loads checkpoint, it returns ErrCheckpointLoaded.
	_, err = gs2.RunGame(provider2)
	if !errors.Is(err, human.ErrCheckpointLoaded) {
		t.Fatalf("expected ErrCheckpointLoaded, got: %v", err)
	}

	// Resume after load.
	_, err = gs2.ResumeGame(provider2)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF from quit after resume, got: %v", err)
	}
}

func TestRunHumanMode_ReplaySave(t *testing.T) {
	sc, err := server.LoadScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}
	gs, err := server.NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	dir := t.TempDir()
	replayPath := dir + "/replay.json"

	// Play 1 tick, save replay, then quit.
	input := strings.NewReader("5\nr\n" + replayPath + "\nq\ny\n")
	out := &strings.Builder{}
	ir := human.NewInputReader(input, out)
	cb := server.NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, out, cb)
	provider.SetCheckpointOps(server.NewServerCheckpointOps(gs))

	_, err = gs.RunGame(provider)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF from quit, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "リプレイを保存しました") {
		t.Errorf("expected replay confirmation in output")
	}

	// Verify replay file exists and can be played back.
	gs2, err := server.NewGameServer(sc, 0)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	result, err := gs2.PlayReplayFrom(replayPath)
	if err != nil {
		t.Fatalf("PlayReplayFrom: %v", err)
	}
	t.Logf("Replay result: %v (tick %d)", result.Status, result.FinalTick)
}
