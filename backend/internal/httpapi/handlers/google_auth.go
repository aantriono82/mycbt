package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/config"
	"mycbt/backend/internal/model"
	"mycbt/backend/internal/repo/masterrepo"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/authsvc"
)

type GoogleAuthHandler struct {
	cfg   config.Config
	auth  *authsvc.Service
	users *userrepo.Repo
	regs  *masterrepo.RegistrationRepo
}

func NewGoogleAuthHandler(cfg config.Config, auth *authsvc.Service, users *userrepo.Repo, regs *masterrepo.RegistrationRepo) *GoogleAuthHandler {
	return &GoogleAuthHandler{cfg: cfg, auth: auth, users: users, regs: regs}
}

func (h *GoogleAuthHandler) Redirect(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		role = "student"
	}

	if strings.TrimSpace(h.cfg.GoogleClientID) == "" || strings.TrimSpace(h.cfg.GoogleRedirectURL) == "" {
		h.redirectWithError(c, "Google OAuth belum dikonfigurasi di backend. Pastikan GOOGLE_CLIENT_ID dan GOOGLE_REDIRECT_URL sudah diset, lalu restart backend.")
		return
	}

	u := "https://accounts.google.com/o/oauth2/v2/auth"
	q := url.Values{}
	q.Set("client_id", h.cfg.GoogleClientID)
	q.Set("redirect_uri", h.cfg.GoogleRedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email")
	q.Set("state", role) // simple state to pass role
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")

	c.Redirect(http.StatusFound, u+"?"+q.Encode())
}

func (h *GoogleAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	// Exchange code for token
	tokenRes, err := h.exchangeCode(code)
	if err != nil {
		h.redirectWithError(c, "Gagal menukar kode akses: "+err.Error())
		return
	}

	// Get user info
	userInfo, err := h.getUserInfo(tokenRes.AccessToken)
	if err != nil {
		h.redirectWithError(c, "Gagal mendapatkan info user dari Google: "+err.Error())
		return
	}

	ctx := c.Request.Context()

	// 1. Try to find user by Google ID
	u, ok, err := h.users.GetByGoogleID(ctx, userInfo.ID)
	if err != nil {
		h.redirectWithError(c, "Database error: "+err.Error())
		return
	}

	// 2. If not found by Google ID, try by Email
	if !ok {
		u, ok, err = h.users.GetByEmail(ctx, userInfo.Email)
		if err != nil {
			h.redirectWithError(c, "Database error: "+err.Error())
			return
		}

		// If found by email but has no Google ID, link it now
		if ok && u.GoogleID == "" {
			if err := h.users.UpdateGoogleID(ctx, u.ID, userInfo.ID); err != nil {
				h.redirectWithError(c, "Gagal menautkan Google ID: "+err.Error())
				return
			}
			u.GoogleID = userInfo.ID
		}
	}

	// 3. If user exists, log them in
	if ok {
		token, exp, err := h.auth.IssueToken(u)
		if err != nil {
			h.redirectWithError(c, "Gagal membuat sesi login: "+err.Error())
			return
		}
		h.redirectWithToken(c, token, exp, u)
		return
	}

	// 4. If user not found, check registration state (by google_id first, then email)
	reg, regOk, err := h.regs.GetByGoogleID(ctx, userInfo.ID)
	if err != nil {
		h.redirectWithError(c, "Database error: "+err.Error())
		return
	}

	if !regOk {
		reg, regOk, err = h.regs.GetByEmail(ctx, userInfo.Email)
		if err != nil {
			h.redirectWithError(c, "Database error: "+err.Error())
			return
		}
	}

	if regOk {
		if reg.Status == "pending" {
			h.redirectWithMessage(c, "Pendaftaran Anda sedang menunggu verifikasi admin.")
			return
		}
		if reg.Status == "rejected" {
			h.redirectWithError(c, "Pendaftaran Anda ditolak. Silakan hubungi admin.")
			return
		}
		if reg.Status == "approved" {
			// Admin already approved this registration. The account should exist; if not, try to find it
			// by the registration username then link google_id.
			u2, ok2, err := h.users.GetByUsername(ctx, reg.Username)
			if err != nil {
				h.redirectWithError(c, "Database error: "+err.Error())
				return
			}
			if !ok2 {
				h.redirectWithError(c, "Pendaftaran Anda sudah disetujui, tetapi akun belum ditemukan. Silakan hubungi admin.")
				return
			}

			// Link google_id if empty; otherwise only allow if it matches.
			if strings.TrimSpace(u2.GoogleID) == "" {
				if err := h.users.UpdateGoogleID(ctx, u2.ID, userInfo.ID); err != nil {
					h.redirectWithError(c, "Gagal menautkan Google ID: "+err.Error())
					return
				}
				u2.GoogleID = userInfo.ID
			} else if u2.GoogleID != userInfo.ID {
				h.redirectWithError(c, "Akun ini sudah tertaut ke Google lain. Silakan hubungi admin.")
				return
			}

			token, exp, err := h.auth.IssueToken(u2)
			if err != nil {
				h.redirectWithError(c, "Gagal membuat sesi login: "+err.Error())
				return
			}
			h.redirectWithToken(c, token, exp, u2)
			return
		}
	}

	// 5. If no user and no registration, they must fill form first
	h.redirectWithError(c, "Email Anda belum terdaftar. Silakan lakukan pendaftaran terlebih dahulu.")
}

