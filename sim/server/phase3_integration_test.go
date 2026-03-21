package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/adapter/ai"
	"github.com/nyasuto/seed/sim/adapter/human"
)

// TestPhase3_AIMode_ErrorRetry verifies that the AI provider sends an error
// message and re-sends the state when the client sends invalid input, then
// recovers when the client sends a valid action.
func TestPhase3_AIMode_ErrorRetry(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
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
		client := newAIClient(outR, inW)
		firstTick := true

		for {
			msg, err := client.readMessage()
			if err != nil {
				clientErr <- err
				return
			}

			switch msg["type"] {
			case "state":
				if firstTick {
					// Send invalid JSON first.
					if _, wErr := fmt.Fprintf(inW, "not-valid-json\n"); wErr != nil {
						clientErr <- wErr
						return
					}
					// Read error message.
					errMsg, rErr := client.readMessage()
					if rErr != nil {
						clientErr <- fmt.Errorf("read error msg: %w", rErr)
						return
					}
					if errMsg["type"] != "error" {
						clientErr <- fmt.Errorf("expected error message, got %v", errMsg["type"])
						return
					}
					// Read re-sent state message.
					reState, rErr := client.readMessage()
					if rErr != nil {
						clientErr <- fmt.Errorf("read re-sent state: %w", rErr)
						return
					}
					if reState["type"] != "state" {
						clientErr <- fmt.Errorf("expected re-sent state, got %v", reState["type"])
						return
					}
					firstTick = false
				}
				// Now send a valid wait action.
				if err := client.waitAction(); err != nil {
					clientErr <- err
					return
				}
			case "game_end":
				clientErr <- nil
				return
			case "error":
				// Unexpected error in non-first-tick; ignore and wait for state.
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

	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}
	t.Logf("AIMode ErrorRetry: status=%v ticks=%d", result.Result.Status, result.TickCount)
}

// TestPhase3_AIMode_InvalidAction verifies that sending an action not in
// valid_actions triggers an error+retry cycle.
func TestPhase3_AIMode_InvalidAction(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
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
		client := newAIClient(outR, inW)
		firstTick := true

		for {
			msg, err := client.readMessage()
			if err != nil {
				clientErr <- err
				return
			}

			switch msg["type"] {
			case "state":
				if firstTick {
					// Send an action that is not in valid_actions.
					bogus := ai.ActionDef{
						Kind:   "dig_room",
						Params: map[string]any{"room_type_id": "nonexistent", "x": 0, "y": 0},
					}
					if err := client.sendAction(bogus); err != nil {
						clientErr <- err
						return
					}
					// Read error message.
					errMsg, rErr := client.readMessage()
					if rErr != nil {
						clientErr <- fmt.Errorf("read error: %w", rErr)
						return
					}
					if errMsg["type"] != "error" {
						clientErr <- fmt.Errorf("expected error, got %v", errMsg["type"])
						return
					}
					// Read re-sent state.
					reState, rErr := client.readMessage()
					if rErr != nil {
						clientErr <- fmt.Errorf("read re-sent state: %w", rErr)
						return
					}
					if reState["type"] != "state" {
						clientErr <- fmt.Errorf("expected state, got %v", reState["type"])
						return
					}
					firstTick = false
				}
				if err := client.waitAction(); err != nil {
					clientErr <- err
					return
				}
			case "game_end":
				clientErr <- nil
				return
			case "error":
				// Could get error if we're in retry cycle; wait for state.
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

	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}
	t.Logf("AIMode InvalidAction: status=%v ticks=%d", result.Result.Status, result.TickCount)
}

// TestPhase3_AIMode_StandardScenario runs AI mode with the standard scenario.
func TestPhase3_AIMode_StandardScenario(t *testing.T) {
	sc, err := LoadBuiltinScenario("standard")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
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
		client := newAIClient(outR, inW)
		for {
			msg, err := client.readMessage()
			if err != nil {
				clientErr <- err
				return
			}
			switch msg["type"] {
			case "state":
				if err := client.waitAction(); err != nil {
					clientErr <- err
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

	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}
	if result.Result.Status != simulation.Won && result.Result.Status != simulation.Lost {
		t.Errorf("expected terminal status, got %v", result.Result.Status)
	}

	t.Logf("AIMode Standard: status=%v reason=%q ticks=%d",
		result.Result.Status, result.Result.Reason, result.TickCount)
}

// TestPhase3_AIMode_ValidActionsContent verifies that state messages contain
// well-formed valid_actions with at least a wait action.
func TestPhase3_AIMode_ValidActionsContent(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
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
		client := newAIClient(outR, inW)
		checkedFirstTick := false

		for {
			msg, err := client.readMessage()
			if err != nil {
				clientErr <- err
				return
			}

			switch msg["type"] {
			case "state":
				if !checkedFirstTick {
					var stateMsg ai.StateMessage
					raw, _ := json.Marshal(msg)
					if err := json.Unmarshal(raw, &stateMsg); err != nil {
						clientErr <- fmt.Errorf("parse state: %w", err)
						return
					}

					// Must have at least wait action.
					hasWait := false
					for _, va := range stateMsg.ValidActions {
						if va.Kind == "wait" {
							hasWait = true
						}
						// Every action must have a kind.
						if va.Kind == "" {
							clientErr <- fmt.Errorf("action with empty kind")
							return
						}
					}
					if !hasWait {
						clientErr <- fmt.Errorf("no wait action in valid_actions")
						return
					}

					// Snapshot should be non-empty.
					if len(stateMsg.Snapshot) == 0 {
						clientErr <- fmt.Errorf("snapshot should be non-empty")
						return
					}
					// Parse snapshot to verify core_hp is present and positive.
					var snapMap map[string]any
					if err := json.Unmarshal(stateMsg.Snapshot, &snapMap); err != nil {
						clientErr <- fmt.Errorf("parse snapshot: %w", err)
						return
					}
					if coreHP, ok := snapMap["core_hp"]; ok {
						if hp, ok := coreHP.(float64); ok && hp <= 0 {
							clientErr <- fmt.Errorf("initial CoreHP should be positive, got %v", hp)
							return
						}
					}

					checkedFirstTick = true
				}
				if err := client.waitAction(); err != nil {
					clientErr <- err
					return
				}
			case "game_end":
				if !checkedFirstTick {
					clientErr <- fmt.Errorf("game ended before first state message")
					return
				}
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

	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}
}

// TestPhase3_AIMode_GameEndMessage verifies that the game_end message
// contains result and summary fields.
func TestPhase3_AIMode_GameEndMessage(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
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
		client := newAIClient(outR, inW)
		for {
			msg, err := client.readMessage()
			if err != nil {
				clientErr <- err
				return
			}
			switch msg["type"] {
			case "state":
				if err := client.waitAction(); err != nil {
					clientErr <- err
					return
				}
			case "game_end":
				// Verify game_end fields.
				result, ok := msg["result"].(string)
				if !ok || (result != "victory" && result != "defeat") {
					clientErr <- fmt.Errorf("expected result victory/defeat, got %v", msg["result"])
					return
				}
				if _, ok := msg["summary"]; !ok {
					clientErr <- fmt.Errorf("game_end missing summary field")
					return
				}
				clientErr <- nil
				return
			}
		}
	}()

	if _, err := gs.RunGame(provider); err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	if cErr := <-clientErr; cErr != nil {
		t.Fatalf("client error: %v", cErr)
	}
}

// TestPhase3_ModeCoexistence verifies that the same scenario can be played
// with both Human Mode and AI Mode providers through the same GameServer,
// and both produce valid terminal results.
func TestPhase3_ModeCoexistence(t *testing.T) {
	// --- AI Mode ---
	t.Run("AIMode", func(t *testing.T) {
		sc, err := LoadBuiltinScenario("tutorial")
		if err != nil {
			t.Fatalf("LoadBuiltinScenario: %v", err)
		}

		gs, err := NewGameServer(sc, 42)
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
			client := newAIClient(outR, inW)
			for {
				msg, err := client.readMessage()
				if err != nil {
					clientErr <- err
					return
				}
				switch msg["type"] {
				case "state":
					if err := client.waitAction(); err != nil {
						clientErr <- err
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
			t.Fatalf("RunGame AI: %v", err)
		}
		if cErr := <-clientErr; cErr != nil {
			t.Fatalf("client error: %v", cErr)
		}
		if result.Result.Status != simulation.Won && result.Result.Status != simulation.Lost {
			t.Errorf("AI mode: expected terminal status, got %v", result.Result.Status)
		}
		t.Logf("AI Mode: status=%v ticks=%d", result.Result.Status, result.TickCount)
	})

	// --- Human Mode ---
	t.Run("HumanMode", func(t *testing.T) {
		sc, err := LoadBuiltinScenario("tutorial")
		if err != nil {
			t.Fatalf("LoadBuiltinScenario: %v", err)
		}

		gs, err := NewGameServer(sc, 42)
		if err != nil {
			t.Fatalf("NewGameServer: %v", err)
		}

		// Scripted: do nothing for first tick, then fast-forward to end.
		input := "5\n6\n300\n"

		out := &bytes.Buffer{}
		ir := human.NewInputReader(strings.NewReader(input), out)
		ctxBuilder := NewGameContextBuilder(gs)
		provider := human.NewHumanProvider(ir, out, ctxBuilder)
		provider.SetCheckpointOps(NewServerCheckpointOps(gs))

		result, err := gs.RunGame(provider)
		if err != nil {
			t.Fatalf("RunGame Human: %v", err)
		}

		if result.Result.Status != simulation.Won && result.Result.Status != simulation.Lost {
			t.Errorf("Human mode: expected terminal status, got %v", result.Result.Status)
		}
		t.Logf("Human Mode: status=%v ticks=%d", result.Result.Status, result.TickCount)
	})
}

// TestPhase3_SameScenarioBothModes verifies that both modes produce the same
// game result when given the same seed and same actions (wait every tick).
func TestPhase3_SameScenarioBothModes(t *testing.T) {
	const seed int64 = 12345

	// --- NoAction provider (baseline) ---
	sc1, _ := LoadBuiltinScenario("tutorial")
	gs1, _ := NewGameServer(sc1, seed)
	baseResult, err := gs1.RunGame(&noopAI_provider{})
	if err != nil {
		t.Fatalf("baseline RunGame: %v", err)
	}

	// --- AI Mode (wait every tick) ---
	sc2, _ := LoadBuiltinScenario("tutorial")
	gs2, _ := NewGameServer(sc2, seed)

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	builder := ai.NewStateBuilder(gs2.Engine)
	aiProvider := ai.NewAIProvider(inR, outW, builder)

	clientErr := make(chan error, 1)
	go func() {
		defer inW.Close()
		client := newAIClient(outR, inW)
		for {
			msg, err := client.readMessage()
			if err != nil {
				clientErr <- err
				return
			}
			switch msg["type"] {
			case "state":
				if err := client.waitAction(); err != nil {
					clientErr <- err
					return
				}
			case "game_end":
				clientErr <- nil
				return
			}
		}
	}()

	aiResult, err := gs2.RunGame(aiProvider)
	if err != nil {
		t.Fatalf("AI RunGame: %v", err)
	}
	if cErr := <-clientErr; cErr != nil {
		t.Fatalf("client error: %v", cErr)
	}

	// Both should produce the same tick count and result.
	if baseResult.TickCount != aiResult.TickCount {
		t.Errorf("TickCount mismatch: baseline=%d ai=%d", baseResult.TickCount, aiResult.TickCount)
	}
	if baseResult.Result.Status != aiResult.Result.Status {
		t.Errorf("Status mismatch: baseline=%v ai=%v", baseResult.Result.Status, aiResult.Result.Status)
	}

	t.Logf("Determinism check: baseline=%v/%d, AI=%v/%d",
		baseResult.Result.Status, baseResult.TickCount,
		aiResult.Result.Status, aiResult.TickCount)
}
