package senju

import "github.com/ponpoko/chaosseed-core/types"

// DefeatResult represents the outcome of processing a defeated beast.
type DefeatResult struct {
	// BeastID is the ID of the defeated beast.
	BeastID int
	// NewState is the state the beast transitioned to (Stunned).
	NewState BeastState
	// RevivalTick is the tick at which the beast will automatically revive.
	RevivalTick types.Tick
	// LevelPenalty is the actual number of levels lost (may be less than configured if beast is level 1).
	LevelPenalty int
	// RevivalHP is the HP the beast will have upon revival.
	RevivalHP int
}

// DefeatProcessor handles beast defeat logic including stunning and revival calculation.
type DefeatProcessor struct {
	params *DefeatParams
}

// NewDefeatProcessor creates a DefeatProcessor with default parameters.
func NewDefeatProcessor() *DefeatProcessor {
	return &DefeatProcessor{
		params: DefaultDefeatParams(),
	}
}

// NewDefeatProcessorWithParams creates a DefeatProcessor with the given parameters.
func NewDefeatProcessorWithParams(params *DefeatParams) *DefeatProcessor {
	return &DefeatProcessor{
		params: params,
	}
}

// ProcessDefeat handles a beast whose HP has reached 0 or below.
// It transitions the beast to Stunned state and returns the revival parameters.
// The beast's HP is set to 0 and its state becomes Stunned.
// After StunnedDuration ticks, the beast should revive with reduced HP and level.
func (dp *DefeatProcessor) ProcessDefeat(beast *Beast, tick types.Tick) DefeatResult {
	beast.State = Stunned
	beast.HP = 0

	revivalTick := tick + types.Tick(dp.params.StunnedDuration)

	levelPenalty := dp.params.LevelPenalty
	newLevel := beast.Level - levelPenalty
	if newLevel < 1 {
		levelPenalty = beast.Level - 1
	}

	revivalHP := max(int(float64(beast.MaxHP)*dp.params.RevivalHPRatio), 1)

	return DefeatResult{
		BeastID:      beast.ID,
		NewState:     Stunned,
		RevivalTick:  revivalTick,
		LevelPenalty: levelPenalty,
		RevivalHP:    revivalHP,
	}
}
