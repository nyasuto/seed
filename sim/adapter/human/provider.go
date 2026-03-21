package human

import (
	"io"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// ContextBuilder provides the detailed game context needed by menus.
// The GameSnapshot only contains aggregate data, but submenus require
// detailed room lists, room types, and upgrade options. Implementations
// typically wrap a reference to the live GameState.
type ContextBuilder interface {
	// BuildCtx returns the context needed by build submenus (dig room, dig corridor).
	BuildCtx(snapshot scenario.GameSnapshot) BuildContext
	// UnitCtx returns the context needed by unit submenus (summon, upgrade).
	UnitCtx(snapshot scenario.GameSnapshot) UnitContext
}

// HumanProvider implements server.ActionProvider for Human Mode.
// It drives an interactive text menu each tick, displays tick results,
// and shows game end summaries.
type HumanProvider struct {
	ir         *InputReader
	out        io.Writer
	ctxBuilder ContextBuilder

	// fastForward tracks remaining fast-forward ticks. When > 0,
	// ProvideActions returns NoAction without prompting the player.
	fastForward int

	// prevSnapshot stores the previous tick's snapshot for delta display.
	prevSnapshot *scenario.GameSnapshot

	// ffStartSnapshot stores the snapshot at the start of fast-forward
	// for summary display after fast-forward completes.
	ffStartSnapshot *scenario.GameSnapshot
}

// NewHumanProvider creates a HumanProvider with the given InputReader,
// output writer, and ContextBuilder.
func NewHumanProvider(ir *InputReader, out io.Writer, ctxBuilder ContextBuilder) *HumanProvider {
	return &HumanProvider{
		ir:         ir,
		out:        out,
		ctxBuilder: ctxBuilder,
	}
}

// ProvideActions implements server.ActionProvider. During fast-forward it
// returns NoAction without prompting. Otherwise it shows the main menu
// and dispatches to submenus.
func (hp *HumanProvider) ProvideActions(snapshot scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	// Fast-forward: skip menu interaction.
	if hp.fastForward > 0 {
		hp.fastForward--
		return []simulation.PlayerAction{simulation.NoAction{}}, nil
	}

	for {
		choice, err := ShowMainMenu(hp.ir)
		if err != nil {
			return nil, err
		}

		action, done, err := hp.handleChoice(choice, snapshot)
		if err != nil {
			return nil, err
		}
		if done {
			if action == nil {
				return []simulation.PlayerAction{simulation.NoAction{}}, nil
			}
			return []simulation.PlayerAction{action}, nil
		}
		// Not done means the submenu returned to main menu (e.g. "back"); loop again.
	}
}

// handleChoice dispatches a MenuChoice to the appropriate submenu or action.
// Returns (action, done, error). done=false means re-show main menu.
func (hp *HumanProvider) handleChoice(choice MenuChoice, snapshot scenario.GameSnapshot) (simulation.PlayerAction, bool, error) {
	switch choice {
	case ChoiceDigRoom:
		ctx := hp.ctxBuilder.BuildCtx(snapshot)
		action, err := ShowDigRoomMenu(hp.ir, ctx)
		if err != nil {
			return nil, false, err
		}
		if action == nil {
			return nil, false, nil // back to main menu
		}
		return action, true, nil

	case ChoiceDigCorridor:
		ctx := hp.ctxBuilder.BuildCtx(snapshot)
		action, err := ShowDigCorridorMenu(hp.ir, ctx)
		if err != nil {
			return nil, false, err
		}
		if action == nil {
			return nil, false, nil
		}
		return action, true, nil

	case ChoiceSummonBeast:
		ctx := hp.ctxBuilder.UnitCtx(snapshot)
		action, err := ShowSummonBeastMenu(hp.ir, ctx)
		if err != nil {
			return nil, false, err
		}
		if action == nil {
			return nil, false, nil
		}
		return action, true, nil

	case ChoiceUpgradeRoom:
		ctx := hp.ctxBuilder.UnitCtx(snapshot)
		action, err := ShowUpgradeRoomMenu(hp.ir, ctx)
		if err != nil {
			return nil, false, err
		}
		if action == nil {
			return nil, false, nil
		}
		return action, true, nil

	case ChoiceDoNothing:
		return nil, true, nil

	case ChoiceFastForward:
		n, err := ReadFastForwardTicks(hp.ir)
		if err != nil {
			return nil, false, err
		}
		// Store snapshot at FF start for summary.
		snap := snapshot
		hp.ffStartSnapshot = &snap
		hp.fastForward = n - 1 // -1 because this tick also counts
		return nil, true, nil

	case ChoiceSave, ChoiceLoad, ChoiceReplay:
		// Save/Load/Replay are handled by CLI integration (Task 2-F).
		PrintMessage(hp.out, "この機能はまだ実装されていません。")
		return nil, false, nil

	case ChoiceQuit:
		confirmed, err := ConfirmQuit(hp.ir)
		if err != nil {
			return nil, false, err
		}
		if confirmed {
			return nil, false, io.EOF
		}
		return nil, false, nil

	default:
		return nil, false, nil
	}
}

// OnTickComplete implements server.ActionProvider. It displays a summary
// of changes since the previous tick. During fast-forward, display is
// skipped; a summary is shown when fast-forward completes.
func (hp *HumanProvider) OnTickComplete(snapshot scenario.GameSnapshot) {
	if hp.fastForward > 0 {
		// Still fast-forwarding; skip display.
		snap := snapshot
		hp.prevSnapshot = &snap
		return
	}

	// If we just finished fast-forwarding, show a summary.
	if hp.ffStartSnapshot != nil {
		FormatFastForwardSummary(hp.out, *hp.ffStartSnapshot, snapshot)
		hp.ffStartSnapshot = nil
		snap := snapshot
		hp.prevSnapshot = &snap
		return
	}

	FormatTickSummary(hp.out, hp.prevSnapshot, snapshot)
	snap := snapshot
	hp.prevSnapshot = &snap
}

// OnGameEnd implements server.ActionProvider. It displays the final
// game result and statistics.
func (hp *HumanProvider) OnGameEnd(result simulation.RunResult) {
	FormatGameEnd(hp.out, result)
}
