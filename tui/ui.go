package tui

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danterolle/loqi/translate"
)

func RunBubbleTea(ctx context.Context, backend translate.Backend, langs translate.LanguageProvider) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n%s\n", r, debug.Stack())
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	m := newModel(ctx, backend, langs)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return
}
