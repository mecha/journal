package main

import (
	"log"
	"strings"

	c "journal-tui/components"
)

func CreateLogs() *c.Panel {
	logText := c.NewText([]string{})

	writer := logText.Writer()
	writer.OnWrite(renderScreen)
	log.SetOutput(writer)

	return c.NewPanel("[4]─Log", logText)
}

type LogWriter struct {
	component *c.Text
}

func (w *LogWriter) Write(data []byte) (int, error) {
	newLines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	w.component.AddLines(newLines)
	w.component.ScrollToBottom()
	return len(data), nil
}
