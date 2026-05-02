package studentexamrepo

import "time"

func isFinalSessionStatus(status string) bool {
	switch status {
	case "submitted", "forced", "expired":
		return true
	default:
		return false
	}
}

func parseRFC3339(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func shouldReplaceBestSession(current, candidate ExamSessionRow) bool {
	currentFinal := isFinalSessionStatus(current.Status)
	candidateFinal := isFinalSessionStatus(candidate.Status)
	if currentFinal != candidateFinal {
		return candidateFinal
	}

	if candidate.Score != current.Score {
		return candidate.Score > current.Score
	}

	if candidate.AttemptNumber != current.AttemptNumber {
		return candidate.AttemptNumber > current.AttemptNumber
	}

	candidateFinished := parseRFC3339(candidate.FinishedAt)
	currentFinished := parseRFC3339(current.FinishedAt)
	if !candidateFinished.Equal(currentFinished) {
		return candidateFinished.After(currentFinished)
	}

	candidateStarted := parseRFC3339(candidate.StartedAt)
	currentStarted := parseRFC3339(current.StartedAt)
	if !candidateStarted.Equal(currentStarted) {
		return candidateStarted.After(currentStarted)
	}

	return candidate.SessionID > current.SessionID
}

func selectBestSessionsByStudent(rows []ExamSessionRow) []ExamSessionRow {
	bestByStudent := make(map[string]ExamSessionRow, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		current, ok := bestByStudent[row.StudentID]
		if !ok {
			bestByStudent[row.StudentID] = row
			order = append(order, row.StudentID)
			continue
		}
		if shouldReplaceBestSession(current, row) {
			bestByStudent[row.StudentID] = row
		}
	}

	out := make([]ExamSessionRow, 0, len(bestByStudent))
	for _, studentID := range order {
		out = append(out, bestByStudent[studentID])
	}
	return out
}

func shouldReplaceBestStudentResult(current, candidate StudentResultSummary) bool {
	currentFinal := isFinalSessionStatus(current.SessionStatus)
	candidateFinal := isFinalSessionStatus(candidate.SessionStatus)
	if currentFinal != candidateFinal {
		return candidateFinal
	}

	if candidate.Score != current.Score {
		return candidate.Score > current.Score
	}

	candidateFinished := parseRFC3339(candidate.SubmittedAt)
	currentFinished := parseRFC3339(current.SubmittedAt)
	if !candidateFinished.Equal(currentFinished) {
		return candidateFinished.After(currentFinished)
	}

	return candidate.SessionID > current.SessionID
}

func selectBestResultsByExam(rows []StudentResultSummary) []StudentResultSummary {
	bestByExam := make(map[string]StudentResultSummary, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		current, ok := bestByExam[row.ExamID]
		if !ok {
			bestByExam[row.ExamID] = row
			order = append(order, row.ExamID)
			continue
		}
		if shouldReplaceBestStudentResult(current, row) {
			bestByExam[row.ExamID] = row
		}
	}

	out := make([]StudentResultSummary, 0, len(bestByExam))
	for _, examID := range order {
		out = append(out, bestByExam[examID])
	}
	return out
}
