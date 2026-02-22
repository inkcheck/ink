package model

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/inkcheck/ink/internal/render"
)

// clearStatusMsg clears the status bar feedback text.
type clearStatusMsg struct{}

// Chapter is the markdown viewer.
type Chapter struct {
	viewport   viewport.Model
	filePath   string
	content    string // raw markdown
	ctx        *ViewContext
	showHelp   bool
	statusText string
	grade      string // cached FK grade
}

// NewChapter creates a new Chapter viewer for the given file.
func NewChapter(ctx *ViewContext, filePath string) Chapter {
	vp := viewport.New(ctx.width, chapterViewportHeight(ctx, false))
	ch := Chapter{
		filePath: filePath,
		ctx:      ctx,
		viewport: vp,
	}
	ch.refresh()
	return ch
}

func (c Chapter) Init() tea.Cmd {
	return nil
}

func (c Chapter) Update(msg tea.Msg) (Chapter, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.Width = c.ctx.width
		c.viewport.Height = chapterViewportHeight(c.ctx, c.showHelp)
		if c.content != "" {
			c.setRenderedContent()
		}
	case ExternalEditorClosedMsg:
		if msg.Err != nil {
			c.statusText = "Editor error: " + msg.Err.Error()
		}
		c.refresh()
		return c, clearStatusAfter(2*time.Second, clearStatusMsg{})
	case clearStatusMsg:
		c.statusText = ""
		return c, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "left", "h", "ctrl+w":
			if c.showHelp {
				c.showHelp = false
				c.viewport.Height = chapterViewportHeight(c.ctx, false)
				return c, nil
			}
			// When there's no book, only esc and ctrl+w close; left/h are ignored
			if !c.ctx.isBook && (msg.String() == "left" || msg.String() == "h") {
				break
			}
			return c, func() tea.Msg { return BackToBookMsg{} }
		case "e":
			return c, func() tea.Msg {
				return OpenEditorMsg{
					FilePath: c.filePath,
					Content:  c.content,
				}
			}
		case "E":
			return c, func() tea.Msg {
				return OpenExternalEditorMsg{FilePath: c.filePath}
			}
		case "y":
			if err := clipboard.WriteAll(c.content); err != nil {
				c.statusText = "Copy failed"
			} else {
				c.statusText = "Copied!"
			}
			return c, clearStatusAfter(2*time.Second, clearStatusMsg{})
		case "r", "ctrl+r":
			c.refresh()
			return c, nil
		case "?":
			c.showHelp = !c.showHelp
			c.viewport.Height = chapterViewportHeight(c.ctx, c.showHelp)
			if c.viewport.PastBottom() {
				c.viewport.GotoBottom()
			}
			return c, nil
		case "b", "pgup":
			c.viewport.ViewUp()
			return c, nil
		case "f", "pgdown":
			c.viewport.ViewDown()
			return c, nil
		case "u", "ctrl+b":
			c.viewport.HalfViewUp()
			return c, nil
		case "d", "ctrl+f":
			c.viewport.HalfViewDown()
			return c, nil
		}
	}

	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)
	return c, cmd
}

const pagerHelpHeight = 3

func chapterViewportHeight(ctx *ViewContext, showHelp bool) int {
	return contentHeight(ctx, chapterChromeHeight, pagerHelpHeight, showHelp)
}

// setRenderedContent renders the current content and sets it on the viewport.
func (c *Chapter) setRenderedContent() {
	rendered := render.Render([]byte(c.content), c.ctx.maxWidth)
	centered := centerContent(rendered, c.viewport.Width, c.ctx.maxWidth)
	c.viewport.SetContent(centered)
}

func (c *Chapter) refresh() {
	raw, err := os.ReadFile(c.filePath)
	if err != nil {
		c.statusText = "Error reading file: " + err.Error()
		return
	}
	c.content = string(raw)
	c.grade = fleschKincaidGrade(c.content)
	c.setRenderedContent()
}

func (c Chapter) helpView() string {
	return renderHelpPane([][]helpEntry{
		{{"k/↑", "up"}, {"j/↓", "down"}, {"b", "page up"}, {"f", "page down"}},
		{{"u", "½ page up"}, {"d", "½ page down"}, {"g", "go to top"}, {"G", "go to bottom"}},
		{{"e", "edit file"}, {"E", "open in $EDITOR"}, {"y", "copy to clipboard"}, {"esc", "back"}},
	}, c.ctx.width)
}

func (c Chapter) statusBarView() string {
	w := c.ctx.width

	left := statusBarBookName(c.ctx.bookName) + statusBarFileName(c.filePath)

	// Scroll percentage
	percent := int(c.viewport.ScrollPercent() * 100)
	percentStr := fmt.Sprintf("%d%%", percent)

	// Right side: status text | percentage | grade | ? Help
	parts := []string{percentStr}
	if c.grade != "" {
		parts = append(parts, c.grade)
	}
	parts = append(parts, "? Help")
	rightText := strings.Join(parts, " | ")
	if c.statusText != "" {
		rightText = statusBarAccentStyle.Render(c.statusText) + "  " + rightText
	}
	right := statusBarHintStyle.Render(rightText)

	return statusBarFill(left, right, w)
}

func (c Chapter) View() string {
	var helpPane string
	if c.showHelp {
		helpPane = c.helpView()
	}
	return layoutView(logo, c.viewport.View(), c.statusBarView(), helpPane)
}
