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

type ExamItemOptionStat struct {
	OptionID      string  `json:"option_id"`
	Label         string  `json:"label"`
	Content       string  `json:"content"`
	IsCorrect     bool    `json:"is_correct"`
	SelectedCount int     `json:"selected_count"`
	SelectedPct   float64 `json:"selected_percent"`
}

type ExamItemAnalysisRow struct {
	QuestionID         string               `json:"question_id"`
	OrderNo            int                  `json:"order_no"`
	QuestionType       string               `json:"question_type"`
	Stem               string               `json:"stem"`
	Participants       int                  `json:"participants"`
	AnsweredCount      int                  `json:"answered_count"`
	CorrectCount       int                  `json:"correct_count"`
	PValuePercent      float64              `json:"p_value_percent"`
	DifficultyLabel    string               `json:"difficulty_label"`
	AutoScorable       bool                 `json:"auto_scorable"`
	UnansweredCount    int                  `json:"unanswered_count"`
	IncorrectCount     int                  `json:"incorrect_count"`
	AnsweredRatePct    float64              `json:"answered_rate_percent"`
	CorrectRateOnAnsPt float64              `json:"correct_rate_on_answered_percent"`
	UpperGroupSize     int                  `json:"upper_group_size"`
	LowerGroupSize     int                  `json:"lower_group_size"`
	UpperCorrectCount  int                  `json:"upper_correct_count"`
	LowerCorrectCount  int                  `json:"lower_correct_count"`
	DiscriminationIdx  float64              `json:"discrimination_index"`
	DiscriminationNote string               `json:"discrimination_label"`
	OptionStats        []ExamItemOptionStat `json:"option_stats,omitempty"`
}

