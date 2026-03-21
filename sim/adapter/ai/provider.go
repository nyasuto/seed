package ai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// AIProvider implements server.ActionProvider for AI Mode.
// It communicates via JSON Lines over stdin/stdout, sending StateMessages
// and receiving ActionMessages.
type AIProvider struct {
	in      *bufio.Scanner
	out     io.Writer
	builder *StateBuilder

	// timeout is the duration to wait for an action response.
	// Zero means no timeout.
	timeout time.Duration

	// maxRetries is the maximum number of retries on invalid input
	// before falling back to a wait action.
	maxRetries int
}

// NewAIProvider creates an AIProvider that reads from in and writes to out.
// The StateBuilder is used to construct StateMessages from snapshots.
func NewAIProvider(in io.Reader, out io.Writer, builder *StateBuilder) *AIProvider {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer
	return &AIProvider{
		in:         scanner,
		out:        out,
		builder:    builder,
		maxRetries: 3,
	}
}

// SetTimeout sets the timeout for reading action input.
// Zero means no timeout (wait indefinitely).
func (ap *AIProvider) SetTimeout(d time.Duration) {
	ap.timeout = d
}

// ProvideActions implements server.ActionProvider. It writes a StateMessage
// to stdout and reads an ActionMessage from stdin. On invalid input it
// sends an ErrorMessage and retries. On EOF, it returns io.EOF to signal
// the game should end.
func (ap *AIProvider) ProvideActions(snapshot scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	stateMsg, err := ap.builder.BuildStateMessage(snapshot)
	if err != nil {
		return nil, fmt.Errorf("build state message: %w", err)
	}
	if stateMsg == nil {
		return []simulation.PlayerAction{simulation.NoAction{}}, nil
	}

	if err := ap.writeLine(stateMsg); err != nil {
		return nil, fmt.Errorf("write state message: %w", err)
	}

	for attempt := 0; attempt <= ap.maxRetries; attempt++ {
		line, err := ap.readLine()
		if err == io.EOF {
			return nil, io.EOF
		}
		if err != nil {
			return nil, fmt.Errorf("read action: %w", err)
		}

		actions, parseErr := ap.parseAndValidate(line, stateMsg.ValidActions)
		if parseErr != nil {
			errMsg := NewErrorMessage(parseErr.Error())
			if wErr := ap.writeLine(errMsg); wErr != nil {
				return nil, fmt.Errorf("write error message: %w", wErr)
			}
			// Re-send state message so the client can retry.
			if wErr := ap.writeLine(stateMsg); wErr != nil {
				return nil, fmt.Errorf("write state message: %w", wErr)
			}
			continue
		}

		return actions, nil
	}

	// Max retries exceeded: fall back to wait.
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}

// OnTickComplete implements server.ActionProvider. In AI Mode this is a
// no-op because all information is included in the next StateMessage.
func (ap *AIProvider) OnTickComplete(_ scenario.GameSnapshot) {}

// OnGameEnd implements server.ActionProvider. It writes a GameEndMessage
// to stdout.
func (ap *AIProvider) OnGameEnd(result simulation.RunResult) {
	resultStr := "defeat"
	if result.Result.Status == simulation.Won {
		resultStr = "victory"
	}

	summaryJSON, _ := json.Marshal(result.Statistics)
	msg := NewGameEndMessage(resultStr, summaryJSON, nil)
	_ = ap.writeLine(msg)
}

// writeLine marshals v to JSON and writes it as a single line.
func (ap *AIProvider) writeLine(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(ap.out, "%s\n", data)
	return err
}

