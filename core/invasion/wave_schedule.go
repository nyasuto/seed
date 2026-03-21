package invasion

import (
	"encoding/json"
	"fmt"
)

// WaveSchedule defines a sequence of invasion waves for a scenario.
// In Phase 4, this is used with a fixed test schedule.
// Phase 6's scenario system will override this with dynamic schedules.
type WaveSchedule struct {
	// Waves is the ordered list of wave configurations.
	Waves []WaveConfig `json:"waves"`
}

// LoadWaveSchedule parses a JSON byte slice into a WaveSchedule.
func LoadWaveSchedule(data []byte) (*WaveSchedule, error) {
	var schedule WaveSchedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return nil, fmt.Errorf("unmarshal wave schedule: %w", err)
	}
	if len(schedule.Waves) == 0 {
		return nil, fmt.Errorf("wave schedule must contain at least one wave")
	}
	for i, w := range schedule.Waves {
		if w.MinInvaders <= 0 || w.MaxInvaders <= 0 {
			return nil, fmt.Errorf("wave %d: min and max invaders must be positive", i)
		}
		if w.MinInvaders > w.MaxInvaders {
			return nil, fmt.Errorf("wave %d: min invaders (%d) must not exceed max invaders (%d)", i, w.MinInvaders, w.MaxInvaders)
		}
		if w.Difficulty <= 0 {
			return nil, fmt.Errorf("wave %d: difficulty must be positive", i)
		}
	}
	return &schedule, nil
}
