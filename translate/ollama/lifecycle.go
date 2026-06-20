package ollama

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 2 * time.Second}

func Reachable(baseURL string) bool {
	resp, err := httpClient.Get(baseURL + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func WaitForReady(seconds int, baseURL string) bool {
	for i := 0; i < seconds; i++ {
		if Reachable(baseURL) {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

func ModelExists(model, baseURL string) bool {
	resp, err := httpClient.Get(baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var tags struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return false
	}
	for _, m := range tags.Models {
		if m.Name == model || strings.HasPrefix(m.Name, model+":") {
			return true
		}
	}
	return false
}

func PullModel(model, baseURL string) error {
	body := map[string]any{"name": model, "stream": true}
	var buf strings.Builder
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("ollama pull: encode body: %w", err)
	}

	pullClient := &http.Client{Timeout: 30 * time.Minute}
	resp, err := pullClient.Post(baseURL+"/api/pull", "application/json", strings.NewReader(buf.String()))
	if err != nil {
		return fmt.Errorf("ollama pull: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var s struct {
			Status    string `json:"status"`
			Total     int64  `json:"total,omitempty"`
			Completed int64  `json:"completed,omitempty"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &s); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ ollama pull: skip malformed line: %v\n", err)
			continue
		}
		renderPullStatus(s.Status, s.Total, s.Completed)
	}
	return scanner.Err()
}