type googleTokenRes struct {
	AccessToken string `json:"access_token"`
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (h *GoogleAuthHandler) exchangeCode(code string) (*googleTokenRes, error) {
	v := url.Values{}
	v.Set("code", code)
	v.Set("client_id", h.cfg.GoogleClientID)
	v.Set("client_secret", h.cfg.GoogleClientSecret)
	v.Set("redirect_uri", h.cfg.GoogleRedirectURL)
	v.Set("grant_type", "authorization_code")

	resp, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google token exchange failed: %d", resp.StatusCode)
	}

	var res googleTokenRes
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (h *GoogleAuthHandler) getUserInfo(accessToken string) (*googleUserInfo, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get google user info: %d", resp.StatusCode)
	}

	var res googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (h *GoogleAuthHandler) redirectWithToken(c *gin.Context, token string, exp time.Time, u model.User) {
	frontendURL := h.cfg.CORSOrigins
	v := url.Values{}
	v.Set("token", token)
	v.Set("user_id", u.ID)
	v.Set("user_name", u.Name)
	v.Set("user_role", u.Role)

	// Choose first origin if multiple
	origin := frontendURL
	if idx := strings.Index(origin, ","); idx != -1 {
		origin = origin[:idx]
	}
	origin = strings.TrimRight(origin, "/")
	// Frontend uses Vite base `/admin-one-vue-tailwind/` and Vue Router hash mode.
	c.Redirect(http.StatusFound, origin+"/admin-one-vue-tailwind/#/auth/google/success?"+v.Encode())
}

func (h *GoogleAuthHandler) redirectWithError(c *gin.Context, msg string) {
	frontendURL := h.cfg.CORSOrigins
	origin := frontendURL
	if idx := strings.Index(origin, ","); idx != -1 {
		origin = origin[:idx]
	}
	origin = strings.TrimRight(origin, "/")
	c.Redirect(http.StatusFound, origin+"/admin-one-vue-tailwind/#/auth/google/error?message="+url.QueryEscape(msg))
}

func (h *GoogleAuthHandler) redirectWithMessage(c *gin.Context, msg string) {
	frontendURL := h.cfg.CORSOrigins
	origin := frontendURL
	if idx := strings.Index(origin, ","); idx != -1 {
		origin = origin[:idx]
	}
	origin = strings.TrimRight(origin, "/")
	c.Redirect(http.StatusFound, origin+"/admin-one-vue-tailwind/#/auth/google/info?message="+url.QueryEscape(msg))
}

func (h *GoogleAuthHandler) SubmitRegistration(c *gin.Context) {
	var req struct {
		Role          string  `json:"role" binding:"required"`
		GoogleID      *string `json:"google_id"`
		GooglePicture *string `json:"google_picture"`
		Name          string  `json:"name" binding:"required"`
		Email         string  `json:"email" binding:"required"`
		Phone         string  `json:"phone" binding:"required"`
		NIS           *string `json:"nis"`
		NIP           *string `json:"nip"`
		Note          *string `json:"note"`
		NISN          *string `json:"nisn"`
		Jenjang       *string `json:"jenjang"`
		Gender        *string `json:"gender"`
		BirthDate     *string `json:"birth_date"`
		SchoolName    *string `json:"school_name"`
		AcademicYear  *string `json:"academic_year"`
		NISSekolah    *string `json:"nis_sekolah"`
		ProgramCode   *string `json:"program_code"`
		LevelName     *string `json:"level_name"`
		GroupName     *string `json:"group_name"`
		MapelCodes    *string `json:"mapel_codes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data: " + err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 1. Check if user already exists with this email
	_, ok, err := h.users.GetByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Email tersebut sudah terdaftar."})
		return
	}

	// 2. Check if registration already exists with this email
	reg, ok, err := h.regs.GetByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if ok {
		switch reg.Status {
		case "pending":
			c.JSON(http.StatusConflict, gin.H{"error": "Pendaftaran dengan email ini sedang diproses (menunggu verifikasi admin)."})
			return
		case "approved":
			c.JSON(http.StatusConflict, gin.H{"error": "Email ini sudah disetujui. Silakan login menggunakan tombol Google."})
			return
		case "rejected":
			// Allow re-register after rejection (admin might want updated data).
		default:
			c.JSON(http.StatusConflict, gin.H{"error": "Pendaftaran dengan email ini tidak dapat diproses. Silakan hubungi admin."})
			return
		}
	}

	var bd *time.Time
	if req.BirthDate != nil && *req.BirthDate != "" {
		if t, err := time.Parse(time.RFC3339, *req.BirthDate); err == nil {
			bd = &t
		}
	}

	str := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	// Generate a temporary username if google_id is missing
	username := "user_" + randHexString(6)
	regNISN := str(req.NISN)
	regNIP := str(req.NIP)
	regNIS := str(req.NIS)
	if req.Role == "student" && regNISN != "" {
		username = regNISN
	} else if req.Role == "teacher" && (regNIP != "" || regNIS != "") {
		if regNIP != "" {
			username = regNIP
		} else {
			username = regNIS
		}
	}

	// 3. Create registration request
	newReg := masterrepo.RegistrationRequest{
		Role:          req.Role,
		Username:      username,
		Name:          req.Name,
		Email:         req.Email,
		GoogleID:      str(req.GoogleID),
		GooglePicture: str(req.GooglePicture),
		Status:        "pending",
		NIS:           regNIS,
		NIP:           regNIP,
		Note:          str(req.Note),
		NISN:          regNISN,
		Jenjang:       str(req.Jenjang),
		Gender:        str(req.Gender),
		BirthDate:     bd,
		SchoolName:    str(req.SchoolName),
		AcademicYear:  str(req.AcademicYear),
		NISSekolah:    str(req.NISSekolah),
		ProgramCode:   str(req.ProgramCode),
		LevelName:     str(req.LevelName),
		GroupName:     str(req.GroupName),
		MapelCodes:    str(req.MapelCodes),
		Phone:         req.Phone,
	}

	if _, err := h.regs.Create(ctx, newReg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pendaftaran: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pendaftaran berhasil dikirim. Silakan tunggu verifikasi admin."})
}

func randHexString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
