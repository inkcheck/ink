package model

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// editorGradeDebounce is the delay before recalculating the FK grade after edits.
const editorGradeDebounce = 500 * time.Millisecond

// editorGradeTickMsg triggers a debounced FK grade recalculation.
type editorGradeTickMsg struct{}

// Editor is the distraction-free markdown editor.
type Editor struct {
	textarea     textarea.Model
	filePath     string
	common       *Common
	saved        bool
	err          error
	savedContent string // content at last save, for unsaved-change detection
	prevContent  string // content at last frame, for change detection
	grade        string // cached FK grade
	gradeDirty   bool   // true when grade needs recalculation
	zenMode      bool // true hides all chrome (Alt+Z)
	showHelp     bool // true shows help pane at the bottom
	confirmClose bool // true when waiting for second esc/ctrl+w to discard unsaved changes
}

// NewEditor creates a new Editor for the given file content.
func NewEditor(common *Common, filePath string, content string) Editor {
	ta := textarea.New()
	ta.SetValue(content)
	ta.ShowLineNumbers = true
	ta.SetWidth(common.ContentWidth())
	ta.SetHeight(editorTextareaHeight(common, false))
	ta.Focus()

	// Move cursor to the beginning of the file
	for ta.Line() > 0 {
		ta.CursorUp()
	}
	ta.CursorStart()

	// Reclaim ctrl+f, ctrl+b, ctrl+t from default bindings
	ta.KeyMap.CharacterForward = key.NewBinding(key.WithKeys("right"))
	ta.KeyMap.CharacterBackward = key.NewBinding(key.WithKeys("left"))
	ta.KeyMap.TransposeCharacterBackward = key.NewBinding(key.WithKeys(""))
	ta.KeyMap.DeleteWordBackward = key.NewBinding(key.WithKeys("alt+backspace"))

	// Custom navigation shortcuts
	ta.KeyMap.InputBegin = key.NewBinding(key.WithKeys("alt+<", "ctrl+home", "ctrl+t"))
	ta.KeyMap.InputEnd = key.NewBinding(key.WithKeys("alt+>", "ctrl+end", "ctrl+g"))

	dim := lipgloss.Color("240")
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(dim)
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(dim)
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(dim)

	return Editor{
		textarea:     ta,
		filePath:     filePath,
		common:       common,
		saved:        true,
		savedContent: content,
		prevContent:  content,
		grade:        fleschKincaidGrade(content),
	}
}

func (e Editor) Init() tea.Cmd {
	return textarea.Blink
}

func (e *Editor) Reload() {
	raw, err := os.ReadFile(e.filePath)
	if err != nil {
		e.err = err
		return
	}
	row := e.textarea.Line()
	col := e.textarea.LineInfo().CharOffset

	content := string(raw)
	e.textarea.SetValue(content)
	e.savedContent = content
	e.prevContent = content
	e.saved = true
	e.err = nil
	e.grade = fleschKincaidGrade(content)
	e.gradeDirty = false

	// Restore cursor position: reset to beginning first (SetValue leaves the
	// cursor at the end and the viewport in a state where CursorUp alone
	// cannot reliably reach line 0), then navigate down to the target row.
	lineCount := e.textarea.LineCount()
	if row >= lineCount {
		row = lineCount - 1
	}
	for e.textarea.Line() > 0 {
		e.textarea.CursorUp()
	}
	e.textarea.CursorStart()
	for i := 0; i < row; i++ {
		e.textarea.CursorDown()
	}
	e.textarea.SetCursor(col)
}

