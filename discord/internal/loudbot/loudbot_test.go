package loudbot

import "testing"

func TestIsItLoud(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected LoudStatus
	}{
		// Bad cases - should not trigger any response
		{"lowercase", "IT'S ONLY VALID IF ALL WORDS ARE UPPERCASE COMPLETELy", StatusBad},
		{"too short", "LOUD SHORT", StatusBad},
		{"low uppercase ratio", "THIS                                         ISN'T                                   VALID", StatusBad},

		// Rejected cases - trigger response but not stored
		{"no vowels", "TBBSSSDDDFFF FDDSSDJJKLLM FRTGBNMV", StatusRejected},
		{"vowels only", "AEIOUOUAEU AUIOEEIUO AUIOUA", StatusRejected},
		{"too many dupes", "BINARY BINARY BINARY BINARY BOO", StatusRejected},
		{"words too small", "IS IT VALID NO", StatusRejected},

		// Valid loud message
		{"valid", "THIS ISN'T FUNNY, BUT AT LEAST IT'S VALID!", StatusLoud},

		// Period at end
		{"ends with period", "THIS IS A VALID MESSAGE THAT ENDS WITH A PERIOD.", StatusRejected},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, reason := IsItLoud(tt.text)
			if status != tt.expected {
				t.Errorf("IsItLoud(%q) = %v (%s), want %v", tt.text, status, reason, tt.expected)
			}
		})
	}
}

func TestIsItLoud_RetardFilter(t *testing.T) {
	status, _ := IsItLoud("WHAT A RETARD HE IS LOLOLOL")
	if status != StatusBad {
		t.Errorf("expected StatusBad for retard filter, got %v", status)
	}

	// Also with extended e's
	status, _ = IsItLoud("WHAT A REEEETARD LOLOLOL")
	if status != StatusBad {
		t.Errorf("expected StatusBad for reeeetard filter, got %v", status)
	}
}
