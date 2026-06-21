package argos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBackend_Translate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"translatedText": "Ciao mondo"}`))
	}))
	defer srv.Close()

	b := NewBackend(srv.URL)
	result, err := b.Translate(context.Background(), "Hello world", "en", "it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Ciao mondo" {
		t.Fatalf("expected %q, got %q", "Ciao mondo", result)
	}
}

func TestBackend_EmptyInput(t *testing.T) {
	b := NewBackend("http://localhost:9999")
	result, err := b.Translate(context.Background(), "", "en", "it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty, got %q", result)
	}
}

func TestBackend_SameLang(t *testing.T) {
	b := NewBackend("http://localhost:9999")
	result, err := b.Translate(context.Background(), "hello", "en", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Fatalf("expected unchanged, got %q", result)
	}
}

func TestBackend_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "something broke"}`))
	}))
	defer srv.Close()

	b := NewBackend(srv.URL)
	_, err := b.Translate(context.Background(), "hello", "en", "it")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBackend_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	b := NewBackend(srv.URL)
	_, err := b.Translate(context.Background(), "hello", "en", "it")
	if err == nil {
		t.Fatal("expected decode error")
	}
}
