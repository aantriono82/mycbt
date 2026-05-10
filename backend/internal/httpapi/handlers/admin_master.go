package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"

	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/httpapi/params"
	"atigacbt/backend/internal/httpapi/pgerr"
	"atigacbt/backend/internal/repo/masterrepo"
	"atigacbt/backend/internal/repo/userrepo"
	"atigacbt/backend/internal/service/authsvc"
	"atigacbt/backend/internal/service/notificationsvc"
)

type AdminMasterHandler struct {
	pool *pgxpool.Pool

	users *userrepo.Repo

	programs      *masterrepo.ProgramsRepo
	levels        *masterrepo.LevelsRepo
	groups        *masterrepo.GroupsRepo
	subjects      *masterrepo.SubjectsRepo
	announcements *masterrepo.AnnouncementsRepo
	teacherSubs   *masterrepo.TeacherSubjectsRepo
	teacherGroups *masterrepo.TeacherGroupsRepo
	teacherLevels *masterrepo.TeacherLevelsRepo
	teachers      *masterrepo.TeachersRepo
	students      *masterrepo.StudentsRepo
	registrations *masterrepo.RegistrationRepo
	sessions      *masterrepo.SessionsRepo
	lookups       *masterrepo.Lookups
	notif         *notificationsvc.Service
}

func NewAdminMasterHandler(
	pool *pgxpool.Pool,
	users *userrepo.Repo,
	programs *masterrepo.ProgramsRepo,
	levels *masterrepo.LevelsRepo,
	groups *masterrepo.GroupsRepo,
	subjects *masterrepo.SubjectsRepo,
	announcements *masterrepo.AnnouncementsRepo,
	teacherSubs *masterrepo.TeacherSubjectsRepo,
	teacherGroups *masterrepo.TeacherGroupsRepo,
	teacherLevels *masterrepo.TeacherLevelsRepo,
	teachers *masterrepo.TeachersRepo,
	students *masterrepo.StudentsRepo,
	registrations *masterrepo.RegistrationRepo,
	sessions *masterrepo.SessionsRepo,
	lookups *masterrepo.Lookups,
	notif *notificationsvc.Service,
) *AdminMasterHandler {
	return &AdminMasterHandler{
		pool:          pool,
		users:         users,
		programs:      programs,
		levels:        levels,
		groups:        groups,
		subjects:      subjects,
		announcements: announcements,
		teacherSubs:   teacherSubs,
		teacherGroups: teacherGroups,
		teacherLevels: teacherLevels,
		teachers:      teachers,
		students:      students,
		registrations: registrations,
		sessions:      sessions,
		lookups:       lookups,
		notif:         notif,
	}
}

