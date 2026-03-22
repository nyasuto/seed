package human

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

func TestFormatTickSummary_FirstTick(t *testing.T) {
	var buf bytes.Buffer
	snap := scenario.GameSnapshot{
		Tick:           1,
		CoreHP:         100,
		ChiPoolBalance: 50.0,
		BeastCount:     2,
		AliveBeasts:    2,
		DefeatedWaves:  0,
		TotalWaves:     5,
	}

	FormatTickSummary(&buf, nil, snap)
	output := buf.String()

	if !strings.Contains(output, "Tick 1") {
		t.Errorf("expected Tick 1, got: %s", output)
	}
	if !strings.Contains(output, "CoreHP: 100") {
		t.Errorf("expected CoreHP: 100, got: %s", output)
	}
	if !strings.Contains(output, "50.0") {
		t.Errorf("expected chi balance 50.0, got: %s", output)
	}
	// Should not contain warnings on first tick.
	if strings.Contains(output, "⚠") {
		t.Errorf("expected no warnings on first tick, got: %s", output)
	}
}

func TestFormatTickSummary_CoreHPDecrease(t *testing.T) {
	var buf bytes.Buffer
	prev := scenario.GameSnapshot{
		Tick:           1,
		CoreHP:         100,
		ChiPoolBalance: 50.0,
		BeastCount:     2,
		AliveBeasts:    2,
	}
	current := scenario.GameSnapshot{
		Tick:           2,
		CoreHP:         75,
		ChiPoolBalance: 45.0,
		BeastCount:     2,
		AliveBeasts:    2,
	}

	FormatTickSummary(&buf, &prev, current)
	output := buf.String()

	if !strings.Contains(output, "-25!") {
		t.Errorf("expected CoreHP decrease -25!, got: %s", output)
	}
}

func TestFormatTickSummary_BeastLost(t *testing.T) {
	var buf bytes.Buffer
	prev := scenario.GameSnapshot{
		Tick:        1,
		CoreHP:      100,
		BeastCount:  3,
		AliveBeasts: 3,
	}
	current := scenario.GameSnapshot{
		Tick:        2,
		CoreHP:      100,
		BeastCount:  3,
		AliveBeasts: 1,
	}

	FormatTickSummary(&buf, &prev, current)
	output := buf.String()

	if !strings.Contains(output, "2体 戦闘不能") {
		t.Errorf("expected beast loss warning, got: %s", output)
	}
}

func TestFormatTickSummary_WaveDefeated(t *testing.T) {
	var buf bytes.Buffer
	prev := scenario.GameSnapshot{
		Tick:          1,
		CoreHP:        100,
		DefeatedWaves: 0,
		TotalWaves:    5,
	}
	current := scenario.GameSnapshot{
		Tick:          2,
		CoreHP:        100,
		DefeatedWaves: 1,
		TotalWaves:    5,
	}

	FormatTickSummary(&buf, &prev, current)
	output := buf.String()

	if !strings.Contains(output, "1波 撃退") {
		t.Errorf("expected wave defeated message, got: %s", output)
	}
}

func TestFormatTickSummary_LowCoreHPWarning(t *testing.T) {
	var buf bytes.Buffer
	current := scenario.GameSnapshot{
		Tick:   5,
		CoreHP: 15,
	}

	FormatTickSummary(&buf, nil, current)
	output := buf.String()

	if !strings.Contains(output, "危険水域") {
		t.Errorf("expected low CoreHP warning, got: %s", output)
	}
}

func TestFormatFastForwardSummary(t *testing.T) {
	var buf bytes.Buffer
	start := scenario.GameSnapshot{
		Tick:           types.Tick(1),
		CoreHP:         100,
		ChiPoolBalance: 50.0,
		BeastCount:     2,
		AliveBeasts:    2,
		DefeatedWaves:  0,
		TotalWaves:     5,
	}
	end := scenario.GameSnapshot{
		Tick:           types.Tick(51),
		CoreHP:         80,
		ChiPoolBalance: 120.0,
		BeastCount:     3,
		AliveBeasts:    3,
		DefeatedWaves:  2,
		TotalWaves:     5,
	}

	FormatFastForwardSummary(&buf, start, end)
	output := buf.String()

	if !strings.Contains(output, "早送り完了: 50ティック") {
		t.Errorf("expected FF summary header, got: %s", output)
	}
	if !strings.Contains(output, "100 → 80") {
		t.Errorf("expected CoreHP change, got: %s", output)
	}
	if !strings.Contains(output, "+70.0") {
		t.Errorf("expected chi delta, got: %s", output)
	}
	if !strings.Contains(output, "+2") {
		t.Errorf("expected wave defeated count, got: %s", output)
	}
}

func TestFormatGameEnd_Won(t *testing.T) {
	var buf bytes.Buffer
	result := simulation.RunResult{
		Result: simulation.GameResult{
			Status:    simulation.Won,
			FinalTick: 200,
			Reason:    "all waves defeated",
		},
		TickCount: 200,
		Statistics: simulation.RunStatistics{
			PeakChi:        300.0,
			WavesDefeated:  10,
			FinalFengShui:  0.9,
			Evolutions:     3,
			DamageDealt:    1000,
			DamageReceived: 400,
			DeficitTicks:   5,
		},
	}

	FormatGameEnd(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "勝利") {
		t.Errorf("expected victory, got: %s", output)
	}
	if !strings.Contains(output, "300.0") {
		t.Errorf("expected peak chi, got: %s", output)
	}
	if !strings.Contains(output, "0.90") {
		t.Errorf("expected feng shui score, got: %s", output)
	}
}

func TestFormatGameEnd_Lost(t *testing.T) {
	var buf bytes.Buffer
	result := simulation.RunResult{
		Result: simulation.GameResult{
			Status:    simulation.Lost,
			FinalTick: 50,
			Reason:    "core HP reached 0",
		},
		TickCount: 50,
		Statistics: simulation.RunStatistics{
			PeakChi:       100.0,
			WavesDefeated: 1,
		},
	}

	FormatGameEnd(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "敗北") {
		t.Errorf("expected defeat, got: %s", output)
	}
	if !strings.Contains(output, "core HP reached 0") {
		t.Errorf("expected reason, got: %s", output)
	}
}
