// Package main is the entry point for chaosseed-sim CLI.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nyasuto/seed/sim/adapter/ai"
	"github.com/nyasuto/seed/sim/adapter/batch"
	"github.com/nyasuto/seed/sim/adapter/human"
	"github.com/nyasuto/seed/sim/balance"
	"github.com/nyasuto/seed/sim/metrics"
	"github.com/nyasuto/seed/sim/server"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	showVersion := flag.Bool("version", false, "show version and exit")
	humanMode := flag.Bool("human", false, "start Human Mode (interactive terminal)")
	aiMode := flag.Bool("ai", false, "start AI Mode (JSON Lines I/O)")
	batchMode := flag.Bool("batch", false, "start Batch Mode (headless statistics)")
	balanceMode := flag.Bool("balance", false, "start Balance Dashboard")
	scenarioName := flag.String("scenario", "tutorial", "scenario name or file path")
	replayPath := flag.String("replay", "", "replay file path to play back")
	aiTimeout := flag.Duration("timeout", 0, "action input timeout for AI Mode (e.g. 30s)")
	games := flag.Int("games", 100, "number of games for batch mode")
	aiStrategy := flag.String("batch-ai", "noop", "AI strategy for batch mode (simple, noop)")
	outputPath := flag.String("output", "", "output file path for batch results")
	format := flag.String("format", "json", "output format for batch mode (json, csv)")
	sweep := flag.String("sweep", "", "parameter sweep spec (e.g. key=v1,v2,v3)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("chaosseed-sim %s\n", version)
		return 0
	}

	selected := 0
	if *humanMode {
		selected++
	}
	if *aiMode {
		selected++
	}
	if *batchMode {
		selected++
	}
	if *balanceMode {
		selected++
	}
	// --replay is a standalone mode.
	if *replayPath != "" {
		selected++
	}

	if selected == 0 {
		fmt.Fprintln(os.Stderr, "error: specify one of --human, --ai, --batch, --balance, or --replay")
		flag.Usage()
		return 1
	}
	if selected > 1 {
		fmt.Fprintln(os.Stderr, "error: specify only one mode at a time")
		return 1
	}

	switch {
	case *replayPath != "":
		if err := runReplayMode(*scenarioName, *replayPath); err != nil {
			fmt.Fprintf(os.Stderr, "replay error: %v\n", err)
			return 1
		}
		return 0
	case *humanMode:
		if err := runHumanMode(*scenarioName); err != nil {
			fmt.Fprintf(os.Stderr, "human mode error: %v\n", err)
			return 1
		}
		return 0
	case *aiMode:
		if err := runAIMode(*scenarioName, *aiTimeout); err != nil {
			fmt.Fprintf(os.Stderr, "ai mode error: %v\n", err)
			return 1
		}
		return 0
	case *batchMode:
		if err := runBatchMode(*scenarioName, *games, *aiStrategy, *outputPath, *format, *sweep); err != nil {
			fmt.Fprintf(os.Stderr, "batch mode error: %v\n", err)
			return 1
		}
		return 0
	case *balanceMode:
		if err := runBalanceMode(*scenarioName, *games); err != nil {
			fmt.Fprintf(os.Stderr, "balance mode error: %v\n", err)
			return 1
		}
		return 0
	}
	return 1
}

// runHumanMode starts the Human Mode interactive game.
func runHumanMode(scenarioName string) error {
	sc, err := server.LoadScenario(scenarioName)
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	seed := time.Now().UnixNano()
	gs, err := server.NewGameServer(sc, seed)
	if err != nil {
		return fmt.Errorf("create game server: %w", err)
	}

	ir := human.NewInputReader(os.Stdin, os.Stdout)
	ctxBuilder := server.NewGameContextBuilder(gs)
	provider := human.NewHumanProvider(ir, os.Stdout, ctxBuilder)
	provider.SetCheckpointOps(server.NewServerCheckpointOps(gs))

	_, err = gs.RunGame(provider)
	if err == nil {
		return nil
	}

	// Handle checkpoint load: restart via ResumeGame loop.
	for errors.Is(err, human.ErrCheckpointLoaded) {
		_, err = gs.ResumeGame(provider)
	}

	// io.EOF means the player quit.
	if errors.Is(err, io.EOF) {
		_, _ = fmt.Fprintln(os.Stdout, "ゲームを終了しました。")
		return nil
	}

	return err
}

