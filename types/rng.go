package types

import "math/rand"

// RNG is the interface for random number generation.
// All game logic that requires randomness must accept an RNG parameter
// to ensure deterministic behavior with the same seed.
type RNG interface {
	// Intn returns a non-negative pseudo-random int in the half-open interval [0,n).
	Intn(n int) int
	// Float64 returns a pseudo-random float64 in the half-open interval [0.0,1.0).
	Float64() float64
}

// seededRNG wraps math/rand.Rand to implement the RNG interface.
type seededRNG struct {
	r *rand.Rand
}

// NewSeededRNG returns a deterministic RNG initialized with the given seed.
func NewSeededRNG(seed int64) RNG {
	return &seededRNG{r: rand.New(rand.NewSource(seed))}
}

func (s *seededRNG) Intn(n int) int {
	return s.r.Intn(n)
}

func (s *seededRNG) Float64() float64 {
	return s.r.Float64()
}
