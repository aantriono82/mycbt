package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/repo/loginlogrepo"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/authsvc"
	"mycbt/backend/internal/storage"
)

type AuthHandler struct {
	auth      *authsvc.Service
	users     *userrepo.Repo
	loginLogs *loginlogrepo.Repo
	store     storage.ObjectStore
}

func NewAuthHandler(auth *authsvc.Service, users *userrepo.Repo, loginLogs *loginlogrepo.Repo, store storage.ObjectStore) *AuthHandler {
	if store == nil {
		store = storage.NewLocalObjectStore("uploads")
	}
	return &AuthHandler{auth: auth, users: users, loginLogs: loginLogs, store: store}
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
			UserID:     &uID,
			Username:   user.Username,
			Role:       user.Role,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
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
				"id":        user.ID,
				"username":  user.Username,
				"role":      user.Role,
				"name":      user.Name,
				"email":     user.Email,
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
			"photo_url": u.PhotoURL,
			"is_active": u.IsActive,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	rawToken := ""
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		rawToken = strings.TrimSpace(parts[1])
	}
	if rawToken != "" && h.auth != nil {
		_ = h.auth.RevokeToken(c.Request.Context(), rawToken)
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

type updateMeReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "unauthorized"}})
		return
	}

	var req updateMeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "name is required"}})
		return
	}

	if err := h.users.UpdateProfile(c.Request.Context(), userID, name, email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to update profile"}})
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
			"photo_url": u.PhotoURL,
			"is_active": u.IsActive,
		},
	})
}

type changePasswordReq struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	Confirmation    string `json:"password_confirmation"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "unauthorized"}})
		return
	}

	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	current := strings.TrimSpace(req.CurrentPassword)
	next := strings.TrimSpace(req.NewPassword)
	confirm := strings.TrimSpace(req.Confirmation)

	if current == "" || next == "" || confirm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "current_password, new_password, password_confirmation required"}})
		return
	}
	if len(next) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "new_password too short (min 8)"}})
		return
	}
	if next != confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "password confirmation mismatch"}})
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
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(current)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "invalid_credentials", "message": "current password invalid"}})
		return
	}
	if current == next {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "new password must be different"}})
		return
	}

	hash, err := authsvc.HashPassword(next)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to process password"}})
		return
	}
	if err := h.users.UpdatePassword(c.Request.Context(), userID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to update password"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

func (h *AuthHandler) UploadPhoto(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "unauthorized"}})
		return
	}

	// Admin can upload for other users
	targetUserID := c.Query("target_user_id")
	if targetUserID != "" && targetUserID != userID {
		role := middleware.GetUserRole(c)
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "only admin can upload for others"}})
			return
		}
		userID = targetUserID
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file is required"}})
		return
	}

	filename := fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())
	photoURL, err := uploadImageToStore(c.Request.Context(), h.store, file, "avatars", filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to save file"}})
		return
	}
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
