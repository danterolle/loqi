package commands

import (
	"fmt"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/translate"
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
			translate.UnloadBackend("ollama", model, cfg.Backend.BaseURL)
			stopProcess(c)
		}
	} else {
		cleanup = func() { translate.UnloadBackend("ollama", model, cfg.Backend.BaseURL) }
	}

	backend, err := translate.NewBackend("ollama", cfg.Backend.BaseURL, model, cfg.Backend.Options, translate.NewDefaultPrompt())
	if err != nil {
		return nil, nil, err
	}

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

	backend, err := translate.NewBackend("llamacpp", cfg.Backend.BaseURL, model, cfg.Backend.Options, translate.NewDefaultPrompt())
	if err != nil {
		return nil, nil, err
	}

	return translate.NewCore(backend, translate.NewStaticLanguages()), cleanup, nil
}