func (h *AdminMasterHandler) ListPrograms(c *gin.Context) {
	items, err := h.programs.List(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type createProgramReq struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (h *AdminMasterHandler) CreateProgram(c *gin.Context) {
	var req createProgramReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	p, err := h.programs.Create(c.Request.Context(), strings.TrimSpace(req.Code), strings.TrimSpace(req.Name))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": p})
}

func (h *AdminMasterHandler) GetProgram(c *gin.Context) {
	it, ok, err := h.programs.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) UpdateProgram(c *gin.Context) {
	var req createProgramReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	it, ok, err := h.programs.Update(c.Request.Context(), c.Param("id"), strings.TrimSpace(req.Code), strings.TrimSpace(req.Name))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) DeleteProgram(c *gin.Context) {
	ok, err := h.programs.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "in use"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) ListLevels(c *gin.Context) {
	items, err := h.levels.List(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type createLevelReq struct {
	Name  string `json:"name"`
	Kelas *int   `json:"kelas"`
}

func (h *AdminMasterHandler) CreateLevel(c *gin.Context) {
	var req createLevelReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	it, err := h.levels.Create(c.Request.Context(), strings.TrimSpace(req.Name), req.Kelas)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *AdminMasterHandler) GetLevel(c *gin.Context) {
	it, ok, err := h.levels.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) UpdateLevel(c *gin.Context) {
	var req createLevelReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	it, ok, err := h.levels.Update(c.Request.Context(), c.Param("id"), strings.TrimSpace(req.Name), req.Kelas)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) DeleteLevel(c *gin.Context) {
	ok, err := h.levels.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "in use"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) ListGroups(c *gin.Context) {
	items, err := h.groups.List(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type createGroupReq struct {
	Name string `json:"name"`
}

func (h *AdminMasterHandler) CreateGroup(c *gin.Context) {
	var req createGroupReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	it, err := h.groups.Create(c.Request.Context(), strings.TrimSpace(req.Name))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *AdminMasterHandler) GetGroup(c *gin.Context) {
	it, ok, err := h.groups.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) UpdateGroup(c *gin.Context) {
	var req createGroupReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	it, ok, err := h.groups.Update(c.Request.Context(), c.Param("id"), strings.TrimSpace(req.Name))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) DeleteGroup(c *gin.Context) {
	ok, err := h.groups.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "in use"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) ListSessions(c *gin.Context) {
	items, err := h.sessions.List(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type createSessionReq struct {
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (h *AdminMasterHandler) CreateSession(c *gin.Context) {
	var req createSessionReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	// Normalisasi format jam (07.30 -> 07:30)
	startTime := strings.ReplaceAll(strings.TrimSpace(req.StartTime), ".", ":")
	endTime := strings.ReplaceAll(strings.TrimSpace(req.EndTime), ".", ":")

	it, err := h.sessions.Create(c.Request.Context(), strings.TrimSpace(req.Name), startTime, endTime)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "Gagal menyimpan sesi. Pastikan format jam benar (HH:MM)"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *AdminMasterHandler) GetSession(c *gin.Context) {
	it, ok, err := h.sessions.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) UpdateSession(c *gin.Context) {
	var req createSessionReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	// Normalisasi format jam (07.30 -> 07:30)
	startTime := strings.ReplaceAll(strings.TrimSpace(req.StartTime), ".", ":")
	endTime := strings.ReplaceAll(strings.TrimSpace(req.EndTime), ".", ":")

	it, ok, err := h.sessions.Update(c.Request.Context(), c.Param("id"), strings.TrimSpace(req.Name), startTime, endTime)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "Gagal memperbarui sesi. Pastikan format jam benar (HH:MM)"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) DeleteSession(c *gin.Context) {
	ok, err := h.sessions.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "in use"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) ListSubjects(c *gin.Context) {
	items, err := h.subjects.List(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type createSubjectReq struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (h *AdminMasterHandler) CreateSubject(c *gin.Context) {
	var req createSubjectReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	it, err := h.subjects.Create(c.Request.Context(), strings.TrimSpace(req.Code), strings.TrimSpace(req.Name))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *AdminMasterHandler) GetSubject(c *gin.Context) {
	it, ok, err := h.subjects.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) UpdateSubject(c *gin.Context) {
	var req createSubjectReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	it, ok, err := h.subjects.Update(c.Request.Context(), c.Param("id"), strings.TrimSpace(req.Code), strings.TrimSpace(req.Name))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) DeleteSubject(c *gin.Context) {
	ok, err := h.subjects.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "in use"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

type adminAnnouncementReq struct {
	Title           string `json:"title"`
	Body            string `json:"body"`
	Category        string `json:"category"`
	IsActive        *bool  `json:"is_active"`
	PublishedAt     string `json:"published_at"`
	ExpiresAt       string `json:"expires_at"`
	TargetLevelID   string `json:"target_level_id"`
	TargetGroupID   string `json:"target_group_id"`
	TargetStudentID string `json:"target_student_id"`
}

func parseRFC3339Optional(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	v := parsed.UTC()
	return &v, nil
}

func targetCount(req adminAnnouncementReq) int {
	n := 0
	if strings.TrimSpace(req.TargetLevelID) != "" {
		n++
	}
	if strings.TrimSpace(req.TargetGroupID) != "" {
		n++
	}
	if strings.TrimSpace(req.TargetStudentID) != "" {
		n++
	}
	return n
}

func (h *AdminMasterHandler) ListAnnouncements(c *gin.Context) {
	q := params.StringQueryTrim(c, "q")
	isActive := params.StringQueryTrim(c, "is_active")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.announcements.List(c.Request.Context(), q, isActive, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": items, "meta": gin.H{"q": q, "is_active": isActive, "limit": limit, "offset": offset, "total": total}})
}

func (h *AdminMasterHandler) CreateAnnouncement(c *gin.Context) {
	var req adminAnnouncementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.Category = strings.TrimSpace(req.Category)
	req.TargetLevelID = strings.TrimSpace(req.TargetLevelID)
	req.TargetGroupID = strings.TrimSpace(req.TargetGroupID)
	req.TargetStudentID = strings.TrimSpace(req.TargetStudentID)

	if req.Title == "" || req.Body == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "title/body required"}})
		return
	}
	if targetCount(req) > 1 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "only one target is allowed (level/group/student)"}})
		return
	}

	publishedAt, err := parseRFC3339Optional(req.PublishedAt)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid published_at (RFC3339)"}})
		return
	}
	expiresAt, err := parseRFC3339Optional(req.ExpiresAt)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid expires_at (RFC3339)"}})
		return
	}
	if publishedAt != nil && expiresAt != nil && expiresAt.Before(*publishedAt) {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "expires_at must be >= published_at"}})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	item, err := h.announcements.Create(
		c.Request.Context(),
		req.Title,
		req.Body,
		req.Category,
		isActive,
		publishedAt,
		expiresAt,
		req.TargetLevelID,
		req.TargetGroupID,
		req.TargetStudentID,
		middleware.GetUserID(c),
	)
	if err != nil {
		switch pgerr.Code(err) {
		case pgerr.CodeForeignKeyViolation:
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid target reference"}})
			return
		case "22P02":
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid uuid format"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(201, gin.H{"data": item})
}

func (h *AdminMasterHandler) GetAnnouncement(c *gin.Context) {
	item, ok, err := h.announcements.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": item})
}

func (h *AdminMasterHandler) UpdateAnnouncement(c *gin.Context) {
	var req adminAnnouncementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.Category = strings.TrimSpace(req.Category)
	req.TargetLevelID = strings.TrimSpace(req.TargetLevelID)
	req.TargetGroupID = strings.TrimSpace(req.TargetGroupID)
	req.TargetStudentID = strings.TrimSpace(req.TargetStudentID)

	if req.Title == "" || req.Body == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "title/body required"}})
		return
	}
	if targetCount(req) > 1 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "only one target is allowed (level/group/student)"}})
		return
	}

	publishedAt, err := parseRFC3339Optional(req.PublishedAt)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid published_at (RFC3339)"}})
		return
	}
	expiresAt, err := parseRFC3339Optional(req.ExpiresAt)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid expires_at (RFC3339)"}})
		return
	}
	if publishedAt != nil && expiresAt != nil && expiresAt.Before(*publishedAt) {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "expires_at must be >= published_at"}})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	item, ok, err := h.announcements.Update(
		c.Request.Context(),
		c.Param("id"),
		req.Title,
		req.Body,
		req.Category,
		isActive,
		publishedAt,
		expiresAt,
		req.TargetLevelID,
		req.TargetGroupID,
		req.TargetStudentID,
	)
	if err != nil {
		switch pgerr.Code(err) {
		case pgerr.CodeForeignKeyViolation:
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid target reference"}})
			return
		case "22P02":
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid uuid format"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": item})
}

