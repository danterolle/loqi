package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) doTranslate(text, source, target string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.core.Backend.Translate(context.Background(), text, source, target)
		return translateResultMsg{text: text, result: result, err: err}
	}
}

func copyClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		return fmt.Errorf("copyClipboard: unsupported platform %s", runtime.GOOS)
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
