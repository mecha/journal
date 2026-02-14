package components

import (
	"strings"
	"unicode/utf8"

	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type InputProps struct {
	State      *InputState
	HideCursor bool
	Mask       string
}

type InputState struct {
	Value  string
	Cursor int
}

func Input(r Renderer, props InputProps) EventHandler {
	state := props.State
	width, _ := r.Size()
	maxRunes := width - 1

	var text string
	if utf8.RuneCountInString(state.Value) <= maxRunes {
		text = state.Value
	} else {
		text = state.Value[len(state.Value)-maxRunes:]
	}

	if len(props.Mask) > 0 {
		text = strings.Repeat(props.Mask, len(text))
	}

	r.Fill(' ', theme.Input())
	r.PutStrStyled(0, 0, text, theme.Input())
	if !props.HideCursor {
		r.ShowCursor(min(maxRunes, state.Cursor), 0)
	}

	return func(ev t.Event) bool {
		switch ev := ev.(type) {
		default:
			return false
		case *t.EventKey:
			switch ev.Key() {
			default:
				return false
			case t.KeyRune:
				if state.Cursor == len(state.Value) {
					state.Value += string(ev.Rune())
				} else {
					state.Value = state.Value[:state.Cursor] + string(ev.Rune()) + state.Value[state.Cursor:]
				}
				state.Cursor += 1

			case t.KeyLeft:
				state.Cursor = max(0, state.Cursor-1)
			case t.KeyRight:
				state.Cursor = min(len(state.Value), state.Cursor+1)

			case t.KeyBackspace, t.KeyBackspace2:
				if len(state.Value) > 0 && state.Cursor > 0 {
					state.Value = state.Value[:state.Cursor-1] + state.Value[state.Cursor:]
					state.Cursor -= 1
				}
			}
		}
		return true
	}
}
