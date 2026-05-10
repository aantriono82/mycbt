package studentexamrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrDiscussionNotAvailable = errors.New("discussion not available")
	ErrDiscussionNotFound     = errors.New("discussion not found")
)

type StudentDiscussion struct {
	ExamID      string                  `json:"exam_id"`
	ExamTitle   string                  `json:"exam_title"`
	SessionID   string                  `json:"session_id"`
	AttemptNo   int                     `json:"attempt_no"`
	MaxAttempts int                     `json:"max_attempts"`
	Items       []StudentDiscussionItem `json:"items"`
}

type StudentDiscussionItem struct {
	OrderNo       int    `json:"order_no"`
	QuestionID    string `json:"question_id"`
	QuestionType  string `json:"question_type"`
	Stem          string `json:"stem"`
	Explanation   string `json:"explanation,omitempty"`
	StudentAnswer any    `json:"student_answer,omitempty"`
	CorrectAnswer any    `json:"correct_answer,omitempty"`
	IsCorrect     *bool  `json:"is_correct,omitempty"`
}

func (r *Repo) GetStudentDiscussionByExam(ctx context.Context, examID, studentID string) (StudentDiscussion, error) {
	const eligibilityQ = `
SELECT e.title,
       e.max_attempts,
       e.show_discussion_to_students,
       (
         SELECT COUNT(*)
         FROM exam_sessions sx
         WHERE sx.exam_id = e.id
           AND sx.student_id = $2
           AND sx.status IN ('submitted','forced','expired')
       ) AS attempts_used
FROM exams e
WHERE e.id = $1`
	var out StudentDiscussion
	var show bool
	var attemptsUsed int
	if err := r.pool.QueryRow(ctx, eligibilityQ, examID, studentID).Scan(&out.ExamTitle, &out.MaxAttempts, &show, &attemptsUsed); err != nil {
		return StudentDiscussion{}, ErrDiscussionNotFound
	}
	if out.MaxAttempts <= 0 {
		out.MaxAttempts = 1
	}
	if !show || attemptsUsed < out.MaxAttempts {
		return StudentDiscussion{}, ErrDiscussionNotAvailable
	}

	const sessionQ = `
SELECT id::text, COALESCE(attempt_no, 1)
FROM exam_sessions
WHERE exam_id = $1
  AND student_id = $2
  AND status IN ('submitted','forced','expired')
ORDER BY attempt_no DESC, finished_at DESC NULLS LAST, started_at DESC, id DESC
LIMIT 1`
	if err := r.pool.QueryRow(ctx, sessionQ, examID, studentID).Scan(&out.SessionID, &out.AttemptNo); err != nil {
		return StudentDiscussion{}, ErrDiscussionNotFound
	}
	out.ExamID = examID

	rows, err := r.pool.Query(ctx, `
SELECT sq.order_no, q.id::text, q.type, q.stem, COALESCE(q.explanation,'')
FROM exam_session_questions sq
JOIN questions q ON q.id = sq.question_id
WHERE sq.exam_session_id = $1
ORDER BY sq.order_no ASC`, out.SessionID)
	if err != nil {
		return StudentDiscussion{}, fmt.Errorf("list discussion questions: %w", err)
	}
	defer rows.Close()

	items := make([]StudentDiscussionItem, 0)
	qids := make([]string, 0)
	for rows.Next() {
		var it StudentDiscussionItem
		if err := rows.Scan(&it.OrderNo, &it.QuestionID, &it.QuestionType, &it.Stem, &it.Explanation); err != nil {
			return StudentDiscussion{}, fmt.Errorf("scan question: %w", err)
		}
		items = append(items, it)
		qids = append(qids, it.QuestionID)
	}
	if err := rows.Err(); err != nil {
		return StudentDiscussion{}, err
	}

	attempts := map[string]json.RawMessage{}
	arows, err := r.pool.Query(ctx, `SELECT question_id::text, answer_json FROM exam_attempts WHERE exam_session_id = $1`, out.SessionID)
	if err != nil {
		return StudentDiscussion{}, fmt.Errorf("list attempts: %w", err)
	}
	for arows.Next() {
		var qid string
		var raw []byte
		if err := arows.Scan(&qid, &raw); err != nil {
			arows.Close()
			return StudentDiscussion{}, fmt.Errorf("scan attempt: %w", err)
		}
		attempts[qid] = json.RawMessage(raw)
	}
	arows.Close()

	mcOpts := map[string][]struct {
		ID        string
		IsCorrect bool
	}{}
	orows, err := r.pool.Query(ctx, `
SELECT question_id::text, id::text, is_correct
FROM question_options
WHERE question_id::text = ANY($1::text[])`, qids)
	if err != nil {
		return StudentDiscussion{}, fmt.Errorf("list options: %w", err)
	}
	for orows.Next() {
		var qid, oid string
		var ok bool
		if err := orows.Scan(&qid, &oid, &ok); err != nil {
			orows.Close()
			return StudentDiscussion{}, fmt.Errorf("scan option: %w", err)
		}
		mcOpts[qid] = append(mcOpts[qid], struct {
			ID        string
			IsCorrect bool
		}{ID: oid, IsCorrect: ok})
	}
	orows.Close()

	shortAnswers := map[string][]string{}
	srows, err := r.pool.Query(ctx, `SELECT question_id::text, answer_text FROM question_short_answers WHERE question_id::text = ANY($1::text[])`, qids)
	if err != nil {
		return StudentDiscussion{}, fmt.Errorf("list short answers: %w", err)
	}
	for srows.Next() {
		var qid, ans string
		if err := srows.Scan(&qid, &ans); err != nil {
			srows.Close()
			return StudentDiscussion{}, fmt.Errorf("scan short answer: %w", err)
		}
		shortAnswers[qid] = append(shortAnswers[qid], strings.TrimSpace(ans))
	}
	srows.Close()

	tfSingle := map[string]bool{}
	tfRows, err := r.pool.Query(ctx, `SELECT question_id::text, correct FROM question_true_false WHERE question_id::text = ANY($1::text[])`, qids)
	if err != nil {
		return StudentDiscussion{}, fmt.Errorf("list true_false: %w", err)
	}
	for tfRows.Next() {
		var qid string
		var c bool
		if err := tfRows.Scan(&qid, &c); err != nil {
			tfRows.Close()
			return StudentDiscussion{}, fmt.Errorf("scan true_false: %w", err)
		}
		tfSingle[qid] = c
	}
	tfRows.Close()

	tfStatements := map[string]map[string]bool{}
	tfsRows, err := r.pool.Query(ctx, `SELECT question_id::text, id, correct FROM question_true_false_statements WHERE question_id::text = ANY($1::text[])`, qids)
	if err != nil {
		return StudentDiscussion{}, fmt.Errorf("list tf statements: %w", err)
	}
	for tfsRows.Next() {
		var qid, sid string
		var c bool
		if err := tfsRows.Scan(&qid, &sid, &c); err != nil {
			tfsRows.Close()
			return StudentDiscussion{}, fmt.Errorf("scan tf statement: %w", err)
		}
		if _, ok := tfStatements[qid]; !ok {
			tfStatements[qid] = map[string]bool{}
		}
		tfStatements[qid][sid] = c
	}
	tfsRows.Close()

	matchPairs := map[string]map[string]string{}
	mpRows, err := r.pool.Query(ctx, `SELECT question_id::text, id::text FROM question_matching_pairs WHERE question_id::text = ANY($1::text[])`, qids)
	if err != nil {
		return StudentDiscussion{}, fmt.Errorf("list matching pairs: %w", err)
	}
	for mpRows.Next() {
		var qid, pid string
		if err := mpRows.Scan(&qid, &pid); err != nil {
			mpRows.Close()
			return StudentDiscussion{}, fmt.Errorf("scan matching pair: %w", err)
		}
		if _, ok := matchPairs[qid]; !ok {
			matchPairs[qid] = map[string]string{}
		}
		matchPairs[qid][pid+":L"] = pid + ":R"
	}
	mpRows.Close()

	for i := range items {
		raw := attempts[items[i].QuestionID]
		var studentAnswer any
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &studentAnswer)
			items[i].StudentAnswer = studentAnswer
		}
		switch items[i].QuestionType {
		case "mc_single":
			correctID := ""
			for _, o := range mcOpts[items[i].QuestionID] {
				if o.IsCorrect {
					correctID = o.ID
					break
				}
			}
			items[i].CorrectAnswer = map[string]any{"selected_option_id": correctID}
			if m, ok := studentAnswer.(map[string]any); ok {
				selected := strings.TrimSpace(toString(m["selected_option_id"]))
				v := selected != "" && selected == correctID
				items[i].IsCorrect = &v
			}
		case "mc_multiple":
			correct := make([]string, 0)
			for _, o := range mcOpts[items[i].QuestionID] {
				if o.IsCorrect {
					correct = append(correct, o.ID)
				}
			}
			sort.Strings(correct)
			items[i].CorrectAnswer = map[string]any{"selected_option_ids": correct}
			if m, ok := studentAnswer.(map[string]any); ok {
				selected := toStringSlice(m["selected_option_ids"])
				sort.Strings(selected)
				v := equalStringSlices(selected, correct)
				items[i].IsCorrect = &v
			}
		case "short_answer":
			acc := shortAnswers[items[i].QuestionID]
			items[i].CorrectAnswer = map[string]any{"accepted_texts": acc}
			if m, ok := studentAnswer.(map[string]any); ok {
				txt := strings.TrimSpace(strings.ToLower(toString(m["text"])))
				correct := false
				for _, a := range acc {
					if txt != "" && txt == strings.ToLower(strings.TrimSpace(a)) {
						correct = true
						break
					}
				}
				items[i].IsCorrect = &correct
			}
		case "true_false":
			if vals, ok := tfStatements[items[i].QuestionID]; ok && len(vals) > 0 {
				items[i].CorrectAnswer = map[string]any{"values": vals}
				if m, ok := studentAnswer.(map[string]any); ok {
					valueMap := map[string]bool{}
					if inner, ok := m["values"].(map[string]any); ok {
						for k, v := range inner {
							b, ok := toBool(v)
							if ok {
								valueMap[k] = b
							}
						}
					}
					correct := len(valueMap) == len(vals)
					for k, v := range vals {
						if valueMap[k] != v {
							correct = false
							break
						}
					}
					items[i].IsCorrect = &correct
				}
			} else if c, ok := tfSingle[items[i].QuestionID]; ok {
				items[i].CorrectAnswer = map[string]any{"value": c}
				if m, ok := studentAnswer.(map[string]any); ok {
					if v, ok := toBool(m["value"]); ok {
						correct := v == c
						items[i].IsCorrect = &correct
					}
				}
			}
		case "matching":
			correctPairs := matchPairs[items[i].QuestionID]
			items[i].CorrectAnswer = map[string]any{"pairs": correctPairs}
			if m, ok := studentAnswer.(map[string]any); ok {
				pairs := map[string]string{}
				if inner, ok := m["pairs"].(map[string]any); ok {
					for k, v := range inner {
						pairs[k] = toString(v)
					}
				}
				correct := len(pairs) == len(correctPairs)
				for l, r := range correctPairs {
					if pairs[l] != r {
						correct = false
						break
					}
				}
				items[i].IsCorrect = &correct
			}
		case "essay":
			// Manual scoring, do not force correctness boolean.
			items[i].CorrectAnswer = map[string]any{"note": "manual_scoring"}
		}
	}

	out.Items = items
	return out, nil
}

func toString(v any) string {
	switch vv := v.(type) {
	case string:
		return vv
	default:
		return ""
	}
}

func toStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		s := strings.TrimSpace(toString(it))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func toBool(v any) (bool, bool) {
	switch vv := v.(type) {
	case bool:
		return vv, true
	case string:
		switch strings.TrimSpace(strings.ToLower(vv)) {
		case "true", "1":
			return true, true
		case "false", "0":
			return false, true
		}
	}
	return false, false
}

func equalStringSlices(a, b []string) bool {
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
