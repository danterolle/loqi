package translate

import (
	"context"
	"testing"
)

func TestMockBackend_ImplementsBackend(t *testing.T) {
	var _ Backend = (*MockBackend)(nil)
}

func TestDefaultPrompt_ImplementsPromptBuilder(t *testing.T) {
	var _ PromptBuilder = (*DefaultPrompt)(nil)
}

func TestStaticLanguages_ImplementsLanguageProvider(t *testing.T) {
	var _ LanguageProvider = (*StaticLanguages)(nil)
}

func TestMockBackend_Translate(t *testing.T) {
	b := NewMockBackend()
	result, err := b.Translate(context.Background(), "hello", "en", "it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "[en->it] hello" {
		t.Fatalf("expected '[en->it] hello', got %q", result)
	}
}

func TestDefaultPrompt_System(t *testing.T) {
	p := NewDefaultPrompt()
	s := p.System()
	if s == "" {
		t.Fatal("system prompt should not be empty")
	}
}

func TestDefaultPrompt_Translate(t *testing.T) {
	p := NewDefaultPrompt()
	result := p.Translate("hello", "en", "it")
	if result == "" {
		t.Fatal("prompt should not be empty")
	}
}

func TestStaticLanguages_List(t *testing.T) {
	l := NewStaticLanguages()
	langs := l.List()
	if len(langs) == 0 {
		t.Fatal("languages list should not be empty")
	}
	if langs[0].Code == "" || langs[0].Name == "" {
		t.Fatal("each language should have Code and Name")
	}
	if langs[0].Code != "ar" {
		t.Fatalf("first language should be 'ar' (sorted), got %q", langs[0].Code)
	}
}

func TestStaticLanguages_Lookup(t *testing.T) {
	l := NewStaticLanguages()
	lang, ok := l.Lookup("it")
	if !ok {
		t.Fatal("expected 'it' to be found")
	}
	if lang.Name != "Italian" {
		t.Fatalf("expected 'Italian', got %q", lang.Name)
	}
	_, ok = l.Lookup("xx")
	if ok {
		t.Fatal("expected 'xx' to not be found")
	}
}

func TestMockBackend_CustomFunc(t *testing.T) {
	b := NewMockBackend()
	b.TranslateFunc = func(ctx context.Context, text, source, target string) (string, error) {
		return "custom", nil
	}
	result, err := b.Translate(context.Background(), "x", "en", "fr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "custom" {
		t.Fatalf("expected 'custom', got %q", result)
	}
}
