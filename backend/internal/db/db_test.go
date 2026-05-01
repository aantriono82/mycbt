package db

import "testing"

func TestGetEnvInt(t *testing.T) {
	t.Setenv("DB_TEST_INT", "42")
	if got := getEnvInt("DB_TEST_INT", 7); got != 42 {
		t.Fatalf("expected env int 42, got %d", got)
	}

	t.Setenv("DB_TEST_BAD", "bad")
	if got := getEnvInt("DB_TEST_BAD", 7); got != 7 {
		t.Fatalf("expected fallback for bad value, got %d", got)
	}

	t.Setenv("DB_TEST_ZERO", "0")
	if got := getEnvInt("DB_TEST_ZERO", 7); got != 7 {
		t.Fatalf("expected fallback for non-positive value, got %d", got)
	}

	if got := getEnvInt("DB_TEST_MISSING", 7); got != 7 {
		t.Fatalf("expected fallback for missing key, got %d", got)
	}
}

func TestMinMax(t *testing.T) {
	if min(1, 2) != 1 || min(5, -1) != -1 {
		t.Fatal("min returned unexpected value")
	}
	if max(1, 2) != 2 || max(5, -1) != 5 {
		t.Fatal("max returned unexpected value")
	}
}
