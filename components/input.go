package components

import (
	"strings"
	"unicode/utf8"

	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type Input struct {
	onEnterFunc func(input *Input)
	mask        rune

	text   string
	cursor int
}

var _ Component = (*Input)(nil)

func NewInput() *Input {
	return &Input{
		onEnterFunc: nil,
		mask:        0,
	}
}

func (in *Input) OnEnter(onEnter func(input *Input)) *Input {
	in.onEnterFunc = onEnter
	return in
}

func (c *Input) Value() string {
	return c.text
}

func (c *Input) SetValue(value string) *Input {
	c.text = value
	c.cursor = len(c.text)
	return c
}

func (c *Input) SetMask(mask rune) *Input {
	c.mask = mask
	return c
}

func (c *Input) HandleEvent(ev t.Event) bool {
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
				c.onEnterFunc(c)
			}
		}
	}
	return true
}

func (c *Input) Render(r Renderer, hasFocus bool) {
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

	r.Fill(' ', theme.Input())
	r.PutStrStyled(0, 0, text, theme.Input())
	if hasFocus {
		r.ShowCursor(min(maxRunes, c.cursor), 0)
	}
}
