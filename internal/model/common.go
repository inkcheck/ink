package model

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inkcheck/readability"
)

// ViewState represents which view is currently active.
type ViewState int

const (
	BookView ViewState = iota
	ChapterView
	EditorView
)

// MinWidth is the minimum usable width for the application.
const MinWidth = 60

// Layout constants for chrome height calculations.
const (
	// logoAndGapHeight accounts for the logo line plus the blank line below it.
	logoAndGapHeight = 2
	// statusBarHeight accounts for the status bar at the bottom.
	statusBarHeight = 1
	// bookChromeHeight is the total chrome for the book view (logo + gap + title + filter + gap + status).
	bookChromeHeight = 6
	// chapterChromeHeight is the total chrome for the chapter view (logo + gap + status).
	chapterChromeHeight = 3
	// editorChromeHeight is the total chrome for the editor view (logo + gap + status).
	editorChromeHeight = 3
)

// Common holds shared state across all views.
type Common struct {
	Width      int
	Height     int
	MaxWidth   int
	BookName   string
	IsBook     bool // true when there is a book view to return to
}

// ContentWidth returns the effective content width, capped at MaxWidth.
func (c *Common) ContentWidth() int { return min(c.Width, c.MaxWidth) }

// Inter-view messages

// OpenChapterMsg requests switching to the Chapter view for the given file.
type OpenChapterMsg struct {
	FilePath string
}

// OpenEditorMsg requests switching to the Editor view.
type OpenEditorMsg struct {
	FilePath string
	Content  string
}

// CloseEditorMsg signals the editor has closed.
type CloseEditorMsg struct {
	FilePath string
}

// OpenExternalEditorMsg requests opening the file in $EDITOR.
type OpenExternalEditorMsg struct {
	FilePath string
}

// ExternalEditorClosedMsg signals the external editor has exited.
type ExternalEditorClosedMsg struct {
	Err error
}

// BackToBookMsg signals returning to the Book view.
type BackToBookMsg struct{}

// FileSavedMsg signals a file was saved successfully.
type FileSavedMsg struct{}

// Shared status bar styles.
var (
	statusBarBookStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	statusBarNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	statusBarHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	statusBarSavedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("120")).
				Background(lipgloss.Color("236"))

	statusBarStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("120")).
				Background(lipgloss.Color("236"))
)

// statusBarBookName renders the book name segment for a status bar.
func statusBarBookName(bookName string) string {
	return statusBarBookStyle.Render(bookName)
}

// statusBarFileName renders the filename segment for a status bar.
func statusBarFileName(filePath string) string {
	return statusBarNameStyle.Render(filepath.Base(filePath))
}

// logo is the pre-rendered application logo.
var logo = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("230")).
	Background(lipgloss.Color("135")).
	Padding(0, 1).
	Render("Ink")

// statusBarFillStyle is the pre-computed fill style for status bars.
var statusBarFillStyle = lipgloss.NewStyle().Background(lipgloss.Color("236"))

// statusBarFill builds a status bar row: left + fill + right, padded to width.
func statusBarFill(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	fill := statusBarFillStyle.Render(strings.Repeat(" ", gap))
	return left + fill + right
}

// fleschKincaidGrade returns a formatted grade string for the given text.
func fleschKincaidGrade(text string) string {
	a := readability.NewAnalysis(text)
	score, err := a.Score(readability.FleschKincaidGrade)
	if err != nil || a.Stats().Words < 10 {
		return ""
	}
	return fmt.Sprintf("Grade %d", int(score))
}

// helpEntry is a key-description pair for help panes.
type helpEntry struct{ Key, Val string }

// renderHelpPane renders a centered, multi-column help pane from the given columns.
func renderHelpPane(cols [][]helpEntry, width int) string {
	bg := lipgloss.AdaptiveColor{Light: "#f2f2f2", Dark: "#1B1B1B"}
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Background(bg)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Background(bg)
	bgStyle := lipgloss.NewStyle().Background(bg)

	maxKeyW := 0
	maxValW := 0
	for _, col := range cols {
		for _, e := range col {
			if w := lipgloss.Width(e.Key); w > maxKeyW {
				maxKeyW = w
			}
			if w := lipgloss.Width(e.Val); w > maxValW {
				maxValW = w
			}
		}
	}
	gap := 1
	colWidth := maxKeyW + gap + maxValW
	colGap := 2

	maxRows := 0
	for _, col := range cols {
		if len(col) > maxRows {
			maxRows = len(col)
		}
	}

	pad := func(s string, w int) string {
		n := w - lipgloss.Width(s)
		if n > 0 {
			return s + bgStyle.Render(strings.Repeat(" ", n))
		}
		return s
	}

	numCols := len(cols)
	totalContentW := numCols*colWidth + (numCols-1)*colGap
	indent := (width - totalContentW) / 2
	if indent < 2 {
		indent = 2
	}

	var lines []string
	for r := 0; r < maxRows; r++ {
		line := bgStyle.Render(strings.Repeat(" ", indent))
		for ci, col := range cols {
			if r < len(col) {
				cell := pad(keyStyle.Render(col[r].Key), maxKeyW) +
					bgStyle.Render(strings.Repeat(" ", gap)) +
					valStyle.Render(col[r].Val)
				line += pad(cell, colWidth)
			} else {
				line += bgStyle.Render(strings.Repeat(" ", colWidth))
			}
			if ci < numCols-1 {
				line += bgStyle.Render(strings.Repeat(" ", colGap))
			}
		}
		lineWidth := lipgloss.Width(line)
		if lineWidth < width {
			line += bgStyle.Render(strings.Repeat(" ", width-lineWidth))
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// centerContent horizontally centers content when the terminal is wider than maxWidth.
func centerContent(content string, termWidth, maxWidth int) string {
	if termWidth <= maxWidth {
		return content
	}
	block := lipgloss.NewStyle().Width(maxWidth).Render(content)
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, block)
}

// countWords counts words in s by iterating runes and counting space-to-non-space transitions.
func countWords(s string) int {
	count := 0
	inWord := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

// clearStatusAfter returns a tea.Cmd that sends msg after duration d.
func clearStatusAfter(d time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return msg })
}

// layoutView assembles the standard view layout: logo, content, status bar, and optional help pane.
func layoutView(logoStr, content, statusBar, helpPane string) string {
	var b strings.Builder
	b.WriteString(logoStr)
	b.WriteString("\n\n")
	b.WriteString(content)
	b.WriteString("\n")
	b.WriteString(statusBar)
	if helpPane != "" {
		b.WriteString("\n")
		b.WriteString(helpPane)
	}
	return b.String()
}

// pluralize returns singular when n == 1, plural otherwise.
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// Duration constants for relativeTime calculations.
const (
	day   = 24 * time.Hour
	month = 30 * day
	year  = 365 * day
)

// relativeTime formats t as a human-readable duration relative to now.
func relativeTime(t time.Time, now time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := max(1, int(math.Round(d.Minutes())))
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < day:
		h := int(math.Round(d.Hours()))
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < month:
		days := int(math.Round(d.Hours() / 24))
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case d < year:
		months := int(math.Round(d.Hours() / 24 / 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(math.Round(d.Hours() / 24 / 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// currentUser returns the current username, falling back across platform env vars.
func currentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return ""
}
