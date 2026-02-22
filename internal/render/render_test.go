package render

import (
	"strings"
	"testing"
)

func TestRenderHeadings(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{"H1", "# Hello World", "Hello World"},
		{"H2", "## Section Two", "Section Two"},
		{"H3", "### Section Three", "Section Three"},
		{"H4", "#### Section Four", "Section Four"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render([]byte(tt.markdown), 80)
			if !strings.Contains(got, tt.want) {
				t.Errorf("Render(%q) = %q, want it to contain %q", tt.markdown, got, tt.want)
			}
		})
	}
}

func TestRenderParagraph(t *testing.T) {
	md := "This is a paragraph of text."
	got := Render([]byte(md), 80)
	if !strings.Contains(got, "This is a paragraph of text.") {
		t.Errorf("Render paragraph: got %q", got)
	}
}

func TestRenderFencedCodeBlock(t *testing.T) {
	md := "```go\nfmt.Println(\"hello\")\n```"
	got := Render([]byte(md), 80)
	if !strings.Contains(got, `fmt.Println("hello")`) {
		t.Errorf("Render code block: got %q", got)
	}
}

func TestRenderBlockquote(t *testing.T) {
	md := "> This is a quote"
	got := Render([]byte(md), 80)
	if !strings.Contains(got, "This is a quote") {
		t.Errorf("Render blockquote: got %q", got)
	}
}

func TestRenderUnorderedList(t *testing.T) {
	md := "- alpha\n- beta\n- gamma"
	got := Render([]byte(md), 80)
	for _, item := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(got, item) {
			t.Errorf("Render unordered list: missing %q in %q", item, got)
		}
	}
	if !strings.Contains(got, "•") {
		t.Errorf("Render unordered list: missing bullet marker in %q", got)
	}
}

func TestRenderOrderedList(t *testing.T) {
	md := "1. first\n2. second\n3. third"
	got := Render([]byte(md), 80)
	for _, item := range []string{"first", "second", "third"} {
		if !strings.Contains(got, item) {
			t.Errorf("Render ordered list: missing %q in %q", item, got)
		}
	}
	if !strings.Contains(got, "1.") {
		t.Errorf("Render ordered list: missing numbering in %q", got)
	}
}

func TestRenderNestedList(t *testing.T) {
	md := "- outer\n  - inner"
	got := Render([]byte(md), 80)
	if !strings.Contains(got, "outer") || !strings.Contains(got, "inner") {
		t.Errorf("Render nested list: got %q", got)
	}
}

func TestRenderTable(t *testing.T) {
	md := "| Name | Age |\n|------|-----|\n| Alice | 30 |\n| Bob | 25 |"
	got := Render([]byte(md), 80)
	for _, cell := range []string{"Name", "Age", "Alice", "30", "Bob", "25"} {
		if !strings.Contains(got, cell) {
			t.Errorf("Render table: missing %q in %q", cell, got)
		}
	}
	// Table should have border characters
	if !strings.Contains(got, "│") {
		t.Errorf("Render table: missing border character in %q", got)
	}
}

func TestRenderInlineElements(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{"bold", "This is **bold** text", "bold"},
		{"italic", "This is *italic* text", "italic"},
		{"code span", "Use `fmt.Println`", "fmt.Println"},
		{"link", "[Go](https://go.dev)", "Go"},
		{"link URL", "[Go](https://go.dev)", "https://go.dev"},
		{"image", "![alt text](image.png)", "[image: alt text]"},
		{"strikethrough", "This is ~~deleted~~ text", "deleted"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render([]byte(tt.markdown), 80)
			if !strings.Contains(got, tt.want) {
				t.Errorf("Render(%q) = %q, want it to contain %q", tt.markdown, got, tt.want)
			}
		})
	}
}

func TestRenderTaskCheckboxes(t *testing.T) {
	md := "- [x] done\n- [ ] todo"
	got := Render([]byte(md), 80)
	if !strings.Contains(got, "☑") {
		t.Errorf("Render task list: missing checked box in %q", got)
	}
	if !strings.Contains(got, "☐") {
		t.Errorf("Render task list: missing unchecked box in %q", got)
	}
}

func TestRenderFrontmatterStripping(t *testing.T) {
	md := "---\ntitle: Test\nauthor: Me\n---\n\n# Hello"
	got := Render([]byte(md), 80)
	if strings.Contains(got, "title:") {
		t.Errorf("Render frontmatter: frontmatter not stripped, got %q", got)
	}
	if !strings.Contains(got, "Hello") {
		t.Errorf("Render frontmatter: content missing after stripping, got %q", got)
	}
}

func TestRenderFrontmatterStrippingCRLF(t *testing.T) {
	md := "---\r\ntitle: Test\r\n---\r\n\r\n# Hello"
	got := Render([]byte(md), 80)
	if strings.Contains(got, "title:") {
		t.Errorf("Render CRLF frontmatter: not stripped, got %q", got)
	}
	if !strings.Contains(got, "Hello") {
		t.Errorf("Render CRLF frontmatter: content missing, got %q", got)
	}
}

func TestRenderThematicBreak(t *testing.T) {
	md := "above\n\n---\n\nbelow"
	got := Render([]byte(md), 80)
	if !strings.Contains(got, "─") {
		t.Errorf("Render thematic break: missing separator in %q", got)
	}
	if !strings.Contains(got, "above") || !strings.Contains(got, "below") {
		t.Errorf("Render thematic break: missing content around break, got %q", got)
	}
}

func TestRenderEmptyInput(t *testing.T) {
	got := Render([]byte(""), 80)
	if got != "" {
		t.Errorf("Render empty: expected empty string, got %q", got)
	}
}

func TestRenderWhitespaceOnly(t *testing.T) {
	got := Render([]byte("   \n\n  "), 80)
	// Should produce no meaningful content
	trimmed := strings.TrimSpace(got)
	if trimmed != "" {
		t.Errorf("Render whitespace: expected empty or whitespace, got %q", got)
	}
}

func TestRenderNoFrontmatterPassthrough(t *testing.T) {
	// Content that starts with --- but isn't valid frontmatter (no closing ---)
	md := "---\nno closing delimiter"
	got := Render([]byte(md), 80)
	// Should render the thematic break or the raw text, not crash
	if got == "" {
		t.Errorf("Render malformed frontmatter: unexpected empty output")
	}
}
