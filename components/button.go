package components

import (
	"strings"

	t "github.com/gdamore/tcell/v2"
	"github.com/mecha/journal/theme"
)

type ButtonProps struct {
	Pos      Pos
	Text     string
	Padding  int
	Shortcut rune
	HasFocus bool
	OnEnter  func()
}

func Button(r Renderer, props ButtonProps) EventHandler {
	btnStyle := theme.Button(props.HasFocus)
	fullText := "  " + props.Text + "  "

	r.PutStrStyled(props.Pos.X, props.Pos.Y, fullText, btnStyle)

	if props.Shortcut > 0 {
		i := strings.IndexRune(props.Text, props.Shortcut)
		if i >= 0 {
			r.PutStrStyled(props.Pos.X+2+i, props.Pos.Y, string(props.Shortcut), btnStyle.Underline(true))
		}
	}

	if props.OnEnter == nil {
		return nil
	}

	return HandleKey(func(ev *t.EventKey) bool {
		shortcut := rune(strings.ToLower(string(props.Shortcut))[0])

		switch ev.Key() {
		case t.KeyEnter:
			if props.HasFocus {
				props.OnEnter()
				return true
			}
		case t.KeyRune:
			if ev.Rune() == shortcut {
				props.OnEnter()
				return true
			}
		}

		return false
	})
}
