package invasion

// RetreatReason represents why an invader decided to retreat.
type RetreatReason int

const (
	// ReasonLowHP means the invader's HP fell below its retreat threshold.
	ReasonLowHP RetreatReason = iota
	// ReasonMoraleBroken means half or more of the wave companions are defeated.
	ReasonMoraleBroken
	// ReasonGoalComplete means the invader achieved its objective.
	ReasonGoalComplete
)

// String returns the name of the retreat reason.
func (r RetreatReason) String() string {
	switch r {
	case ReasonLowHP:
		return "LowHP"
	case ReasonMoraleBroken:
		return "MoraleBroken"
	case ReasonGoalComplete:
		return "GoalComplete"
	default:
		return "Unknown"
	}
}

// RetreatEvaluator determines whether an invader should retreat from the cave.
type RetreatEvaluator struct {
	classRegistry *InvaderClassRegistry
}

// NewRetreatEvaluator creates a new RetreatEvaluator with the given class registry.
func NewRetreatEvaluator(registry *InvaderClassRegistry) *RetreatEvaluator {
	return &RetreatEvaluator{classRegistry: registry}
}

// ShouldRetreat evaluates whether the given invader should retreat.
// It checks three conditions in order:
//  1. Goal achieved → retreat to carry spoils home
//  2. HP ≤ MaxHP × RetreatThreshold → retreat due to low health
//  3. Half or more of wave companions defeated → retreat due to morale break
//
// The waveInvaders parameter includes all invaders in the same wave (including the subject).
// Returns true and the reason if the invader should retreat.
func (re *RetreatEvaluator) ShouldRetreat(invader *Invader, waveInvaders []*Invader) (bool, RetreatReason) {
	// Already retreating or defeated — no re-evaluation needed.
	if invader.State == Retreating || invader.State == Defeated {
		return false, 0
	}

	// 3. Goal achieved → retreat with spoils.
	if invader.State == GoalAchieved {
		return true, ReasonGoalComplete
	}

	// 1. HP ≤ MaxHP × RetreatThreshold → retreat.
	class, err := re.classRegistry.Get(invader.ClassID)
	if err == nil && invader.MaxHP > 0 {
		threshold := float64(invader.MaxHP) * class.RetreatThreshold
		if float64(invader.HP) <= threshold {
			return true, ReasonLowHP
		}
	}

	// 2. Half or more of wave companions defeated → morale break.
	if len(waveInvaders) > 1 {
		defeatedCount := 0
		for _, inv := range waveInvaders {
			if inv.State == Defeated {
				defeatedCount++
			}
		}
		if defeatedCount*2 >= len(waveInvaders) {
			return true, ReasonMoraleBroken
		}
	}

	return false, 0
}
