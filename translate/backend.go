package translate

import "context"

type Backend interface {
	Translate(ctx context.Context, text, source, target string) (string, error)
}
