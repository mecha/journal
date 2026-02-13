package components

import (
	"unicode/utf8"

	"github.com/mecha/journal/theme"
	"github.com/mecha/journal/utils"

	t "github.com/gdamore/tcell/v2"
)

type ListState[Item any] struct {
	Cursor   int
	VScroll  int
	HScroll  int
	LastSize Size
}

type ListProps[Item any] struct {
	Items        []Item
	ShowSelected bool
	RenderFunc   ListRenderFunc[Item]
	OnEnter      ListItemFunc[Item]
	OnSelect     ListItemFunc[Item]
}

type ListRenderFunc[Item any] func(item Item) string
type ListItemFunc[Item any] func(i int, item Item)

func DrawList[Item any](r Renderer, s *ListState[Item], p ListProps[Item]) EventHandler {
	width, height := r.Size()
	s.LastSize = Size{width, height}

	for i := range height {
		index := s.VScroll + i

		if index < len(p.Items) {
			itemStr := p.RenderFunc(p.Items[index])
			text := utils.ScrollString(itemStr, s.HScroll, width, " ")

			isSelected := p.ShowSelected && index == s.Cursor
			style := theme.ListItem(isSelected)

			r.PutStrStyled(0, i, text, style)
		}
	}

	return func(ev t.Event) bool {
		moveCursor := func(offset int) {
			numItems := len(p.Items)
			s.Cursor = max(0, min(numItems-1, s.Cursor+offset))
			h := max(3, s.LastSize.H)

			pageSize := max(0, h-2)
			topTarget := s.Cursor - 2
			bottomTarget := s.Cursor + 2 - pageSize

			switch {
			case topTarget < s.VScroll:
				s.VScroll = max(0, topTarget)
			case bottomTarget > s.VScroll:
				s.VScroll = min(numItems-pageSize, bottomTarget)
			}
		}

		switch ev := ev.(type) {
		case *t.EventKey:
			switch ev.Key() {
			case t.KeyRune:
				switch ev.Rune() {
				case 'k':
					moveCursor(-1)
				case 'j':
					moveCursor(1)
				case ',':
					moveCursor(-s.LastSize.H - 2)
				case '.':
					moveCursor(s.LastSize.H - 2)
				case '<':
					s.Cursor = 0
				case '>':
					s.Cursor = len(p.Items) - 1
				}
			case t.KeyUp:
				moveCursor(-1)
			case t.KeyDown:
				moveCursor(1)
			case t.KeyLeft:
				s.HScroll = max(0, s.HScroll-1)
			case t.KeyRight:
				maxLength := 0
				for _, item := range p.Items {
					length := utf8.RuneCountInString(p.RenderFunc(item))
					if length > maxLength {
						maxLength = length
					}
				}
				maxHScroll := maxLength - s.LastSize.W
				s.HScroll = min(maxHScroll, s.HScroll+1)
			case t.KeyEnter:
				if p.OnEnter != nil && s.Cursor < len(p.Items) {
					p.OnEnter(s.Cursor, p.Items[s.Cursor])
				}
			}
		}
		return false
	}
}
