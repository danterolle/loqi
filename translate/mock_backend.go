package translate

import "context"

var _ Backend = (*MockBackend)(nil)

type MockBackend struct {
	TranslateFunc func(ctx context.Context, text, source, target string) (string, error)
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		TranslateFunc: func(ctx context.Context, text, source, target string) (string, error) {
			return "[" + source + "->" + target + "] " + text, nil
		},
	}
}

func (b *MockBackend) Translate(ctx context.Context, text, source, target string) (string, error) {
	return b.TranslateFunc(ctx, text, source, target)
}
