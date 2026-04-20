package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"mycbt/backend/internal/model"
	"mycbt/backend/internal/repo/ltirepo"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/authsvc"
)

type LTIHandler struct {
	lti    *ltirepo.Repo
	user   *userrepo.Repo
	auth   *authsvc.Service
	appURL string
}

func NewLTIHandler(lti *ltirepo.Repo, user *userrepo.Repo, auth *authsvc.Service, appURL string) *LTIHandler {
	return &LTIHandler{lti: lti, user: user, auth: auth, appURL: appURL}
}

// ─── LOGIN INITIATION ─────────────────────────────────────────────────────────

func (h *LTIHandler) LoginInitiation(c *gin.Context) {
	iss := c.Query("iss")
	loginHint := c.Query("login_hint")
	targetLinkURI := c.Query("target_link_uri")
	ltiMessageHit := c.Query("lti_message_hint")

	if iss == "" || loginHint == "" || targetLinkURI == "" {
		c.JSON(400, gin.H{"error": "missing required parameters"})
		return
	}

	platform, ok, err := h.lti.GetPlatformByIssuer(c.Request.Context(), iss)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal error lookup platform"})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": "platform not registered in AtigaCBT"})
		return
	}

	state := randURLString(16)
	nonce := randURLString(16)

	// In a real app, we might store state in a cookie or session.
	// For simplicity, we'll just store the nonce in DB.
	if err := h.lti.StoreNonce(c.Request.Context(), nonce, 5*time.Minute); err != nil {
		c.JSON(500, gin.H{"error": "failed to store nonce"})
		return
	}

	u, _ := url.Parse(platform.OIDCAuthURL)
	q := u.Query()
	q.Set("scope", "openid")
	q.Set("response_type", "id_token")
	q.Set("client_id", platform.ClientID)
	q.Set("redirect_uri", targetLinkURI)
	q.Set("login_hint", loginHint)
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("response_mode", "form_post")
	if ltiMessageHit != "" {
		q.Set("lti_message_hint", ltiMessageHit)
	}
	u.RawQuery = q.Encode()

	// Set state in a cookie for launch verification
	c.SetCookie("lti_state", state, 300, "/", "", false, true)

	c.Redirect(http.StatusFound, u.String())
}

// ─── LTI LAUNCH ───────────────────────────────────────────────────────────────

