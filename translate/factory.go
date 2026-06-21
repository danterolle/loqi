package translate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	httpclient "github.com/danterolle/loqi/translate/http"
	"github.com/danterolle/loqi/translate/llamacpp"
	"github.com/danterolle/loqi/translate/ollama"
)

const defaultMaxTokens = 2048

func NewBackend(backendType, baseURL, model string, options map[string]any, prompt *chatPrompt) (Backend, error) {
	config := httpclient.BackendConfig{
		BaseURL:     baseURL,
		Model:       model,
		Prompt:      prompt,
		Client:      httpclient.NewHTTPClient(),
		MaxTokens:   intOption(options, "num_predict", defaultMaxTokens),
		Temperature: floatOption(options, "temperature", 0.0),
		TopP:        floatOption(options, "top_p", 1.0),
	}
	config.Client.Timeout = durationOption(options, "timeout", 2*time.Minute)

	switch backendType {
	case "ollama":
		return ollama.NewBackend(config), nil
	case "llamacpp":
		return llamacpp.NewBackend(config), nil
	default:
		return nil, fmt.Errorf("unsupported backend type: %q", backendType)
	}
}

func UnloadBackend(backendType, model, baseURL string) error {
	if backendType != "ollama" {
		return nil
	}
	body := map[string]string{"model": model, "keep_alive": "0m", "unload": "true"}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unload: encode: %w", err)
	}
	client := httpclient.NewHTTPClient()
	client.Timeout = 30 * time.Second
	resp, err := client.Post(baseURL+"/api/generate", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("unload: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unload: unexpected status: %s", resp.Status)
	}
	return nil
}

func optionAsFloat64(options map[string]any, key string) (float64, bool) {
	v, ok := options[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		if err == nil {
			return f, true
		}
		fmt.Fprintf(os.Stderr, "  ⚠ config: %q has invalid value %q; using default\n", key, n)
		return 0, false
	}
	return 0, false
}

func intOption(options map[string]any, key string, defaultVal int) int {
	v, ok := optionAsFloat64(options, key)
	if !ok {
		return defaultVal
	}
	return int(v)
}

func floatOption(options map[string]any, key string, defaultVal float64) float64 {
	v, ok := optionAsFloat64(options, key)
	if !ok {
		return defaultVal
	}
	return v
}

func durationOption(options map[string]any, key string, defaultVal time.Duration) time.Duration {
	v, ok := optionAsFloat64(options, key)
	if !ok {
		return defaultVal
	}
	return time.Duration(v) * time.Second
}
