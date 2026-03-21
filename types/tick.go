package types

// Tick represents a unit of in-game time.
// It is the smallest time unit in the simulation; all time-based
// operations (chi flow, senju actions, invasions, etc.) advance in ticks.
type Tick uint64
