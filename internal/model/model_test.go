package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func tempDirWithFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestViewRoutingBookView(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{"test.md": "# Hello"})
	m := New(dir, 80)
	view := m.View()
	// Book view should contain the book name (derived from directory)
	bookName := dirToBookName(filepath.Base(dir))
	if !strings.Contains(view, bookName) {
		t.Errorf("BookView: View() missing book name %q", bookName)
	}
}

func TestViewRoutingChapterView(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"readme.md": "# Readme\n\nContent here.",
	})
	m := NewFromFile(filepath.Join(dir, "readme.md"), 80)
	view := m.View()
	// Chapter view should show the rendered markdown content
	if !strings.Contains(view, "Readme") {
		t.Errorf("ChapterView: View() missing chapter content")
	}
}

func TestWindowSizeMsgRespectsMinWidth(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{"test.md": "# Hello"})
	m := New(dir, 80)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 20, Height: 24})
	um := updated.(Model)
	if um.common.Width < MinWidth {
		t.Errorf("WindowSizeMsg: Width = %d, want >= %d", um.common.Width, MinWidth)
	}
}

func TestOpenChapterMsgSwitchesToChapterView(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"chapter.md": "# Chapter\n\nText content.",
	})
	m := New(dir, 80)
	updated, _ := m.Update(OpenChapterMsg{FilePath: filepath.Join(dir, "chapter.md")})
	um := updated.(Model)
	if um.view != ChapterView {
		t.Errorf("OpenChapterMsg: view = %v, want ChapterView", um.view)
	}
}

func TestBackToBookMsgReturnsToBookView(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"chapter.md": "# Chapter\n\nText here.",
	})
	m := New(dir, 80)
	// First go to chapter
	updated, _ := m.Update(OpenChapterMsg{FilePath: filepath.Join(dir, "chapter.md")})
	um := updated.(Model)
	// Then back to book
	updated2, _ := um.Update(BackToBookMsg{})
	um2 := updated2.(Model)
	if um2.view != BookView {
		t.Errorf("BackToBookMsg: view = %v, want BookView", um2.view)
	}
}

func TestBackToBookMsgQuitsWhenNoBook(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"single.md": "# Single\n\nSolo file content.",
	})
	m := NewFromFile(filepath.Join(dir, "single.md"), 80)
	_, cmd := m.Update(BackToBookMsg{})
	if cmd == nil {
		t.Fatal("BackToBookMsg (no book): expected non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("BackToBookMsg (no book): expected tea.QuitMsg, got %T", msg)
	}
}

func TestOpenEditorMsgSwitchesToEditorView(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"edit.md": "# Edit\n\nEditable content.",
	})
	m := New(dir, 80)
	updated, _ := m.Update(OpenEditorMsg{
		FilePath: filepath.Join(dir, "edit.md"),
		Content:  "# Edit\n\nEditable content.",
	})
	um := updated.(Model)
	if um.view != EditorView {
		t.Errorf("OpenEditorMsg: view = %v, want EditorView", um.view)
	}
}

func TestCloseEditorMsgReturnsToChapterView(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"edit.md": "# Edit\n\nContent for editing.",
	})
	m := New(dir, 80)
	// Go to chapter first
	updated, _ := m.Update(OpenChapterMsg{FilePath: filepath.Join(dir, "edit.md")})
	um := updated.(Model)
	// Open editor
	updated2, _ := um.Update(OpenEditorMsg{
		FilePath: filepath.Join(dir, "edit.md"),
		Content:  "# Edit\n\nContent for editing.",
	})
	um2 := updated2.(Model)
	// Close editor
	updated3, _ := um2.Update(CloseEditorMsg{FilePath: filepath.Join(dir, "edit.md")})
	um3 := updated3.(Model)
	if um3.view != ChapterView {
		t.Errorf("CloseEditorMsg: view = %v, want ChapterView", um3.view)
	}
}
