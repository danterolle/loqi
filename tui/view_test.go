package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/danterolle/loqi/translate"
)

func newTestModel(t *testing.T) Model {
	t.Helper()
	m := newModel(context.Background(), translate.NewMockBackend(), translate.NewStaticLanguages(), "", "")
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

func TestWrap_CJK_no_spaces(t *testing.T) {
	input := "你好世界你好世界"
	width := 10
	got := wrap(input, width)
	for _, line := range strings.Split(got, "\n") {
		if runewidth.StringWidth(line) > width {
			t.Fatalf("line %q has display width %d, exceeds %d", line, runewidth.StringWidth(line), width)
		}
	}
	if runewidth.StringWidth(got) > runewidth.StringWidth(input) {
		t.Fatal("wrapped output should not exceed input display width")
	}
}

func TestWrap_CJK_with_spaces(t *testing.T) {
	input := "你好 世界 你好 世界"
	width := 8
	got := wrap(input, width)
	for _, line := range strings.Split(got, "\n") {
		if runewidth.StringWidth(line) > width {
			t.Fatalf("line %q has display width %d, exceeds %d", line, runewidth.StringWidth(line), width)
		}
	}
}

func TestWrap_ascii(t *testing.T) {
	input := "hello world foo bar"
	width := 10
	got := wrap(input, width)
	for _, line := range strings.Split(got, "\n") {
		if len(line) > width {
			t.Fatalf("line %q exceeds width %d", line, width)
		}
	}
}

func TestWrap_empty(t *testing.T) {
	if got := wrap("", 10); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestWrap_zero_width(t *testing.T) {
	input := "hello"
	if got := wrap(input, 0); got != input {
		t.Fatalf("expected unchanged, got %q", got)
	}
}
