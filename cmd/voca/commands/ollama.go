package commands

import (
	"fmt"
	"os/exec"

	"github.com/danterolle/voca/translate/ollama"
)

func SetupOllama(model, baseURL string) (cmd *exec.Cmd, started bool, err error) {
	if _, err := exec.LookPath("ollama"); err != nil {
		return nil, false, fmt.Errorf("ollama not found — install from https://ollama.com")
	}

	if !ollama.Reachable(baseURL) {
		logDiag("  ◆ Starting Ollama... ")
		cmd = exec.Command("ollama", "serve")
		if err := cmd.Start(); err != nil {
			return nil, false, fmt.Errorf("failed to start Ollama: %w", err)
		}
		started = true
		if !ollama.WaitForReady(30, baseURL) {
			stopProcess(cmd)
			return nil, started, fmt.Errorf("timeout waiting for Ollama to start")
		}
		logDiag("online\n")
	}

	if !ollama.ModelExists(model, baseURL) {
		logDiag("  ◆ Pulling %s...\n", model)
		if err := ollama.PullModel(model, baseURL); err != nil {
			stopProcess(cmd)
			return nil, started, fmt.Errorf("pull failed: %w", err)
		}
		logDiag("  ◆ Model ready\n")
	}

	return cmd, started, nil
}
