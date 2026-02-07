package main

import (
	"strings"

	c "journal-tui/components"
)

type LogWriter struct {
	component *c.Text
}

func (w *LogWriter) Write(data []byte) (int, error) {
	newLines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	w.component.AddLines(newLines)
	w.component.ScrollToBottom()
	return len(data), nil
}
