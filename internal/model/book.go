package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/inkcheck/ink/internal/render"
)

// clearBookStatusMsg clears the Book status bar feedback text.
type clearBookStatusMsg struct{}

// Book is the file browser view.
type Book struct {
	list        list.Model
	ctx         *ViewContext
	bookName    string
	dir         string
	rootDir     string
	naming      bool
	input       textinput.Model
	statusText  string
	showHelp    bool
	preFiltered bool // true when built from explicit file args (no directory navigation)
}

// newBookList creates a configured list.Model for the book view.
func newBookList(items []list.Item, ctx *ViewContext) list.Model {
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, ctx.contentWidth(), ctx.height-bookChromeHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.KeyMap.PrevPage.SetKeys("pgup", "b", "u", "ctrl+b")
	l.KeyMap.NextPage.SetKeys("pgdown", "f", "d", "ctrl+f")
	return l
}

// NewBook creates a new Book file browser for the given directory.
func NewBook(ctx *ViewContext, dir string) Book {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	items, err := scanDir(absDir)
	if err != nil {
		items = nil
	}

	return Book{
		list:     newBookList(items, ctx),
		ctx:      ctx,
		bookName: dirToBookName(absDir),
		dir:      absDir,
		rootDir:  absDir,
	}
}

// NewBookFromFiles creates a Book view from explicit file/directory paths
// instead of scanning a directory. Used when ink is called with multiple args.
func NewBookFromFiles(ctx *ViewContext, files []string) Book {
	var items []list.Item
	for _, f := range files {
		absPath, err := filepath.Abs(f)
		if err != nil {
			absPath = f
		}
		info, err := os.Stat(absPath)
		if err != nil {
			continue
		}
		if info.IsDir() {
			mc := countMarkdownFiles(absPath)
			if mc > 0 {
				items = append(items, dirItem{
					name:    filepath.Base(absPath),
					path:    absPath,
					mdCount: mc,
				})
			}
		} else {
			items = append(items, fileItem{
				name:    filepath.Base(absPath),
				path:    absPath,
				modTime: info.ModTime(),
			})
		}
	}

	// Derive common parent directory
	parentDir := commonParentDir(files)

	return Book{
		list:        newBookList(items, ctx),
		ctx:         ctx,
		bookName:    dirToBookName(parentDir),
		dir:         parentDir,
		rootDir:     parentDir,
		preFiltered: true,
	}
}

func (b *Book) changeDir(dir string) {
	b.dir = dir
	b.bookName = dirToBookName(dir)
	b.ctx.bookName = b.bookName
	items, err := scanDir(dir)
	if err != nil {
		b.statusText = "Error: " + err.Error()
		return
	}
	b.list.SetItems(items)
	b.list.ResetSelected()
}

// createFile validates the name, writes a new markdown file with frontmatter,
// and refreshes the directory listing.
func (b *Book) createFile(raw string) tea.Cmd {
	name := strings.TrimSpace(raw)
	if name == "" {
		b.naming = false
		return nil
	}
	if !strings.HasSuffix(strings.ToLower(name), ".md") {
		name += ".md"
	}
	filePath := filepath.Join(b.dir, name)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		b.naming = false
		b.statusText = "Invalid filename"
		return clearStatusAfter(2*time.Second, clearBookStatusMsg{})
	}
	rel, err := filepath.Rel(b.dir, absPath)
	if err != nil || strings.HasPrefix(rel, "..") || strings.Contains(rel, string(os.PathSeparator)) {
		b.naming = false
		b.statusText = "Invalid filename"
		return clearStatusAfter(2*time.Second, clearBookStatusMsg{})
	}
	title := strings.TrimSuffix(name, filepath.Ext(name))
	frontmatter := fmt.Sprintf("---\ntitle: %q\nauthor: %s\ndate: %s\n---\n",
		title, currentUser(), time.Now().Format(time.RFC3339))
	if err := os.WriteFile(absPath, []byte(frontmatter), 0644); err != nil {
		b.naming = false
		b.statusText = "Error: " + err.Error()
		return clearStatusAfter(2*time.Second, clearBookStatusMsg{})
	}
	b.naming = false
	b.changeDir(b.dir)
	return nil
}

func (b Book) Init() tea.Cmd {
	return nil
}

