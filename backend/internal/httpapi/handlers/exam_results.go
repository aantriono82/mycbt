package handlers

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/httpapi/params"
	"mycbt/backend/internal/repo/examrepo"
	"mycbt/backend/internal/repo/ltirepo"
	"mycbt/backend/internal/repo/studentexamrepo"
	"mycbt/backend/internal/service/ltisvc"
	"mycbt/backend/internal/service/notificationsvc"
	"mycbt/backend/internal/util/simplepdf"
)

type ExamResultsHandler struct {
	ex     *examrepo.Repo
	lti    *ltirepo.Repo
	ltiAGS *ltisvc.AGSService
	st     *studentexamrepo.Repo
	notif  *notificationsvc.Service
}

func NewExamResultsHandler(ex *examrepo.Repo, lti *ltirepo.Repo, ltiAGS *ltisvc.AGSService, st *studentexamrepo.Repo, notif *notificationsvc.Service) *ExamResultsHandler {
	return &ExamResultsHandler{ex: ex, lti: lti, ltiAGS: ltiAGS, st: st, notif: notif}
}

func (h *ExamResultsHandler) authorizeExam(c *gin.Context) (examrepo.Exam, bool) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	examID := c.Param("id")
	exam, ok, err := h.ex.Get(c.Request.Context(), examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return examrepo.Exam{}, false
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return examrepo.Exam{}, false
	}
	if role == "teacher" {
		tid, tidOK, tErr := h.ex.TeacherIDByUserID(c.Request.Context(), userID)
		if tErr != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return examrepo.Exam{}, false
		}
		if !tidOK || exam.TeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return examrepo.Exam{}, false
		}
	}
	return exam, true
}

