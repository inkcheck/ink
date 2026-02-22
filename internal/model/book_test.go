package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommonParentDir(t *testing.T) {
	t.Run("empty paths", func(t *testing.T) {
		got := commonParentDir(nil)
		// Should return abs of "."
		abs, _ := filepath.Abs(".")
		if got != abs {
			t.Errorf("commonParentDir(nil) = %q, want %q", got, abs)
		}
	})

	t.Run("single file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.md")
		os.WriteFile(f, []byte("# Test"), 0644)
		got := commonParentDir([]string{f})
		if got != dir {
			t.Errorf("commonParentDir single = %q, want %q", got, dir)
		}
	})

	t.Run("sibling files", func(t *testing.T) {
		dir := t.TempDir()
		f1 := filepath.Join(dir, "a.md")
		f2 := filepath.Join(dir, "b.md")
		os.WriteFile(f1, []byte("# A"), 0644)
		os.WriteFile(f2, []byte("# B"), 0644)
		got := commonParentDir([]string{f1, f2})
		if got != dir {
			t.Errorf("commonParentDir siblings = %q, want %q", got, dir)
		}
	})

	t.Run("nested files", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		os.MkdirAll(sub, 0755)
		f1 := filepath.Join(dir, "a.md")
		f2 := filepath.Join(sub, "b.md")
		os.WriteFile(f1, []byte("# A"), 0644)
		os.WriteFile(f2, []byte("# B"), 0644)
		got := commonParentDir([]string{f1, f2})
		if got != dir {
			t.Errorf("commonParentDir nested = %q, want %q", got, dir)
		}
	})
}

func TestBookListHeight(t *testing.T) {
	ctx := &ViewContext{width: 80, height: 30, maxWidth: 80}

	t.Run("default", func(t *testing.T) {
		h := bookListHeight(ctx, false, false)
		expected := ctx.height - bookChromeHeight
		if h != expected {
			t.Errorf("bookListHeight() = %d, want %d", h, expected)
		}
	})

	t.Run("with help", func(t *testing.T) {
		h := bookListHeight(ctx, true, false)
		expected := ctx.height - bookChromeHeight - bookHelpHeight
		if h != expected {
			t.Errorf("bookListHeight(help) = %d, want %d", h, expected)
		}
	})

	t.Run("with filtering", func(t *testing.T) {
		h := bookListHeight(ctx, false, true)
		expected := ctx.height - bookChromeHeight + 1
		if h != expected {
			t.Errorf("bookListHeight(filtering) = %d, want %d", h, expected)
		}
	})

	t.Run("minimum height 1", func(t *testing.T) {
		small := &ViewContext{width: 80, height: 3, maxWidth: 80}
		h := bookListHeight(small, true, false)
		if h < 1 {
			t.Errorf("bookListHeight(small) = %d, want >= 1", h)
		}
	})
}

func TestBookViewContainsBookName(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"readme.md": "# Hello",
	})
	ctx := &ViewContext{width: 80, height: 30, maxWidth: 80, isBook: true}
	book := NewBook(ctx, dir)
	view := book.View()

	bookName := dirToBookName(dir)
	if !strings.Contains(view, bookName) {
		t.Errorf("Book.View() missing book name %q", bookName)
	}
}

func TestBookViewContainsFileNames(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"chapter-one.md": "# Chapter One",
		"chapter-two.md": "# Chapter Two",
	})
	ctx := &ViewContext{width: 80, height: 30, maxWidth: 80, isBook: true}
	book := NewBook(ctx, dir)
	view := book.View()

	if !strings.Contains(view, "chapter-one.md") {
		t.Errorf("Book.View() missing file name chapter-one.md")
	}
	if !strings.Contains(view, "chapter-two.md") {
		t.Errorf("Book.View() missing file name chapter-two.md")
	}
}

func TestNewBookFromFilesPreFiltered(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		"a.md": "# A",
		"b.md": "# B",
	})
	files := []string{
		filepath.Join(dir, "a.md"),
		filepath.Join(dir, "b.md"),
	}
	ctx := &ViewContext{width: 80, height: 30, maxWidth: 80, isBook: true}
	book := NewBookFromFiles(ctx, files)
	if !book.preFiltered {
		t.Error("NewBookFromFiles: expected preFiltered to be true")
	}
}

func TestNewBookSkipsHiddenFiles(t *testing.T) {
	dir := tempDirWithFiles(t, map[string]string{
		".hidden.md":  "# Hidden",
		"visible.md":  "# Visible",
	})
	ctx := &ViewContext{width: 80, height: 30, maxWidth: 80, isBook: true}
	book := NewBook(ctx, dir)
	view := book.View()

	if strings.Contains(view, ".hidden.md") {
		t.Error("Book.View() should not show hidden files")
	}
	if !strings.Contains(view, "visible.md") {
		t.Error("Book.View() should show visible files")
	}
}