func (h *AdminMasterHandler) DeleteAnnouncement(c *gin.Context) {
	ok, err := h.announcements.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

type blastAnnouncementReq struct {
	Channels []string `json:"channels"` // "email", "whatsapp"
}

func (h *AdminMasterHandler) BlastAnnouncement(c *gin.Context) {
	annID := c.Param("id")
	var req blastAnnouncementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	ann, ok, err := h.announcements.Get(c.Request.Context(), annID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "announcement not found"}})
		return
	}

	// Fetch target students
	students, _, err := h.students.ListByTarget(c.Request.Context(), ann.TargetLevelID, ann.TargetGroupID, ann.TargetStudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to fetch target students"}})
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

	for _, s := range students {
		if useEmail && s.Email != "" {
			err := h.notif.SendEmail(c.Request.Context(), s.Email, ann.Title, ann.Body)
			if err != nil {
				failedCount++
			} else {
				sentCount++
			}
		}
		if useWA && s.Phone != "" {
			err := h.notif.SendWhatsApp(c.Request.Context(), s.Phone, fmt.Sprintf("*%s*\n\n%s", ann.Title, ann.Body))
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

func (h *AdminMasterHandler) ListTeachers(c *gin.Context) {
	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.teachers.List(c.Request.Context(), q, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items, "meta": gin.H{"q": q, "limit": limit, "offset": offset, "total": total}})
}

func (h *AdminMasterHandler) ListStudents(c *gin.Context) {
	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.students.List(c.Request.Context(), q, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items, "meta": gin.H{"q": q, "limit": limit, "offset": offset, "total": total}})
}

func (h *AdminMasterHandler) GetTeacher(c *gin.Context) {
	it, ok, err := h.teachers.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) GetTeacherSubjects(c *gin.Context) {
	teacherID := c.Param("id")
	if _, ok, err := h.teachers.Get(c.Request.Context(), teacherID); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	} else if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	items, err := h.teacherSubs.ListByTeacherID(c.Request.Context(), teacherID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type setTeacherSubjectsReq struct {
	SubjectIDs   []string `json:"subject_ids"`
	SubjectCodes []string `json:"subject_codes"`
}

func (h *AdminMasterHandler) SetTeacherSubjects(c *gin.Context) {
	teacherID := c.Param("id")
	if _, ok, err := h.teachers.Get(c.Request.Context(), teacherID); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	} else if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	var req setTeacherSubjectsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	subjectIDs := make([]string, 0)
	if len(req.SubjectCodes) > 0 {
		codes := make([]string, 0, len(req.SubjectCodes))
		for _, cc := range req.SubjectCodes {
			cc = strings.TrimSpace(cc)
			if cc == "" {
				continue
			}
			codes = append(codes, strings.ToUpper(cc))
		}
		ids, missing, err := h.lookups.SubjectIDsByCodesStrict(c.Request.Context(), codes)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if len(missing) > 0 {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "unknown subject_codes: " + strings.Join(missing, ",")}})
			return
		}
		subjectIDs = append(subjectIDs, ids...)
	} else {
		for _, id := range req.SubjectIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			subjectIDs = append(subjectIDs, id)
		}

		missing, err := h.subjects.MissingIDs(c.Request.Context(), subjectIDs)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if len(missing) > 0 {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "unknown subject_ids: " + strings.Join(missing, ",")}})
			return
		}
	}

	if err := h.teacherSubs.Replace(c.Request.Context(), teacherID, subjectIDs); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) GetTeacherGroups(c *gin.Context) {
	teacherID := c.Param("id")
	if _, ok, err := h.teachers.Get(c.Request.Context(), teacherID); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	} else if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	items, err := h.teacherGroups.ListByTeacherID(c.Request.Context(), teacherID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type setTeacherGroupsReq struct {
	GroupIDs   []string `json:"group_ids"`
	GroupNames []string `json:"group_names"`
}

func (h *AdminMasterHandler) SetTeacherGroups(c *gin.Context) {
	teacherID := c.Param("id")
	if _, ok, err := h.teachers.Get(c.Request.Context(), teacherID); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	} else if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	var req setTeacherGroupsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	groupIDs := make([]string, 0)
	if len(req.GroupNames) > 0 {
		ids, err := h.lookups.GroupIDsByNames(c.Request.Context(), req.GroupNames)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		groupIDs = ids
	} else {
		for _, id := range req.GroupIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			groupIDs = append(groupIDs, id)
		}
	}

	if err := h.teacherGroups.Replace(c.Request.Context(), teacherID, groupIDs); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) GetTeacherLevels(c *gin.Context) {
	teacherID := c.Param("id")
	if _, ok, err := h.teachers.Get(c.Request.Context(), teacherID); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	} else if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	items, err := h.teacherLevels.ListByTeacherID(c.Request.Context(), teacherID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type setTeacherLevelsReq struct {
	LevelIDs   []string `json:"level_ids"`
	LevelNames []string `json:"level_names"`
}

func (h *AdminMasterHandler) SetTeacherLevels(c *gin.Context) {
	teacherID := c.Param("id")
	if _, ok, err := h.teachers.Get(c.Request.Context(), teacherID); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	} else if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	var req setTeacherLevelsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	levelIDs := make([]string, 0)
	if len(req.LevelNames) > 0 {
		ids, err := h.lookups.LevelIDsByNames(c.Request.Context(), req.LevelNames)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		levelIDs = ids
	} else {
		for _, id := range req.LevelIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			levelIDs = append(levelIDs, id)
		}
	}

	if err := h.teacherLevels.Replace(c.Request.Context(), teacherID, levelIDs); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

type updateTeacherReq struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	NIP      string  `json:"nip"`
	Jenjang  string  `json:"jenjang"`
	IsActive *bool   `json:"is_active"`
	SchoolID *string `json:"school_id"`
}

type createTeacherReq struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	NIP        string `json:"nip"`
	Jenjang    string `json:"jenjang"`
	MapelCodes string `json:"mapel_codes"`
	GroupNames string `json:"group_names"`
	LevelNames string  `json:"level_names"`
	SchoolID   *string `json:"school_id"`
}

