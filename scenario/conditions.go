package scenario

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// ErrUnknownConditionType indicates that NewCondition was called with a
// ConditionDef whose Type is not recognised by the factory.
var ErrUnknownConditionType = errors.New("unknown condition type")

// NewCondition creates a ConditionEvaluator from a data-driven ConditionDef.
// Returns ErrUnknownConditionType when the def.Type is not supported.
func NewCondition(def ConditionDef) (ConditionEvaluator, error) {
	switch def.Type {
	case "survive_until":
		return newSurviveUntil(def.Params)
	case "defeat_all_waves":
		return &defeatAllWaves{}, nil
	case "fengshui_score":
		return newFengshuiScore(def.Params)
	case "chi_pool":
		return newChiPool(def.Params)
	case "core_destroyed":
		return &coreDestroyed{}, nil
	case "all_beasts_defeated":
		return &allBeastsDefeated{}, nil
	case "bankrupt":
		return newBankrupt(def.Params)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownConditionType, def.Type)
	}
}

// surviveUntilParams holds the typed parameters for surviveUntil.
type surviveUntilParams struct {
	Ticks float64 `json:"ticks"`
}

// surviveUntil evaluates true when the current tick reaches the target
// and CoreHP is still positive.
type surviveUntil struct {
	ticks types.Tick
}

func newSurviveUntil(params json.RawMessage) (*surviveUntil, error) {
	var p surviveUntilParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("survive_until params: %w", err)
	}
	if p.Ticks == 0 {
		return nil, fmt.Errorf("survive_until: missing required parameter \"ticks\"")
	}
	return &surviveUntil{ticks: types.Tick(p.Ticks)}, nil
}

// Evaluate returns true when the game has survived to at least the
// specified tick with a positive CoreHP.
func (c *surviveUntil) Evaluate(snap GameSnapshot) bool {
	return snap.Tick >= c.ticks && snap.CoreHP > 0
}

// defeatAllWaves evaluates true when all invasion waves have been defeated.
type defeatAllWaves struct{}

// Evaluate returns true when DefeatedWaves equals TotalWaves and
// TotalWaves is positive.
func (c *defeatAllWaves) Evaluate(snap GameSnapshot) bool {
	return snap.TotalWaves > 0 && snap.DefeatedWaves >= snap.TotalWaves
}

// fengshuiScoreParams holds the typed parameters for fengshuiScore.
type fengshuiScoreParams struct {
	Threshold float64 `json:"threshold"`
}

// fengshuiScore evaluates true when the cave fengshui score meets or
// exceeds a threshold.
type fengshuiScore struct {
	threshold float64
}

func newFengshuiScore(params json.RawMessage) (*fengshuiScore, error) {
	var p fengshuiScoreParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("fengshui_score params: %w", err)
	}
	if p.Threshold == 0 {
		return nil, fmt.Errorf("fengshui_score: missing required parameter \"threshold\"")
	}
	return &fengshuiScore{threshold: p.Threshold}, nil
}

// Evaluate returns true when the snapshot's fengshui score is at or above
// the configured threshold.
func (c *fengshuiScore) Evaluate(snap GameSnapshot) bool {
	return snap.CaveFengShuiScore >= c.threshold
}

// chiPoolParams holds the typed parameters for chiPool.
type chiPoolParams struct {
	Threshold float64 `json:"threshold"`
}

// chiPool evaluates true when the chi pool balance meets or exceeds a
// threshold.
type chiPool struct {
	threshold float64
}

func newChiPool(params json.RawMessage) (*chiPool, error) {
	var p chiPoolParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("chi_pool params: %w", err)
	}
	if p.Threshold == 0 {
		return nil, fmt.Errorf("chi_pool: missing required parameter \"threshold\"")
	}
	return &chiPool{threshold: p.Threshold}, nil
}

// Evaluate returns true when the snapshot's chi pool balance is at or
// above the configured threshold.
func (c *chiPool) Evaluate(snap GameSnapshot) bool {
	return snap.ChiPoolBalance >= c.threshold
}

// coreDestroyed evaluates true when the core's HP has dropped to zero
// or below (lose condition).
type coreDestroyed struct{}

// Evaluate returns true when CoreHP is zero or negative.
func (c *coreDestroyed) Evaluate(snap GameSnapshot) bool {
	return snap.CoreHP <= 0
}

// allBeastsDefeated evaluates true when no beasts remain alive (lose
// condition).
type allBeastsDefeated struct{}

// Evaluate returns true when AliveBeasts is zero.
func (c *allBeastsDefeated) Evaluate(snap GameSnapshot) bool {
	return snap.AliveBeasts == 0
}

// bankruptParams holds the typed parameters for bankrupt.
type bankruptParams struct {
	Ticks float64 `json:"ticks"`
}

// bankrupt evaluates true when the player has been running a chi deficit
// for too many consecutive ticks (lose condition).
type bankrupt struct {
	ticksThreshold int
}

func newBankrupt(params json.RawMessage) (*bankrupt, error) {
	var p bankruptParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("bankrupt params: %w", err)
	}
	if p.Ticks == 0 {
		return nil, fmt.Errorf("bankrupt: missing required parameter \"ticks\"")
	}
	return &bankrupt{ticksThreshold: int(p.Ticks)}, nil
}

// Evaluate returns true when ConsecutiveDeficitTicks meets or exceeds
// the configured threshold.
func (c *bankrupt) Evaluate(snap GameSnapshot) bool {
	return snap.ConsecutiveDeficitTicks >= c.ticksThreshold
}
