package translate

import (
	"context"
	"sync"
)

const batchWorkers = 3

type jsonTranslator struct {
	core   *Core
	from   string
	to     string
	errCh  chan error
	cancel context.CancelFunc
	sem    chan struct{}
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
	if len(entries) == 0 {
		return
	}
	ch := make(chan entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)

	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < batchWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range ch {
				childCopy := e.val
				t.processNode(ctx, &childCopy)
				if ctx.Err() != nil {
					return
				}
				mu.Lock()
				v[e.key] = childCopy
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
}

func (t *jsonTranslator) processSliceNode(ctx context.Context, v []any) {
	if len(v) == 0 {
		return
	}
	ch := make(chan int, len(v))
	for i := range v {
		ch <- i
	}
	close(ch)

	var wg sync.WaitGroup
	for i := 0; i < batchWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range ch {
				childCopy := v[idx]
				t.processNode(ctx, &childCopy)
				if ctx.Err() != nil {
					return
				}
				v[idx] = childCopy
			}
		}()
	}
	wg.Wait()
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
