package commands

import "testing"

func TestRollPenalty(t *testing.T) {
	tests := []struct {
		count    int
		expected int
	}{
		{0, 0},
		{1, -1},
		{2, -3},
		{3, -6},
		{4, -9},
		{5, -12},
		{14, -108},
		{100, -108}, // beyond array length, should return last
	}

	for _, tt := range tests {
		result := rollPenalty(tt.count)
		if result != tt.expected {
			t.Errorf("rollPenalty(%d) = %d, want %d", tt.count, result, tt.expected)
		}
	}
}

func TestPlaceText(t *testing.T) {
	tests := []struct {
		place    int
		expected string
	}{
		{1, "FIRST"},
		{2, "SECOND"},
		{10, "TENTH"},
		{11, "NUMBER 11"},
	}

	for _, tt := range tests {
		result := placeText(tt.place)
		if result != tt.expected {
			t.Errorf("placeText(%d) = %q, want %q", tt.place, result, tt.expected)
		}
	}
}

func TestUserlistText(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"ALICE"}, "ALICE"},
		{[]string{"BOB", "ALICE"}, "ALICE AND BOB"},
		{[]string{"CHARLIE", "BOB", "ALICE"}, "ALICE, BOB, AND CHARLIE"},
		{[]string{"DAVE", "CHARLIE", "BOB", "ALICE"}, "ALICE, BOB, CHARLIE, AND DAVE"},
	}

	for _, tt := range tests {
		result := userlistText(tt.input)
		if result != tt.expected {
			t.Errorf("userlistText(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMax(t *testing.T) {
	if max(2, 5) != 5 {
		t.Error("max(2, 5) should be 5")
	}
	if max(10, 3) != 10 {
		t.Error("max(10, 3) should be 10")
	}
}
