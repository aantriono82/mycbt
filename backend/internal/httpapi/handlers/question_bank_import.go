package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/repo/questionbankrepo"
)

const maxDocxBytes = 20 << 20 // 20MB

func (h *QuestionBankHandler) ImportDocxPreview(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	setID := c.Param("id")
	set, ok, err := h.qb.GetSet(c.Request.Context(), setID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || set.OwnerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "file required"}})
		return
	}
	if fh.Size <= 0 || fh.Size > maxDocxBytes {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "file too large"}})
		return
	}
	f, err := fh.Open()
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "cannot open file"}})
		return
	}
	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(io.LimitReader(f, maxDocxBytes+1))
	if err != nil || int64(len(b)) > maxDocxBytes {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "cannot read file"}})
		return
	}

	text, err := extractTextFromDOCX(b)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid docx"}})
		return
	}

	qs, warnings := parseQuestionsFromPlainText(text)
	c.JSON(200, gin.H{
		"data": gin.H{
			"questions": qs,
			"warnings":  warnings,
		},
		"meta": gin.H{
			"question_set_id": setID,
			"count":           len(qs),
		},
	})
}

func (h *QuestionBankHandler) ImportDocx(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	setID := c.Param("id")
	set, ok, err := h.qb.GetSet(c.Request.Context(), setID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || set.OwnerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "file required"}})
		return
	}
	if fh.Size <= 0 || fh.Size > maxDocxBytes {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "file too large"}})
		return
	}
	f, err := fh.Open()
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "cannot open file"}})
		return
	}
	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(io.LimitReader(f, maxDocxBytes+1))
	if err != nil || int64(len(b)) > maxDocxBytes {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "cannot read file"}})
		return
	}
	text, err := extractTextFromDOCX(b)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid docx"}})
		return
	}

	qs, warnings := parseQuestionsFromPlainText(text)
	if len(qs) == 0 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "no questions detected"}})
		return
	}

	existing, err := h.qb.ListQuestions(c.Request.Context(), setID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	maxOrderNo := 0
	for _, item := range existing {
		if item.OrderNo > maxOrderNo {
			maxOrderNo = item.OrderNo
		}
	}

	ins := make([]questionbankrepo.CreateQuestionInput, 0, len(qs))
	for idx, q := range qs {
		cReq := createQuestionReq{
			Type:       q.Type,
			Stem:       q.Stem,
			OrderNo:    maxOrderNo + idx + 1,
			Options:    q.Options,
			Pairs:      q.MatchingPairs,
			Answers:    q.ShortAnswers,
			Correct:    nil,
			RubricText: "",
			MaxScore:   nil,
		}
		if q.TrueFalse != nil {
			cReq.Correct = &q.TrueFalse.Correct
		}
		if q.Essay != nil {
			cReq.RubricText = q.Essay.RubricText
			cReq.MaxScore = q.Essay.MaxScore
		}

		in, err := validateAndBuildCreateQuestionInput(cReq)
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": fmt.Sprintf("invalid question order_no=%d: %s", q.OrderNo, err.Error())}})
			return
		}
		ins = append(ins, in)
	}

	created, err := h.qb.CreateQuestionsBulk(c.Request.Context(), setID, ins)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{
		"data": gin.H{
			"questions": created,
			"warnings":  warnings,
		},
		"meta": gin.H{
			"question_set_id": setID,
			"count":           len(created),
		},
	})
}

func extractTextFromDOCX(b []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", err
	}
	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return "", fmt.Errorf("missing document.xml")
	}
	rc, err := docFile.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = rc.Close() }()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	return extractTextFromDocumentXML(raw)
}