func (h *LTIHandler) Launch(c *gin.Context) {
	idToken := c.PostForm("id_token")
	state := c.PostForm("state")

	cookieState, err := c.Cookie("lti_state")
	if err != nil || cookieState != state {
		c.JSON(400, gin.H{"error": "invalid state or session expired"})
		return
	}

	// 1. Unverified parse to get issuer
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid token format"})
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid claims"})
		return
	}
	iss, _ := claims["iss"].(string)

	platform, ok, err := h.lti.GetPlatformByIssuer(c.Request.Context(), iss)
	if !ok || err != nil {
		c.JSON(404, gin.H{"error": "issuer not found"})
		return
	}

	// 2. Verify Token Signature using JWKS
	publicKey, err := h.fetchJWKSPublicKey(platform.JWKSURL, token.Header["kid"].(string))
	if err != nil {
		c.JSON(401, gin.H{"error": "failed to verify token signature: " + err.Error()})
		return
	}

	verifiedToken, err := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil || !verifiedToken.Valid {
		c.JSON(401, gin.H{"error": "token verification failed"})
		return
	}
	if verifiedClaims, ok := verifiedToken.Claims.(jwt.MapClaims); ok {
		claims = verifiedClaims
	}

	// 3. Verify Nonce
	nonce, _ := claims["nonce"].(string)
	if valid, _ := h.lti.UseNonce(c.Request.Context(), nonce); !valid {
		c.JSON(401, gin.H{"error": "invalid or reused nonce"})
		return
	}

	// 4. Map or Provision User
	sub, _ := claims["sub"].(string)
	localUserID, found, err := h.lti.FindUserByLTI(c.Request.Context(), platform.ID, sub)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal error finding linked user"})
		return
	}

	if !found {
		// Auto-provision student if it's a student launch
		// LTI 1.3 usually provides name and email
		name, _ := claims["name"].(string)
		email, _ := claims["email"].(string)

		// In a real scenario, we might want to ask user to link or auto-create.
		// Let's at least check if email already exists
		existingUser, ok, err := h.user.GetByUsername(c.Request.Context(), email)
		if err == nil && ok {
			localUserID = existingUser.ID
		} else {
			// Create new student
			hash, _ := authsvc.HashPassword(randURLString(12))
			var err error
			localUserID, err = h.user.Create(c.Request.Context(), model.User{
				Username:     email, // Use email as username for LTI users
				PasswordHash: hash,
				Role:         "student",
				Name:         name,
				Email:        email,
				IsActive:     true,
			})
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to provision new user"})
				return
			}
		}

		// Map them
		h.lti.LinkUser(c.Request.Context(), platform.ID, sub, localUserID)
	}

	// 4.5 Fetch user object for token issuance
	user, ok, err := h.user.GetByID(c.Request.Context(), localUserID)
	if err != nil || !ok {
		c.JSON(500, gin.H{"error": "failed to load linked user"})
		return
	}

	// 5. Check Message Type
	msgType, _ := claims["https://purl.imsglobal.org/spec/lti/claim/message_type"].(string)
	deploymentID, _ := claims["https://purl.imsglobal.org/spec/lti/claim/deployment_id"].(string)

	if msgType == "LtiDeepLinkingRequest" {
		// Store deep linking info for later submission
		dlSettings, _ := claims["https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings"].(map[string]any)
		returnURL, _ := dlSettings["deep_link_return_url"].(string)
		data, _ := dlSettings["data"].(string)

		sessionID, err := h.lti.CreateSession(c.Request.Context(), ltirepo.LTISession{
			PlatformID:   platform.ID,
			LocalUserID:  localUserID,
			MessageType:  msgType,
			ReturnURL:    returnURL,
			Data:         data,
			DeploymentID: deploymentID,
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		})
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to create LTI session"})
			return
		}

		// Issue a short local token for the teacher to pick things in the UI
		localToken, _, _ := h.auth.IssueToken(user)
		c.Redirect(http.StatusFound, fmt.Sprintf("/#/teacher/lti/picker?session_id=%s&token=%s", sessionID, localToken))
		return
	}

	// 6. Issue local JWT and Redirect to Dashboard
	localToken, _, err := h.auth.IssueToken(user)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate session"})
		return
	}

	// Redirect to frontend with token
	targetURL := "/#/student/dashboard"
	if role := user.Role; role == "teacher" {
		targetURL = "/#/teacher/dashboard"
	} else if role == "admin" {
		targetURL = "/#/admin/dashboard"
	}

	// Resource Link Launch: check if there's a specific resource
	if custom, ok := claims["https://purl.imsglobal.org/spec/lti/claim/custom"].(map[string]any); ok {
		if examID, ok := custom["exam_id"].(string); ok {
			h.captureAGSLaunch(c, platform, deploymentID, localUserID, sub, examID, claims)
			targetURL = "/#/student/exams/" + examID + "/join"
		}
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("%s?token=%s", targetURL, localToken))
}

func (h *LTIHandler) captureAGSLaunch(c *gin.Context, platform ltirepo.Platform, deploymentID, localUserID, ltiSub, examID string, claims jwt.MapClaims) {
	endpointClaim, _ := claims["https://purl.imsglobal.org/spec/lti-ags/claim/endpoint"].(map[string]any)
	if len(endpointClaim) == 0 {
		return
	}

	resourceLinkClaim, _ := claims["https://purl.imsglobal.org/spec/lti/claim/resource_link"].(map[string]any)
	resourceLinkID, _ := resourceLinkClaim["id"].(string)
	lineItemURL, _ := endpointClaim["lineitem"].(string)
	lineItemsURL, _ := endpointClaim["lineitems"].(string)
	scopeValues, _ := endpointClaim["scope"].([]any)
	if resourceLinkID == "" || lineItemURL == "" || examID == "" {
		return
	}

	scopes := make([]string, 0, len(scopeValues))
	for _, item := range scopeValues {
		value, _ := item.(string)
		value = strings.TrimSpace(value)
		if value != "" {
			scopes = append(scopes, value)
		}
	}

	_ = h.lti.UpsertAGSLaunchContext(c.Request.Context(), ltirepo.AGSLaunchContext{
		PlatformID:     platform.ID,
		DeploymentID:   deploymentID,
		ResourceLinkID: resourceLinkID,
		ExamID:         examID,
		LocalUserID:    localUserID,
		LTISub:         ltiSub,
		LineItemURL:    lineItemURL,
		LineItemsURL:   lineItemsURL,
		ScopeText:      strings.Join(scopes, " "),
	})
}

