package handlers

import (
	"strings"
	"testing"
)

func TestParseQuestionsFromPlainText_MCSingleWithKey(t *testing.T) {
	input := `
1. Ibu kota Indonesia adalah?
A. Jakarta
B. Bandung
Kunci: A
`

	qs, warnings := parseQuestionsFromPlainText(input)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(qs) != 1 {
		t.Fatalf("expected 1 question, got %d", len(qs))
	}
	q := qs[0]
	if q.Type != "mc_single" {
		t.Fatalf("expected type mc_single, got %s", q.Type)
	}
	if len(q.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(q.Options))
	}
	if !q.Options[0].IsCorrect {
		t.Fatalf("expected option A to be correct")
	}
	if q.Options[1].IsCorrect {
		t.Fatalf("expected option B to be incorrect")
	}
}

func TestParseQuestionsFromPlainText_MultipleTypes(t *testing.T) {
	input := `
1) Pernyataan ini benar [true_false]
Answer: benar

2) Sebutkan mamalia [short_answer]
Answer: kucing | sapi

3) Pasangkan angka [matching]
1 => satu
2 => dua
`

	qs, warnings := parseQuestionsFromPlainText(input)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(qs) != 3 {
		t.Fatalf("expected 3 questions, got %d", len(qs))
	}

	if qs[0].Type != "true_false" || qs[0].TrueFalse == nil || !qs[0].TrueFalse.Correct {
		t.Fatalf("unexpected true_false parse result: %+v", qs[0])
	}
	if qs[1].Type != "short_answer" || len(qs[1].ShortAnswers) != 2 {
		t.Fatalf("unexpected short_answer parse result: %+v", qs[1])
	}
	if qs[2].Type != "matching" || len(qs[2].MatchingPairs) != 2 {
		t.Fatalf("unexpected matching parse result: %+v", qs[2])
	}
}

func TestParseQuestionsFromPlainText_DefaultEssayAndMissingKeyWarning(t *testing.T) {
	input := `
1. Soal pilihan tanpa kunci
A. Pilihan A
B. Pilihan B

2. Jelaskan proses fotosintesis.
`

	qs, warnings := parseQuestionsFromPlainText(input)
	if len(qs) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(qs))
	}
	if qs[0].Type != "mc_single" {
		t.Fatalf("expected first question type mc_single, got %s", qs[0].Type)
	}
	if qs[1].Type != "essay" {
		t.Fatalf("expected second question type essay, got %s", qs[1].Type)
	}
	if qs[1].Essay == nil {
		t.Fatalf("expected essay payload to be initialized")
	}

	warnText := strings.Join(warnings, " | ")
	if !strings.Contains(warnText, "no correct option marked") {
		t.Fatalf("expected warning for missing option key, got %v", warnings)
	}
	if !strings.Contains(warnText, "defaulted to essay") {
		t.Fatalf("expected warning for default essay type, got %v", warnings)
	}
}

func TestExtractTextFromDocumentXML_SynthesizesListMarkersForParser(t *testing.T) {
	// Simulates a DOCX where numbering/bullets are stored in paragraph properties (w:numPr),
	// not as literal "1."/"A." text nodes.
	docXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="7"/></w:numPr></w:pPr>
      <w:r><w:t>Ibu kota Indonesia adalah?</w:t></w:r>
    </w:p>
    <w:p>
      <w:pPr><w:numPr><w:ilvl w:val="1"/><w:numId w:val="7"/></w:numPr></w:pPr>
      <w:r><w:t>Jakarta</w:t></w:r>
    </w:p>
    <w:p>
      <w:pPr><w:numPr><w:ilvl w:val="1"/><w:numId w:val="7"/></w:numPr></w:pPr>
      <w:r><w:t>Bandung</w:t></w:r>
    </w:p>
    <w:p>
      <w:r><w:t>Kunci: A</w:t></w:r>
    </w:p>
  </w:body>
</w:document>`

	text, err := extractTextFromDocumentXML([]byte(docXML))
	if err != nil {
		t.Fatalf("extractTextFromDocumentXML error: %v", err)
	}

	qs, warnings := parseQuestionsFromPlainText(text)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v (text=%q)", warnings, text)
	}
	if len(qs) != 1 {
		t.Fatalf("expected 1 question, got %d (text=%q)", len(qs), text)
	}
	q := qs[0]
	if q.Type != "mc_single" {
		t.Fatalf("expected type mc_single, got %s (text=%q)", q.Type, text)
	}
	if len(q.Options) != 2 {
		t.Fatalf("expected 2 options, got %d (text=%q)", len(q.Options), text)
	}
	if q.Options[0].Label != "A" || q.Options[1].Label != "B" {
		t.Fatalf("expected labels A/B, got %+v (text=%q)", q.Options, text)
	}
	if !q.Options[0].IsCorrect {
		t.Fatalf("expected option A to be correct (text=%q)", text)
	}
	if q.Options[1].IsCorrect {
		t.Fatalf("expected option B to be incorrect (text=%q)", text)
	}
}
