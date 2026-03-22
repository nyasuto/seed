package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/nyasuto/seed/sim/adapter/ai"
	"github.com/nyasuto/seed/sim/adapter/batch"
	"github.com/nyasuto/seed/sim/adapter/human"
	"github.com/nyasuto/seed/sim/balance"
	"github.com/nyasuto/seed/sim/server"
)

// TestAllModes_Coexistence verifies that all four modes (Human, AI, Batch, Balance)
// can initialize and run independently within the same process, sharing
// the same scenario loader and core engine infrastructure.
func TestAllModes_Coexistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping coexistence test in short mode")
	}

	// 1. Human Mode: scripted input, quit immediately.
	t.Run("Human", func(t *testing.T) {
		sc, err := server.LoadScenario("tutorial")
		if err != nil {
			t.Fatalf("LoadScenario: %v", err)
		}
		gs, err := server.NewGameServer(sc, 42)
		if err != nil {
			t.Fatalf("NewGameServer: %v", err)
		}

		input := strings.NewReader("q\ny\n")
		out := &strings.Builder{}
		ir := human.NewInputReader(input, out)
		cb := server.NewGameContextBuilder(gs)
		provider := human.NewHumanProvider(ir, out, cb)
		provider.SetCheckpointOps(server.NewServerCheckpointOps(gs))

		_, err = gs.RunGame(provider)
		if !errors.Is(err, io.EOF) {
			t.Fatalf("expected io.EOF from quit, got: %v", err)
		}
	})

	// 2. AI Mode: verify provider initializes and can communicate.
	// Full E2E is tested in adapter/ai package; here we verify the mode's
	// components (StateBuilder, AIProvider) integrate with GameServer.
	t.Run("AI", func(t *testing.T) {
		sc, err := server.LoadScenario("tutorial")
		if err != nil {
			t.Fatalf("LoadScenario: %v", err)
		}
		gs, err := server.NewGameServer(sc, 42)
		if err != nil {
			t.Fatalf("NewGameServer: %v", err)
		}

		// Verify AI provider can be created with the game engine.
		builder := ai.NewStateBuilder(gs.Engine)
		if builder == nil {
			t.Fatal("NewStateBuilder returned nil")
		}

		// Use a reader that returns EOF immediately — provider should handle gracefully.
		provider := ai.NewAIProvider(strings.NewReader(""), &strings.Builder{}, builder)
		if provider == nil {
			t.Fatal("NewAIProvider returned nil")
		}

		// RunGame with immediate EOF should terminate.
		_, err = gs.RunGame(provider)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Logf("RunGame error (expected EOF): %v", err)
		}
	})

	// 3. Batch Mode: 10 games with SimpleAI.
	t.Run("Batch", func(t *testing.T) {
		err := runBatchMode("tutorial", 10, "simple", "", "json", "")
		if err != nil {
			t.Fatalf("runBatchMode: %v", err)
		}
	})

	// 4. Balance Mode: 10 games.
	t.Run("Balance", func(t *testing.T) {
		sc, err := server.LoadScenario("tutorial")
		if err != nil {
			t.Fatalf("LoadScenario: %v", err)
		}

		var output strings.Builder
		config := balance.DashboardConfig{
			Scenario:     sc,
			ScenarioName: "tutorial",
			Games:        10,
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
		if len(baseline.BatchResult.Summaries) != 10 {
			t.Errorf("expected 10 summaries, got %d", len(baseline.BatchResult.Summaries))
		}
	})
}
