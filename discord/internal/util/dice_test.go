package util

import (
	"math/rand"
	"testing"
)

func TestDiceRoll(t *testing.T) {
	// With a fixed seed, results should be deterministic
	rng := rand.New(rand.NewSource(42))

	// Roll 1d6 multiple times - all should be between 1 and 6
	for i := 0; i < 100; i++ {
		result := DiceRoll(1, 6, rng)
		if result < 1 || result > 6 {
			t.Errorf("DiceRoll(1, 6) = %d, want 1-6", result)
		}
	}

	// Roll 3d15 - should be between 3 and 45
	rng = rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		result := DiceRoll(3, 15, rng)
		if result < 3 || result > 45 {
			t.Errorf("DiceRoll(3, 15) = %d, want 3-45", result)
		}
	}
}

func TestDiceRollDeterministic(t *testing.T) {
	rng1 := rand.New(rand.NewSource(123))
	rng2 := rand.New(rand.NewSource(123))

	for i := 0; i < 10; i++ {
		r1 := DiceRoll(3, 15, rng1)
		r2 := DiceRoll(3, 15, rng2)
		if r1 != r2 {
			t.Errorf("same seed produced different results: %d vs %d", r1, r2)
		}
	}
}
