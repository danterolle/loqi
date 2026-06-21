package translate

import (
	"context"
	"sync"
)

const batchWorkers = 3

type jsonTranslator struct {
	tr     *Translator
	from   string
	to     string
	errCh  chan error
	cancel context.CancelFunc
}

func translateJSON(ctx context.Context, tr *Translator, data *any, from, to string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	t := &jsonTranslator{
		tr:     tr,
		from:   from,
		to:     to,
		errCh:  make(chan error, 1),
		cancel: cancel,
	}

	t.processNode(ctx, data)

	select {
	case err := <-t.errCh:
		return err
	default:
		if err := ctx.Err(); err != nil {
			return err
		}
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

func (t *jsonTranslator) processItems(ctx context.Context, n int, fn func(i int)) {
	if n == 0 {
		return
	}
	ch := make(chan int, n)
	for i := 0; i < n; i++ {
		ch <- i
	}
	close(ch)

	workers := min(batchWorkers, n)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				fn(i)
				if ctx.Err() != nil {
					return
				}
			}
		}()
	}
	wg.Wait()
}

// processMapNode extracts values into vals, then dispatches them to
// concurrent workers. json.Unmarshal produces distinct objects per map
// key, so workers never share underlying map/slice references.
func (t *jsonTranslator) processMapNode(ctx context.Context, v map[string]any) {
	keys := make([]string, 0, len(v))
	vals := make([]any, 0, len(v))
	for k, val := range v {
		keys = append(keys, k)
		vals = append(vals, val)
	}
	if len(keys) == 0 {
		return
	}

	var mu sync.Mutex
	t.processItems(ctx, len(keys), func(i int) {
		childCopy := vals[i]
		t.processNode(ctx, &childCopy)
		mu.Lock()
		v[keys[i]] = childCopy
		mu.Unlock()
	})
}

func (t *jsonTranslator) processSliceNode(ctx context.Context, v []any) {
	var mu sync.Mutex
	t.processItems(ctx, len(v), func(i int) {
		childCopy := v[i]
		t.processNode(ctx, &childCopy)
		mu.Lock()
		v[i] = childCopy
		mu.Unlock()
	})
}

func (t *jsonTranslator) translateString(ctx context.Context, val *any) {
	v := (*val).(string)
	result, err := t.tr.Translate(ctx, v, t.from, t.to)

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
