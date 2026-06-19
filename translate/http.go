package translate

import (
	"context"
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
