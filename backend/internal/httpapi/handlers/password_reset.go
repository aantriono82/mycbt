package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/service/authsvc"
)

type PasswordResetHandler struct {
	svc *authsvc.PasswordResetService
}

func NewPasswordResetHandler(svc *authsvc.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{svc: svc}
}

func (h *PasswordResetHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email tidak valid"})
		return
	}

	if err := h.svc.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses permintaan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tautan reset kata sandi telah dikirim ke email Anda."})
}

func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kata sandi Anda telah berhasil diatur ulang."})
}
