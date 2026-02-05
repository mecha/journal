package components

import (
	"journal-tui/render"
	"strings"

	t "github.com/gdamore/tcell/v2"
)

type MarkdownComponent struct {
	ComponentPos
	ComponentSize

	title   string
	content []string
	vscroll int
	hscroll int
}

var _ Component = (*MarkdownComponent)(nil)

func NewMarkdownComponent(title, markdown string) *MarkdownComponent {
	c := &MarkdownComponent{
		ComponentPos{0, 0},
		ComponentSize{80, 30},
		title,
		[]string{},
		0,
		0,
	}
	c.SetContent(markdown)
	return c
}

func (c *MarkdownComponent) SetContent(content string) {
	c.content = strings.Split(content, "\n")
}

func (c *MarkdownComponent) HandleEvent(ev t.Event) bool {
	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyUp:
			c.vscroll = max(0, c.vscroll-1)
		case t.KeyDown:
			c.vscroll = min(len(c.content)-1, c.vscroll+1)
		case t.KeyLeft:
			c.hscroll = max(0, c.hscroll-1)
		case t.KeyRight:
			maxWidth := render.MaxLength(c.content)
			c.hscroll = min(maxWidth, c.hscroll+1)
		}
	}
	return false
}

func (c *MarkdownComponent) Render(screen t.Screen, hasFocus bool) {
	render.Panel(screen, c.title, c.x, c.y, c.w, c.h, render.RoundedBorders, hasFocus)

	maxLineLen := max(0, c.w-2)
	maxNumLines := max(0, c.h-2)

	topLine := max(0, c.vscroll)
	botLine := min(len(c.content), topLine+maxNumLines)
	visibleLines := c.content[topLine:botLine]

	for i, line := range visibleLines {
		// screen.PutStr(c.x+1, c.y+1+i, render.ScrollString(line, c.hscroll, maxLineLen, " "))
		screen.PutStr(c.x+1, c.y+1+i, render.FixedString(line, maxLineLen, " "))
	}
}
