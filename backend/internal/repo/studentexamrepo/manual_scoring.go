package studentexamrepo

import (
	"context"
	"fmt"
	"strings"
)

type EssayAttempt struct {
	QuestionID      string `json:"question_id"`
	QuestionStem    string `json:"question_stem"`
	QuestionOrder   int    `json:"question_order"`
	AnswerText      string `json:"answer_text"`
	ManualScore     *int   `json:"manual_score"`
	ManualFeedback  string `json:"manual_feedback"`
	MaxScore        int    `json:"max_score"`
	RubricText      string `json:"rubric_text"`
}

func (r *Repo) ListEssayAttempts(ctx context.Context, sessionID string) ([]EssayAttempt, error) {
	const q = `
SELECT q.id::text, q.stem, sq.order_no, 
       COALESCE(a.answer_json->>'text', ''),
       a.manual_score,
       COALESCE(a.manual_feedback, ''),
       COALESCE(qe.max_score, 100),
       COALESCE(qe.rubric_text, '')
FROM exam_session_questions sq
JOIN questions q ON q.id = sq.question_id
LEFT JOIN exam_attempts a ON a.exam_session_id = sq.exam_session_id AND a.question_id = sq.question_id
LEFT JOIN question_essays qe ON qe.question_id = q.id
WHERE sq.exam_session_id = $1 AND q.type = 'essay'
ORDER BY sq.order_no ASC`

	rows, err := r.pool.Query(ctx, q, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list essay attempts: %w", err)
	}
	defer rows.Close()

	out := []EssayAttempt{}
	for rows.Next() {
		var it EssayAttempt
		var feedback, rubric string
		if err := rows.Scan(&it.QuestionID, &it.QuestionStem, &it.QuestionOrder, &it.AnswerText, &it.ManualScore, &feedback, &it.MaxScore, &rubric); err != nil {
			return nil, fmt.Errorf("scan essay attempt: %w", err)
		}
		it.ManualFeedback = feedback
		it.RubricText = rubric
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Repo) SaveManualScoring(ctx context.Context, sessionID, questionID string, score int, feedback string) error {
	// Verify it's an essay question for this session.
	var qType string
	err := r.pool.QueryRow(ctx, `
SELECT q.type 
FROM exam_session_questions sq 
JOIN questions q ON q.id = sq.question_id 
WHERE sq.exam_session_id = $1 AND sq.question_id = $2`, sessionID, questionID).Scan(&qType)
	if err != nil {
		return fmt.Errorf("verify question: %w", err)
	}
	if qType != "essay" {
		return fmt.Errorf("question is not an essay")
	}

	// Update or insert (if student somehow skipped it but teacher still wants to score it? 
	// Usually attempts exist if answered, but we'll use upsert for safety).
	// Actually, if it's not answered, answer_json should be empty object.
	const upsert = `
INSERT INTO exam_attempts (exam_session_id, question_id, answer_json, manual_score, manual_feedback, updated_at)
VALUES ($1::uuid, $2::uuid, '{}'::jsonb, $3, $4, now())
ON CONFLICT (exam_session_id, question_id)
DO UPDATE SET manual_score = EXCLUDED.manual_score, 
             manual_feedback = EXCLUDED.manual_feedback, 
             updated_at = now()`

	_, err = r.pool.Exec(ctx, upsert, sessionID, questionID, score, strings.TrimSpace(feedback))
	if err != nil {
		return fmt.Errorf("upsert manual score: %w", err)
	}

	return nil
}
