package components

import t "github.com/gdamore/tcell/v2"

type Toggle struct {
	content Component
	shown   bool
}

var _ Component = (*Toggle)(nil)

func NewToggle(content Component) *Toggle {
	return &Toggle{content, false}
}

func (m *Toggle) IsShown() bool {
	return m.shown
}

func (m *Toggle) Show(show bool) *Toggle {
	m.shown = show
	return m
}

func (m *Toggle) HandleEvent(ev t.Event) bool {
	return m.HandleEvent(ev)
}

func (m *Toggle) Render(screen t.Screen, region Rect, hasFocus bool) {
	if m.shown {
		m.content.Render(screen, region, hasFocus)
	}
}
