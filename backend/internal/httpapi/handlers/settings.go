package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/repo/masterrepo"
)

type SettingsHandler struct {
	settings *masterrepo.SettingsRepo
}

func NewSettingsHandler(settings *masterrepo.SettingsRepo) *SettingsHandler {
	return &SettingsHandler{settings: settings}
}

func (h *SettingsHandler) GetSchoolIdentity(c *gin.Context) {
	data, err := h.settings.GetSchoolIdentity(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

type putSchoolIdentityReq struct {
	SchoolName    string `json:"school_name"`
	Address       string `json:"address"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Website       string `json:"website"`
	PrincipalName string `json:"principal_name"`
	LogoURL       string `json:"logo_url"`
}

func (h *SettingsHandler) PutSchoolIdentity(c *gin.Context) {
	var req putSchoolIdentityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.SchoolName) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "validation", "message": "school_name is required"}})
		return
	}

	data, err := h.settings.UpsertSchoolIdentity(c.Request.Context(), masterrepo.SchoolIdentity{
		SchoolName:    strings.TrimSpace(req.SchoolName),
		Address:       strings.TrimSpace(req.Address),
		Phone:         strings.TrimSpace(req.Phone),
		Email:         strings.TrimSpace(req.Email),
		Website:       strings.TrimSpace(req.Website),
		PrincipalName: strings.TrimSpace(req.PrincipalName),
		LogoURL:       strings.TrimSpace(req.LogoURL),
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *SettingsHandler) UploadSchoolLogo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file is required"}})
		return
	}
	if file.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file is empty"}})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "unsupported image format (png/jpg/jpeg/webp)"}})
		return
	}

	targetDir := filepath.Join("uploads", "logos")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to prepare upload directory"}})
		return
	}

	filename := fmt.Sprintf("school_logo_%d%s", time.Now().UnixNano(), ext)
	targetPath := filepath.Join(targetDir, filename)
	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to save uploaded file"}})
		return
	}

	identity, err := h.settings.GetSchoolIdentity(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	identity.LogoURL = "/" + filepath.ToSlash(targetPath)
	data, err := h.settings.UpsertSchoolIdentity(c.Request.Context(), identity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"logo_url": data.LogoURL,
		},
	})
}

func (h *SettingsHandler) GetSystem(c *gin.Context) {
	data, err := h.settings.GetSystem(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

type putSystemReq struct {
	Timezone            string `json:"timezone"`
	TokenRequired       bool   `json:"token_required"`
	AllowResetLogin     bool   `json:"allow_reset_login"`
	MaxActiveSessions   int    `json:"max_active_sessions"`
	AttendanceRequireIP bool   `json:"attendance_require_ip"`
}

func (h *SettingsHandler) PutSystem(c *gin.Context) {
	var req putSystemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.Timezone) == "" {
		req.Timezone = "Asia/Jakarta"
	}
	if req.MaxActiveSessions < 1 || req.MaxActiveSessions > 5 {
		c.JSON(400, gin.H{"error": gin.H{"code": "validation", "message": "max_active_sessions must be 1..5"}})
		return
	}

	data, err := h.settings.UpsertSystem(c.Request.Context(), masterrepo.SystemSettings{
		Timezone:            strings.TrimSpace(req.Timezone),
		TokenRequired:       req.TokenRequired,
		AllowResetLogin:     req.AllowResetLogin,
		MaxActiveSessions:   req.MaxActiveSessions,
		AttendanceRequireIP: req.AttendanceRequireIP,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *SettingsHandler) GetSMTP(c *gin.Context) {
	data, err := h.settings.GetSMTP(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

type putSMTPReq struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	From     string `json:"from"`
	UseTLS   bool   `json:"use_tls"`
}

func (h *SettingsHandler) PutSMTP(c *gin.Context) {
	var req putSMTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	data, err := h.settings.UpsertSMTP(c.Request.Context(), masterrepo.SMTPConfig{
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		From:     req.From,
		UseTLS:   req.UseTLS,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *SettingsHandler) GetWhatsApp(c *gin.Context) {
	data, err := h.settings.GetWhatsApp(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

type putWhatsAppReq struct {
	APIURL string `json:"api_url"`
	APIKey string `json:"api_key"`
	Sender string `json:"sender"`
}

func (h *SettingsHandler) PutWhatsApp(c *gin.Context) {
	var req putWhatsAppReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	data, err := h.settings.UpsertWhatsApp(c.Request.Context(), masterrepo.WhatsAppConfig{
		APIURL: req.APIURL,
		APIKey: req.APIKey,
		Sender: req.Sender,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}
