package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBatchMode_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "results.json")

	err := runBatchMode("tutorial", 10, "simple", outPath, "json", "")
	if err != nil {
		t.Fatalf("runBatchMode: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	// Verify it's valid JSON with expected structure.
	var report map[string]any
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check required top-level keys.
	for _, key := range []string{"config", "summary", "breakage_report", "raw_metrics"} {
		if _, ok := report[key]; !ok {
			t.Errorf("missing key %q in JSON report", key)
		}
	}

	// Verify config section.
	config := report["config"].(map[string]any)
	if config["scenario"] != "tutorial" {
		t.Errorf("config.scenario = %v, want tutorial", config["scenario"])
	}
	if config["games"].(float64) != 10 {
		t.Errorf("config.games = %v, want 10", config["games"])
	}
	if config["ai"] != "simple" {
		t.Errorf("config.ai = %v, want simple", config["ai"])
	}
}

func TestRunBatchMode_CSVOutput(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "results.csv")

	err := runBatchMode("tutorial", 5, "noop", outPath, "csv", "")
	if err != nil {
		t.Fatalf("runBatchMode: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// Header + 5 game rows.
	if len(lines) != 6 {
		t.Errorf("CSV lines = %d, want 6 (header + 5 games)", len(lines))
	}

	// Verify header.
	if !strings.HasPrefix(lines[0], "game,result,reason") {
		t.Errorf("unexpected CSV header: %s", lines[0])
	}
}

func TestRunBatchMode_Sweep(t *testing.T) {
	err := runBatchMode("tutorial", 5, "noop", "", "json", "initial_state.starting_chi=100.0,200.0,500.0")
	if err != nil {
		t.Fatalf("runBatchMode with sweep: %v", err)
	}
}

func TestRunBatchMode_InvalidAI(t *testing.T) {
	err := runBatchMode("tutorial", 5, "unknown", "", "json", "")
	if err == nil {
		t.Fatal("expected error for unknown AI strategy")
	}
}

func TestRunBatchMode_InvalidFormat(t *testing.T) {
	err := runBatchMode("tutorial", 5, "noop", "", "xml", "")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}
