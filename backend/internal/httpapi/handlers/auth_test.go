package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"atigacbt/backend/internal/model"
	"atigacbt/backend/internal/repo/loginlogrepo"
	"atigacbt/backend/internal/service/authsvc"
)

type mockAuthService struct {
	token      string
	expiresAt  time.Time
	user       model.User
	loginErr   error
	lastUser   string
	lastPass   string
	revokedRaw string
	revokeErr  error
}

func (m *mockAuthService) Login(_ context.Context, username, password string) (string, time.Time, model.User, error) {
	m.lastUser = username
	m.lastPass = password
	return m.token, m.expiresAt, m.user, m.loginErr
}

func (m *mockAuthService) RevokeToken(_ context.Context, tokenString string) error {
	m.revokedRaw = tokenString
	return m.revokeErr
}

type mockUserStore struct {
	getByIDUser        model.User
	getByIDOK          bool
	getByIDErr         error
	updatedProfileID   string
	updatedProfileName string
	updatedProfileMail string
	updateProfileErr   error
	updatedPasswordID  string
	updatedPassword    string
	updatedPasswordRaw string
	updatePasswordErr  error
	updatePhotoID      string
	updatePhotoURL     string
	updatePhotoErr     error
}

func (m *mockUserStore) GetByID(_ context.Context, id string) (model.User, bool, error) {
	return m.getByIDUser, m.getByIDOK, m.getByIDErr
}

func (m *mockUserStore) UpdateProfile(_ context.Context, id, name, email string) error {
	m.updatedProfileID = id
	m.updatedProfileName = name
	m.updatedProfileMail = email
	return m.updateProfileErr
}

func (m *mockUserStore) UpdatePassword(_ context.Context, id string, hash string, plain string) error {
	m.updatedPasswordID = id
	m.updatedPassword = hash
	m.updatedPasswordRaw = plain
	return m.updatePasswordErr
}

func (m *mockUserStore) UpdatePhoto(_ context.Context, id string, photoURL string) error {
	m.updatePhotoID = id
	m.updatePhotoURL = photoURL
	return m.updatePhotoErr
}

type mockLoginLogStore struct {
	inserted   []loginlogrepo.LoginLog
	pruneDays  []int
	insertErr  error
	pruneError error
}

func (m *mockLoginLogStore) Insert(_ context.Context, in loginlogrepo.LoginLog) error {
	m.inserted = append(m.inserted, in)
	return m.insertErr
}

func (m *mockLoginLogStore) PruneOlderThan(_ context.Context, days int) (int64, error) {
	m.pruneDays = append(m.pruneDays, days)
	return 0, m.pruneError
}

func performJSONRequest(t *testing.T, handler gin.HandlerFunc, method, path, body string, setup func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	if setup != nil {
		setup(c)
	}
	handler(c)
	return w
}

