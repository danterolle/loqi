package llamacpp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	httpclient "github.com/danterolle/loqi/translate/http"
)

type promptBuilder interface {
	System() string
	Translate(text, source, target string) string
}

type chatCompletionRequest struct {
	Model       string               `json:"model"`
	Messages    []httpclient.Message `json:"messages"`
	Temperature float64              `json:"temperature,omitempty"`
	TopP        float64              `json:"top_p,omitempty"`
	MaxTokens   int                  `json:"max_tokens,omitempty"`
	Stream      bool                 `json:"stream"`
}

type Backend struct {
	BaseURL     string
	Model       string
	Prompt      promptBuilder
	Client      *http.Client
	MaxTokens   int
	Temperature float64
	TopP        float64
}

func NewBackend(baseURL, model string, prompt promptBuilder) *Backend {
	return &Backend{
		BaseURL: baseURL,
		Model:   model,
		Prompt:  prompt,
		Client:  httpclient.NewHTTPClient(),
	}
}

func (b *Backend) Translate(ctx context.Context, text, source, target string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}
	if source == target {
		return text, nil
	}

	return httpclient.PostJSON(ctx, b.Client, b.BaseURL+"/v1/chat/completions", "llamacpp",
		b.buildRequestBody(text, source, target),
		func(data []byte) (string, error) {
			var cr struct {
				Choices []struct {
					Message httpclient.Message `json:"message"`
				} `json:"choices"`
			}
			if err := json.Unmarshal(data, &cr); err != nil {
				return "", fmt.Errorf("llamacpp: decode: %w", err)
			}
			if len(cr.Choices) == 0 {
				return "", fmt.Errorf("llamacpp: empty response")
			}
			return strings.TrimSpace(cr.Choices[0].Message.Content), nil
		})
}

func (b *Backend) buildRequestBody(text, source, target string) chatCompletionRequest {
	return chatCompletionRequest{
		Model: b.Model,
		Messages: []httpclient.Message{
			{Role: "system", Content: b.Prompt.System()},
			{Role: "user", Content: b.Prompt.Translate(text, source, target)},
		},
		Temperature: b.Temperature,
		TopP:        b.TopP,
		MaxTokens:   b.MaxTokens,
		Stream:      false,
	}
}
