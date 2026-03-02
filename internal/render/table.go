package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	east "github.com/yuin/goldmark/extension/ast"
)

// renderTable renders a GFM table with proportional column widths and text wrapping.
func renderTable(buf *strings.Builder, table *east.Table, source []byte, maxWidth int) {
	var rows [][]string
	var isHeader []bool

	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		var cells []string
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			cells = append(cells, renderInlineChildren(cell, source))
		}
		rows = append(rows, cells)
		_, hdr := row.(*east.TableHeader)
		isHeader = append(isHeader, hdr)
	}

	if len(rows) == 0 {
		return
	}

	numCols := 0
	for _, r := range rows {
		if len(r) > numCols {
			numCols = len(r)
		}
	}

	colWidths := computeColumnWidths(rows, numCols, maxWidth)

	var sepParts []string
	for _, w := range colWidths {
		sepParts = append(sepParts, strings.Repeat("─", w+2))
	}
	separator := "├" + strings.Join(sepParts, "┼") + "┤"

	for i, row := range rows {
		renderTableRow(buf, row, colWidths, numCols, table.Alignments, isHeader[i])
		if isHeader[i] {
			buf.WriteString(TableBorderStyle.Render(separator))
			buf.WriteString("\n")
		}
	}
	buf.WriteString("\n")
}

// computeColumnWidths returns column widths that fit within maxWidth,
// using natural widths when possible and proportional distribution otherwise.
func computeColumnWidths(rows [][]string, numCols, maxWidth int) []int {
	natural := make([]int, numCols)
	for _, r := range rows {
		for i, cell := range r {
			if w := lipgloss.Width(cell); w > natural[i] {
				natural[i] = w
			}
		}
	}

	// Overhead: 1 leading │ + numCols × (1 space + 1 space + 1 │)
	overhead := 1 + numCols*3
	available := maxWidth - overhead
	if available < numCols {
		available = numCols
	}

	totalNatural := 0
	for _, w := range natural {
		totalNatural += w
	}

	if totalNatural <= available {
		return natural
	}

	const minColWidth = 5
	widths := make([]int, numCols)
	remaining := available
	for i, nat := range natural {
		w := nat * available / totalNatural
		if w < minColWidth {
			w = minColWidth
		}
		if w > remaining {
			w = remaining
		}
		widths[i] = w
		remaining -= w
	}
	for i := range widths {
		if remaining <= 0 {
			break
		}
		widths[i]++
		remaining--
	}

	return widths
}

// renderTableRow renders a single table row, wrapping cell content and
// aligning multi-line output across columns.
func renderTableRow(buf *strings.Builder, row []string, colWidths []int, numCols int, alignments []east.Alignment, isHeader bool) {
	cellLines := make([][]string, numCols)
	maxLines := 0

	for j := 0; j < numCols; j++ {
		cell := ""
		if j < len(row) {
			cell = row[j]
		}
		wrapped := lipgloss.NewStyle().Width(colWidths[j]).Render(cell)
		lines := strings.Split(wrapped, "\n")
		cellLines[j] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for line := 0; line < maxLines; line++ {
		var out strings.Builder
		out.WriteString(TableBorderStyle.Render("│"))
		for j := 0; j < numCols; j++ {
			content := ""
			if line < len(cellLines[j]) {
				content = cellLines[j][line]
			}
			align := east.AlignNone
			if j < len(alignments) {
				align = alignments[j]
			}
			padded := " " + alignCell(content, colWidths[j], align) + " "
			if isHeader {
				out.WriteString(TableHeaderStyle.Render(padded))
			} else {
				out.WriteString(TableCellStyle.Render(padded))
			}
			out.WriteString(TableBorderStyle.Render("│"))
		}
		buf.WriteString(out.String())
		buf.WriteString("\n")
	}
}

// alignCell pads content to width according to alignment.
func alignCell(s string, width int, align east.Alignment) string {
	gap := width - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	switch align {
	case east.AlignCenter:
		left := gap / 2
		right := gap - left
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	case east.AlignRight:
		return strings.Repeat(" ", gap) + s
	default:
		return s + strings.Repeat(" ", gap)
	}
}
