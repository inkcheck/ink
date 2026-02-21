package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/inkcheck/ink/internal/model"
)

func main() {
	width := flag.Int("w", 80, "max content width")
	flag.Parse()
	if *width < 1 {
		*width = 1
	}
	if *width > 200 {
		*width = 200
	}

	args := flag.Args()

	var m tea.Model
	switch {
	case len(args) == 0:
		// No args: browse current directory
		m = model.New(".", *width)

	case len(args) == 1:
		arg := args[0]
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if info.IsDir() {
			// Single directory arg: browse that directory
			m = model.New(arg, *width)
		} else {
			// Single file arg: must be .md
			if !strings.HasSuffix(strings.ToLower(arg), ".md") {
				fmt.Fprintf(os.Stderr, "Error: %s is not a markdown file\n", arg)
				os.Exit(1)
			}
			m = model.NewFromFile(arg, *width)
		}

	default:
		// Multiple args (e.g. shell glob expansion): collect valid paths
		var files []string
		for _, arg := range args {
			info, err := os.Stat(arg)
			if err != nil {
				continue // skip non-existent paths
			}
			if info.IsDir() {
				files = append(files, arg)
			} else if strings.HasSuffix(strings.ToLower(arg), ".md") {
				files = append(files, arg)
			}
		}
		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no markdown files found in arguments\n")
			os.Exit(1)
		}
		m = model.NewFromFiles(files, *width)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
