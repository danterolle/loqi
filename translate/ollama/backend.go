package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/danterolle/voca/translate"
)

type chatRequest struct {
	Model    string              `json:"model"`
	Messages []translate.Message `json:"messages"`
	Stream   bool                `json:"stream"`
	Options  map[string]any      `json:"options"`
}

type chatResponse struct {
	Message translate.Message `json:"message"`
}

type Backend struct {
	BaseURL     string
	Model       string
	Prompt      translate.PromptBuilder
	Client      *http.Client
	NumPredict  int
	Temperature float64
	TopP        float64
}

func NewBackend(baseURL, model string, prompt translate.PromptBuilder) *Backend {
	return &Backend{
		BaseURL: baseURL,
		Model:   model,
		Prompt:  prompt,
		Client:  translate.NewHTTPClient(),
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

	body, err := translate.DoTranslate(ctx, b.Client, req, "ollama")
	if err != nil {
		return "", err
	}
	defer body.Close()

	var cr chatResponse
	if err := json.NewDecoder(body).Decode(&cr); err != nil {
		return "", fmt.Errorf("ollama: decode: %w", err)
	}

	return strings.TrimSpace(cr.Message.Content), nil
}

func (b *Backend) buildRequest(ctx context.Context, text, source, target string) (*http.Request, error) {
	options := map[string]any{
		"temperature": b.Temperature,
		"top_p":       b.TopP,
	}
	if b.NumPredict > 0 {
		options["num_predict"] = b.NumPredict
	}

	body := chatRequest{
		Model: b.Model,
		Messages: []translate.Message{
			{Role: "system", Content: b.Prompt.System()},
			{Role: "user", Content: b.Prompt.Translate(text, source, target)},
		},
		Stream:  false,
		Options: options,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("ollama: encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/api/chat", &buf)
	if err != nil {
		return nil, fmt.Errorf("ollama: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
