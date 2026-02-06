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

func (l *List) OnEnter(onEnter ListOnEnterFunc) *List {
	l.onEnter = onEnter
	return l
}

func (l *List) AddItem(item string) {
	l.items = append(l.items, item)
	l.MoveCursor(0)
}

func (l *List) SetItems(items []string) bool {
	equal := slices.Equal(l.items, items)
	if !equal {
		l.items = items
		l.MoveCursor(0)
	}
	return !equal
}

func (l *List) MoveCursor(n int) {
	l.cursor = max(0, min(len(l.items)-1, l.cursor+n))
	h := max(3, l.lastSize.H)

	pageSize := max(0, h-2)
	topTarget := l.cursor - 2
	bottomTarget := l.cursor + 2 - pageSize

	switch {
	case topTarget < l.vscroll:
		l.vscroll = max(0, topTarget)
	case bottomTarget > l.vscroll:
		l.vscroll = min(len(l.items)-pageSize, bottomTarget)
	}
}

func (l *List) Up()       { l.MoveCursor(-1) }
func (l *List) Down()     { l.MoveCursor(1) }
func (l *List) PageUp()   { l.MoveCursor(-l.lastSize.H - 2) }
func (l *List) PageDown() { l.MoveCursor(l.lastSize.H - 2) }
func (l *List) Top()      { l.cursor = 0 }
func (l *List) Bottom()   { l.cursor = len(l.items) - 1 }

func (l *List) ScrollLeft() {
	l.hscroll = max(0, l.hscroll-1)
}

func (l *List) ScrollRight() {
	maxWidth := render.MaxLength(l.items) - l.lastSize.W + 2
	l.hscroll = min(maxWidth, l.hscroll+1)
}

func (l *List) HandleEvent(ev t.Event) bool {
	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyRune:
			switch ev.Rune() {
			case 'k':
				l.Up()
			case 'j':
				l.Down()
			case ',':
				l.PageUp()
			case '.':
				l.PageDown()
			case '<':
				l.Top()
			case '>':
				l.Bottom()
			}
		case t.KeyUp:
			l.Up()
		case t.KeyDown:
			l.Down()
		case t.KeyLeft:
			l.ScrollLeft()
		case t.KeyRight:
			l.ScrollRight()
		case t.KeyEnter:
			if l.cursor < len(l.items) {
				l.onEnter(l.cursor, l.items[l.cursor])
			}
		}
	}
	return false
}

func (l *List) Render(screen t.Screen, bounds Rect, hasFocus bool) {
	x, y, w, h := bounds.XYWH()
	l.lastSize = bounds.Size

	for i := range h {
		index := l.vscroll + i

		style := t.StyleDefault
		if hasFocus && index == l.cursor {
			style = theme.CalendarSelect
		}

		if index < len(l.items) {
			text := render.ScrollString(l.items[index], l.hscroll, w, " ")
			screen.PutStrStyled(x, y+i, text, style)
		}
	}
}
