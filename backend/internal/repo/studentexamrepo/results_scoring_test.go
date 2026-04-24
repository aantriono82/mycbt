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

func TestScoreMCMultiplePartial(t *testing.T) {
	correct := map[string]bool{"a": true, "c": true}
	score, detail := scoreMCMultiple([]byte(`{"selected_option_ids":["a","d"]}`), correct, "partial")
	if score != 0 {
		t.Fatalf("expected mc_multiple partial score 0 (1 benar - 1 salah), got %d", score)
	}
	if detail.CorrectCount != 1 || detail.WrongCount != 2 {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	score2, _ := scoreMCMultiple([]byte(`{"selected_option_ids":["a"]}`), correct, "partial")
	if score2 != 50 {
		t.Fatalf("expected mc_multiple partial score 50, got %d", score2)
	}
}

func TestScoreTrueFalsePartial(t *testing.T) {
	keys := map[string]bool{"s1": true, "s2": false, "s3": true}
	score, detail := scoreTrueFalse([]byte(`{"values":{"s1":true,"s2":false}}`), keys, "partial")
	if score != 67 {
		t.Fatalf("expected true_false partial score 67, got %d", score)
	}
	if detail.CorrectCount != 2 || detail.WrongCount != 1 {
		t.Fatalf("unexpected detail: %+v", detail)
	}
}

func TestScoreMatchingPartial(t *testing.T) {
	pairs := []string{"p1", "p2", "p3"}
	score, detail := scoreMatching([]byte(`{"pairs":{"p1:L":"p1:R","p2:L":"wrong","p3:L":"p3:R"}}`), pairs, "partial")
	if score != 67 {
		t.Fatalf("expected matching partial score 67, got %d", score)
	}
	if detail.CorrectCount != 2 || detail.WrongCount != 1 {
		t.Fatalf("unexpected detail: %+v", detail)
	}
}

func TestNormalizeText(t *testing.T) {
	got := normalizeText("  Jakarta\t Pusat \n")
	if got != "jakarta pusat" {
		t.Fatalf("unexpected normalize result: %q", got)
	}
}
