package llamacpp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	httpclient "github.com/danterolle/loqi/translate/http"
)

type chatCompletionRequest struct {
	Model       string               `json:"model"`
	Messages    []httpclient.Message `json:"messages"`
	Temperature float64              `json:"temperature"`
	TopP        float64              `json:"top_p"`
	MaxTokens   int                  `json:"max_tokens,omitempty"`
	Stream      bool                 `json:"stream"`
}

type Backend struct {
	Config httpclient.BackendConfig
}

func NewBackend(config httpclient.BackendConfig) *Backend {
	if config.Client == nil {
		config.Client = httpclient.NewHTTPClient()
	}
	return &Backend{Config: config}
}

func (b *Backend) Translate(ctx context.Context, text, source, target string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}
	if source == target {
		return text, nil
	}

	return httpclient.PostJSON(ctx, b.Config.Client, b.Config.BaseURL+"/v1/chat/completions", "llamacpp",
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
		Model: b.Config.Model,
		Messages: []httpclient.Message{
			{Role: "system", Content: b.Config.Prompt.System()},
			{Role: "user", Content: b.Config.Prompt.Translate(text, source, target)},
		},
		Temperature: b.Config.Temperature,
		TopP:        b.Config.TopP,
		MaxTokens:   b.Config.MaxTokens,
		Stream:      false,
	}
}
