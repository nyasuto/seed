package testutil

import "testing"

func TestFixedRNG_Intn(t *testing.T) {
	tests := []struct {
		name     string
		intValue int
		n        int
		want     int
	}{
		{"zero value", 0, 10, 0},
		{"value less than n", 3, 10, 3},
		{"value equal to n", 10, 10, 0},
		{"value greater than n", 15, 10, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := &FixedRNG{IntValue: tt.intValue}
			got := rng.Intn(tt.n)
			if got != tt.want {
				t.Errorf("FixedRNG{IntValue: %d}.Intn(%d) = %d, want %d", tt.intValue, tt.n, got, tt.want)
			}
			// Verify it returns the same value every time
			got2 := rng.Intn(tt.n)
			if got2 != tt.want {
				t.Errorf("FixedRNG should return same value on repeated calls, got %d then %d", got, got2)
			}
		})
	}
}

func TestFixedRNG_Float64(t *testing.T) {
	rng := &FixedRNG{FloatValue: 0.42}
	got := rng.Float64()
	if got != 0.42 {
		t.Errorf("FixedRNG{FloatValue: 0.42}.Float64() = %f, want 0.42", got)
	}
	// Verify it returns the same value every time
	got2 := rng.Float64()
	if got2 != 0.42 {
		t.Errorf("FixedRNG should return same value on repeated calls, got %f then %f", got, got2)
	}
}

func TestNewTestRNG_Deterministic(t *testing.T) {
	rng1 := NewTestRNG(42)
	rng2 := NewTestRNG(42)

	// Same seed should produce same sequence
	for i := 0; i < 10; i++ {
		v1 := rng1.Intn(100)
		v2 := rng2.Intn(100)
		if v1 != v2 {
			t.Errorf("iteration %d: NewTestRNG(42) produced different values: %d vs %d", i, v1, v2)
		}
	}
}

func TestNewTestRNG_DifferentSeeds(t *testing.T) {
	rng1 := NewTestRNG(1)
	rng2 := NewTestRNG(2)

	// Different seeds should (very likely) produce different sequences
	same := true
	for i := 0; i < 10; i++ {
		if rng1.Intn(1000) != rng2.Intn(1000) {
			same = false
			break
		}
	}
	if same {
		t.Error("NewTestRNG with different seeds produced identical sequences")
	}
}
