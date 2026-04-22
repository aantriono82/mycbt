package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/httpapi/middleware"
)

type UploadsHandler struct{}

func NewUploadsHandler() *UploadsHandler {
	return &UploadsHandler{}
}

func (h *UploadsHandler) UploadImage(c *gin.Context) {
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
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "unsupported image format (png/jpg/jpeg/gif/webp)"}})
		return
	}

	targetDir := filepath.Join("uploads", "editor-images")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to prepare upload directory"}})
		return
	}

	userID := middleware.GetUserID(c)
	suffix := randHex(6)
	filename := fmt.Sprintf("img_%s_%d_%s%s", userID, time.Now().UnixNano(), suffix, ext)
	targetPath := filepath.Join(targetDir, filename)
	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to save uploaded file"}})
		return
	}

	url := "/" + filepath.ToSlash(targetPath)
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"url": url,
		},
	})
}

func randHex(nbytes int) string {
	if nbytes <= 0 {
		return ""
	}
	buf := make([]byte, nbytes)
	if _, err := rand.Read(buf); err != nil {
		// Fallback to time-based value; still good enough for filename uniqueness.
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