func (h *AdminMasterHandler) CreateTeacher(c *gin.Context) {
	var req createTeacherReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "username/password/name required"}})
		return
	}

	hash, err := authsvc.HashPassword(strings.TrimSpace(req.Password))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	subjectCodes := splitCSVUpper(req.MapelCodes)
	subjectIDs, err := h.lookups.SubjectIDsByCodes(c.Request.Context(), subjectCodes)
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error lookup subjects"}})
		return
	}

	groupNames := splitCSVTrim(req.GroupNames)
	groupIDs, err := h.lookups.GroupIDsByNames(c.Request.Context(), groupNames)
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error lookup groups"}})
		return
	}

	levelNames := splitCSVTrim(req.LevelNames)
	levelIDs, err := h.lookups.LevelIDsByNames(c.Request.Context(), levelNames)
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error lookup levels"}})
		return
	}

	teacherID, _, err := h.teachers.CreateTeacher(
		c.Request.Context(),
		strings.TrimSpace(req.Username),
		hash,
		strings.TrimSpace(req.Password),
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.Email),
		strings.TrimSpace(req.Phone),
		strings.TrimSpace(req.NIP),
		strings.TrimSpace(req.Jenjang),
		"",
		subjectIDs,
		groupIDs,
		levelIDs,
		req.SchoolID,
	)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error: " + err.Error()}})
		return
	}

	it, ok, err := h.teachers.Get(c.Request.Context(), teacherID)
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "created teacher not found"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *AdminMasterHandler) UpdateTeacher(c *gin.Context) {
	var req updateTeacherReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Name) == "" || req.IsActive == nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "username/name/is_active required"}})
		return
	}
	var passHash string
	if strings.TrimSpace(req.Password) != "" {
		hh, err := authsvc.HashPassword(strings.TrimSpace(req.Password))
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		passHash = hh
	}
	it, ok, err := h.teachers.UpdateTeacher(
		c.Request.Context(),
		c.Param("id"),
		strings.TrimSpace(req.Username),
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.Email),
		strings.TrimSpace(req.Phone),
		strings.TrimSpace(req.NIP),
		strings.TrimSpace(req.Jenjang),
		*req.IsActive,
		passHash,
		strings.TrimSpace(req.Password),
		req.SchoolID,
	)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) DeleteTeacher(c *gin.Context) {
	ok, err := h.teachers.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) GetStudent(c *gin.Context) {
	it, ok, err := h.students.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

type updateStudentReq struct {
	Username      string  `json:"username"`
	Password      string  `json:"password"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	NIS           string  `json:"nis"`
	ParticipantNo string  `json:"participant_no"`
	Jenjang       string  `json:"jenjang"`
	ProgramID     string  `json:"program_id"`
	LevelID       string  `json:"level_id"`
	GroupID       string  `json:"group_id"`
	IsActive      *bool   `json:"is_active"`
	SchoolID      *string `json:"school_id"`
}

type createStudentReq struct {
	Username      string  `json:"username"`
	Password      string  `json:"password"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	NIS           string  `json:"nis"`
	ParticipantNo string  `json:"participant_no"`
	Jenjang       string  `json:"jenjang"`
	ProgramID     string  `json:"program_id"`
	LevelID       string  `json:"level_id"`
	GroupID       string  `json:"group_id"`
	SchoolID      *string `json:"school_id"`
}

func (h *AdminMasterHandler) CreateStudent(c *gin.Context) {
	var req createStudentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.NIS) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "username/password/name/nis required"}})
		return
	}

	hash, err := authsvc.HashPassword(strings.TrimSpace(req.Password))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	studentID, _, err := h.students.CreateStudent(
		c.Request.Context(),
		strings.TrimSpace(req.Username),
		hash,
		strings.TrimSpace(req.Password),
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.Email),
		strings.TrimSpace(req.Phone),
		strings.TrimSpace(req.NIS),
		strings.TrimSpace(req.ParticipantNo),
		strings.TrimSpace(req.Jenjang),
		strings.TrimSpace(req.ProgramID),
		strings.TrimSpace(req.LevelID),
		strings.TrimSpace(req.GroupID),
		"",
		req.SchoolID,
	)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid reference"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	it, _, _ := h.students.Get(c.Request.Context(), studentID)
	c.JSON(201, gin.H{"data": it})
}

