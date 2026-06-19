package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)

	case debounceMsg:
		return m.handleDebounce(msg)

	case translateResultMsg:
		return m.handleTranslateResult(msg)
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.ready = true
	m.width = msg.Width
	contentH := msg.Height - 6
	if contentH < 4 {
		contentH = 4
	}
	m.textarea.SetWidth(msg.Width - 4)
	m.textarea.SetHeight(contentH / 2)
	return m, nil
}

func (m Model) handleDebounce(msg debounceMsg) (Model, tea.Cmd) {
	if msg.seq != m.translateSeq {
		return m, nil
	}
	m.leadingDone = false
	text := m.textarea.Value()
	if text == "" || m.output == text {
		return m, nil
	}
	src := m.langCodes[m.srcIdx]
	tgt := m.langCodes[m.tgtIdx]
	m.status = fmt.Sprintf("Translating... (%s -> %s)", m.langNames[src], m.langNames[tgt])
	return m, m.doTranslate(text, src, tgt)
}

func (m Model) handleTranslateResult(msg translateResultMsg) (Model, tea.Cmd) {
	if msg.text != m.textarea.Value() {
		return m, nil
	}
	if msg.err != nil {
		m.status = fmt.Sprintf("Error: %v", msg.err)
	} else {
		if msg.result != "" {
			m.output = msg.result
		}
		m.status = "Ready."
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.focused == focusSrcLang || m.focused == focusTgtLang {
		return m.handleLangKey(msg)
	}
	return m.handleInputKey(msg)
}

func (m Model) handleLangKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		return m.advanceFocus(), nil
	case "shift+tab":
		return m.retreatFocus(), nil
	case "up":
		return m.adjustLangIndex(-1), nil
	case "down":
		return m.adjustLangIndex(1), nil
	case "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) adjustLangIndex(delta int) Model {
	switch m.focused {
	case focusSrcLang:
		idx := m.srcIdx + delta
		if idx >= 0 && idx < len(m.langCodes) {
			m.srcIdx = idx
		}
	case focusTgtLang:
		idx := m.tgtIdx + delta
		if idx >= 0 && idx < len(m.langCodes) {
			m.tgtIdx = idx
		}
	}
	return m
}

func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		return m.advanceFocus(), nil
	case "shift+tab":
		return m.retreatFocus(), nil
	case "ctrl+y":
		if m.output != "" {
			if err := copyClipboard(m.output); err != nil {
				m.status = fmt.Sprintf("Clipboard error: %v", err)
			} else {
				m.status = "Copied to clipboard!"
			}
		}
		return m, nil
	case "ctrl+l":
		m.textarea.Reset()
		m.output = ""
		m.translateSeq++
		m.leadingDone = false
		m.status = "Cleared."
		return m, nil
	case "ctrl+t":
		if m.langCodes[m.srcIdx] != "auto" {
			m.srcIdx, m.tgtIdx = m.tgtIdx, m.srcIdx
			m.status = "Languages swapped."
		}
		return m, nil
	case "ctrl+c", "esc":
		return m, tea.Quit
	}

	return m.handleTextChange(msg)
}

func (m Model) handleTextChange(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	before := m.textarea.Value()
	m.textarea, cmd = m.textarea.Update(msg)
	if m.textarea.Value() != before {
		m.translateSeq++
		if m.leadingDone {
			seq := m.translateSeq
			cmd = tea.Batch(cmd, tea.Tick(debounceDuration, func(t time.Time) tea.Msg {
				return debounceMsg{seq: seq}
			}))
		} else {
			m.leadingDone = true
			text := m.textarea.Value()
			src := m.langCodes[m.srcIdx]
			tgt := m.langCodes[m.tgtIdx]
			m.status = fmt.Sprintf("Translating... (%s -> %s)", m.langNames[src], m.langNames[tgt])
			cmd = tea.Batch(cmd, m.doTranslate(text, src, tgt))
		}
	}
	return m, cmd
}

func (m Model) advanceFocus() Model {
	m.textarea.Blur()
	switch m.focused {
	case focusSrcLang:
		m.focused = focusTgtLang
	case focusTgtLang:
		m.focused = focusInput
		m.textarea.Focus()
	case focusInput:
		m.focused = focusSrcLang
	}
	return m
}

func (m Model) retreatFocus() Model {
	m.textarea.Blur()
	switch m.focused {
	case focusSrcLang:
		m.focused = focusInput
		m.textarea.Focus()
	case focusTgtLang:
		m.focused = focusSrcLang
	case focusInput:
		m.focused = focusTgtLang
	}
	return m
}
