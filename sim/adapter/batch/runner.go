package batch

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"sync"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/sim/metrics"
	"github.com/nyasuto/seed/sim/server"
)

// AIType specifies the AI strategy used for batch games.
type AIType string

const (
	// AISimple uses core's SimpleAIPlayer strategy.
	AISimple AIType = "simple"
	// AINoop uses a no-op provider that takes no actions (passive observer).
	AINoop AIType = "noop"
)

// BatchConfig holds configuration for a batch run.
type BatchConfig struct {
	// Scenario is the scenario to run.
	Scenario *scenario.Scenario
	// Games is the number of games to execute.
	Games int
	// BaseSeed is the base RNG seed; each game gets BaseSeed + gameIndex.
	BaseSeed int64
	// AI is the AI strategy to use.
	AI AIType
	// Parallel is the number of goroutines; 0 means runtime.NumCPU().
	Parallel int
	// Progress is an optional writer for progress messages (typically os.Stderr).
	// If nil, no progress is reported.
	Progress io.Writer
}

// BatchResult holds the results of a completed batch run.
type BatchResult struct {
	// Summaries contains a GameSummary for each game, ordered by game index.
	Summaries []metrics.GameSummary
	// BreakageReport is the aggregate breakage analysis across all games.
	BreakageReport metrics.BreakageReport
	// BreakageData contains per-game breakage data, ordered by game index.
	BreakageData []metrics.BreakageData
}

// BatchRunner executes multiple games in parallel and collects results.
type BatchRunner struct {
	config BatchConfig
}

// NewBatchRunner creates a BatchRunner with the given configuration.
func NewBatchRunner(config BatchConfig) (*BatchRunner, error) {
	if config.Scenario == nil {
		return nil, fmt.Errorf("scenario must not be nil")
	}
	if config.Games <= 0 {
		return nil, fmt.Errorf("games must be positive, got %d", config.Games)
	}
	if config.Parallel < 0 {
		return nil, fmt.Errorf("parallel must be non-negative, got %d", config.Parallel)
	}
	return &BatchRunner{config: config}, nil
}

// gameResult holds the output of a single game execution, keyed by index.
type gameResult struct {
	index    int
	summary  metrics.GameSummary
	breakage metrics.BreakageData
	err      error
}

// Run executes the batch and returns aggregated results.
func (br *BatchRunner) Run() (*BatchResult, error) {
	parallel := br.config.Parallel
	if parallel <= 0 {
		parallel = runtime.NumCPU()
	}
	if parallel > br.config.Games {
		parallel = br.config.Games
	}

	results := make([]gameResult, br.config.Games)
	jobs := make(chan int, br.config.Games)

	// Progress tracking.
	var mu sync.Mutex
	completed := 0

	var wg sync.WaitGroup
	wg.Add(parallel)

	for w := 0; w < parallel; w++ {
		go func() {
			defer wg.Done()
			for idx := range jobs {
				gr := br.runSingleGame(idx)
				results[idx] = gr

				if br.config.Progress != nil {
					mu.Lock()
					completed++
					c := completed
					// Report progress periodically: every 10% or at completion.
					total := br.config.Games
					if c == total || (total >= 10 && c%(total/10) == 0) {
						fmt.Fprintf(br.config.Progress, "%d/%d games completed...\n", c, total)
					}
					mu.Unlock()
				}
			}
		}()
	}

	for i := 0; i < br.config.Games; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	// Collect results in order.
	summaries := make([]metrics.GameSummary, br.config.Games)
	breakageData := make([]metrics.BreakageData, br.config.Games)

	for i, r := range results {
		if r.err != nil {
			return nil, fmt.Errorf("game %d failed: %w", i, r.err)
		}
		summaries[i] = r.summary
		breakageData[i] = r.breakage
	}

	// Aggregate breakage: average per-game metrics, then detect.
	// B10 layout entropy requires room positions which are not available
	// in GameSnapshot; it will be computed externally when position data
	// is available.
	aggBreakage := aggregateBreakageData(breakageData, 0)
	report := metrics.DetectBreakage(aggBreakage, metrics.BreakageThresholds{})

	return &BatchResult{
		Summaries:      summaries,
		BreakageReport: report,
		BreakageData:   breakageData,
	}, nil
}

