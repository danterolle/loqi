package setup

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"

	"github.com/danterolle/loqi/translate/argos"
	"github.com/danterolle/loqi/translate/llamacpp"
	"github.com/danterolle/loqi/translate/ollama"
)

func StopProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	cmd.Process.Kill()
	cmd.Wait()
}

func SetupOllama(model, baseURL string, diag DiagFunc) (cmd *exec.Cmd, started bool, err error) {
	if !ollama.Reachable(context.Background(), baseURL) {
		if _, err := exec.LookPath("ollama"); err != nil {
			return nil, false, fmt.Errorf("ollama not found — install from https://ollama.com")
		}
		diag("  ◆ Starting Ollama... ")
		cmd = exec.Command("ollama", "serve")
		if err := cmd.Start(); err != nil {
			return nil, false, fmt.Errorf("failed to start Ollama: %w", err)
		}
		started = true
		if !ollama.WaitForReady(30, baseURL) {
			StopProcess(cmd)
			return nil, started, fmt.Errorf("timeout waiting for Ollama to start")
		}
		diag("online\n")
	}

	if !ollama.ModelExists(model, baseURL) {
		diag("  ◆ Pulling %s...\n", model)
		if err := ollama.PullModel(model, baseURL); err != nil {
			StopProcess(cmd)
			return nil, started, fmt.Errorf("pull failed: %w", err)
		}
		diag("  ◆ Model ready\n")
	}

	return cmd, started, nil
}

func SetupLlamaCpp(model, baseURL, modelPath string, serverArgs []string, diag DiagFunc) (cmd *exec.Cmd, started bool, err error) {
	if llamacpp.ServerRunning(baseURL) {
		diag("  ◆ Waiting for model to load...\n")
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

	diag("  ◆ Starting llama-server on %s...\n", u.Host)
	cmd = exec.Command("llama-server", args...)
	if err := cmd.Start(); err != nil {
		return nil, false, fmt.Errorf("failed to start llama-server: %w", err)
	}

	started = true
	if !llamacpp.WaitForModelReady(60, baseURL) {
		StopProcess(cmd)
		return cmd, started, fmt.Errorf("timeout waiting for llama-server to load model")
	}

	diag("  ◆ Online\n")
	return cmd, started, nil
}

func SetupArgos(baseURL string, diag DiagFunc) (cmd *exec.Cmd, started bool, err error) {
	if argos.Reachable(baseURL) {
		return nil, false, nil
	}

	diag("  ◆ Starting argos server...\n")
	cmd, err = argos.StartServer(baseURL)
	if err != nil {
		return nil, false, err
	}
	diag("  ◆ Online\n")
	return cmd, true, nil
}
