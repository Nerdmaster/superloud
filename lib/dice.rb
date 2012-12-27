# Simple module for simulating die rolls.  With the API here, it would be
# trivial to swap out a more powerful random number generator.
module Dice
  # Seed RNG for reproducible results.  Or don't.  Returns prior seed.
  def self.seed(seed)
    return srand(seed)
  end

  # Rolls num sides-sided dice.  So for a 2d6 roll, you'd call Dice.roll(2, 6)
  def self.roll(num, sides)
    total = 0
    num.downto(1) { total += rand(sides) + 1 }

    return total
  end
end
