package translate

type Core struct {
	Backend   Backend
	Prompt    PromptBuilder
	Languages LanguageProvider
	Model     string
}

func NewCore(backend Backend, prompt PromptBuilder, langs LanguageProvider, model string) *Core {
	return &Core{
		Backend:   backend,
		Prompt:    prompt,
		Languages: langs,
		Model:     model,
	}
}
