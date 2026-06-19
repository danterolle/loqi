package llamacpp

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

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type Backend struct {
	BaseURL     string
	Model       string
	Prompt      translate.PromptBuilder
	Client      *http.Client
	MaxTokens   int
	Temperature float64
	TopP        float64
}

func NewBackend(baseURL, model string, prompt translate.PromptBuilder) *Backend {
	return &Backend{
		BaseURL:     baseURL,
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   2048,
		Temperature: 0.0,
		TopP:        1.0,
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

	req, err := b.buildRequest(ctx, text, source, target)
	if err != nil {
		return "", err
	}

	resp, err := b.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llamacpp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llamacpp: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var cr chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("llamacpp: decode: %w", err)
	}

	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("llamacpp: empty response")
	}

	return strings.TrimSpace(cr.Choices[0].Message.Content), nil
}

func (b *Backend) buildRequest(ctx context.Context, text, source, target string) (*http.Request, error) {
	body := chatCompletionRequest{
		Model: b.Model,
		Messages: []chatMessage{
			{Role: "system", Content: b.Prompt.System()},
			{Role: "user", Content: b.Prompt.Translate(text, source, target)},
		},
		Temperature: b.Temperature,
		TopP:        b.TopP,
		MaxTokens:   b.MaxTokens,
		Stream:      false,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("llamacpp: encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/v1/chat/completions", &buf)
	if err != nil {
		return nil, fmt.Errorf("llamacpp: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
