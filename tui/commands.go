package tui

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danterolle/voca/translate"
)

func (m Model) doTranslate(text, source, target string) tea.Cmd {
	return func() tea.Msg {
		result, err := translate.Translate(text, source, target, translate.DefaultModel)
		return translateResultMsg{text: text, result: result, err: err}
	}
}

func copyClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
