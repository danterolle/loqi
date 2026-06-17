package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danterolle/voca/translate"
	"github.com/danterolle/voca/tui"
)

func main() {
	startedOllama := false

	printBanner()

	if _, err := exec.LookPath("ollama"); err != nil {
		fmt.Printf("  ✖ ollama not found. Install it from https://ollama.com\n")
		os.Exit(1)
	}

	if !ollamaReachable() {
		fmt.Printf("  ◆ Starting Ollama... ")
		cmd := exec.Command("ollama", "serve")
		if err := cmd.Start(); err != nil {
			fmt.Printf("\n  ✖ Failed to start Ollama: %v\n", err)
			os.Exit(1)
		}
		startedOllama = true
		if !waitForOllama(30) {
			fmt.Printf("timeout waiting for Ollama to start\n")
			cmd.Process.Kill()
			os.Exit(1)
		}
		fmt.Printf("online\n")
	}

	if !modelExists() {
		fmt.Printf("  ◆ Pulling %s...\n", translate.DefaultModel)
		if err := pullModel(); err != nil {
			fmt.Printf("  ✖ Pull failed: %v\n", err)
			if startedOllama {
				pkill("ollama")
			}
			os.Exit(1)
		}
		fmt.Printf("  ◆ Model ready\n")
	}

	fmt.Printf("\n  Starting terminal interface...\n\n")

	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
	}

	if startedOllama {
		pkill("ollama")
	}
}

func printBanner() {
	fmt.Println()
	fmt.Println("  ██╗   ██╗ ██████╗  ██████╗ █████╗ ")
	fmt.Println("  ██║   ██║██╔═══██╗██╔════╝██╔══██╗")
	fmt.Println("  ██║   ██║██║   ██║██║     ███████║")
	fmt.Println("  ╚██╗ ██╔╝██║   ██║██║     ██╔══██║")
	fmt.Println("   ╚████╔╝ ╚██████╔╝╚██████╗██║  ██║")
	fmt.Println("    ╚═══╝   ╚═════╝  ╚═════╝╚═╝  ╚═╝")
	fmt.Println()
}

func ollamaReachable() bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func waitForOllama(seconds int) bool {
	for i := 0; i < seconds; i++ {
		if ollamaReachable() {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

func modelExists() bool {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var tags struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	json.NewDecoder(resp.Body).Decode(&tags)
	for _, m := range tags.Models {
		if m.Name == translate.DefaultModel || strings.HasPrefix(m.Name, translate.DefaultModel+":") {
			return true
		}
	}
	return false
}

func pullModel() error {
	body := map[string]any{"name": translate.DefaultModel, "stream": true}
	var buf strings.Builder
	json.NewEncoder(&buf).Encode(body)

	resp, err := http.Post("http://localhost:11434/api/pull", "application/json", strings.NewReader(buf.String()))
	if err != nil {
		return fmt.Errorf("ollama pull: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var s struct {
			Status    string `json:"status"`
			Total     int64  `json:"total,omitempty"`
			Completed int64  `json:"completed,omitempty"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &s); err != nil {
			continue
		}
		if s.Total > 0 {
			pct := float64(s.Completed) / float64(s.Total) * 100
			bar := progressBar(pct, 30)
			fmt.Printf("\r     %s  %.0f%%", bar, pct)
		} else if s.Status == "success" {
			fmt.Printf("\r     %s  100%%\n", progressBar(100, 30))
		} else if strings.Contains(s.Status, "pulling") {
			parts := strings.SplitN(s.Status, " ", 2)
			if len(parts) == 2 {
				short := parts[1]
				if len(short) > 12 {
					short = short[:12]
				}
				fmt.Printf("\r     Pulling %s...", short)
			}
		} else if s.Status == "verifying sha256 digest" {
			fmt.Printf("\r     Verifying...")
		} else if s.Status == "writing manifest" {
			fmt.Printf("\r     Writing manifest...")
		} else {
			fmt.Printf("\r     %s", s.Status)
		}
	}
	return scanner.Err()
}

func progressBar(pct float64, width int) string {
	filled := int(pct * float64(width) / 100)
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

func pkill(name string) {
	exec.Command("pkill", "-f", name).Run()
}
