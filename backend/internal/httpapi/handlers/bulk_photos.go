package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/repo/masterrepo"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/storage"
)

type BulkPhotoHandler struct {
	students *masterrepo.StudentsRepo
	users    *userrepo.Repo
	store    storage.ObjectStore
}

func NewBulkPhotoHandler(students *masterrepo.StudentsRepo, users *userrepo.Repo, store storage.ObjectStore) *BulkPhotoHandler {
	return &BulkPhotoHandler{students: students, users: users, store: store}
}

type bulkPhotoResult struct {
	Filename string `json:"filename"`
	Key      string `json:"key"` // NIS / username yang dikenali
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	PhotoURL string `json:"photo_url,omitempty"`
}

// BulkUploadPhotos menerima sebuah file ZIP berisi foto-foto siswa.
// Konvensi penamaan file dalam ZIP: <NIS>.<ext> atau <username>.<ext>
// Contoh: 1234567890.jpg  atau  siswa.budi.jpg
// Handler akan mencocokkan nama file (tanpa ekstensi) ke NIS atau username siswa,
// kemudian menyimpan foto dan memperbarui photo_url di database.
func (h *BulkPhotoHandler) BulkUploadPhotos(c *gin.Context) {
	role := middleware.GetUserRole(c)
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "only admin can bulk-upload photos"}})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file ZIP diperlukan (field: file)"}})
		return
	}

	// Validasi ekstensi
	if strings.ToLower(filepath.Ext(file.Filename)) != ".zip" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "hanya file .zip yang diterima"}})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal membuka file"}})
		return
	}
	defer src.Close()

	rawBytes, err := io.ReadAll(io.LimitReader(src, 50<<20)) // max 50 MB ZIP
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal membaca file"}})
		return
	}

	zr, err := zip.NewReader(bytes.NewReader(rawBytes), int64(len(rawBytes)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "file ZIP tidak valid"}})
		return
	}

	ctx := c.Request.Context()
	allowedExts := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
	}

	results := make([]bulkPhotoResult, 0, len(zr.File))
	var matched, uploaded, skipped int

	for _, zf := range zr.File {
		// Abaikan direktori dan file tersembunyi
		base := filepath.Base(zf.Name)
		if zf.FileInfo().IsDir() || strings.HasPrefix(base, ".") || strings.HasPrefix(base, "__MACOSX") {
			continue
		}

		ext := strings.ToLower(filepath.Ext(base))
		contentType, ok := allowedExts[ext]
		if !ok {
			results = append(results, bulkPhotoResult{
				Filename: base,
				Status:   "skipped",
				Error:    "ekstensi tidak didukung (hanya jpg/jpeg/png/webp)",
			})
			skipped++
			continue
		}

		// Key = nama file tanpa ekstensi = NIS atau username
		key := strings.TrimSuffix(base, filepath.Ext(base))
		key = strings.TrimSpace(key)
		if key == "" {
			skipped++
			continue
		}

		student, found, lookupErr := h.students.FindByNISOrUsername(ctx, key)
		if lookupErr != nil {
			results = append(results, bulkPhotoResult{
				Filename: base,
				Key:      key,
				Status:   "error",
				Error:    fmt.Sprintf("error lookup: %v", lookupErr),
			})
			continue
		}
		if !found {
			results = append(results, bulkPhotoResult{
				Filename: base,
				Key:      key,
				Status:   "not_found",
				Error:    "siswa tidak ditemukan (cek NIS/username)",
			})
			skipped++
			continue
		}
		matched++

		// Buka entry ZIP dan upload
		rc, openErr := zf.Open()
		if openErr != nil {
			results = append(results, bulkPhotoResult{
				Filename: base,
				Key:      key,
				Status:   "error",
				Error:    fmt.Sprintf("gagal buka entry ZIP: %v", openErr),
			})
			continue
		}

		photoURL, uploadErr := uploadRawToStore(ctx, h.store, rc, zf.UncompressedSize64, contentType, "avatars", student.UserID)
		rc.Close()
		if uploadErr != nil {
			results = append(results, bulkPhotoResult{
				Filename: base,
				Key:      key,
				Status:   "error",
				Error:    fmt.Sprintf("gagal upload: %v", uploadErr),
			})
			continue
		}

		if dbErr := h.users.UpdatePhoto(ctx, student.UserID, photoURL); dbErr != nil {
			results = append(results, bulkPhotoResult{
				Filename: base,
				Key:      key,
				Status:   "error",
				Error:    fmt.Sprintf("gagal update DB: %v", dbErr),
			})
			continue
		}

		uploaded++
		results = append(results, bulkPhotoResult{
			Filename: base,
			Key:      key,
			Status:   "ok",
			PhotoURL: photoURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_files": len(results),
			"matched":     matched,
			"uploaded":    uploaded,
			"skipped":     skipped,
			"results":     results,
		},
	})
}

// uploadRawToStore mengunggah data dari io.ReadCloser langsung ke object store.
func uploadRawToStore(ctx context.Context, store storage.ObjectStore, r io.Reader, size uint64, contentType, dirPrefix, userID string) (string, error) {
	if store == nil {
		return "", fmt.Errorf("object store tidak dikonfigurasi")
	}
	objectKey := strings.Trim(
		filepath.ToSlash(filepath.Join(
			strings.TrimSpace(dirPrefix),
			fmt.Sprintf("%s_%d.jpg", strings.TrimSpace(userID), time.Now().UnixNano()),
		)),
		"/",
	)
	url, err := store.PutObject(ctx, objectKey, contentType, r, int64(size))
	if err != nil {
		return "", err
	}
	return url, nil
}
