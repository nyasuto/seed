package sim

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/adapter/ai"
	"github.com/nyasuto/seed/sim/adapter/batch"
	"github.com/nyasuto/seed/sim/adapter/human"
	"github.com/nyasuto/seed/sim/metrics"
	"github.com/nyasuto/seed/sim/server"
)

// TestPhase4_AllModesCoexistence verifies that Human, AI, and Batch modes
// all produce valid terminal results for the same scenario.
func TestPhase4_AllModesCoexistence(t *testing.T) {
	const seed int64 = 42

	t.Run("HumanMode", func(t *testing.T) {
		sc, err := server.LoadBuiltinScenario("tutorial")
		if err != nil {
			t.Fatalf("LoadBuiltinScenario: %v", err)
		}
		gs, err := server.NewGameServer(sc, seed)
		if err != nil {
			t.Fatalf("NewGameServer: %v", err)
		}

		// Script: wait (5) → fast-forward (6) → 300 ticks.
		input := "5\n6\n300\n"
		out := &bytes.Buffer{}
		ir := human.NewInputReader(strings.NewReader(input), out)
		ctxBuilder := server.NewGameContextBuilder(gs)
		provider := human.NewHumanProvider(ir, out, ctxBuilder)
		provider.SetCheckpointOps(server.NewServerCheckpointOps(gs))

		result, err := gs.RunGame(provider)
		if err != nil {
			t.Fatalf("RunGame Human: %v", err)
		}
		assertTerminalResult(t, "Human", result)
	})

	t.Run("AIMode", func(t *testing.T) {
		sc, err := server.LoadBuiltinScenario("tutorial")
		if err != nil {
			t.Fatalf("LoadBuiltinScenario: %v", err)
		}
		gs, err := server.NewGameServer(sc, seed)
		if err != nil {
			t.Fatalf("NewGameServer: %v", err)
		}

		inR, inW := io.Pipe()
		outR, outW := io.Pipe()

		builder := ai.NewStateBuilder(gs.Engine)
		provider := ai.NewAIProvider(inR, outW, builder)

		clientErr := make(chan error, 1)
		go func() {
			defer func() { _ = inW.Close() }()
			client := newTestAIClient(outR, inW)
			for {
				msg, err := client.readMessage()
				if err != nil {
					clientErr <- err
					return
				}
				switch msg["type"] {
				case "state":
					if err := client.waitAction(); err != nil {
						clientErr <- err
						return
					}
				case "game_end":
					clientErr <- nil
					return
				}
			}
		}()

		result, err := gs.RunGame(provider)
		if err != nil {
			t.Fatalf("RunGame AI: %v", err)
		}
		if cErr := <-clientErr; cErr != nil {
			t.Fatalf("client error: %v", cErr)
		}
		assertTerminalResult(t, "AI", result)
	})

	t.Run("BatchMode", func(t *testing.T) {
		sc, err := server.LoadBuiltinScenario("tutorial")
		if err != nil {
			t.Fatalf("LoadBuiltinScenario: %v", err)
		}

		runner, err := batch.NewBatchRunner(batch.BatchConfig{
			Scenario: sc,
			Games:    10,
			BaseSeed: seed,
			AI:       batch.AISimple,
			Parallel: 2,
		})
		if err != nil {
			t.Fatalf("NewBatchRunner: %v", err)
		}

		batchResult, err := runner.Run()
		if err != nil {
			t.Fatalf("Run: %v", err)
		}

		if len(batchResult.Summaries) != 10 {
			t.Fatalf("expected 10 summaries, got %d", len(batchResult.Summaries))
		}

		for i, s := range batchResult.Summaries {
			if s.Result != simulation.Won && s.Result != simulation.Lost {
				t.Errorf("game %d: unexpected result %v", i, s.Result)
			}
		}
		t.Logf("Batch: %d games, %d alerts, %d clean metrics",
			len(batchResult.Summaries),
			len(batchResult.BreakageReport.Alerts),
			len(batchResult.BreakageReport.Clean))
	})
}

