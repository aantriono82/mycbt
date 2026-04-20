package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/repo/loginlogrepo"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/authsvc"
)

type AuthHandler struct {
	auth      *authsvc.Service
	users     *userrepo.Repo
	loginLogs *loginlogrepo.Repo
}

func NewAuthHandler(auth *authsvc.Service, users *userrepo.Repo, loginLogs *loginlogrepo.Repo) *AuthHandler {
	return &AuthHandler{auth: auth, users: users, loginLogs: loginLogs}
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	token, exp, user, err := h.auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case authsvc.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "invalid_credentials", "message": "invalid credentials"}})
		case authsvc.ErrUserInactive:
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "user_inactive", "message": "user inactive"}})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		}
		return
	}

	// Best-effort: do not block login if activity log insert fails.
	if h.loginLogs != nil {
		uID := user.ID
		_ = h.loginLogs.Insert(c.Request.Context(), loginlogrepo.LoginLog{
			UserID:    &uID,
			Username:  user.Username,
			Role:      user.Role,
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			LoggedInAt: time.Now().UTC(),
		})
		// Enforce retention default: keep last 30 days.
		_, _ = h.loginLogs.PruneOlderThan(c.Request.Context(), 30)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_at":   exp.Format(time.RFC3339),
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"role":     user.Role,
				"name":     user.Name,
				"email":    user.Email,
				"photo_url": user.PhotoURL,
			},
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "unauthorized"}})
		return
	}

	u, ok, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "unauthorized"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":        u.ID,
			"username":  u.Username,
			"role":      u.Role,
			"name":      u.Name,
			"email":     u.Email,
			"photo_url":  u.PhotoURL,
			"is_active": u.IsActive,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Stateless JWT for now.
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

func (h *AuthHandler) UploadPhoto(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "unauthorized"}})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file is required"}})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%v%v%v", userID, time.Now().Unix(), ext)
	dst := filepath.Join("./uploads/avatars", filename)

	if err := os.MkdirAll("./uploads/avatars", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to create directory"}})
		return
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to save file"}})
		return
	}

	photoURL := "/uploads/avatars/" + filename
	if err := h.users.UpdatePhoto(c.Request.Context(), userID, photoURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to update database"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"photo_url": photoURL,
		},
	})
}
