package middleware_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gin-contrib/cors"

	"atigacbt/backend/internal/config"
	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/model"
	"atigacbt/backend/internal/service/authsvc"
)

func makeAuthService(t *testing.T) *authsvc.Service {
	t.Helper()
	svc, err := authsvc.New(config.Config{
		JWTSecret:     "test-secret-key-123",
		JWTIssuer:     "atigacbt-test",
		JWTTTLMinutes: "120",
	}, nil)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return svc
}

func issueExpiredToken(t *testing.T, secret, issuer string) string {
	t.Helper()
	claims := authsvc.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   "u-expired",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-20 * time.Minute)),
		},
		Role:     "admin",
		Username: "expired-user",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	return s
}

func prettyBody(b string) string {
	b = strings.TrimSpace(b)
	if b == "" {
		return "<empty>"
	}
	var js any
	if err := json.Unmarshal([]byte(b), &js); err == nil {
		out, _ := json.Marshal(js)
		return string(out)
	}
	return b
}

func TestScenarioAuthMiddlewareJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := makeAuthService(t)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/protected", middleware.RequireAuth(auth), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "user_id": middleware.GetUserID(c)})
	})

	t.Run("without authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		body := rec.Body.String()
		t.Logf("auth/no-header -> status=%d headers=%v body=%s", rec.Code, rec.Header(), prettyBody(body))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		if !strings.Contains(body, "missing Authorization header") {
			t.Fatalf("expected missing header message, got %s", body)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expired := issueExpiredToken(t, "test-secret-key-123", "atigacbt-test")
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expired)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		body := rec.Body.String()
		t.Logf("auth/expired-token -> status=%d headers=%v body=%s", rec.Code, rec.Header(), prettyBody(body))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		// Current middleware returns generic "invalid token" for expired token.
		if !strings.Contains(body, "invalid token") {
			t.Fatalf("expected invalid token message, got %s", body)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		token, _, err := auth.IssueToken(model.User{
			ID:       "u-valid",
			Role:     "admin",
			Username: "admin",
		})
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		body := rec.Body.String()
		t.Logf("auth/valid-token -> status=%d headers=%v body=%s", rec.Code, rec.Header(), prettyBody(body))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(body, "\"ok\":true") {
			t.Fatalf("expected handler execution, got body %s", body)
		}
	})
}

func TestScenarioRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(gin.Recovery())
	scope := "test-rate-limit-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	r.GET("/limited", middleware.RateLimit(scope, 2, time.Minute), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.RemoteAddr = "203.0.113.10:12345"
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		body := rec.Body.String()
		t.Logf("rate-limit/request-%d -> status=%d retry-after=%q body=%s", i, rec.Code, rec.Header().Get("Retry-After"), prettyBody(body))

		if i <= 2 && rec.Code != http.StatusOK {
			t.Fatalf("request %d should pass with 200, got %d", i, rec.Code)
		}
		if i == 3 {
			if rec.Code != http.StatusTooManyRequests {
				t.Fatalf("request 3 should be 429, got %d", rec.Code)
			}
			if strings.TrimSpace(rec.Header().Get("Retry-After")) == "" {
				t.Fatalf("expected Retry-After header on 429")
			}
			if !strings.Contains(body, "too many requests") {
				t.Fatalf("expected too many requests message, got %s", body)
			}
		}
	}
}

func TestScenarioCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		Env:         "release",
		CORSOrigins: "http://allowed.local",
	}
	r := gin.New()
	r.Use(gin.Recovery())
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	if cfg.Env == "debug" || cfg.Env == "" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = []string{"http://allowed.local"}
	}
	r.Use(cors.New(corsConfig))
	r.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	t.Run("origin not registered", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
		req.Header.Set("Origin", "http://blocked.local")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		body := rec.Body.String()

		t.Logf("cors/disallowed-origin -> status=%d allow-origin=%q body=%s", rec.Code, rec.Header().Get("Access-Control-Allow-Origin"), prettyBody(body))
	})

	t.Run("preflight options", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/ping", nil)
		req.Header.Set("Origin", "http://allowed.local")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Authorization,Content-Type")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		body := rec.Body.String()

		t.Logf(
			"cors/preflight -> status=%d allow-origin=%q allow-methods=%q allow-headers=%q body=%s",
			rec.Code,
			rec.Header().Get("Access-Control-Allow-Origin"),
			rec.Header().Get("Access-Control-Allow-Methods"),
			rec.Header().Get("Access-Control-Allow-Headers"),
			prettyBody(body),
		)
	})
}

func TestScenarioPanicRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/panic-test", func(c *gin.Context) {
		panic("forced panic for testing")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic-test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	body := rec.Body.String()

	encodedBody := base64.StdEncoding.EncodeToString([]byte(body))
	t.Logf("panic/recovery -> status=%d headers=%v body(base64)=%s", rec.Code, rec.Header(), encodedBody)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on panic recovery, got %d", rec.Code)
	}
}
