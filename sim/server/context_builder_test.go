package server

import (
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/adapter/human"
)

func TestGameContextBuilder_BuildCtx_NilEngine(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	cb := NewGameContextBuilder(gs)
	// Engine is nil before game starts.
	ctx := cb.BuildCtx(scenario.GameSnapshot{})
	if len(ctx.RoomTypes) != 0 {
		t.Error("expected empty room types when engine is nil")
	}
	unitCtx := cb.UnitCtx(scenario.GameSnapshot{})
	if len(unitCtx.SummonOptions) != 0 {
		t.Error("expected empty summon options when engine is nil")
	}
}

func TestGameContextBuilder_DuringGame(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	cb := NewGameContextBuilder(gs)

	// Use a provider that captures context mid-game.
	var capturedBuild human.BuildContext
	var capturedUnit human.UnitContext
	provider := &capturingProvider{
		cb: cb,
		onCapture: func(b human.BuildContext, u human.UnitContext) {
			capturedBuild = b
			capturedUnit = u
		},
	}

	_, _ = gs.RunGame(provider)

	// BuildContext assertions
	if capturedBuild.CaveWidth == 0 {
		t.Error("expected non-zero CaveWidth")
	}
	if capturedBuild.CaveHeight == 0 {
		t.Error("expected non-zero CaveHeight")
	}
	if len(capturedBuild.RoomTypes) == 0 {
		t.Error("expected room types")
	}
	if len(capturedBuild.Rooms) == 0 {
		t.Error("expected at least one room (dragon hole)")
	}

	// Verify dragon hole is present
	found := false
	for _, r := range capturedBuild.Rooms {
		if r.TypeID == "dragon_hole" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected dragon_hole room in rooms list")
	}

	// UnitContext assertions
	if len(capturedUnit.SummonOptions) == 0 {
		t.Error("expected summon options")
	}
	if capturedUnit.ChiBalance < 0 {
		t.Error("expected non-negative chi balance")
	}
}

func TestServerCheckpointOps_SaveLoadReplay(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	ops := NewServerCheckpointOps(gs)

	// No engine → error
	if err := ops.SaveCheckpoint("/tmp/nope.json"); err == nil {
		t.Error("expected error when no engine active")
	}
	if err := ops.SaveReplay("/tmp/nope.json"); err == nil {
		t.Error("expected error when no engine active")
	}

	// Run game with a provider that saves/loads mid-game
	dir := t.TempDir()
	savePath := dir + "/save.json"
	replayPath := dir + "/replay.json"

	provider := &checkpointTestProvider{
		ops:        ops,
		savePath:   savePath,
		replayPath: replayPath,
	}
	_, _ = gs.RunGame(provider)

	if !provider.saveOK {
		t.Error("expected save to succeed during game")
	}
	if !provider.replayOK {
		t.Error("expected replay save to succeed during game")
	}

	// Load the checkpoint.
	if err := ops.LoadCheckpoint(savePath); err != nil {
		t.Fatalf("LoadCheckpoint: %v", err)
	}
	if gs.Engine() == nil {
		t.Error("expected engine to be restored after load")
	}
}

// capturingProvider captures BuildCtx/UnitCtx on the first tick, then plays NoAction.
type capturingProvider struct {
	cb        *GameContextBuilder
	onCapture func(human.BuildContext, human.UnitContext)
	captured  bool
}

func (p *capturingProvider) ProvideActions(snap scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	if !p.captured {
		p.captured = true
		b := p.cb.BuildCtx(snap)
		u := p.cb.UnitCtx(snap)
		p.onCapture(b, u)
	}
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}
func (p *capturingProvider) OnTickComplete(_ scenario.GameSnapshot) {}
func (p *capturingProvider) OnGameEnd(_ simulation.RunResult)       {}

// checkpointTestProvider saves checkpoint and replay on first tick.
type checkpointTestProvider struct {
	ops        human.CheckpointOps
	savePath   string
	replayPath string
	saveOK     bool
	replayOK   bool
	done       bool
}

func (p *checkpointTestProvider) ProvideActions(_ scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	if !p.done {
		p.done = true
		if err := p.ops.SaveCheckpoint(p.savePath); err == nil {
			p.saveOK = true
		}
		if err := p.ops.SaveReplay(p.replayPath); err == nil {
			p.replayOK = true
		}
	}
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}
func (p *checkpointTestProvider) OnTickComplete(_ scenario.GameSnapshot) {}
func (p *checkpointTestProvider) OnGameEnd(_ simulation.RunResult)       {}
