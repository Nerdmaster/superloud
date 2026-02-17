package loudbot

import (
	"regexp"
	"strings"
	"unicode"
)

// LoudStatus represents the result of checking if a message is loud.
type LoudStatus int

const (
	StatusBad      LoudStatus = iota // Contains lowercase or too short
	StatusRejected                   // Loud enough to respond but not store
	StatusLoud                       // Loud enough to respond AND store
)

var retardRegex = regexp.MustCompile(`(?i)re+tard`)

// IsItLoud checks if the given text meets the criteria for being "loud".
// Returns the status and a reason string.
func IsItLoud(text string) (LoudStatus, string) {
	// No lowercase letters allowed
	for _, r := range text {
		if unicode.IsLower(r) {
			return StatusBad, "includes lowercase letters"
		}
	}

	length := len(text)
	if length < 11 {
		return StatusBad, "too short"
	}

	// Count uppercase letters
	uppercaseCount := 0
	for _, r := range text {
		if unicode.IsUpper(r) {
			uppercaseCount++
		}
	}

	if float64(uppercaseCount) < float64(length)*0.60 {
		return StatusBad, "too low uppercase ratio"
	}

	if retardRegex.MatchString(text) {
		return StatusBad, "shut up"
	}

	// Split into words (sequences of uppercase letters and apostrophes)
	wordRegex := regexp.MustCompile(`[A-Z']+`)
	allWords := wordRegex.FindAllString(text, -1)

	// Count all letters (uppercase only at this point since we rejected lowercase)
	letters := 0
	for _, r := range text {
		if unicode.IsUpper(r) {
			letters++
		}
	}

	// Filter out nonsense words
	var words []string
	for _, word := range allWords {
		// Single-letter words (I, A) are always valid
		if len(word) == 1 {
			words = append(words, word)
			continue
		}

		// No vowels? NO WORD!
		if !containsVowel(word) {
			continue
		}

		// Vowels only? NO WORD!
		if isAllVowels(word) {
			continue
		}

		words = append(words, word)
	}

	// Fewer than 2 unique words or half the words aren't unique
	unique := uniqueStrings(words)
	if len(unique) < 2 || len(unique) <= len(words)/2 {
		return StatusRejected, "too few unique words"
	}

	// Average letters per word must be at least 3
	if len(words) > 0 && letters/len(words) < 3 {
		return StatusRejected, "words are too small"
	}

	// No periods at the end
	if strings.HasSuffix(text, ".") {
		return StatusRejected, "friggin' periods"
	}

	return StatusLoud, "loud"
}

func containsVowel(word string) bool {
	for _, r := range word {
		switch r {
		case 'A', 'E', 'I', 'O', 'U', 'Y':
			return true
		}
	}
	return false
}

func isAllVowels(word string) bool {
	for _, r := range word {
		switch r {
		case 'A', 'E', 'I', 'O', 'U':
			continue
		default:
			return false
		}
	}
	return true
}

func uniqueStrings(ss []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