func (h *AdminMasterHandler) UpdateStudent(c *gin.Context) {
	var req updateStudentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.NIS) == "" || req.IsActive == nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "username/name/nis/is_active required"}})
		return
	}
	var passHash string
	if strings.TrimSpace(req.Password) != "" {
		hh, err := authsvc.HashPassword(strings.TrimSpace(req.Password))
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		passHash = hh
	}
	it, ok, err := h.students.UpdateStudent(
		c.Request.Context(),
		c.Param("id"),
		strings.TrimSpace(req.Username),
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.Email),
		strings.TrimSpace(req.Phone),
		strings.TrimSpace(req.NIS),
		strings.TrimSpace(req.ParticipantNo),
		strings.TrimSpace(req.Jenjang),
		strings.TrimSpace(req.ProgramID),
		strings.TrimSpace(req.LevelID),
		strings.TrimSpace(req.GroupID),
		*req.IsActive,
		passHash,
		strings.TrimSpace(req.Password),
		req.SchoolID,
	)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid reference"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *AdminMasterHandler) DeleteStudent(c *gin.Context) {
	ok, err := h.students.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) ListPendingRegistrations(c *gin.Context) {
	items, err := h.registrations.ListPending(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

func (h *AdminMasterHandler) ListRegistrations(c *gin.Context) {
	status := params.StringQueryTrim(c, "status")
	role := params.StringQueryTrim(c, "role")
	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.registrations.List(c.Request.Context(), status, role, q, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": items, "meta": gin.H{"status": status, "role": role, "q": q, "limit": limit, "offset": offset, "total": total}})
}

func (h *AdminMasterHandler) GetRegistration(c *gin.Context) {
	it, ok, err := h.registrations.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	// password hash is excluded by json tag
	c.JSON(200, gin.H{"data": it})
}

type patchRegistrationReq struct {
	Role        *string `json:"role"`
	Username    *string `json:"username"`
	Name        *string `json:"name"`
	Email       *string `json:"email"`
	NIS         *string `json:"nis"`
	NIP         *string `json:"nip"`
	ProgramCode *string `json:"program_code"`
	LevelName   *string `json:"level_name"`
	GroupName   *string `json:"group_name"`
	MapelCodes  *string `json:"mapel_codes"`
}

func (h *AdminMasterHandler) PatchRegistration(c *gin.Context) {
	var req patchRegistrationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	it, ok, err := h.registrations.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if it.Status != "pending" {
		c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "only pending registrations can be edited"}})
		return
	}

	applyStr := func(dst *string, src *string) {
		if src == nil {
			return
		}
		*dst = strings.TrimSpace(*src)
	}

	applyStr(&it.Role, req.Role)
	originalUsername := it.Username
	applyStr(&it.Username, req.Username)
	applyStr(&it.Name, req.Name)
	applyStr(&it.Email, req.Email)
	applyStr(&it.NIS, req.NIS)
	applyStr(&it.NIP, req.NIP)
	applyStr(&it.ProgramCode, req.ProgramCode)
	applyStr(&it.LevelName, req.LevelName)
	applyStr(&it.GroupName, req.GroupName)
	applyStr(&it.MapelCodes, req.MapelCodes)

	if it.Role != "student" && it.Role != "teacher" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "role must be student or teacher"}})
		return
	}
	if it.Username == "" || it.Name == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "username and name required"}})
		return
	}
	if it.Role == "student" && it.NIS == "" {
		if it.NISN != "" {
			it.NIS = it.NISN
		} else if it.Username != "" {
			it.NIS = it.Username
		} else {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "nis required for student"}})
			return
		}
	}

	// If username changed, validate it doesn't collide with existing users.
	if h.users != nil && it.Username != "" && it.Username != originalUsername {
		if _, ok, err := h.users.GetByUsername(c.Request.Context(), it.Username); err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		} else if ok {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "username already exists"}})
			return
		}
	}

	updated, ok, err := h.registrations.UpdatePending(c.Request.Context(), it)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	c.JSON(200, gin.H{"data": updated})
}

type decideReq struct {
	Note string `json:"note"`
}

type actionError struct {
	status  int
	code    string
	message string
}

func (e *actionError) Error() string { return e.message }

func newActionError(status int, code string, message string) error {
	return &actionError{status: status, code: code, message: message}
}

func writeActionError(c *gin.Context, err error) {
	var ae *actionError
	if errors.As(err, &ae) {
		c.JSON(ae.status, gin.H{"error": gin.H{"code": ae.code, "message": ae.message}})
		return
	}
	c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
}

func findExistingTeacherTx(ctx context.Context, tx pgx.Tx, username string, nip string) (teacherID string, userID string, ok bool, err error) {
	const q = `
SELECT t.id::text, u.id::text
FROM teachers t
JOIN users u ON u.id = t.user_id
WHERE u.username = $1
   OR (NULLIF($2,'') IS NOT NULL AND t.nip = $2)
LIMIT 1`
	if err := tx.QueryRow(ctx, q, strings.TrimSpace(username), strings.TrimSpace(nip)).Scan(&teacherID, &userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", false, nil
		}
		return "", "", false, fmt.Errorf("find existing teacher: %w", err)
	}
	return teacherID, userID, true, nil
}

func findExistingStudentTx(ctx context.Context, tx pgx.Tx, username string, nis string) (studentID string, userID string, ok bool, err error) {
	const q = `
SELECT s.id::text, u.id::text
FROM students s
JOIN users u ON u.id = s.user_id
WHERE u.username = $1
   OR (NULLIF($2,'') IS NOT NULL AND s.nis = $2)
LIMIT 1`
	if err := tx.QueryRow(ctx, q, strings.TrimSpace(username), strings.TrimSpace(nis)).Scan(&studentID, &userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", false, nil
		}
		return "", "", false, fmt.Errorf("find existing student: %w", err)
	}
	return studentID, userID, true, nil
}

