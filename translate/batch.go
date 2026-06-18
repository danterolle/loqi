package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const batchWorkers = 3

type jsonTranslator struct {
	core   *Core
	from   string
	to     string
	wg     sync.WaitGroup
	errCh  chan error
	cancel context.CancelFunc
	sem    chan struct{}
}

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

	t := &jsonTranslator{
		core:   core,
		from:   from,
		to:     to,
		errCh:  make(chan error, 1),
		cancel: cancel,
		sem:    make(chan struct{}, batchWorkers),
	}

	t.processNode(ctx, data)

	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	select {
	case err := <-t.errCh:
		return err
	default:
		return nil
	}
}

func (t *jsonTranslator) processNode(ctx context.Context, val *any) {
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
		t.translateString(ctx, val)

	case map[string]any:
		t.processMapNode(ctx, v)

	case []any:
		t.processSliceNode(ctx, v)
	}
}

func (t *jsonTranslator) processMapNode(ctx context.Context, v map[string]any) {
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
		t.wg.Add(1)
		go func() {
			defer t.wg.Done()
			childCopy := e.val
			t.processNode(ctx, &childCopy)
			if ctx.Err() != nil {
				return
			}
			mu.Lock()
			v[e.key] = childCopy
			mu.Unlock()
		}()
	}
}

func (t *jsonTranslator) processSliceNode(ctx context.Context, v []any) {
	for i, child := range v {
		t.wg.Add(1)
		go func() {
			defer t.wg.Done()
			childCopy := child
			t.processNode(ctx, &childCopy)
			if ctx.Err() != nil {
				return
			}
			v[i] = childCopy
		}()
	}
}

func (t *jsonTranslator) translateString(ctx context.Context, val *any) {
	select {
	case t.sem <- struct{}{}:
	case <-ctx.Done():
		return
	}

	v := (*val).(string)
	result, err := t.core.Translate(ctx, v, t.from, t.to)
	<-t.sem

	if err != nil {
		select {
		case t.errCh <- err:
		default:
		}
		t.cancel()
		return
	}
	*val = result
}
