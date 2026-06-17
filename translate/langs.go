package translate

type Language struct {
	Code string
	Name string
}

type LanguageProvider interface {
	List() []Language
	Lookup(code string) (Language, bool)
}