func (h *ExamResultsHandler) List(c *gin.Context) {
	nowUTC := time.Now().UTC()

	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.st.ListExamSessionsWithScore(c.Request.Context(), c.Param("id"), studentexamrepo.ListExamSessionsFilter{
		Q:      q,
		Limit:  limit,
		Offset: offset,
		NowUTC: nowUTC,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{
		"data": items,
		"meta": gin.H{
			"exam":   exam,
			"q":      q,
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

func (h *ExamResultsHandler) Attendance(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, targeted, attended, err := h.st.ListExamAttendanceParticipants(c.Request.Context(), c.Param("id"), studentexamrepo.ExamAttendanceFilter{
		Q:      q,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	attendanceRate := 0
	if targeted > 0 {
		attendanceRate = int(math.Round((float64(attended) / float64(targeted)) * 100))
	}

	c.JSON(200, gin.H{
		"data": items,
		"meta": gin.H{
			"exam":                    exam,
			"q":                       q,
			"limit":                   limit,
			"offset":                  offset,
			"total":                   total,
			"targeted_students":       targeted,
			"attended_students":       attended,
			"attendance_rate_percent": attendanceRate,
		},
	})
}

func (h *ExamResultsHandler) SyncLTIScores(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}
	if h.lti == nil || h.ltiAGS == nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "LTI AGS service is not configured"}})
		return
	}

	targets, err := h.lti.ListAGSScoreTargets(c.Request.Context(), exam.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to load LTI score targets"}})
		return
	}

	synced := 0
	skipped := 0
	failed := 0
	details := make([]string, 0, 5)
	nowUTC := time.Now().UTC()

	for _, target := range targets {
		sum, err := h.st.ComputeAutoScoreAny(c.Request.Context(), target.SessionID, nowUTC)
		if err != nil {
			failed++
			if len(details) < 5 {
				details = append(details, fmt.Sprintf("session %s: gagal hitung nilai", target.SessionID))
			}
			continue
		}
		if sum.PendingGrading > 0 {
			skipped++
			if len(details) < 5 {
				details = append(details, fmt.Sprintf("session %s: essay belum selesai dikoreksi", target.SessionID))
			}
			continue
		}

		timestamp := target.FinishedAt
		if timestamp.IsZero() {
			timestamp = nowUTC
		}

		err = h.ltiAGS.PublishScore(c.Request.Context(), ltisvc.PublishScoreInput{
			Platform:    target.Platform,
			LineItemURL: target.LineItemURL,
			LTISub:      target.LTISub,
			Score:       float64(sum.Score),
			Timestamp:   timestamp,
		}, target.ScopeText)
		if err != nil {
			failed++
			if len(details) < 5 {
				details = append(details, fmt.Sprintf("session %s: %s", target.SessionID, err.Error()))
			}
			continue
		}
		synced++
	}

	c.JSON(200, gin.H{
		"data": gin.H{
			"exam_id":       exam.ID,
			"exam_title":    exam.Title,
			"target_count":  len(targets),
			"synced_count":  synced,
			"skipped_count": skipped,
			"failed_count":  failed,
			"details":       details,
		},
	})
}

func (h *ExamResultsHandler) ItemAnalysis(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	items, sessions, err := h.st.ListExamItemAnalysis(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{
		"data": items,
		"meta": gin.H{
			"exam":           exam,
			"total_items":    len(items),
			"total_sessions": sessions,
		},
	})
}

type ItemSuggestion struct {
	QuestionID string   `json:"question_id"`
	Tips       []string `json:"tips"`
	Priority   string   `json:"priority"` // low, medium, high
}

func (h *ExamResultsHandler) ItemSuggestions(c *gin.Context) {
	_, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	items, _, err := h.st.ListExamItemAnalysis(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	suggestions := make([]ItemSuggestion, 0)
	for _, row := range items {
		if !row.AutoScorable {
			continue
		}

		s := ItemSuggestion{
			QuestionID: row.QuestionID,
			Tips:       make([]string, 0),
			Priority:   "low",
		}

		// Analysis logic
		if row.DiscriminationIdx < 0 {
			s.Tips = append(s.Tips, "⚠️ Daya pembeda negatif: Siswa berkemampuan tinggi justru lebih banyak menjawab salah. Periksa kembali validitas kunci jawaban.")
			s.Priority = "high"
		} else if row.DiscriminationIdx < 20 {
			s.Tips = append(s.Tips, "⚠️ Daya pembeda rendah: Soal kurang efektif membedakan siswa pintar dan kurang. Pertimbangkan revisi pada pilihan jawaban.")
			s.Priority = "medium"
		}

		if row.PValuePercent < 30 {
			if row.DiscriminationIdx < 20 {
				s.Tips = append(s.Tips, "❌ Soal terlalu sulit dan membingungkan. Periksa apakah materi ini sudah diajarkan atau kalimat soal terlalu kompleks.")
				s.Priority = "high"
			} else {
				s.Tips = append(s.Tips, "ℹ️ Soal tergolong sulit, namun masih valid secara statistik untuk menyeleksi siswa.")
			}
		} else if row.PValuePercent > 90 {
			s.Tips = append(s.Tips, "✅ Soal sangat mudah. Bagus untuk memotivasi, namun kurang menantang untuk ujian seleksi.")
		}

		// Distractor check
		var bestCorrect float64
		for _, opt := range row.OptionStats {
			if opt.IsCorrect {
				bestCorrect = float64(opt.SelectedCount)
			}
		}
		for _, opt := range row.OptionStats {
			if !opt.IsCorrect && opt.SelectedCount > 0 && float64(opt.SelectedCount) > bestCorrect {
				s.Tips = append(s.Tips, fmt.Sprintf("🧐 Distraktor %s lebih banyak dipilih dari kunci. Periksa apakah ada jebakan atau kebenaran parsial di opsi tersebut.", opt.Label))
				if s.Priority != "high" {
					s.Priority = "medium"
				}
			}
		}

		if len(s.Tips) > 0 {
			suggestions = append(suggestions, s)
		}
	}

	c.JSON(200, gin.H{"data": suggestions})
}

func (h *ExamResultsHandler) Export(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	nowUTC := time.Now().UTC()
	items, _, err := h.st.ListExamSessionsWithScore(c.Request.Context(), c.Param("id"), studentexamrepo.ListExamSessionsFilter{
		Q:      "",
		Limit:  5000,
		Offset: 0,
		NowUTC: nowUTC,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	scoreDist, err := h.st.GetExamScoreDistribution(c.Request.Context(), c.Param("id"), nowUTC)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	itemRows, itemSessions, err := h.st.ListExamItemAnalysis(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	f := excelize.NewFile()
	const resultsSheet = "Results"
	f.SetSheetName("Sheet1", resultsSheet)

	headers := []string{
		"No",
		"Nama",
		"Username",
		"NIS",
		"Status",
		"Mulai",
		"Selesai",
		"Dijawab",
		"Auto Scorable",
		"Benar",
		"Nilai",
		"Koreksi Essay",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(resultsSheet, cell, h)
	}

	for i, row := range items {
		r := i + 2
		f.SetCellValue(resultsSheet, fmt.Sprintf("A%d", r), i+1)
		f.SetCellValue(resultsSheet, fmt.Sprintf("B%d", r), row.StudentName)
		f.SetCellValue(resultsSheet, fmt.Sprintf("C%d", r), row.StudentUsername)
		f.SetCellValue(resultsSheet, fmt.Sprintf("D%d", r), row.StudentNIS)
		f.SetCellValue(resultsSheet, fmt.Sprintf("E%d", r), row.Status)
		f.SetCellValue(resultsSheet, fmt.Sprintf("F%d", r), row.StartedAt)
		f.SetCellValue(resultsSheet, fmt.Sprintf("G%d", r), row.FinishedAt)
		f.SetCellValue(resultsSheet, fmt.Sprintf("H%d", r), fmt.Sprintf("%d/%d", row.AnsweredQuestions, row.TotalQuestions))
		f.SetCellValue(resultsSheet, fmt.Sprintf("I%d", r), row.AutoScorable)
		f.SetCellValue(resultsSheet, fmt.Sprintf("J%d", r), row.CorrectCount)
		f.SetCellValue(resultsSheet, fmt.Sprintf("K%d", r), row.Score)
		gradingStatus := "-"
		if row.PendingGradingCount > 0 || row.ManualScoredCount > 0 {
			gradingStatus = fmt.Sprintf("%d/%d", row.ManualScoredCount, row.ManualScoredCount+row.PendingGradingCount)
		}
		f.SetCellValue(resultsSheet, fmt.Sprintf("L%d", r), gradingStatus)
	}

	f.SetCellValue(resultsSheet, "M1", "Ujian")
	f.SetCellValue(resultsSheet, "N1", exam.Title)
	f.SetCellValue(resultsSheet, "M2", "Generated At (UTC)")
	f.SetCellValue(resultsSheet, "N2", nowUTC.Format(time.RFC3339))

	styleID, styleErr := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#EEF2FF"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if styleErr == nil {
		f.SetCellStyle(resultsSheet, "A1", "K1", styleID)
	}
	_ = f.SetPanes(resultsSheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
	})
	_ = f.SetColWidth(resultsSheet, "A", "A", 6)
	_ = f.SetColWidth(resultsSheet, "B", "B", 24)
	_ = f.SetColWidth(resultsSheet, "C", "D", 14)
	_ = f.SetColWidth(resultsSheet, "E", "G", 14)
	_ = f.SetColWidth(resultsSheet, "H", "I", 12)
	_ = f.SetColWidth(resultsSheet, "J", "K", 10)
	_ = f.SetColWidth(resultsSheet, "L", "L", 14)
	_ = f.SetColWidth(resultsSheet, "M", "N", 24)

	const distSheet = "ScoreDistribution"
	if _, nerr := f.NewSheet(distSheet); nerr == nil {
		f.SetCellValue(distSheet, "A1", "Range")
		f.SetCellValue(distSheet, "B1", "Count")
		f.SetCellValue(distSheet, "C1", "Percent")
		if styleErr == nil {
			f.SetCellStyle(distSheet, "A1", "C1", styleID)
		}
		for i, bin := range scoreDist.DistributionBin {
			r := i + 2
			f.SetCellValue(distSheet, fmt.Sprintf("A%d", r), bin.Label)
			f.SetCellValue(distSheet, fmt.Sprintf("B%d", r), bin.Count)
			f.SetCellValue(distSheet, fmt.Sprintf("C%d", r), bin.Percent)
		}
		f.SetCellValue(distSheet, "E1", "Ujian")
		f.SetCellValue(distSheet, "F1", exam.Title)
		f.SetCellValue(distSheet, "E2", "Total Sessions")
		f.SetCellValue(distSheet, "F2", scoreDist.TotalSessions)
		f.SetCellValue(distSheet, "E3", "Submitted")
		f.SetCellValue(distSheet, "F3", scoreDist.SubmittedCount)
		f.SetCellValue(distSheet, "E4", "Expired")
		f.SetCellValue(distSheet, "F4", scoreDist.ExpiredCount)
		f.SetCellValue(distSheet, "E5", "Min")
		f.SetCellValue(distSheet, "F5", scoreDist.MinScore)
		f.SetCellValue(distSheet, "E6", "Average")
		f.SetCellValue(distSheet, "F6", scoreDist.AverageScore)
		f.SetCellValue(distSheet, "E7", "Median")
		f.SetCellValue(distSheet, "F7", scoreDist.MedianScore)
		f.SetCellValue(distSheet, "E8", "Max")
		f.SetCellValue(distSheet, "F8", scoreDist.MaxScore)
		f.SetCellValue(distSheet, "E9", "Generated At (UTC)")
		f.SetCellValue(distSheet, "F9", nowUTC.Format(time.RFC3339))
		_ = f.SetColWidth(distSheet, "A", "C", 14)
		_ = f.SetColWidth(distSheet, "E", "F", 22)
	}

	const itemSheet = "ItemAnalysis"
	if _, nerr := f.NewSheet(itemSheet); nerr == nil {
		itemHeaders := []string{
			"No", "Order", "Tipe", "Stem", "Peserta", "Jawab", "Benar",
			"P-Value(%)", "Kategori", "D-Index(%)", "D-Label", "Upper", "Lower", "Distraktor",
		}
		for i, h := range itemHeaders {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(itemSheet, cell, h)
		}
		if styleErr == nil {
			f.SetCellStyle(itemSheet, "A1", "N1", styleID)
		}
		for i, row := range itemRows {
			r := i + 2
			f.SetCellValue(itemSheet, fmt.Sprintf("A%d", r), i+1)
			f.SetCellValue(itemSheet, fmt.Sprintf("B%d", r), row.OrderNo)
			f.SetCellValue(itemSheet, fmt.Sprintf("C%d", r), row.QuestionType)
			f.SetCellValue(itemSheet, fmt.Sprintf("D%d", r), row.Stem)
			f.SetCellValue(itemSheet, fmt.Sprintf("E%d", r), row.Participants)
			f.SetCellValue(itemSheet, fmt.Sprintf("F%d", r), row.AnsweredCount)
			f.SetCellValue(itemSheet, fmt.Sprintf("G%d", r), row.CorrectCount)
			f.SetCellValue(itemSheet, fmt.Sprintf("H%d", r), row.PValuePercent)
			f.SetCellValue(itemSheet, fmt.Sprintf("I%d", r), row.DifficultyLabel)
			f.SetCellValue(itemSheet, fmt.Sprintf("J%d", r), row.DiscriminationIdx)
			f.SetCellValue(itemSheet, fmt.Sprintf("K%d", r), row.DiscriminationNote)
			f.SetCellValue(itemSheet, fmt.Sprintf("L%d", r), row.UpperCorrectCount)
			f.SetCellValue(itemSheet, fmt.Sprintf("M%d", r), row.LowerCorrectCount)
			distractor := "-"
			if len(row.OptionStats) > 0 {
				parts := make([]string, 0, len(row.OptionStats))
				for _, opt := range row.OptionStats {
					keyMark := ""
					if opt.IsCorrect {
						keyMark = "*"
					}
					parts = append(parts, fmt.Sprintf("%s%s:%d(%.2f%%)", opt.Label, keyMark, opt.SelectedCount, opt.SelectedPct))
				}
				distractor = strings.Join(parts, " | ")
			}
			f.SetCellValue(itemSheet, fmt.Sprintf("N%d", r), distractor)
		}
		f.SetCellValue(itemSheet, "P1", "Ujian")
		f.SetCellValue(itemSheet, "Q1", exam.Title)
		f.SetCellValue(itemSheet, "P2", "Total Items")
		f.SetCellValue(itemSheet, "Q2", len(itemRows))
		f.SetCellValue(itemSheet, "P3", "Total Sessions")
		f.SetCellValue(itemSheet, "Q3", itemSessions)
		_ = f.SetColWidth(itemSheet, "A", "C", 12)
		_ = f.SetColWidth(itemSheet, "D", "D", 44)
		_ = f.SetColWidth(itemSheet, "E", "M", 12)
		_ = f.SetColWidth(itemSheet, "N", "N", 48)
		_ = f.SetColWidth(itemSheet, "P", "Q", 22)
	}

	difficultItems := 0
	lowDiscItems := 0
	negativeDiscItems := 0
	nonAutoItems := 0
	for _, row := range itemRows {
		if !row.AutoScorable {
			nonAutoItems++
			continue
		}
		if row.DifficultyLabel == "sulit" {
			difficultItems++
		}
		if row.DiscriminationIdx < 0 {
			negativeDiscItems++
		}
		if row.DiscriminationIdx < 20 {
			lowDiscItems++
		}
	}

	recommendations := make([]string, 0, 6)
	if scoreDist.AverageScore < 70 {
		recommendations = append(recommendations, "Rata-rata nilai < 70: pertimbangkan remedial dan review materi inti.")
	}
	if len(itemRows) > 0 && difficultItems*3 > len(itemRows) {
		recommendations = append(recommendations, "Butir sulit > 33%: evaluasi tingkat kesulitan blueprint dan distribusi level kognitif.")
	}
	if len(itemRows) > 0 && lowDiscItems*3 > len(itemRows) {
		recommendations = append(recommendations, "Daya pembeda rendah pada banyak butir: revisi stem, opsi, dan kunci jawaban.")
	}
	if negativeDiscItems > 0 {
		recommendations = append(recommendations, "Ada butir dengan daya pembeda negatif: prioritaskan audit kunci/distraktor pada butir tersebut.")
	}
	if scoreDist.ExpiredCount > 0 {
		recommendations = append(recommendations, "Ada sesi expired: cek kecukupan durasi ujian dan kesiapan perangkat/jaringan.")
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Kualitas ujian relatif stabil. Lanjutkan bank soal dengan review butir periodik.")
	}

	const summarySheet = "ExecutiveSummary"
	if _, nerr := f.NewSheet(summarySheet); nerr == nil {
		neutralStyle, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F3F4F6"}},
		})
		goodStyle, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#DCFCE7"}},
		})
		warnStyle, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FEF3C7"}},
		})
		badStyle, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FEE2E2"}},
		})

		f.SetCellValue(summarySheet, "A1", "Exam")
		f.SetCellValue(summarySheet, "B1", exam.Title)
		f.SetCellValue(summarySheet, "A2", "Generated At (UTC)")
		f.SetCellValue(summarySheet, "B2", nowUTC.Format(time.RFC3339))
		f.SetCellValue(summarySheet, "A4", "KPI")
		f.SetCellValue(summarySheet, "B4", "Value")
		f.SetCellValue(summarySheet, "C4", "Status")
		if styleErr == nil {
			f.SetCellStyle(summarySheet, "A4", "C4", styleID)
		}

		kpis := []struct {
			Key string
			Val any
		}{
			{"Total Sessions", scoreDist.TotalSessions},
			{"Submitted Sessions", scoreDist.SubmittedCount},
			{"Expired Sessions", scoreDist.ExpiredCount},
			{"Average Score", scoreDist.AverageScore},
			{"Median Score", scoreDist.MedianScore},
			{"Min Score", scoreDist.MinScore},
			{"Max Score", scoreDist.MaxScore},
			{"Total Items", len(itemRows)},
			{"Difficult Items", difficultItems},
			{"Low Discrimination Items", lowDiscItems},
			{"Negative Discrimination Items", negativeDiscItems},
			{"Non Auto-Scorable Items", nonAutoItems},
		}
		for i, it := range kpis {
			r := i + 5
			f.SetCellValue(summarySheet, fmt.Sprintf("A%d", r), it.Key)
			f.SetCellValue(summarySheet, fmt.Sprintf("B%d", r), it.Val)
			// Visual + textual KPI indicator for quick review in exported report.
			valueCell := fmt.Sprintf("B%d", r)
			statusCell := fmt.Sprintf("C%d", r)
			statusText := "INFO"
			styleToApply := neutralStyle
			switch it.Key {
			case "Average Score":
				v := scoreDist.AverageScore
				if v >= 80 {
					statusText = "OK"
					styleToApply = goodStyle
				} else if v >= 70 {
					statusText = "WARN"
					styleToApply = warnStyle
				} else {
					statusText = "ALERT"
					styleToApply = badStyle
				}
			case "Expired Sessions":
				if scoreDist.ExpiredCount == 0 {
					statusText = "OK"
					styleToApply = goodStyle
				} else if scoreDist.ExpiredCount <= 2 {
					statusText = "WARN"
					styleToApply = warnStyle
				} else {
					statusText = "ALERT"
					styleToApply = badStyle
				}
			case "Low Discrimination Items":
				if lowDiscItems == 0 {
					statusText = "OK"
					styleToApply = goodStyle
				} else if lowDiscItems*3 <= len(itemRows) {
					statusText = "WARN"
					styleToApply = warnStyle
				} else {
					statusText = "ALERT"
					styleToApply = badStyle
				}
			case "Negative Discrimination Items":
				if negativeDiscItems == 0 {
					statusText = "OK"
					styleToApply = goodStyle
				} else {
					statusText = "ALERT"
					styleToApply = badStyle
				}
			case "Difficult Items":
				if len(itemRows) == 0 || difficultItems*4 <= len(itemRows) {
					statusText = "OK"
					styleToApply = goodStyle
				} else if difficultItems*3 <= len(itemRows) {
					statusText = "WARN"
					styleToApply = warnStyle
				} else {
					statusText = "ALERT"
					styleToApply = badStyle
				}
			default:
				statusText = "INFO"
				styleToApply = neutralStyle
			}
			f.SetCellValue(summarySheet, statusCell, statusText)
			f.SetCellStyle(summarySheet, valueCell, valueCell, styleToApply)
			f.SetCellStyle(summarySheet, statusCell, statusCell, styleToApply)
		}

		startRecRow := 5 + len(kpis) + 2
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", startRecRow), "Recommendations")
		if styleErr == nil {
			f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", startRecRow), fmt.Sprintf("B%d", startRecRow), styleID)
		}
		for i, rec := range recommendations {
			r := startRecRow + 1 + i
			f.SetCellValue(summarySheet, fmt.Sprintf("A%d", r), i+1)
			f.SetCellValue(summarySheet, fmt.Sprintf("B%d", r), rec)
			f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", r), fmt.Sprintf("B%d", r), warnStyle)
		}
		_ = f.SetColWidth(summarySheet, "A", "A", 28)
		_ = f.SetColWidth(summarySheet, "B", "B", 96)
		_ = f.SetColWidth(summarySheet, "C", "C", 12)
		if idx, idxErr := f.GetSheetIndex(summarySheet); idxErr == nil {
			f.SetActiveSheet(idx)
		}
	}

	buf := bytes.NewBuffer(nil)
	if err := f.Write(buf); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to generate xlsx"}})
		return
	}

	filename := fmt.Sprintf("exam-results-%s.xlsx", url.QueryEscape(c.Param("id")))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *ExamResultsHandler) ExportPDF(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}
	nowUTC := time.Now().UTC()
	items, _, err := h.st.ListExamSessionsWithScore(c.Request.Context(), c.Param("id"), studentexamrepo.ListExamSessionsFilter{
		Q:      "",
		Limit:  5000,
		Offset: 0,
		NowUTC: nowUTC,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	doc := simplepdf.NewA4Landscape()
	y := 570.0
	lineH := 15.0

	headers := []string{"No", "NIS", "Nama", "Status", "Benar", "Salah", "Kosong", "Skor", "Waktu Selesai", "Durasi"}
	xs := []float64{24, 56, 120, 320, 382, 428, 474, 524, 570, 730}
	drawHeader := func(pageNum int) float64 {
		top := 570.0
		doc.AddText(simplepdf.Text{X: 24, Y: top, Size: 16, Body: "Rekap Hasil Ujian"})
		top -= 18
		doc.AddText(simplepdf.Text{X: 24, Y: top, Size: 11, Body: "Ujian: " + exam.Title})
		top -= 14
		doc.AddText(simplepdf.Text{X: 24, Y: top, Size: 10, Body: "Generated (UTC): " + nowUTC.Format(time.RFC3339)})
		doc.AddText(simplepdf.Text{X: 760, Y: top, Size: 10, Body: fmt.Sprintf("Hal: %d", pageNum)})
		top -= 18
		for i, htxt := range headers {
			doc.AddText(simplepdf.Text{X: xs[i], Y: top, Size: 9, Body: htxt})
		}
		top -= 8
		doc.AddLine(24, top, 818, top)
		return top - 12
	}

	pageNum := 1
	y = drawHeader(pageNum)

	for i, row := range items {
		if y < 26 {
			doc.AddPage()
			pageNum++
			y = drawHeader(pageNum)
		}
		wrong := row.AutoScorable - row.CorrectCount
		if wrong < 0 {
			wrong = 0
		}
		blank := row.TotalQuestions - row.AnsweredQuestions
		if blank < 0 {
			blank = 0
		}
		finished := row.FinishedAt
		if strings.TrimSpace(finished) == "" {
			finished = "-"
		}
		doc.AddText(simplepdf.Text{X: xs[0], Y: y, Size: 8, Body: fmt.Sprintf("%d", i+1)})
		doc.AddText(simplepdf.Text{X: xs[1], Y: y, Size: 8, Body: row.StudentNIS})
		doc.AddText(simplepdf.Text{X: xs[2], Y: y, Size: 8, Body: truncateText(row.StudentName, 32)})
		doc.AddText(simplepdf.Text{X: xs[3], Y: y, Size: 8, Body: row.Status})
		doc.AddText(simplepdf.Text{X: xs[4], Y: y, Size: 8, Body: fmt.Sprintf("%d", row.CorrectCount)})
		doc.AddText(simplepdf.Text{X: xs[5], Y: y, Size: 8, Body: fmt.Sprintf("%d", wrong)})
		doc.AddText(simplepdf.Text{X: xs[6], Y: y, Size: 8, Body: fmt.Sprintf("%d", blank)})
		doc.AddText(simplepdf.Text{X: xs[7], Y: y, Size: 8, Body: fmt.Sprintf("%d", row.Score)})
		doc.AddText(simplepdf.Text{X: xs[8], Y: y, Size: 8, Body: truncateText(finished, 26)})
		doc.AddText(simplepdf.Text{X: xs[9], Y: y, Size: 8, Body: formatDuration(row.DurationSeconds)})
		y -= lineH
	}

	pdfBytes, err := doc.Bytes()
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to generate pdf"}})
		return
	}
	filename := fmt.Sprintf("exam-results-%s.pdf", url.QueryEscape(c.Param("id")))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "application/pdf", pdfBytes)
}

