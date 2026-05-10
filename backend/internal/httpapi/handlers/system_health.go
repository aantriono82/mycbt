package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/httpapi/middleware"
)

type SystemHealthHandler struct{}

func NewSystemHealthHandler() *SystemHealthHandler {
	return &SystemHealthHandler{}
}

func (h *SystemHealthHandler) Get(c *gin.Context) {
	now := time.Now()
	snapshot := middleware.ReadRuntimeSnapshot(now)
	queue := middleware.ReadQueueSnapshot()
	historyLimit := 20
	if raw := c.Query("history_limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			historyLimit = n
		}
	}
	queueHistory := middleware.ReadQueueHistorySnapshot(historyLimit)

	c.JSON(200, gin.H{
		"data": gin.H{
			"health":        "ok",
			"metrics":       snapshot,
			"queue":         queue,
			"queue_history": queueHistory,
		},
	})
}

func (h *SystemHealthHandler) ClearQueueHistory(c *gin.Context) {
	middleware.ClearQueueHistory()
	c.JSON(200, gin.H{
		"data": gin.H{
			"ok": true,
		},
	})
}
