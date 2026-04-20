package studentexamrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type StudentResultSummary struct {
	SessionID string `json:"session_id"`
	ExamID    string `json:"exam_id"`

	ExamTitle     string `json:"exam_title"`
	SubjectName   string `json:"subject"`
	SessionStatus string `json:"session_status"`
	SubmittedAt   string `json:"submitted_at,omitempty"`

	TotalQuestions    int `json:"total_questions"`
	AnsweredQuestions int `json:"answered_questions"`
	AutoScorable      int `json:"auto_scorable_questions"`
	CorrectCount      int `json:"correct_count"`
	Score             int `json:"score"`
	ManualScoredCount int `json:"manual_scored_count"`
	PendingGradingCount int `json:"pending_grading_count"`
}

type ListStudentResultsFilter struct {
	Limit  int
	Offset int
}

func (r *Repo) ListStudentResults(ctx context.Context, studentID string, f ListStudentResultsFilter) ([]StudentResultSummary, int, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id::text,
       s.exam_id::text,
       e.title,
       sub.name,
       s.status,
       COALESCE(to_char(s.finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS finished_at
FROM exam_sessions s
JOIN exams e ON e.id = s.exam_id
JOIN subjects sub ON sub.id = e.subject_id
WHERE s.student_id = $1
  AND s.status <> 'in_progress'
ORDER BY s.finished_at DESC NULLS LAST, s.started_at DESC
LIMIT $2 OFFSET $3`, studentID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list results: %w", err)
	}
	defer rows.Close()

	out := []StudentResultSummary{}
	for rows.Next() {
		var it StudentResultSummary
		var finished string
		if err := rows.Scan(&it.SessionID, &it.ExamID, &it.ExamTitle, &it.SubjectName, &it.SessionStatus, &finished); err != nil {
			return nil, 0, fmt.Errorf("scan result: %w", err)
		}
		if finished != "" {
			it.SubmittedAt = finished
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM exam_sessions
WHERE student_id = $1 AND status <> 'in_progress'`, studentID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count results: %w", err)
	}

	for i := range out {
		sum, err := r.ComputeAutoScore(ctx, out[i].SessionID, studentID, time.Now().UTC())
		if err != nil {
			return nil, 0, err
		}
		out[i].TotalQuestions = sum.TotalQuestions
		out[i].AnsweredQuestions = sum.AnsweredQuestions
		out[i].AutoScorable = sum.AutoScorable
		out[i].CorrectCount = sum.CorrectCount
		out[i].Score = sum.Score
		out[i].ManualScoredCount = sum.ManualScored
		out[i].PendingGradingCount = sum.PendingGrading
	}

	return out, total, nil
}

type AutoScoreSummary struct {
	TotalQuestions    int `json:"total_questions"`
	AnsweredQuestions int `json:"answered_questions"`
	AutoScorable      int `json:"auto_scorable_questions"`
	CorrectCount      int `json:"correct_count"`
	Score             int `json:"score"`
	ManualScored      int `json:"manual_scored_count"`
	PendingGrading    int `json:"pending_grading_count"`
	TotalMaxScore     int `json:"total_max_score"`
	TotalActualScore  int `json:"total_actual_score"`
}

type qinfo struct {
	ID   string
	Type string
}

func (r *Repo) ComputeAutoScore(ctx context.Context, sessionID, studentID string, nowUTC time.Time) (AutoScoreSummary, error) {
	// Keep session status consistent.
	if _, err := r.expireSessionIfNeeded(ctx, sessionID, studentID, nowUTC); err != nil && err != ErrSessionNotFound {
		return AutoScoreSummary{}, err
	}

	qrows, err := r.pool.Query(ctx, `
SELECT q.id::text, q.type
FROM exam_session_questions sq
JOIN exam_sessions s ON s.id = sq.exam_session_id
JOIN questions q ON q.id = sq.question_id
WHERE sq.exam_session_id = $1 AND s.student_id = $2
ORDER BY sq.order_no ASC`, sessionID, studentID)
	if err != nil {
		return AutoScoreSummary{}, fmt.Errorf("load session questions: %w", err)
	}
	defer qrows.Close()

	qs := []qinfo{}
	for qrows.Next() {
		var it qinfo
		if err := qrows.Scan(&it.ID, &it.Type); err != nil {
			return AutoScoreSummary{}, fmt.Errorf("scan question: %w", err)
		}
		qs = append(qs, it)
	}
	if err := qrows.Err(); err != nil {
		return AutoScoreSummary{}, err
	}
	if len(qs) == 0 {
		return AutoScoreSummary{}, nil
	}

	return r.computeAutoScoreFromQuestions(ctx, sessionID, qs)
}

func (r *Repo) ComputeAutoScoreAny(ctx context.Context, sessionID string, nowUTC time.Time) (AutoScoreSummary, error) {
	if _, err := r.expireSessionIfNeededAny(ctx, sessionID, nowUTC); err != nil && err != ErrSessionNotFound {
		return AutoScoreSummary{}, err
	}

	qrows, err := r.pool.Query(ctx, `
SELECT q.id::text, q.type
FROM exam_session_questions sq
JOIN questions q ON q.id = sq.question_id
WHERE sq.exam_session_id = $1
ORDER BY sq.order_no ASC`, sessionID)
	if err != nil {
		return AutoScoreSummary{}, fmt.Errorf("load session questions: %w", err)
	}
	defer qrows.Close()

	qs := []qinfo{}
	for qrows.Next() {
		var it qinfo
		if err := qrows.Scan(&it.ID, &it.Type); err != nil {
			return AutoScoreSummary{}, fmt.Errorf("scan question: %w", err)
		}
		qs = append(qs, it)
	}
	if err := qrows.Err(); err != nil {
		return AutoScoreSummary{}, err
	}
	if len(qs) == 0 {
		return AutoScoreSummary{}, nil
	}

	return r.computeAutoScoreFromQuestions(ctx, sessionID, qs)
}

func (r *Repo) computeAutoScoreFromQuestions(ctx context.Context, sessionID string, qs []qinfo) (AutoScoreSummary, error) {
	attempts := map[string]struct {
		Answer      json.RawMessage
		ManualScore *int
	}{}
	arows, err := r.pool.Query(ctx, `
SELECT question_id::text, answer_json, manual_score
FROM exam_attempts
WHERE exam_session_id = $1`, sessionID)
	if err != nil {
		return AutoScoreSummary{}, fmt.Errorf("load attempts: %w", err)
	}
	defer arows.Close()
	for arows.Next() {
		var qid string
		var raw []byte
		var ms *int
		if err := arows.Scan(&qid, &raw, &ms); err != nil {
			return AutoScoreSummary{}, fmt.Errorf("scan attempt: %w", err)
		}
		attempts[qid] = struct {
			Answer      json.RawMessage
			ManualScore *int
		}{Answer: json.RawMessage(raw), ManualScore: ms}
	}
	if err := arows.Err(); err != nil {
		return AutoScoreSummary{}, err
	}

	total := len(qs)
	answered := 0
	autoTotal := 0
	correct := 0

	// Prepare correct keys lookup.
	mcIDs := []string{}
	tfIDs := []string{}
	saIDs := []string{}
	matchIDs := []string{}

	essayIDs := []string{}
	for _, q := range qs {
		if _, ok := attempts[q.ID]; ok {
			answered++
		}
		switch q.Type {
		case "mc_single", "mc_multiple":
			mcIDs = append(mcIDs, q.ID)
		case "true_false":
			tfIDs = append(tfIDs, q.ID)
		case "short_answer":
			saIDs = append(saIDs, q.ID)
		case "matching":
			matchIDs = append(matchIDs, q.ID)
		case "essay":
			essayIDs = append(essayIDs, q.ID)
		}
	}

	mcCorrect := map[string]map[string]bool{} // qid -> optionIDs correct
	if len(mcIDs) > 0 {
		rows, err := r.pool.Query(ctx, `
SELECT question_id::text, id::text, is_correct
FROM question_options
WHERE question_id::text = ANY($1::text[])`, mcIDs)
		if err != nil {
			return AutoScoreSummary{}, fmt.Errorf("load mc keys: %w", err)
		}
		for rows.Next() {
			var qid, oid string
			var isCorrect bool
			if err := rows.Scan(&qid, &oid, &isCorrect); err != nil {
				rows.Close()
				return AutoScoreSummary{}, fmt.Errorf("scan mc key: %w", err)
			}
			m, ok := mcCorrect[qid]
			if !ok {
				m = map[string]bool{}
				mcCorrect[qid] = m
			}
			if isCorrect {
				m[oid] = true
			}
		}
		rows.Close()
	}

	tfCorrect := map[string]map[string]bool{} // qid -> stID -> correctValue
	if len(tfIDs) > 0 {
		rows, err := r.pool.Query(ctx, `
SELECT question_id::text, id, correct
FROM question_true_false_statements
WHERE question_id::text = ANY($1::text[])`, tfIDs)
		if err != nil {
			return AutoScoreSummary{}, fmt.Errorf("load tf keys: %w", err)
		}
		for rows.Next() {
			var qid, stid string
			var v bool
			if err := rows.Scan(&qid, &stid, &v); err != nil {
				rows.Close()
				return AutoScoreSummary{}, fmt.Errorf("scan tf key: %w", err)
			}
			m, ok := tfCorrect[qid]
			if !ok {
				m = map[string]bool{}
				tfCorrect[qid] = m
			}
			m[stid] = v
		}
		rows.Close()

		// FALLBACK: If a question has no statements in the new table, check the old table
		// This handles legacy data that wasn't migrated.
		for _, qid := range tfIDs {
			if len(tfCorrect[qid]) == 0 {
				var v bool
				err := r.pool.QueryRow(ctx, `SELECT correct FROM question_true_false WHERE question_id = $1`, qid).Scan(&v)
				if err == nil {
					tfCorrect[qid] = map[string]bool{"legacy": v}
				}
			}
		}
	}

	saCorrect := map[string][]string{}
	if len(saIDs) > 0 {
		rows, err := r.pool.Query(ctx, `
SELECT question_id::text, answer_text
FROM question_short_answers
WHERE question_id::text = ANY($1::text[])`, saIDs)
		if err != nil {
			return AutoScoreSummary{}, fmt.Errorf("load short_answer keys: %w", err)
		}
		for rows.Next() {
			var qid, ans string
			if err := rows.Scan(&qid, &ans); err != nil {
				rows.Close()
				return AutoScoreSummary{}, fmt.Errorf("scan short_answer key: %w", err)
			}
			saCorrect[qid] = append(saCorrect[qid], normalizeText(ans))
		}
		rows.Close()
	}

	matchPairs := map[string][]string{} // qid -> pairIDs
	if len(matchIDs) > 0 {
		rows, err := r.pool.Query(ctx, `
SELECT question_id::text, id::text
FROM question_matching_pairs
WHERE question_id::text = ANY($1::text[])`, matchIDs)
		if err != nil {
			return AutoScoreSummary{}, fmt.Errorf("load matching keys: %w", err)
		}
		for rows.Next() {
			var qid, pid string
			if err := rows.Scan(&qid, &pid); err != nil {
				rows.Close()
				return AutoScoreSummary{}, fmt.Errorf("scan matching key: %w", err)
			}
			matchPairs[qid] = append(matchPairs[qid], pid)
		}
		rows.Close()
		for qid := range matchPairs {
			sort.Strings(matchPairs[qid])
		}
	}

	essayMaxScores := map[string]int{}
	if len(essayIDs) > 0 {
		rows, err := r.pool.Query(ctx, `
SELECT question_id::text, COALESCE(max_score, 100)
FROM question_essays
WHERE question_id::text = ANY($1::text[])`, essayIDs)
		if err != nil {
			return AutoScoreSummary{}, fmt.Errorf("load essay max scores: %w", err)
		}
		for rows.Next() {
			var qid string
			var maxS int
			if err := rows.Scan(&qid, &maxS); err != nil {
				rows.Close()
				return AutoScoreSummary{}, fmt.Errorf("scan essay max score: %w", err)
			}
			essayMaxScores[qid] = maxS
		}
		rows.Close()
	}

	totalMax := 0
	totalActual := 0
	manualScored := 0
	pendingGrading := 0

	for _, q := range qs {
		att, ok := attempts[q.ID]
		raw := att.Answer
		switch q.Type {
		case "mc_single":
			autoTotal++
			totalMax += 100 // Scale to 100 per question internally
			if ok && isCorrectMCSingle(raw, mcCorrect[q.ID]) {
				correct++
				totalActual += 100
			}

		case "mc_multiple":
			autoTotal++
			totalMax += 100
			if ok && isCorrectMCMultiple(raw, mcCorrect[q.ID]) {
				correct++
				totalActual += 100
			}

		case "true_false":
			autoTotal++
			totalMax += 100
			if ok && isCorrectTrueFalse(raw, tfCorrect[q.ID]) {
				correct++
				totalActual += 100
			}

		case "short_answer":
			autoTotal++
			totalMax += 100
			if ok && isCorrectShortAnswer(raw, saCorrect[q.ID]) {
				correct++
				totalActual += 100
			}

		case "matching":
			autoTotal++
			totalMax += 100
			if ok && isCorrectMatching(raw, matchPairs[q.ID]) {
				correct++
				totalActual += 100
			}

		case "essay":
			maxS := essayMaxScores[q.ID]
			if maxS <= 0 {
				maxS = 100
			}
			totalMax += maxS
			if ok {
				if att.ManualScore != nil {
					totalActual += *att.ManualScore
					manualScored++
				} else {
					pendingGrading++
				}
			} else {
				// not answered, 0 points
			}

		default:
			// unknown type, skip
		}
	}

	score := 0
	if totalMax > 0 {
		score = int(math.Round(float64(totalActual) / float64(totalMax) * 100))
	}

	return AutoScoreSummary{
		TotalQuestions:    total,
		AnsweredQuestions: answered,
		AutoScorable:      autoTotal,
		CorrectCount:      correct,
		Score:             score,
		ManualScored:      manualScored,
		PendingGrading:    pendingGrading,
		TotalMaxScore:     totalMax,
		TotalActualScore:  totalActual,
	}, nil
}

func normalizeText(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	// collapse whitespace
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isCorrectMCSingle(raw json.RawMessage, correctOptions map[string]bool) bool {
	var req struct {
		SelectedOptionID string `json:"selected_option_id"`
	}
	if json.Unmarshal(raw, &req) != nil {
		return false
	}
	selected := strings.TrimSpace(req.SelectedOptionID)
	return selected != "" && correctOptions[selected]
}

func isCorrectMCMultiple(raw json.RawMessage, correctOptions map[string]bool) bool {
	var req struct {
		SelectedOptionIDs []string `json:"selected_option_ids"`
	}
	if json.Unmarshal(raw, &req) != nil {
		return false
	}
	want := keys(correctOptions)
	got := make([]string, 0, len(req.SelectedOptionIDs))
	seen := map[string]bool{}
	for _, id := range req.SelectedOptionIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		got = append(got, id)
	}
	sort.Strings(want)
	sort.Strings(got)
	return equalStrings(want, got)
}

func isCorrectTrueFalse(raw json.RawMessage, correctLabels map[string]bool) bool {
	var req struct {
		Value  *bool           `json:"value"`  // legacy
		Values map[string]bool `json:"values"` // new multi-statement
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return false
	}

	// Legacy single-toggle check
	if req.Value != nil {
		if want, ok := correctLabels["legacy"]; ok {
			return *req.Value == want
		}
	}

	// Multi-statement check: All statements must be correct
	if len(correctLabels) == 0 {
		return false
	}
	if len(req.Values) == 0 && len(correctLabels) > 0 {
		return false
	}

	for stid, correctVal := range correctLabels {
		if stid == "legacy" {
			continue
		}
		got, ok := req.Values[stid]
		if !ok || got != correctVal {
			return false
		}
	}
	return true
}

func isCorrectShortAnswer(raw json.RawMessage, acceptable []string) bool {
	var req struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &req) != nil {
		return false
	}
	v := normalizeText(req.Text)
	if v == "" {
		return false
	}
	for _, want := range acceptable {
		if v == want {
			return true
		}
	}
	return false
}

func isCorrectMatching(raw json.RawMessage, pairIDs []string) bool {
	var req struct {
		Pairs map[string]string `json:"pairs"`
	}
	if json.Unmarshal(raw, &req) != nil || len(req.Pairs) == 0 || len(pairIDs) == 0 {
		return false
	}
	for _, pid := range pairIDs {
		left := pid + ":L"
		rightWant := pid + ":R"
		got := strings.TrimSpace(req.Pairs[left])
		if got != rightWant {
			return false
		}
	}
	return true
}
