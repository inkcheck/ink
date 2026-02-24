package model

// Inter-view messages

// OpenChapterMsg requests switching to the Chapter view for the given file.
type OpenChapterMsg struct {
	FilePath string
}

// OpenEditorMsg requests switching to the Editor view.
type OpenEditorMsg struct {
	FilePath string
	Content  string
}

// CloseEditorMsg signals the editor has closed.
type CloseEditorMsg struct {
	FilePath string
}

// OpenExternalEditorMsg requests opening the file in $EDITOR.
type OpenExternalEditorMsg struct {
	FilePath string
}

// ExternalEditorClosedMsg signals the external editor has exited.
type ExternalEditorClosedMsg struct {
	Err error
}

// OpenMetricsMsg requests switching to the Metrics view.
type OpenMetricsMsg struct {
	FilePath string
}

// CloseMetricsMsg signals the metrics view has closed.
type CloseMetricsMsg struct{}

// BackToBookMsg signals returning to the Book view.
type BackToBookMsg struct{}

// FileSavedMsg signals a file was saved successfully.
type FileSavedMsg struct{}
