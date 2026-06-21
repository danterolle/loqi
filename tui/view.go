package tui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
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
		b.WriteString(inputBoxStyle.Render(m.inputContentView()))
		b.WriteString("\n\n")
		b.WriteString(outputBoxStyle.Render(m.outputContentView()))
	}

	b.WriteString("\n")
	b.WriteString(m.statusView())
	b.WriteString("\n")
	b.WriteString(m.helpView())
	return b.String()
}

func (m Model) inputContentView() string {
	return inputLabelStyle.Render("Input") + "\n" + m.textarea.View()
}

func (m Model) outputContentView() string {
	if m.output != "" {
		return outputLabelStyle.Render("Output") + "\n" + wrap(m.output, m.width-4)
	}
	return outputLabelStyle.Render("Output") + "\n" + subtleStyle.Render("Translation will appear here...")
}

func (m Model) headerView() string {
	var b strings.Builder
	srcName := m.langNames[m.langCodes[m.sourceIdx]]
	tgtName := m.langNames[m.langCodes[m.targetIdx]]

	title := "loqi"
	if m.version != "" {
		title += " " + m.version
		if m.commit != "" {
			title += " " + m.commit[:min(7, len(m.commit))]
		}
	}
	b.WriteString(headerStyle.Render(title))
	b.WriteString("  ")
	if m.focused == focusSrcLang {
		b.WriteString(subtleStyle.Render("From:"))
		b.WriteString(" ")
		b.WriteString(inputLabelStyle.Bold(true).Render(srcName))
	} else {
		fmt.Fprintf(&b, "From: %s", srcName)
	}
	b.WriteString("  ->  ")
	if m.focused == focusTgtLang {
		b.WriteString(subtleStyle.Render("To:"))
		b.WriteString(" ")
		b.WriteString(outputLabelStyle.Bold(true).Render(tgtName))
	} else {
		fmt.Fprintf(&b, "To: %s", tgtName)
	}
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

	fmt.Fprintf(&b, "  %s language (↑↓ to navigate, Tab to confirm)\n\n", label)

	total := len(m.langCodes)
	start := max(0, min(idx-langListVisible/2, total-langListVisible))
	end := min(start+langListVisible, total)

	if start > 0 {
		b.WriteString("    ...\n")
	}
	for i := start; i < end; i++ {
		cursor := "  "
		style := subtleStyle.Render
		if i == idx {
			cursor = " >"
			style = inputLabelStyle.Bold(true).Render
		}
		code := m.langCodes[i]
		name := m.langNames[code]
		fmt.Fprintf(&b, "%s %s\n", cursor, style(fmt.Sprintf("%-5s %s", code, name)))
	}
	if end < total {
		b.WriteString("    ...\n")
	}

	return b.String()
}

func (m Model) statusView() string {
	return strings.Repeat("─", max(m.width-2, 0)) + "\n" + m.status
}

func (m Model) helpView() string {
	return helpStyle.Render("ctrl+y:copy  ctrl+l:clear  ctrl+t:swap  tab:next  ctrl+c:quit")
}

func wrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var result strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		prefix := line[:len(line)-len(trimmed)]
		prefLen := runewidth.StringWidth(prefix)

		words := strings.Fields(trimmed)
		if len(words) == 0 {
			result.WriteString(line)
			result.WriteByte('\n')
			continue
		}
		n := prefLen
		result.WriteString(prefix)
		for _, w := range words {
			wLen := runewidth.StringWidth(w)
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
					rw := runewidth.RuneWidth(r)
					if n+rw > width {
						result.WriteByte('\n')
						result.WriteString(prefix)
						n = prefLen
					}
					result.WriteRune(r)
					n += rw
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
