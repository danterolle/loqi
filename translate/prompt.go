package translate

type PromptBuilder interface {
	System() string
	Translate(text, source, target string) string
}
