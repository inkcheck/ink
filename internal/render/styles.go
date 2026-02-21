package render

import "github.com/charmbracelet/lipgloss"

var (
	H1Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("63")).
		Padding(0, 1)

	H2Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginTop(1)

	H3Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")).
		MarginTop(1)

	H4Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("105"))

	ParagraphStyle = lipgloss.NewStyle().
			MarginBottom(1)

	CodeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(1, 2).
			MarginBottom(1)

	InlineCodeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("213"))

	BlockquoteStyle = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("240")).
			PaddingLeft(2).
			MarginTop(1).
			MarginBottom(1)

	LinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("87")).
			Underline(true)

	EmphasisStyle = lipgloss.NewStyle().
			Italic(true)

	StrongStyle = lipgloss.NewStyle().
			Bold(true)

	ThematicBreakStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginTop(1).
				MarginBottom(1)

	StrikethroughStyle = lipgloss.NewStyle().
				Strikethrough(true).
				Foreground(lipgloss.Color("245"))

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170"))

	TableCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	TableBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
)