// runAIMode starts the AI Mode game with JSON Lines I/O.
func runAIMode(scenarioName string, timeout time.Duration) error {
	sc, err := server.LoadScenario(scenarioName)
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	seed := time.Now().UnixNano()
	gs, err := server.NewGameServer(sc, seed)
	if err != nil {
		return fmt.Errorf("create game server: %w", err)
	}

	builder := ai.NewStateBuilder(gs.Engine)
	provider := ai.NewAIProvider(os.Stdin, os.Stdout, builder)
	if timeout > 0 {
		provider.SetTimeout(timeout)
	}

	_, err = gs.RunGame(provider)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// runBatchMode executes batch mode with the given parameters.
func runBatchMode(scenarioName string, games int, aiStrategy string, outputPath string, format string, sweepSpec string) error {
	aiType := batch.AIType(aiStrategy)
	switch aiType {
	case batch.AISimple, batch.AINoop:
	default:
		return fmt.Errorf("unknown AI strategy: %q (use simple or noop)", aiStrategy)
	}

	// Sweep mode: run parameter sweep.
	if sweepSpec != "" {
		return runBatchSweep(scenarioName, games, aiType, outputPath, format, sweepSpec)
	}

	sc, err := server.LoadScenario(scenarioName)
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	config := batch.BatchConfig{
		Scenario: sc,
		Games:    games,
		BaseSeed: 42,
		AI:       aiType,
		Progress: os.Stderr,
	}

	runner, err := batch.NewBatchRunner(config)
	if err != nil {
		return fmt.Errorf("create batch runner: %w", err)
	}

	result, err := runner.Run()
	if err != nil {
		return fmt.Errorf("batch run: %w", err)
	}

	return writeBatchOutput(result, scenarioName, games, string(aiType), outputPath, format)
}

// runBatchSweep executes a parameter sweep.
func runBatchSweep(scenarioName string, games int, aiType batch.AIType, outputPath string, format string, sweepSpec string) error {
	param, err := batch.ParseSweepParam(sweepSpec)
	if err != nil {
		return fmt.Errorf("parse sweep: %w", err)
	}

	// Load raw JSON for sweep modification.
	scenarioJSON, err := loadScenarioJSON(scenarioName)
	if err != nil {
		return fmt.Errorf("load scenario JSON: %w", err)
	}

	baseConfig := batch.BatchConfig{
		Games:    games,
		BaseSeed: 42,
		AI:       aiType,
		Progress: os.Stderr,
	}

	results, err := batch.RunSweep(scenarioJSON, param, baseConfig)
	if err != nil {
		return fmt.Errorf("sweep: %w", err)
	}

	// For sweep, output each result with the parameter value.
	for _, sr := range results {
		label := fmt.Sprintf("%s=%s", sr.ParamKey, sr.ParamValue)
		fmt.Fprintf(os.Stderr, "\n--- %s ---\n", label)
		if err := writeBatchOutput(sr.Result, scenarioName, games, string(aiType), "", format); err != nil {
			return fmt.Errorf("write output for %s: %w", label, err)
		}
	}

	// If output path specified, write the last result (or all as array — for now, last).
	if outputPath != "" {
		lastResult := results[len(results)-1].Result
		return writeBatchOutput(lastResult, scenarioName, games, string(aiType), outputPath, format)
	}

	return nil
}

// loadScenarioJSON loads raw scenario JSON bytes from a builtin name or file path.
func loadScenarioJSON(nameOrPath string) ([]byte, error) {
	data, err := server.LoadBuiltinScenarioJSON(nameOrPath)
	if err == nil {
		return data, nil
	}
	return os.ReadFile(nameOrPath)
}

// writeBatchOutput writes batch results to stdout or a file.
func writeBatchOutput(result *batch.BatchResult, scenarioName string, games int, ai string, outputPath string, format string) error {
	reportConfig := metrics.ReportConfig{
		Scenario: scenarioName,
		Games:    games,
		AI:       ai,
	}

	var output string
	var err error

	switch format {
	case "json":
		output, err = metrics.GenerateJSON(reportConfig, result.Summaries, result.BreakageData, result.BreakageReport)
		if err != nil {
			return fmt.Errorf("generate JSON: %w", err)
		}
	case "csv":
		output = metrics.GenerateCSV(result.Summaries)
	default:
		return fmt.Errorf("unknown format: %q (use json or csv)", format)
	}

	if outputPath == "" {
		fmt.Println(output)
		return nil
	}

	if err := os.WriteFile(outputPath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	fmt.Fprintf(os.Stderr, "results written to %s\n", outputPath)
	return nil
}

// runBalanceMode starts the Balance Dashboard.
func runBalanceMode(scenarioName string, games int) error {
	sc, err := server.LoadScenario(scenarioName)
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	config := balance.DashboardConfig{
		Scenario:     sc,
		ScenarioName: scenarioName,
		Games:        games,
		AI:           batch.AISimple,
		BaseSeed:     42,
		Output:       os.Stdout,
		Input:        os.Stdin,
	}

	dash, err := balance.NewDashboard(config)
	if err != nil {
		return fmt.Errorf("create dashboard: %w", err)
	}

	baseline, err := dash.Run()
	if err != nil {
		return fmt.Errorf("run baseline: %w", err)
	}

	// If there are alerts, show sweep suggestions.
	alerts := baseline.BatchResult.BreakageReport.Alerts
	if len(alerts) == 0 {
		return nil
	}

	// Generate and display suggestions for each alert.
	_, _ = fmt.Fprintln(os.Stdout)
	_, _ = fmt.Fprintln(os.Stdout, "--- Sweep Suggestions ---")
	for _, alert := range alerts {
		suggestions := balance.SuggestSweep(alert)
		_, _ = fmt.Fprint(os.Stdout, balance.FormatSuggestions(suggestions))

		// Run sweep for each suggestion and compare.
		scenarioJSON, jsonErr := loadScenarioJSON(scenarioName)
		if jsonErr != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot load scenario JSON for sweep: %v\n", jsonErr)
			continue
		}

		for _, s := range suggestions {
			sweepParam := batch.SweepParam{
				Key:    s.ParamKey,
				Values: s.Values,
			}
			baseConfig := batch.BatchConfig{
				Games:    games,
				BaseSeed: 42,
				AI:       batch.AISimple,
			}
			sweepResults, sweepErr := batch.RunSweep(scenarioJSON, sweepParam, baseConfig)
			if sweepErr != nil {
				fmt.Fprintf(os.Stderr, "warning: sweep failed for %s: %v\n", s.ParamKey, sweepErr)
				continue
			}

			comparison := balance.CompareResults(baseline.BatchResult.BreakageReport, alert.MetricID, sweepResults)
			_, _ = fmt.Fprint(os.Stdout, balance.FormatComparison(comparison))
		}
	}

	return nil
}

// runReplayMode loads and plays back a replay file.
func runReplayMode(scenarioName string, replayPath string) error {
	sc, err := server.LoadScenario(scenarioName)
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	gs, err := server.NewGameServer(sc, 0)
	if err != nil {
		return fmt.Errorf("create game server: %w", err)
	}

	result, err := gs.PlayReplayFrom(replayPath)
	if err != nil {
		return fmt.Errorf("play replay: %w", err)
	}

	fmt.Printf("リプレイ完了: %s (Tick %d, %s)\n", result.Status, result.FinalTick, result.Reason)
	return nil
}
