package components

import (
	"journal-tui/render"
	"strings"
	"unicode/utf8"

	t "github.com/gdamore/tcell/v2"
)

type InputComponent struct {
	title        string
	onEnterFunc  func(value string)
	clearOnEnter bool
	mask         rune

	text   string
	cursor int
}

var _ Component = (*InputComponent)(nil)

func NewInputComponent(title string, onEnter func(value string)) *InputComponent {
	return &InputComponent{
		title:       title,
		onEnterFunc: onEnter,
		mask:        0,
	}
}

func (in *InputComponent) ClearOnEnter(clearOnEnter bool) *InputComponent {
	in.clearOnEnter = clearOnEnter
	return in
}

func (c *InputComponent) SetValue(value string) *InputComponent {
	c.text = value
	c.cursor = len(c.text)
	return c
}

func (c *InputComponent) SetMask(mask rune) *InputComponent {
	c.mask = mask
	return c
}

func (c *InputComponent) HandleEvent(ev t.Event) bool {
	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyRune:
			if c.cursor == len(c.text) {
				c.text += string(ev.Rune())
			} else {
				c.text = c.text[:c.cursor] + string(ev.Rune()) + c.text[c.cursor:]
			}
			c.cursor++

		case t.KeyLeft:
			c.cursor = max(0, c.cursor-1)
		case t.KeyRight:
			c.cursor = min(len(c.text), c.cursor+1)

		case t.KeyBackspace, t.KeyBackspace2:
			if len(c.text) > 0 && c.cursor > 0 {
				c.text = c.text[:c.cursor-1] + c.text[c.cursor:]
				c.cursor--
			}

		case t.KeyEnter:
			go c.onEnterFunc(c.text)
			c.SetValue("")
		}
	}
	return false
}

func (c *InputComponent) Render(screen t.Screen, bounds Rect, hasFocus bool) {
	x, y, w, h := bounds.XYWH()

	render.Panel(screen, c.title, x, y, w, h, render.RoundedBorders, hasFocus)

	text := c.text
	maxLength := w - 3
	if len(c.text) > maxLength {
		text = c.text[:len(c.text)-maxLength]
	}

	if c.mask != 0 {
		text = strings.Repeat(string(c.mask), utf8.RuneCountInString(text))
	}

	screen.PutStr(x+1, y+1, text)
	if hasFocus {
		screen.ShowCursor(x+1+c.cursor, y+1)
	}
}
