package util

import "math/rand"

// DiceRoll rolls num dice each with the given number of sides.
// For example, DiceRoll(3, 15, rng) rolls 3d15.
func DiceRoll(num, sides int, rng *rand.Rand) int {
	total := 0
	for i := 0; i < num; i++ {
		total += rng.Intn(sides) + 1
	}
	return total
}
