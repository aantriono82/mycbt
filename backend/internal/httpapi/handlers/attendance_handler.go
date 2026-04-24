package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/httpapi/pgerr"
	"mycbt/backend/internal/repo/masterrepo"
	"mycbt/backend/internal/repo/studentexamrepo"
)

type AttendanceHandler struct {
	sessions *masterrepo.AttendanceSessionsRepo
	student  *studentexamrepo.Repo
	settings *masterrepo.SettingsRepo
}

func NewAttendanceHandler(sessions *masterrepo.AttendanceSessionsRepo, student *studentexamrepo.Repo, settings *masterrepo.SettingsRepo) *AttendanceHandler {
	return &AttendanceHandler{
		sessions: sessions,
		student:  student,
		settings: settings,
	}
}

type createAttendanceSessionReq struct {
	Lat          *float64 `json:"lat"`
	Lon          *float64 `json:"lon"`
	RadiusMeters int      `json:"radius_meters"`
	DurationMin  int      `json:"duration_minutes"`
}

func (h *AttendanceHandler) CreateSession(c *gin.Context) {
	var req createAttendanceSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	examID := c.Param("id")
	if examID == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "exam_id required"}})
		return
	}

	if req.RadiusMeters <= 0 {
		req.RadiusMeters = 50 // Default 50m
	}
	if req.DurationMin <= 0 {
		req.DurationMin = 15 // Default 15 min
	}

	// Generate a unique token
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	expiresAt := time.Now().Add(time.Duration(req.DurationMin) * time.Minute)

	// Cleanup expired
	_, _ = h.sessions.DeleteExpired(c.Request.Context())

	session, err := h.sessions.Create(c.Request.Context(), examID, token, req.Lat, req.Lon, req.RadiusMeters, expiresAt)
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "exam not found"}})
			return
		}
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to create attendance session"}})
		return
	}

	c.JSON(201, gin.H{"data": session})
}

func (h *AttendanceHandler) ListActiveSessions(c *gin.Context) {
	examID := c.Param("id")
	items, err := h.sessions.ListActiveByExam(c.Request.Context(), examID)
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to list active sessions"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type scanAttendanceReq struct {
	Token    string   `json:"token"`
	Lat      *float64 `json:"lat"`
	Lon      *float64 `json:"lon"`
	Accuracy *float64 `json:"accuracy"`
}

func (h *AttendanceHandler) ScanAttendance(c *gin.Context) {
	userID := middleware.GetUserID(c)
	st, ok, err := h.student.StudentByUserID(c.Request.Context(), userID)
	if err != nil || !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}

	var req scanAttendanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "token required"}})
		return
	}

	// 1. Verify token
	sess, ok, err := h.sessions.GetByToken(c.Request.Context(), token)
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to verify token"}})
		return
	}
	if !ok {
		c.JSON(400, gin.H{"error": gin.H{"code": "invalid_token", "message": "invalid or expired token"}})
		return
	}
	if sess.ExpiresAt.Before(time.Now()) {
		c.JSON(400, gin.H{"error": gin.H{"code": "expired_token", "message": "token has expired"}})
		return
	}

	// 2. Verify Geofence if required
	if sess.Lat != nil && sess.Lon != nil {
		if req.Lat == nil || req.Lon == nil {
			c.JSON(400, gin.H{"error": gin.H{"code": "location_required", "message": "location is required for this session"}})
			return
		}

		dist := haversine(*sess.Lat, *sess.Lon, *req.Lat, *req.Lon)
		if dist > float64(sess.RadiusMeters) {
			c.JSON(403, gin.H{
				"error": gin.H{
					"code":    "outside_geofence",
					"message": "you are outside the allowed area",
				},
				"meta": gin.H{
					"distance_meters": int(dist),
					"allowed_radius":  sess.RadiusMeters,
				},
			})
			return
		}
	}

	// 3. Upsert attendance
	clientIP := net.ParseIP(c.ClientIP())
	item, err := h.student.UpsertAttendance(c.Request.Context(), sess.ExamID, st.StudentID, "QR Scan", clientIP, time.Now().UTC(), studentexamrepo.AttendanceOption{
		Lat:       req.Lat,
		Lon:       req.Lon,
		Accuracy:  req.Accuracy,
		IsQR:      true,
		SessionID: sess.ID,
	})
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "failed to record attendance"}})
		return
	}

	c.JSON(200, gin.H{"data": item})
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	dphi := (lat2 - lat1) * math.Pi / 180
	dlambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dphi/2)*math.Sin(dphi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(dlambda/2)*math.Sin(dlambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
