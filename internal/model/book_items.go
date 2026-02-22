package model

import (
	"fmt"
	"time"
)

// fileItem represents a markdown file in the list.
type fileItem struct {
	name    string
	path    string
	modTime time.Time
}

func (f fileItem) Title() string       { return f.name }
func (f fileItem) Description() string { return relativeTime(f.modTime, time.Now()) }
func (f fileItem) FilterValue() string { return f.name }

// dirItem represents a navigable folder in the list.
type dirItem struct {
	name    string
	path    string
	mdCount int
}

func (d dirItem) Title() string { return d.name + "/" }
func (d dirItem) Description() string {
	return fmt.Sprintf("%d %s", d.mdCount, pluralize(d.mdCount, "document", "documents"))
}
func (d dirItem) FilterValue() string { return d.name }
