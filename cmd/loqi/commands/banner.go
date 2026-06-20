package commands

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

func printBanner(quiet bool) {
	if quiet {
		return
	}

	colors := []lipgloss.Color{
		"255", "230", "229", "227", "221", "215", "209", "203",
	}
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))

	lines := []string{
		"dP                   oo",
		"88                     ",
		"88 .d8888b. .d8888b. dP",
		"88 88'  `88 88'  `88 88",
		"88 88.  .88 88.  .88 88",
		"dP `88888P' `8888P88 dP",
		"                  88   ",
		"                  dP    ",
	}

	fmt.Fprintln(os.Stderr)
	for i, line := range lines {
		c := colors[i]
		if i >= len(colors) {
			c = colors[len(colors)-1]
		}
		fmt.Fprintln(os.Stderr, lipgloss.NewStyle().Foreground(c).Render(line))
	}
	fmt.Fprintln(os.Stderr)
	if Version != "" {
		fmt.Fprintln(os.Stderr, "       "+accent.Bold(true).Render(Version))
	}
	fmt.Fprintln(os.Stderr, "   "+accent.Render("LOcal Quiet Interpreter"))
	fmt.Fprintln(os.Stderr)
}
