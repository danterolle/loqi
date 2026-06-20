package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danterolle/loqi/translate"
)

func RunBubbleTea(ctx context.Context, backend translate.Backend, langs translate.LanguageProvider) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			// FIXME: log stack trace before recovering?
		}
	}()

	m := newModel(ctx, backend, langs)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return
}
