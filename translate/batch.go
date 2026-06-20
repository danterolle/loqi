package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func Batch(ctx context.Context, tr *Translator, input []byte, from, to string) ([]byte, error) {
	if json.Valid(input) {
		var data any
		if err := json.Unmarshal(input, &data); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		if err := translateJSON(ctx, tr, &data, from, to); err != nil {
			return nil, err
		}
		return json.MarshalIndent(data, "", "  ")
	}

	text := strings.TrimSpace(string(input))
	if text == "" {
		return nil, fmt.Errorf("empty input")
	}

	result, err := tr.Translate(ctx, text, from, to)
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}
