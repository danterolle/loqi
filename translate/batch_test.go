package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestBatch_PlainText(t *testing.T) {
	core := NewCore(NewMockBackend(), NewDefaultPrompt(), NewStaticLanguages(), "test")
	result, err := Batch(context.Background(), core, []byte("hello"), "en", "it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "[en->it] hello" {
		t.Fatalf("expected '[en->it] hello', got %q", string(result))
	}
}

func assertJSONResult(t *testing.T, got []byte, want any) {
	t.Helper()
	var gotVal any
	if err := json.Unmarshal(got, &gotVal); err != nil {
		t.Fatalf("invalid JSON result: %v\n%s", err, string(got))
	}
	wantJSON, _ := json.Marshal(want)
	var wantVal any
	json.Unmarshal(wantJSON, &wantVal)

	gotStr, _ := json.Marshal(gotVal)
	wantStr, _ := json.Marshal(wantVal)
	if string(gotStr) != string(wantStr) {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", string(wantJSON), string(got))
	}
}

func TestBatch_FromFixture(t *testing.T) {
	core := NewCore(NewMockBackend(), NewDefaultPrompt(), NewStaticLanguages(), "test")
	input, err := os.ReadFile("../test_data/i18n.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	result, err := Batch(context.Background(), core, input, "en", "it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got any
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("invalid JSON result: %v", err)
	}
	// Verify flat: all string leaf values were translated
	m := got.(map[string]any)
	app := m["app"].(map[string]any)
	if app["title"] != "[en->it] Welcome to Voca" {
		t.Fatalf("expected translated title, got %q", app["title"])
	}
	// Verify non-string leaf values preserved
	settings := m["settings"].(map[string]any)
	if settings["notifications"] != true {
		t.Fatal("non-string field 'notifications' should be preserved")
	}
	if settings["items_per_page"] != float64(25) {
		t.Fatal("non-string field 'items_per_page' should be preserved")
	}
	// Verify long sentences translated
	errors := m["errors"].(map[string]any)
	if errors["not_found"] != "[en->it] The requested resource could not be found on this server. Please check the URL and try again." {
		t.Fatalf("unexpected long translation, got %q", errors["not_found"])
	}
	// Verify nested structure preserved
	menu := m["menu"].(map[string]any)
	file := menu["file"].(map[string]any)
	if file["save"] != "[en->it] Save" {
		t.Fatalf("expected translated 'Save', got %q", file["save"])
	}
}

func TestBatch_FlatJSON(t *testing.T) {
	core := NewCore(NewMockBackend(), NewDefaultPrompt(), NewStaticLanguages(), "test")
	input := []byte(`{"a": "hello", "b": "world"}`)
	result, err := Batch(context.Background(), core, input, "en", "it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertJSONResult(t, result, map[string]any{
		"a": "[en->it] hello",
		"b": "[en->it] world",
	})
}

func TestBatch_NestedJSON(t *testing.T) {
	core := NewCore(NewMockBackend(), NewDefaultPrompt(), NewStaticLanguages(), "test")
	input := []byte(`{"greeting": {"en": "hello", "fr": "bonjour"}, "count": 42}`)
	result, err := Batch(context.Background(), core, input, "en", "it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertJSONResult(t, result, map[string]any{
		"count": float64(42),
		"greeting": map[string]any{
			"en": "[en->it] hello",
			"fr": "[en->it] bonjour",
		},
	})
}

func TestBatch_JSONPreservesNonString(t *testing.T) {
	core := NewCore(NewMockBackend(), NewDefaultPrompt(), NewStaticLanguages(), "test")
	input := []byte(`{"name": "hello", "count": 42, "active": true, "tags": null}`)
	result, err := Batch(context.Background(), core, input, "en", "it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertJSONResult(t, result, map[string]any{
		"name":   "[en->it] hello",
		"count":  float64(42),
		"active": true,
		"tags":   nil,
	})
}

func TestBatch_JSONErrorPropagation(t *testing.T) {
	mock := NewMockBackend()
	mock.TranslateFunc = func(ctx context.Context, text, source, target string) (string, error) {
		if text == "fail" {
			return "", fmt.Errorf("translation failed")
		}
		return "[" + source + "->" + target + "] " + text, nil
	}
	core := NewCore(mock, NewDefaultPrompt(), NewStaticLanguages(), "test")
	input := []byte(`{"ok": "hello", "bad": "fail"}`)
	_, err := Batch(context.Background(), core, input, "en", "it")
	if err == nil {
		t.Fatal("expected error for failing translation")
	}
}
