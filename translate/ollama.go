package translate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var Languages = map[string]string{
	"auto": "Auto",
	"en":   "English",
	"it":   "Italian",
	"fr":   "French",
	"de":   "German",
	"es":   "Spanish",
	"pt":   "Portuguese",
	"nl":   "Dutch",
	"pl":   "Polish",
	"ru":   "Russian",
	"ja":   "Japanese",
	"zh":   "Chinese",
	"ko":   "Korean",
	"ar":   "Arabic",
	"tr":   "Turkish",
	"cs":   "Czech",
	"sv":   "Swedish",
	"da":   "Danish",
	"fi":   "Finnish",
	"el":   "Greek",
	"ro":   "Romanian",
	"hu":   "Hungarian",
	"vi":   "Vietnamese",
	"th":   "Thai",
	"hi":   "Hindi",
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string         `json:"model"`
	Messages []message      `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options"`
}

type chatResponse struct {
	Message message `json:"message"`
}

func buildPrompt(text, source, target string) string {
	if source == "auto" {
		return fmt.Sprintf("You are a professional translator. Translate the following text into %s.\nText:\n%s\n\nTranslation:", Languages[target], text)
	}
	return fmt.Sprintf("You are a professional translator. Translate the following text from %s to %s.\nText:\n%s\n\nTranslation:", Languages[source], Languages[target], text)
}

func Translate(text, source, target, model string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}
	if source == target {
		return text, nil
	}

	body := chatRequest{
		Model: model,
		Messages: []message{
			{Role: "system", Content: "You are a translator. Translate the text provided by the user into the requested language. Output ONLY the translation, no explanations or extra text."},
			{Role: "user", Content: buildPrompt(text, source, target)},
		},
		Stream: false,
		Options: map[string]any{
			"temperature": 0.1,
			"num_predict": 1024,
			"top_p":       0.9,
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", &buf)
	if err != nil {
		return "", fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	return strings.TrimSpace(cr.Message.Content), nil
}
