package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type MaintenanceHandler struct {
	dbURL string
}

func NewMaintenanceHandler(dbURL string) *MaintenanceHandler {
	return &MaintenanceHandler{dbURL: dbURL}
}

// Backup handles database export using pg_dump
func (h *MaintenanceHandler) Backup(c *gin.Context) {
	// pg_dump -d dbURL
	cmd := exec.Command("pg_dump", "--dbname="+h.dbURL, "--no-owner", "--no-acl")
	
	// Set headers for file download
	filename := fmt.Sprintf("atigacbt_backup_%s.sql", time.Now().Format("20060102_150405"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")

	// Directly pipe pg_dump output to gin response writer
	cmd.Stdout = c.Writer
	cmd.Stderr = os.Stderr // Log errors to server stderr

	if err := cmd.Run(); err != nil {
		// If we already started writing the body, this error might be too late to send via JSON
		// but typically gin won't flush until enough data is written or handler exits.
		// However, it's better to log it.
		fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
	}
}

// Restore handles database import from an uploaded SQL file
func (h *MaintenanceHandler) Restore(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file is required"}})
		return
	}

	// Save file temporarily
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("restore_%d.sql", time.Now().UnixNano()))
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to save temporary file"}})
		return
	}
	defer os.Remove(tempPath)

	// Execute restore using psql
	// We use --set ON_ERROR_STOP=on to stop on first error
	cmd := exec.Command("psql", "--dbname="+h.dbURL, "-f", tempPath, "--set", "ON_ERROR_STOP=on")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "restore_failed",
				"message": "database restore failed",
				"detail":  string(output),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true, "message": "database restored successfully"}})
}
