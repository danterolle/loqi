package config

import (
	"os"
	"testing"
)

func TestDefault_BackendType(t *testing.T) {
	cfg := Default()
	if cfg.Backend.Type != "ollama" {
		t.Fatalf("expected ollama, got %s", cfg.Backend.Type)
	}
}

func TestDefault_BackendModel(t *testing.T) {
	cfg := Default()
	if cfg.Backend.Model != DefaultModel {
		t.Fatalf("expected %s, got %s", DefaultModel, cfg.Backend.Model)
	}
}

func TestDefault_BackendBaseURL(t *testing.T) {
	cfg := Default()
	if cfg.Backend.BaseURL != DefaultBaseURL {
		t.Fatalf("expected %s, got %s", DefaultBaseURL, cfg.Backend.BaseURL)
	}
}

func TestDefault_BackendOptions(t *testing.T) {
	cfg := Default()
	if cfg.Backend.Options == nil {
		t.Fatal("options should not be nil")
	}
	if cfg.Backend.Options["num_predict"] != float64(2048) {
		t.Fatalf("expected num_predict 2048, got %v", cfg.Backend.Options["num_predict"])
	}
	if cfg.Backend.Options["timeout"] != float64(120) {
		t.Fatalf("expected timeout 120, got %v", cfg.Backend.Options["timeout"])
	}
	if cfg.Backend.Options["temperature"] != float64(0) {
		t.Fatalf("expected temperature 0, got %v", cfg.Backend.Options["temperature"])
	}
	if cfg.Backend.Options["top_p"] != float64(1) {
		t.Fatalf("expected top_p 1, got %v", cfg.Backend.Options["top_p"])
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Backend.Model != DefaultModel {
		t.Fatalf("expected defaults on missing file, got %s", cfg.Backend.Model)
	}
}

func TestLoad_ExplicitFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for explicit nonexistent config file")
	}
}

func TestLoad_CustomFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.yaml"
	data := []byte("backend:\n  model: custom-model\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Backend.Model != "custom-model" {
		t.Fatalf("expected custom-model, got %s", cfg.Backend.Model)
	}
	if cfg.Backend.BaseURL != DefaultBaseURL {
		t.Fatalf("expected default base_url %s, got %s", DefaultBaseURL, cfg.Backend.BaseURL)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/bad.yaml"
	if err := os.WriteFile(path, []byte("backend:\n  model: [invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/empty.yaml"
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Backend.Model != DefaultModel {
		t.Fatalf("expected defaults from empty file, got %s", cfg.Backend.Model)
	}
}

func TestLoad_PartialOverride(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/partial.yaml"
	data := []byte("backend:\n  base_url: http://custom:11434\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Backend.BaseURL != "http://custom:11434" {
		t.Fatalf("expected custom url, got %s", cfg.Backend.BaseURL)
	}
	if cfg.Backend.Model != DefaultModel {
		t.Fatalf("expected default model, got %s", cfg.Backend.Model)
	}
}