func (e Editor) Update(msg tea.Msg) (Editor, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.textarea.SetWidth(e.common.ContentWidth())
		e.textarea.SetHeight(editorTextareaHeight(e.common, e.showHelp))
	case editorGradeTickMsg:
		if e.gradeDirty {
			e.grade = fleschKincaidGrade(e.textarea.Value())
			e.gradeDirty = false
		}
		return e, nil
	case tea.KeyMsg:
		k := msg.String()
		// Reset close confirmation on any key that isn't esc/ctrl+w
		if k != "esc" && k != "ctrl+w" {
			e.confirmClose = false
		}
		switch k {
		case "ctrl+s":
			content := e.textarea.Value()
			err := os.WriteFile(e.filePath, []byte(content), 0644)
			if err != nil {
				e.err = err
				return e, nil
			}
			e.saved = true
			e.err = nil
			e.savedContent = content
			return e, func() tea.Msg {
				return FileSavedMsg{}
			}
		case "ctrl+f":
			// Loop-based scroll: the textarea widget does not expose a half-page
			// scroll method, so we move the cursor one line at a time.
			for i := 0; i < e.textarea.Height()/2; i++ {
				e.textarea.CursorDown()
			}
			return e, nil
		case "ctrl+b":
			// See ctrl+f comment above.
			for i := 0; i < e.textarea.Height()/2; i++ {
				e.textarea.CursorUp()
			}
			return e, nil
		case "ctrl+r":
			e.Reload()
			return e, nil
		case "alt+?", "alt+/":
			e.showHelp = !e.showHelp
			e.textarea.SetHeight(editorTextareaHeight(e.common, e.showHelp))
			return e, nil
		case "alt+z":
			e.zenMode = !e.zenMode
			if e.zenMode {
				e.textarea.ShowLineNumbers = false
				e.textarea.SetPromptFunc(editorGutterWidth, func(lineIdx int) string {
					return strings.Repeat(" ", editorGutterWidth)
				})
				e.textarea.SetWidth(e.common.ContentWidth())
			} else {
				e.textarea.ShowLineNumbers = true
				e.textarea.SetPromptFunc(0, nil)
				e.textarea.Prompt = lipgloss.ThickBorder().Left + " "
				dim := lipgloss.Color("240")
				e.textarea.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(dim)
				e.textarea.SetWidth(e.common.ContentWidth())
			}
			return e, nil
		case "esc", "ctrl+w":
			if !e.saved && !e.confirmClose {
				e.confirmClose = true
				return e, nil
			}
			return e, func() tea.Msg {
				return CloseEditorMsg{
					FilePath: e.filePath,
				}
			}
		}
	}

	var cmd tea.Cmd
	e.textarea, cmd = e.textarea.Update(msg)

	// Detect content changes for unsaved-state and debounced grade
	content := e.textarea.Value()
	if content != e.prevContent {
		if content != e.savedContent {
			e.saved = false
		} else {
			e.saved = true
		}
		e.gradeDirty = true
		e.prevContent = content
		gradeCmd := tea.Tick(editorGradeDebounce, func(time.Time) tea.Msg {
			return editorGradeTickMsg{}
		})
		return e, tea.Batch(cmd, gradeCmd)
	}

	return e, cmd
}

func (e Editor) statusBarView() string {
	w := e.common.Width

	left := statusBarBookName(e.common.BookName) + statusBarFileName(e.filePath)

	// Word count + grade + status + hints
	words := countWords(e.textarea.Value())
	wordCount := fmt.Sprintf("%d words", words)

	parts := []string{wordCount}
	if e.grade != "" {
		parts = append(parts, e.grade)
	}
	if e.confirmClose {
		warnStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Background(lipgloss.Color("236"))
		parts = append(parts, warnStyle.Render("Unsaved! Press again to close"))
	} else if e.err != nil {
		errStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("236"))
		parts = append(parts, errStyle.Render(e.err.Error()))
	} else if e.saved {
		parts = append(parts, statusBarSavedStyle.Render("Saved"))
	}
	parts = append(parts, "⌥? help")

	right := statusBarHintStyle.Render(strings.Join(parts, " | "))

	return statusBarFill(left, right, w)
}

const editorHelpHeight = 3

// editorGutterWidth is the width of the line number gutter (4 digits + 2 prompt chars).
const editorGutterWidth = 6

func editorTextareaHeight(common *Common, showHelp bool) int {
	h := common.Height - editorChromeHeight
	if showHelp {
		h -= editorHelpHeight
	}
	if h < 1 {
		h = 1
	}
	return h
}

func (e Editor) helpView() string {
	return renderHelpPane([][]helpEntry{
		{{"^F", "½ page down"}, {"^B", "½ page up"}, {"^T", "go to top"}},
		{{"^G", "go to end"}, {"^S", "save"}, {"^R", "reload"}},
		{{"^W", "close"}, {"⌥Z", "zen mode"}, {"⌥?", "toggle help"}},
	}, e.common.Width)
}

func (e Editor) View() string {
	var logoStr, statusBar string
	if !e.zenMode {
		logoStr = logo
		statusBar = e.statusBarView()
	}
	content := centerContent(e.textarea.View(), e.common.Width, e.common.MaxWidth)
	var helpPane string
	if e.showHelp {
		helpPane = e.helpView()
	}
	return layoutView(logoStr, content, statusBar, helpPane)
}
