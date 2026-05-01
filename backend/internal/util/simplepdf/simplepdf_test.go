package simplepdf

import (
	"strings"
	"testing"
)

func TestEscapePDFText(t *testing.T) {
	t.Parallel()

	got := escapePDFText(`a\(b)` + "\n\tc")
	if strings.Contains(got, "\n") || strings.Contains(got, "\t") {
		t.Fatalf("expected control chars escaped, got %q", got)
	}
	if !strings.Contains(got, `\\`) || !strings.Contains(got, `\(`) || !strings.Contains(got, `\)`) {
		t.Fatalf("expected slashes and parens escaped, got %q", got)
	}
}

func TestDoc_AddTextAddLineAndBytes(t *testing.T) {
	t.Parallel()

	doc := NewA4Landscape()
	doc.AddText(Text{X: 10, Y: 20, Size: 0, Body: "Hello"})
	doc.AddText(Text{X: 0, Y: 0, Size: 12, Body: "   "})
	doc.AddLine(1, 2, 3, 4)
	doc.AddPage()
	doc.AddText(Text{X: 15, Y: 25, Size: 11, Body: "World"})

	pdf, err := doc.Bytes()
	if err != nil {
		t.Fatalf("Bytes error: %v", err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-1.4") {
		t.Fatalf("expected PDF header, got %q", string(pdf[:8]))
	}
	if !strings.Contains(string(pdf), "/Count 2") {
		t.Fatalf("expected 2 pages in output")
	}
	if !strings.Contains(string(pdf), "Hello") || !strings.Contains(string(pdf), "World") {
		t.Fatalf("expected text content in PDF output")
	}
}
