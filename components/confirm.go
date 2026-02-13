package components

import (
	"github.com/mecha/journal/theme"
	"github.com/mecha/journal/utils"

	t "github.com/gdamore/tcell/v2"
)

type ConfirmProps struct {
	Yes, No  string
	Message  string
	Value    bool
	Borders  BorderSet
	Style    t.Style
	OnSelect func(value bool)
	OnChoice func(accepted bool)
}

func Confirm(r Renderer, hasFocus bool, props ConfirmProps) EventHandler {
	maxWidth, _ := r.Size()

	btnsWidth := len(props.Yes) + 4 + len(props.No) + 4 + 2
	width := max(maxWidth, btnsWidth)

	lines := utils.WrapString(props.Message, width-2)
	height := len(lines) + 3

	region := CenteredRegion(r, width, height)
	region.Fill(' ', theme.Dialog())

	btnHandler := Box(region, BoxProps{
		Borders: props.Borders,
		Style:   props.Style,
		Children: func(r Renderer) EventHandler {
			w, h := r.Size()

			for i, line := range lines {
				r.PutStr(0, i, line)
			}

			noBtnX := w - 1 - len(props.No) - 4
			yesBtnX := noBtnX - 1 - len(props.Yes) - 4

			noHandler := Button(r, ButtonProps{
				Pos:      Pos{noBtnX, h - 1},
				Text:     props.No,
				Shortcut: 'N',
				HasFocus: props.Value == false,
				OnEnter:  func() { props.OnChoice(false) },
			})

			yesHandler := Button(r, ButtonProps{
				Pos:      Pos{yesBtnX, h - 1},
				Text:     props.Yes,
				Shortcut: 'Y',
				HasFocus: props.Value == true,
				OnEnter:  func() { props.OnChoice(true) },
			})

			return func(ev t.Event) bool {
				if noHandler != nil && noHandler(ev) {
					return true
				}
				if yesHandler != nil && yesHandler(ev) {
					return true
				}
				return false
			}
		},
	})

	return HandleKey(func(ev *t.EventKey) bool {
		if btnHandler != nil && btnHandler(ev) {
			return true
		}

		switch ev.Key() {
		case t.KeyEsc:
			if props.OnChoice != nil {
				props.OnChoice(false)
			}
			return true

		case t.KeyLeft, t.KeyRight, t.KeyTab:
			if props.OnSelect != nil {
				props.OnSelect(!props.Value)
			}
			return true
		}

		return false
	})
}