// TestPhase4_BatchFullPipeline verifies the full Batch Mode pipeline:
// batch run → metrics collection → breakage detection → report generation.
func TestPhase4_BatchFullPipeline(t *testing.T) {
	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	// 1. Batch run.
	runner, err := batch.NewBatchRunner(batch.BatchConfig{
		Scenario: sc,
		Games:    20,
		BaseSeed: 100,
		AI:       batch.AISimple,
		Parallel: 2,
	})
	if err != nil {
		t.Fatalf("NewBatchRunner: %v", err)
	}

	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// 2. Verify summaries are populated.
	if len(result.Summaries) != 20 {
		t.Fatalf("expected 20 summaries, got %d", len(result.Summaries))
	}
	for i, s := range result.Summaries {
		if s.TotalTicks == 0 {
			t.Errorf("game %d: TotalTicks should be > 0", i)
		}
	}

	// 3. Verify breakage report structure.
	report := result.BreakageReport
	totalMetrics := len(report.Alerts) + len(report.Clean)
	if totalMetrics == 0 {
		t.Error("breakage report has no alerts and no clean metrics")
	}

	// Critical metrics (B04, B06, B07, B08) should not alert on tutorial.
	criticalIDs := map[string]bool{"B04": true, "B06": true, "B07": true, "B08": true}
	for _, a := range report.Alerts {
		if criticalIDs[a.MetricID] {
			t.Errorf("critical alert %s should not appear on tutorial: %s (value=%.4f)",
				a.MetricID, a.BrokenSign, a.Value)
		}
	}

	// 4. Verify breakage data is populated.
	if len(result.BreakageData) != 20 {
		t.Fatalf("expected 20 breakage data entries, got %d", len(result.BreakageData))
	}

	// 5. Verify JSON report generation.
	jsonStr, err := metrics.GenerateJSON(
		metrics.ReportConfig{Scenario: "tutorial", Games: 20, AI: "simple"},
		result.Summaries, result.BreakageData, report,
	)
	if err != nil {
		t.Fatalf("GenerateJSON: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("invalid JSON report: %v", err)
	}
	for _, key := range []string{"config", "summary", "breakage_report", "raw_metrics"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("JSON report missing %q section", key)
		}
	}

	// 6. Verify CSV report generation.
	csv := metrics.GenerateCSV(result.Summaries)
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	// Header + 20 data lines.
	if len(lines) != 21 {
		t.Errorf("CSV: expected 21 lines (header + 20 games), got %d", len(lines))
	}

	t.Logf("Pipeline: %d games, JSON=%d bytes, CSV=%d lines, alerts=%d, clean=%d",
		len(result.Summaries), len(jsonStr), len(lines),
		len(report.Alerts), len(report.Clean))
}

