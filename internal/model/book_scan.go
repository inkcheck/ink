package model

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

// isMarkdownFile reports whether name has a .md extension (case-insensitive).
func isMarkdownFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".md")
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
		} else if isMarkdownFile(name) {
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

// skipDirs contains directory names to exclude when scanning for markdown files.
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
}

func countMarkdownFiles(dir string) int {
	count := 0
	dirDepth := strings.Count(dir, string(os.PathSeparator))
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || skipDirs[name] {
				return filepath.SkipDir
			}
		}
		depth := strings.Count(path, string(os.PathSeparator)) - dirDepth
		if d.IsDir() && depth > 3 {
			return filepath.SkipDir
		}
		if !d.IsDir() && isMarkdownFile(d.Name()) {
			count++
		}
		return nil
	})
	return count
}
