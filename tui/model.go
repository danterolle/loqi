package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/danterolle/loqi/translate"
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
	sourceIdx int
	targetIdx int

	textarea  textarea.Model
	output    string
	lastInput string

	focused           focusField
	status            string
	ready             bool
	width             int
	translateSequence int
	leadingInProgress bool
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
	sourceIdx, targetIdx := 0, len(list)-1
	if targetIdx < 0 {
		targetIdx = 0
	}
	for i, l := range list {
		codes[i] = l.Code
		names[l.Code] = l.Name
		if l.Code == "auto" {
			sourceIdx = i
		}
		if l.Code == "en" {
			targetIdx = i
		}
	}

	return Model{
		ctx:       ctx,
		backend:   backend,
		langCodes: codes,
		langNames: names,
		sourceIdx: sourceIdx,
		targetIdx: targetIdx,
		textarea:  ta,
		focused:   focusInput,
		status:    "Ready. Select languages and start typing.",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
