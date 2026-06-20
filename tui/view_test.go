package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danterolle/loqi/translate"
)

func newTestModel(t *testing.T) Model {
	t.Helper()
	m := newModel(context.Background(), translate.NewMockBackend(), translate.NewStaticLanguages())
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return mm.(Model)
}

func TestViewShowsTranslationResult(t *testing.T) {
	m := newTestModel(t)
	m.textarea.SetValue("hello")

	mm, _ := m.Update(translateResultMsg{text: "hello", result: "ciao"})
	m = mm.(Model)

	view := m.View()
	if !strings.Contains(view, "ciao") {
		t.Fatal("view should contain the translated text")
	}
}

func TestViewStaleResultDoesNotOverwrite(t *testing.T) {
	m := newTestModel(t)
	m.textarea.SetValue("newest")
	m.output = "newest result"

	mm, _ := m.Update(translateResultMsg{text: "stale text", result: "should not appear"})
	m = mm.(Model)

	view := m.View()
	if strings.Contains(view, "should not appear") {
		t.Fatal("stale result should not overwrite newer output")
	}
	if !strings.Contains(view, "newest result") {
		t.Fatal("original output should be preserved")
	}
}

func TestViewEmptyResultDoesNotClearOutput(t *testing.T) {
	m := newTestModel(t)
	m.textarea.SetValue("hello")
	m.output = "existing"

	mm, _ := m.Update(translateResultMsg{text: "hello", result: ""})
	m = mm.(Model)

	view := m.View()
	if !strings.Contains(view, "existing") {
		t.Fatal("empty result should not clear existing output")
	}
}

func TestViewErrorPreservesOutputAndShowsError(t *testing.T) {
	m := newTestModel(t)
	m.textarea.SetValue("hello")
	m.output = "existing"

	mm, _ := m.Update(translateResultMsg{text: "hello", err: errors.New("mock error")})
	m = mm.(Model)

	view := m.View()
	if !strings.Contains(view, "existing") {
		t.Fatal("error should preserve existing output")
	}
	if !strings.Contains(view, "Error:") {
		t.Fatal("error should be visible in status")
	}
}

func TestViewClearResetsOutputAndInput(t *testing.T) {
	m := newTestModel(t)
	m.textarea.SetValue("hello")
	m.output = "ciao"
	m.leadingInProgress = true

	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	m = mm.(Model)

	view := m.View()
	if strings.Contains(view, "ciao") {
		t.Fatal("output should be cleared after Ctrl+L")
	}
	if !strings.Contains(view, "Cleared.") {
		t.Fatal("status should show Cleared after Ctrl+L")
	}
}

func TestViewLanguageListAppearsOnTab(t *testing.T) {
	m := newTestModel(t)

	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mm.(Model)

	view := m.View()
	if !strings.Contains(view, "Source language") {
		t.Fatal("view should show language list when focus is on language")
	}
	if !strings.Contains(view, "auto") {
		t.Fatal("language list should contain language codes")
	}
}
