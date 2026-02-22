package model

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the root application model that routes between views.
type Model struct {
	common     *Common
	view       ViewState
	book       Book
	chapter    Chapter
	editor     Editor
}

// New creates the root model.
func New(dir string, maxWidth int) Model {
	if maxWidth < MinWidth {
		maxWidth = MinWidth
	}
	common := &Common{Width: 80, Height: 24, MaxWidth: maxWidth, IsBook: true}
	book := NewBook(common, dir)
	common.BookName = book.bookName

	return Model{
		common: common,
		view:   BookView,
		book:   book,
	}
}

// NewFromFile creates a model that opens a single markdown file directly in ChapterView.
// Pressing back/esc quits the app instead of returning to BookView.
func NewFromFile(filePath string, maxWidth int) Model {
	if maxWidth < MinWidth {
		maxWidth = MinWidth
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	common := &Common{Width: 80, Height: 24, MaxWidth: maxWidth}
	common.BookName = filepath.Base(absPath)
	chapter := NewChapter(common, absPath)

	return Model{
		common:  common,
		view:    ChapterView,
		chapter: chapter,
	}
}

// NewFromFiles creates a model that shows a filtered BookView with the given file/dir paths.
func NewFromFiles(files []string, maxWidth int) Model {
	if maxWidth < MinWidth {
		maxWidth = MinWidth
	}
	common := &Common{Width: 80, Height: 24, MaxWidth: maxWidth, IsBook: true}
	book := NewBookFromFiles(common, files)
	common.BookName = book.bookName

	return Model{
		common: common,
		view:   BookView,
		book:   book,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width
		if w < MinWidth {
			w = MinWidth
		}
		m.common.Width = w
		m.common.Height = msg.Height

		// Forward resize to all initialized views so they don't have stale dimensions.
		if m.book.common != nil {
			m.book, _ = m.book.Update(msg)
		}
		if m.chapter.common != nil {
			m.chapter, _ = m.chapter.Update(msg)
		}
		if m.editor.common != nil {
			m.editor, _ = m.editor.Update(msg)
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case OpenChapterMsg:
		m.chapter = NewChapter(m.common, msg.FilePath)
		m.view = ChapterView
		return m, nil

	case OpenExternalEditorMsg:
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		parts := strings.Fields(editor)
		c := exec.Command(parts[0], append(parts[1:], msg.FilePath)...)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return ExternalEditorClosedMsg{Err: err}
		})

	case ExternalEditorClosedMsg:
		// Route through chapter's Update so it manages its own state
		var cmd tea.Cmd
		m.chapter, cmd = m.chapter.Update(msg)
		m.view = ChapterView
		return m, cmd

	case OpenEditorMsg:
		m.editor = NewEditor(m.common, msg.FilePath, msg.Content)
		m.view = EditorView
		return m, m.editor.Init()

	case CloseEditorMsg:
		// Refresh chapter content after editing
		m.chapter.refresh()
		m.view = ChapterView
		return m, nil

	case FileSavedMsg:
		// File saved, stay in editor
		return m, nil

	case BackToBookMsg:
		if !m.common.IsBook {
			return m, tea.Quit
		}
		m.view = BookView
		return m, nil
	}

	// Route to active view
	var cmd tea.Cmd
	switch m.view {
	case BookView:
		m.book, cmd = m.book.Update(msg)
	case ChapterView:
		m.chapter, cmd = m.chapter.Update(msg)
	case EditorView:
		m.editor, cmd = m.editor.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.view {
	case ChapterView:
		return m.chapter.View()
	case EditorView:
		return m.editor.View()
	default:
		return m.book.View()
	}
}