func (h *AdminMasterHandler) approveRegistration(ctx context.Context, id string, note string) error {
	note = strings.TrimSpace(note)

	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin approval transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const sel = `
SELECT role, username, name, COALESCE(email,''), COALESCE(phone,''), COALESCE(password_hash,''), COALESCE(password_plain,''),
       COALESCE(nis,''), COALESCE(nip,''), COALESCE(program_code,''), COALESCE(level_name,''), COALESCE(group_name,''), COALESCE(mapel_codes,''),
       COALESCE(google_id,''), COALESCE(nisn,'')
FROM registration_requests
WHERE id = $1 AND status = 'pending'
FOR UPDATE`
	var role, username, name, email, phone, passwordHash, passwordPlain, nis, nip, programCode, levelName, groupName, mapelCodes, googleID, nisn string
	if err := tx.QueryRow(ctx, sel, id).Scan(&role, &username, &name, &email, &phone, &passwordHash, &passwordPlain, &nis, &nip, &programCode, &levelName, &groupName, &mapelCodes, &googleID, &nisn); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Distinguish between missing id and non-pending status.
			if it, ok, gerr := h.registrations.Get(ctx, id); gerr == nil && ok {
				if it.Status != "pending" {
					return newActionError(409, "conflict", "registration is not pending")
				}
				return newActionError(404, "not_found", "not found")
			}
			return newActionError(404, "not_found", "not found")
		}
		return fmt.Errorf("load pending registration: %w", err)
	}

	// If no password_hash and no google_id (manual form registration), auto-generate
	// a temporary password equal to the username so the account is usable.
	if strings.TrimSpace(passwordHash) == "" && strings.TrimSpace(googleID) == "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(username), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generate fallback password: %w", err)
		}
		passwordHash = string(hashed)
		passwordPlain = username
	}

	switch role {
	case "teacher":
		codes := splitCSVUpper(mapelCodes)
		var subjectIDs []string
		if len(codes) > 0 {
			sids, missing, err := h.lookups.SubjectIDsByCodesStrict(ctx, codes)
			if err != nil {
				return fmt.Errorf("lookup teacher subjects: %w", err)
			}
			if len(missing) > 0 {
				return newActionError(400, "bad_request", "unknown mapel_codes: "+strings.Join(missing, ","))
			}
			subjectIDs = sids
		}
		if _, spErr := tx.Exec(ctx, `SAVEPOINT sp_create_user`); spErr != nil {
			return fmt.Errorf("create savepoint teacher approval: %w", spErr)
		}
		if teacherID, userID, err := h.teachers.CreateTeacherTx(ctx, tx, username, passwordHash, passwordPlain, name, email, phone, nip, "", googleID, subjectIDs, nil, nil, nil); err != nil {
			if pgerr.Code(err) == pgerr.CodeUniqueViolation {
				// Roll back the failed insert attempt to clear the aborted-tx state.
				_, _ = tx.Exec(ctx, `ROLLBACK TO SAVEPOINT sp_create_user`)

				// Idempotent approve: if the account was already created (e.g. from import/manual create),
				// mark the registration as approved instead of failing with "duplicate".
				existingTeacherID, existingUserID, ok, ferr := findExistingTeacherTx(ctx, tx, username, nip)
				if ferr != nil {
					return fmt.Errorf("find existing teacher after duplicate: %w", ferr)
				}
				if !ok {
					return newActionError(409, "conflict", "duplicate")
				}

				// Best-effort: link google_id if provided and user doesn't have one yet.
				if strings.TrimSpace(googleID) != "" {
					if _, uerr := tx.Exec(ctx, `
UPDATE users
SET google_id = NULLIF($1,''), updated_at = now()
WHERE id = $2 AND (google_id IS NULL OR google_id = '')`, googleID, existingUserID); uerr != nil {
						if pgerr.Code(uerr) == pgerr.CodeUniqueViolation {
							return newActionError(409, "conflict", "google_id already linked")
						}
						return fmt.Errorf("link teacher google_id: %w", uerr)
					}
				}
				// Ensure the account is active.
				_, _ = tx.Exec(ctx, `UPDATE users SET is_active = true, updated_at = now() WHERE id = $1`, existingUserID)

				// Best-effort: ensure subjects mapping exists (no-op if already there).
				if len(subjectIDs) > 0 {
					const insMap = `INSERT INTO teacher_subjects (teacher_id, subject_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
					for _, sid := range subjectIDs {
						if _, mapErr := tx.Exec(ctx, insMap, existingTeacherID, sid); mapErr != nil {
							return fmt.Errorf("ensure teacher subject mapping: %w", mapErr)
						}
					}
				}
			} else {
				return fmt.Errorf("create teacher account from registration: %w", err)
			}
		} else {
			_ = teacherID
			_ = userID
			_, _ = tx.Exec(ctx, `RELEASE SAVEPOINT sp_create_user`)
		}
	case "student":
		// Fallback: if NIS is empty, try NISN, then fallback to Username.
		if strings.TrimSpace(nis) == "" {
			if strings.TrimSpace(nisn) != "" {
				nis = nisn
			} else if strings.TrimSpace(username) != "" {
				nis = username
			} else {
				return newActionError(400, "bad_request", "nis required")
			}
		}

		var programID, levelID, groupID string
		if strings.TrimSpace(programCode) != "" {
			id, ok, err := h.lookups.ProgramIDByCode(ctx, programCode)
			if err != nil {
				return fmt.Errorf("lookup program by code: %w", err)
			}
			if !ok {
				// Log or ignore if lookup fails; don't block approval
				programID = ""
			} else {
				programID = id
			}
		}
		if strings.TrimSpace(levelName) != "" {
			id, ok, err := h.lookups.LevelIDByName(ctx, levelName)
			if err != nil {
				return fmt.Errorf("lookup level by name: %w", err)
			}
			if !ok {
				levelID = ""
			} else {
				levelID = id
			}
		}
		if strings.TrimSpace(groupName) != "" {
			id, ok, err := h.lookups.GroupIDByName(ctx, groupName)
			if err != nil {
				return fmt.Errorf("lookup group by name: %w", err)
			}
			if !ok {
				groupID = ""
			} else {
				groupID = id
			}
		}

		if _, spErr := tx.Exec(ctx, `SAVEPOINT sp_create_user`); spErr != nil {
			return fmt.Errorf("create savepoint student approval: %w", spErr)
		}
		if studentID, userID, err := h.students.CreateStudentTx(ctx, tx, username, passwordHash, passwordPlain, name, email, phone, nis, "", "", programID, levelID, groupID, googleID, nil); err != nil {
			if pgerr.Code(err) == pgerr.CodeUniqueViolation {
				_, _ = tx.Exec(ctx, `ROLLBACK TO SAVEPOINT sp_create_user`)

				existingStudentID, existingUserID, ok, ferr := findExistingStudentTx(ctx, tx, username, nis)
				if ferr != nil {
					return fmt.Errorf("find existing student after duplicate: %w", ferr)
				}
				if !ok {
					return newActionError(409, "conflict", "duplicate")
				}

				if strings.TrimSpace(googleID) != "" {
					if _, uerr := tx.Exec(ctx, `
UPDATE users
SET google_id = NULLIF($1,''), updated_at = now()
WHERE id = $2 AND (google_id IS NULL OR google_id = '')`, googleID, existingUserID); uerr != nil {
						if pgerr.Code(uerr) == pgerr.CodeUniqueViolation {
							return newActionError(409, "conflict", "google_id already linked")
						}
						return fmt.Errorf("link student google_id: %w", uerr)
					}
				}
				_, _ = tx.Exec(ctx, `UPDATE users SET is_active = true, updated_at = now() WHERE id = $1`, existingUserID)
				_ = existingStudentID
			} else if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
				return newActionError(409, "conflict", "invalid reference")
			} else {
				return fmt.Errorf("create student account from registration: %w", err)
			}
		} else {
			_ = studentID
			_ = userID
			_, _ = tx.Exec(ctx, `RELEASE SAVEPOINT sp_create_user`)
		}
	default:
		return newActionError(400, "bad_request", "invalid role")
	}

	const upd = `
UPDATE registration_requests
SET status = 'approved', note = NULLIF($2,''), decided_at = now()
WHERE id = $1`
	if _, err := tx.Exec(ctx, upd, id, note); err != nil {
		return fmt.Errorf("mark registration approved: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit registration approval: %w", err)
	}

	return nil
}

func (h *AdminMasterHandler) ApproveRegistration(c *gin.Context) {
	var req decideReq
	_ = c.ShouldBindJSON(&req)

	if err := h.approveRegistration(c.Request.Context(), c.Param("id"), req.Note); err != nil {
		writeActionError(c, err)
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

type bulkApproveRegistrationsReq struct {
	Role  string `json:"role"`
	Q     string `json:"q"`
	Note  string `json:"note"`
	Limit *int   `json:"limit"`
}

func (h *AdminMasterHandler) BulkApproveRegistrations(c *gin.Context) {
	var req bulkApproveRegistrationsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	role := strings.TrimSpace(req.Role)
	if role != "" && role != "student" && role != "teacher" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "role must be student or teacher"}})
		return
	}

	limit := 200
	if req.Limit != nil {
		limit = *req.Limit
	}
	if limit < 1 || limit > 500 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "limit must be between 1 and 500"}})
		return
	}

	items, total, err := h.registrations.List(c.Request.Context(), "pending", role, strings.TrimSpace(req.Q), limit, 0)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	failures := make([]gin.H, 0)
	approved := 0
	failed := 0
	const maxFailureDetails = 25

	for _, item := range items {
		if err := h.approveRegistration(c.Request.Context(), item.ID, req.Note); err != nil {
			failed++
			message := "internal error"
			var ae *actionError
			if errors.As(err, &ae) {
				message = ae.message
			}
			if len(failures) < maxFailureDetails {
				failures = append(failures, gin.H{
					"id":       item.ID,
					"role":     item.Role,
					"username": item.Username,
					"name":     item.Name,
					"message":  message,
				})
			}
			continue
		}
		approved++
	}

	processed := len(items)
	remaining := total - processed
	if remaining < 0 {
		remaining = 0
	}

	c.JSON(200, gin.H{"data": gin.H{
		"matched_total":     total,
		"processed":         processed,
		"approved":          approved,
		"failed":            failed,
		"remaining":         remaining,
		"failure_details":   failures,
		"failure_truncated": failed > len(failures),
	}})
}

func (h *AdminMasterHandler) RejectRegistration(c *gin.Context) {
	var req decideReq
	_ = c.ShouldBindJSON(&req)

	ctx := c.Request.Context()
	id := c.Param("id")

	ok, err := h.registrations.DecidePending(ctx, id, "rejected", strings.TrimSpace(req.Note))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		if it, ok2, gerr := h.registrations.Get(ctx, id); gerr == nil && ok2 {
			if it.Status != "pending" {
				c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "registration is not pending"}})
				return
			}
		}
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *AdminMasterHandler) ResetRegistration(c *gin.Context) {
	var req decideReq
	_ = c.ShouldBindJSON(&req)

	ctx := c.Request.Context()
	id := c.Param("id")

	err := h.registrations.Decide(ctx, id, "pending", strings.TrimSpace(req.Note))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

// Excel templates + import

func (h *AdminMasterHandler) TeachersTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	f.SetSheetName("Sheet1", "DATA")
	_ = f.SetSheetRow("DATA", "A1", &[]string{"username", "password", "nama", "email", "nip", "mapel_codes"})
	_ = f.SetSheetRow("DATA", "A2", &[]string{"guru01", "password123", "Budi Santoso", "budi@example.com", "1987654321", "MAT,IPA"})

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to generate template"}})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=template-guru.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *AdminMasterHandler) StudentsTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	f.SetSheetName("Sheet1", "DATA")
	_ = f.SetSheetRow("DATA", "A1", &[]string{"username", "password", "nama", "email", "nis", "participant_no", "jenjang", "program_code", "level", "group", "phone"})
	_ = f.SetSheetRow("DATA", "A2", &[]string{"siswa01", "password123", "Siti Aminah", "siti@example.com", "1234567890", "ASAT-2026-001", "SMA", "IPA", "Kelas 10", "X IPA-1", "6281234567890"})

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to generate template"}})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=template-siswa.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

type importResult struct {
	Inserted int           `json:"inserted"`
	Errors   []importError `json:"errors"`
}

type importError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

func (h *AdminMasterHandler) ImportTeachers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "file is required"}})
		return
	}

	fh, err := file.Open()
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "failed to open file"}})
		return
	}
	defer fh.Close()

	f, err := excelize.OpenReader(fh)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid xlsx"}})
		return
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows("DATA")
	if err != nil || len(rows) < 2 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "missing DATA sheet or rows"}})
		return
	}

	res := importResult{Inserted: 0, Errors: []importError{}}

	// Expect fixed headers (row 1).
	// username, password, nama, email, nip, mapel_codes
	for i := 1; i < len(rows); i++ {
		rowNum := i + 1
		r := rows[i]
		get := func(idx int) string {
			if idx >= len(r) {
				return ""
			}
			return strings.TrimSpace(r[idx])
		}

		username := get(0)
		password := get(1)
		name := get(2)
		email := get(3)
		nip := get(4)
		mapelCodes := get(5)
		phone := get(6)

		if username == "" || password == "" || name == "" {
			res.Errors = append(res.Errors, importError{Row: rowNum, Message: "username/password/nama required"})
			continue
		}

		hash, err := authsvc.HashPassword(password)
		if err != nil {
			res.Errors = append(res.Errors, importError{Row: rowNum, Message: "failed to hash password"})
			continue
		}

		codes := splitCSVUpper(mapelCodes)
		subjectIDs, err := h.lookups.SubjectIDsByCodes(c.Request.Context(), codes)
		if err != nil {
			res.Errors = append(res.Errors, importError{Row: rowNum, Message: "failed subject lookup"})
			continue
		}

		_, _, err = h.teachers.CreateTeacher(c.Request.Context(), username, hash, password, name, email, phone, nip, "", "", subjectIDs, nil, nil, nil)
		if err != nil {
			res.Errors = append(res.Errors, importError{Row: rowNum, Message: "insert failed (duplicate?)"})
			continue
		}

		res.Inserted++
	}

	c.JSON(200, gin.H{"data": res})
}

func (h *AdminMasterHandler) ImportStudents(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "file is required"}})
		return
	}

	fh, err := file.Open()
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "failed to open file"}})
		return
	}
	defer fh.Close()

	f, err := excelize.OpenReader(fh)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid xlsx"}})
		return
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows("DATA")
	if err != nil || len(rows) < 2 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "missing DATA sheet or rows"}})
		return
	}

	res := importResult{Inserted: 0, Errors: []importError{}}

	// Expect fixed headers (row 1).
	// username, password, nama, email, nis, participant_no, jenjang, program_code, level, group, phone
	for i := 1; i < len(rows); i++ {
		rowNum := i + 1
		r := rows[i]
		get := func(idx int) string {
			if idx >= len(r) {
				return ""
			}
			return strings.TrimSpace(r[idx])
		}

		username := get(0)
		password := get(1)
		name := get(2)
		email := get(3)
		nis := get(4)
		participantNo := get(5)
		jenjang := get(6)
		programCode := get(7)
		levelName := get(8)
		groupName := get(9)
		phone := get(10)

		if username == "" || password == "" || name == "" || nis == "" {
			res.Errors = append(res.Errors, importError{Row: rowNum, Message: "username/password/nama/nis required"})
			continue
		}

		hash, err := authsvc.HashPassword(password)
		if err != nil {
			res.Errors = append(res.Errors, importError{Row: rowNum, Message: "failed to hash password"})
			continue
		}

		var programID, levelID, groupID string
		if programCode != "" {
			if id, ok, err := h.lookups.ProgramIDByCode(c.Request.Context(), programCode); err == nil && ok {
				programID = id
			}
		}
		if levelName != "" {
			if id, ok, err := h.lookups.LevelIDByName(c.Request.Context(), levelName); err == nil && ok {
				levelID = id
			}
		}
		if groupName != "" {
			if id, ok, err := h.lookups.GroupIDByName(c.Request.Context(), groupName); err == nil && ok {
				groupID = id
			}
		}

		_, _, err = h.students.CreateStudent(c.Request.Context(), username, hash, password, name, email, phone, nis, participantNo, jenjang, programID, levelID, groupID, "", nil)
		if err != nil {
			res.Errors = append(res.Errors, importError{Row: rowNum, Message: "insert failed (duplicate?)"})
			continue
		}

		res.Inserted++
	}

	c.JSON(200, gin.H{"data": res})
}

func (h *AdminMasterHandler) GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()

	var totalStudents, totalTeachers, totalQuestionSets, totalExams, pendingRegs int
	var globalAvg float64

	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM students`).Scan(&totalStudents)
	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'teacher'`).Scan(&totalTeachers)
	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM question_sets`).Scan(&totalQuestionSets)
	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exams`).Scan(&totalExams)
	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM registration_requests WHERE status = 'pending'`).Scan(&pendingRegs)
	h.pool.QueryRow(ctx, `SELECT COALESCE(AVG(score), 0) FROM exam_sessions WHERE status = 'submitted'`).Scan(&globalAvg)

	c.JSON(200, gin.H{
		"data": gin.H{
			"total_students":        totalStudents,
			"active_students":       totalStudents, // Technically all in 'students' are active, registration_requests are separate
			"pending_registrations": pendingRegs,
			"total_teachers":        totalTeachers,
			"total_question_sets":   totalQuestionSets,
			"total_exams":           totalExams,
			"average_score":         globalAvg,
		},
	})
}

