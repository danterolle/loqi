package setup

import (
	"fmt"
	"os/exec"

	"github.com/danterolle/loqi/config"
	"github.com/danterolle/loqi/translate"
)

type DiagFunc func(format string, args ...any)

func SetupRun(cfg *config.Config, model string, diag DiagFunc, banner func()) (*translate.Translator, func() error, error) {
	if banner != nil {
		banner()
	}

	var (
		serverStarter func() (*exec.Cmd, bool, error)
		backendType   string
		unloadOnClose bool
	)

	switch cfg.Backend.Type {
	case "ollama":
		serverStarter = func() (*exec.Cmd, bool, error) {
			return SetupOllama(model, cfg.Backend.BaseURL, diag)
		}
		backendType = "ollama"
		unloadOnClose = true
	case "llamacpp":
		serverStarter = func() (*exec.Cmd, bool, error) {
			return SetupLlamaCpp(model, cfg.Backend.BaseURL, cfg.Backend.ModelPath, cfg.Backend.ServerArgs, diag)
		}
		backendType = "llamacpp"
		unloadOnClose = false
	default:
		return nil, nil, fmt.Errorf("unsupported backend type: %q", cfg.Backend.Type)
	}

	serverCmd, started, err := serverStarter()
	if err != nil {
		return nil, nil, err
	}

	var cleanup func() error
	if started && serverCmd != nil {
		c := serverCmd
		cleanup = func() error {
			var err error
			if unloadOnClose {
				err = translate.UnloadBackend(backendType, model, cfg.Backend.BaseURL)
			}
			StopProcess(c)
			return err
		}
	} else if unloadOnClose {
		cleanup = func() error { return translate.UnloadBackend(backendType, model, cfg.Backend.BaseURL) }
	} else {
		cleanup = func() error { return nil }
	}

	backend, err := translate.NewBackend(&translate.NewBackendConfig{
		Type:    backendType,
		BaseURL: cfg.Backend.BaseURL,
		Model:   model,
		Options: cfg.Backend.Options,
		Prompt:  translate.NewChatPrompt(),
	})
	if err != nil {
		return nil, nil, err
	}

	return translate.NewTranslator(backend, translate.NewStaticLanguages()), cleanup, nil
}
