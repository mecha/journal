package components

import t "github.com/gdamore/tcell/v2"

type KeyHandler struct {
	content     Component
	handlerFunc KeyHandlerFunc
}

var _ Component = (*KeyHandler)(nil)

type KeyHandlerFunc func(ev *t.EventKey) bool

func NewKeyHandler(content Component, handlerFunc KeyHandlerFunc) *KeyHandler {
	return &KeyHandler{content, handlerFunc}
}

func (kh *KeyHandler) HandleEvent(ev t.Event) bool {
	var consumed = false
	switch ev := ev.(type) {
	case *t.EventKey:
		consumed = kh.handlerFunc(ev)
	}

	if !consumed {
		consumed = kh.content.HandleEvent(ev)
	}

	return consumed
}

func (kh *KeyHandler) Render(r Renderer, hasFocus bool) {
	kh.content.Render(r, hasFocus)
}
