package components

import (
	"unicode/utf8"

	"github.com/mecha/journal/theme"
	"github.com/mecha/journal/utils"

	t "github.com/gdamore/tcell/v2"
)

type ListProps[Item any] struct {
	State        *ListState[Item]
	Items        []Item
	ShowSelected bool
	RenderFunc   ListRenderFunc[Item]
	OnEnter      ListItemFunc[Item]
	OnSelect     ListItemFunc[Item]
}

type ListState[Item any] struct {
	Cursor   int
	VScroll  int
	HScroll  int
	LastSize Size
}

type ListRenderFunc[Item any] func(item Item) string
type ListItemFunc[Item any] func(i int, item Item)

func List[Item any](r Renderer, props ListProps[Item]) EventHandler {
	width, height := r.Size()
	state := props.State
	state.LastSize = Size{width, height}

	rendered := make([]string, 0, min(height, len(props.Items)))

	for i := range height {
		index := state.VScroll + i

		if index < len(props.Items) {
			itemStr := props.RenderFunc(props.Items[index])
			text := utils.ScrollString(itemStr, state.HScroll, width, " ")
			rendered = append(rendered, text)

			isSelected := props.ShowSelected && index == state.Cursor
			style := theme.ListItem(isSelected)

			r.PutStrStyled(0, i, text, style)
		}
	}

	return HandleKey(func(ev *t.EventKey) bool {
		moveCursor := func(offset int) {
			numItems := len(props.Items)
			state.Cursor = max(0, min(numItems-1, state.Cursor+offset))
			h := max(3, state.LastSize.H)

			pageSize := max(0, h-2)
			topTarget := state.Cursor - 2
			bottomTarget := state.Cursor + 2 - pageSize

			switch {
			case topTarget < state.VScroll:
				state.VScroll = max(0, topTarget)
			case bottomTarget > state.VScroll:
				state.VScroll = min(numItems-pageSize, bottomTarget)
			}

			if props.OnSelect != nil {
				props.OnSelect(state.Cursor, props.Items[state.Cursor])
			}
		}

		switch ev.Key() {
		case t.KeyRune:
			switch ev.Rune() {
			case 'k':
				moveCursor(-1)
			case 'j':
				moveCursor(1)
			case ',':
				moveCursor(-state.LastSize.H - 2)
			case '.':
				moveCursor(state.LastSize.H - 2)
			case '<':
				state.Cursor = 0
			case '>':
				state.Cursor = len(props.Items) - 1
			}
		case t.KeyUp:
			moveCursor(-1)
		case t.KeyDown:
			moveCursor(1)
		case t.KeyLeft:
			state.HScroll = max(0, state.HScroll-1)
		case t.KeyRight:
			maxLength := 0
			for _, item := range props.Items {
				length := utf8.RuneCountInString(props.RenderFunc(item))
				if length > maxLength {
					maxLength = length
				}
			}
			maxHScroll := maxLength - state.LastSize.W
			state.HScroll = min(maxHScroll, state.HScroll+1)
		case t.KeyEnter:
			if props.OnEnter != nil && state.Cursor < len(props.Items) {
				props.OnEnter(state.Cursor, props.Items[state.Cursor])
			}
		}

		return false
	})
}
