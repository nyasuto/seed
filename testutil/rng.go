package testutil

import "github.com/ponpoko/chaosseed-core/types"

// FixedRNG is a mock RNG that always returns the same fixed values.
// Useful for tests where you need completely predictable random output.
type FixedRNG struct {
	IntValue   int
	FloatValue float64
}

// Intn always returns f.IntValue % n.
func (f *FixedRNG) Intn(n int) int {
	return f.IntValue % n
}

// Float64 always returns f.FloatValue.
func (f *FixedRNG) Float64() float64 {
	return f.FloatValue
}

// NewTestRNG returns a deterministic RNG seeded with the given value.
// This is a convenience wrapper around types.NewSeededRNG for use in tests.
func NewTestRNG(seed int64) types.RNG {
	return types.NewSeededRNG(seed)
}
