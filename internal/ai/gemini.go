package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client é a interface que o app usa para gerar sugestões.
type Client interface {
	Sugerir(ctx context.Context, in PromptInput) (string, error)
}

// GeminiClient chama a API oficial do Google Gemini (generativelanguage.googleapis.com).
type GeminiClient struct {
	HTTP     *http.Client
	APIKey   string
	Model    string // ex: "gemini-1.5-flash"
	APIBase  string // default: https://generativelanguage.googleapis.com/v1beta
}

// NewGemini cria um cliente pronto. O modelo padrão é gemini-2.5-flash
// (a série 1.5 foi descontinuada em 2025).
func NewGemini(apiKey string) *GeminiClient {
	return &GeminiClient{
		HTTP:    &http.Client{Timeout: 30 * time.Second},
		APIKey:  apiKey,
		Model:   "gemini-2.5-flash",
		APIBase: "https://generativelanguage.googleapis.com/v1beta",
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiReq struct {
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  map[string]any  `json:"generationConfig,omitempty"`
}

type geminiResp struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Sugerir monta o prompt e chama a Gemini. Retorna só o texto.
func (g *GeminiClient) Sugerir(ctx context.Context, in PromptInput) (string, error) {
	if g.APIKey == "" {
		return "", errors.New("gemini: API key vazia")
	}
	sys, user := MontaPrompt(in)
	body := geminiReq{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: sys}}},
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: user}}},
		},
		GenerationConfig: map[string]any{
			"temperature":     0.6,
			"maxOutputTokens": 256,
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.APIBase, g.Model, g.APIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("gemini: status %d: %s", resp.StatusCode, string(raw))
	}
	var out geminiResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("gemini: %s", out.Error.Message)
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini: resposta vazia")
	}
	texto := out.Candidates[0].Content.Parts[0].Text
	texto = strings.TrimSpace(texto)
	return texto, nil
}
