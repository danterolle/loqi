package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const batchWorkers = 3

var translateSem = make(chan struct{}, batchWorkers)

func Batch(ctx context.Context, core *Core, input []byte, from, to string) ([]byte, error) {
	if json.Valid(input) {
		var data any
		if err := json.Unmarshal(input, &data); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		if err := translateJSON(ctx, core, &data, from, to); err != nil {
			return nil, err
		}
		return json.MarshalIndent(data, "", "  ")
	}

	text := strings.TrimSpace(string(input))
	if text == "" {
		return nil, fmt.Errorf("empty input")
	}

	result, err := core.Translate(ctx, text, from, to)
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

func translateJSON(ctx context.Context, core *Core, data *any, from, to string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	processNode(ctx, core, data, from, to, &wg, errCh, cancel)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func processNode(ctx context.Context, core *Core, val *any, from, to string, wg *sync.WaitGroup, errCh chan error, cancel context.CancelFunc) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	switch v := (*val).(type) {
	case string:
		if v == "" {
			return
		}
		translateString(ctx, core, val, from, to, errCh, cancel)

	case map[string]any:
		type entry struct {
			key string
			val any
		}
		entries := make([]entry, 0, len(v))
		for k, child := range v {
			entries = append(entries, entry{k, child})
		}
		var mu sync.Mutex
		for _, e := range entries {
			wg.Add(1)
			go func() {
				defer wg.Done()
				childCopy := e.val
				processNode(ctx, core, &childCopy, from, to, wg, errCh, cancel)
				if ctx.Err() != nil {
					return
				}
				mu.Lock()
				v[e.key] = childCopy
				mu.Unlock()
			}()
		}

	case []any:
		for i, child := range v {
			wg.Add(1)
			go func() {
				defer wg.Done()
				childCopy := child
				processNode(ctx, core, &childCopy, from, to, wg, errCh, cancel)
				if ctx.Err() != nil {
					return
				}
				v[i] = childCopy
			}()
		}
	}
}

func translateString(ctx context.Context, core *Core, val *any, from, to string, errCh chan error, cancel context.CancelFunc) {
	select {
	case translateSem <- struct{}{}:
	case <-ctx.Done():
		return
	}

	v := (*val).(string)
	result, err := core.Translate(ctx, v, from, to)
	<-translateSem

	if err != nil {
		select {
		case errCh <- err:
		default:
		}
		cancel()
		return
	}
	*val = result
}
