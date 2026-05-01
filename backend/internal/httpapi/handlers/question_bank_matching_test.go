package handlers

import (
	"testing"

	"atigacbt/backend/internal/repo/questionbankrepo"
)

func TestValidateMatching_AllowsEmptyLeftForDistractor(t *testing.T) {
	req := createQuestionReq{
		Type:    "matching",
		Stem:    "Cocokkan pasangan",
		OrderNo: 1,
		Pairs: []questionbankrepo.MatchingPair{
			{LeftContent: "2+2", RightContent: "4", OrderNo: 1},
			{LeftContent: "", RightContent: "5", OrderNo: 2}, // distractor
		},
	}

	in, err := validateAndBuildCreateQuestionInput(req)
	if err != nil {
		t.Fatalf("expected matching with empty left to be valid, got error: %v", err)
	}
	if got := len(in.MatchingPairs); got != 2 {
		t.Fatalf("expected 2 pairs, got %d", got)
	}
	if in.MatchingPairs[1].LeftContent != "" {
		t.Fatalf("expected second left_content empty, got %q", in.MatchingPairs[1].LeftContent)
	}
}

func TestValidateMatching_RejectsEmptyRight(t *testing.T) {
	req := createQuestionReq{
		Type:    "matching",
		Stem:    "Cocokkan pasangan",
		OrderNo: 1,
		Pairs: []questionbankrepo.MatchingPair{
			{LeftContent: "A", RightContent: "B", OrderNo: 1},
			{LeftContent: "", RightContent: "", OrderNo: 2},
		},
	}

	_, err := validateAndBuildCreateQuestionInput(req)
	if err == nil {
		t.Fatal("expected error for empty right_content, got nil")
	}
}
