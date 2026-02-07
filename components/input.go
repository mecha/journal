package components

import (
	"strings"
	"unicode/utf8"

	t "github.com/gdamore/tcell/v2"
)

type InputComponent struct {
	onEnterFunc  func(value string)
	clearOnEnter bool
	mask         rune

	text   string
	cursor int
}

var _ Component = (*InputComponent)(nil)

func NewInputComponent() *InputComponent {
	return &InputComponent{
		onEnterFunc: func(value string) {},
		mask:        0,
	}
}

func (in *InputComponent) OnEnter(onEnter func(value string)) *InputComponent {
	in.onEnterFunc = onEnter
	return in
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
	default:
		return false
	case *t.EventKey:
		switch ev.Key() {
		default:
			return false
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
			if c.onEnterFunc != nil {
				c.onEnterFunc(c.text)
				if c.clearOnEnter {
					c.SetValue("")
				}
			}
		}
	}
	return true
}

func (c *InputComponent) Render(r Renderer, hasFocus bool) {
	width, _ := r.Size()
	maxRunes := width - 1

	var text string
	if utf8.RuneCountInString(c.text) <= maxRunes {
		text = c.text
	} else {
		text = c.text[len(c.text)-maxRunes:]
	}

	if c.mask != 0 {
		text = strings.Repeat(string(c.mask), utf8.RuneCountInString(text))
	}

	r.PutStr(0, 0, text)
	if hasFocus {
		r.ShowCursor(min(maxRunes, c.cursor), 0)
	}
}
