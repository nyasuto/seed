package types

import "testing"

func TestNewSeededRNG_Deterministic(t *testing.T) {
	rng1 := NewSeededRNG(42)
	rng2 := NewSeededRNG(42)

	for i := 0; i < 100; i++ {
		v1 := rng1.Intn(1000)
		v2 := rng2.Intn(1000)
		if v1 != v2 {
			t.Fatalf("iteration %d: same seed produced different Intn results: %d vs %d", i, v1, v2)
		}
	}
}

func TestNewSeededRNG_DeterministicFloat64(t *testing.T) {
	rng1 := NewSeededRNG(99)
	rng2 := NewSeededRNG(99)

	for i := 0; i < 100; i++ {
		v1 := rng1.Float64()
		v2 := rng2.Float64()
		if v1 != v2 {
			t.Fatalf("iteration %d: same seed produced different Float64 results: %f vs %f", i, v1, v2)
		}
	}
}

func TestNewSeededRNG_DifferentSeeds(t *testing.T) {
	rng1 := NewSeededRNG(1)
	rng2 := NewSeededRNG(2)

	same := true
	for i := 0; i < 10; i++ {
		if rng1.Intn(1000) != rng2.Intn(1000) {
			same = false
			break
		}
	}
	if same {
		t.Error("different seeds produced identical sequences")
	}
}

func TestNewSeededRNG_IntnRange(t *testing.T) {
	rng := NewSeededRNG(123)
	for i := 0; i < 1000; i++ {
		v := rng.Intn(10)
		if v < 0 || v >= 10 {
			t.Fatalf("Intn(10) returned %d, want [0,10)", v)
		}
	}
}

func TestNewSeededRNG_Float64Range(t *testing.T) {
	rng := NewSeededRNG(456)
	for i := 0; i < 1000; i++ {
		v := rng.Float64()
		if v < 0.0 || v >= 1.0 {
			t.Fatalf("Float64() returned %f, want [0.0,1.0)", v)
		}
	}
}
