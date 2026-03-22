package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/adapter/ai"
)

// aiClient simulates an AI client that reads state messages from the server
// and writes action responses via JSON Lines over pipes.
type aiClient struct {
	outR    io.Reader // reads server output
	inW     io.Writer // writes to server input
	scanner *bufio.Scanner
}

func newAIClient(outR io.Reader, inW io.Writer) *aiClient {
	scanner := bufio.NewScanner(outR)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	return &aiClient{outR: outR, inW: inW, scanner: scanner}
}

// readMessage reads the next JSON Lines message from the server.
func (c *aiClient) readMessage() (map[string]any, error) {
	if !c.scanner.Scan() {
		if err := c.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	var msg map[string]any
	if err := json.Unmarshal(c.scanner.Bytes(), &msg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return msg, nil
}

// sendAction writes an action message to the server.
func (c *aiClient) sendAction(actions ...ai.ActionDef) error {
	msg := ai.ActionMessage{
		Type:    "action",
		Actions: actions,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.inW, "%s\n", data)
	return err
}

// waitAction is a convenience for sending a single wait action.
func (c *aiClient) waitAction() error {
	return c.sendAction(ai.ActionDef{Kind: "wait", Params: map[string]any{}})
}

// TestE2E_AIMode_WaitOnly runs a full tutorial scenario where the AI
// client responds with "wait" every tick, completing the game via JSON Lines.
func TestE2E_AIMode_WaitOnly(t *testing.T) {
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

	// Client goroutine: read state messages and respond with wait.
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

	// Wait for client goroutine to finish.
	if cErr := <-clientErr; cErr != nil {
		t.Fatalf("client error: %v", cErr)
	}

	// Tutorial: 300 ticks, survive_until win condition.
	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}
	if result.Result.Status != simulation.Won && result.Result.Status != simulation.Lost {
		t.Errorf("expected terminal status, got %v", result.Result.Status)
	}

	t.Logf("E2E AI WaitOnly: status=%v reason=%q ticks=%d",
		result.Result.Status, result.Result.Reason, result.TickCount)
}

// TestE2E_AIMode_BasicStrategy runs a tutorial scenario where the AI client
// performs dig_room + summon_beast in the first few ticks, then waits.
func TestE2E_AIMode_BasicStrategy(t *testing.T) {
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

	// Client goroutine: implement a basic strategy.
	clientErr := make(chan error, 1)
	go func() {
		defer inW.Close()
		client := newAIClient(outR, inW)
		tick := 0
		digDone := false
		summonDone := false

		for {
			msg, err := client.readMessage()
			if err != nil {
				clientErr <- err
				return
			}

			switch msg["type"] {
			case "state":
				tick++
				var stateMsg ai.StateMessage
				raw, _ := json.Marshal(msg)
				if err := json.Unmarshal(raw, &stateMsg); err != nil {
					clientErr <- fmt.Errorf("parse state: %w", err)
					return
				}

				action := chooseAction(stateMsg.ValidActions, &digDone, &summonDone)
				if err := client.sendAction(action); err != nil {
					clientErr <- fmt.Errorf("send action at tick %d: %w", tick, err)
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

	t.Logf("E2E AI BasicStrategy: status=%v reason=%q ticks=%d",
		result.Result.Status, result.Result.Reason, result.TickCount)
}

// chooseAction picks an action from valid_actions based on a simple strategy:
// 1. First opportunity: dig a room (pick the first dig_room action)
// 2. Second opportunity: summon a beast (pick the first summon_beast action)
// 3. Otherwise: wait
func chooseAction(validActions []ai.ValidAction, digDone, summonDone *bool) ai.ActionDef {
	if !*digDone {
		for _, va := range validActions {
			if va.Kind == "dig_room" {
				*digDone = true
				return ai.ActionDef{Kind: va.Kind, Params: va.Params}
			}
		}
	}

	if !*summonDone {
		for _, va := range validActions {
			if va.Kind == "summon_beast" {
				*summonDone = true
				return ai.ActionDef{Kind: va.Kind, Params: va.Params}
			}
		}
	}

	return ai.ActionDef{Kind: "wait", Params: map[string]any{}}
}
