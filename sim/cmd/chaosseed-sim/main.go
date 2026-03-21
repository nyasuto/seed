// Package main is the entry point for chaosseed-sim CLI.
package main

import (
	"flag"
	"fmt"
	"os"
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

	if selected == 0 {
		fmt.Fprintln(os.Stderr, "error: specify one of --human, --ai, --batch, or --balance")
		flag.Usage()
		return 1
	}
	if selected > 1 {
		fmt.Fprintln(os.Stderr, "error: specify only one mode at a time")
		return 1
	}

	switch {
	case *humanMode:
		fmt.Fprintln(os.Stderr, "human mode: not implemented")
	case *aiMode:
		fmt.Fprintln(os.Stderr, "ai mode: not implemented")
	case *batchMode:
		fmt.Fprintln(os.Stderr, "batch mode: not implemented")
	case *balanceMode:
		fmt.Fprintln(os.Stderr, "balance mode: not implemented")
	}
	return 1
}
