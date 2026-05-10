package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/repo/masterrepo"
	"atigacbt/backend/internal/repo/userrepo"
	"atigacbt/backend/internal/storage"
)

type TeacherSchoolHandler struct {
	schools *masterrepo.SchoolRepo
	users   *userrepo.Repo
	store   storage.ObjectStore
}

func NewTeacherSchoolHandler(schools *masterrepo.SchoolRepo, users *userrepo.Repo, store storage.ObjectStore) *TeacherSchoolHandler {
	if store == nil {
		store = storage.NewLocalObjectStore("uploads")
	}
	return &TeacherSchoolHandler{schools: schools, users: users, store: store}
}

func (h *TeacherSchoolHandler) getSchoolID(c *gin.Context) (string, error) {
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)
	if userIDStr == "" {
		return "", fmt.Errorf("unauthorized")
	}

	user, ok, err := h.users.GetByID(c.Request.Context(), userIDStr)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("user not found")
	}
	if user.SchoolID == nil || *user.SchoolID == "" {
		return "", fmt.Errorf("no school associated with this account")
	}
	return *user.SchoolID, nil
}

func (h *TeacherSchoolHandler) Get(c *gin.Context) {
	schoolID, err := h.getSchoolID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": err.Error()}})
		return
	}

	school, ok, err := h.schools.GetByID(c.Request.Context(), schoolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "not_found", "message": "school not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": school})
}

func (h *TeacherSchoolHandler) Update(c *gin.Context) {
	schoolID, err := h.getSchoolID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": err.Error()}})
		return
	}

	var req schoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "validation", "message": "school name is required"}})
		return
	}

	err = h.schools.Update(c.Request.Context(), masterrepo.School{
		ID:            schoolID,
		Name:          strings.TrimSpace(req.Name),
		LogoURL:       req.LogoURL,
		Address:       req.Address,
		Phone:         req.Phone,
		Email:         req.Email,
		Website:       req.Website,
		PrincipalName: req.PrincipalName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": schoolID}})
}

func (h *TeacherSchoolHandler) UploadLogo(c *gin.Context) {
	schoolID, err := h.getSchoolID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": err.Error()}})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file is required"}})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("school_%s_%d", schoolID, time.Now().UnixNano())
	logoURL, err := uploadImageToStore(c.Request.Context(), h.store, file, "logos", filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to save uploaded file"}})
		return
	}

	// Update school record with new logo
	school, ok, err := h.schools.GetByID(c.Request.Context(), schoolID)
	if err == nil && ok {
		school.LogoURL = logoURL
		_ = h.schools.Update(c.Request.Context(), school)
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"logo_url": logoURL, "ext": ext}})
}
