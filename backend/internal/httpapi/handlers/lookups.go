package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/repo/masterrepo"
)

type LookupsHandler struct {
	subjects *masterrepo.SubjectsRepo
	sessions *masterrepo.SessionsRepo
	levels   *masterrepo.LevelsRepo
	groups   *masterrepo.GroupsRepo
	students *masterrepo.StudentsRepo
	teachers *masterrepo.TeachersRepo
	programs *masterrepo.ProgramsRepo
}

func NewLookupsHandler(subjects *masterrepo.SubjectsRepo, sessions *masterrepo.SessionsRepo, levels *masterrepo.LevelsRepo, groups *masterrepo.GroupsRepo, students *masterrepo.StudentsRepo, teachers *masterrepo.TeachersRepo, programs *masterrepo.ProgramsRepo) *LookupsHandler {
	return &LookupsHandler{subjects: subjects, sessions: sessions, levels: levels, groups: groups, students: students, teachers: teachers, programs: programs}
}

func (h *LookupsHandler) ListSubjects(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	var (
		items []masterrepo.Subject
		err   error
	)
	if role == "teacher" {
		var ok bool
		items, ok, err = h.subjects.ListForTeacherUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "teacher not registered"}})
			return
		}
	} else {
		items, err = h.subjects.List(c.Request.Context())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
func (h *LookupsHandler) ListSessions(c *gin.Context) {
	items, err := h.sessions.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *LookupsHandler) ListPrograms(c *gin.Context) {
	items, err := h.programs.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *LookupsHandler) ListLevels(c *gin.Context) {

	var (
		items []masterrepo.Level
		err   error
	)
	items, err = h.levels.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *LookupsHandler) ListGroups(c *gin.Context) {

	var (
		items []masterrepo.Group
		err   error
	)
	items, err = h.groups.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *LookupsHandler) ListStudents(c *gin.Context) {

	var (
		items []masterrepo.Student
		total int
		err   error
	)
	q := c.Query("q")
	items, total, err = h.students.List(c.Request.Context(), q, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"total": total}})
}
func (h *LookupsHandler) ListTeachers(c *gin.Context) {
	items, _, err := h.teachers.List(c.Request.Context(), "", 500, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *LookupsHandler) ListMyAssignments(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if role != "teacher" {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{
			"levels":   []string{},
			"groups":   []string{},
			"subjects": []string{},
		}})
		return
	}

	ctx := c.Request.Context()
	levels, _, _ := h.levels.ListForTeacherUserID(ctx, userID)
	groups, _, _ := h.groups.ListForTeacherUserID(ctx, userID)
	subjects, _, _ := h.subjects.ListForTeacherUserID(ctx, userID)

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"levels":   levels,
		"groups":   groups,
		"subjects": subjects,
	}})
}
