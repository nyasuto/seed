package batch

import (
	"bytes"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/server"
)

func loadTutorialScenario(t *testing.T) *scenario.Scenario {
	t.Helper()
	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario(tutorial): %v", err)
	}
	return sc
}

func TestBatchRunner_100Games(t *testing.T) {
	sc := loadTutorialScenario(t)

	runner, err := NewBatchRunner(BatchConfig{
		Scenario: sc,
		Games:    100,
		BaseSeed: 42,
		AI:       AINoop,
		Parallel: 4,
	})
	if err != nil {
		t.Fatalf("NewBatchRunner: %v", err)
	}

	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Summaries) != 100 {
		t.Errorf("Summaries count = %d, want 100", len(result.Summaries))
	}

	// Every game should have completed (Won or Lost).
	for i, s := range result.Summaries {
		if s.Result != simulation.Won && s.Result != simulation.Lost {
			t.Errorf("game %d: unexpected result %v", i, s.Result)
		}
		if s.TotalTicks == 0 {
			t.Errorf("game %d: TotalTicks = 0", i)
		}
	}

	// BreakageData should be populated for all games.
	if len(result.BreakageData) != 100 {
		t.Errorf("BreakageData count = %d, want 100", len(result.BreakageData))
	}
}

func TestBatchRunner_Determinism(t *testing.T) {
	sc := loadTutorialScenario(t)

	run := func() *BatchResult {
		runner, err := NewBatchRunner(BatchConfig{
			Scenario: sc,
			Games:    20,
			BaseSeed: 123,
			AI:       AINoop,
			Parallel: 4,
		})
		if err != nil {
			t.Fatalf("NewBatchRunner: %v", err)
		}
		result, err := runner.Run()
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		return result
	}

	r1 := run()
	r2 := run()

	if len(r1.Summaries) != len(r2.Summaries) {
		t.Fatalf("summary count mismatch: %d vs %d", len(r1.Summaries), len(r2.Summaries))
	}

	for i := range r1.Summaries {
		s1, s2 := r1.Summaries[i], r2.Summaries[i]
		if s1.Result != s2.Result {
			t.Errorf("game %d: result mismatch %v vs %v", i, s1.Result, s2.Result)
		}
		if s1.TotalTicks != s2.TotalTicks {
			t.Errorf("game %d: tick mismatch %d vs %d", i, s1.TotalTicks, s2.TotalTicks)
		}
		if s1.FinalCoreHP != s2.FinalCoreHP {
			t.Errorf("game %d: coreHP mismatch %d vs %d", i, s1.FinalCoreHP, s2.FinalCoreHP)
		}
	}
}

func TestBatchRunner_Progress(t *testing.T) {
	sc := loadTutorialScenario(t)

	var buf bytes.Buffer
	runner, err := NewBatchRunner(BatchConfig{
		Scenario: sc,
		Games:    10,
		BaseSeed: 42,
		AI:       AINoop,
		Parallel: 2,
		Progress: &buf,
	})
	if err != nil {
		t.Fatalf("NewBatchRunner: %v", err)
	}

	_, err = runner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected progress output, got empty")
	}
	// Should contain the final completion message.
	if !bytes.Contains([]byte(output), []byte("10/10 games completed...")) {
		t.Errorf("progress output missing final message: %q", output)
	}
}

func TestBatchRunner_GameSummaryAggregation(t *testing.T) {
	sc := loadTutorialScenario(t)

	runner, err := NewBatchRunner(BatchConfig{
		Scenario: sc,
		Games:    10,
		BaseSeed: 42,
		AI:       AINoop,
		Parallel: 1,
	})
	if err != nil {
		t.Fatalf("NewBatchRunner: %v", err)
	}

	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// All summaries should have valid data.
	for i, s := range result.Summaries {
		if s.Reason == "" {
			t.Errorf("game %d: Reason is empty", i)
		}
	}

	// BreakageReport should have alerts + clean covering all 11 metrics.
	total := len(result.BreakageReport.Alerts) + len(result.BreakageReport.Clean)
	if total != 11 {
		t.Errorf("BreakageReport covers %d metrics, want 11", total)
	}
}

func TestNewBatchRunner_Validation(t *testing.T) {
	sc := loadTutorialScenario(t)

	tests := []struct {
		name    string
		config  BatchConfig
		wantErr bool
	}{
		{
			name:    "nil scenario",
			config:  BatchConfig{Scenario: nil, Games: 10},
			wantErr: true,
		},
		{
			name:    "zero games",
			config:  BatchConfig{Scenario: sc, Games: 0},
			wantErr: true,
		},
		{
			name:    "negative games",
			config:  BatchConfig{Scenario: sc, Games: -1},
			wantErr: true,
		},
		{
			name:    "negative parallel",
			config:  BatchConfig{Scenario: sc, Games: 10, Parallel: -1},
			wantErr: true,
		},
		{
			name:    "valid config",
			config:  BatchConfig{Scenario: sc, Games: 10, BaseSeed: 1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBatchRunner(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBatchRunner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
