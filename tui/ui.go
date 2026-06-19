package tui

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danterolle/voca/translate"
)

func RunBubbleTea(ctx context.Context, core *translate.Core) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "  ✖ panic: %v\n", r)
			os.Exit(1)
		}
	}()

	m := newModel(ctx, core)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
