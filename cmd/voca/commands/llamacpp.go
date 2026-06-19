package commands

import (
	"fmt"
	"net/url"
	"os/exec"

	"github.com/danterolle/voca/translate/llamacpp"
)

func SetupLlamaCpp(model, baseURL, modelPath string, serverArgs []string) (cmd *exec.Cmd, started bool, err error) {
	if llamacpp.ServerRunning(baseURL) {
		fmt.Printf("  ◆ Waiting for model to load...\n")
		if !llamacpp.WaitForModelReady(60, baseURL) {
			return nil, false, fmt.Errorf("model not ready at %s", baseURL)
		}
		return nil, false, nil
	}

	if _, lookupErr := exec.LookPath("llama-server"); lookupErr != nil {
		return nil, false, fmt.Errorf("llama-server not found — install from https://github.com/ggml-org/llama.cpp")
	}

	if modelPath == "" {
		return nil, false, fmt.Errorf("llama-server not running at %s — set model_path in config to auto-start, or start it manually", baseURL)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, false, fmt.Errorf("invalid base_url: %w", err)
	}

	args := []string{"--model", modelPath, "--host", u.Hostname()}
	if port := u.Port(); port != "" {
		args = append(args, "--port", port)
	}
	args = append(args, serverArgs...)

	fmt.Printf("  ◆ Starting llama-server on %s...\n", u.Host)
	cmd = exec.Command("llama-server", args...)
	if err := cmd.Start(); err != nil {
		return nil, false, fmt.Errorf("failed to start llama-server: %w", err)
	}

	started = true
	if !llamacpp.WaitForModelReady(60, baseURL) {
		_ = cmd.Process.Kill()
		return cmd, started, fmt.Errorf("timeout waiting for llama-server to load model")
	}

	fmt.Printf("  ◆ Online\n")
	return cmd, started, nil
}