// extractTextFromDocumentXML extracts paragraph text from word/document.xml.
//
// Important: In DOCX, list numbering (e.g. "1.", "A.") is usually not part of w:t text;
// it's encoded in paragraph properties (w:numPr). Our plain-text parser expects explicit
// markers, so we synthesize "N." for ilvl=0 and "A." for ilvl=1 to make the import
// robust for typical "question list + options list" authoring in MS Word.
func extractTextFromDocumentXML(raw []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(raw))

	var (
		inText      bool
		inParagraph int
		inNumPr     bool

		sb       strings.Builder
		out      []string
		hasNumPr bool
		ilvl     = -1

		qNo       int
		nextOpt   rune = 'A'
		gotQStart bool
	)

	resetPara := func() {
		sb.Reset()
		hasNumPr = false
		ilvl = -1
		inText = false
		inNumPr = false
	}

	flushPara := func() {
		s := strings.TrimSpace(sb.String())
		if s == "" {
			return
		}

		line := s
		if hasNumPr {
			switch ilvl {
			case 0:
				// If the number is already present as literal text, don't double-prefix.
				if reQuestionStart.MatchString(s) {
					gotQStart = true
					nextOpt = 'A'
					break
				}
				qNo++
				nextOpt = 'A'
				gotQStart = true
				line = fmt.Sprintf("%d. %s", qNo, s)
			case 1:
				// Only synthesize options if we already saw a question start.
				// Avoid prefixing lines like "Kunci: A" when they happen to be inside a list.
				if reOption.MatchString(s) {
					break
				}
				lt := strings.ToLower(strings.TrimSpace(s))
				if gotQStart && nextOpt <= 'F' && !strings.HasPrefix(lt, "kunci:") && !strings.HasPrefix(lt, "answer:") {
					line = fmt.Sprintf("%c. %s", nextOpt, s)
					nextOpt++
				}
			}
		}

		out = append(out, line)
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				inParagraph++
				if inParagraph == 1 {
					resetPara()
				}
			case "t":
				if inParagraph > 0 {
					inText = true
				}
			case "tab":
				if inParagraph > 0 {
					sb.WriteByte('\t')
				}
			case "br":
				if inParagraph > 0 {
					sb.WriteByte('\n')
				}
			case "numPr":
				if inParagraph > 0 {
					inNumPr = true
					hasNumPr = true
				}
			case "ilvl":
				if inParagraph > 0 && inNumPr {
					for _, a := range t.Attr {
						if a.Name.Local == "val" {
							if v, err := strconv.Atoi(strings.TrimSpace(a.Value)); err == nil {
								ilvl = v
							}
						}
					}
				}
			case "numId":
				if inParagraph > 0 && inNumPr {
					// Presence is enough for our heuristic.
					hasNumPr = true
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "numPr":
				inNumPr = false
			case "p":
				if inParagraph > 0 {
					inParagraph--
					if inParagraph == 0 {
						flushPara()
					}
				}
			}
		case xml.CharData:
			if inParagraph > 0 && inText {
				sb.Write([]byte(t))
			}
		}
	}

	return strings.Join(out, "\n"), nil
}

var reQuestionStart = regexp.MustCompile(`^\s*(\d+)\s*[\.\)]\s*(.*)$`)
var reOption = regexp.MustCompile(`^\s*([A-Fa-f])\s*[\.\)]\s*(.+)$`)

func parseQuestionsFromPlainText(text string) ([]questionbankrepo.Question, []string) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := []questionbankrepo.Question{}
	warnings := []string{}

	var cur *questionbankrepo.Question
	flush := func() {
		if cur == nil {
			return
		}
		cur.Stem = strings.TrimSpace(cur.Stem)
		if cur.Type == "" {
			// Heuristic: if options exist => mc_single, else essay.
			if len(cur.Options) > 0 {
				cur.Type = "mc_single"
				warnings = append(warnings, fmt.Sprintf("Q%d: type missing, defaulted to mc_single", cur.OrderNo))
			} else {
				cur.Type = "essay"
				warnings = append(warnings, fmt.Sprintf("Q%d: type missing, defaulted to essay", cur.OrderNo))
			}
		}
		if (cur.Type == "mc_single" || cur.Type == "mc_multiple") && len(cur.Options) > 0 {
			correct := 0
			for _, o := range cur.Options {
				if o.IsCorrect {
					correct++
				}
			}
			if correct == 0 {
				warnings = append(warnings, fmt.Sprintf("Q%d: no correct option marked (missing Answer:)", cur.OrderNo))
			}
		}
		out = append(out, *cur)
		cur = nil
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		if m := reQuestionStart.FindStringSubmatch(line); m != nil {
			flush()
			no := len(out) + 1
			cur = &questionbankrepo.Question{
				Type:          "",
				Stem:          strings.TrimSpace(m[2]),
				OrderNo:       no,
				Options:       []questionbankrepo.QuestionOption{},
				MatchingPairs: []questionbankrepo.MatchingPair{},
				ShortAnswers:  []questionbankrepo.ShortAnswer{},
			}
			// Inline type tag: [mc_single]
			if strings.Contains(cur.Stem, "[") && strings.Contains(cur.Stem, "]") {
				if i := strings.Index(cur.Stem, "["); i >= 0 {
					if j := strings.Index(cur.Stem[i:], "]"); j >= 0 {
						tag := strings.ToLower(strings.TrimSpace(cur.Stem[i+1 : i+j]))
						if isSupportedType(tag) {
							cur.Type = tag
							cur.Stem = strings.TrimSpace(strings.TrimSpace(cur.Stem[:i]) + " " + strings.TrimSpace(cur.Stem[i+j+1:]))
						}
					}
				}
			}
			continue
		}

		if cur == nil {
			// Ignore preface text
			continue
		}

		lt := strings.ToLower(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			tag := strings.ToLower(strings.TrimSpace(strings.Trim(line, "[]")))
			if isSupportedType(tag) {
				cur.Type = tag
				continue
			}
		}

		if m := reOption.FindStringSubmatch(line); m != nil {
			lbl := strings.ToUpper(strings.TrimSpace(m[1]))
			content := strings.TrimSpace(m[2])
			cur.Options = append(cur.Options, questionbankrepo.QuestionOption{Label: lbl, Content: content})
			if cur.Type == "" {
				cur.Type = "mc_single"
			}
			continue
		}

		if strings.HasPrefix(lt, "answer:") || strings.HasPrefix(lt, "kunci:") {
			ans := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			if ans != "" {
				applyAnswerToken(cur, ans, &warnings)
			}
			continue
		}

		// Matching pair convention: "left => right" lines
		if strings.Contains(line, "=>") {
			parts := strings.SplitN(line, "=>", 2)
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			if left != "" && right != "" {
				cur.MatchingPairs = append(cur.MatchingPairs, questionbankrepo.MatchingPair{LeftContent: left, RightContent: right, OrderNo: len(cur.MatchingPairs) + 1})
				if cur.Type == "" {
					cur.Type = "matching"
				}
				continue
			}
		}

		if cur.Stem == "" {
			cur.Stem = line
		} else {
			cur.Stem += "\n" + line
		}
	}
	flush()

	// Post-process: normalize order_no sequential and default essay payload.
	for i := range out {
		out[i].OrderNo = i + 1
		out[i].Options = normalizeOptionLabels(out[i].Options)
		if out[i].Type == "essay" && out[i].Essay == nil {
			out[i].Essay = &questionbankrepo.Essay{}
		}
		if out[i].Type == "true_false" && out[i].TrueFalse == nil {
			warnings = append(warnings, fmt.Sprintf("Q%d: true_false missing correct (Answer:)", out[i].OrderNo))
		}
		if out[i].Type == "matching" && len(out[i].MatchingPairs) < 2 {
			warnings = append(warnings, fmt.Sprintf("Q%d: matching pairs less than 2", out[i].OrderNo))
		}
		if out[i].Type == "short_answer" && len(out[i].ShortAnswers) < 1 {
			warnings = append(warnings, fmt.Sprintf("Q%d: short_answer missing answers (Answer:)", out[i].OrderNo))
		}
	}

	return out, warnings
}