func (r *Repo) ListExamItemAnalysis(ctx context.Context, examID string) ([]ExamItemAnalysisRow, int, error) {
	type qmeta struct {
		ID      string
		OrderNo int
		QType   string
		Stem    string
	}
	type optionMeta struct {
		ID        string
		Label     string
		Content   string
		IsCorrect bool
	}

	qRows, err := r.pool.Query(ctx, `
SELECT DISTINCT ON (q.id)
  q.id::text,
  q.order_no,
  q.type,
  q.stem
FROM exam_question_sets eqs
JOIN questions q ON q.question_set_id = eqs.question_set_id
WHERE eqs.exam_id = $1
ORDER BY q.id, q.order_no ASC`, examID)
	if err != nil {
		return nil, 0, fmt.Errorf("list item analysis questions: %w", err)
	}
	defer qRows.Close()

	questions := make([]qmeta, 0)
	questionByID := map[string]qmeta{}
	for qRows.Next() {
		var it qmeta
		if scanErr := qRows.Scan(&it.ID, &it.OrderNo, &it.QType, &it.Stem); scanErr != nil {
			return nil, 0, fmt.Errorf("scan item question: %w", scanErr)
		}
		questions = append(questions, it)
		questionByID[it.ID] = it
	}
	if rowsErr := qRows.Err(); rowsErr != nil {
		return nil, 0, rowsErr
	}
	sort.SliceStable(questions, func(i, j int) bool {
		if questions[i].OrderNo == questions[j].OrderNo {
			return questions[i].ID < questions[j].ID
		}
		return questions[i].OrderNo < questions[j].OrderNo
	})

	var totalSessions int
	if err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM exam_sessions
WHERE exam_id = $1 AND status <> 'in_progress'`, examID).Scan(&totalSessions); err != nil {
		return nil, 0, fmt.Errorf("count sessions for item analysis: %w", err)
	}

	participantsByQ := map[string]int{}
	pRows, err := r.pool.Query(ctx, `
SELECT sq.question_id::text, COUNT(*)::int
FROM exam_session_questions sq
JOIN exam_sessions s ON s.id = sq.exam_session_id
WHERE s.exam_id = $1
  AND s.status <> 'in_progress'
GROUP BY sq.question_id`, examID)
	if err != nil {
		return nil, 0, fmt.Errorf("count participants by question: %w", err)
	}
	for pRows.Next() {
		var qid string
		var count int
		if scanErr := pRows.Scan(&qid, &count); scanErr != nil {
			pRows.Close()
			return nil, 0, fmt.Errorf("scan participants by question: %w", scanErr)
		}
		participantsByQ[qid] = count
	}
	pRows.Close()

	answeredByQ := map[string]int{}
	correctByQ := map[string]int{}
	mcSelectedByQ := map[string]map[string]int{}
	answeredBySession := map[string]map[string]bool{} // sessionID -> qid -> correct

	mcIDs := make([]string, 0)
	tfIDs := make([]string, 0)
	saIDs := make([]string, 0)
	matchIDs := make([]string, 0)
	for _, q := range questions {
		switch q.QType {
		case "mc_single", "mc_multiple":
			mcIDs = append(mcIDs, q.ID)
		case "true_false":
			tfIDs = append(tfIDs, q.ID)
		case "short_answer":
			saIDs = append(saIDs, q.ID)
		case "matching":
			matchIDs = append(matchIDs, q.ID)
		}
	}

	mcCorrect := map[string]map[string]bool{}
	mcOptionsByQ := map[string][]optionMeta{}
	if len(mcIDs) > 0 {
		rows, qErr := r.pool.Query(ctx, `
SELECT question_id::text, id::text, label, content, is_correct
FROM question_options
WHERE question_id::text = ANY($1::text[])
ORDER BY question_id, label, id`, mcIDs)
		if qErr != nil {
			return nil, 0, fmt.Errorf("load mc keys for item analysis: %w", qErr)
		}
		for rows.Next() {
			var qid string
			var opt optionMeta
			var isCorrect bool
			if scanErr := rows.Scan(&qid, &opt.ID, &opt.Label, &opt.Content, &isCorrect); scanErr != nil {
				rows.Close()
				return nil, 0, fmt.Errorf("scan mc key for item analysis: %w", scanErr)
			}
			opt.IsCorrect = isCorrect
			mcOptionsByQ[qid] = append(mcOptionsByQ[qid], opt)

			m, ok := mcCorrect[qid]
			if !ok {
				m = map[string]bool{}
				mcCorrect[qid] = m
			}
			if isCorrect {
				m[opt.ID] = true
			}
		}
		rows.Close()
	}

	tfCorrect := map[string]bool{}
	if len(tfIDs) > 0 {
		rows, qErr := r.pool.Query(ctx, `
SELECT question_id::text, correct
FROM question_true_false
WHERE question_id::text = ANY($1::text[])`, tfIDs)
		if qErr != nil {
			return nil, 0, fmt.Errorf("load tf keys for item analysis: %w", qErr)
		}
		for rows.Next() {
			var qid string
			var v bool
			if scanErr := rows.Scan(&qid, &v); scanErr != nil {
				rows.Close()
				return nil, 0, fmt.Errorf("scan tf key for item analysis: %w", scanErr)
			}
			tfCorrect[qid] = v
		}
		rows.Close()
	}

	saCorrect := map[string][]string{}
	if len(saIDs) > 0 {
		rows, qErr := r.pool.Query(ctx, `
SELECT question_id::text, answer_text
FROM question_short_answers
WHERE question_id::text = ANY($1::text[])`, saIDs)
		if qErr != nil {
			return nil, 0, fmt.Errorf("load short-answer keys for item analysis: %w", qErr)
		}
		for rows.Next() {
			var qid, answer string
			if scanErr := rows.Scan(&qid, &answer); scanErr != nil {
				rows.Close()
				return nil, 0, fmt.Errorf("scan short-answer key for item analysis: %w", scanErr)
			}
			saCorrect[qid] = append(saCorrect[qid], normalizeText(answer))
		}
		rows.Close()
	}

	matchPairs := map[string][]string{}
	if len(matchIDs) > 0 {
		rows, qErr := r.pool.Query(ctx, `
SELECT question_id::text, id::text
FROM question_matching_pairs
WHERE question_id::text = ANY($1::text[])
  AND NULLIF(TRIM(left_content), '') IS NOT NULL 
  AND TRIM(left_content) != '<p></p>'
  AND TRIM(left_content) != '<p><br></p>'`, matchIDs)
		if qErr != nil {
			return nil, 0, fmt.Errorf("load matching keys for item analysis: %w", qErr)
		}
		for rows.Next() {
			var qid, pid string
			if scanErr := rows.Scan(&qid, &pid); scanErr != nil {
				rows.Close()
				return nil, 0, fmt.Errorf("scan matching key for item analysis: %w", scanErr)
			}
			matchPairs[qid] = append(matchPairs[qid], pid)
		}
		rows.Close()
		for qid := range matchPairs {
			sort.Strings(matchPairs[qid])
		}
	}

	attemptRows, err := r.pool.Query(ctx, `
SELECT a.exam_session_id::text, a.question_id::text, q.type, a.answer_json
FROM exam_attempts a
JOIN exam_sessions s ON s.id = a.exam_session_id
JOIN questions q ON q.id = a.question_id
WHERE s.exam_id = $1
  AND s.status <> 'in_progress'`, examID)
	if err != nil {
		return nil, 0, fmt.Errorf("list attempts for item analysis: %w", err)
	}
	for attemptRows.Next() {
		var sid, qid, qType string
		var raw []byte
		if scanErr := attemptRows.Scan(&sid, &qid, &qType, &raw); scanErr != nil {
			attemptRows.Close()
			return nil, 0, fmt.Errorf("scan attempt for item analysis: %w", scanErr)
		}
		if _, exists := questionByID[qid]; !exists {
			continue
		}
		answeredByQ[qid]++

		isCorrect := false
		switch qType {
		case "mc_single":
			var req struct {
				SelectedOptionID string `json:"selected_option_id"`
			}
			if json.Unmarshal(raw, &req) == nil {
				req.SelectedOptionID = strings.TrimSpace(req.SelectedOptionID)
				if req.SelectedOptionID != "" {
					if _, ok := mcSelectedByQ[qid]; !ok {
						mcSelectedByQ[qid] = map[string]int{}
					}
					mcSelectedByQ[qid][req.SelectedOptionID]++
				}
				isCorrect = req.SelectedOptionID != "" && mcCorrect[qid][req.SelectedOptionID]
			}
		case "mc_multiple":
			var req struct {
				SelectedOptionIDs []string `json:"selected_option_ids"`
			}
			if json.Unmarshal(raw, &req) == nil {
				want := keys(mcCorrect[qid])
				got := make([]string, 0, len(req.SelectedOptionIDs))
				seen := map[string]bool{}
				for _, id := range req.SelectedOptionIDs {
					id = strings.TrimSpace(id)
					if id == "" || seen[id] {
						continue
					}
					seen[id] = true
					got = append(got, id)
					if _, ok := mcSelectedByQ[qid]; !ok {
						mcSelectedByQ[qid] = map[string]int{}
					}
					mcSelectedByQ[qid][id]++
				}
				sort.Strings(want)
				sort.Strings(got)
				isCorrect = equalStrings(want, got)
			}
		case "true_false":
			var req struct {
				Value *bool `json:"value"`
			}
			if json.Unmarshal(raw, &req) == nil && req.Value != nil {
				isCorrect = tfCorrect[qid] == *req.Value
			}
		case "short_answer":
			var req struct {
				Text string `json:"text"`
			}
			if json.Unmarshal(raw, &req) == nil {
				v := normalizeText(req.Text)
				for _, want := range saCorrect[qid] {
					if v != "" && v == want {
						isCorrect = true
						break
					}
				}
			}
		case "matching":
			var req struct {
				Pairs map[string]string `json:"pairs"`
			}
			if json.Unmarshal(raw, &req) == nil && len(req.Pairs) > 0 {
				allOK := true
				for _, pid := range matchPairs[qid] {
					left := pid + ":L"
					rightWant := pid + ":R"
					got := strings.TrimSpace(req.Pairs[left])
					if got != rightWant {
						allOK = false
						break
					}
				}
				isCorrect = allOK && len(matchPairs[qid]) > 0
			}
		}
		if isCorrect {
			correctByQ[qid]++
		}
		if _, ok := answeredBySession[sid]; !ok {
			answeredBySession[sid] = map[string]bool{}
		}
		answeredBySession[sid][qid] = isCorrect
	}
	attemptRows.Close()

	sessionRows, err := r.pool.Query(ctx, `
SELECT s.id::text
FROM exam_sessions s
WHERE s.exam_id = $1
  AND s.status <> 'in_progress'`, examID)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions for discrimination: %w", err)
	}
	type sessionScore struct {
		ID    string
		Score int
	}
	sessionScores := make([]sessionScore, 0)
	for sessionRows.Next() {
		var sid string
		if scanErr := sessionRows.Scan(&sid); scanErr != nil {
			sessionRows.Close()
			return nil, 0, fmt.Errorf("scan session id for discrimination: %w", scanErr)
		}
		sum, sErr := r.ComputeAutoScoreAny(ctx, sid, timeNowUTC())
		if sErr != nil {
			sessionRows.Close()
			return nil, 0, fmt.Errorf("compute session score for discrimination: %w", sErr)
		}
		sessionScores = append(sessionScores, sessionScore{ID: sid, Score: sum.Score})
	}
	sessionRows.Close()

	sort.SliceStable(sessionScores, func(i, j int) bool {
		if sessionScores[i].Score == sessionScores[j].Score {
			return sessionScores[i].ID < sessionScores[j].ID
		}
		return sessionScores[i].Score > sessionScores[j].Score
	})

	groupN := int(math.Ceil(float64(len(sessionScores)) * 0.27))
	if groupN < 1 && len(sessionScores) > 0 {
		groupN = 1
	}
	if groupN > len(sessionScores) {
		groupN = len(sessionScores)
	}
	upperIDs := map[string]bool{}
	lowerIDs := map[string]bool{}
	for i := 0; i < groupN; i++ {
		upperIDs[sessionScores[i].ID] = true
	}
	for i := len(sessionScores) - groupN; i < len(sessionScores); i++ {
		if i >= 0 {
			lowerIDs[sessionScores[i].ID] = true
		}
	}

	out := make([]ExamItemAnalysisRow, 0, len(questions))
	for _, q := range questions {
		participants := participantsByQ[q.ID]
		answered := answeredByQ[q.ID]
		correct := correctByQ[q.ID]
		auto := isAutoScorableType(q.QType)

		pValue := 0.0
		if auto && participants > 0 {
			pValue = round2((float64(correct) / float64(participants)) * 100.0)
		}

		answeredRate := 0.0
		if participants > 0 {
			answeredRate = round2((float64(answered) / float64(participants)) * 100.0)
		}

		correctOnAnswered := 0.0
		if auto && answered > 0 {
			correctOnAnswered = round2((float64(correct) / float64(answered)) * 100.0)
		}

		difficulty := "-"
		if auto && participants > 0 {
			difficulty = difficultyFromPValue(pValue)
		}

		incorrect := answered - correct
		if incorrect < 0 {
			incorrect = 0
		}
		unanswered := participants - answered
		if unanswered < 0 {
			unanswered = 0
		}

		optionStats := make([]ExamItemOptionStat, 0)
		if q.QType == "mc_single" || q.QType == "mc_multiple" {
			for _, opt := range mcOptionsByQ[q.ID] {
				selectedCount := mcSelectedByQ[q.ID][opt.ID]
				selectedPct := 0.0
				if participants > 0 {
					selectedPct = round2((float64(selectedCount) / float64(participants)) * 100.0)
				}
				optionStats = append(optionStats, ExamItemOptionStat{
					OptionID:      opt.ID,
					Label:         opt.Label,
					Content:       opt.Content,
					IsCorrect:     opt.IsCorrect,
					SelectedCount: selectedCount,
					SelectedPct:   selectedPct,
				})
			}
		}

		upperCorrect := 0
		lowerCorrect := 0
		for sid := range upperIDs {
			if answeredBySession[sid][q.ID] {
				upperCorrect++
			}
		}
		for sid := range lowerIDs {
			if answeredBySession[sid][q.ID] {
				lowerCorrect++
			}
		}
		discrimination := 0.0
		discriminationLabel := "-"
		if auto && groupN > 0 {
			discrimination = round2((float64(upperCorrect-lowerCorrect) / float64(groupN)) * 100.0)
			discriminationLabel = discriminationFromIndex(discrimination)
		}

		out = append(out, ExamItemAnalysisRow{
			QuestionID:         q.ID,
			OrderNo:            q.OrderNo,
			QuestionType:       q.QType,
			Stem:               q.Stem,
			Participants:       participants,
			AnsweredCount:      answered,
			CorrectCount:       correct,
			PValuePercent:      pValue,
			DifficultyLabel:    difficulty,
			AutoScorable:       auto,
			UnansweredCount:    unanswered,
			IncorrectCount:     incorrect,
			AnsweredRatePct:    answeredRate,
			CorrectRateOnAnsPt: correctOnAnswered,
			UpperGroupSize:     groupN,
			LowerGroupSize:     groupN,
			UpperCorrectCount:  upperCorrect,
			LowerCorrectCount:  lowerCorrect,
			DiscriminationIdx:  discrimination,
			DiscriminationNote: discriminationLabel,
			OptionStats:        optionStats,
		})
	}

	return out, totalSessions, nil
}

func isAutoScorableType(t string) bool {
	switch t {
	case "mc_single", "mc_multiple", "true_false", "short_answer", "matching":
		return true
	default:
		return false
	}
}

func difficultyFromPValue(p float64) string {
	if p >= 80 {
		return "mudah"
	}
	if p >= 30 {
		return "sedang"
	}
	return "sulit"
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func discriminationFromIndex(idx float64) string {
	if idx >= 40 {
		return "sangat baik"
	}
	if idx >= 30 {
		return "baik"
	}
	if idx >= 20 {
		return "cukup"
	}
	if idx >= 0 {
		return "kurang"
	}
	return "negatif"
}

func timeNowUTC() time.Time {
	return time.Now().UTC()
}
