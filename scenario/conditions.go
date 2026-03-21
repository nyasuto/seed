package scenario

import (
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

// surviveUntil evaluates true when the current tick reaches the target
// and CoreHP is still positive.
type surviveUntil struct {
	ticks types.Tick
}

func newSurviveUntil(params map[string]any) (*surviveUntil, error) {
	ticks, err := paramFloat64(params, "ticks")
	if err != nil {
		return nil, fmt.Errorf("survive_until: %w", err)
	}
	return &surviveUntil{ticks: types.Tick(ticks)}, nil
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

// fengshuiScore evaluates true when the cave fengshui score meets or
// exceeds a threshold.
type fengshuiScore struct {
	threshold float64
}

func newFengshuiScore(params map[string]any) (*fengshuiScore, error) {
	threshold, err := paramFloat64(params, "threshold")
	if err != nil {
		return nil, fmt.Errorf("fengshui_score: %w", err)
	}
	return &fengshuiScore{threshold: threshold}, nil
}

// Evaluate returns true when the snapshot's fengshui score is at or above
// the configured threshold.
func (c *fengshuiScore) Evaluate(snap GameSnapshot) bool {
	return snap.CaveFengShuiScore >= c.threshold
}

// chiPool evaluates true when the chi pool balance meets or exceeds a
// threshold.
type chiPool struct {
	threshold float64
}

func newChiPool(params map[string]any) (*chiPool, error) {
	threshold, err := paramFloat64(params, "threshold")
	if err != nil {
		return nil, fmt.Errorf("chi_pool: %w", err)
	}
	return &chiPool{threshold: threshold}, nil
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

// bankrupt evaluates true when the player has been running a chi deficit
// for too many consecutive ticks (lose condition).
type bankrupt struct {
	ticksThreshold int
}

func newBankrupt(params map[string]any) (*bankrupt, error) {
	ticks, err := paramFloat64(params, "ticks")
	if err != nil {
		return nil, fmt.Errorf("bankrupt: %w", err)
	}
	return &bankrupt{ticksThreshold: int(ticks)}, nil
}

// Evaluate returns true when ConsecutiveDeficitTicks meets or exceeds
// the configured threshold.
func (c *bankrupt) Evaluate(snap GameSnapshot) bool {
	return snap.ConsecutiveDeficitTicks >= c.ticksThreshold
}

// paramFloat64 extracts a float64 parameter by key from a params map.
// JSON numbers unmarshal as float64, so this handles the common case.
func paramFloat64(params map[string]any, key string) (float64, error) {
	v, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("missing required parameter %q", key)
	}
	f, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("parameter %q must be a number, got %T", key, v)
	}
	return f, nil
}
