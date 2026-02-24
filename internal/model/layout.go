package model

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout constants for chrome height calculations.
const (
	// bookChromeHeight is the total chrome for the book view (logo + gap + title + filter + gap + status).
	bookChromeHeight = 6
	// chapterChromeHeight is the total chrome for the chapter view (logo + gap + status).
	chapterChromeHeight = 3
	// editorChromeHeight is the total chrome for the editor view (logo + gap + status).
	editorChromeHeight = 3
	// metricsChromeHeight is the total chrome for the metrics view (logo + gap + status).
	metricsChromeHeight = 3
)

// logo is the pre-rendered application logo.
var logo = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("230")).
	Background(lipgloss.Color("135")).
	Padding(0, 1).
	Render("Ink")

// contentHeight calculates the available content height after subtracting chrome.
func contentHeight(ctx *ViewContext, chromeHeight, helpHeight int, showHelp bool) int {
	h := ctx.height - chromeHeight
	if showHelp {
		h -= helpHeight
	}
	return max(h, 1)
}

// centerContent horizontally centers content when the terminal is wider than maxWidth.
func centerContent(content string, termWidth, maxWidth int) string {
	if termWidth <= maxWidth {
		return content
	}
	block := lipgloss.NewStyle().Width(maxWidth).Render(content)
	return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, block)
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
