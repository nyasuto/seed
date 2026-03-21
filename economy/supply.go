package economy

import "github.com/ponpoko/chaosseed-core/fengshui"

// SupplyCalculator calculates per-tick chi supply from dragon veins.
type SupplyCalculator struct {
	params *SupplyParams
}

// NewSupplyCalculator creates a SupplyCalculator with the given parameters.
func NewSupplyCalculator(params *SupplyParams) *SupplyCalculator {
	return &SupplyCalculator{params: params}
}

// CalcTickSupply computes the chi supply for a single tick.
//
// Formula:
//  1. baseSupply = len(veins) × BaseSupplyPerVein
//  2. fillBonus  = averageChiFillRatio × ChiRatioSupplyWeight
//  3. fengShuiMul = linear map of caveScore [0,1] → [FengShuiMinMultiplier, FengShuiMaxMultiplier]
//  4. totalSupply = (baseSupply + fillBonus) × fengShuiMul
func (sc *SupplyCalculator) CalcTickSupply(veins []fengshui.DragonVein, roomChis map[int]*fengshui.RoomChi, caveScore float64) float64 {
	if len(veins) == 0 {
		return 0
	}

	// 1. base supply
	baseSupply := float64(len(veins)) * sc.params.BaseSupplyPerVein

	// 2. fill ratio bonus
	avgRatio := averageChiFillRatio(roomChis)
	fillBonus := avgRatio * sc.params.ChiRatioSupplyWeight

	// 3. feng shui multiplier (linear interpolation)
	fengShuiMul := sc.params.FengShuiMinMultiplier +
		caveScore*(sc.params.FengShuiMaxMultiplier-sc.params.FengShuiMinMultiplier)

	// 4. total
	return (baseSupply + fillBonus) * fengShuiMul
}

// averageChiFillRatio returns the average Current/Capacity ratio across all rooms.
// Returns 0 if the map is empty.
func averageChiFillRatio(roomChis map[int]*fengshui.RoomChi) float64 {
	if len(roomChis) == 0 {
		return 0
	}
	var sum float64
	for _, rc := range roomChis {
		if rc.Capacity > 0 {
			sum += rc.Current / rc.Capacity
		}
	}
	return sum / float64(len(roomChis))
}
