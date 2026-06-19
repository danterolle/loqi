package tui

import (
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) doTranslate(text, source, target string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.backend.Translate(m.ctx, text, source, target)
		return translateResultMsg{text: text, result: result, err: err}
	}
}

func copyClipboard(text string) error {
	return clipboard.WriteAll(text)
}
