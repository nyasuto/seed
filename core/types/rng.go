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

// CheckpointableRNG extends RNG with the ability to capture and restore
// internal state for checkpoint/restore functionality.
type CheckpointableRNG interface {
	RNG
	// RNGState returns the current state of the RNG for checkpointing.
	RNGState() RNGState
}

// RNGState captures the state needed to reconstruct a CheckpointableRNG.
type RNGState struct {
	// Seed is the original seed used to create the RNG.
	Seed int64
	// Calls is the number of underlying Source.Int63() calls made so far.
	Calls int64
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

// countingSource wraps a rand.Source and counts every Int63 call.
type countingSource struct {
	src   rand.Source
	seed  int64
	calls int64
}

func (cs *countingSource) Int63() int64 {
	cs.calls++
	return cs.src.Int63()
}

func (cs *countingSource) Seed(seed int64) {
	cs.src.Seed(seed)
	cs.seed = seed
	cs.calls = 0
}

// checkpointableRNG wraps a counting source to implement CheckpointableRNG.
type checkpointableRNG struct {
	r   *rand.Rand
	src *countingSource
}

// NewCheckpointableRNG returns a deterministic RNG that tracks its internal
// state for checkpoint/restore. Use this instead of NewSeededRNG when
// checkpoint support is needed.
func NewCheckpointableRNG(seed int64) CheckpointableRNG {
	src := &countingSource{
		src:  rand.NewSource(seed),
		seed: seed,
	}
	return &checkpointableRNG{
		r:   rand.New(src),
		src: src,
	}
}

// RestoreRNG reconstructs a CheckpointableRNG from a previously captured
// RNGState. The underlying source is fast-forwarded to the saved position.
func RestoreRNG(state RNGState) CheckpointableRNG {
	src := rand.NewSource(state.Seed)
	// Fast-forward the source to the saved position.
	for i := int64(0); i < state.Calls; i++ {
		src.Int63()
	}
	cs := &countingSource{
		src:   src,
		seed:  state.Seed,
		calls: state.Calls,
	}
	return &checkpointableRNG{
		r:   rand.New(cs),
		src: cs,
	}
}

func (c *checkpointableRNG) Intn(n int) int {
	return c.r.Intn(n)
}

func (c *checkpointableRNG) Float64() float64 {
	return c.r.Float64()
}

func (c *checkpointableRNG) RNGState() RNGState {
	return RNGState{
		Seed:  c.src.seed,
		Calls: c.src.calls,
	}
}
