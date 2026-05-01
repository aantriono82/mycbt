package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/storage"
)

type UploadsHandler struct {
	store storage.ObjectStore
}

func NewUploadsHandler(store storage.ObjectStore) *UploadsHandler {
	if store == nil {
		store = storage.NewLocalObjectStore("uploads")
	}
	return &UploadsHandler{store: store}
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

	userID := middleware.GetUserID(c)
	suffix := randHex(6)
	filename := fmt.Sprintf("img_%s_%s", userID, suffix)
	url, err := uploadImageToStore(c.Request.Context(), h.store, file, "editor-images", filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to save uploaded file"}})
		return
	}

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
