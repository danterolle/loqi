package commands

import (
	"fmt"
	"time"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/translate"
	"github.com/danterolle/voca/translate/llamacpp"
	"github.com/danterolle/voca/translate/ollama"
)

func SetupRun(cfg *config.Config, model string) (*translate.Core, func(), error) {
	printBanner()

	switch cfg.Backend.Type {
	case "ollama":
		return setupOllama(cfg, model)
	case "llamacpp":
		return setupLlamaCpp(cfg, model)
	default:
		return nil, nil, fmt.Errorf("unsupported backend type: %q", cfg.Backend.Type)
	}
}

func setupOllama(cfg *config.Config, model string) (*translate.Core, func(), error) {
	ollamaCmd, started, err := SetupOllama(model, cfg.Backend.BaseURL)
	if err != nil {
		return nil, nil, err
	}

	var cleanup func()
	if started && ollamaCmd != nil {
		c := ollamaCmd
		cleanup = func() {
			ollama.UnloadModel(model, cfg.Backend.BaseURL)
			stopProcess(c)
		}
	} else {
		cleanup = func() { ollama.UnloadModel(model, cfg.Backend.BaseURL) }
	}

	backend := ollama.NewBackend(cfg.Backend.BaseURL, model, translate.NewDefaultPrompt())
	backend.NumPredict = intOption(cfg.Backend.Options, "num_predict", 2048)
	backend.Client.Timeout = durationOption(cfg.Backend.Options, "timeout", 2*time.Minute)
	backend.Temperature = floatOption(cfg.Backend.Options, "temperature", 0.0)
	backend.TopP = floatOption(cfg.Backend.Options, "top_p", 1.0)

	return translate.NewCore(backend, translate.NewStaticLanguages()), cleanup, nil
}

func setupLlamaCpp(cfg *config.Config, model string) (*translate.Core, func(), error) {
	llamaCmd, started, err := SetupLlamaCpp(model, cfg.Backend.BaseURL, cfg.Backend.ModelPath, cfg.Backend.ServerArgs)
	if err != nil {
		return nil, nil, err
	}

	var cleanup func()
	if started && llamaCmd != nil {
		c := llamaCmd
		cleanup = func() { stopProcess(c) }
	} else {
		cleanup = func() {}
	}

	backend := llamacpp.NewBackend(cfg.Backend.BaseURL, model, translate.NewDefaultPrompt())
	backend.MaxTokens = intOption(cfg.Backend.Options, "num_predict", 2048)
	backend.Client.Timeout = durationOption(cfg.Backend.Options, "timeout", 2*time.Minute)
	backend.Temperature = floatOption(cfg.Backend.Options, "temperature", 0.0)
	backend.TopP = floatOption(cfg.Backend.Options, "top_p", 1.0)

	return translate.NewCore(backend, translate.NewStaticLanguages()), cleanup, nil
}

func readFloatOption(options map[string]any, key string) (float64, bool) {
	v, ok := options[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}

func intOption(options map[string]any, key string, defaultVal int) int {
	v, ok := readFloatOption(options, key)
	if !ok {
		return defaultVal
	}
	return int(v)
}

func floatOption(options map[string]any, key string, defaultVal float64) float64 {
	v, ok := readFloatOption(options, key)
	if !ok {
		return defaultVal
	}
	return v
}

func durationOption(options map[string]any, key string, defaultVal time.Duration) time.Duration {
	v, ok := readFloatOption(options, key)
	if !ok {
		return defaultVal
	}
	return time.Duration(v) * time.Second
}
