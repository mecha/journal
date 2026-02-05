package components

import (
	"journal-tui/render"
	"journal-tui/theme"
	"slices"

	t "github.com/gdamore/tcell/v2"
)

type ListComponent struct {
	ComponentPos
	ComponentSize
	title   string
	items   []string
	onEnter ListItemOnEnterFunc

	cursor, vscroll, hscroll int
}

type ListItemOnEnterFunc func(i int, item string)

var _ Component = (*ListComponent)(nil)

func NewListComponent(title string, x, y, w, h int) *ListComponent {
	h = max(3, h)
	return &ListComponent{
		ComponentPos{x, y},
		ComponentSize{w, h},
		title,
		[]string{},
		func(i int, item string) {},
		0,
		0,
		0,
	}
}

func (p *ListComponent) OnEnter(onEnter ListItemOnEnterFunc) {
	p.onEnter = onEnter
}

func (p *ListComponent) AddItem(item string) {
	p.items = append(p.items, item)
	p.MoveCursor(0)
}

func (p *ListComponent) SetItems(items []string) bool {
	equal := slices.Equal(p.items, items)
	if !equal {
		p.items = items
		p.MoveCursor(0)
	}
	return !equal
}

func (p *ListComponent) MoveCursor(n int) {
	p.cursor = max(0, min(len(p.items)-1, p.cursor+n))

	pageSize := max(0, p.h-2)
	topTarget := p.cursor - 2
	bottomTarget := p.cursor + 2 - pageSize

	switch {
	case topTarget < p.vscroll:
		p.vscroll = max(0, topTarget)
	case bottomTarget > p.vscroll:
		p.vscroll = min(len(p.items)-pageSize, bottomTarget)
	}
}

func (p *ListComponent) Up()       { p.MoveCursor(-1) }
func (p *ListComponent) Down()     { p.MoveCursor(1) }
func (p *ListComponent) PageUp()   { p.MoveCursor(-p.h - 2) }
func (p *ListComponent) PageDown() { p.MoveCursor(p.h - 2) }
func (p *ListComponent) Top()      { p.cursor = 0 }
func (p *ListComponent) Bottom()   { p.cursor = len(p.items) - 1 }

func (p *ListComponent) ScrollLeft() {
	p.hscroll = max(0, p.hscroll-1)
}

func (p *ListComponent) ScrollRight() {
	maxWidth := render.MaxLength(p.items) - p.w + 2
	p.hscroll = min(maxWidth, p.hscroll+1)
}

func (p *ListComponent) HandleEvent(ev t.Event) bool {
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
			p.onEnter(p.cursor, p.items[p.cursor])
		}
	}
	return false
}

func (p *ListComponent) Render(screen t.Screen, hasFocus bool) {
	render.Panel(screen, p.title, p.x, p.y, p.w, p.h, render.RoundedBorders, hasFocus)

	for i := range p.h - 2 {
		index := p.vscroll + i

		style := t.StyleDefault
		if hasFocus && index == p.cursor {
			style = theme.CalendarSelect
		}

		if index < len(p.items) {
			text := render.ScrollString(p.items[index], p.hscroll, p.w-2, " ")
			screen.PutStrStyled(p.x+1, p.y+i+1, text, style)
		}
	}
}
