package examrepo

import (
	"strings"
	"testing"
)

func TestNormalizeTokenLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   int
		want int
	}{
		{in: 0, want: 6},
		{in: -5, want: 6},
		{in: 1, want: 4},
		{in: 4, want: 4},
		{in: 8, want: 8},
		{in: 20, want: 12},
	}

	for _, tt := range tests {
		if got := normalizeTokenLength(tt.in); got != tt.want {
			t.Fatalf("normalizeTokenLength(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestRandToken_UsesAllowedAlphabetAndLength(t *testing.T) {
	t.Parallel()

	token, err := randToken(10)
	if err != nil {
		t.Fatalf("randToken error: %v", err)
	}
	if len(token) != 10 {
		t.Fatalf("expected token length 10, got %d", len(token))
	}
	for _, ch := range token {
		if !strings.ContainsRune(tokenAlphabet, ch) {
			t.Fatalf("unexpected rune %q in token %q", ch, token)
		}
	}
}
