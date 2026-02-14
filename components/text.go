package components

import (
	"github.com/mecha/journal/utils"

	t "github.com/gdamore/tcell/v2"
)

type TextProps struct {
	State *TextState
	Style t.Style
}

type TextState struct {
	Scroll Pos
	Lines  []string
}

func Text(r Renderer, props TextProps) EventHandler {
	state := props.State
	width, height := r.Size()
	maxLength := utils.MaxLength(state.Lines)
	setScroll := func(pos Pos) {
		state.Scroll.X = max(0, min(maxLength-width, pos.X))
		state.Scroll.Y = max(0, min(len(state.Lines)-height, pos.Y))
	}
	setScroll(state.Scroll)

	topLine := max(0, state.Scroll.Y)
	lastLine := min(len(state.Lines), topLine+height)

	for i, line := range state.Lines[topLine:lastLine] {
		if len(line) == 0 {
			continue
		}
		left := max(0, state.Scroll.X)
		right := min(len(line), state.Scroll.X+width)
		row := utils.FixedString(line[left:right], width, " ")
		r.PutStrStyled(0, i, row, props.Style)
	}

	return HandleKey(func(ev *t.EventKey) bool {
		switch ev.Key() {
		default:
			return false
		case t.KeyRune:
			switch ev.Rune() {
			default:
				return false
			case 'h':
				setScroll(state.Scroll.Add(-1, 0))
			case 'j':
				setScroll(state.Scroll.Add(0, 1))
			case 'k':
				setScroll(state.Scroll.Add(0, -1))
			case 'l':
				setScroll(state.Scroll.Add(1, 0))
			case ',':
				setScroll(state.Scroll.Add(0, -10))
			case '.':
				setScroll(state.Scroll.Add(0, 10))
			}
		case t.KeyLeft:
			setScroll(state.Scroll.Add(-1, 0))
		case t.KeyDown:
			setScroll(state.Scroll.Add(0, 1))
		case t.KeyUp:
			setScroll(state.Scroll.Add(0, -1))
		case t.KeyRight:
			setScroll(state.Scroll.Add(1, 0))
		}
		return true
	})
}
