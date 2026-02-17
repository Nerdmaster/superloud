package util

import "testing"

func TestUserHash(t *testing.T) {
	// Same input should always produce the same hash
	h1 := UserHash("123456789")
	h2 := UserHash("123456789")
	if h1 != h2 {
		t.Errorf("same input produced different hashes: %d vs %d", h1, h2)
	}

	// Different inputs should produce different hashes
	h3 := UserHash("987654321")
	if h1 == h3 {
		t.Errorf("different inputs produced same hash: %d", h1)
	}
}
