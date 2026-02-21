package model

import (
	"fmt"
	"math"
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

// fileItem represents a markdown file in the list.
type fileItem struct {
	name    string
	path    string
	modTime time.Time
}

func (f fileItem) Title() string       { return f.name }
func (f fileItem) Description() string { return relativeTime(f.modTime) }
func (f fileItem) FilterValue() string { return f.name }

// dirItem represents a navigable folder in the list.
type dirItem struct {
	name    string
	path    string
	mdCount int
}

func (d dirItem) Title() string { return d.name + "/" }
func (d dirItem) Description() string {
	if d.mdCount == 1 {
		return "1 document"
	}
	return fmt.Sprintf("%d documents", d.mdCount)
}
func (d dirItem) FilterValue() string { return d.name }

// Book is the file browser view.
type Book struct {
	list       list.Model
	common     *Common
	bookName   string
	dir        string
	rootDir    string
	naming     bool
	input      textinput.Model
	statusText string
	showHelp   bool
	preFiltered  bool // true when built from explicit file args (no directory navigation)
}

// NewBook creates a new Book file browser for the given directory.
func NewBook(common *Common, dir string) Book {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	items, err := scanDir(absDir)
	if err != nil {
		items = nil
	}
	delegate := list.NewDefaultDelegate()
	listWidth := common.ContentWidth()
	l := list.New(items, delegate, listWidth, common.Height-bookChromeHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.KeyMap.PrevPage.SetKeys("pgup", "b", "u", "ctrl+b")
	l.KeyMap.NextPage.SetKeys("pgdown", "f", "d", "ctrl+f")

	return Book{
		list:     l,
		common:   common,
		bookName: dirToBookName(absDir),
		dir:      absDir,
		rootDir:  absDir,
	}
}

// NewBookFromFiles creates a Book view from explicit file/directory paths
// instead of scanning a directory. Used when ink is called with multiple args.
func NewBookFromFiles(common *Common, files []string) Book {
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

	delegate := list.NewDefaultDelegate()
	listWidth := common.ContentWidth()
	l := list.New(items, delegate, listWidth, common.Height-bookChromeHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.KeyMap.PrevPage.SetKeys("pgup", "b", "u", "ctrl+b")
	l.KeyMap.NextPage.SetKeys("pgdown", "f", "d", "ctrl+f")

	return Book{
		list:      l,
		common:    common,
		bookName:  dirToBookName(parentDir),
		dir:       parentDir,
		rootDir:   parentDir,
		preFiltered: true,
	}
}

// commonParentDir finds the common parent directory of a list of paths.
func commonParentDir(paths []string) string {
	if len(paths) == 0 {
		abs, _ := filepath.Abs(".")
		return abs
	}
	abs, err := filepath.Abs(paths[0])
	if err != nil {
		abs = paths[0]
	}
	parent := filepath.Dir(abs)
	for _, p := range paths[1:] {
		a, err := filepath.Abs(p)
		if err != nil {
			a = p
		}
		d := filepath.Dir(a)
		for !strings.HasPrefix(d+string(os.PathSeparator), parent+string(os.PathSeparator)) &&
			parent != string(os.PathSeparator) && parent != "." {
			parent = filepath.Dir(parent)
		}
	}
	return parent
}

func dirToBookName(dir string) string {
	name := filepath.Base(dir)
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ToUpper(name)
	return name
}

func scanDir(dir string) ([]list.Item, error) {
	var dirs []list.Item
	var files []list.Item
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if e.IsDir() {
			subPath := filepath.Join(dir, name)
			mc := countMarkdownFiles(subPath)
			if mc > 0 {
				dirs = append(dirs, dirItem{
					name:    name,
					path:    subPath,
					mdCount: mc,
				})
			}
		} else if strings.HasSuffix(strings.ToLower(name), ".md") {
			info, err := e.Info()
			var modTime time.Time
			if err == nil {
				modTime = info.ModTime()
			}
			files = append(files, fileItem{
				name:    name,
				path:    filepath.Join(dir, name),
				modTime: modTime,
			})
		}
	}
	// Directories first, then files
	return append(dirs, files...), nil
}

func countMarkdownFiles(dir string) int {
	count := 0
	dirDepth := strings.Count(dir, string(os.PathSeparator))
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
		}
		depth := strings.Count(path, string(os.PathSeparator)) - dirDepth
		if d.IsDir() && depth > 3 {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			count++
		}
		return nil
	})
	return count
}