func (b Book) Update(msg tea.Msg) (Book, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		filtering := b.list.FilterState() == list.Filtering
		b.list.SetSize(b.ctx.contentWidth(), bookListHeight(b.ctx, b.showHelp, filtering))
	case clearBookStatusMsg:
		b.statusText = ""
		return b, nil
	case tea.KeyMsg:
		// Handle naming mode input
		if b.naming {
			switch msg.String() {
			case "enter":
				return b, b.createFile(b.input.Value())
			case "esc":
				b.naming = false
				return b, nil
			}
			var cmd tea.Cmd
			b.input, cmd = b.input.Update(msg)
			return b, cmd
		}
		// Don't intercept keys when filtering is active
		if b.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "enter", "right", "l":
			selected := b.list.SelectedItem()
			switch item := selected.(type) {
			case dirItem:
				b.changeDir(item.path)
				return b, nil
			case fileItem:
				return b, func() tea.Msg {
					return OpenChapterMsg{FilePath: item.path}
				}
			}
		case "backspace", "left", "h":
			if !b.preFiltered && b.dir != b.rootDir {
				b.changeDir(filepath.Dir(b.dir))
				return b, nil
			}
		case "n":
			if b.preFiltered {
				b.statusText = "Not allowed"
				return b, clearStatusAfter(2*time.Second, clearBookStatusMsg{})
			}
			ti := textinput.New()
			ti.Placeholder = "filename.md"
			ti.Focus()
			ti.CharLimit = 255
			b.input = ti
			b.naming = true
			return b, ti.Cursor.BlinkCmd()
		case "r", "ctrl+r":
			b.changeDir(b.dir)
			return b, nil
		case "esc":
			if b.showHelp {
				b.showHelp = false
				filtering := b.list.FilterState() == list.Filtering
				b.list.SetSize(b.ctx.contentWidth(), bookListHeight(b.ctx, b.showHelp, filtering))
				return b, nil
			}
		case "?":
			b.showHelp = !b.showHelp
			filtering := b.list.FilterState() == list.Filtering
			b.list.SetSize(b.ctx.contentWidth(), bookListHeight(b.ctx, b.showHelp, filtering))
			return b, nil
		case "ctrl+w":
			return b, tea.Quit
		}
	}

	var cmd tea.Cmd
	prevFilterState := b.list.FilterState()
	b.list, cmd = b.list.Update(msg)
	// Resize the list when filter state changes so the filter input line
	// doesn't steal a row from the visible items.
	filtering := b.list.FilterState() == list.Filtering
	if b.list.FilterState() != prevFilterState {
		b.list.SetSize(b.ctx.contentWidth(), bookListHeight(b.ctx, b.showHelp, filtering))
	}
	return b, cmd
}

const bookHelpHeight = 3

func bookListHeight(ctx *ViewContext, showHelp bool, filtering bool) int {
	h := contentHeight(ctx, bookChromeHeight, bookHelpHeight, showHelp)
	// When filtering, the list component uses one row for the filter input;
	// give it an extra row so the visible item count stays the same.
	if filtering {
		h++
	}
	return h
}

func (b Book) helpView() string {
	return renderHelpPane([][]helpEntry{
		{{"k/↑", "up"}, {"j/↓", "down"}, {"enter", "open"}},
		{{"backspace", "back"}, {"n", "new file"}, {"/", "filter"}},
		{{"r", "reload"}, {"?", "toggle help"}, {"ctrl+w", "quit"}},
	}, b.ctx.width)
}

func (b Book) statusBarView() string {
	w := b.ctx.width

	if b.naming {
		promptStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
		label := promptStyle.Render("New file:")
		inputStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(0, 1)
		input := inputStyle.Render(b.input.View())
		left := label + input
		return statusBarFill(left, "", w)
	}

	left := statusBarBookName(b.bookName)

	// Right side: status text + hints
	n := b.docCount()
	hints := fmt.Sprintf("%d %s | ? help", n, pluralize(n, "document", "documents"))
	if b.statusText != "" {
		hints = statusBarStatusStyle.Render(b.statusText) + "  " + hints
	}
	right := statusBarHintStyle.Render(hints)

	return statusBarFill(left, right, w)
}

func (b Book) docCount() int {
	count := 0
	for _, item := range b.list.Items() {
		if _, ok := item.(fileItem); ok {
			count++
		}
	}
	return count
}

func (b Book) View() string {
	title := render.H1Style.Render(b.bookName)
	// Reserve a blank line for the filter input so the list doesn't jump
	// when "/" is pressed. When filtering is active, the list component
	// renders its own filter input line, so we drop the placeholder.
	filtering := b.list.FilterState() == list.Filtering
	filterLine := "\n"
	if filtering {
		filterLine = ""
	}
	content := centerContent(title+"\n"+filterLine+"\n"+b.list.View(), b.ctx.width, b.ctx.maxWidth)
	var helpPane string
	if b.showHelp {
		helpPane = b.helpView()
	}
	return layoutView(logo, content, b.statusBarView(), helpPane)
}
