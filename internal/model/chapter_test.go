package model

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestChapterViewLineCount(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"test.md": "# Test\n\n" + strings.Repeat("Line of text.\n", 100),
	})
	ctx := &ViewContext{width: 80, height: 45, maxWidth: 80}
	ch := NewChapter(ctx, filepath.Join(dir, "test.md"))

	viewNoHelp := ch.View()
	linesNoHelp := strings.Count(viewNoHelp, "\n") + 1

	t.Logf("Without help: %d lines (expected %d)", linesNoHelp, ctx.height)
	if linesNoHelp != ctx.height {
		t.Errorf("View() without help: got %d lines, want %d", linesNoHelp, ctx.height)
	}

	// Toggle help on
	ch, _ = ch.Update(tea.KeyPressMsg{Code: '?', Text: "?"})

	if !ch.help.Visible() {
		t.Fatal("help should be visible after toggle")
	}

	viewWithHelp := ch.View()
	linesWithHelp := strings.Count(viewWithHelp, "\n") + 1

	t.Logf("With help: %d lines (expected %d)", linesWithHelp, ctx.height)
	if linesWithHelp != ctx.height {
		t.Errorf("View() with help: got %d lines, want %d", linesWithHelp, ctx.height)
	}

	// Verify logo is present in both
	if !strings.Contains(viewNoHelp, "Ink") {
		t.Error("View() without help: missing logo")
	}
	if !strings.Contains(viewWithHelp, "Ink") {
		t.Error("View() with help: missing logo")
	}
}