// TestPhase4_BatchDeterminism verifies that running the same batch config
// twice produces identical results.
func TestPhase4_BatchDeterminism(t *testing.T) {
	sc, err := server.LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	runBatch := func() *batch.BatchResult {
		runner, err := batch.NewBatchRunner(batch.BatchConfig{
			Scenario: sc,
			Games:    10,
			BaseSeed: 777,
			AI:       batch.AISimple,
			Parallel: 1, // single-threaded for determinism
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

	r1 := runBatch()
	r2 := runBatch()

	for i := range r1.Summaries {
		s1, s2 := r1.Summaries[i], r2.Summaries[i]
		if s1.TotalTicks != s2.TotalTicks {
			t.Errorf("game %d: TotalTicks mismatch: %d vs %d", i, s1.TotalTicks, s2.TotalTicks)
		}
		if s1.Result != s2.Result {
			t.Errorf("game %d: Result mismatch: %v vs %v", i, s1.Result, s2.Result)
		}
		if s1.RoomsBuilt != s2.RoomsBuilt {
			t.Errorf("game %d: RoomsBuilt mismatch: %d vs %d", i, s1.RoomsBuilt, s2.RoomsBuilt)
		}
	}
}

// TestPhase4_SweepIntegration verifies that parameter sweeping produces
// results for each parameter value via the RunSweep API.
func TestPhase4_SweepIntegration(t *testing.T) {
	scenarioJSON, err := server.LoadBuiltinScenarioJSON("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenarioJSON: %v", err)
	}

	param, err := batch.ParseSweepParam("initial_state.starting_chi=50,200,500")
	if err != nil {
		t.Fatalf("ParseSweepParam: %v", err)
	}

	results, err := batch.RunSweep(scenarioJSON, param, batch.BatchConfig{
		Games:    5,
		BaseSeed: 42,
		AI:       batch.AISimple,
		Parallel: 2,
	})
	if err != nil {
		t.Fatalf("RunSweep: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 sweep results, got %d", len(results))
	}

	for i, r := range results {
		if r.Result == nil {
			t.Errorf("result[%d] is nil", i)
			continue
		}
		if len(r.Result.Summaries) != 5 {
			t.Errorf("result[%d].Summaries count = %d, want 5", i, len(r.Result.Summaries))
		}
		t.Logf("sweep %s=%s: %d alerts, %d clean",
			r.ParamKey, r.ParamValue,
			len(r.Result.BreakageReport.Alerts),
			len(r.Result.BreakageReport.Clean))
	}
}

// --- Test helpers ---

// testAIClient simulates an AI client for testing.
type testAIClient struct {
	reader  io.Reader
	writer  io.Writer
	scanner interface {
		Scan() bool
		Bytes() []byte
		Err() error
	}
}

func newTestAIClient(outR io.Reader, inW io.Writer) *testAIClient {
	scanner := newLineScanner(outR)
	return &testAIClient{reader: outR, writer: inW, scanner: scanner}
}

func (c *testAIClient) readMessage() (map[string]any, error) {
	if !c.scanner.Scan() {
		if err := c.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	var msg map[string]any
	if err := json.Unmarshal(c.scanner.Bytes(), &msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *testAIClient) waitAction() error {
	msg := ai.ActionMessage{
		Type:    "action",
		Actions: []ai.ActionDef{{Kind: "wait", Params: map[string]any{}}},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = c.writer.Write(append(data, '\n'))
	return err
}

type lineScanner struct {
	*bytes.Reader
	buf    []byte
	line   []byte
	reader io.Reader
	err    error
}

func newLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{reader: r, buf: make([]byte, 0, 1024*1024)}
}

func (s *lineScanner) Scan() bool {
	// Read until we find a newline.
	for {
		for i, b := range s.buf {
			if b == '\n' {
				s.line = s.buf[:i]
				s.buf = s.buf[i+1:]
				return true
			}
		}
		tmp := make([]byte, 4096)
		n, err := s.reader.Read(tmp)
		if n > 0 {
			s.buf = append(s.buf, tmp[:n]...)
		}
		if err != nil {
			if len(s.buf) > 0 {
				s.line = s.buf
				s.buf = nil
				return true
			}
			if err != io.EOF {
				s.err = err
			}
			return false
		}
	}
}

func (s *lineScanner) Bytes() []byte { return s.line }
func (s *lineScanner) Err() error    { return s.err }

func assertTerminalResult(t *testing.T, mode string, result simulation.RunResult) {
	t.Helper()
	if result.Result.Status != simulation.Won && result.Result.Status != simulation.Lost {
		t.Errorf("%s: expected terminal status, got %v", mode, result.Result.Status)
	}
	if result.TickCount == 0 {
		t.Errorf("%s: expected TickCount > 0", mode)
	}
	t.Logf("%s: status=%v ticks=%d", mode, result.Result.Status, result.TickCount)
}
