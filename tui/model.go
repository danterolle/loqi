package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/danterolle/voca/translate"
)

type focusField int

const (
	focusSrcLang focusField = iota
	focusTgtLang
	focusInput
)

const debounceDuration = 600 * time.Millisecond
const langListVisible = 15

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
	ctx       context.Context
	backend   translate.Backend
	langCodes []string
	langNames map[string]string
	srcIdx    int
	tgtIdx    int

	textarea  textarea.Model
	output    string
	lastInput string

	focused      focusField
	status       string
	ready        bool
	width        int
	translateSeq int
	leadingDone  bool
}

func newModel(ctx context.Context, backend translate.Backend, langs translate.LanguageProvider) Model {
	ta := textarea.New()
	ta.Placeholder = "Type text to translate..."
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.Focus()

	list := langs.List()
	codes := make([]string, len(list))
	names := make(map[string]string, len(list))
	srcIdx, tgtIdx := 0, 1
	for i, l := range list {
		codes[i] = l.Code
		names[l.Code] = l.Name
		if l.Code == "auto" {
			srcIdx = i
		}
		if l.Code == "en" {
			tgtIdx = i
		}
	}

	return Model{
		ctx:       ctx,
		backend:   backend,
		langCodes: codes,
		langNames: names,
		srcIdx:    srcIdx,
		tgtIdx:    tgtIdx,
		textarea:  ta,
		focused:   focusInput,
		status:    "Ready. Select languages and start typing.",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
