package ltisvc

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"mycbt/backend/internal/repo/ltirepo"
)

const AGSScoreScope = "https://purl.imsglobal.org/spec/lti-ags/scope/score"

type AGSService struct {
	client *http.Client
}

func NewAGSService(client *http.Client) *AGSService {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &AGSService{client: client}
}

type PublishScoreInput struct {
	Platform    ltirepo.Platform
	LineItemURL string
	LTISub      string
	Score       float64
	Timestamp   time.Time
}

func (s *AGSService) PublishScore(ctx context.Context, in PublishScoreInput, scopeText string) error {
	if strings.TrimSpace(in.LineItemURL) == "" {
		return fmt.Errorf("missing line item url")
	}
	if !hasScope(scopeText, AGSScoreScope) {
		return fmt.Errorf("platform launch does not grant AGS score scope")
	}

	token, err := s.fetchAccessToken(ctx, in.Platform, AGSScoreScope)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"userId":           in.LTISub,
		"scoreGiven":       in.Score,
		"scoreMaximum":     100.0,
		"activityProgress": "Completed",
		"gradingProgress":  "FullyGraded",
		"timestamp":        in.Timestamp.UTC().Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal ags score payload: %w", err)
	}

	endpoint := strings.TrimRight(in.LineItemURL, "/") + "/scores"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build ags score request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.ims.lis.v1.score+json")
	req.Header.Set("Accept", "application/json, application/vnd.ims.lis.v1.score+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send ags score request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ags score sync rejected with status %d", resp.StatusCode)
	}
	return nil
}

func (s *AGSService) fetchAccessToken(ctx context.Context, platform ltirepo.Platform, scope string) (string, error) {
	assertion, err := buildClientAssertion(platform)
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", scope)
	form.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	form.Set("client_assertion", assertion)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, platform.OIDCTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build ags token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request ags access token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode ags token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ags token request rejected with status %d", resp.StatusCode)
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return "", fmt.Errorf("ags token response missing access_token")
	}
	return tokenResp.AccessToken, nil
}

func buildClientAssertion(platform ltirepo.Platform) (string, error) {
	key, err := parsePrivateKey(platform.ToolPrivateKey)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": platform.ClientID,
		"sub": platform.ClientID,
		"aud": platform.OIDCTokenURL,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"jti": randomString(24),
	})
	signed, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("sign ags client assertion: %w", err)
	}
	return signed, nil
}

func parsePrivateKey(raw string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fmt.Errorf("decode tool private key: invalid PEM")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse tool private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("parse tool private key: unsupported key type")
	}
	return key, nil
}

func hasScope(scopeText, want string) bool {
	for _, scope := range strings.Fields(scopeText) {
		if strings.TrimSpace(scope) == want {
			return true
		}
	}
	return false
}

func randomString(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf)
}
