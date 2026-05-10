package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/repo/masterrepo"
	"atigacbt/backend/internal/storage"
)

type SchoolHandler struct {
	repo  *masterrepo.SchoolRepo
	store storage.ObjectStore
}

func NewSchoolHandler(repo *masterrepo.SchoolRepo, store storage.ObjectStore) *SchoolHandler {
	if store == nil {
		store = storage.NewLocalObjectStore("uploads")
	}
	return &SchoolHandler{repo: repo, store: store}
}

func (h *SchoolHandler) List(c *gin.Context) {
	data, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *SchoolHandler) Get(c *gin.Context) {
	id := c.Param("id")
	data, ok, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "school not found"}})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

type schoolReq struct {
	Name          string `json:"name"`
	LogoURL       string `json:"logo_url"`
	Address       string `json:"address"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Website       string `json:"website"`
	PrincipalName string `json:"principal_name"`
}

func (h *SchoolHandler) Create(c *gin.Context) {
	var req schoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "validation", "message": "name is required"}})
		return
	}

	id, err := h.repo.Create(c.Request.Context(), masterrepo.School{
		Name:          strings.TrimSpace(req.Name),
		LogoURL:       req.LogoURL,
		Address:       req.Address,
		Phone:         req.Phone,
		Email:         req.Email,
		Website:       req.Website,
		PrincipalName: req.PrincipalName,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": gin.H{"id": id}})
}

func (h *SchoolHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req schoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	err := h.repo.Update(c.Request.Context(), masterrepo.School{
		ID:            id,
		Name:          strings.TrimSpace(req.Name),
		LogoURL:       req.LogoURL,
		Address:       req.Address,
		Phone:         req.Phone,
		Email:         req.Email,
		Website:       req.Website,
		PrincipalName: req.PrincipalName,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"id": id}})
}

func (h *SchoolHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(204, nil)
}

func (h *SchoolHandler) UploadLogo(c *gin.Context) {
	id := c.Param("id")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file is required"}})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("school_%s_%d", id, time.Now().UnixNano())
	logoURL, err := uploadImageToStore(c.Request.Context(), h.store, file, "logos", filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "failed to save uploaded file"}})
		return
	}

	// Update school record with new logo
	school, ok, err := h.repo.GetByID(c.Request.Context(), id)
	if err == nil && ok {
		school.LogoURL = logoURL
		_ = h.repo.Update(c.Request.Context(), school)
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"logo_url": logoURL, "ext": ext}})
}
