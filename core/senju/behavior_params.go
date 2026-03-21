package senju

import (
	_ "embed"
	"encoding/json"
)

//go:embed behavior_params_data.json
var defaultBehaviorParamsJSON []byte

// BehaviorParams holds the tunable parameters for beast AI behaviors.
type BehaviorParams struct {
	// FleeHPThreshold is the HP ratio (HP/MaxHP) at or below which a beast flees.
	FleeHPThreshold float64 `json:"flee_hp_threshold"`
	// ChaseTimeoutTicks is the maximum number of ticks a beast will chase
	// before giving up and returning to its previous behavior.
	ChaseTimeoutTicks int `json:"chase_timeout_ticks"`
	// PatrolRestTicks is the number of ticks a patrolling beast stays in each
	// room before advancing to the next room in its route.
	PatrolRestTicks int `json:"patrol_rest_ticks"`
}

// DefaultBehaviorParams returns the default behavior parameters loaded from
// the embedded JSON data.
func DefaultBehaviorParams() *BehaviorParams {
	p, _ := LoadBehaviorParams(defaultBehaviorParamsJSON)
	return p
}

// LoadBehaviorParams parses behavior parameters from JSON data.
func LoadBehaviorParams(data []byte) (*BehaviorParams, error) {
	var p BehaviorParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
