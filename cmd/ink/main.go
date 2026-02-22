package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/inkcheck/ink/internal/model"
)

func parseFlags() int {
	width := flag.Int("w", 80, "max content width")
	flag.Parse()
	if *width < 1 {
		*width = 1
	}
	if *width > 200 {
		*width = 200
	}
	return *width
}

func resolveModel(args []string, width int) (tea.Model, error) {
	switch {
	case len(args) == 0:
		return model.New(".", width), nil

	case len(args) == 1:
		arg := args[0]
		info, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			return model.New(arg, width), nil
		}
		if !strings.HasSuffix(strings.ToLower(arg), ".md") {
			return nil, fmt.Errorf("%s is not a markdown file", arg)
		}
		return model.NewFromFile(arg, width), nil

	default:
		var files []string
		for _, arg := range args {
			info, err := os.Stat(arg)
			if err != nil {
				continue
			}
			if info.IsDir() {
				files = append(files, arg)
			} else if strings.HasSuffix(strings.ToLower(arg), ".md") {
				files = append(files, arg)
			}
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("no markdown files found in arguments")
		}
		return model.NewFromFiles(files, width), nil
	}
}

func main() {
	width := parseFlags()
	m, err := resolveModel(flag.Args(), width)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
