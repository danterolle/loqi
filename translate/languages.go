package translate

import (
	"sort"
	"strings"
)

var languages = map[string]string{
	"auto": "Auto",
	"en":   "English",
	"it":   "Italian",
	"fr":   "French",
	"de":   "German",
	"es":   "Spanish",
	"pt":   "Portuguese",
	"nl":   "Dutch",
	"pl":   "Polish",
	"ru":   "Russian",
	"ja":   "Japanese",
	"zh":   "Chinese",
	"ko":   "Korean",
	"ar":   "Arabic",
	"tr":   "Turkish",
	"cs":   "Czech",
	"sv":   "Swedish",
	"da":   "Danish",
	"fi":   "Finnish",
	"el":   "Greek",
	"ro":   "Romanian",
	"hu":   "Hungarian",
	"vi":   "Vietnamese",
	"th":   "Thai",
	"hi":   "Hindi",
}

var langCodes = func() []string {
	codes := make([]string, 0, len(languages))
	for code := range languages {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}()

var _ LanguageProvider = (*staticLanguages)(nil)

type staticLanguages struct{}

func NewStaticLanguages() *staticLanguages {
	return &staticLanguages{}
}

func (s *staticLanguages) List() []Language {
	result := make([]Language, len(langCodes))
	for i, code := range langCodes {
		result[i] = Language{Code: code, Name: languages[code]}
	}
	return result
}

func IsValidLang(code string) bool {
	_, ok := languages[code]
	return ok
}

func ListSupported() string {
	var b strings.Builder
	b.WriteString("supported: ")
	for i, code := range langCodes {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(code)
	}
	return b.String()
}
