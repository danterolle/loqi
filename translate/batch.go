package translate

import (
	"context"
	"encoding/json"
	"fmt"
)

func TranslateJSON(ctx context.Context, core *Core, input []byte, from, to string) ([]byte, error) {
	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	translateValue(ctx, core, &data, from, to)

	return json.MarshalIndent(data, "", "  ")
}

func translateValue(ctx context.Context, core *Core, val *any, from, to string) {
	switch v := (*val).(type) {
	case string:
		if v == "" {
			return
		}
		result, err := core.Backend.Translate(ctx, v, from, to)
		if err == nil {
			*val = result
		}
	case map[string]any:
		for k, child := range v {
			childCopy := child
			translateValue(ctx, core, &childCopy, from, to)
			v[k] = childCopy
		}
	case []any:
		for i, child := range v {
			childCopy := child
			translateValue(ctx, core, &childCopy, from, to)
			v[i] = childCopy
		}
	}
}
