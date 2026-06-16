package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"

	"github.com/danterolle/voca/translate"
)

type focusField int

const (
	focusSrcLang focusField = iota
	focusTgtLang
	focusInput
)

const debounceDuration = 600 * time.Millisecond

type (
	debounceMsg struct {
		seq int
	}
	translateResultMsg struct {
		text   string
		result string
		err    error
	}
)

type Model struct {
	langCodes []string
	langNames map[string]string
	srcIdx    int
	tgtIdx    int

	textarea textarea.Model
	output   string

	focused      focusField
	status       string
	ready        bool
	width        int
	height       int
	translateSeq int
	leadingDone  bool
}

func InitialModel() Model {
	ta := textarea.New()
	ta.Placeholder = "Type text to translate..."
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.Focus()

	codes := []string{"auto", "en", "it", "fr", "de", "es", "pt", "nl", "pl",
		"ru", "ja", "zh", "ko", "ar", "tr", "cs", "sv", "da", "fi", "el",
		"ro", "hu", "vi", "th", "hi"}

	return Model{
		langCodes: codes,
		langNames: translate.Languages,
		srcIdx:    0,
		tgtIdx:    1,
		textarea:  ta,
		focused:   focusInput,
		status:    "Ready. Select languages and start typing.",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height
		contentH := msg.Height - 6
		if contentH < 4 {
			contentH = 4
		}
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(contentH / 2)

	case tea.KeyMsg:
		return m.handleKey(msg)

	case debounceMsg:
		if msg.seq != m.translateSeq {
			return m, nil
		}
		m.leadingDone = false
		text := m.pendingText()
		if text == "" || m.output == text {
			return m, nil
		}
		src := m.langCodes[m.srcIdx]
		tgt := m.langCodes[m.tgtIdx]
		m.status = fmt.Sprintf("Translating... (%s -> %s)", m.langNames[src], m.langNames[tgt])
		return m, m.doTranslate(text, src, tgt)

	case translateResultMsg:
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

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m Model) pendingText() string {
	return m.textarea.Value()
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.focused == focusSrcLang || m.focused == focusTgtLang {
		switch msg.String() {
		case "tab":
			return m.advanceFocus(), nil
		case "shift+tab":
			return m.retreatFocus(), nil
		case "left":
			if m.focused == focusSrcLang && m.srcIdx > 0 {
				m.srcIdx--
			} else if m.focused == focusTgtLang && m.tgtIdx > 0 {
				m.tgtIdx--
			}
		case "right":
			if m.focused == focusSrcLang && m.srcIdx < len(m.langCodes)-1 {
				m.srcIdx++
			} else if m.focused == focusTgtLang && m.tgtIdx < len(m.langCodes)-1 {
				m.tgtIdx++
			}
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.String() {
	case "tab":
		return m.advanceFocus(), nil
	case "shift+tab":
		return m.retreatFocus(), nil
	case "ctrl+y":
		if m.output != "" {
			if err := copyClipboard(m.output); err == nil {
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
		if m.srcIdx != 0 {
			m.srcIdx, m.tgtIdx = m.tgtIdx, m.srcIdx
			m.status = "Languages swapped."
		}
		return m, nil
	case "ctrl+c", "esc":
		return m, tea.Quit
	}

	var cmd tea.Cmd
	before := m.textarea.Value()
	m.textarea, cmd = m.textarea.Update(msg)
	if m.textarea.Value() != before {
		m.translateSeq++
		seq := m.translateSeq
		cmd = tea.Batch(cmd, tea.Tick(debounceDuration, func(t time.Time) tea.Msg {
			return debounceMsg{seq: seq}
		}))
		if !m.leadingDone {
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

func (m Model) doTranslate(text, source, target string) tea.Cmd {
	return func() tea.Msg {
		result, err := translate.Translate(text, source, target, translate.DefaultModel)
		return translateResultMsg{text: text, result: result, err: err}
	}
}

func copyClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
