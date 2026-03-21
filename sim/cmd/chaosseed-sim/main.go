// Package main is the entry point for chaosseed-sim CLI.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nyasuto/seed/sim/adapter/human"
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
		fmt.Fprintln(os.Stderr, "ai mode: not implemented")
	case *batchMode:
		fmt.Fprintln(os.Stderr, "batch mode: not implemented")
	case *balanceMode:
		fmt.Fprintln(os.Stderr, "balance mode: not implemented")
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
		fmt.Fprintln(os.Stdout, "ゲームを終了しました。")
		return nil
	}

	return err
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
