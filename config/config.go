package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultConfigDir = ".config/voca"
const DefaultConfigFile = "config.yaml"
const DefaultModel = "gemma4:e2b-it-qat"
const DefaultBaseURL = "http://localhost:11434"

type Config struct {
	Backend BackendConfig `yaml:"backend"`
}

type BackendConfig struct {
	Type    string         `yaml:"type"`
	Model   string         `yaml:"model"`
	BaseURL string         `yaml:"base_url"`
	Options map[string]any `yaml:"options"`
}

func Default() *Config {
	return &Config{
		Backend: BackendConfig{
			Type:    "ollama",
			Model:   DefaultModel,
			BaseURL: DefaultBaseURL,
			Options: map[string]any{
				"temperature": 0.0,
				"num_predict": float64(2048),
				"top_p":       1.0,
				"timeout":     float64(120),
			},
		},
	}
}

func Load(cfgPath string) (*Config, error) {
	cfg := Default()

	paths, explicit := resolvePaths(cfgPath)
	for _, p := range paths {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			if explicit && os.IsNotExist(err) {
				return nil, fmt.Errorf("config: %s not found", p)
			}
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("config: reading %s: %w", p, err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("config: parsing %s: %w", p, err)
		}
		break
	}

	return cfg, nil
}

func resolvePaths(cfgPath string) ([]string, bool) {
	if cfgPath != "" {
		return []string{cfgPath}, true
	}
	if env := os.Getenv("VOCA_CONFIG"); env != "" {
		return []string{env}, true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, true
	}
	return []string{filepath.Join(home, DefaultConfigDir, DefaultConfigFile)}, false
}
