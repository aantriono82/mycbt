package studentexamrepo

import "testing"

func TestNormalizeScoringMode(t *testing.T) {
	if got := normalizeScoringMode(" absolute "); got != "absolute" {
		t.Fatalf("expected absolute, got %q", got)
	}
	if got := normalizeScoringMode("PARTIAL"); got != "partial" {
		t.Fatalf("expected partial, got %q", got)
	}
	if got := normalizeScoringMode("weird-mode"); got != "partial" {
		t.Fatalf("expected unknown mode fallback to partial, got %q", got)
	}
}

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

func TestScoreMCMultipleAbsolute(t *testing.T) {
	correct := map[string]bool{"a": true, "c": true}

	score, detail := scoreMCMultiple([]byte(`{"selected_option_ids":["a","c"]}`), correct, "absolute")
	if score != 100 || !detail.FullyCorrect {
		t.Fatalf("expected full absolute score, got score=%d detail=%+v", score, detail)
	}

	score2, detail2 := scoreMCMultiple([]byte(`{"selected_option_ids":["a"]}`), correct, "absolute")
	if score2 != 0 || detail2.FullyCorrect {
		t.Fatalf("expected incomplete absolute score 0, got score=%d detail=%+v", score2, detail2)
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

func TestScoreTrueFalseAbsoluteAndEmptyKeyset(t *testing.T) {
	keys := map[string]bool{"s1": true, "s2": false}
	score, detail := scoreTrueFalse([]byte(`{"values":{"s1":true}}`), keys, "absolute")
	if score != 0 || detail.FullyCorrect {
		t.Fatalf("expected non-full absolute score 0, got score=%d detail=%+v", score, detail)
	}

	score2, detail2 := scoreTrueFalse([]byte(`{}`), map[string]bool{}, "partial")
	if score2 != 0 || detail2.WrongCount != 0 {
		t.Fatalf("expected empty keyset score 0, got score=%d detail=%+v", score2, detail2)
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

func TestScoreMatchingAbsoluteAndEmptyPairs(t *testing.T) {
	pairs := []string{"p1", "p2"}

	score, detail := scoreMatching([]byte(`{"pairs":{"p1:L":"p1:R","p2:L":"p2:R"}}`), pairs, "absolute")
	if score != 100 || !detail.FullyCorrect {
		t.Fatalf("expected full absolute matching score, got score=%d detail=%+v", score, detail)
	}

	score2, detail2 := scoreMatching([]byte(`{"pairs":{"p1:L":"p1:R"}}`), pairs, "absolute")
	if score2 != 0 || detail2.FullyCorrect {
		t.Fatalf("expected incomplete absolute matching score 0, got score=%d detail=%+v", score2, detail2)
	}

	score3, detail3 := scoreMatching([]byte(`{}`), nil, "partial")
	if score3 != 0 || detail3.MaxScore != 100 {
		t.Fatalf("expected empty pair set score 0, got score=%d detail=%+v", score3, detail3)
	}
}

func TestNormalizeText(t *testing.T) {
	got := normalizeText("  Jakarta\t Pusat \n")
	if got != "jakarta pusat" {
		t.Fatalf("unexpected normalize result: %q", got)
	}
}

func TestResolveQuestionWeightsEven(t *testing.T) {
	qs := []qinfo{
		{ID: "q1", Weight: 1},
		{ID: "q2", Weight: 1},
		{ID: "q3", Weight: 1},
		{ID: "q4", Weight: 1},
	}
	w := resolveQuestionWeights(qs)
	if len(w) != 4 {
		t.Fatalf("unexpected weight size: %d", len(w))
	}
	if w["q1"] != 2500 || w["q2"] != 2500 || w["q3"] != 2500 || w["q4"] != 2500 {
		t.Fatalf("unexpected even distribution: %+v", w)
	}
}

func TestResolveQuestionWeightsOdd(t *testing.T) {
	qs := []qinfo{
		{ID: "q1", Weight: 1},
		{ID: "q2", Weight: 1},
		{ID: "q3", Weight: 1},
	}
	w := resolveQuestionWeights(qs)
	if len(w) != 3 {
		t.Fatalf("unexpected weight size: %d", len(w))
	}
	total := int(w["q1"] + w["q2"] + w["q3"])
	if total != 10000 {
		t.Fatalf("odd distribution must sum to 10000, got %d", total)
	}
	if !(w["q1"] == 3334 && w["q2"] == 3333 && w["q3"] == 3333) {
		t.Fatalf("unexpected odd distribution: %+v", w)
	}
}

func TestResolveQuestionWeightsCustom(t *testing.T) {
	qs := []qinfo{
		{ID: "q1", Weight: 2},
		{ID: "q2", Weight: 3},
		{ID: "q3", Weight: 5},
	}
	w := resolveQuestionWeights(qs)
	if w["q1"] != 2 || w["q2"] != 3 || w["q3"] != 5 {
		t.Fatalf("custom weight should be preserved, got %+v", w)
	}
}

func TestResolveQuestionWeights_NonPositiveCustomFallback(t *testing.T) {
	qs := []qinfo{
		{ID: "q1", Weight: 0},
		{ID: "q2", Weight: -2},
		{ID: "q3", Weight: 4},
	}
	w := resolveQuestionWeights(qs)
	if w["q1"] != 1 || w["q2"] != 1 || w["q3"] != 4 {
		t.Fatalf("non-positive custom weight should fallback to 1, got %+v", w)
	}
}
