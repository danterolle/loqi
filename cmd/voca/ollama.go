package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/danterolle/voca/translate/ollama"
)

func setupOllama(model string) (*exec.Cmd, bool) {
	if _, err := exec.LookPath("ollama"); err != nil {
		fmt.Fprintf(os.Stderr, "ollama not found. Install it from https://ollama.com\n")
		os.Exit(1)
	}

	started := false
	var cmd *exec.Cmd

	if !ollama.Reachable() {
		fmt.Printf("  ◆ Starting Ollama... ")
		cmd = exec.Command("ollama", "serve")
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "\n  ✖ Failed to start Ollama: %v\n", err)
			os.Exit(1)
		}
		started = true
		if !ollama.WaitForReady(30) {
			fmt.Fprintf(os.Stderr, "  ✖ timeout waiting for Ollama to start\n")
			if cmd != nil {
				_ = cmd.Process.Kill()
			}
			os.Exit(1)
		}
		fmt.Printf("online\n")
	}

	if !ollama.ModelExists(model) {
		fmt.Printf("  ◆ Pulling %s...\n", model)
		if err := ollama.PullModel(model); err != nil {
			fmt.Fprintf(os.Stderr, "  ✖ Pull failed: %v\n", err)
			if started && cmd != nil {
				_ = cmd.Process.Kill()
			}
			os.Exit(1)
		}
		fmt.Printf("  ◆ Model ready\n")
	}

	return cmd, started
}
