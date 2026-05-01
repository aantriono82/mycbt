package ltisvc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"atigacbt/backend/internal/repo/ltirepo"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func makeTestRSAKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey error: %v", err)
	}
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return string(pem.EncodeToMemory(block))
}

func TestHasScope(t *testing.T) {
	t.Parallel()

	if !hasScope("foo "+AGSScoreScope+" bar", AGSScoreScope) {
		t.Fatal("expected scope to be found")
	}
	if hasScope("foo bar", AGSScoreScope) {
		t.Fatal("expected scope to be absent")
	}
}

func TestParsePrivateKey_InvalidPEM(t *testing.T) {
	t.Parallel()

	if _, err := parsePrivateKey("not-pem"); err == nil {
		t.Fatal("expected parsePrivateKey to fail for invalid pem")
	}
}

func TestBuildClientAssertion(t *testing.T) {
	t.Parallel()

	privateKeyPEM := makeTestRSAKeyPEM(t)
	platform := ltirepo.Platform{
		ClientID:       "client-1",
		OIDCTokenURL:   "https://lms.example.com/token",
		ToolPrivateKey: privateKeyPEM,
	}

	signed, err := buildClientAssertion(platform)
	if err != nil {
		t.Fatalf("buildClientAssertion error: %v", err)
	}

	key, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("parsePrivateKey error: %v", err)
	}
	parsed, err := jwt.Parse(signed, func(token *jwt.Token) (any, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("jwt parse error: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("expected signed assertion to be valid")
	}
	claims, _ := parsed.Claims.(jwt.MapClaims)
	if claims["iss"] != "client-1" || claims["sub"] != "client-1" || claims["aud"] != "https://lms.example.com/token" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestFetchAccessToken_ValidatesResponse(t *testing.T) {
	t.Parallel()

	privateKeyPEM := makeTestRSAKeyPEM(t)
	svc := NewAGSService(newTestClient(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Fatalf("unexpected content-type %q", ct)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll error: %v", err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("ParseQuery error: %v", err)
		}
		if values.Get("scope") != AGSScoreScope {
			t.Fatalf("unexpected scope %q", values.Get("scope"))
		}
		if values.Get("client_assertion") == "" {
			t.Fatal("expected client_assertion")
		}
		respBody, err := json.Marshal(map[string]any{"access_token": "token-123"})
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(respBody))),
		}, nil
	}))

	token, err := svc.fetchAccessToken(context.Background(), ltirepo.Platform{
		ClientID:       "client-1",
		OIDCTokenURL:   "https://lms.example.com/token",
		ToolPrivateKey: privateKeyPEM,
	}, AGSScoreScope)
	if err != nil {
		t.Fatalf("fetchAccessToken error: %v", err)
	}
	if token != "token-123" {
		t.Fatalf("expected token-123, got %q", token)
	}
}

func TestFetchAccessToken_MissingTokenAndRejectedStatus(t *testing.T) {
	t.Parallel()

	privateKeyPEM := makeTestRSAKeyPEM(t)
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "missing token", status: http.StatusOK, body: `{}`},
		{name: "rejected", status: http.StatusBadRequest, body: `{"access_token":"x"}`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := NewAGSService(newTestClient(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tt.status,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				}, nil
			}))

			_, err := svc.fetchAccessToken(context.Background(), ltirepo.Platform{
				ClientID:       "client-1",
				OIDCTokenURL:   "https://lms.example.com/token",
				ToolPrivateKey: privateKeyPEM,
			}, AGSScoreScope)
			if err == nil {
				t.Fatal("expected fetchAccessToken error")
			}
		})
	}
}

func TestPublishScore_ValidatesInputAndSendsRequest(t *testing.T) {
	t.Parallel()

	privateKeyPEM := makeTestRSAKeyPEM(t)
	var gotAuth string
	var gotPayload map[string]any

	svc := NewAGSService(newTestClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/token":
			respBody, err := json.Marshal(map[string]any{"access_token": "token-123"})
			if err != nil {
				t.Fatalf("Marshal token body: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(respBody))),
			}, nil
		case "/line-item/scores":
			gotAuth = r.Header.Get("Authorization")
			if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "vnd.ims.lis.v1.score+json") {
				t.Fatalf("unexpected score content-type %q", ct)
			}
			if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
				t.Fatalf("decode score payload: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
			return nil, errors.New("unexpected path")
		}
	}))

	err := svc.PublishScore(context.Background(), PublishScoreInput{
		Platform: ltirepo.Platform{
			ClientID:       "client-1",
			OIDCTokenURL:   "https://lms.example.com/token",
			ToolPrivateKey: privateKeyPEM,
		},
		LineItemURL: "https://lms.example.com/line-item/",
		LTISub:      "student-1",
		Score:       87,
		Timestamp:   time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	}, "foo "+AGSScoreScope)
	if err != nil {
		t.Fatalf("PublishScore error: %v", err)
	}
	if gotAuth != "Bearer token-123" {
		t.Fatalf("expected bearer token, got %q", gotAuth)
	}
	if gotPayload["userId"] != "student-1" || gotPayload["scoreGiven"] != float64(87) {
		t.Fatalf("unexpected payload: %+v", gotPayload)
	}
}

func TestPublishScore_RejectsMissingURLScopeAndNon2xx(t *testing.T) {
	t.Parallel()

	privateKeyPEM := makeTestRSAKeyPEM(t)
	svc := NewAGSService(&http.Client{Timeout: time.Second})
	platform := ltirepo.Platform{ClientID: "client-1", OIDCTokenURL: "https://example.com/token", ToolPrivateKey: privateKeyPEM}

	if err := svc.PublishScore(context.Background(), PublishScoreInput{Platform: platform}, AGSScoreScope); err == nil {
		t.Fatal("expected error for missing line item url")
	}
	if err := svc.PublishScore(context.Background(), PublishScoreInput{Platform: platform, LineItemURL: "https://example.com/item"}, "foo"); err == nil {
		t.Fatal("expected error for missing AGS scope")
	}

	err := NewAGSService(newTestClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/token" {
			respBody, marshalErr := json.Marshal(map[string]any{"access_token": "token-123"})
			if marshalErr != nil {
				t.Fatalf("Marshal token body: %v", marshalErr)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(respBody))),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})).PublishScore(context.Background(), PublishScoreInput{
		Platform: ltirepo.Platform{
			ClientID:       "client-1",
			OIDCTokenURL:   "https://example.com/token",
			ToolPrivateKey: privateKeyPEM,
		},
		LineItemURL: "https://example.com/line-item",
		LTISub:      "student-1",
		Score:       87,
		Timestamp:   time.Now(),
	}, AGSScoreScope)
	if err == nil {
		t.Fatal("expected non-2xx publish error")
	}
}