func TestAuthHandlerLogin_InvalidJSON(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	h := &AuthHandler{auth: &mockAuthService{}}

	w := performJSONRequest(t, h.Login, http.MethodPost, "/login", "{", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandlerLogin_SuccessAndLogsActivity(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	logs := &mockLoginLogStore{}
	auth := &mockAuthService{
		token:     "jwt-token",
		expiresAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		user: model.User{
			ID:       "user-1",
			Username: "alice",
			Role:     "admin",
			Name:     "Alice",
			Email:    "alice@example.com",
		},
	}
	h := &AuthHandler{auth: auth, loginLogs: logs}

	w := performJSONRequest(t, h.Login, http.MethodPost, "/login", `{"username":"alice","password":"secret123"}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if auth.lastUser != "alice" || auth.lastPass != "secret123" {
		t.Fatalf("unexpected login args: %q %q", auth.lastUser, auth.lastPass)
	}
	if len(logs.inserted) != 1 || len(logs.pruneDays) != 1 || logs.pruneDays[0] != 30 {
		t.Fatalf("expected login activity and prune, got inserted=%d prune=%v", len(logs.inserted), logs.pruneDays)
	}

	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			User        struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Data.AccessToken != "jwt-token" || resp.Data.TokenType != "Bearer" || resp.Data.User.ID != "user-1" {
		t.Fatalf("unexpected response payload: %+v", resp)
	}
}

func TestAuthHandlerLogin_MapsAuthErrors(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		err  error
		code int
	}{
		{name: "invalid credentials", err: authsvc.ErrInvalidCredentials, code: http.StatusUnauthorized},
		{name: "inactive", err: authsvc.ErrUserInactive, code: http.StatusForbidden},
		{name: "internal", err: errors.New("db down"), code: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &AuthHandler{auth: &mockAuthService{loginErr: tt.err}}
			w := performJSONRequest(t, h.Login, http.MethodPost, "/login", `{"username":"alice","password":"secret123"}`, nil)
			if w.Code != tt.code {
				t.Fatalf("expected %d, got %d", tt.code, w.Code)
			}
		})
	}
}

func TestAuthHandlerLogout_RevokesBearerTokenWhenPresent(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	auth := &mockAuthService{}
	h := &AuthHandler{auth: auth}

	w := performJSONRequest(t, h.Logout, http.MethodPost, "/logout", `{}`, func(c *gin.Context) {
		c.Request.Header.Set("Authorization", "Bearer jwt-token")
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if auth.revokedRaw != "jwt-token" {
		t.Fatalf("expected token to be revoked, got %q", auth.revokedRaw)
	}
}

func TestAuthHandlerMe_RejectsMissingContextUser(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	h := &AuthHandler{users: &mockUserStore{}}

	w := performJSONRequest(t, h.Me, http.MethodGet, "/me", `{}`, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthHandlerUpdateMe_ValidatesNameAndPersistsProfile(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	users := &mockUserStore{
		getByIDUser: model.User{ID: "user-1", Username: "alice", Name: "Alice Updated", Email: "alice@example.com", Role: "teacher", IsActive: true},
		getByIDOK:   true,
	}
	h := &AuthHandler{users: users}

	invalid := performJSONRequest(t, h.UpdateMe, http.MethodPut, "/me", `{"name":"  ","email":"a@example.com"}`, func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", invalid.Code)
	}

	w := performJSONRequest(t, h.UpdateMe, http.MethodPut, "/me", `{"name":" Alice Updated ","email":" alice@example.com "}`, func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if users.updatedProfileID != "user-1" || users.updatedProfileName != "Alice Updated" || users.updatedProfileMail != "alice@example.com" {
		t.Fatalf("unexpected profile update args: id=%q name=%q email=%q", users.updatedProfileID, users.updatedProfileName, users.updatedProfileMail)
	}
}

func TestAuthHandlerChangePassword_ValidatesAndUpdatesHash(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	currentHash, err := authsvc.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	users := &mockUserStore{
		getByIDUser: model.User{ID: "user-1", PasswordHash: currentHash},
		getByIDOK:   true,
	}
	h := &AuthHandler{users: users}

	mismatch := performJSONRequest(t, h.ChangePassword, http.MethodPost, "/me/password", `{"current_password":"secret123","new_password":"newpass123","password_confirmation":"different"}`, func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	if mismatch.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", mismatch.Code)
	}

	w := performJSONRequest(t, h.ChangePassword, http.MethodPost, "/me/password", `{"current_password":"secret123","new_password":"newpass123","password_confirmation":"newpass123"}`, func(c *gin.Context) {
		c.Set("user_id", "user-1")
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if users.updatedPasswordID != "user-1" {
		t.Fatalf("expected password update for user-1, got %q", users.updatedPasswordID)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(users.updatedPassword), []byte("newpass123")); err != nil {
		t.Fatalf("expected stored hash to match new password: %v", err)
	}
}

func TestAuthHandlerUploadPhoto_RejectsNonAdminTargetOverride(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	h := &AuthHandler{users: &mockUserStore{}}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("dummy", "x"); err != nil {
		t.Fatalf("WriteField error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer close error: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/me/photo?target_user_id=user-2", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Set("user_id", "user-1")
	c.Set("user_role", "student")
	h.UploadPhoto(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
