package model

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

	statusBarPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	statusBarInputStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Padding(0, 1)
)

// statusBarBookName renders the book name segment for a status bar.
func statusBarBookName(name string) string {
	return statusBarBookStyle.Render(name)
}

// statusBarFileName renders the filename segment for a status bar.
func statusBarFileName(filePath string) string {
	return statusBarNameStyle.Render(filepath.Base(filePath))
}

// statusBarFillStyle is the pre-computed fill style for status bars.
var statusBarFillStyle = lipgloss.NewStyle().Background(lipgloss.Color("236"))

// renderStatusBar builds a complete status bar with consistent styling.
// parts are plain-text, view-specific segments (e.g. word count, grade, status).
// helpKey is the help hint (e.g. "? help", "⌥? help").
func renderStatusBar(ctx *ViewContext, left string, parts []string, helpKey string) string {
	if ctx.mouseEnabled {
		parts = append(parts, "↕")
	}
	parts = append(parts, helpKey)
	right := statusBarHintStyle.Render(strings.Join(parts, " | "))
	return statusBarFill(left, right, ctx.width)
}

// statusBarFill builds a status bar row: left + fill + right, padded to width.
func statusBarFill(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	fill := statusBarFillStyle.Render(strings.Repeat(" ", gap))
	return left + fill + right
}