func (b *Book) changeDir(dir string) {
	b.dir = dir
	b.bookName = dirToBookName(dir)
	b.common.BookName = b.bookName
	items, err := scanDir(dir)
	if err != nil {
		b.statusText = "Error: " + err.Error()
		return
	}
	b.list.SetItems(items)
	b.list.ResetSelected()
}

func (b Book) Init() tea.Cmd {
	return nil
}

func (b Book) Update(msg tea.Msg) (Book, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.list.SetSize(b.common.ContentWidth(), bookListHeight(b.common, b.showHelp))
	case clearBookStatusMsg:
		b.statusText = ""
		return b, nil
	case tea.KeyMsg:
		// Handle naming mode input
		if b.naming {
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(b.input.Value())
				if name == "" {
					b.naming = false
					return b, nil
				}
				if !strings.HasSuffix(strings.ToLower(name), ".md") {
					name += ".md"
				}
				filePath := filepath.Join(b.dir, name)
				absPath, err := filepath.Abs(filePath)
				if err != nil {
					b.naming = false
					b.statusText = "Invalid filename"
					return b, clearStatusAfter(2*time.Second, clearBookStatusMsg{})
				}
				rel, err := filepath.Rel(b.dir, absPath)
				if err != nil || strings.HasPrefix(rel, "..") || strings.Contains(rel, string(os.PathSeparator)) {
					b.naming = false
					b.statusText = "Invalid filename"
					return b, clearStatusAfter(2*time.Second, clearBookStatusMsg{})
				}
				title := strings.TrimSuffix(name, filepath.Ext(name))
				user := currentUser()
				frontmatter := fmt.Sprintf("---\ntitle: %q\nauthor: %s\ndate: %s\n---\n",
					title, user, time.Now().Format(time.RFC3339))
				if err := os.WriteFile(absPath, []byte(frontmatter), 0644); err != nil {
					b.naming = false
					b.statusText = "Error: " + err.Error()
					return b, clearStatusAfter(2*time.Second, clearBookStatusMsg{})
				}
				b.naming = false
				b.changeDir(b.dir)
				return b, nil
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
				b.list.SetSize(b.common.ContentWidth(), bookListHeight(b.common, b.showHelp))
				return b, nil
			}
		case "?":
			b.showHelp = !b.showHelp
			b.list.SetSize(b.common.ContentWidth(), bookListHeight(b.common, b.showHelp))
			return b, nil
		case "ctrl+w":
			return b, tea.Quit
		}
	}

	var cmd tea.Cmd
	b.list, cmd = b.list.Update(msg)
	return b, cmd
}

const bookHelpHeight = 3

func bookListHeight(common *Common, showHelp bool) int {
	h := common.Height - bookChromeHeight
	if showHelp {
		h -= bookHelpHeight
	}
	if h < 1 {
		h = 1
	}
	return h
}

func (b Book) helpView() string {
	return renderHelpPane([][]helpEntry{
		{{"k/↑", "up"}, {"j/↓", "down"}, {"enter", "open"}},
		{{"backspace", "back"}, {"n", "new file"}, {"/", "filter"}},
		{{"r", "reload"}, {"?", "toggle help"}, {"ctrl+w", "quit"}},
	}, b.common.Width)
}

func (b Book) statusBarView() string {
	w := b.common.Width

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
	hints := fmt.Sprintf("%d %s | ? help", b.docCount(), pluralize(b.docCount(), "document", "documents"))
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

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func (b Book) View() string {
	title := render.H1Style.Render(b.bookName)
	content := centerContent(title+"\n\n"+b.list.View(), b.common.Width, b.common.MaxWidth)
	var helpPane string
	if b.showHelp {
		helpPane = b.helpView()
	}
	return layoutView(logo, content, b.statusBarView(), helpPane)
}

// Duration constants for relativeTime calculations.
const (
	day   = 24 * time.Hour
	month = 30 * day
	year  = 365 * day
)

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
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
