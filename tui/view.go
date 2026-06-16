package tui

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	var b strings.Builder

	srcName := m.langNames[m.langCodes[m.srcIdx]]
	tgtName := m.langNames[m.langCodes[m.tgtIdx]]

	b.WriteString(headerStyle.Render("voca"))
	b.WriteString("  ")
	if m.focused == focusSrcLang {
		b.WriteString(subtleStyle.Render("From:"))
		b.WriteString(" ")
		b.WriteString(inputStyle.Bold(true).Render(srcName))
	} else {
		b.WriteString(fmt.Sprintf("From: %s", srcName))
	}
	b.WriteString("  ->  ")
	if m.focused == focusTgtLang {
		b.WriteString(subtleStyle.Render("To:"))
		b.WriteString(" ")
		b.WriteString(outputStyle.Bold(true).Render(tgtName))
	} else {
		b.WriteString(fmt.Sprintf("To: %s", tgtName))
	}
	b.WriteString("\n\n")

	b.WriteString(inputStyle.Render("Input"))
	b.WriteString("\n")
	b.WriteString(m.textarea.View())
	b.WriteString("\n\n")

	b.WriteString(outputStyle.Render("Output"))
	b.WriteString("\n")
	if m.output != "" {
		b.WriteString(m.output)
		b.WriteString("\n")
	} else {
		b.WriteString(subtleStyle.Render("Translation will appear here..."))
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", max(m.width-2, 0)))
	b.WriteString("\n")
	b.WriteString(m.status)
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("ctrl+y:copy  ctrl+l:clear  ctrl+t:swap  ctrl+c:quit  tab:next"))

	return b.String()
}
