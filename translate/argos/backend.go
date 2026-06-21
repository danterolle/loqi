package argos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Backend struct {
	BaseURL string
	Client  *http.Client
}

func NewBackend(baseURL string) *Backend {
	return &Backend{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type translateRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type translateResponse struct {
	TranslatedText string `json:"translatedText"`
	Error          string `json:"error,omitempty"`
}

func (b *Backend) Translate(ctx context.Context, text, source, target string) (string, error) {
	if text == "" {
		return "", nil
	}
	if source == target {
		return text, nil
	}

	reqBody := translateRequest{
		Q:      text,
		Source: source,
		Target: target,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("argos: encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/translate", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("argos: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("argos: do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("argos: read: %w", err)
	}

	var tr translateResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("argos: decode: %w", err)
	}
	if tr.Error != "" {
		return "", fmt.Errorf("argos: %s", tr.Error)
	}

	return tr.TranslatedText, nil
}
