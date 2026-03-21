package scenario

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewCommand_SpawnWave(t *testing.T) {
	def := CommandDef{
		Type:   "spawn_wave",
		Params: json.RawMessage(`{"difficulty": 1.5, "min_invaders": 3, "max_invaders": 6}`),
	}
	cmd, err := NewCommand(def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sw, ok := cmd.(*SpawnWaveCommand)
	if !ok {
		t.Fatalf("expected *SpawnWaveCommand, got %T", cmd)
	}
	if sw.Difficulty != 1.5 {
		t.Errorf("Difficulty = %v, want 1.5", sw.Difficulty)
	}
	if sw.MinInvaders != 3 {
		t.Errorf("MinInvaders = %d, want 3", sw.MinInvaders)
	}
	if sw.MaxInvaders != 6 {
		t.Errorf("MaxInvaders = %d, want 6", sw.MaxInvaders)
	}
	desc := cmd.Execute()
	if desc != "spawn wave: difficulty=1.5 invaders=3-6" {
		t.Errorf("Execute() = %q", desc)
	}
}

func TestNewCommand_SpawnWave_MinExceedsMax(t *testing.T) {
	def := CommandDef{
		Type:   "spawn_wave",
		Params: json.RawMessage(`{"difficulty": 1.0, "min_invaders": 10, "max_invaders": 5}`),
	}
	_, err := NewCommand(def)
	if err == nil {
		t.Fatal("expected error for min > max")
	}
}

func TestNewCommand_ModifyChi(t *testing.T) {
	def := CommandDef{
		Type:   "modify_chi",
		Params: json.RawMessage(`{"amount": -50.0}`),
	}
	cmd, err := NewCommand(def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mc, ok := cmd.(*ModifyChiCommand)
	if !ok {
		t.Fatalf("expected *ModifyChiCommand, got %T", cmd)
	}
	if mc.Amount != -50.0 {
		t.Errorf("Amount = %v, want -50", mc.Amount)
	}
	desc := cmd.Execute()
	if desc != "modify chi: -50.0" {
		t.Errorf("Execute() = %q", desc)
	}
}

func TestNewCommand_ModifyConstraint(t *testing.T) {
	def := CommandDef{
		Type:   "modify_constraint",
		Params: json.RawMessage(`{"constraint": "max_rooms", "value": 10.0}`),
	}
	cmd, err := NewCommand(def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mc, ok := cmd.(*ModifyConstraintCommand)
	if !ok {
		t.Fatalf("expected *ModifyConstraintCommand, got %T", cmd)
	}
	if mc.Constraint != "max_rooms" {
		t.Errorf("Constraint = %q, want %q", mc.Constraint, "max_rooms")
	}
	if mc.Value != 10.0 {
		t.Errorf("Value = %v, want 10", mc.Value)
	}
	desc := cmd.Execute()
	if desc != "modify constraint: max_rooms=10.0" {
		t.Errorf("Execute() = %q", desc)
	}
}

func TestNewCommand_Message(t *testing.T) {
	def := CommandDef{
		Type:   "message",
		Params: json.RawMessage(`{"text": "boss incoming!"}`),
	}
	cmd, err := NewCommand(def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mc, ok := cmd.(*MessageCommand)
	if !ok {
		t.Fatalf("expected *MessageCommand, got %T", cmd)
	}
	if mc.Text != "boss incoming!" {
		t.Errorf("Text = %q", mc.Text)
	}
	if cmd.Execute() != "boss incoming!" {
		t.Errorf("Execute() = %q", cmd.Execute())
	}
}

func TestNewCommand_UnknownType(t *testing.T) {
	def := CommandDef{
		Type:   "explode",
		Params: json.RawMessage(`{}`),
	}
	_, err := NewCommand(def)
	if !errors.Is(err, ErrUnknownCommandType) {
		t.Errorf("expected ErrUnknownCommandType, got %v", err)
	}
}

func TestNewCommand_MissingParams(t *testing.T) {
	tests := []struct {
		name string
		def  CommandDef
	}{
		{"spawn_wave missing difficulty", CommandDef{Type: "spawn_wave", Params: json.RawMessage(`{"min_invaders": 1, "max_invaders": 3}`)}},
		{"modify_chi missing amount", CommandDef{Type: "modify_chi", Params: json.RawMessage(`{}`)}},
		{"modify_constraint missing constraint", CommandDef{Type: "modify_constraint", Params: json.RawMessage(`{"value": 1.0}`)}},
		{"modify_constraint missing value", CommandDef{Type: "modify_constraint", Params: json.RawMessage(`{"constraint": "x"}`)}},
		{"message missing text", CommandDef{Type: "message", Params: json.RawMessage(`{}`)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCommand(tt.def)
			if err == nil {
				t.Error("expected error for missing params")
			}
		})
	}
}
