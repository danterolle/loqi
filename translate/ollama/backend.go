package ollama

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

type chatRequest struct {
	Model    string               `json:"model"`
	Messages []httpclient.Message `json:"messages"`
	Stream   bool                 `json:"stream"`
	Options  map[string]any       `json:"options"`
}

type Backend struct {
	BaseURL     string
	Model       string
	Prompt      promptBuilder
	Client      *http.Client
	NumPredict  int
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

	return httpclient.PostJSON(ctx, b.Client, b.BaseURL+"/api/chat", "ollama",
		b.buildRequestBody(text, source, target),
		func(data []byte) (string, error) {
			var cr struct {
				Message httpclient.Message `json:"message"`
			}
			if err := json.Unmarshal(data, &cr); err != nil {
				return "", fmt.Errorf("ollama: decode: %w", err)
			}
			return strings.TrimSpace(cr.Message.Content), nil
		})
}

func (b *Backend) buildRequestBody(text, source, target string) chatRequest {
	options := map[string]any{
		"temperature": b.Temperature,
		"top_p":       b.TopP,
	}
	if b.NumPredict > 0 {
		options["num_predict"] = b.NumPredict
	}

	return chatRequest{
		Model: b.Model,
		Messages: []httpclient.Message{
			{Role: "system", Content: b.Prompt.System()},
			{Role: "user", Content: b.Prompt.Translate(text, source, target)},
		},
		Stream:  false,
		Options: options,
	}
}
