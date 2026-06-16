package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danterolle/voca/tui"
)

func main() {
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
