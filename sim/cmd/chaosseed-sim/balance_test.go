package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nyasuto/seed/sim/balance"
)

func TestRunBalanceMode_Starts(t *testing.T) {
	// Verify that balance mode runs successfully with a small game count.
	err := runBalanceMode("tutorial", 10)
	if err != nil {
		t.Fatalf("runBalanceMode: %v", err)
	}
}

func TestApplyParameter_UpdatesAndBackups(t *testing.T) {
	dir := t.TempDir()
	scenarioPath := filepath.Join(dir, "scenario.json")

	original := map[string]any{
		"id": "test",
		"initial_state": map[string]any{
			"starting_chi":    200.0,
			"terrain_density": 0.05,
		},
	}
	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scenarioPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Apply parameter change.
	backupPath, err := balance.ApplyParameter(scenarioPath, "initial_state.starting_chi", "150")
	if err != nil {
		t.Fatalf("ApplyParameter: %v", err)
	}

	// Verify backup exists.
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("backup file was not created")
	}

	// Verify scenario was updated.
	updatedData, err := os.ReadFile(scenarioPath)
	if err != nil {
		t.Fatal(err)
	}
	var updated map[string]any
	if err := json.Unmarshal(updatedData, &updated); err != nil {
		t.Fatal(err)
	}
	initialState := updated["initial_state"].(map[string]any)
	if got := initialState["starting_chi"].(float64); got != 150.0 {
		t.Errorf("starting_chi = %v, want 150.0", got)
	}

	// Verify backup has original value.
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	var backup map[string]any
	if err := json.Unmarshal(backupData, &backup); err != nil {
		t.Fatal(err)
	}
	backupInitial := backup["initial_state"].(map[string]any)
	if got := backupInitial["starting_chi"].(float64); got != 200.0 {
		t.Errorf("backup starting_chi = %v, want 200.0", got)
	}
}
