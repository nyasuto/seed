package balance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyParameter_UpdatesScenarioJSON(t *testing.T) {
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

	backupPath, err := ApplyParameter(scenarioPath, "initial_state.starting_chi", "150")
	if err != nil {
		t.Fatalf("ApplyParameter failed: %v", err)
	}

	// Verify backup was created.
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("backup file was not created")
	}

	// Verify backup contains original data.
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	var backupMap map[string]any
	if err := json.Unmarshal(backupData, &backupMap); err != nil {
		t.Fatalf("parsing backup: %v", err)
	}
	initialState := backupMap["initial_state"].(map[string]any)
	if got := initialState["starting_chi"].(float64); got != 200.0 {
		t.Errorf("backup starting_chi = %v, want 200.0", got)
	}

	// Verify scenario was updated.
	updatedData, err := os.ReadFile(scenarioPath)
	if err != nil {
		t.Fatalf("reading updated scenario: %v", err)
	}
	var updatedMap map[string]any
	if err := json.Unmarshal(updatedData, &updatedMap); err != nil {
		t.Fatalf("parsing updated scenario: %v", err)
	}
	updatedInitial := updatedMap["initial_state"].(map[string]any)
	if got := updatedInitial["starting_chi"].(float64); got != 150.0 {
		t.Errorf("updated starting_chi = %v, want 150.0", got)
	}
}

func TestApplyParameter_BackupFileCreated(t *testing.T) {
	dir := t.TempDir()
	scenarioPath := filepath.Join(dir, "test.json")

	data := []byte(`{"id":"test","value":10}`)
	if err := os.WriteFile(scenarioPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	backupPath, err := ApplyParameter(scenarioPath, "value", "20")
	if err != nil {
		t.Fatalf("ApplyParameter failed: %v", err)
	}

	expectedBackup := scenarioPath + ".bak"
	if backupPath != expectedBackup {
		t.Errorf("backup path = %q, want %q", backupPath, expectedBackup)
	}

	if _, err := os.Stat(expectedBackup); os.IsNotExist(err) {
		t.Fatal("backup file does not exist")
	}
}

func TestApplyParameter_NonexistentFile(t *testing.T) {
	_, err := ApplyParameter("/nonexistent/path.json", "key", "value")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestApplyParameter_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	scenarioPath := filepath.Join(dir, "scenario.json")

	data := []byte(`{"nested":{"value":10}}`)
	if err := os.WriteFile(scenarioPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Try to set a sub-key on a non-object value.
	_, err := ApplyParameter(scenarioPath, "nested.value.subkey", "20")
	if err == nil {
		t.Fatal("expected error for invalid key path")
	}
}

func TestSetJSONPath_FloatValue(t *testing.T) {
	data := []byte(`{"a":{"b":1.0}}`)
	result, err := setJSONPath(data, "a.b", "2.5")
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]any
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatal(err)
	}
	a := m["a"].(map[string]any)
	if got := a["b"].(float64); got != 2.5 {
		t.Errorf("a.b = %v, want 2.5", got)
	}
}

func TestSetJSONPath_BoolValue(t *testing.T) {
	data := []byte(`{"flag":false}`)
	result, err := setJSONPath(data, "flag", "true")
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]any
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatal(err)
	}
	if got := m["flag"].(bool); !got {
		t.Errorf("flag = %v, want true", got)
	}
}
