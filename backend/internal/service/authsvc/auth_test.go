package authsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"atigacbt/backend/internal/config"
	"atigacbt/backend/internal/model"
)

type stubUserReader struct {
	user model.User
	ok   bool
	err  error
}

func (s stubUserReader) GetByUsername(_ context.Context, username string) (model.User, bool, error) {
	if s.err != nil {
		return model.User{}, false, s.err
	}
	if s.ok && s.user.Username == username {
		return s.user, true, nil
	}
	return model.User{}, false, nil
}

type stubBlocklist struct {
	revoked map[string]bool
	lastKey string
	lastTTL time.Duration
}

func (s *stubBlocklist) Revoke(_ context.Context, tokenHash string, ttl time.Duration) error {
	if s.revoked == nil {
		s.revoked = map[string]bool{}
	}
	s.revoked[tokenHash] = true
	s.lastKey = tokenHash
	s.lastTTL = ttl
	return nil
}

func (s *stubBlocklist) IsRevoked(_ context.Context, tokenHash string) (bool, error) {
	return s.revoked[tokenHash], nil
}

func TestNew_ValidatesConfig(t *testing.T) {
	t.Parallel()

	_, err := New(config.Config{JWTTTLMinutes: "15"}, nil)
	if err == nil || err.Error() != "JWT_SECRET is required" {
		t.Fatalf("expected missing secret error, got %v", err)
	}

	_, err = New(config.Config{JWTSecret: "secret", JWTTTLMinutes: "0"}, nil)
	if err == nil || err.Error() != "JWT_TTL_MINUTES must be a positive integer" {
		t.Fatalf("expected ttl validation error, got %v", err)
	}
}

func TestLogin_SuccessClearsPasswordHash(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}

	svc := &Service{
		users: stubUserReader{
			user: model.User{
				ID:           "user-1",
				Username:     "alice",
				PasswordHash: hash,
				Role:         "admin",
				IsActive:     true,
			},
			ok: true,
		},
		secret: []byte("jwt-secret"),
		issuer: "atigacbt",
		ttl:    15 * time.Minute,
	}

	token, exp, user, err := svc.Login(context.Background(), "alice", "secret123")
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be issued")
	}
	if time.Until(exp) <= 0 {
		t.Fatalf("expected future expiry, got %v", exp)
	}
	if user.PasswordHash != "" {
		t.Fatalf("expected password hash to be cleared, got %q", user.PasswordHash)
	}

	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.Subject != "user-1" || claims.Role != "admin" || claims.Username != "alice" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestLogin_RejectsInvalidCredentialsAndInactiveUser(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}

	tests := []struct {
		name string
		svc  *Service
		err  error
	}{
		{
			name: "unknown user",
			svc: &Service{
				users:  stubUserReader{},
				secret: []byte("jwt-secret"),
				issuer: "atigacbt",
				ttl:    15 * time.Minute,
			},
			err: ErrInvalidCredentials,
		},
		{
			name: "inactive user",
			svc: &Service{
				users: stubUserReader{
					user: model.User{
						Username:     "alice",
						PasswordHash: string(hash),
						IsActive:     false,
					},
					ok: true,
				},
				secret: []byte("jwt-secret"),
				issuer: "atigacbt",
				ttl:    15 * time.Minute,
			},
			err: ErrUserInactive,
		},
		{
			name: "wrong password",
			svc: &Service{
				users: stubUserReader{
					user: model.User{
						Username:     "alice",
						PasswordHash: string(hash),
						IsActive:     true,
					},
					ok: true,
				},
				secret: []byte("jwt-secret"),
				issuer: "atigacbt",
				ttl:    15 * time.Minute,
			},
			err: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, _, err := tt.svc.Login(context.Background(), "alice", "wrong")
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, got %v", tt.err, err)
			}
		})
	}
}

func TestParseToken_RejectsEmptyAndRevokedToken(t *testing.T) {
	t.Parallel()

	svc := &Service{
		secret: []byte("jwt-secret"),
		issuer: "atigacbt",
		ttl:    15 * time.Minute,
	}

	if _, err := svc.ParseToken(" "); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized for empty token, got %v", err)
	}

	token, _, err := svc.IssueToken(model.User{ID: "user-1", Username: "alice", Role: "teacher"})
	if err != nil {
		t.Fatalf("IssueToken error: %v", err)
	}

	blocklist := &stubBlocklist{revoked: map[string]bool{tokenHash(token): true}}
	svc.SetBlocklist(blocklist)

	if _, err := svc.ParseToken(token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized for revoked token, got %v", err)
	}
}

func TestRevokeToken_StoresHashAndTTL(t *testing.T) {
	t.Parallel()

	svc := &Service{
		secret: []byte("jwt-secret"),
		issuer: "atigacbt",
		ttl:    5 * time.Minute,
	}

	token, _, err := svc.IssueToken(model.User{ID: "user-1", Username: "alice", Role: "teacher"})
	if err != nil {
		t.Fatalf("IssueToken error: %v", err)
	}

	blocklist := &stubBlocklist{}
	svc.SetBlocklist(blocklist)

	if err := svc.RevokeToken(context.Background(), token); err != nil {
		t.Fatalf("RevokeToken error: %v", err)
	}
	if blocklist.lastKey != tokenHash(token) {
		t.Fatalf("expected revoked key %q, got %q", tokenHash(token), blocklist.lastKey)
	}
	if blocklist.lastTTL <= 0 {
		t.Fatalf("expected positive ttl, got %v", blocklist.lastTTL)
	}
}
