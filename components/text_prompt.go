package components

import (
	"journal-tui/theme"

	t "github.com/gdamore/tcell/v2"
)

type InputPrompt struct {
	title    string
	input    *Input
	callback InputPromptCallback
}

type InputPromptCallback = func(input *Input, cancelled bool)

func NewInputPrompt(title string, input *Input, callback InputPromptCallback) *InputPrompt {
	return &InputPrompt{
		title:    title,
		input:    input,
		callback: callback,
	}
}

func (p *InputPrompt) Input() *Input {
	return p.input
}

func (p *InputPrompt) HandleEvent(ev t.Event) bool {
	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyEsc:
			p.callback(p.input, true)
			return true
		case t.KeyEnter:
			p.callback(p.input, false)
			return true
		}
	}
	return p.input.HandleEvent(ev)
}

func (p *InputPrompt) Render(r Renderer, hasFocus bool) {
	r.Fill(' ', theme.Dialog())
	panelRegion := DrawPanel(r, p.title, theme.Borders(hasFocus, theme.Dialog()))
	p.input.Render(panelRegion, hasFocus)
}
