package authsvc

import (
	"crypto/sha256"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"mycbt/backend/internal/config"
	"mycbt/backend/internal/model"
	"mycbt/backend/internal/repo/userrepo"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user inactive")
	ErrUnauthorized       = errors.New("unauthorized")
)

type Service struct {
	users *userrepo.Repo

	secret []byte
	issuer string
	ttl    time.Duration

	blocklist TokenBlocklist
}

type Claims struct {
	jwt.RegisteredClaims
	Role     string `json:"role"`
	Username string `json:"username"`
}

type TokenBlocklist interface {
	Revoke(ctx context.Context, tokenHash string, ttl time.Duration) error
	IsRevoked(ctx context.Context, tokenHash string) (bool, error)
}

func New(cfg config.Config, users *userrepo.Repo) (*Service, error) {
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	min, err := strconv.Atoi(cfg.JWTTTLMinutes)
	if err != nil || min <= 0 {
		return nil, fmt.Errorf("JWT_TTL_MINUTES must be a positive integer")
	}

	return &Service{
		users:  users,
		secret: []byte(cfg.JWTSecret),
		issuer: cfg.JWTIssuer,
		ttl:    time.Duration(min) * time.Minute,
	}, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (token string, expiresAt time.Time, user model.User, err error) {
	u, ok, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", time.Time{}, model.User{}, err
	}
	if !ok {
		return "", time.Time{}, model.User{}, ErrInvalidCredentials
	}
	if !u.IsActive {
		return "", time.Time{}, model.User{}, ErrUserInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", time.Time{}, model.User{}, ErrInvalidCredentials
	}

	expiresAt = time.Now().UTC().Add(s.ttl)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   u.ID,
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Role:     u.Role,
		Username: u.Username,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, model.User{}, fmt.Errorf("sign jwt: %w", err)
	}

	// Don't return password hash.
	u.PasswordHash = ""
	return token, expiresAt, u, nil
}

func (s *Service) IssueToken(user model.User) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().UTC().Add(s.ttl)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.ID,
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Role:     user.Role,
		Username: user.Username,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString(s.secret)
	return token, expiresAt, err
}

func (s *Service) ParseToken(tokenString string) (Claims, error) {
	return s.parseToken(tokenString, true)
}

func (s *Service) parseToken(tokenString string, checkRevoked bool) (Claims, error) {
	var c Claims
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return Claims{}, ErrUnauthorized
	}
	if checkRevoked && s.blocklist != nil {
		revoked, err := s.blocklist.IsRevoked(context.Background(), tokenHash(tokenString))
		if err == nil && revoked {
			return Claims{}, ErrUnauthorized
		}
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer))
	if err != nil {
		return Claims{}, ErrUnauthorized
	}
	if !parsed.Valid {
		return Claims{}, ErrUnauthorized
	}

	return c, nil
}

func (s *Service) SetBlocklist(blocklist TokenBlocklist) {
	s.blocklist = blocklist
}

func (s *Service) RevokeToken(ctx context.Context, tokenString string) error {
	if s.blocklist == nil {
		return nil
	}

	claims, err := s.parseToken(tokenString, false)
	if err != nil {
		return err
	}
	ttl := time.Minute
	if claims.ExpiresAt != nil {
		if d := time.Until(claims.ExpiresAt.Time); d > 0 {
			ttl = d
		}
	}
	return s.blocklist.Revoke(ctx, tokenHash(tokenString), ttl)
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
