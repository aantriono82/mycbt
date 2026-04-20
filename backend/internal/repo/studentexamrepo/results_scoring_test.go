package studentexamrepo

import "testing"

func TestIsCorrectMCSingle(t *testing.T) {
	correct := map[string]bool{"opt-a": true}
	if !isCorrectMCSingle([]byte(`{"selected_option_id":"opt-a"}`), correct) {
		t.Fatalf("expected mc_single correct answer to pass")
	}
	if isCorrectMCSingle([]byte(`{"selected_option_id":"opt-b"}`), correct) {
		t.Fatalf("expected mc_single wrong answer to fail")
	}
}

func TestIsCorrectMCMultiple(t *testing.T) {
	correct := map[string]bool{"a": true, "c": true}
	if !isCorrectMCMultiple([]byte(`{"selected_option_ids":["c","a"]}`), correct) {
		t.Fatalf("expected exact mc_multiple match to pass")
	}
	if isCorrectMCMultiple([]byte(`{"selected_option_ids":["a"]}`), correct) {
		t.Fatalf("expected incomplete mc_multiple answer to fail")
	}
	if isCorrectMCMultiple([]byte(`{"selected_option_ids":["a","c","d"]}`), correct) {
		t.Fatalf("expected extra option mc_multiple answer to fail")
	}
}

func TestIsCorrectTrueFalse(t *testing.T) {
	if !isCorrectTrueFalse([]byte(`{"value":true}`), map[string]bool{"legacy": true}) {
		t.Fatalf("expected true_false correct to pass")
	}
	if isCorrectTrueFalse([]byte(`{"value":false}`), map[string]bool{"legacy": true}) {
		t.Fatalf("expected true_false wrong to fail")
	}

	// legacy correct=false should also be supported
	if !isCorrectTrueFalse([]byte(`{"value":false}`), map[string]bool{"legacy": false}) {
		t.Fatalf("expected true_false legacy correct=false to pass")
	}
}

func TestIsCorrectShortAnswer(t *testing.T) {
	acceptable := []string{"kucing", "sapi"}
	if !isCorrectShortAnswer([]byte(`{"text":"  KUcing "}`), acceptable) {
		t.Fatalf("expected short_answer normalized value to pass")
	}
	if isCorrectShortAnswer([]byte(`{"text":"anjing"}`), acceptable) {
		t.Fatalf("expected short_answer wrong value to fail")
	}
}

func TestIsCorrectMatching(t *testing.T) {
	pairs := []string{"p1", "p2"}
	if !isCorrectMatching([]byte(`{"pairs":{"p1:L":"p1:R","p2:L":"p2:R"}}`), pairs) {
		t.Fatalf("expected matching correct answer to pass")
	}
	if isCorrectMatching([]byte(`{"pairs":{"p1:L":"p1:R","p2:L":"wrong"}}`), pairs) {
		t.Fatalf("expected matching wrong answer to fail")
	}
}

func TestNormalizeText(t *testing.T) {
	got := normalizeText("  Jakarta\t Pusat \n")
	if got != "jakarta pusat" {
		t.Fatalf("unexpected normalize result: %q", got)
	}
}
