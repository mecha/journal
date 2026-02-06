package components

import (
	"journal-tui/render"
	"strings"

	t "github.com/gdamore/tcell/v2"
)

type Text struct {
	lines []string
	style t.Style
	pad   string
}

var _ Component = (*Text)(nil)

func NewText(text string) *Text {
	c := &Text{nil, t.StyleDefault, " "}
	c.SetText(text)
	return c
}

func (t *Text) Style(style t.Style) *Text {
	t.style = style
	return t
}

func (t *Text) Pad(pad string) *Text {
	t.pad = pad
	return t
}

func (t *Text) SetText(text string) *Text {
	t.lines = strings.Split(text, "\n")
	return t
}

func (t *Text) HandleEvent(ev t.Event) bool {
	return false
}

func (t *Text) Render(screen t.Screen, bounds Rect, hasFocus bool) {
	x, y, w, h := bounds.XYWH()
	visibleLines := t.lines
	if h < len(t.lines) {
		visibleLines = t.lines[:h]
	}

	for i, line := range visibleLines {
		screen.PutStrStyled(x, y+i, render.FixedString(line, w, t.pad), t.style)
	}
}
