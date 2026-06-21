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
	b.WriteString(m.headerView())
	b.WriteString("\n\n")

	if m.focused == focusSrcLang || m.focused == focusTgtLang {
		b.WriteString(m.languageListView())
	} else {
		b.WriteString(inputStyle.Render("Input"))
		b.WriteString("\n")
		b.WriteString(m.textarea.View())
		b.WriteString("\n\n")
		b.WriteString(m.outputView())
	}

	b.WriteString(m.statusView())
	return b.String()
}

func (m Model) headerView() string {
	var b strings.Builder
	srcName := m.langNames[m.langCodes[m.sourceIdx]]
	tgtName := m.langNames[m.langCodes[m.targetIdx]]

	b.WriteString(headerStyle.Render("loqi"))
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
	return b.String()
}

func (m Model) outputView() string {
	var b strings.Builder
	b.WriteString(outputStyle.Render("Output"))
	b.WriteString("\n")
	if m.output != "" {
		b.WriteString(wrap(m.output, m.width-4))
		b.WriteString("\n")
	} else {
		b.WriteString(subtleStyle.Render("Translation will appear here..."))
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) statusView() string {
	var b strings.Builder
	b.WriteString(strings.Repeat("─", max(m.width-2, 0)))
	b.WriteString("\n")
	b.WriteString(m.status)
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("ctrl+y:copy  ctrl+l:clear  ctrl+t:swap  ctrl+c:quit  tab:next"))
	return b.String()
}

func (m Model) languageListView() string {
	var b strings.Builder

	label := "Source"
	idx := m.sourceIdx
	if m.focused == focusTgtLang {
		label = "Target"
		idx = m.targetIdx
	}

	b.WriteString(fmt.Sprintf("  %s language (↑↓ to navigate, Tab to confirm)\n\n", label))

	total := len(m.langCodes)
	start := idx - langListVisible/2
	if start < 0 {
		start = 0
	}
	if start+langListVisible > total {
		start = total - langListVisible
	}
	if start < 0 {
		start = 0
	}
	end := start + langListVisible
	if end > total {
		end = total
	}

	if start > 0 {
		b.WriteString("    ...\n")
	}
	for i := start; i < end; i++ {
		cursor := "  "
		style := subtleStyle.Render
		if i == idx {
			cursor = " >"
			style = inputStyle.Bold(true).Render
		}
		code := m.langCodes[i]
		name := m.langNames[code]
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, style(fmt.Sprintf("%-5s %s", code, name))))
	}
	if end < total {
		b.WriteString("    ...\n")
	}

	return b.String()
}

func wrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var result strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		prefix := line[:len(line)-len(trimmed)]
		prefLen := len([]rune(prefix))

		words := strings.Fields(trimmed)
		if len(words) == 0 {
			result.WriteString(line)
			result.WriteByte('\n')
			continue
		}
		n := prefLen
		result.WriteString(prefix)
		for _, w := range words {
			wLen := len([]rune(w))
			if n > prefLen && n+1+wLen > width {
				result.WriteByte('\n')
				result.WriteString(prefix)
				n = prefLen
			}
			if wLen > width {
				if n > prefLen {
					result.WriteByte('\n')
					result.WriteString(prefix)
					n = prefLen
				}
				for _, r := range w {
					if n >= width {
						result.WriteByte('\n')
						result.WriteString(prefix)
						n = prefLen
					}
					result.WriteRune(r)
					n++
				}
				continue
			}
			if n > prefLen {
				result.WriteByte(' ')
				n++
			}
			result.WriteString(w)
			n += wLen
		}
		result.WriteByte('\n')
	}
	return strings.TrimRight(result.String(), "\n")
}
