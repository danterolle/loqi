package translate

import (
	"context"
	"encoding/json"
	"fmt"
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