// ─── HELPERS ──────────────────────────────────────────────────────────────────

func randURLString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:n]
}

func (h *LTIHandler) fetchJWKSPublicKey(jwksURL, kid string) (*rsa.PublicKey, error) {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid && key.Kty == "RSA" {
			nb, _ := base64.RawURLEncoding.DecodeString(key.N)
			eb, _ := base64.RawURLEncoding.DecodeString(key.E)
			var e int
			for _, b := range eb {
				e = e<<8 | int(b)
			}
			return &rsa.PublicKey{
				N: new(big.Int).SetBytes(nb),
				E: e,
			}, nil
		}
	}

	return nil, errors.New("kid not found in JWKS")
}

// ─── ADMIN: Generate New Tool Keys ─────────────────────────────────────────────

func (h *LTIHandler) GenerateKeys(c *gin.Context) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate keys"})
		return
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	pubKeyBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	c.JSON(200, gin.H{
		"private_key": string(privateKeyPEM),
		"public_key":  string(publicKeyPEM),
	})
}

// ─── ADMIN: Platform CRUD ──────────────────────────────────────────────────────

func (h *LTIHandler) ListPlatforms(c *gin.Context) {
	platforms, err := h.lti.ListPlatforms(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": platforms})
}

func (h *LTIHandler) CreatePlatform(c *gin.Context) {
	var p ltirepo.Platform
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(400, gin.H{"error": "invalid payload"})
		return
	}

	res, err := h.lti.CreatePlatform(c.Request.Context(), p)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"data": res})
}

func (h *LTIHandler) GetPlatform(c *gin.Context) {
	id := c.Param("id")
	p, ok, err := h.lti.GetPlatformByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": "platform not found"})
		return
	}
	c.JSON(200, gin.H{"data": p})
}

func (h *LTIHandler) SubmitDeepLink(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		ExamID    string `json:"exam_id" binding:"required"`
		Title     string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid payload"})
		return
	}

	session, ok, err := h.lti.GetSession(c.Request.Context(), req.SessionID)
	if !ok || err != nil {
		c.JSON(404, gin.H{"error": "LTI session not found or expired"})
		return
	}

	platform, ok, err := h.lti.GetPlatformByID(c.Request.Context(), session.PlatformID)
	if !ok || err != nil {
		c.JSON(404, gin.H{"error": "platform not found"})
		return
	}

	// Sign the Deep Linking Response
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   platform.ClientID,
		"sub":   platform.ClientID,
		"aud":   platform.Issuer, // Audience is the LMS issuer
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(5 * time.Minute).Unix(),
		"nonce": randURLString(16),
		"https://purl.imsglobal.org/spec/lti/claim/message_type":  "LtiDeepLinkingResponse",
		"https://purl.imsglobal.org/spec/lti/claim/version":       "1.3.0",
		"https://purl.imsglobal.org/spec/lti/claim/deployment_id": session.DeploymentID,
		"https://purl.imsglobal.org/spec/lti-dl/claim/data":       session.Data,
		"https://purl.imsglobal.org/spec/lti-dl/claim/content_items": []any{
			map[string]any{
				"type":  "ltiResourceLink",
				"title": req.Title,
				"url":   fmt.Sprintf("%s/api/v1/lti/launch", h.appURL),
				"lineItem": map[string]any{
					"scoreMaximum": 100,
					"label":        req.Title,
					"resourceId":   req.ExamID,
					"tag":          "mycbt-exam",
				},
				"custom": map[string]string{
					"exam_id": req.ExamID,
				},
			},
		},
	})

	privPEM := []byte(platform.ToolPrivateKey)
	block, _ := pem.Decode(privPEM)
	if block == nil {
		c.JSON(500, gin.H{"error": "failed to decode private key"})
		return
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to parse private key"})
		return
	}

	signedToken, err := token.SignedString(privKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to sign response"})
		return
	}

	c.JSON(200, gin.H{
		"data": gin.H{
			"jwt":        signedToken,
			"return_url": session.ReturnURL,
		},
	})
}
