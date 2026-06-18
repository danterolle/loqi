package translate

import "context"

type Core struct {
	Backend   Backend
	Languages LanguageProvider
}

func NewCore(backend Backend, langs LanguageProvider) *Core {
	return &Core{
		Backend:   backend,
		Languages: langs,
	}
}

func (c *Core) Translate(ctx context.Context, text, source, target string) (string, error) {
	return c.Backend.Translate(ctx, text, source, target)
}
