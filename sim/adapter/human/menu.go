package human

import (
	"fmt"
	"io"
)

// MenuChoice represents the player's selection from the main menu.
type MenuChoice int

const (
	// ChoiceDigRoom opens the room digging submenu.
	ChoiceDigRoom MenuChoice = iota + 1
	// ChoiceDigCorridor opens the corridor digging submenu.
	ChoiceDigCorridor
	// ChoiceSummonBeast opens the beast summoning submenu.
	ChoiceSummonBeast
	// ChoiceUpgradeRoom opens the room upgrade submenu.
	ChoiceUpgradeRoom
	// ChoiceDoNothing advances one tick with no action.
	ChoiceDoNothing
	// ChoiceFastForward advances multiple ticks.
	ChoiceFastForward
)

const (
	// ChoiceSave triggers a checkpoint save.
	ChoiceSave MenuChoice = 100
	// ChoiceLoad triggers a checkpoint load.
	ChoiceLoad MenuChoice = 101
	// ChoiceReplay triggers a replay save.
	ChoiceReplay MenuChoice = 102
	// ChoiceQuit exits the game.
	ChoiceQuit MenuChoice = 103
)

// menuText is the formatted main menu string.
const menuText = `
=== メインメニュー ===
  1. 部屋を掘る
  2. 通路を掘る
  3. 仙獣を召喚する
  4. 部屋をアップグレードする
  5. 何もしない（1ティック進める）
  6. 早送り（Nティック）
  ---
  s. セーブ  l. ロード  r. リプレイ保存  q. 終了
`

// ShowMainMenu displays the main menu and reads the player's choice.
// It returns the selected MenuChoice. On invalid input, it re-prompts.
func ShowMainMenu(ir *InputReader) (MenuChoice, error) {
	for {
		fmt.Fprint(ir.out, menuText)
		line, err := ir.ReadLine("選択> ")
		if err != nil {
			return 0, err
		}

		choice, ok := parseMenuChoice(line)
		if !ok {
			fmt.Fprintf(ir.out, "無効な選択です: %q\n", line)
			continue
		}
		return choice, nil
	}
}

// parseMenuChoice converts user input to a MenuChoice.
func parseMenuChoice(input string) (MenuChoice, bool) {
	switch input {
	case "1":
		return ChoiceDigRoom, true
	case "2":
		return ChoiceDigCorridor, true
	case "3":
		return ChoiceSummonBeast, true
	case "4":
		return ChoiceUpgradeRoom, true
	case "5":
		return ChoiceDoNothing, true
	case "6":
		return ChoiceFastForward, true
	case "s", "S":
		return ChoiceSave, true
	case "l", "L":
		return ChoiceLoad, true
	case "r", "R":
		return ChoiceReplay, true
	case "q", "Q":
		return ChoiceQuit, true
	default:
		return 0, false
	}
}

// ReadFastForwardTicks prompts the player for the number of ticks to fast-forward.
func ReadFastForwardTicks(ir *InputReader) (int, error) {
	n, err := ir.ReadIntInRange("何ティック進めますか？ (1-1000)> ", 1, 1000)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ConfirmQuit asks the player to confirm quitting.
func ConfirmQuit(ir *InputReader) (bool, error) {
	return ir.ReadYesNo("本当に終了しますか？ (y/n)> ")
}

// PrintMessage writes a message line to the output.
func PrintMessage(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}

// PrintError writes an error message to the output.
func PrintError(w io.Writer, err error) {
	fmt.Fprintf(w, "エラー: %v\n", err)
}
