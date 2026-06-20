package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Minute,
	}
}

func DoTranslate(ctx context.Context, client *http.Client, req *http.Request, errPrefix string) (io.ReadCloser, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errPrefix, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s %s", errPrefix, resp.Status, strings.TrimSpace(string(body)))
	}
	return resp.Body, nil
}

func PostJSON(ctx context.Context, client *http.Client, url, errPrefix string, body any, extract func([]byte) (string, error)) (string, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", fmt.Errorf("%s: encode: %w", errPrefix, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return "", fmt.Errorf("%s: request: %w", errPrefix, err)
	}
	req.Header.Set("Content-Type", "application/json")

	respBody, err := DoTranslate(ctx, client, req, errPrefix)
	if err != nil {
		return "", err
	}
	defer respBody.Close()

	data, err := io.ReadAll(respBody)
	if err != nil {
		return "", fmt.Errorf("%s: read: %w", errPrefix, err)
	}

	return extract(data)
}