func isSupportedType(t string) bool {
	switch t {
	case "mc_single", "mc_multiple", "matching", "short_answer", "essay", "true_false":
		return true
	default:
		return false
	}
}

func normalizeOptionLabels(opts []questionbankrepo.QuestionOption) []questionbankrepo.QuestionOption {
	seen := map[string]bool{}
	out := make([]questionbankrepo.QuestionOption, 0, len(opts))
	for _, o := range opts {
		lbl := strings.ToUpper(strings.TrimSpace(o.Label))
		if lbl == "" {
			continue
		}
		if seen[lbl] {
			continue
		}
		seen[lbl] = true
		o.Label = lbl
		o.Content = strings.TrimSpace(o.Content)
		out = append(out, o)
	}
	return out
}

func applyAnswerToken(q *questionbankrepo.Question, ans string, warnings *[]string) {
	u := strings.ToUpper(strings.TrimSpace(ans))
	// Option answer: "A" or "A,C"
	if reOptionAnswer(u) {
		parts := qbSplitCSVUpper(u)
		for i := range q.Options {
			q.Options[i].IsCorrect = false
		}
		for _, p := range parts {
			for i := range q.Options {
				if q.Options[i].Label == p {
					q.Options[i].IsCorrect = true
				}
			}
		}
		if len(parts) > 1 {
			q.Type = "mc_multiple"
		} else if q.Type == "" {
			q.Type = "mc_single"
		}
		return
	}

	lt := strings.ToLower(strings.TrimSpace(ans))
	if lt == "true" || lt == "benar" || lt == "ya" || lt == "y" {
		q.Type = "true_false"
		q.TrueFalse = &questionbankrepo.TrueFalse{Correct: true}
		return
	}
	if lt == "false" || lt == "salah" || lt == "tidak" || lt == "n" {
		q.Type = "true_false"
		q.TrueFalse = &questionbankrepo.TrueFalse{Correct: false}
		return
	}

	// Short answer: "abc | def"
	q.Type = "short_answer"
	parts := strings.Split(ans, "|")
	q.ShortAnswers = []questionbankrepo.ShortAnswer{}
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		q.ShortAnswers = append(q.ShortAnswers, questionbankrepo.ShortAnswer{AnswerText: p, OrderNo: i + 1})
	}
	if len(q.ShortAnswers) == 0 {
		*warnings = append(*warnings, fmt.Sprintf("Q%d: Answer: present but empty", q.OrderNo))
	}
}

func reOptionAnswer(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r >= 'A' && r <= 'F' {
			continue
		}
		if r == ',' || r == ' ' {
			continue
		}
		return false
	}
	return true
}

func qbSplitCSVUpper(s string) []string {
	parts := strings.Split(s, ",")
	out := []string{}
	for _, p := range parts {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
