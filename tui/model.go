package tui

import (
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
