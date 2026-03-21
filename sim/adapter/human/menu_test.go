package human

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestShowMainMenu_ValidChoices(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  MenuChoice
	}{
		{name: "dig room", input: "1\n", want: ChoiceDigRoom},
		{name: "dig corridor", input: "2\n", want: ChoiceDigCorridor},
		{name: "summon beast", input: "3\n", want: ChoiceSummonBeast},
		{name: "upgrade room", input: "4\n", want: ChoiceUpgradeRoom},
		{name: "do nothing", input: "5\n", want: ChoiceDoNothing},
		{name: "fast forward", input: "6\n", want: ChoiceFastForward},
		{name: "save", input: "s\n", want: ChoiceSave},
		{name: "save upper", input: "S\n", want: ChoiceSave},
		{name: "load", input: "l\n", want: ChoiceLoad},
		{name: "replay", input: "r\n", want: ChoiceReplay},
		{name: "quit", input: "q\n", want: ChoiceQuit},
		{name: "quit upper", input: "Q\n", want: ChoiceQuit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := NewInputReader(strings.NewReader(tt.input), io.Discard)
			got, err := ShowMainMenu(ir)
			if err != nil {
				t.Fatalf("ShowMainMenu() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ShowMainMenu() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestShowMainMenu_InvalidThenValid(t *testing.T) {
	var buf bytes.Buffer
	// "x" is invalid, then "7" is invalid, then "1" is valid
	ir := NewInputReader(strings.NewReader("x\n7\n1\n"), &buf)
	got, err := ShowMainMenu(ir)
	if err != nil {
		t.Fatalf("ShowMainMenu() unexpected error: %v", err)
	}
	if got != ChoiceDigRoom {
		t.Errorf("ShowMainMenu() = %d, want %d", got, ChoiceDigRoom)
	}
	output := buf.String()
	// Should show error messages for both invalid inputs
	if strings.Count(output, "無効な選択です") != 2 {
		t.Errorf("expected 2 error messages, got output: %q", output)
	}
}

func TestShowMainMenu_EOF(t *testing.T) {
	ir := NewInputReader(strings.NewReader(""), io.Discard)
	_, err := ShowMainMenu(ir)
	if err != io.EOF {
		t.Errorf("ShowMainMenu() error = %v, want io.EOF", err)
	}
}

func TestShowMainMenu_DisplaysMenuText(t *testing.T) {
	var buf bytes.Buffer
	ir := NewInputReader(strings.NewReader("5\n"), &buf)
	_, err := ShowMainMenu(ir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// Verify key menu items are displayed
	for _, item := range []string{"部屋を掘る", "通路を掘る", "仙獣を召喚する", "アップグレード", "何もしない", "早送り", "セーブ", "ロード", "リプレイ", "終了"} {
		if !strings.Contains(output, item) {
			t.Errorf("menu output missing %q", item)
		}
	}
}

func TestReadFastForwardTicks(t *testing.T) {
	ir := NewInputReader(strings.NewReader("50\n"), io.Discard)
	got, err := ReadFastForwardTicks(ir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 50 {
		t.Errorf("ReadFastForwardTicks() = %d, want 50", got)
	}
}

func TestReadFastForwardTicks_OutOfRange(t *testing.T) {
	var buf bytes.Buffer
	ir := NewInputReader(strings.NewReader("0\n1001\n100\n"), &buf)
	got, err := ReadFastForwardTicks(ir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 100 {
		t.Errorf("ReadFastForwardTicks() = %d, want 100", got)
	}
}

func TestConfirmQuit_Yes(t *testing.T) {
	ir := NewInputReader(strings.NewReader("y\n"), io.Discard)
	got, err := ConfirmQuit(ir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("ConfirmQuit() = false, want true")
	}
}

func TestConfirmQuit_No(t *testing.T) {
	ir := NewInputReader(strings.NewReader("n\n"), io.Discard)
	got, err := ConfirmQuit(ir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("ConfirmQuit() = true, want false")
	}
}

func TestMenuTransition_MultipleSelections(t *testing.T) {
	// Simulate a session: select "5" (do nothing), then re-show menu and select "q" (quit)
	input := "5\nq\n"
	ir := NewInputReader(strings.NewReader(input), io.Discard)

	first, err := ShowMainMenu(ir)
	if err != nil {
		t.Fatalf("first ShowMainMenu() error: %v", err)
	}
	if first != ChoiceDoNothing {
		t.Errorf("first choice = %d, want %d", first, ChoiceDoNothing)
	}

	second, err := ShowMainMenu(ir)
	if err != nil {
		t.Fatalf("second ShowMainMenu() error: %v", err)
	}
	if second != ChoiceQuit {
		t.Errorf("second choice = %d, want %d", second, ChoiceQuit)
	}
}
