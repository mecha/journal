package components

import (
	"journal-tui/render"
	"strings"

	t "github.com/gdamore/tcell/v2"
)

type TextComponent struct {
	ComponentPos
	ComponentSize
	lines []string
	style t.Style
	pad   string
}

var _ Component = (*TextComponent)(nil)

func NewTextComponent(text string, style t.Style, pad string) *TextComponent {
	c := &TextComponent{
		ComponentPos{0, 0},
		ComponentSize{0, 1},
		nil,
		style,
		pad,
	}
	c.SetText(text)
	return c
}

func (c *TextComponent) SetText(text string) *TextComponent {
	c.lines = strings.Split(text, "\n")
	return c
}

func (c *TextComponent) HandleEvent(ev t.Event) bool {
	return false
}

func (c *TextComponent) Render(screen t.Screen, hasFocus bool) {
	visibleLines := c.lines
	if c.h < len(c.lines) {
		visibleLines = c.lines[:c.h]
	}

	for i, line := range visibleLines {
		screen.PutStrStyled(c.x, c.y+i, render.FixedString(line, c.w, c.pad), c.style)
	}
}
