package components

import t "github.com/gdamore/tcell/v2"

type FocusToggle struct {
	content Component
}

var _ Component = (*FocusToggle)(nil)

func NewFocusToggle(content Component) *FocusToggle {
	return &FocusToggle{content}
}

func (ft *FocusToggle) HandleEvent(ev t.Event) bool {
	return ft.content.HandleEvent(ev)
}

func (ft *FocusToggle) Render(r Renderer, hasFocus bool) {
	if hasFocus {
		ft.content.Render(r, hasFocus)
	}
}
