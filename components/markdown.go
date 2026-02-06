package components

import (
	"journal-tui/render"
	"strings"

	t "github.com/gdamore/tcell/v2"
)

type Markdown struct {
	content []string
	vscroll int
	hscroll int
}

var _ Component = (*Markdown)(nil)

func NewMarkdownComponent(markdown string) *Markdown {
	c := &Markdown{
		[]string{},
		0,
		0,
	}
	c.SetContent(markdown)
	return c
}

func (c *Markdown) SetContent(content string) {
	c.content = strings.Split(content, "\n")
}

func (c *Markdown) HandleEvent(ev t.Event) bool {
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

func (c *Markdown) Render(screen t.Screen, bounds Rect, hasFocus bool) {
	x, y, w, h := bounds.XYWH()

	maxLineLen := max(0, w)
	maxNumLines := max(0, h)

	topLine := max(0, c.vscroll)
	botLine := min(len(c.content), topLine+maxNumLines)
	visibleLines := c.content[topLine:botLine]

	for i, line := range visibleLines {
		// screen.PutStr(x, y+i, render.ScrollString(line, c.hscroll, maxLineLen, " "))
		screen.PutStr(x, y+i, render.FixedString(line, maxLineLen, " "))
	}
}