// readLine reads a single line from the input. It respects the configured
// timeout by using a goroutine and channel.
func (ap *AIProvider) readLine() (string, error) {
	if ap.timeout <= 0 {
		if !ap.in.Scan() {
			if err := ap.in.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		return ap.in.Text(), nil
	}

	type scanResult struct {
		text string
		err  error
	}
	ch := make(chan scanResult, 1)
	go func() {
		if !ap.in.Scan() {
			if err := ap.in.Err(); err != nil {
				ch <- scanResult{err: err}
			} else {
				ch <- scanResult{err: io.EOF}
			}
			return
		}
		ch <- scanResult{text: ap.in.Text()}
	}()

	select {
	case r := <-ch:
		return r.text, r.err
	case <-time.After(ap.timeout):
		// Timeout: return empty to trigger wait action.
		return "", errTimeout
	}
}

// errTimeout is a sentinel error for action input timeout.
var errTimeout = fmt.Errorf("action input timeout")

// parseAndValidate parses a JSON line into an ActionMessage, validates
// the actions against the valid action list, and converts them to
// PlayerAction values.
func (ap *AIProvider) parseAndValidate(line string, validActions []ValidAction) ([]simulation.PlayerAction, error) {
	var msg ActionMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if msg.Type != "action" {
		return nil, fmt.Errorf("expected message type \"action\", got %q", msg.Type)
	}
	if len(msg.Actions) == 0 {
		return nil, fmt.Errorf("actions array must not be empty")
	}

	var actions []simulation.PlayerAction
	for _, ad := range msg.Actions {
		if !isValidAction(ad, validActions) {
			return nil, fmt.Errorf("action %q with params %v is not in valid_actions", ad.Kind, ad.Params)
		}
		pa, err := convertAction(ad)
		if err != nil {
			return nil, err
		}
		actions = append(actions, pa)
	}
	return actions, nil
}

// isValidAction checks if an ActionDef matches any entry in the valid actions list.
func isValidAction(ad ActionDef, validActions []ValidAction) bool {
	for _, va := range validActions {
		if va.Kind == ad.Kind && matchParams(ad, va) {
			return true
		}
	}
	return false
}

// matchParams checks if the action params match a valid action's params.
func matchParams(ad ActionDef, va ValidAction) bool {
	switch ad.Kind {
	case "wait":
		return true
	case "dig_room":
		return matchNumParam(ad.Params, va.Params, "x") &&
			matchNumParam(ad.Params, va.Params, "y") &&
			matchStrParam(ad.Params, va.Params, "room_type_id")
	case "dig_corridor":
		return matchNumParam(ad.Params, va.Params, "from_room_id") &&
			matchNumParam(ad.Params, va.Params, "to_room_id")
	case "summon_beast":
		return matchStrParam(ad.Params, va.Params, "element")
	case "upgrade_room":
		return matchNumParam(ad.Params, va.Params, "room_id")
	case "evolve_beast":
		return matchNumParam(ad.Params, va.Params, "beast_id") &&
			matchStrParam(ad.Params, va.Params, "to_species_id")
	case "place_beast":
		return matchStrParam(ad.Params, va.Params, "species_id") &&
			matchNumParam(ad.Params, va.Params, "room_id")
	default:
		return false
	}
}

// matchNumParam compares a numeric parameter in both maps, handling JSON
// number types (float64) with integer comparison.
func matchNumParam(adParams, vaParams map[string]any, key string) bool {
	av, ok := adParams[key]
	if !ok {
		return false
	}
	bv, ok := vaParams[key]
	if !ok {
		return false
	}
	return toInt(av) == toInt(bv)
}

// matchStrParam compares a string parameter in both maps.
func matchStrParam(adParams, vaParams map[string]any, key string) bool {
	av, ok := adParams[key]
	if !ok {
		return false
	}
	bv, ok := vaParams[key]
	if !ok {
		return false
	}
	return fmt.Sprint(av) == fmt.Sprint(bv)
}

// toInt converts a value to int, handling float64 (from JSON) and int.
func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(math.Round(n))
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return -1
	}
}

// convertAction converts an ActionDef to a simulation.PlayerAction.
func convertAction(ad ActionDef) (simulation.PlayerAction, error) {
	switch ad.Kind {
	case "wait":
		return simulation.NoAction{}, nil
	case "dig_room":
		return simulation.DigRoomAction{
			RoomTypeID: fmt.Sprint(ad.Params["room_type_id"]),
			Pos: types.Pos{
				X: toInt(ad.Params["x"]),
				Y: toInt(ad.Params["y"]),
			},
			Width:  toInt(ad.Params["width"]),
			Height: toInt(ad.Params["height"]),
		}, nil
	case "dig_corridor":
		return simulation.DigCorridorAction{
			FromRoomID: toInt(ad.Params["from_room_id"]),
			ToRoomID:   toInt(ad.Params["to_room_id"]),
		}, nil
	case "summon_beast":
		elem, err := parseElement(fmt.Sprint(ad.Params["element"]))
		if err != nil {
			return nil, err
		}
		return simulation.SummonBeastAction{Element: elem}, nil
	case "upgrade_room":
		return simulation.UpgradeRoomAction{
			RoomID: toInt(ad.Params["room_id"]),
		}, nil
	case "evolve_beast":
		return simulation.EvolveBeastAction{
			BeastID: toInt(ad.Params["beast_id"]),
		}, nil
	case "place_beast":
		return simulation.PlaceBeastAction{
			SpeciesID: fmt.Sprint(ad.Params["species_id"]),
			RoomID:    toInt(ad.Params["room_id"]),
		}, nil
	default:
		return nil, fmt.Errorf("unknown action kind: %s", ad.Kind)
	}
}

// parseElement converts a string element name to types.Element.
func parseElement(s string) (types.Element, error) {
	switch strings.ToLower(s) {
	case "wood":
		return types.Wood, nil
	case "fire":
		return types.Fire, nil
	case "earth":
		return types.Earth, nil
	case "metal":
		return types.Metal, nil
	case "water":
		return types.Water, nil
	default:
		return 0, fmt.Errorf("unknown element: %s", s)
	}
}
