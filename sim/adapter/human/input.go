package human

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// InputReader provides helper methods for reading and validating user input
// from an io.Reader. It supports scripted input via io.Reader for testing.
type InputReader struct {
	scanner *bufio.Scanner
	out     io.Writer
}

// NewInputReader creates an InputReader that reads from r and writes prompts to w.
func NewInputReader(r io.Reader, w io.Writer) *InputReader {
	return &InputReader{
		scanner: bufio.NewScanner(r),
		out:     w,
	}
}

// ReadLine reads a single line of input. Returns io.EOF when input is exhausted.
func (ir *InputReader) ReadLine(prompt string) (string, error) {
	fmt.Fprint(ir.out, prompt)
	if !ir.scanner.Scan() {
		if err := ir.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(ir.scanner.Text()), nil
}

// ReadInt reads a line and parses it as an integer. On invalid input, it
// prints an error message and retries until a valid integer is entered or
// input is exhausted.
func (ir *InputReader) ReadInt(prompt string) (int, error) {
	for {
		line, err := ir.ReadLine(prompt)
		if err != nil {
			return 0, err
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			fmt.Fprintf(ir.out, "無効な入力です。数字を入力してください: %q\n", line)
			continue
		}
		return n, nil
	}
}

// ReadIntInRange reads an integer that must be in [min, max]. On out-of-range
// input, it prints an error message and retries.
func (ir *InputReader) ReadIntInRange(prompt string, min, max int) (int, error) {
	for {
		n, err := ir.ReadInt(prompt)
		if err != nil {
			return 0, err
		}
		if n < min || n > max {
			fmt.Fprintf(ir.out, "範囲外です。%d〜%d の数字を入力してください。\n", min, max)
			continue
		}
		return n, nil
	}
}

// ReadYesNo reads a yes/no response. Accepts "y", "yes", "n", "no" (case-insensitive).
// Returns true for yes, false for no.
func (ir *InputReader) ReadYesNo(prompt string) (bool, error) {
	for {
		line, err := ir.ReadLine(prompt)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(line) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintf(ir.out, "無効な入力です。y/n で答えてください。\n")
		}
	}
}

// ReadCoord reads X,Y coordinates (two integers on separate prompts).
func (ir *InputReader) ReadCoord(promptX, promptY string) (x, y int, err error) {
	x, err = ir.ReadInt(promptX)
	if err != nil {
		return 0, 0, err
	}
	y, err = ir.ReadInt(promptY)
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}
