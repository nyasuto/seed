package batch

import (
	"testing"
	"time"

	"github.com/nyasuto/seed/core/simulation"
)

// TestD002Verification runs the tutorial scenario with SimpleAI across 1,000
// games and reports BreakageReport results. The tutorial scenario is designed
// to be easy (low terrain density, generous chi, single wave), so some
// breakage alerts (B03, B05, B11) are expected by design.
//
// This test verifies:
//   - All 1,000 games complete successfully
//   - No critical alerts (B04, B06, B07, B08) appear
//   - Known tutorial-specific alerts are documented
func TestD002Verification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping D002 verification in short mode")
	}

	sc := loadTutorialScenario(t)

	runner, err := NewBatchRunner(BatchConfig{
		Scenario: sc,
		Games:    1000,
		BaseSeed: 42,
		AI:       AISimple,
		Parallel: 0, // use all CPUs
	})
	if err != nil {
		t.Fatalf("NewBatchRunner: %v", err)
	}

	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Summaries) != 1000 {
		t.Fatalf("expected 1000 summaries, got %d", len(result.Summaries))
	}

	// Every game should complete.
	for i, s := range result.Summaries {
		if s.Result != simulation.Won && s.Result != simulation.Lost {
			t.Errorf("game %d: unexpected result %v", i, s.Result)
		}
	}

	// Log all alerts for visibility.
	for _, a := range result.BreakageReport.Alerts {
		t.Logf("alert: %s: %s (value=%.4f, threshold=%.4f, direction=%s)",
			a.MetricID, a.BrokenSign, a.Value, a.Threshold, a.Direction)
	}
	t.Logf("clean metrics: %v", result.BreakageReport.Clean)

	// Known tutorial-specific alerts that are acceptable due to the easy
	// scenario design (low terrain density, single wave, generous chi).
	knownTutorialAlerts := map[string]bool{
		"B03": true, // terrain_density=0.05, intentionally weak constraints
		"B05": true, // single wave at tick 100, no construction overlap
		"B11": true, // starting_chi=200 with low costs, surplus expected
	}

	// Critical alerts that should NOT appear even for tutorial.
	for _, a := range result.BreakageReport.Alerts {
		if !knownTutorialAlerts[a.MetricID] {
			t.Errorf("unexpected alert %s: %s (value=%.4f, threshold=%.4f)",
				a.MetricID, a.BrokenSign, a.Value, a.Threshold)
		}
	}
}

// TestBatchRunner_SimpleAI verifies that SimpleAI actually takes actions
// and produces different results from noop.
func TestBatchRunner_SimpleAI(t *testing.T) {
	sc := loadTutorialScenario(t)

	runWith := func(ai AIType) *BatchResult {
		runner, err := NewBatchRunner(BatchConfig{
			Scenario: sc,
			Games:    10,
			BaseSeed: 42,
			AI:       ai,
			Parallel: 2,
		})
		if err != nil {
			t.Fatalf("NewBatchRunner(%s): %v", ai, err)
		}
		result, err := runner.Run()
		if err != nil {
			t.Fatalf("Run(%s): %v", ai, err)
		}
		return result
	}

	simple := runWith(AISimple)
	noop := runWith(AINoop)

	// SimpleAI should build rooms; noop should not.
	simpleRooms := 0
	noopRooms := 0
	for _, s := range simple.Summaries {
		simpleRooms += s.RoomsBuilt
	}
	for _, s := range noop.Summaries {
		noopRooms += s.RoomsBuilt
	}

	if simpleRooms <= noopRooms {
		t.Errorf("SimpleAI should build more rooms than Noop: simple=%d, noop=%d", simpleRooms, noopRooms)
	}
}

// TestBatchPerformance_1000Games verifies that 1,000 games with SimpleAI
// complete within 5 minutes.
func TestBatchPerformance_1000Games(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	sc := loadTutorialScenario(t)

	runner, err := NewBatchRunner(BatchConfig{
		Scenario: sc,
		Games:    1000,
		BaseSeed: 42,
		AI:       AISimple,
		Parallel: 0,
	})
	if err != nil {
		t.Fatalf("NewBatchRunner: %v", err)
	}

	start := time.Now()
	_, err = runner.Run()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	maxDuration := 5 * time.Minute
	if elapsed > maxDuration {
		t.Errorf("1000 games took %v, exceeds limit of %v", elapsed, maxDuration)
	}
	t.Logf("1000 games completed in %v", elapsed)
}