func truncateText(s string, limit int) string {
	if limit <= 3 || len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "-"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (h *ExamResultsHandler) ExportItemAnalysis(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	nowUTC := time.Now().UTC()
	items, sessions, err := h.st.ListExamItemAnalysis(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	f := excelize.NewFile()
	const sheet = "ItemAnalysis"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"No",
		"Order",
		"Tipe",
		"Stem",
		"Peserta",
		"Jawab",
		"Benar",
		"P-Value(%)",
		"Kategori",
		"D-Index(%)",
		"D-Label",
		"Upper Correct",
		"Lower Correct",
		"Distraktor",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, row := range items {
		r := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), row.OrderNo)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), row.QuestionType)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), row.Stem)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), row.Participants)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), row.AnsweredCount)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), row.CorrectCount)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", r), row.PValuePercent)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", r), row.DifficultyLabel)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", r), row.DiscriminationIdx)
		f.SetCellValue(sheet, fmt.Sprintf("K%d", r), row.DiscriminationNote)
		f.SetCellValue(sheet, fmt.Sprintf("L%d", r), row.UpperCorrectCount)
		f.SetCellValue(sheet, fmt.Sprintf("M%d", r), row.LowerCorrectCount)

		distractor := "-"
		if len(row.OptionStats) > 0 {
			parts := make([]string, 0, len(row.OptionStats))
			for _, opt := range row.OptionStats {
				keyMark := ""
				if opt.IsCorrect {
					keyMark = "*"
				}
				parts = append(parts, fmt.Sprintf("%s%s:%d(%.2f%%)", opt.Label, keyMark, opt.SelectedCount, opt.SelectedPct))
			}
			distractor = strings.Join(parts, " | ")
		}
		f.SetCellValue(sheet, fmt.Sprintf("N%d", r), distractor)
	}

	f.SetCellValue(sheet, "P1", "Ujian")
	f.SetCellValue(sheet, "Q1", exam.Title)
	f.SetCellValue(sheet, "P2", "Total Items")
	f.SetCellValue(sheet, "Q2", len(items))
	f.SetCellValue(sheet, "P3", "Total Sessions")
	f.SetCellValue(sheet, "Q3", sessions)
	f.SetCellValue(sheet, "P4", "Generated At (UTC)")
	f.SetCellValue(sheet, "Q4", nowUTC.Format(time.RFC3339))

	styleID, styleErr := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#ECFEFF"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if styleErr == nil {
		f.SetCellStyle(sheet, "A1", "N1", styleID)
	}
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
	})
	_ = f.SetColWidth(sheet, "A", "A", 6)
	_ = f.SetColWidth(sheet, "B", "C", 12)
	_ = f.SetColWidth(sheet, "D", "D", 44)
	_ = f.SetColWidth(sheet, "E", "M", 14)
	_ = f.SetColWidth(sheet, "N", "N", 48)
	_ = f.SetColWidth(sheet, "P", "Q", 22)

	buf := bytes.NewBuffer(nil)
	if err := f.Write(buf); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to generate xlsx"}})
		return
	}

	filename := fmt.Sprintf("item-analysis-%s.xlsx", url.QueryEscape(c.Param("id")))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *ExamResultsHandler) ScoreDistribution(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	data, err := h.st.GetExamScoreDistribution(c.Request.Context(), c.Param("id"), time.Now().UTC())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{
		"data": data,
		"meta": gin.H{
			"exam": exam,
		},
	})
}
func (h *ExamResultsHandler) ListEssays(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	sessionID := c.Param("sessionId")
	// Safety: check if session belongs to exam
	var sessExamID string
	err := h.ex.Pool().QueryRow(c.Request.Context(), `SELECT exam_id::text FROM exam_sessions WHERE id = $1`, sessionID).Scan(&sessExamID)
	if err != nil {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "session not found"}})
		return
	}
	if sessExamID != exam.ID {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "session does not belong to this exam"}})
		return
	}

	items, err := h.st.ListEssayAttempts(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

func (h *ExamResultsHandler) SaveEssayScore(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	sessionID := c.Param("sessionId")
	// Safety: check if session belongs to exam
	var sessExamID string
	err := h.ex.Pool().QueryRow(c.Request.Context(), `SELECT exam_id::text FROM exam_sessions WHERE id = $1`, sessionID).Scan(&sessExamID)
	if err != nil {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "session not found"}})
		return
	}
	if sessExamID != exam.ID {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "session does not belong to this exam"}})
		return
	}

	var req struct {
		QuestionID string `json:"question_id" binding:"required"`
		Score      int    `json:"score"`
		Feedback   string `json:"feedback"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid request body"}})
		return
	}

	if err := h.st.SaveManualScoring(c.Request.Context(), sessionID, req.QuestionID, req.Score, req.Feedback); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to save score: " + err.Error()}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

type blastResultsReq struct {
	Channels []string `json:"channels"` // "email", "whatsapp"
}

func (h *ExamResultsHandler) BlastResults(c *gin.Context) {
	exam, ok := h.authorizeExam(c)
	if !ok {
		return
	}

	var req blastResultsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	nowUTC := time.Now().UTC()
	items, _, err := h.st.ListExamSessionsWithScore(c.Request.Context(), exam.ID, studentexamrepo.ListExamSessionsFilter{
		Q:      "",
		Limit:  5000,
		Offset: 0,
		NowUTC: nowUTC,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to fetch results"}})
		return
	}

	useEmail := false
	useWA := false
	for _, ch := range req.Channels {
		if ch == "email" {
			useEmail = true
		}
		if ch == "whatsapp" {
			useWA = true
		}
	}

	sentCount := 0
	failedCount := 0

	for _, row := range items {
		// Only blast if submitted or forced
		if row.Status != "submitted" && row.Status != "forced" {
			continue
		}

		subject := fmt.Sprintf("Hasil Ujian: %s", exam.Title)
		body := fmt.Sprintf(`Halo <b>%s</b>,<br><br>Berikut adalah hasil ujian Anda:<br>
Ujian: %s<br>
Nilai: <b>%d</b><br>
Benar: %d/%d<br><br>
Terima kasih.`, row.StudentName, exam.Title, row.Score, row.CorrectCount, row.TotalQuestions)

		waMsg := fmt.Sprintf("*Hasil Ujian: %s*\n\nHalo %s,\n\nNilai Anda: *%d*\nBenar: %d/%d\n\nTerima kasih.",
			exam.Title, row.StudentName, row.Score, row.CorrectCount, row.TotalQuestions)

		if useEmail && row.StudentEmail != "" {
			err := h.notif.SendEmail(c.Request.Context(), row.StudentEmail, subject, body)
			if err != nil {
				failedCount++
			} else {
				sentCount++
			}
		}
		if useWA && row.StudentPhone != "" {
			err := h.notif.SendWhatsApp(c.Request.Context(), row.StudentPhone, waMsg)
			if err != nil {
				failedCount++
			} else {
				sentCount++
			}
		}
	}

	c.JSON(200, gin.H{
		"data": gin.H{
			"sent_count":   sentCount,
			"failed_count": failedCount,
		},
	})
}
