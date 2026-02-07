package components

import t "github.com/gdamore/tcell/v2"

type Component interface {
	HandleEvent(ev t.Event) (consumed bool)
	Render(renderer Renderer, hasFocus bool)
}

type RenderFunc func(renderer Renderer, hasFocus bool)
type EventHandlerFunc func(ev t.Event) (consumed bool)

type customComponent struct {
	RenderFunc
	EventHandlerFunc
}

func NewComponent(render RenderFunc, eventHandler EventHandlerFunc) *customComponent {
	return &customComponent{render, eventHandler}
}

func (f *customComponent) HandleEvent(ev t.Event) bool {
	if f.EventHandlerFunc != nil {
		return f.EventHandlerFunc(ev)
	}
	return false
}

func (f *customComponent) Render(r Renderer, hasFocus bool) {
	if f.RenderFunc != nil {
		f.RenderFunc(r, hasFocus)
	}
}
