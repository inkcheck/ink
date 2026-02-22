package model

import (
	"fmt"
	"math"
	"os"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
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

// ViewContext holds shared state across all views.
type ViewContext struct {
	width    int
	height   int
	maxWidth int
	bookName string
	isBook   bool // true when there is a book view to return to
}

// newViewContext creates a ViewContext with maxWidth clamped to MinWidth.
func newViewContext(maxWidth int, isBook bool) *ViewContext {
	return &ViewContext{
		width:    80,
		height:   24,
		maxWidth: max(maxWidth, MinWidth),
		isBook:   isBook,
	}
}

// contentWidth returns the effective content width, capped at maxWidth.
func (c *ViewContext) contentWidth() int { return min(c.width, c.maxWidth) }

// fleschKincaidGrade returns a formatted grade string for the given text.
func fleschKincaidGrade(text string) string {
	a := readability.NewAnalysis(text)
	score, err := a.Score(readability.FleschKincaidGrade)
	if err != nil || a.Stats().Words < 10 {
		return ""
	}
	return fmt.Sprintf("Grade %d", int(score))
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
		return fmt.Sprintf("%d %s ago", m, pluralize(m, "minute", "minutes"))
	case d < day:
		h := int(math.Round(d.Hours()))
		return fmt.Sprintf("%d %s ago", h, pluralize(h, "hour", "hours"))
	case d < month:
		days := int(math.Round(d.Hours() / 24))
		return fmt.Sprintf("%d %s ago", days, pluralize(days, "day", "days"))
	case d < year:
		months := int(math.Round(d.Hours() / 24 / 30))
		return fmt.Sprintf("%d %s ago", months, pluralize(months, "month", "months"))
	default:
		years := int(math.Round(d.Hours() / 24 / 365))
		return fmt.Sprintf("%d %s ago", years, pluralize(years, "year", "years"))
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
