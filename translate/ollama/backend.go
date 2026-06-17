package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/danterolle/voca/translate"
)

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string         `json:"model"`
	Messages []message      `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options"`
}

type chatResponse struct {
	Message message `json:"message"`
}

type Backend struct {
	BaseURL string
	Model   string
	Prompt  translate.PromptBuilder
	Client  *http.Client
}

func NewBackend(baseURL, model string, prompt translate.PromptBuilder) *Backend {
	return &Backend{
		BaseURL: baseURL,
		Model:   model,
		Prompt:  prompt,
		Client: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func (b *Backend) Translate(ctx context.Context, text, source, target string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}
	if source == target {
		return text, nil
	}

	body := chatRequest{
		Model: b.Model,
		Messages: []message{
			{Role: "system", Content: b.Prompt.System()},
			{Role: "user", Content: b.Prompt.Translate(text, source, target)},
		},
		Stream: false,
		Options: map[string]any{
			"temperature": 0.0,
			"num_predict": 512,
			"top_p":       1.0,
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/api/chat", &buf)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	return strings.TrimSpace(cr.Message.Content), nil
}
