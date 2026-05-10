package essayaisvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"atigacbt/backend/internal/config"
)

type SuggestionInput struct {
	QuestionText string
	RubricText   string
	AnswerText   string
	MaxScore     int
}

type Suggestion struct {
	RecommendedScore int      `json:"recommended_score"`
	Feedback         string   `json:"feedback"`
	ReasoningSummary []string `json:"reasoning_summary"`
	Confidence       string   `json:"confidence"`
	Provider         string   `json:"provider"`
	Model            string   `json:"model"`
}

type Service struct {
	enabled       bool
	provider      string
	model         string
	timeout       time.Duration
	maxTokens     int
	geminiAPIKey  string
	ollamaBaseURL string
	httpClient    *http.Client
}

func New(cfg config.Config) *Service {
	enabled := strings.EqualFold(strings.TrimSpace(cfg.EssayAIEnabled), "true") || strings.TrimSpace(cfg.EssayAIEnabled) == "1"
	timeoutMS, _ := strconv.Atoi(strings.TrimSpace(cfg.EssayAITimeoutMS))
	if timeoutMS <= 0 {
		timeoutMS = 15000
	}
	maxTokens, _ := strconv.Atoi(strings.TrimSpace(cfg.EssayAIMaxTokens))
	if maxTokens <= 0 {
		maxTokens = 800
	}
	return &Service{
		enabled:       enabled,
		provider:      strings.ToLower(strings.TrimSpace(cfg.EssayAIProvider)),
		model:         strings.TrimSpace(cfg.EssayAIModel),
		timeout:       time.Duration(timeoutMS) * time.Millisecond,
		maxTokens:     maxTokens,
		geminiAPIKey:  strings.TrimSpace(cfg.GeminiAPIKey),
		ollamaBaseURL: strings.TrimSpace(cfg.OllamaBaseURL),
		httpClient:    &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond},
	}
}

func (s *Service) Enabled() bool {
	return s != nil && s.enabled
}

func (s *Service) Suggest(ctx context.Context, in SuggestionInput) (Suggestion, error) {
	if s == nil || !s.enabled {
		return Suggestion{}, errors.New("essay ai is disabled")
	}
	switch s.provider {
	case "gemini":
		return s.suggestGemini(ctx, in)
	case "ollama":
		return s.suggestOllama(ctx, in)
	default:
		return Suggestion{}, fmt.Errorf("unsupported essay ai provider: %s", s.provider)
	}
}

func (s *Service) suggestGemini(ctx context.Context, in SuggestionInput) (Suggestion, error) {
	if s.geminiAPIKey == "" {
		return Suggestion{}, errors.New("gemini api key is not configured")
	}
	model := s.model
	if model == "" {
		model = "gemini-2.5-flash"
	}
	body := map[string]any{
		"contents": []any{
			map[string]any{
				"parts": []any{
					map[string]any{"text": buildPrompt(in)},
				},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": s.maxTokens,
			"temperature":     0.2,
		},
	}
	raw, err := s.postJSON(ctx, fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, s.geminiAPIKey), body, nil)
	if err != nil {
		return Suggestion{}, err
	}
	var resp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return Suggestion{}, fmt.Errorf("parse gemini response: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return Suggestion{}, errors.New("gemini returned empty response")
	}
	sug, err := parseSuggestionJSON(resp.Candidates[0].Content.Parts[0].Text, in.MaxScore)
	if err != nil {
		return Suggestion{}, err
	}
	sug.Provider = "gemini"
	sug.Model = model
	return sug, nil
}

func (s *Service) suggestOllama(ctx context.Context, in SuggestionInput) (Suggestion, error) {
	base := s.ollamaBaseURL
	if base == "" {
		base = "http://127.0.0.1:11434"
	}
	model := s.model
	if model == "" {
		model = "llama3.1:8b-instruct"
	}
	body := map[string]any{
		"model":  model,
		"prompt": buildPrompt(in),
		"stream": false,
		"options": map[string]any{
			"num_predict": s.maxTokens,
			"temperature": 0.2,
		},
	}
	raw, err := s.postJSON(ctx, strings.TrimRight(base, "/")+"/api/generate", body, nil)
	if err != nil {
		return Suggestion{}, err
	}
	var resp struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return Suggestion{}, fmt.Errorf("parse ollama response: %w", err)
	}
	sug, err := parseSuggestionJSON(resp.Response, in.MaxScore)
	if err != nil {
		return Suggestion{}, err
	}
	sug.Provider = "ollama"
	sug.Model = model
	return sug, nil
}

func (s *Service) postJSON(ctx context.Context, endpoint string, payload any, headers map[string]string) ([]byte, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ai provider: %w", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ai response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := strings.TrimSpace(string(out))
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return nil, fmt.Errorf("ai provider status %d: %s", resp.StatusCode, msg)
	}
	return out, nil
}

func buildPrompt(in SuggestionInput) string {
	return fmt.Sprintf(`Anda asisten penilai essay. Kembalikan JSON valid saja tanpa teks lain.
Skala score final harus 0 sampai %d.
Jika jawaban kosong atau tidak relevan, score rendah.
Gunakan confidence: low|medium|high.
reasoning_summary maksimal 4 poin, setiap poin ringkas.

Pertanyaan:
%s

Rubrik/Kunci (opsional):
%s

Jawaban siswa:
%s

Format JSON:
{
  "recommended_score": 0,
  "confidence": "low|medium|high",
  "reasoning_summary": ["..."],
  "feedback": "..."
}`, in.MaxScore, in.QuestionText, in.RubricText, in.AnswerText)
}

func parseSuggestionJSON(raw string, maxScore int) (Suggestion, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return Suggestion{}, errors.New("ai response is not valid json")
	}
	var out Suggestion
	if err := json.Unmarshal([]byte(raw[start:end+1]), &out); err != nil {
		return Suggestion{}, fmt.Errorf("parse suggestion json: %w", err)
	}
	if out.Confidence == "" {
		out.Confidence = "low"
	}
	out.Confidence = strings.ToLower(strings.TrimSpace(out.Confidence))
	if out.Confidence != "low" && out.Confidence != "medium" && out.Confidence != "high" {
		out.Confidence = "low"
	}
	if out.RecommendedScore < 0 {
		out.RecommendedScore = 0
	}
	if out.RecommendedScore > maxScore {
		out.RecommendedScore = maxScore
	}
	out.Feedback = strings.TrimSpace(out.Feedback)
	if len(out.ReasoningSummary) > 4 {
		out.ReasoningSummary = out.ReasoningSummary[:4]
	}
	return out, nil
}
