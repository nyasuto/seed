package economy

import "github.com/ponpoko/chaosseed-core/types"

// DeficitSeverity represents the severity level of a chi deficit.
type DeficitSeverity int

const (
	// None means no deficit occurred.
	None DeficitSeverity = iota
	// Mild means the shortage is below the mild threshold.
	Mild
	// Moderate means the shortage is between mild and moderate thresholds.
	Moderate
	// Severe means the shortage exceeds the moderate threshold.
	Severe
)

// DeficitResult holds the outcome of deficit processing for a single tick.
type DeficitResult struct {
	// Shortage is the amount of chi that could not be paid.
	Shortage float64
	// Severity is the deficit severity level.
	Severity DeficitSeverity
	// GrowthPenalty is the multiplier applied to beast growth (1.0 = no penalty).
	GrowthPenalty float64
	// CapacityPenalty is the multiplier applied to room chi capacity (1.0 = no penalty).
	CapacityPenalty float64
	// TrapDisabled indicates whether traps are disabled due to deficit.
	TrapDisabled bool
	// BeastHPDrain is the HP reduction applied to beasts.
	BeastHPDrain int
}

// DeficitProcessor handles the case where maintenance costs exceed
// the chi pool balance.
type DeficitProcessor struct {
	params *DeficitParams
}

// NewDeficitProcessor creates a DeficitProcessor with the given parameters.
func NewDeficitProcessor(params *DeficitParams) *DeficitProcessor {
	return &DeficitProcessor{params: params}
}

// ProcessDeficit processes a maintenance payment against the chi pool.
// If the pool cannot cover the full maintenance cost, it withdraws what
// it can and determines penalties based on the shortage ratio.
func (dp *DeficitProcessor) ProcessDeficit(chiPool *ChiPool, maintenance MaintenanceBreakdown, tick types.Tick) DeficitResult {
	if maintenance.Total <= 0 {
		return DeficitResult{
			GrowthPenalty:   1.0,
			CapacityPenalty: 1.0,
		}
	}

	// Attempt to withdraw the full maintenance cost.
	err := chiPool.Withdraw(maintenance.Total, Deficit, "maintenance deficit processing", tick)
	if err == nil {
		// Fully paid — no deficit.
		return DeficitResult{
			GrowthPenalty:   1.0,
			CapacityPenalty: 1.0,
		}
	}

	// Partial payment occurred. Calculate shortage and ratio.
	// After partial withdrawal, balance is 0. The shortage is the unpaid portion.
	// Look at the last transaction to figure out the actual amount withdrawn.
	shortage := maintenance.Total - chiPool.History[len(chiPool.History)-1].Amount

	ratio := shortage / maintenance.Total

	result := DeficitResult{
		Shortage:        shortage,
		GrowthPenalty:   1.0,
		CapacityPenalty: 1.0,
	}

	switch {
	case ratio >= dp.params.ModerateThreshold:
		result.Severity = Severe
		result.GrowthPenalty = 0.0
		result.CapacityPenalty = dp.params.ModerateCapacityPenalty
		result.TrapDisabled = dp.params.SevereTrapDisable
		result.BeastHPDrain = dp.params.SevereHPDrain
	case ratio >= dp.params.MildThreshold:
		result.Severity = Moderate
		result.GrowthPenalty = 0.0
		result.CapacityPenalty = dp.params.ModerateCapacityPenalty
	default:
		result.Severity = Mild
		result.GrowthPenalty = dp.params.MildGrowthPenalty
	}

	return result
}
