package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/httpapi/pgerr"
	"atigacbt/backend/internal/repo/masterrepo"
	"atigacbt/backend/internal/repo/userrepo"
	"atigacbt/backend/internal/service/authsvc"
)

type RegistrationPublicHandler struct {
	registrations *masterrepo.RegistrationRepo
	users         *userrepo.Repo
}

func NewRegistrationPublicHandler(regs *masterrepo.RegistrationRepo, users *userrepo.Repo) *RegistrationPublicHandler {
	return &RegistrationPublicHandler{registrations: regs, users: users}
}

type registerReq struct {
	Role        string `json:"role"` // student | teacher
	Username    string `json:"username"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	NIS         string `json:"nis"`
	NIP         string `json:"nip"`
	ProgramCode string `json:"program_code"`
	LevelName   string `json:"level_name"`
	GroupName   string `json:"group_name"`
	MapelCodes  string `json:"mapel_codes"`
}

func (h *RegistrationPublicHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	role := strings.TrimSpace(req.Role)
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)

	if role != "student" && role != "teacher" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "role must be student or teacher"}})
		return
	}
	if username == "" || password == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "username/password/name required"}})
		return
	}
	if len(password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "password too short (min 8)"}})
		return
	}
	if role == "student" && strings.TrimSpace(req.NIS) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "nis required"}})
		return
	}

	// Fast feedback for duplicate usernames.
	if h.users != nil {
		if _, ok, err := h.users.GetByUsername(c.Request.Context(), username); err == nil && ok {
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "conflict", "message": "username already exists"}})
			return
		}
	}

	hash, err := authsvc.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	id, err := h.registrations.Create(c.Request.Context(), masterrepo.RegistrationRequest{
		Role:         role,
		Username:     username,
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		NIS:          strings.TrimSpace(req.NIS),
		NIP:          strings.TrimSpace(req.NIP),
		ProgramCode:  strings.TrimSpace(req.ProgramCode),
		LevelName:    strings.TrimSpace(req.LevelName),
		GroupName:    strings.TrimSpace(req.GroupName),
		MapelCodes:   strings.TrimSpace(req.MapelCodes),
	})
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"id":     id,
			"status": "pending",
		},
	})
}

func (h *RegistrationPublicHandler) Status(c *gin.Context) {
	it, ok, err := h.registrations.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         it.ID,
			"role":       it.Role,
			"username":   it.Username,
			"name":       it.Name,
			"status":     it.Status,
			"note":       it.Note,
			"decided_at": it.DecidedAt,
		},
	})
}
