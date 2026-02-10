package components

import (
	"fmt"
	"unicode/utf8"

	"github.com/mecha/journal/theme"
	"github.com/mecha/journal/utils"

	t "github.com/gdamore/tcell/v2"
)

type List[Item any] struct {
	items      []Item
	renderFunc ListRenderFunc[Item]
	onEnter    ListItemFunc[Item]
	onSelect   ListItemFunc[Item]

	cursor   int
	vscroll  int
	hscroll  int
	lastSize Size
}

type ListRenderFunc[Item any] func(item Item) string
type ListItemFunc[Item any] func(i int, item Item)

func NewList[Item any](items []Item) *List[Item] {
	return &List[Item]{
		items:      items,
		renderFunc: func(item Item) string { return fmt.Sprintf("%v", item) },
		onEnter:    func(i int, item Item) {},
	}
}

func (l *List[Item]) RenderWith(renderFunc ListRenderFunc[Item]) *List[Item] {
	l.renderFunc = renderFunc
	return l
}

func (l *List[Item]) OnEnter(onEnter ListItemFunc[Item]) *List[Item] {
	l.onEnter = onEnter
	return l
}

func (l *List[Item]) OnSelect(onSelect ListItemFunc[Item]) *List[Item] {
	l.onSelect = onSelect
	return l
}

func (l *List[Item]) AddItem(item Item) {
	l.items = append(l.items, item)
	l.MoveCursor(0)
}

func (l *List[Item]) SetItems(items []Item) {
	l.items = items
	l.MoveCursor(0)
}

func (l *List[Item]) MoveCursor(n int) {
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

	if l.onSelect != nil && l.cursor < len(l.items) {
		l.onSelect(l.cursor, l.items[l.cursor])
	}
}

func (l *List[Item]) Up()       { l.MoveCursor(-1) }
func (l *List[Item]) Down()     { l.MoveCursor(1) }
func (l *List[Item]) PageUp()   { l.MoveCursor(-l.lastSize.H - 2) }
func (l *List[Item]) PageDown() { l.MoveCursor(l.lastSize.H - 2) }
func (l *List[Item]) Top()      { l.cursor = 0 }
func (l *List[Item]) Bottom()   { l.cursor = len(l.items) - 1 }

func (l *List[Item]) ScrollLeft() {
	l.hscroll = max(0, l.hscroll-1)
}

func (l *List[Item]) ScrollRight() {
	maxLength := 0
	for _, item := range l.items {
		length := utf8.RuneCountInString(l.renderFunc(item))
		if length > maxLength {
			maxLength = length
		}
	}
	maxHScroll := maxLength - l.lastSize.W
	l.hscroll = min(maxHScroll, l.hscroll+1)
}

func (l *List[Item]) HandleEvent(ev t.Event) bool {
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

func (l *List[Item]) Render(r Renderer, hasFocus bool) {
	width, height := r.Size()
	l.lastSize = Size{width, height}

	for i := range height {
		index := l.vscroll + i

		if index < len(l.items) {
			itemStr := l.renderFunc(l.items[index])
			text := utils.ScrollString(itemStr, l.hscroll, width, " ")

			isSelected := hasFocus && index == l.cursor
			style := theme.ListItem(isSelected)

			r.PutStrStyled(0, i, text, style)
		}
	}
}
