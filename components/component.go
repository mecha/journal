package components

import t "github.com/gdamore/tcell/v2"

type Component interface {
	HandleEvent(ev t.Event) (consumed bool)
	Render(renderer Renderer, hasFocus bool)
}

type customComponent[S any] struct {
	state S
	RenderFunc[S]
	EventHandlerFunc[S]
}

type RenderFunc[S any] func(state S, renderer Renderer, hasFocus bool)
type EventHandlerFunc[S any] func(state S, ev t.Event) (consumed bool)
type EventHandler func(ev t.Event) bool

func NewComponent[S any](state S, render RenderFunc[S], eventHandler EventHandlerFunc[S]) *customComponent[S] {
	return &customComponent[S]{state, render, eventHandler}
}

func (f *customComponent[State]) Render(r Renderer, hasFocus bool) {
	if f.RenderFunc != nil {
		f.RenderFunc(f.state, r, hasFocus)
	}
}

func (f *customComponent[State]) HandleEvent(ev t.Event) bool {
	if f.EventHandlerFunc != nil {
		return f.EventHandlerFunc(f.state, ev)
	}
	return false
}