// runSingleGame executes one game and returns its results.
func (br *BatchRunner) runSingleGame(index int) gameResult {
	seed := br.config.BaseSeed + int64(index)

	gs, err := server.NewGameServer(br.config.Scenario, seed)
	if err != nil {
		return gameResult{index: index, err: err}
	}

	var provider server.ActionProvider
	switch br.config.AI {
	case AISimple:
		provider = &simpleAIBatchProvider{gs: gs}
	default:
		provider = &batchProvider{}
	}

	result, err := gs.RunGame(provider)
	if err != nil {
		return gameResult{index: index, err: err}
	}

	summary := gs.Collector().OnGameEnd(&result)
	bd := gs.Collector().BreakageMetrics()

	return gameResult{
		index:    index,
		summary:  *summary,
		breakage: bd,
	}
}

// batchProvider implements server.ActionProvider for batch mode.
// It always returns NoAction (the game runs passively).
type batchProvider struct{}

func (p *batchProvider) ProvideActions(_ scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}

func (p *batchProvider) OnTickComplete(_ scenario.GameSnapshot) {}

func (p *batchProvider) OnGameEnd(_ simulation.RunResult) {}

// simpleAIBatchProvider implements server.ActionProvider using core's
// SimpleAIPlayer. The AI is lazily created on the first ProvideActions
// call because the engine's GameState is only available after RunGame
// creates the engine.
type simpleAIBatchProvider struct {
	gs *server.GameServer
	ai simulation.AIPlayer
}

func (p *simpleAIBatchProvider) ProvideActions(snapshot scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	if p.ai == nil {
		p.ai = simulation.NewSimpleAIPlayer(p.gs.Engine().State)
	}
	actions := p.ai.DecideActions(snapshot)
	return actions, nil
}

func (p *simpleAIBatchProvider) OnTickComplete(_ scenario.GameSnapshot) {}

func (p *simpleAIBatchProvider) OnGameEnd(_ simulation.RunResult) {}

// aggregateBreakageData combines per-game BreakageData into a single
// representative BreakageData for threshold detection. Boolean flags use
// majority vote; numeric values use the mean.
func aggregateBreakageData(data []metrics.BreakageData, layoutEntropy float64) metrics.BreakageData {
	n := len(data)
	if n == 0 {
		return metrics.BreakageData{}
	}

	var (
		b01Sum, b02Sum                       int
		firstWaveCount                       int
		b03Sum, b05Sum                       float64
		b04Count                             int
		b06Count, b07Count, b08Count         int
		b09Sum, b11Sum                       float64
		b09Count                             int
	)

	for _, d := range data {
		if d.FirstWaveRecorded {
			firstWaveCount++
			b01Sum += d.B01
			b02Sum += d.B02
		}
		b03Sum += d.B03
		if d.B04ZeroBuildable {
			b04Count++
		}
		b05Sum += d.B05
		if d.B06Stomp {
			b06Count++
		}
		if d.B07EarlyWipe {
			b07Count++
		}
		if d.B08Perfection {
			b08Count++
		}
		if d.B09RoomLevelRatio > 0 {
			b09Sum += d.B09RoomLevelRatio
			b09Count++
		}
		b11Sum += d.B11SurplusRate
	}

	agg := metrics.BreakageData{
		B10LayoutEntropy: layoutEntropy,
	}

	if firstWaveCount > 0 {
		agg.FirstWaveRecorded = true
		agg.B01 = b01Sum / firstWaveCount
		agg.B02 = b02Sum / firstWaveCount
	}

	agg.B03 = b03Sum / float64(n)
	agg.B04ZeroBuildable = b04Count > n/2 // majority
	agg.B05 = b05Sum / float64(n)
	agg.B06Stomp = b06Count > n/2
	agg.B07EarlyWipe = b07Count > n/2
	agg.B08Perfection = b08Count > n/2
	if b09Count > 0 {
		agg.B09RoomLevelRatio = b09Sum / float64(b09Count)
	}
	agg.B11SurplusRate = b11Sum / float64(n)

	return agg
}

// SortSummariesByTicks sorts a slice of GameSummary by TotalTicks ascending.
func SortSummariesByTicks(summaries []metrics.GameSummary) {
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].TotalTicks < summaries[j].TotalTicks
	})
}
