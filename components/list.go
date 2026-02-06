package components

import (
	"journal-tui/render"
	"journal-tui/theme"
	"slices"

	t "github.com/gdamore/tcell/v2"
)

type List struct {
	items   []string
	onEnter ListOnEnterFunc

	cursor   int
	vscroll  int
	hscroll  int
	lastSize Size
}

type ListOnEnterFunc func(i int, item string)

var _ Component = (*List)(nil)

func NewList(items []string) *List {
	return &List{
		items,
		func(i int, item string) {},
		0,
		0,
		0,
		Size{0, 0},
	}
}

func (p *List) OnEnter(onEnter ListOnEnterFunc) {
	p.onEnter = onEnter
}

func (p *List) AddItem(item string) {
	p.items = append(p.items, item)
	p.MoveCursor(0)
}

func (p *List) SetItems(items []string) bool {
	equal := slices.Equal(p.items, items)
	if !equal {
		p.items = items
		p.MoveCursor(0)
	}
	return !equal
}

func (p *List) MoveCursor(n int) {
	p.cursor = max(0, min(len(p.items)-1, p.cursor+n))
	h := max(3, p.lastSize.H)

	pageSize := max(0, h-2)
	topTarget := p.cursor - 2
	bottomTarget := p.cursor + 2 - pageSize

	switch {
	case topTarget < p.vscroll:
		p.vscroll = max(0, topTarget)
	case bottomTarget > p.vscroll:
		p.vscroll = min(len(p.items)-pageSize, bottomTarget)
	}
}

func (p *List) Up()       { p.MoveCursor(-1) }
func (p *List) Down()     { p.MoveCursor(1) }
func (p *List) PageUp()   { p.MoveCursor(-p.lastSize.H - 2) }
func (p *List) PageDown() { p.MoveCursor(p.lastSize.H - 2) }
func (p *List) Top()      { p.cursor = 0 }
func (p *List) Bottom()   { p.cursor = len(p.items) - 1 }

func (p *List) ScrollLeft() {
	p.hscroll = max(0, p.hscroll-1)
}

func (p *List) ScrollRight() {
	maxWidth := render.MaxLength(p.items) - p.lastSize.W + 2
	p.hscroll = min(maxWidth, p.hscroll+1)
}

func (p *List) HandleEvent(ev t.Event) bool {
	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyRune:
			switch ev.Rune() {
			case 'k':
				p.Up()
			case 'j':
				p.Down()
			case ',':
				p.PageUp()
			case '.':
				p.PageDown()
			case '<':
				p.Top()
			case '>':
				p.Bottom()
			}
		case t.KeyUp:
			p.Up()
		case t.KeyDown:
			p.Down()
		case t.KeyLeft:
			p.ScrollLeft()
		case t.KeyRight:
			p.ScrollRight()
		case t.KeyEnter:
			if p.cursor < len(p.items) {
				p.onEnter(p.cursor, p.items[p.cursor])
			}
		}
	}
	return false
}

func (p *List) Render(screen t.Screen, bounds Rect, hasFocus bool) {
	x, y, w, h := bounds.XYWH()
	p.lastSize = bounds.Size

	for i := range h {
		index := p.vscroll + i

		style := t.StyleDefault
		if hasFocus && index == p.cursor {
			style = theme.CalendarSelect
		}

		if index < len(p.items) {
			text := render.ScrollString(p.items[index], p.hscroll, w, " ")
			screen.PutStrStyled(x, y+i, text, style)
		}
	}
}
