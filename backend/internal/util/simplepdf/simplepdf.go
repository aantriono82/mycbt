package simplepdf

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	a4LandscapeWidth  = 842.0
	a4LandscapeHeight = 595.0
)

type Text struct {
	X    float64
	Y    float64
	Size int
	Body string
}

type Doc struct {
	pages [][]string
}

func NewA4Landscape() *Doc {
	return &Doc{pages: [][]string{make([]string, 0, 128)}}
}

func (d *Doc) AddPage() {
	d.pages = append(d.pages, make([]string, 0, 128))
}

func (d *Doc) AddText(t Text) {
	if strings.TrimSpace(t.Body) == "" {
		return
	}
	if t.Size <= 0 {
		t.Size = 10
	}
	escaped := escapePDFText(t.Body)
	page := d.ensureCurrentPage()
	*page = append(*page,
		fmt.Sprintf("BT /F1 %d Tf %.2f %.2f Td (%s) Tj ET", t.Size, t.X, t.Y, escaped),
	)
}

func (d *Doc) AddLine(x1, y1, x2, y2 float64) {
	page := d.ensureCurrentPage()
	*page = append(*page, fmt.Sprintf("%.2f %.2f m %.2f %.2f l S", x1, y1, x2, y2))
}

func (d *Doc) Bytes() ([]byte, error) {
	if len(d.pages) == 0 {
		d.pages = [][]string{make([]string, 0, 1)}
	}

	pageContents := make([][]byte, 0, len(d.pages))
	for _, pageLines := range d.pages {
		var content bytes.Buffer
		content.WriteString("0.5 w\n")
		for _, line := range pageLines {
			content.WriteString(line)
			content.WriteByte('\n')
		}
		pageContents = append(pageContents, content.Bytes())
	}

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")

	pageCount := len(pageContents)
	totalObjects := 3 + 2*pageCount // 1 catalog + 1 pages + N page + 1 font + N content
	offsets := make([]int, totalObjects+1)
	writeObj := func(id int, body string) {
		offsets[id] = out.Len()
		out.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", id, body))
	}

	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	pageObjectStart := 3
	fontObjectID := pageObjectStart + pageCount
	contentObjectStart := fontObjectID + 1

	kids := make([]string, 0, pageCount)
	for i := 0; i < pageCount; i++ {
		kids = append(kids, fmt.Sprintf("%d 0 R", pageObjectStart+i))
	}
	writeObj(2, fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), pageCount))

	for i := 0; i < pageCount; i++ {
		pageObjID := pageObjectStart + i
		contentObjID := contentObjectStart + i
		writeObj(pageObjID, fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>", a4LandscapeWidth, a4LandscapeHeight, fontObjectID, contentObjID))
	}
	writeObj(fontObjectID, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	for i := 0; i < pageCount; i++ {
		contentObjID := contentObjectStart + i
		contentBytes := pageContents[i]
		writeObj(contentObjID, fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(contentBytes), contentBytes))
	}

	xrefPos := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n", totalObjects+1))
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i <= totalObjects; i++ {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", totalObjects+1))
	out.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF", xrefPos))
	return out.Bytes(), nil
}

func (d *Doc) ensureCurrentPage() *[]string {
	if len(d.pages) == 0 {
		d.pages = [][]string{make([]string, 0, 128)}
	}
	return &d.pages[len(d.pages)-1]
}

func escapePDFText(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"(", "\\(",
		")", "\\)",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	)
	return replacer.Replace(s)
}
