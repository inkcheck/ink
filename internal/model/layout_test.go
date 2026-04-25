package model

import "testing"

func TestHelpPaneToggleVisibleHide(t *testing.T) {
	hp := NewHelpPane([][]helpEntry{
		{{"a", "b"}},
	})
	if hp.Visible() {
		t.Fatal("expected hidden initially")
	}
	hp.Toggle()
	if !hp.Visible() {
		t.Fatal("expected visible after toggle")
	}
	hp.Toggle()
	if hp.Visible() {
		t.Fatal("expected hidden after second toggle")
	}
	hp.Toggle()
	hp.Hide()
	if hp.Visible() {
		t.Fatal("expected hidden after Hide")
	}
}

func TestHelpPaneHeightIfVisible(t *testing.T) {
	hp := NewHelpPane([][]helpEntry{
		{{"a", "b"}, {"c", "d"}},
		{{"e", "f"}},
	})
	if hp.HeightIfVisible() != 0 {
		t.Fatalf("expected 0 when hidden, got %d", hp.HeightIfVisible())
	}
	hp.Toggle()
	if hp.HeightIfVisible() != 2 {
		t.Fatalf("expected 2 when visible, got %d", hp.HeightIfVisible())
	}
}

func TestHelpPaneHeightMaxAcrossColumns(t *testing.T) {
	hp := NewHelpPane([][]helpEntry{
		{{"a", "b"}},
		{{"c", "d"}, {"e", "f"}, {"g", "h"}},
		{{"i", "j"}, {"k", "l"}},
	})
	if hp.height != 3 {
		t.Fatalf("expected height 3 (max column length), got %d", hp.height)
	}
}

func TestHelpPaneViewHidden(t *testing.T) {
	hp := NewHelpPane([][]helpEntry{
		{{"a", "b"}},
	})
	if hp.View(80) != "" {
		t.Fatal("expected empty string when hidden")
	}
}

func TestNewHelpPaneEmpty(t *testing.T) {
	hp := NewHelpPane(nil)
	if hp.height != 0 {
		t.Fatalf("expected height 0 for nil entries, got %d", hp.height)
	}
	hp2 := NewHelpPane([][]helpEntry{})
	if hp2.height != 0 {
		t.Fatalf("expected height 0 for empty entries, got %d", hp2.height)
	}
}