func splitCSVUpper(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, strings.ToUpper(p))
	}
	return out
}

func splitCSVTrim(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
func (h *AdminMasterHandler) SwitchUserRole(c *gin.Context) {
	userID := c.Param("id")
	ctx := c.Request.Context()

	u, ok, err := h.users.GetByID(ctx, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "user not found"}})
		return
	}

	newRole := ""
	if u.Role == "teacher" {
		newRole = "student"
	} else if u.Role == "student" {
		newRole = "teacher"
	} else {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "cannot switch role for " + u.Role}})
		return
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	defer tx.Rollback(ctx)

	// Update user role
	if _, err := tx.Exec(ctx, `UPDATE users SET role = $1, updated_at = now() WHERE id = $2`, newRole, userID); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to update user role"}})
		return
	}

	if newRole == "teacher" {
		var exists bool
		_ = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teachers WHERE user_id = $1)`, userID).Scan(&exists)
		if !exists {
			// nip is UNIQUE nullable — use NULL so no collision
			if _, err := tx.Exec(ctx, `INSERT INTO teachers (user_id, nip, jenjang) VALUES ($1, NULL, '')`, userID); err != nil {
				c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to create teacher metadata: " + err.Error()}})
				return
			}
		}
	} else if newRole == "student" {
		var exists bool
		_ = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM students WHERE user_id = $1)`, userID).Scan(&exists)
		if !exists {
			// nis is NOT NULL + UNIQUE — use a unique placeholder derived from userID
			nisPlaceholder := "SW-" + userID
			if _, err := tx.Exec(ctx, `INSERT INTO students (user_id, nis, jenjang, updated_at) VALUES ($1, $2, '', now())`, userID, nisPlaceholder); err != nil {
				c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to create student metadata: " + err.Error()}})
				return
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to commit transaction"}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true, "new_role": newRole}})
}
