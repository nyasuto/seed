package human

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReadLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "normal input", input: "hello\n", want: "hello"},
		{name: "trimmed spaces", input: "  hello  \n", want: "hello"},
		{name: "empty input", input: "\n", want: ""},
		{name: "EOF", input: "", wantErr: io.EOF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := NewInputReader(strings.NewReader(tt.input), io.Discard)
			got, err := ir.ReadLine("")
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ReadLine() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadLine() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ReadLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadInt_Valid(t *testing.T) {
	ir := NewInputReader(strings.NewReader("42\n"), io.Discard)
	got, err := ir.ReadInt("")
	if err != nil {
		t.Fatalf("ReadInt() unexpected error: %v", err)
	}
	if got != 42 {
		t.Errorf("ReadInt() = %d, want 42", got)
	}
}

func TestReadInt_InvalidThenValid(t *testing.T) {
	var buf bytes.Buffer
	ir := NewInputReader(strings.NewReader("abc\n7\n"), &buf)
	got, err := ir.ReadInt("")
	if err != nil {
		t.Fatalf("ReadInt() unexpected error: %v", err)
	}
	if got != 7 {
		t.Errorf("ReadInt() = %d, want 7", got)
	}
	if !strings.Contains(buf.String(), "無効な入力です") {
		t.Errorf("expected error message in output, got: %q", buf.String())
	}
}

func TestReadInt_EOF(t *testing.T) {
	ir := NewInputReader(strings.NewReader("abc\n"), io.Discard)
	_, err := ir.ReadInt("")
	if err != io.EOF {
		t.Errorf("ReadInt() error = %v, want io.EOF", err)
	}
}

func TestReadIntInRange_Valid(t *testing.T) {
	ir := NewInputReader(strings.NewReader("3\n"), io.Discard)
	got, err := ir.ReadIntInRange("", 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

func TestReadIntInRange_OutOfRangeThenValid(t *testing.T) {
	var buf bytes.Buffer
	ir := NewInputReader(strings.NewReader("10\n0\n3\n"), &buf)
	got, err := ir.ReadIntInRange("", 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Errorf("got %d, want 3", got)
	}
	if !strings.Contains(buf.String(), "範囲外です") {
		t.Errorf("expected range error in output, got: %q", buf.String())
	}
}

func TestReadYesNo(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "y", input: "y\n", want: true},
		{name: "yes", input: "yes\n", want: true},
		{name: "Y", input: "Y\n", want: true},
		{name: "n", input: "n\n", want: false},
		{name: "no", input: "no\n", want: false},
		{name: "NO", input: "NO\n", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := NewInputReader(strings.NewReader(tt.input), io.Discard)
			got, err := ir.ReadYesNo("")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ReadYesNo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadYesNo_InvalidThenValid(t *testing.T) {
	var buf bytes.Buffer
	ir := NewInputReader(strings.NewReader("maybe\ny\n"), &buf)
	got, err := ir.ReadYesNo("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Errorf("ReadYesNo() = false, want true")
	}
	if !strings.Contains(buf.String(), "y/n") {
		t.Errorf("expected y/n error in output, got: %q", buf.String())
	}
}

func TestReadCoord(t *testing.T) {
	ir := NewInputReader(strings.NewReader("5\n10\n"), io.Discard)
	x, y, err := ir.ReadCoord("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if x != 5 || y != 10 {
		t.Errorf("ReadCoord() = (%d, %d), want (5, 10)", x, y)
	}
}
