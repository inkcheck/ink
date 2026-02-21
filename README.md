# Ink

A terminal markdown viewer and distraction-free editor built on [Charm](https://charm.sh) libraries.

## Install

### Homebrew

```bash
brew install inkcheck/tap/ink
```

### Go

```bash
go install github.com/inkcheck/ink/cmd/ink@latest
```

### Build from source

```bash
git clone https://github.com/inkcheck/ink.git
cd ink
go install ./cmd/ink
```

## Usage

```bash
ink              # browse .md files in current directory
ink /some/path   # browse .md files in a specific directory
ink -w 100       # set max content width (default: 80)
```

## Key Bindings

### Book (file browser)

| Key        | Action              |
|------------|---------------------|
| j/k/arrows | Navigate            |
| enter/l    | Open file or folder |
| h/left     | Go to parent folder |
| n          | Create new file     |
| /          | Filter files        |
| ctrl+w     | Quit                |

### Chapter (viewer)

| Key        | Action              |
|------------|---------------------|
| j/k/arrows | Scroll              |
| b/f        | Page up/down        |
| u/d        | Half page up/down   |
| g/G        | Top/bottom          |
| e          | Open editor         |
| E          | Open in $EDITOR     |
| y          | Copy to clipboard   |
| ?          | Toggle help         |
| esc        | Back to Book        |

### Editor

| Key    | Action         |
|--------|----------------|
| ctrl+s | Save file      |
| ctrl+f | Half page down |
| ctrl+b | Half page up   |
| ctrl+t | Go to top      |
| ctrl+g | Go to bottom   |
| ctrl+w | Close editor   |
| esc    | Close editor   |

`ctrl+c` quits from any view.

## Features

- Markdown rendering with styled headings, code blocks, blockquotes, and lists
- Distraction-free editor with live word count
- Flesch-Kincaid readability grade in viewer and editor
- Directory browsing with subdirectory navigation
- Clipboard copy support
- External editor integration via $EDITOR
- Centered content on wide terminals

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - list, viewport, textarea components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - terminal styling
- [Goldmark](https://github.com/yuin/goldmark) - markdown parsing
- [Readability](https://github.com/inkcheck/readability) - Flesch-Kincaid scoring
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) - AI-assisted development

## License

[MIT](LICENSE)
