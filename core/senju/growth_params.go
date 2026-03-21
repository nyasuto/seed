package senju

import "encoding/json"

// GrowthParams holds the parameters that control beast growth behavior.
type GrowthParams struct {
	// BaseEXPPerTick is the base amount of EXP gained each tick.
	BaseEXPPerTick int `json:"base_exp_per_tick"`
	// LevelUpBase is the base EXP required for the first level up.
	LevelUpBase int `json:"level_up_base"`
	// LevelUpPerLevel is the additional EXP required per current level.
	LevelUpPerLevel int `json:"level_up_per_level"`
	// ChiConsumptionPerTick is the amount of chi consumed per tick for growth.
	ChiConsumptionPerTick float64 `json:"chi_consumption_per_tick"`
	// MaxLevel is the maximum level a beast can reach.
	MaxLevel int `json:"max_level"`
}

// DefaultGrowthParams returns the default growth parameters.
func DefaultGrowthParams() *GrowthParams {
	return &GrowthParams{
		BaseEXPPerTick:        10,
		LevelUpBase:           100,
		LevelUpPerLevel:       50,
		ChiConsumptionPerTick: 2.0,
		MaxLevel:              50,
	}
}

// LoadGrowthParams parses growth parameters from JSON data.
func LoadGrowthParams(data []byte) (*GrowthParams, error) {
	var p GrowthParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// LevelUpThreshold returns the EXP required to level up from the given level.
// Formula: LevelUpBase + LevelUpPerLevel × currentLevel
func (p *GrowthParams) LevelUpThreshold(level int) int {
	return p.LevelUpBase + p.LevelUpPerLevel*level
}
