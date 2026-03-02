package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// mdParser is a reusable Goldmark parser instance with GFM support
// (Table, Strikethrough, Linkify, TaskList).
var mdParser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

// stripFrontMatter removes YAML front matter (--- delimited) from the start of source.
func stripFrontMatter(source []byte) []byte {
	if !bytes.HasPrefix(source, []byte("---")) {
		return source
	}
	// Normalize \r\n to \n for consistent delimiter matching
	normalized := bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))
	end := bytes.Index(normalized[3:], []byte("\n---"))
	if end < 0 {
		return source
	}
	// Skip past closing --- and the newline after it
	rest := normalized[3+end+4:]
	return bytes.TrimLeft(rest, "\n")
}

// BottomMargin is the number of blank lines appended after rendered content.
const BottomMargin = 4

// Render converts markdown source to lipgloss-styled terminal output.
func Render(source []byte, maxWidth int) string {
	source = stripFrontMatter(source)
	reader := text.NewReader(source)
	doc := mdParser.Parser().Parse(reader)

	var buf strings.Builder
	renderNode(&buf, doc, source, 0, maxWidth)

	result := buf.String()
	// Trim trailing whitespace
	result = strings.TrimRight(result, "\n")
	return result + strings.Repeat("\n", BottomMargin)
}

func renderNode(buf *strings.Builder, node ast.Node, source []byte, depth int, maxWidth int) {
	switch n := node.(type) {
	case *ast.Document:
		renderChildren(buf, n, source, depth, maxWidth)

	case *ast.Heading:
		content := renderInlineChildren(n, source)
		var styled string
		switch n.Level {
		case 1:
			badge := H1Style.Render(content)
			styled = lipgloss.NewStyle().Width(maxWidth).Render(badge)
		case 2:
			styled = H2Style.Width(maxWidth).Render(content)
		case 3:
			styled = H3Style.Width(maxWidth).Render(content)
		default:
			styled = H4Style.Width(maxWidth).Render(content)
		}
		buf.WriteString(styled)
		buf.WriteString("\n\n")

	case *ast.Paragraph:
		content := renderInlineChildren(n, source)
		styled := ParagraphStyle.Width(maxWidth).Render(content)
		buf.WriteString(styled)
		buf.WriteString("\n")

	case *ast.FencedCodeBlock, *ast.CodeBlock:
		var code bytes.Buffer
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			code.Write(line.Value(source))
		}
		text := strings.TrimRight(code.String(), "\n")
		styled := CodeBlockStyle.Width(maxWidth).Render(text)
		buf.WriteString(styled)
		buf.WriteString("\n\n")

	case *ast.Blockquote:
		// Border (1) + PaddingLeft (2) = 3 chars of overhead
		innerWidth := maxWidth - 3
		var inner strings.Builder
		renderChildren(&inner, n, source, depth+1, innerWidth)
		content := strings.TrimRight(inner.String(), "\n")
		styled := BlockquoteStyle.Width(maxWidth).Render(content)
		buf.WriteString(styled)
		buf.WriteString("\n\n")

	case *ast.List:
		renderChildren(buf, n, source, depth, maxWidth)
		buf.WriteString("\n")

	case *ast.ListItem:
		// Separate the item's own text from any nested lists so nested
		// lists start on their own line instead of being appended inline.
		var textBuf strings.Builder
		var listBuf strings.Builder
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if _, ok := child.(*ast.List); ok {
				renderNode(&listBuf, child, source, depth+1, maxWidth)
			} else {
				renderNode(&textBuf, child, source, depth+1, maxWidth)
			}
		}
		content := strings.TrimRight(textBuf.String(), "\n")
		indent := strings.Repeat("  ", depth)
		marker := "• "
		if parent, ok := n.Parent().(*ast.List); ok && parent.IsOrdered() {
			idx := parent.Start
			for sib := n.Parent().FirstChild(); sib != nil; sib = sib.NextSibling() {
				if sib == n {
					break
				}
				idx++
			}
			marker = fmt.Sprintf("%d. ", idx)
		}
		buf.WriteString(indent + marker + content + "\n")
		if listBuf.Len() > 0 {
			buf.WriteString(listBuf.String())
		}

	case *east.Table:
		renderTable(buf, n, source, maxWidth)

	case *ast.ThematicBreak:
		styled := ThematicBreakStyle.Width(maxWidth).Render("────────────────────────────────────────")
		buf.WriteString(styled)
		buf.WriteString("\n\n")

	case *ast.TextBlock:
		content := renderInlineChildren(n, source)
		buf.WriteString(content)

	default:
		renderChildren(buf, node, source, depth, maxWidth)
	}
}

func renderChildren(buf *strings.Builder, node ast.Node, source []byte, depth int, maxWidth int) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		renderNode(buf, child, source, depth, maxWidth)
	}
}

// renderInlineChildren collects inline content from a block node.
func renderInlineChildren(node ast.Node, source []byte) string {
	var buf strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		renderInline(&buf, child, source)
	}
	return buf.String()
}

func renderInline(buf *strings.Builder, node ast.Node, source []byte) {
	switch n := node.(type) {
	case *ast.Text:
		buf.Write(n.Segment.Value(source))
		if n.SoftLineBreak() {
			buf.WriteString(" ")
		}
		if n.HardLineBreak() {
			buf.WriteString("\n")
		}

	case *ast.String:
		buf.Write(n.Value)

	case *ast.CodeSpan:
		var code strings.Builder
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if t, ok := child.(*ast.Text); ok {
				code.Write(t.Segment.Value(source))
			}
		}
		styled := InlineCodeStyle.Render(code.String())
		buf.WriteString(styled)

	case *ast.Emphasis:
		content := renderInlineChildren(n, source)
		if n.Level == 2 {
			buf.WriteString(StrongStyle.Render(content))
		} else {
			buf.WriteString(EmphasisStyle.Render(content))
		}

	case *ast.Link:
		content := renderInlineChildren(n, source)
		url := string(n.Destination)
		styled := LinkStyle.Render(content + " (" + url + ")")
		buf.WriteString(styled)

	case *ast.AutoLink:
		url := string(n.URL(source))
		styled := LinkStyle.Render(url)
		buf.WriteString(styled)

	case *ast.Image:
		alt := renderInlineChildren(n, source)
		buf.WriteString("[image: " + alt + "]")

	case *ast.RawHTML:
		segments := n.Segments
		for i := 0; i < segments.Len(); i++ {
			seg := segments.At(i)
			buf.Write(seg.Value(source))
		}

	case *east.Strikethrough:
		content := renderInlineChildren(n, source)
		buf.WriteString(StrikethroughStyle.Render(content))

	case *east.TaskCheckBox:
		if n.IsChecked {
			buf.WriteString("☑ ")
		} else {
			buf.WriteString("☐ ")
		}

	default:
		// Try to render children for unknown inline nodes
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			renderInline(buf, child, source)
		}
	}
}

